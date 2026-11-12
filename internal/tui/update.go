package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case embeddedShellOutputMsg:
		return m.handleEmbeddedShellOutput(msg)

	case embeddedShellExitedMsg:
		return m.handleEmbeddedShellExited(msg)

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
		if m.embeddedShell != nil {
			width, height := embeddedShellSize(m.width, m.height)
			if err := m.embeddedShell.Resize(width, height); err != nil {
				m.showToast(fmt.Sprintf(m.i18n.ToastShellError, m.shellContainerName, err))
				m.closeEmbeddedShell()
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.embeddedShell != nil {
			if m.shellFocused {
				return m.handleEmbeddedShellKey(msg)
			}
			switch msg.Type {
			case tea.KeyCtrlX:
				m.closeEmbeddedShell()
				return m, nil
			}
			if msg.String() == "e" || msg.String() == "E" {
				m.shellFocused = true
				return m, nil
			}
		}

		// Active modals intercept keyboard input.
		if m.activeModal != ModalNone {
			return m.handleModalInput(msg)
		}

		// Container search and filter mode.
		if m.isFiltering && m.activeTab == TabDocker {
			return m.handleFilterInput(msg)
		}

		// Help modal.
		if m.showHelp {
			return m.handleHelpInput(msg)
		}

		// Main input mode.
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
