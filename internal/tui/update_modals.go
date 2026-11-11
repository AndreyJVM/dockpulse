package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleModalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.activeModal == ModalComposeDown {
		switch msg.String() {
		case "esc", "q", "n", "N":
			m.closeModal()
			m.composeDownTarget = ""
			return m, nil
		case "y", "Y":
			target := m.composeDownTarget
			m.closeModal()
			m.composeDownTarget = ""
			m.showToast(fmt.Sprintf(m.i18n.ToastDowningCompose, target))
			return m, downComposeCmd(target, m.dockerInfo.Containers, m.i18n)
		}
		return m, nil
	}

	if m.activeModal == ModalRemoveContainer {
		switch msg.String() {
		case "esc", "q", "n", "N":
			m.closeModal()
			m.removeContainerTargetID = ""
			m.removeContainerTargetName = ""
			return m, nil
		case "y", "Y":
			targetID := m.removeContainerTargetID
			targetName := m.removeContainerTargetName
			m.closeModal()
			m.removeContainerTargetID = ""
			m.removeContainerTargetName = ""
			m.showToast(fmt.Sprintf(m.i18n.ToastRemovingCont, targetName))
			return m, removeContainerCmd(targetID, targetName, m.i18n)
		}
		return m, nil
	}

	if m.activeModal == ModalExec {
		return m.handleExecModalInput(msg)
	}

	if m.activeModal == ModalPrune {
		switch msg.String() {
		case "q":
			m.closeModal()
			m.pruneConfirm = 0
			return m, nil
		case "esc":
			if m.pruneConfirm > 0 {
				m.pruneConfirm = 0
			} else {
				m.closeModal()
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
				m.closeModal()
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
		m.closeModal()
		return m, nil

	case "up", "k":
		if m.activeModal == ModalLogs && m.logsScroll > 0 {
			m.logsScroll--
		}
		return m, nil

	case "down", "j":
		if m.activeModal == ModalLogs && m.logsScroll < len(m.logsLines)-10 {
			m.logsScroll++
		}
		return m, nil

	case "pgup":
		if m.activeModal == ModalLogs {
			m.logsScroll = max(0, m.logsScroll-10)
		}
		return m, nil

	case "pgdown":
		if m.activeModal == ModalLogs {
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

func (m Model) launchExecProcess(opts monitor.ExecOptions) (tea.Model, tea.Cmd) {
	targetID := m.execTargetID
	targetName := m.execTargetName
	m.closeModal()
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
		m.closeModal()
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

