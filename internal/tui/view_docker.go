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
		dimStyle.Render(m.i18n.ServiceLabel), m.formatDockerStatus(m.dockerInfo.Status),
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

	if len(containers) == 0 {
		s.WriteString("  " + yellowStyle.Render(fmt.Sprintf(m.i18n.NoContainersFound, m.filterQuery)) + "\n")
		s.WriteString("  " + dimStyle.Render(m.i18n.PressEscToClear) + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %-14s %-20s %-20s %-10s %-16s %-10s %-12s %s\n",
		dimStyle.Render(m.i18n.ColID), dimStyle.Render(m.i18n.ColName), dimStyle.Render(m.i18n.ColImage),
		dimStyle.Render(m.i18n.ColCPU), dimStyle.Render(m.i18n.ColMem), dimStyle.Render(m.i18n.ColState),
		dimStyle.Render(m.i18n.ColHealth), dimStyle.Render(m.i18n.ColPorts),
	))
	s.WriteString("  " + dimStyle.Render(strings.Repeat("─", max(m.width-4, 96))) + "\n")

	for i, c := range containers {
		prefix := "  "
		if i == m.cursorIdx {
			prefix = greenStyle.Render("► ")
		}

		cpuStr := "-"
		memStr := "-"
		if c.State == "running" {
			cpuStr = formatPercent(c.CPUPerc)
			if c.MemLimitMB > 0 {
				memStr = fmt.Sprintf("%.1fM (%.1f%%)", c.MemUsageMB, c.MemPerc)
			} else {
				memStr = fmt.Sprintf("%.1fM", c.MemUsageMB)
			}
		}

		healthBadge := dimStyle.Render("-")
		switch c.Health {
		case "healthy":
			healthBadge = greenStyle.Render("[healthy]")
		case "unhealthy":
			healthBadge = redStyle.Render("[unhealthy]")
		case "starting":
			healthBadge = yellowStyle.Render("[starting]")
		}

		stateStyle := dimStyle
		if c.State == "running" {
			stateStyle = greenStyle
		} else if c.State == "paused" {
			stateStyle = yellowStyle
		}

		row := fmt.Sprintf("%-14s %-20s %-20s %-10s %-16s %-10s %-12s %s",
			c.ID[:min(12, len(c.ID))],
			boldStyle.Render(truncate(c.Name, 18)),
			truncate(c.Image, 18),
			cpuStr,
			memStr,
			stateStyle.Render(c.State),
			healthBadge,
			c.Ports,
		)

		if i == m.cursorIdx {
			s.WriteString(fmt.Sprintf("%s%s\n", prefix, selectedRowStyle.Render(row)))
		} else {
			s.WriteString(fmt.Sprintf("%s%s\n", prefix, row))
		}
	}

	return s.String()
}
