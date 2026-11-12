package tui

import (
	"strings"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/tui/modals"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
	"github.com/AndreyJVM/dockpulse/internal/tui/views"
)

// View renders the complete TUI interface for the current model state.
func (m Model) View() string {
	var b strings.Builder

	// 1. Header
	b.WriteString(views.RenderHeader(views.HeaderContext{
		Version:               m.version,
		LastUpdate:            m.lastUpdate,
		DockerInfo:            m.dockerInfo,
		FormattedDockerStatus: m.formatDockerStatus(m.dockerInfo.Status),
		I18n:                  &m.i18n,
	}) + "\n\n")

	// 2. Tabs
	tabs := []string{
		m.i18n.TabDocker,
		m.i18n.TabCompose,
		m.i18n.TabNetworks,
		m.i18n.TabVolumes,
		m.i18n.TabSwarm,
	}
	b.WriteString(views.RenderTabs(tabs, int(m.activeTab), m.width) + "\n")

	// Action message toast
	if m.actionMessage != "" && time.Now().Before(m.actionMessageExpiry) {
		b.WriteString(theme.YellowStyle.Render("[!] "+m.actionMessage) + "\n\n")
	}

	// 3. Tab contents or active modal
	switch m.activeModal {
	case ModalLogs:
		b.WriteString(m.renderLogsModal())
	case ModalInspect:
		b.WriteString(m.renderInspectModal())
	case ModalComposeDown:
		b.WriteString(m.renderComposeDownModal())
	case ModalRemoveContainer:
		b.WriteString(m.renderRemoveContainerModal())
	case ModalExec:
		b.WriteString(m.renderExecModal())
	case ModalPrune:
		b.WriteString(m.renderPruneModal())
	default:
		if m.showHelp {
			b.WriteString("\n" + m.renderHelpBox() + "\n")
		} else {
			switch m.activeTab {
			case TabDocker:
				b.WriteString(m.renderDockerTab())
			case TabCompose:
				b.WriteString(m.renderComposeTab())
			case TabNetworks:
				b.WriteString(m.renderNetworksTab())
			case TabVolumes:
				b.WriteString(m.renderVolumesTab())
			case TabSwarm:
				b.WriteString(m.renderSwarmTab())
			}
		}
	}

	// 4. Footer
	b.WriteString(views.RenderFooter(views.FooterContext{
		ActiveTab:    int(m.activeTab),
		ActiveModal:  string(m.activeModal),
		PruneConfirm: m.pruneConfirm,
		IsFiltering:  m.isFiltering,
		FilterQuery:  m.filterQuery,
		Width:        m.width,
		I18n:         &m.i18n,
	}))

	return b.String()
}

func (m Model) renderDockerTab() string {
	containers := m.getFilteredContainers()
	sortNames := []string{
		m.i18n.SortModeDefault,
		m.i18n.SortModeCPU,
		m.i18n.SortModeMem,
		m.i18n.SortModeName,
		m.i18n.SortModeStatus,
	}

	return views.RenderDocker(views.DockerContext{
		DockerInfo:           m.dockerInfo,
		Containers:           containers,
		TotalContainersCount: len(m.dockerInfo.Containers),
		CursorIdx:            m.cursorIdx,
		IsFiltering:          m.isFiltering,
		FilterQuery:          m.filterQuery,
		SortMode:             m.containerSortMode,
		SortNames:            sortNames,
		Width:                m.width,
		I18n:                 &m.i18n,
	})
}

func (m Model) renderComposeTab() string {
	return views.RenderCompose(views.ComposeContext{
		Projects:     m.compose,
		CursorIdx:    m.composeCursorIdx,
		FormatStatus: m.formatComposeStatus,
		I18n:         &m.i18n,
	})
}

func (m Model) renderNetworksTab() string {
	return views.RenderNetworks(views.NetworksContext{
		Networks:  m.networks,
		CursorIdx: m.networkCursorIdx,
		Error:     m.networkErr,
		I18n:      &m.i18n,
	})
}

func (m Model) renderVolumesTab() string {
	return views.RenderVolumes(views.VolumesContext{
		Volumes:   m.volumes,
		CursorIdx: m.volumeCursorIdx,
		Error:     m.volumeErr,
		I18n:      &m.i18n,
	})
}

func (m Model) renderSwarmTab() string {
	return views.RenderSwarm(views.SwarmContext{
		Swarm:        m.swarm,
		CursorIdx:    m.swarmCursorIdx,
		FormatStatus: m.formatSwarmStatus,
		I18n:         &m.i18n,
	})
}

func (m Model) renderLogsModal() string {
	return modals.RenderLogs(m.logsLines, m.logsScroll, m.logsTargetName, m.logsTargetID, m.width, &m.i18n)
}

func (m Model) renderInspectModal() string {
	return modals.RenderInspect(m.inspectTarget, &m.i18n)
}

func (m Model) renderPruneModal() string {
	return modals.RenderPrune(m.pruneUsage, m.pruneUsageErr, m.pruneConfirm, &m.i18n)
}

func (m Model) renderComposeDownModal() string {
	return modals.RenderComposeDown(m.composeDownTarget, &m.i18n)
}

func (m Model) renderRemoveContainerModal() string {
	return modals.RenderRemoveContainer(m.removeContainerTargetName, m.removeContainerTargetID, &m.i18n)
}

func (m Model) renderExecModal() string {
	return modals.RenderExec(modals.ExecModalContext{
		ContainerName: m.execTargetName,
		ContainerID:   m.execTargetID,
		CursorIdx:     m.execCursorIdx,
		IsCustomInput: m.execIsCustomInput,
		CustomCmd:     m.execCustomCmd,
		I18n:          &m.i18n,
	})
}

func (m Model) renderHelpBox() string {
	return modals.RenderHelp(m.refreshInterval, &m.i18n)
}
