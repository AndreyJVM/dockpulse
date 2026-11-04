package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	var statusBadges []string
	if m.isFiltering {
		s.WriteString("  " + yellowStyle.Render(fmt.Sprintf(m.i18n.FilterPrompt, m.filterQuery)) + "  " + dimStyle.Render(m.i18n.FilterHint) + "\n\n")
	} else {
		if m.filterQuery != "" {
			statusBadges = append(statusBadges, activeTabStyle.Render(fmt.Sprintf(m.i18n.FilterActiveBadge, m.filterQuery))+"  "+dimStyle.Render(fmt.Sprintf(m.i18n.FilterResultsCount, len(containers), len(m.dockerInfo.Containers))))
		}
		if m.containerSortMode != SortDefault {
			sortNames := []string{
				m.i18n.SortModeDefault,
				m.i18n.SortModeCPU,
				m.i18n.SortModeMem,
				m.i18n.SortModeName,
				m.i18n.SortModeStatus,
			}
			if m.containerSortMode < len(sortNames) {
				statusBadges = append(statusBadges, cyanStyle.Render(fmt.Sprintf(m.i18n.SortBadge, sortNames[m.containerSortMode])))
			}
		}
		if len(statusBadges) > 0 {
			s.WriteString("  " + strings.Join(statusBadges, "   ") + "\n\n")
		}
	}

	if len(containers) == 0 {
		s.WriteString("  " + yellowStyle.Render(fmt.Sprintf(m.i18n.NoContainersFound, m.filterQuery)) + "\n")
		s.WriteString("  " + dimStyle.Render(m.i18n.PressEscToClear) + "\n")
		return s.String()
	}

	headerText := fmt.Sprintf("  %-8s  %-20s  %-22s  %-20s  %-7s  %-14s  %s",
		m.i18n.ColID,
		m.i18n.ColName,
		m.i18n.ColImage,
		m.i18n.ColStatus,
		m.i18n.ColCPU,
		m.i18n.ColMem,
		m.i18n.ColPorts,
	)
	s.WriteString(dimStyle.Render(headerText) + "\n")
	s.WriteString("  " + dimStyle.Render(strings.Repeat("─", max(m.width-4, 96))) + "\n")

	for i, c := range containers {
		prefix := "  "
		if i == m.cursorIdx {
			prefix = greenStyle.Render("► ")
		}

		cID := padRight(truncate(c.ID, 8), 8)
		cName := padRight(truncate(c.Name, 20), 20)
		cImg := padRight(truncate(c.Image, 22), 22)

		statusText, statusStyle := formatContainerStatus(c.State, c.Health)
		statusPadded := padRight(truncate(statusText, 20), 20)

		cpuRaw := "-"
		memRaw := "-"
		if c.State == "running" {
			cpuRaw = fmt.Sprintf("%.1f%%", c.CPUPerc)
			if c.MemLimitMB > 0 {
				memRaw = fmt.Sprintf("%.1fM (%.1f%%)", c.MemUsageMB, c.MemPerc)
			} else {
				memRaw = fmt.Sprintf("%.1fM", c.MemUsageMB)
			}
		}
		cpuPadded := padLeft(cpuRaw, 7)
		memPadded := padRight(truncate(memRaw, 14), 14)
		portsPreview := truncate(c.Ports, max(20, m.width-105))

		if i == m.cursorIdx {
			plainRow := fmt.Sprintf("%-8s  %-20s  %-22s  %-20s  %-7s  %-14s  %s",
				cID, cName, cImg, statusPadded, cpuPadded, memPadded, portsPreview,
			)
			s.WriteString(fmt.Sprintf("%s%s\n", prefix, selectedRowStyle.Render(plainRow)))
		} else {
			row := fmt.Sprintf("%-8s  %-20s  %-22s  %s  %s  %-14s  %s",
				cID,
				boldStyle.Render(cName),
				dimStyle.Render(cImg),
				statusStyle.Render(statusPadded),
				styleCPU(c.CPUPerc, c.State, cpuPadded),
				dimStyle.Render(memPadded),
				portsPreview,
			)
			s.WriteString(fmt.Sprintf("%s%s\n", prefix, row))
		}
	}

	// Bottom Details Panel for selected container
	if len(containers) > 0 && m.cursorIdx < len(containers) {
		sel := containers[m.cursorIdx]
		s.WriteString("\n")
		cardWidth := max(m.width-4, 96)
		title := fmt.Sprintf(" %s: %s (%s) ", m.i18n.SelectedContainerDetails, sel.Name, sel.ID)
		topDashLen := cardWidth - len([]rune(title)) - 4
		if topDashLen < 2 {
			topDashLen = 2
		}
		s.WriteString("  " + cyanStyle.Render("┌─") + boldStyle.Render(title) + cyanStyle.Render(strings.Repeat("─", topDashLen)+"┐") + "\n")

		// Ports
		portsDisplay := sel.Ports
		if portsDisplay == "" {
			portsDisplay = dimStyle.Render(m.i18n.DetailNone)
		} else {
			portsDisplay = cyanStyle.Render(portsDisplay)
		}
		s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
			cyanStyle.Render("│"),
			boldStyle.Render(m.i18n.DetailPortsLabel),
			portsDisplay,
		))

		// Volumes / Mounts
		if len(sel.Mounts) == 0 {
			s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
				cyanStyle.Render("│"),
				boldStyle.Render(m.i18n.DetailMountsLabel),
				dimStyle.Render(m.i18n.DetailNone),
			))
		} else {
			for idx, mnt := range sel.Mounts {
				label := ""
				if idx == 0 {
					label = boldStyle.Render(fmt.Sprintf("%-18s", m.i18n.DetailMountsLabel))
				} else {
					label = strings.Repeat(" ", 18)
				}
				s.WriteString(fmt.Sprintf("  %s %s %s\n",
					cyanStyle.Render("│"),
					label,
					yellowStyle.Render("• ")+mnt,
				))
			}
		}

		// Compose Info
		if sel.ComposeProject != "" {
			s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
				cyanStyle.Render("│"),
				boldStyle.Render(m.i18n.DetailComposeLabel),
				fmt.Sprintf("%s / %s", boldStyle.Render(sel.ComposeProject), sel.ComposeService),
			))
		}

		s.WriteString("  " + cyanStyle.Render("└"+strings.Repeat("─", cardWidth-2)+"┘") + "\n")
	}

	return s.String()
}

func formatContainerStatus(state, health string) (string, lipgloss.Style) {
	if state == "running" {
		switch health {
		case "healthy":
			return "● [healthy]", greenStyle
		case "unhealthy":
			return "▲ [unhealthy]", redStyle
		case "starting":
			return "⏳ [starting]", yellowStyle
		default:
			return "● running", greenStyle
		}
	} else if state == "paused" {
		return "⏸ paused", yellowStyle
	} else if state == "restarting" {
		return "↻ restarting", yellowStyle
	}
	if state == "" {
		state = "stopped"
	}
	return "○ " + state, dimStyle
}

func styleCPU(val float64, state, paddedStr string) string {
	if state != "running" {
		return dimStyle.Render(paddedStr)
	}
	if val >= 80 {
		return redStyle.Render(paddedStr)
	} else if val >= 50 {
		return yellowStyle.Render(paddedStr)
	}
	return greenStyle.Render(paddedStr)
}
