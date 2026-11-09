package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case execFinishedMsg:
		return m.handleExecFinished(msg)

	case actionFinishedMsg:
		if msg.err != nil {
			m.showToast(msg.err.Error())
		} else if msg.message != "" {
			m.showToast(msg.message)
		}
		if msg.refresh {
			return m, fetchDataCmd()
		}
		return m, nil

	case dataFetchedMsg:
		m.applyFetchedData(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Если активен ввод фильтра
		if m.isFiltering {
			return m.handleFilterInput(msg)
		}

		// Если активно модальное окно
		if m.activeModal != "" {
			return m.handleModalInput(msg)
		}

		// Если открыта справка
		if m.showHelp {
			return m.handleHelpInput(msg)
		}

		// Основной режим
		return m.handleMainKeys(msg)

	case tickMsg:
		return m, tea.Batch(
			tickCmd(m.refreshInterval),
			fetchDataCmd(),
		)
	}
	return m, nil
}

func (m Model) handleExecFinished(msg execFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		if exitErr, ok := msg.err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				m.showToast(fmt.Sprintf(m.i18n.ToastShellEnded, msg.containerName))
			} else {
				m.showToast(fmt.Sprintf(m.i18n.ToastShellEndedCode, msg.containerName, exitErr.ExitCode()))
			}
		} else {
			m.showToast(fmt.Sprintf(m.i18n.ToastShellError, msg.containerName, msg.err))
		}
	} else {
		m.showToast(fmt.Sprintf(m.i18n.ToastShellEnded, msg.containerName))
	}
	m.fetchAllData()
	return m, nil
}

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

func (m Model) handleModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.activeModal == "compose_down" {
		switch msg.String() {
		case "esc", "q", "n", "N":
			m.activeModal = ""
			m.composeDownTarget = ""
			return m, nil
		case "y", "Y":
			target := m.composeDownTarget
			m.activeModal = ""
			m.composeDownTarget = ""
			m.showToast(fmt.Sprintf(m.i18n.ToastDowningCompose, target))
			return m, downComposeCmd(target, m.dockerInfo.Containers, m.i18n)
		}
		return m, nil
	}

	if m.activeModal == "remove_container" {
		switch msg.String() {
		case "esc", "q", "n", "N":
			m.activeModal = ""
			m.removeContainerTargetID = ""
			m.removeContainerTargetName = ""
			return m, nil
		case "y", "Y":
			targetID := m.removeContainerTargetID
			targetName := m.removeContainerTargetName
			m.activeModal = ""
			m.removeContainerTargetID = ""
			m.removeContainerTargetName = ""
			m.showToast(fmt.Sprintf(m.i18n.ToastRemovingCont, targetName))
			return m, removeContainerCmd(targetID, targetName, m.i18n)
		}
		return m, nil
	}

	if m.activeModal == "exec" {
		return m.handleExecModalInput(msg)
	}

	if m.activeModal == "prune" {
		switch msg.String() {
		case "q":
			m.activeModal = ""
			m.pruneConfirm = 0
			return m, nil
		case "esc":
			if m.pruneConfirm > 0 {
				m.pruneConfirm = 0
			} else {
				m.activeModal = ""
			}
			return m, nil
		case "1", "2", "3", "4":
			if m.pruneConfirm == 0 {
				switch msg.String() {
				case "1":
					m.pruneConfirm = 1
				case "2":
					m.pruneConfirm = 2
				case "3":
					m.pruneConfirm = 3
				case "4":
					m.pruneConfirm = 4
				}
			}
			return m, nil
		case "y", "Y":
			if m.pruneConfirm > 0 {
				opt := m.pruneConfirm
				m.activeModal = ""
				m.pruneConfirm = 0
				m.showToast(m.i18n.ToastPruning)
				return m, pruneCmd(opt, m.i18n)
			}
			return m, nil
		case "n", "N":
			if m.pruneConfirm > 0 {
				m.pruneConfirm = 0
			}
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "esc":
		m.activeModal = ""
		return m, nil

	case "up", "k":
		if m.activeModal == "logs" && m.logsScroll > 0 {
			m.logsScroll--
		}
		return m, nil

	case "down", "j":
		if m.activeModal == "logs" && m.logsScroll < len(m.logsLines)-10 {
			m.logsScroll++
		}
		return m, nil

	case "pgup":
		if m.activeModal == "logs" {
			m.logsScroll = max(0, m.logsScroll-10)
		}
		return m, nil

	case "pgdown":
		if m.activeModal == "logs" {
			m.logsScroll = min(max(0, len(m.logsLines)-10), m.logsScroll+10)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleHelpInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "?":
		m.showHelp = false
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

func (m Model) launchExecProcess(opts monitor.ExecOptions) (tea.Model, tea.Cmd) {
	targetID := m.execTargetID
	targetName := m.execTargetName
	m.activeModal = ""
	m.execTargetID = ""
	m.execTargetName = ""
	m.execIsCustomInput = false
	m.execCustomCmd = ""

	cmd := monitor.GetContainerExecCommandWithOptions(targetID, opts)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return execFinishedMsg{
			containerName: targetName,
			err:           err,
		}
	})
}

func (m Model) handleExecModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.execIsCustomInput {
		switch msg.String() {
		case "esc":
			m.execIsCustomInput = false
			return m, nil
		case "enter":
			if strings.TrimSpace(m.execCustomCmd) == "" {
				m.execIsCustomInput = false
				return m, nil
			}
			return m.launchExecProcess(monitor.ExecOptions{
				Mode:      "custom",
				CustomCmd: m.execCustomCmd,
			})
		case "backspace":
			if len(m.execCustomCmd) > 0 {
				run := []rune(m.execCustomCmd)
				m.execCustomCmd = string(run[:len(run)-1])
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
				m.execCustomCmd += msg.String()
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "esc", "q":
		m.activeModal = ""
		m.execTargetID = ""
		m.execTargetName = ""
		return m, nil
	case "up", "k":
		if m.execCursorIdx > 0 {
			m.execCursorIdx--
		} else {
			m.execCursorIdx = 5
		}
		return m, nil
	case "down", "j":
		if m.execCursorIdx < 5 {
			m.execCursorIdx++
		} else {
			m.execCursorIdx = 0
		}
		return m, nil
	case "1", "b", "B":
		return m.launchExecProcess(monitor.ExecOptions{Mode: "auto"})
	case "2", "s", "S":
		return m.launchExecProcess(monitor.ExecOptions{Mode: "sh"})
	case "3", "z", "Z":
		return m.launchExecProcess(monitor.ExecOptions{Mode: "zsh"})
	case "4", "a", "A":
		return m.launchExecProcess(monitor.ExecOptions{Mode: "attach"})
	case "5", "r", "R":
		return m.launchExecProcess(monitor.ExecOptions{Mode: "root"})
	case "6", "c", "C":
		m.execIsCustomInput = true
		m.execCustomCmd = ""
		return m, nil
	case "enter":
		switch m.execCursorIdx {
		case 0:
			return m.launchExecProcess(monitor.ExecOptions{Mode: "auto"})
		case 1:
			return m.launchExecProcess(monitor.ExecOptions{Mode: "sh"})
		case 2:
			return m.launchExecProcess(monitor.ExecOptions{Mode: "zsh"})
		case 3:
			return m.launchExecProcess(monitor.ExecOptions{Mode: "attach"})
		case 4:
			return m.launchExecProcess(monitor.ExecOptions{Mode: "root"})
		case 5:
			m.execIsCustomInput = true
			m.execCustomCmd = ""
			return m, nil
		}
	}
	return m, nil
}

