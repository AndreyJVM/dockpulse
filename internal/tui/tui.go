package tui

import (
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/shell"
	tea "github.com/charmbracelet/bubbletea"
)

// TabID represents an active tab in the dashboard.
type TabID int

const (
	// TabDocker displays Docker containers, statuses, metrics, and actions.
	TabDocker TabID = 1
	// TabCompose displays Docker Compose projects and aggregated service statuses.
	TabCompose TabID = 2
	// TabNetworks displays Docker networks and their configuration.
	TabNetworks TabID = 3
	// TabVolumes displays Docker volumes and their usage details.
	TabVolumes TabID = 4
	// TabSwarm displays Swarm cluster nodes and services.
	TabSwarm TabID = 5
	// TabCount represents the total number of primary tabs.
	TabCount = 5
)

// ModalType represents an active modal dialog.
type ModalType string

type embeddedShellSession interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Feed(data []byte) error
	Render() string
	Resize(width, height int) error
	Wait() error
	Close() error
}

var startEmbeddedShell = func(cmd *exec.Cmd, width, height int) (embeddedShellSession, error) {
	return shell.Start(cmd, width, height)
}

const (
	// ModalNone indicates no modal is active.
	ModalNone ModalType = ""
	// ModalLogs displays container or service logs.
	ModalLogs ModalType = "logs"
	// ModalInspect displays detailed container inspection data.
	ModalInspect ModalType = "inspect"
	// ModalComposeDown displays confirmation for docker compose down.
	ModalComposeDown ModalType = "compose_down"
	// ModalPrune displays Docker system prune options and confirmations.
	ModalPrune ModalType = "prune"
	// ModalRemoveContainer displays confirmation for removing a container.
	ModalRemoveContainer ModalType = "remove_container"
	// ModalExec displays the interactive container shell launcher.
	ModalExec ModalType = "exec"
)

const (
	SortDefault = iota
	SortCPU
	SortMem
	SortName
	SortStatus
	sortModeCount
)

func tickCmd(d time.Duration) tea.Cmd {
	if d <= 0 {
		d = 2 * time.Second
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model contains the TUI state.
type Model struct {
	version         string
	activeTab       TabID
	refreshInterval time.Duration
	showHelp        bool
	width           int
	height          int

	i18n i18n.Dictionary

	// Navigation and modal state.
	cursorIdx           int
	composeCursorIdx    int
	networkCursorIdx    int
	volumeCursorIdx     int
	swarmCursorIdx      int
	activeModal         ModalType
	execTargetID        string
	execTargetName      string
	execCursorIdx       int
	execIsCustomInput   bool
	execCustomCmd       string
	embeddedShell       embeddedShellSession
	shellFocused        bool
	shellContainerName  string
	logsTargetName      string
	logsTargetID        string
	logsLines           []string
	logsScroll          int
	inspectTarget       monitor.ContainerDetail
	actionMessage       string
	actionMessageExpiry time.Time

	// Compose down modal.
	composeDownTarget string

	// Remove container modal.
	removeContainerTargetID   string
	removeContainerTargetName string

	// Prune modal.
	pruneUsage    monitor.DockerDiskUsage
	pruneUsageErr string
	pruneConfirm  int // 0 - menu, 1 - containers, 2 - images, 3 - volumes, 4 - all

	// Container sorting: default, CPU desc, memory desc, name asc, or status.
	containerSortMode int

	// Container search and filtering.
	isFiltering bool
	filterQuery string

	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	networks   []monitor.DockerNetwork
	volumes    []monitor.DockerVolume
	swarm      monitor.SwarmSummary
	networkErr string
	volumeErr  string
	lastUpdate string
}

// NewModel creates a TUI model with a refresh interval and initial tab.
func NewModel(refreshInterval time.Duration, initialTab int, version string) Model {
	tab := TabID(initialTab)
	if tab < TabDocker || tab > TabCount {
		tab = TabDocker
	}
	if refreshInterval <= 0 {
		refreshInterval = 2 * time.Second
	}
	m := Model{
		version:         version,
		activeTab:       tab,
		refreshInterval: refreshInterval,
		i18n:            i18n.Get(),
	}
	m.fetchAllData()
	return m
}

func (m *Model) applyFetchedData(d dataFetchedMsg) {
	m.dockerInfo = d.dockerInfo
	m.compose = d.compose
	m.networks = d.networks
	m.volumes = d.volumes
	m.swarm = d.swarm
	m.networkErr = d.networkErr
	m.volumeErr = d.volumeErr
	m.lastUpdate = d.timestamp

	m.clampCursor()
	if m.composeCursorIdx >= len(m.compose) {
		m.composeCursorIdx = max(0, len(m.compose)-1)
	}
	if m.swarmCursorIdx >= len(m.swarm.Services) {
		m.swarmCursorIdx = max(0, len(m.swarm.Services)-1)
	}
}

func (m *Model) fetchAllData() {
	dockerInfo := monitor.GetDockerSummary()
	networks, networkErr := monitor.GetDockerNetworks()
	volumes, volumeErr := monitor.GetDockerVolumes()
	m.applyFetchedData(dataFetchedMsg{
		dockerInfo: dockerInfo,
		compose:    monitor.GetComposeProjects(dockerInfo.Containers),
		networks:   networks,
		volumes:    volumes,
		swarm:      monitor.GetSwarmSummary(),
		networkErr: formatFetchError(networkErr),
		volumeErr:  formatFetchError(volumeErr),
		timestamp:  time.Now().Format("15:04:05"),
	})
}

func containerStatusScore(c monitor.ContainerItem) int {
	switch {
	case c.State == "running" && c.Health == "unhealthy":
		return 1
	case c.State == "running" && c.Health == "starting":
		return 2
	case c.State == "running":
		return 3
	case c.State == "restarting":
		return 4
	case c.State == "paused":
		return 5
	default:
		return 6
	}
}

func (m Model) getFilteredContainers() []monitor.ContainerItem {
	var list []monitor.ContainerItem
	trimmedQuery := strings.TrimSpace(m.filterQuery)
	if trimmedQuery == "" {
		list = make([]monitor.ContainerItem, len(m.dockerInfo.Containers))
		copy(list, m.dockerInfo.Containers)
	} else {
		q := strings.ToLower(trimmedQuery)
		for _, c := range m.dockerInfo.Containers {
			if strings.Contains(strings.ToLower(c.Name), q) ||
				strings.Contains(strings.ToLower(c.Image), q) ||
				strings.Contains(strings.ToLower(c.ID), q) ||
				strings.Contains(strings.ToLower(c.ComposeProject), q) ||
				strings.Contains(strings.ToLower(c.Health), q) {
				list = append(list, c)
			}
		}
	}

	switch m.containerSortMode {
	case SortCPU:
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].CPUPerc != list[j].CPUPerc {
				return list[i].CPUPerc > list[j].CPUPerc
			}
			return list[i].Name < list[j].Name
		})
	case SortMem:
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].MemUsageMB != list[j].MemUsageMB {
				return list[i].MemUsageMB > list[j].MemUsageMB
			}
			return list[i].Name < list[j].Name
		})
	case SortName:
		sort.SliceStable(list, func(i, j int) bool {
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		})
	case SortStatus:
		sort.SliceStable(list, func(i, j int) bool {
			si, sj := containerStatusScore(list[i]), containerStatusScore(list[j])
			if si != sj {
				return si < sj
			}
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		})
	}

	return list
}

func (m *Model) clampCursor() {
	containers := m.getFilteredContainers()
	if len(containers) == 0 {
		m.cursorIdx = 0
	} else if m.cursorIdx >= len(containers) {
		m.cursorIdx = len(containers) - 1
	} else if m.cursorIdx < 0 {
		m.cursorIdx = 0
	}

	if len(m.compose) == 0 {
		m.composeCursorIdx = 0
	} else if m.composeCursorIdx >= len(m.compose) {
		m.composeCursorIdx = len(m.compose) - 1
	} else if m.composeCursorIdx < 0 {
		m.composeCursorIdx = 0
	}

	if len(m.swarm.Services) == 0 {
		m.swarmCursorIdx = 0
	} else if m.swarmCursorIdx >= len(m.swarm.Services) {
		m.swarmCursorIdx = len(m.swarm.Services) - 1
	} else if m.swarmCursorIdx < 0 {
		m.swarmCursorIdx = 0
	}

	if len(m.networks) == 0 {
		m.networkCursorIdx = 0
	} else if m.networkCursorIdx >= len(m.networks) {
		m.networkCursorIdx = len(m.networks) - 1
	} else if m.networkCursorIdx < 0 {
		m.networkCursorIdx = 0
	}

	if len(m.volumes) == 0 {
		m.volumeCursorIdx = 0
	} else if m.volumeCursorIdx >= len(m.volumes) {
		m.volumeCursorIdx = len(m.volumes) - 1
	} else if m.volumeCursorIdx < 0 {
		m.volumeCursorIdx = 0
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd(m.refreshInterval)
}

func (m *Model) showToast(msg string) {
	m.actionMessage = msg
	m.actionMessageExpiry = time.Now().Add(4 * time.Second)
}

// nextTab switches to the next tab cyclically.
func (m *Model) nextTab() {
	m.activeTab++
	if m.activeTab > TabCount {
		m.activeTab = TabDocker
	}
}

// prevTab switches to the previous tab cyclically.
func (m *Model) prevTab() {
	m.activeTab--
	if m.activeTab < TabDocker {
		m.activeTab = TabCount
	}
}

// setTab switches directly to the specified tab if within valid range.
func (m *Model) setTab(tab TabID) {
	if tab >= TabDocker && tab <= TabCount {
		m.activeTab = tab
	}
}

// moveCursor adjusts cursor index for the active tab list by delta with cyclic wrapping.
func (m *Model) moveCursor(delta int) {
	switch m.activeTab {
	case TabDocker:
		containers := m.getFilteredContainers()
		if len(containers) > 0 {
			m.cursorIdx = (m.cursorIdx + delta) % len(containers)
			if m.cursorIdx < 0 {
				m.cursorIdx += len(containers)
			}
		}
	case TabCompose:
		if len(m.compose) > 0 {
			m.composeCursorIdx = (m.composeCursorIdx + delta) % len(m.compose)
			if m.composeCursorIdx < 0 {
				m.composeCursorIdx += len(m.compose)
			}
		}
	case TabNetworks:
		if len(m.networks) > 0 {
			m.networkCursorIdx = (m.networkCursorIdx + delta) % len(m.networks)
			if m.networkCursorIdx < 0 {
				m.networkCursorIdx += len(m.networks)
			}
		}
	case TabVolumes:
		if len(m.volumes) > 0 {
			m.volumeCursorIdx = (m.volumeCursorIdx + delta) % len(m.volumes)
			if m.volumeCursorIdx < 0 {
				m.volumeCursorIdx += len(m.volumes)
			}
		}
	case TabSwarm:
		if len(m.swarm.Services) > 0 {
			m.swarmCursorIdx = (m.swarmCursorIdx + delta) % len(m.swarm.Services)
			if m.swarmCursorIdx < 0 {
				m.swarmCursorIdx += len(m.swarm.Services)
			}
		}
	}
}

func formatFetchError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// openModal activates the specified modal dialog.
func (m *Model) openModal(modal ModalType) {
	m.activeModal = modal
}

// closeModal dismisses the active modal dialog and cleans ephemeral input state.
func (m *Model) closeModal() {
	m.activeModal = ModalNone
	m.execIsCustomInput = false
	m.execCustomCmd = ""
}
