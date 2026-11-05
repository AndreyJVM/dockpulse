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
	b.WriteString(views.RenderHeader(m.version, m.lastUpdate) + "\n\n")

	// 2. Tabs
	tabs := []string{
		m.i18n.TabHost,
		m.i18n.TabDocker,
		m.i18n.TabCompose,
		m.i18n.TabSwarm,
	}
	b.WriteString(views.RenderTabs(tabs, m.activeTab, m.width) + "\n")

	// Action message toast
	if m.actionMessage != "" && time.Now().Before(m.actionMessageExpiry) {
		b.WriteString(theme.YellowStyle.Render("[!] "+m.actionMessage) + "\n\n")
	}

	// 3. Tab contents or active modal
	if m.activeModal == "logs" {
		b.WriteString(m.renderLogsModal())
	} else if m.activeModal == "inspect" {
		b.WriteString(m.renderInspectModal())
	} else if m.activeModal == "compose_down" {
		b.WriteString(m.renderComposeDownModal())
	} else if m.activeModal == "prune" {
		b.WriteString(m.renderPruneModal())
	} else if m.showHelp {
		b.WriteString("\n" + m.renderHelpBox() + "\n")
	} else {
		switch m.activeTab {
		case 1:
			b.WriteString(m.renderHostTab())
		case 2:
			b.WriteString(m.renderDockerTab())
		case 3:
			b.WriteString(m.renderComposeTab())
		case 4:
			b.WriteString(m.renderSwarmTab())
		}
	}

	// 4. Footer
	b.WriteString(views.RenderFooter(views.FooterContext{
		ActiveTab:    m.activeTab,
		ActiveModal:  m.activeModal,
		PruneConfirm: m.pruneConfirm,
		IsFiltering:  m.isFiltering,
		FilterQuery:  m.filterQuery,
		Width:        m.width,
		I18n:         &m.i18n,
	}))

	return b.String()
}

func (m Model) renderHostTab() string {
	return views.RenderHost(views.HostContext{
		SysInfo:               m.sysInfo,
		CPUInfo:               m.cpuInfo,
		MemInfo:               m.memInfo,
		DockerInfo:            m.dockerInfo,
		CPUHistory:            m.cpuHistory,
		MemHistory:            m.memHistory,
		FormattedUptime:       m.formatUptime(m.sysInfo.UptimeSec),
		FormattedDockerStatus: m.formatDockerStatus(m.dockerInfo.Status),
		I18n:                  &m.i18n,
	})
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
		DockerInfo:            m.dockerInfo,
		Containers:            containers,
		TotalContainersCount:  len(m.dockerInfo.Containers),
		CursorIdx:             m.cursorIdx,
		IsFiltering:           m.isFiltering,
		FilterQuery:           m.filterQuery,
		SortMode:              m.containerSortMode,
		SortNames:             sortNames,
		Width:                 m.width,
		FormattedDockerStatus: m.formatDockerStatus(m.dockerInfo.Status),
		I18n:                  &m.i18n,
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

func (m Model) renderHelpBox() string {
	return modals.RenderHelp(m.refreshInterval, &m.i18n)
}
