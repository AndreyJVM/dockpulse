package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.isFiltering = false
		m.filterQuery = ""
		m.clampCursor()
		return m, nil
	case tea.KeyEnter:
		m.isFiltering = false
		return m, nil
	case tea.KeyBackspace:
		if len(m.filterQuery) > 0 {
			runes := []rune(m.filterQuery)
			m.filterQuery = string(runes[:len(runes)-1])
			m.clampCursor()
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		m.filterQuery += msg.String()
		m.clampCursor()
		return m, nil
	}
	return m, nil
}

func (m Model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?":
		m.showHelp = true
		return m, nil

	// Tab navigation.
	case "tab", "right":
		m.nextTab()
	case "shift+tab", "left":
		m.prevTab()

	case "1":
		m.setTab(TabHost)
	case "2":
		m.setTab(TabDocker)
	case "3":
		m.setTab(TabCompose)
	case "4":
		m.setTab(TabSwarm)

	case "r", "R":
		m.showToast(m.i18n.ToastRefreshing)
		return m, fetchDataCmd()

	// Container sorting on the Docker tab.
	case "o", "O":
		if m.activeTab == TabDocker && !m.isFiltering {
			m.containerSortMode = (m.containerSortMode + 1) % sortModeCount
			m.clampCursor()
			sortNames := []string{
				m.i18n.SortModeDefault,
				m.i18n.SortModeCPU,
				m.i18n.SortModeMem,
				m.i18n.SortModeName,
				m.i18n.SortModeStatus,
			}
			if m.containerSortMode >= 0 && m.containerSortMode < len(sortNames) {
				m.showToast(fmt.Sprintf(m.i18n.SortBadge, sortNames[m.containerSortMode]))
			}
			return m, nil
		}

	// Container search.
	case "/":
		if m.activeTab == TabDocker {
			m.isFiltering = true
			return m, nil
		}

	case "esc":
		if m.filterQuery != "" {
			m.filterQuery = ""
			m.clampCursor()
			return m, nil
		}

	// List navigation for Docker, Compose, and Swarm.
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)

	// Container, Compose project, or Swarm service logs.
	case "l", "L":
		if m.activeTab == TabDocker {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.logsTargetName = c.Name
				m.logsTargetID = c.ID
				logs, err := monitor.GetContainerLogs(c.ID, 120)
				if err != nil {
					m.showToast(m.i18n.ToastLogsError + err.Error())
				} else {
					m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
					m.logsScroll = max(0, len(m.logsLines)-18)
					m.openModal(ModalLogs)
				}
			}
		} else if m.activeTab == TabCompose && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.logsTargetName = "Compose: " + proj.Name
			m.logsTargetID = proj.Name
			logs, err := monitor.GetComposeProjectLogs(proj.Name, m.dockerInfo.Containers, 150)
			if err != nil {
				m.showToast(m.i18n.ToastLogsError + err.Error())
			} else {
				m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
				m.logsScroll = max(0, len(m.logsLines)-18)
				m.openModal(ModalLogs)
			}
		} else if m.activeTab == TabSwarm && len(m.swarm.Services) > 0 && m.swarmCursorIdx < len(m.swarm.Services) {
			svc := m.swarm.Services[m.swarmCursorIdx]
			m.logsTargetName = "Swarm: " + svc.Name
			m.logsTargetID = svc.ID
			logs, err := monitor.GetSwarmServiceLogs(svc.ID, 150)
			if err != nil {
				m.showToast(m.i18n.ToastLogsError + err.Error())
			} else {
				m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
				m.logsScroll = max(0, len(m.logsLines)-18)
				m.openModal(ModalLogs)
			}
		}

	// Container inspection.
	case "enter", "i", "I":
		containers := m.getFilteredContainers()
		if m.activeTab == TabDocker && len(containers) > 0 && m.cursorIdx < len(containers) {
			c := containers[m.cursorIdx]
			detail, err := monitor.GetContainerDetail(c.ID)
			if err != nil {
				m.showToast(m.i18n.ToastInspectError + err.Error())
			} else {
				m.inspectTarget = detail
				m.openModal(ModalInspect)
			}
		}

	// Start or stop a container or Compose project.
	case "s", "S":
		if m.activeTab == TabDocker {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				if c.State == "running" {
					m.showToast(fmt.Sprintf(m.i18n.ToastStoppingContainer, c.Name))
					return m, stopContainerCmd(c.ID, c.Name, m.i18n)
				}
				m.showToast(fmt.Sprintf(m.i18n.ToastStartingContainer, c.Name))
				return m, startContainerCmd(c.ID, c.Name, m.i18n)
			}
		} else if m.activeTab == TabCompose && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.showToast(fmt.Sprintf(m.i18n.ToastStoppingCompose, proj.Name))
			return m, stopComposeCmd(proj.Name, m.dockerInfo.Containers, m.i18n)
		}

	// Start a Compose project.
	case "u", "U":
		if m.activeTab == TabCompose && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.showToast(fmt.Sprintf(m.i18n.ToastStartingCompose, proj.Name))
			return m, startComposeCmd(proj.Name, m.dockerInfo.Containers, m.i18n)
		}

	// Down a Compose project or remove a container.
	case "d", "D":
		if m.activeTab == TabDocker && !m.isFiltering {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.removeContainerTargetID = c.ID
				m.removeContainerTargetName = c.Name
				m.openModal(ModalRemoveContainer)
				return m, nil
			}
		} else if m.activeTab == TabCompose && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.composeDownTarget = proj.Name
			m.openModal(ModalComposeDown)
			return m, nil
		}

	// Open an interactive container shell.
	case "e", "E":
		if m.activeTab == TabDocker {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				if c.State != "running" {
					m.showToast(fmt.Sprintf(m.i18n.ToastCannotConnect, c.Name))
					return m, nil
				}
				if !monitor.IsDockerCLIAvailable() {
					m.showToast(m.i18n.ToastDockerNotFound)
					return m, nil
				}
				m.openModal(ModalExec)
				m.execTargetID = c.ID
				m.execTargetName = c.Name
				m.execCursorIdx = 0
				m.execIsCustomInput = false
				m.execCustomCmd = ""
				return m, nil
			}
		}

	// Force-kill a container with SIGKILL.
	case "K":
		if m.activeTab == TabDocker && !m.isFiltering {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				if c.State == "running" {
					m.showToast(fmt.Sprintf(m.i18n.ToastKillingCont, c.Name))
					return m, killContainerCmd(c.ID, c.Name, m.i18n)
				}
				m.showToast(fmt.Sprintf(m.i18n.ToastCannotConnect, c.Name))
				return m, nil
			}
		}

	// Remove a container.
	case "x", "X":
		if m.activeTab == TabDocker && !m.isFiltering {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.removeContainerTargetID = c.ID
				m.removeContainerTargetName = c.Name
				m.openModal(ModalRemoveContainer)
				return m, nil
			}
		}

	// Restart a container or Compose project.
	case "p", "P":
		if m.activeTab == TabDocker {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.showToast(fmt.Sprintf(m.i18n.ToastRestartingCont, c.Name))
				return m, restartContainerCmd(c.ID, c.Name, m.i18n)
			}
		} else if m.activeTab == TabCompose && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.showToast(fmt.Sprintf(m.i18n.ToastRestartingCompose, proj.Name))
			return m, restartComposeCmd(proj.Name, m.dockerInfo.Containers, m.i18n)
		}

	// Prune unused Docker resources.
	case "c", "C":
		if (m.activeTab == TabDocker || m.activeTab == TabCompose) && !m.isFiltering {
			m.openModal(ModalPrune)
			m.pruneConfirm = 0
			usage, err := monitor.GetDockerDiskUsage()
			m.pruneUsage = usage
			if err != nil {
				m.pruneUsageErr = err.Error()
			} else {
				m.pruneUsageErr = ""
			}
			return m, nil
		}
	}
	return m, nil
}
