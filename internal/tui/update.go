package tui

import (
	"fmt"
	"os/exec"

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
		// Если активно модальное окно — клавиши перехватываются им
		if m.activeModal != ModalNone {
			return m.handleModalInput(msg)
		}

		// Режим поиска / фильтрации контейнеров
		if m.isFiltering && m.activeTab == TabDocker {
			return m.handleFilterInput(msg)
		}

		// Модальное окно справки
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
