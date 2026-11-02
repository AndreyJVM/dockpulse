package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func (m Model) renderDockerTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(m.i18n.DockerTabTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %s\n\n",
		dimStyle.Render(m.i18n.ServiceLabel), m.dockerInfo.StatusDisplay,
		dimStyle.Render(m.i18n.ImagesCountLabel), m.dockerInfo.ImagesCount,
		dimStyle.Render(m.i18n.VersionLabel), m.dockerInfo.Version,
	))

	if len(m.dockerInfo.Containers) == 0 {
		if m.dockerInfo.Status == "running" {
			s.WriteString("  " + dimStyle.Render(m.i18n.ContainersNotRunning) + "\n")
		} else {
			s.WriteString("  " + yellowStyle.Render(m.i18n.DockerDaemonDown) + "\n")
		}
		return s.String()
	}

	containers := m.getFilteredContainers()

	if m.isFiltering {
		s.WriteString("  " + yellowStyle.Render(fmt.Sprintf(m.i18n.FilterPrompt, m.filterQuery)) + "  " + dimStyle.Render(m.i18n.FilterHint) + "\n\n")
	} else if m.filterQuery != "" {
		s.WriteString("  " + activeTabStyle.Render(fmt.Sprintf(m.i18n.FilterActiveBadge, m.filterQuery)) + "  " + dimStyle.Render(fmt.Sprintf(m.i18n.FilterResultsCount, len(containers), len(m.dockerInfo.Containers))) + "\n\n")
	}

	if len(containers) == 0 && m.filterQuery != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf(m.i18n.NoContainersFound, m.filterQuery)) + "\n")
		s.WriteString("  " + dimStyle.Render(m.i18n.PressEscToClear) + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("   %-12s %-18s %-18s %-8s %-15s %-9s %-11s %-18s\n",
		dimStyle.Render(m.i18n.ColID), dimStyle.Render(m.i18n.ColName), dimStyle.Render(m.i18n.ColImage),
		dimStyle.Render(m.i18n.ColCPU), dimStyle.Render(m.i18n.ColMem),
		dimStyle.Render(m.i18n.ColState), dimStyle.Render(m.i18n.ColHealth), dimStyle.Render(m.i18n.ColPorts),
	))
	s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 114)) + "\n")

	for i, c := range containers {
		s.WriteString(m.renderContainerRow(c, i == m.cursorIdx))
	}

	return s.String()
}

func (m Model) renderContainerRow(c monitor.ContainerItem, isSelected bool) string {
	stateColored := formatContainerState(c.State)
	healthText, healthColored := formatContainerHealth(c.Health)
	cpuStr, cpuColored := formatContainerCPU(c.CPUPerc, c.State)
	memStr, memColored := formatContainerMem(c.MemUsageMB, c.MemPerc, c.State)

	prefix := "  "
	rowContent := fmt.Sprintf("%-12s %-18s %-18s %s %s %s %s %s",
		c.ID, boldStyle.Render(truncate(c.Name, 16)), truncate(c.Image, 16),
		cpuColored, memColored,
		stateColored, healthColored, truncate(c.Ports, 18),
	)

	if isSelected {
		prefix = greenStyle.Render("► ")
		rowContent = selectedRowStyle.Render(fmt.Sprintf("%-12s %-18s %-18s %-8s %-15s %-9s %-11s %-18s",
			c.ID, truncate(c.Name, 16), truncate(c.Image, 16),
			cpuStr, truncate(memStr, 15),
			c.State, healthText, truncate(c.Ports, 18),
		))
	}

	return fmt.Sprintf("%s%s\n", prefix, rowContent)
}

func formatContainerState(state string) string {
	formatted := fmt.Sprintf("%-9s", state)
	switch state {
	case "running":
		return greenStyle.Render(formatted)
	case "paused":
		return yellowStyle.Render(formatted)
	default:
		return dimStyle.Render(formatted)
	}
}

func formatContainerHealth(health string) (string, string) {
	healthText := "-"
	switch health {
	case "healthy":
		healthText = "[healthy]"
	case "unhealthy":
		healthText = "[unhealthy]"
	case "starting":
		healthText = "[starting]"
	}

	formatted := fmt.Sprintf("%-11s", healthText)
	switch health {
	case "healthy":
		return healthText, greenStyle.Render(formatted)
	case "unhealthy":
		return healthText, redStyle.Render(formatted)
	case "starting":
		return healthText, yellowStyle.Render(formatted)
	default:
		return healthText, dimStyle.Render(formatted)
	}
}

func formatContainerCPU(cpuPerc float64, state string) (string, string) {
	if state != "running" {
		cpuStr := "-"
		return cpuStr, dimStyle.Render(fmt.Sprintf("%-8s", cpuStr))
	}

	cpuStr := fmt.Sprintf("%.1f%%", cpuPerc)
	formatted := fmt.Sprintf("%-8s", cpuStr)
	if cpuPerc > 80.0 {
		return cpuStr, redStyle.Render(formatted)
	} else if cpuPerc > 50.0 {
		return cpuStr, yellowStyle.Render(formatted)
	} else if cpuPerc > 0.0 {
		return cpuStr, formatted
	}
	return cpuStr, dimStyle.Render(formatted)
}

func formatContainerMem(memUsageMB, memPerc float64, state string) (string, string) {
	if state != "running" || memUsageMB <= 0 {
		memStr := "-"
		return memStr, dimStyle.Render(fmt.Sprintf("%-15s", memStr))
	}

	var memStr string
	if memPerc > 0 {
		memStr = fmt.Sprintf("%.1fM (%.1f%%)", memUsageMB, memPerc)
	} else {
		memStr = fmt.Sprintf("%.1f MB", memUsageMB)
	}

	formatted := fmt.Sprintf("%-15s", truncate(memStr, 15))
	if memPerc > 85.0 {
		return memStr, redStyle.Render(formatted)
	} else if memPerc > 70.0 {
		return memStr, yellowStyle.Render(formatted)
	}
	return memStr, formatted
}
