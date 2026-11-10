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

	// Навигация по вкладкам
	case "tab", "right":
		m.activeTab++
		if m.activeTab > 4 {
			m.activeTab = 1
		}
	case "shift+tab", "left":
		m.activeTab--
		if m.activeTab < 1 {
			m.activeTab = 4
		}

	case "1":
		m.activeTab = 1
	case "2":
		m.activeTab = 2
	case "3":
		m.activeTab = 3
	case "4":
		m.activeTab = 4

	case "r", "R":
		m.showToast(m.i18n.ToastRefreshing)
		return m, fetchDataCmd()

	// Сортировка списка контейнеров (Вкладка 2)
	case "o", "O":
		if m.activeTab == 2 && !m.isFiltering {
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

	// Поиск по списку контейнеров
	case "/":
		if m.activeTab == 2 {
			m.isFiltering = true
			return m, nil
		}

	case "esc":
		if m.filterQuery != "" {
			m.filterQuery = ""
			m.clampCursor()
			return m, nil
		}

	// Навигация в списке контейнеров (Вкладка 2), Compose (Вкладка 3), Swarm (Вкладка 4)
	case "up", "k":
		if m.activeTab == 2 {
			containers := m.getFilteredContainers()
			if len(containers) > 0 {
				m.cursorIdx--
				if m.cursorIdx < 0 {
					m.cursorIdx = len(containers) - 1
				}
			}
		} else if m.activeTab == 3 && len(m.compose) > 0 {
			m.composeCursorIdx--
			if m.composeCursorIdx < 0 {
				m.composeCursorIdx = len(m.compose) - 1
			}
		} else if m.activeTab == 4 && len(m.swarm.Services) > 0 {
			m.swarmCursorIdx--
			if m.swarmCursorIdx < 0 {
				m.swarmCursorIdx = len(m.swarm.Services) - 1
			}
		}
	case "down", "j":
		if m.activeTab == 2 {
			containers := m.getFilteredContainers()
			if len(containers) > 0 {
				m.cursorIdx++
				if m.cursorIdx >= len(containers) {
					m.cursorIdx = 0
				}
			}
		} else if m.activeTab == 3 && len(m.compose) > 0 {
			m.composeCursorIdx++
			if m.composeCursorIdx >= len(m.compose) {
				m.composeCursorIdx = 0
			}
		} else if m.activeTab == 4 && len(m.swarm.Services) > 0 {
			m.swarmCursorIdx++
			if m.swarmCursorIdx >= len(m.swarm.Services) {
				m.swarmCursorIdx = 0
			}
		}

	// Просмотр логов контейнера, проекта Compose или службы Swarm
	case "l", "L":
		if m.activeTab == 2 {
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
					m.activeModal = "logs"
				}
			}
		} else if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.logsTargetName = "Compose: " + proj.Name
			m.logsTargetID = proj.Name
			logs, err := monitor.GetComposeProjectLogs(proj.Name, m.dockerInfo.Containers, 150)
			if err != nil {
				m.showToast(m.i18n.ToastLogsError + err.Error())
			} else {
				m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
				m.logsScroll = max(0, len(m.logsLines)-18)
				m.activeModal = "logs"
			}
		} else if m.activeTab == 4 && len(m.swarm.Services) > 0 && m.swarmCursorIdx < len(m.swarm.Services) {
			svc := m.swarm.Services[m.swarmCursorIdx]
			m.logsTargetName = "Swarm: " + svc.Name
			m.logsTargetID = svc.ID
			logs, err := monitor.GetSwarmServiceLogs(svc.ID, 150)
			if err != nil {
				m.showToast(m.i18n.ToastLogsError + err.Error())
			} else {
				m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
				m.logsScroll = max(0, len(m.logsLines)-18)
				m.activeModal = "logs"
			}
		}

	// Детали контейнера (Inspect)
	case "enter", "i", "I":
		containers := m.getFilteredContainers()
		if m.activeTab == 2 && len(containers) > 0 && m.cursorIdx < len(containers) {
			c := containers[m.cursorIdx]
			detail, err := monitor.GetContainerDetail(c.ID)
			if err != nil {
				m.showToast(m.i18n.ToastInspectError + err.Error())
			} else {
				m.inspectTarget = detail
				m.activeModal = "inspect"
			}
		}

	// Старт / Стоп контейнера или проекта Compose
	case "s", "S":
		if m.activeTab == 2 {
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
		} else if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.showToast(fmt.Sprintf(m.i18n.ToastStoppingCompose, proj.Name))
			return m, stopComposeCmd(proj.Name, m.dockerInfo.Containers, m.i18n)
		}

	// Запуск проекта Compose
	case "u", "U":
		if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.showToast(fmt.Sprintf(m.i18n.ToastStartingCompose, proj.Name))
			return m, startComposeCmd(proj.Name, m.dockerInfo.Containers, m.i18n)
		}

	// Сворачивание проекта Compose (Down) или Удаление контейнера (Delete)
	case "d", "D":
		if m.activeTab == 2 && !m.isFiltering {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.removeContainerTargetID = c.ID
				m.removeContainerTargetName = c.Name
				m.activeModal = "remove_container"
				return m, nil
			}
		} else if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.composeDownTarget = proj.Name
			m.activeModal = "compose_down"
			return m, nil
		}

	// Подключение к Shell контейнера (Exec)
	case "e", "E":
		if m.activeTab == 2 {
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
				m.activeModal = "exec"
				m.execTargetID = c.ID
				m.execTargetName = c.Name
				m.execCursorIdx = 0
				m.execIsCustomInput = false
				m.execCustomCmd = ""
				return m, nil
			}
		}

	// Принудительное завершение контейнера (Kill / SIGKILL)
	case "K":
		if m.activeTab == 2 && !m.isFiltering {
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

	// Удаление контейнера (Remove)
	case "x", "X":
		if m.activeTab == 2 && !m.isFiltering {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.removeContainerTargetID = c.ID
				m.removeContainerTargetName = c.Name
				m.activeModal = "remove_container"
				return m, nil
			}
		}

	// Перезапуск контейнера или проекта Compose
	case "p", "P":
		if m.activeTab == 2 {
			containers := m.getFilteredContainers()
			if len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				m.showToast(fmt.Sprintf(m.i18n.ToastRestartingCont, c.Name))
				return m, restartContainerCmd(c.ID, c.Name, m.i18n)
			}
		} else if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
			proj := m.compose[m.composeCursorIdx]
			m.showToast(fmt.Sprintf(m.i18n.ToastRestartingCompose, proj.Name))
			return m, restartComposeCmd(proj.Name, m.dockerInfo.Containers, m.i18n)
		}

	// Очистка неиспользуемых ресурсов Docker (Prune)
	case "c", "C":
		if (m.activeTab == 2 || m.activeTab == 3) && !m.isFiltering {
			m.activeModal = "prune"
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

