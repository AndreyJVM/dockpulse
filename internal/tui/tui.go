package tui

import (
	"strings"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type execFinishedMsg struct {
	containerName string
	err           error
}

func tickCmd(d time.Duration) tea.Cmd {
	if d <= 0 {
		d = 2 * time.Second
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model представляет состояние TUI.
type Model struct {
	version         string
	activeTab       int
	refreshInterval time.Duration
	showHelp        bool
	width           int
	height          int

	// Навигация и модальные окна
	cursorIdx           int
	composeCursorIdx    int
	swarmCursorIdx      int
	activeModal         string // "", "logs", "inspect", "compose_down", "prune"
	logsTargetName      string
	logsTargetID        string
	logsLines           []string
	logsScroll          int
	inspectTarget       monitor.ContainerDetail
	actionMessage       string
	actionMessageExpiry time.Time

	// Compose down модальное окно
	composeDownTarget string

	// Prune модальное окно
	pruneUsage    monitor.DockerDiskUsage
	pruneUsageErr string
	pruneConfirm  int // 0 - меню, 1 - containers, 2 - images, 3 - volumes, 4 - all

	// История метрик (Sparklines)
	cpuHistory []float64
	memHistory []float64

	// Поиск и фильтрация контейнеров
	isFiltering bool
	filterQuery string

	sysInfo    monitor.SystemInfo
	cpuInfo    monitor.CPUInfo
	memInfo    monitor.MemoryInfo
	disksInfo  []monitor.PartitionInfo
	netInfo    monitor.NetworkInfo
	topProcs   []monitor.ProcessInfo
	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	swarm      monitor.SwarmSummary
	lastUpdate string
}

// NewModel создает новую TUI модель с настраиваемым интервалом обновления и начальной вкладкой.
func NewModel(refreshInterval time.Duration, initialTab int, version string) Model {
	if initialTab < 1 || initialTab > 4 {
		initialTab = 1
	}
	if refreshInterval <= 0 {
		refreshInterval = 2 * time.Second
	}
	m := Model{
		version:         version,
		activeTab:       initialTab,
		refreshInterval: refreshInterval,
	}
	m.fetchAllData()
	return m
}

func (m *Model) applyFetchedData(d dataFetchedMsg) {
	m.sysInfo = d.sysInfo
	m.cpuInfo = d.cpuInfo
	m.memInfo = d.memInfo
	m.disksInfo = d.disksInfo
	m.netInfo = d.netInfo
	m.topProcs = d.topProcs
	m.dockerInfo = d.dockerInfo
	m.compose = d.compose
	m.swarm = d.swarm
	m.lastUpdate = d.timestamp
	m.recordMetricHistory(m.cpuInfo.Percent, m.memInfo.Percent)

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
	m.applyFetchedData(dataFetchedMsg{
		sysInfo:    monitor.GetSystemInfo(),
		cpuInfo:    monitor.GetCPUInfo(),
		memInfo:    monitor.GetMemoryInfo(),
		disksInfo:  monitor.GetDisksInfo(),
		netInfo:    monitor.GetNetworkInfo(),
		topProcs:   monitor.GetTopProcesses(5),
		dockerInfo: dockerInfo,
		compose:    monitor.GetComposeProjects(dockerInfo.Containers),
		swarm:      monitor.GetSwarmSummary(),
		timestamp:  time.Now().Format("15:04:05"),
	})
}

func (m Model) getFilteredContainers() []monitor.ContainerItem {
	if strings.TrimSpace(m.filterQuery) == "" {
		return m.dockerInfo.Containers
	}
	q := strings.ToLower(strings.TrimSpace(m.filterQuery))
	var res []monitor.ContainerItem
	for _, c := range m.dockerInfo.Containers {
		if strings.Contains(strings.ToLower(c.Name), q) ||
			strings.Contains(strings.ToLower(c.Image), q) ||
			strings.Contains(strings.ToLower(c.ID), q) ||
			strings.Contains(strings.ToLower(c.ComposeProject), q) ||
			strings.Contains(strings.ToLower(c.Health), q) {
			res = append(res, c)
		}
	}
	return res
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
}

func (m Model) Init() tea.Cmd {
	return tickCmd(m.refreshInterval)
}

func (m *Model) showToast(msg string) {
	m.actionMessage = msg
	m.actionMessageExpiry = time.Now().Add(4 * time.Second)
}

func (m *Model) recordMetricHistory(cpu, mem float64) {
	const maxHistory = 30
	m.cpuHistory = append(m.cpuHistory, cpu)
	if len(m.cpuHistory) > maxHistory {
		m.cpuHistory = m.cpuHistory[len(m.cpuHistory)-maxHistory:]
	}
	m.memHistory = append(m.memHistory, mem)
	if len(m.memHistory) > maxHistory {
		m.memHistory = m.memHistory[len(m.memHistory)-maxHistory:]
	}
}
