package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// DockerContext contains data needed to render the Docker Containers tab.
type DockerContext struct {
	DockerInfo            monitor.DockerSummary
	Containers            []monitor.ContainerItem
	TotalContainersCount  int
	CursorIdx             int
	IsFiltering           bool
	FilterQuery           string
	SortMode              int
	SortNames             []string
	Width                 int
	FormattedDockerStatus string
	I18n                  *i18n.Dictionary
}

// RenderDocker renders the Docker Containers tab.
func RenderDocker(ctx DockerContext) string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render(ctx.I18n.DockerTabTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %s\n\n",
		theme.DimStyle.Render(ctx.I18n.ServiceLabel), ctx.FormattedDockerStatus,
		theme.DimStyle.Render(ctx.I18n.ImagesCountLabel), ctx.DockerInfo.ImagesCount,
		theme.DimStyle.Render(ctx.I18n.VersionLabel), ctx.DockerInfo.Version,
	))

	if len(ctx.DockerInfo.Containers) == 0 {
		if ctx.DockerInfo.Status == "running" {
			s.WriteString("  " + theme.DimStyle.Render(ctx.I18n.ContainersNotRunning) + "\n")
		} else {
			s.WriteString("  " + theme.YellowStyle.Render(ctx.I18n.DockerDaemonDown) + "\n")
		}
		return s.String()
	}

	var statusBadges []string
	if ctx.IsFiltering {
		s.WriteString("  " + theme.YellowStyle.Render(fmt.Sprintf(ctx.I18n.FilterPrompt, ctx.FilterQuery)) + "  " + theme.DimStyle.Render(ctx.I18n.FilterHint) + "\n\n")
	} else {
		if ctx.FilterQuery != "" {
			statusBadges = append(statusBadges, theme.ActiveTabStyle.Render(fmt.Sprintf(ctx.I18n.FilterActiveBadge, ctx.FilterQuery))+"  "+theme.DimStyle.Render(fmt.Sprintf(ctx.I18n.FilterResultsCount, len(ctx.Containers), ctx.TotalContainersCount)))
		}
		if ctx.SortMode != 0 && ctx.SortMode < len(ctx.SortNames) {
			statusBadges = append(statusBadges, theme.CyanStyle.Render(fmt.Sprintf(ctx.I18n.SortBadge, ctx.SortNames[ctx.SortMode])))
		}
		if len(statusBadges) > 0 {
			s.WriteString("  " + strings.Join(statusBadges, "   ") + "\n\n")
		}
	}

	if len(ctx.Containers) == 0 {
		s.WriteString("  " + theme.YellowStyle.Render(fmt.Sprintf(ctx.I18n.NoContainersFound, ctx.FilterQuery)) + "\n")
		s.WriteString("  " + theme.DimStyle.Render(ctx.I18n.PressEscToClear) + "\n")
		return s.String()
	}

	headerText := fmt.Sprintf("  %-8s  %-20s  %-22s  %-20s  %-7s  %-14s  %s",
		ctx.I18n.ColID,
		ctx.I18n.ColName,
		ctx.I18n.ColImage,
		ctx.I18n.ColStatus,
		ctx.I18n.ColCPU,
		ctx.I18n.ColMem,
		ctx.I18n.ColPorts,
	)
	s.WriteString(theme.DimStyle.Render(headerText) + "\n")
	s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", max(ctx.Width-4, 96))) + "\n")

	for i, c := range ctx.Containers {
		prefix := "  "
		if i == ctx.CursorIdx {
			prefix = theme.GreenStyle.Render("► ")
		}

		cID := components.PadRight(components.Truncate(c.ID, 8), 8)
		cName := components.PadRight(components.Truncate(c.Name, 20), 20)
		cImg := components.PadRight(components.Truncate(c.Image, 22), 22)

		statusText, statusStyle := formatContainerStatus(c.State, c.Health)
		statusPadded := components.PadRight(components.Truncate(statusText, 20), 20)

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
		cpuPadded := components.PadLeft(cpuRaw, 7)
		memPadded := components.PadRight(components.Truncate(memRaw, 14), 14)
		portsPreview := components.Truncate(c.Ports, max(20, ctx.Width-105))

		if i == ctx.CursorIdx {
			plainRow := fmt.Sprintf("%-8s  %-20s  %-22s  %-20s  %-7s  %-14s  %s",
				cID, cName, cImg, statusPadded, cpuPadded, memPadded, portsPreview,
			)
			s.WriteString(fmt.Sprintf("%s%s\n", prefix, theme.SelectedRowStyle.Render(plainRow)))
		} else {
			row := fmt.Sprintf("%-8s  %-20s  %-22s  %s  %s  %-14s  %s",
				cID,
				theme.BoldStyle.Render(cName),
				theme.DimStyle.Render(cImg),
				statusStyle.Render(statusPadded),
				styleCPU(c.CPUPerc, c.State, cpuPadded),
				theme.DimStyle.Render(memPadded),
				portsPreview,
			)
			s.WriteString(fmt.Sprintf("%s%s\n", prefix, row))
		}
	}

	// Bottom Details Panel for selected container
	if len(ctx.Containers) > 0 && ctx.CursorIdx < len(ctx.Containers) {
		sel := ctx.Containers[ctx.CursorIdx]
		s.WriteString("\n")
		cardWidth := max(ctx.Width-4, 96)
		title := fmt.Sprintf(" %s: %s (%s) ", ctx.I18n.SelectedContainerDetails, sel.Name, sel.ID)
		topDashLen := cardWidth - len([]rune(title)) - 4
		if topDashLen < 2 {
			topDashLen = 2
		}
		s.WriteString("  " + theme.CyanStyle.Render("┌─") + theme.BoldStyle.Render(title) + theme.CyanStyle.Render(strings.Repeat("─", topDashLen)+"┐") + "\n")

		// Ports
		portsDisplay := sel.Ports
		if portsDisplay == "" {
			portsDisplay = theme.DimStyle.Render(ctx.I18n.DetailNone)
		} else {
			portsDisplay = theme.CyanStyle.Render(portsDisplay)
		}
		s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
			theme.CyanStyle.Render("│"),
			theme.BoldStyle.Render(ctx.I18n.DetailPortsLabel),
			portsDisplay,
		))

		// Volumes / Mounts
		if len(sel.Mounts) == 0 {
			s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
				theme.CyanStyle.Render("│"),
				theme.BoldStyle.Render(ctx.I18n.DetailMountsLabel),
				theme.DimStyle.Render(ctx.I18n.DetailNone),
			))
		} else {
			for idx, mnt := range sel.Mounts {
				label := ""
				if idx == 0 {
					label = theme.BoldStyle.Render(fmt.Sprintf("%-18s", ctx.I18n.DetailMountsLabel))
				} else {
					label = strings.Repeat(" ", 18)
				}
				s.WriteString(fmt.Sprintf("  %s %s %s\n",
					theme.CyanStyle.Render("│"),
					label,
					theme.YellowStyle.Render("• ")+mnt,
				))
			}
		}

		// Network I/O
		netIODisplay := ""
		if sel.State == "running" {
			netIODisplay = fmt.Sprintf("%s / %s",
				theme.GreenStyle.Render(monitor.FormatBytes(int64(sel.NetRxBytes))),
				theme.CyanStyle.Render(monitor.FormatBytes(int64(sel.NetTxBytes))),
			)
		} else {
			netIODisplay = theme.DimStyle.Render("0 B / 0 B")
		}
		s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
			theme.CyanStyle.Render("│"),
			theme.BoldStyle.Render(ctx.I18n.DetailNetIOLabel),
			netIODisplay,
		))

		// Block I/O
		blockIODisplay := ""
		if sel.State == "running" {
			blockIODisplay = fmt.Sprintf("%s / %s",
				theme.GreenStyle.Render(monitor.FormatBytes(int64(sel.BlockReadBytes))),
				theme.YellowStyle.Render(monitor.FormatBytes(int64(sel.BlockWriteBytes))),
			)
		} else {
			blockIODisplay = theme.DimStyle.Render("0 B / 0 B")
		}
		s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
			theme.CyanStyle.Render("│"),
			theme.BoldStyle.Render(ctx.I18n.DetailBlockIOLabel),
			blockIODisplay,
		))

		// Compose Info
		if sel.ComposeProject != "" {
			s.WriteString(fmt.Sprintf("  %s %-18s %s\n",
				theme.CyanStyle.Render("│"),
				theme.BoldStyle.Render(ctx.I18n.DetailComposeLabel),
				fmt.Sprintf("%s / %s", theme.BoldStyle.Render(sel.ComposeProject), sel.ComposeService),
			))
		}

		s.WriteString("  " + theme.CyanStyle.Render("└"+strings.Repeat("─", cardWidth-2)+"┘") + "\n")
	}

	return s.String()
}

func formatContainerStatus(state, health string) (string, lipgloss.Style) {
	if state == "running" {
		switch health {
		case "healthy":
			return "● [healthy]", theme.GreenStyle
		case "unhealthy":
			return "▲ [unhealthy]", theme.RedStyle
		case "starting":
			return "⏳ [starting]", theme.YellowStyle
		default:
			return "● running", theme.GreenStyle
		}
	} else if state == "paused" {
		return "⏸ paused", theme.YellowStyle
	} else if state == "restarting" {
		return "↻ restarting", theme.YellowStyle
	}
	if state == "" {
		state = "stopped"
	}
	return "○ " + state, theme.DimStyle
}

func styleCPU(val float64, state, paddedStr string) string {
	if state != "running" {
		return theme.DimStyle.Render(paddedStr)
	}
	if val >= 80 {
		return theme.RedStyle.Render(paddedStr)
	} else if val >= 50 {
		return theme.YellowStyle.Render(paddedStr)
	}
	return theme.GreenStyle.Render(paddedStr)
}
