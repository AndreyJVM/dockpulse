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

const (
	colWidthID     = 8
	colWidthName   = 20
	colWidthImage  = 22
	colWidthStatus = 20
	colWidthCPU    = 7
	colWidthMem    = 14
	minPortsWidth  = 20
	minCardWidth   = 96
	detailColWidth = 18
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
	dict := getDictionary(ctx.I18n)

	renderDockerHeader(&ctx, dict, &s)

	if len(ctx.DockerInfo.Containers) == 0 {
		renderEmptyContainersNotice(&ctx, dict, &s)
		return s.String()
	}

	renderFilterAndSortBadges(&ctx, dict, &s)

	if len(ctx.Containers) == 0 {
		renderNoContainersFoundNotice(&ctx, dict, &s)
		return s.String()
	}

	renderContainerTable(&ctx, dict, &s)
	renderContainerDetailsCard(&ctx, dict, &s)

	return s.String()
}

func getDictionary(dict *i18n.Dictionary) *i18n.Dictionary {
	if dict != nil {
		return dict
	}
	d := i18n.Get()
	return &d
}

func renderDockerHeader(ctx *DockerContext, dict *i18n.Dictionary, s *strings.Builder) {
	s.WriteString(theme.TitleStyle.Render(dict.DockerTabTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %s\n\n",
		theme.DimStyle.Render(dict.ServiceLabel), ctx.FormattedDockerStatus,
		theme.DimStyle.Render(dict.ImagesCountLabel), ctx.DockerInfo.ImagesCount,
		theme.DimStyle.Render(dict.VersionLabel), ctx.DockerInfo.Version,
	))
}

func renderEmptyContainersNotice(ctx *DockerContext, dict *i18n.Dictionary, s *strings.Builder) {
	if ctx.DockerInfo.Status == "running" {
		s.WriteString("  " + theme.DimStyle.Render(dict.ContainersNotRunning) + "\n")
	} else {
		s.WriteString("  " + theme.YellowStyle.Render(dict.DockerDaemonDown) + "\n")
	}
}

func renderFilterAndSortBadges(ctx *DockerContext, dict *i18n.Dictionary, s *strings.Builder) {
	if ctx.IsFiltering {
		s.WriteString("  " + theme.YellowStyle.Render(fmt.Sprintf(dict.FilterPrompt, ctx.FilterQuery)) + "  " + theme.DimStyle.Render(dict.FilterHint) + "\n\n")
		return
	}

	var statusBadges []string
	if ctx.FilterQuery != "" {
		statusBadges = append(statusBadges, theme.ActiveTabStyle.Render(fmt.Sprintf(dict.FilterActiveBadge, ctx.FilterQuery))+"  "+theme.DimStyle.Render(fmt.Sprintf(dict.FilterResultsCount, len(ctx.Containers), ctx.TotalContainersCount)))
	}
	if ctx.SortMode > 0 && ctx.SortMode < len(ctx.SortNames) {
		statusBadges = append(statusBadges, theme.CyanStyle.Render(fmt.Sprintf(dict.SortBadge, ctx.SortNames[ctx.SortMode])))
	}
	if len(statusBadges) > 0 {
		s.WriteString("  " + strings.Join(statusBadges, "   ") + "\n\n")
	}
}

func renderNoContainersFoundNotice(ctx *DockerContext, dict *i18n.Dictionary, s *strings.Builder) {
	s.WriteString("  " + theme.YellowStyle.Render(fmt.Sprintf(dict.NoContainersFound, ctx.FilterQuery)) + "\n")
	s.WriteString("  " + theme.DimStyle.Render(dict.PressEscToClear) + "\n")
}

func renderContainerTable(ctx *DockerContext, dict *i18n.Dictionary, s *strings.Builder) {
	headerText := fmt.Sprintf("  %-8s  %-20s  %-22s  %-20s  %-7s  %-14s  %s",
		dict.ColID,
		dict.ColName,
		dict.ColImage,
		dict.ColStatus,
		dict.ColCPU,
		dict.ColMem,
		dict.ColPorts,
	)
	s.WriteString(theme.DimStyle.Render(headerText) + "\n")
	s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", max(ctx.Width-4, minCardWidth))) + "\n")

	for i, c := range ctx.Containers {
		isSelected := (i == ctx.CursorIdx)
		s.WriteString(renderContainerRow(c, isSelected, ctx.Width))
	}
}

func formatContainerMetrics(c monitor.ContainerItem) (string, string) {
	if c.State != "running" {
		return "-", "-"
	}

	cpuRaw := fmt.Sprintf("%.1f%%", c.CPUPerc)
	var memRaw string
	if c.MemLimitMB > 0 {
		memRaw = fmt.Sprintf("%.1fM (%.1f%%)", c.MemUsageMB, c.MemPerc)
	} else {
		memRaw = fmt.Sprintf("%.1fM", c.MemUsageMB)
	}
	return cpuRaw, memRaw
}

func renderContainerRow(c monitor.ContainerItem, isSelected bool, termWidth int) string {
	prefix := "  "
	if isSelected {
		prefix = theme.GreenStyle.Render("► ")
	}

	cID := components.PadRight(components.Truncate(c.ID, colWidthID), colWidthID)
	cName := components.PadRight(components.Truncate(c.Name, colWidthName), colWidthName)
	cImg := components.PadRight(components.Truncate(c.Image, colWidthImage), colWidthImage)

	statusText, statusStyle := formatContainerStatus(c.State, c.Health)
	statusPadded := components.PadRight(components.Truncate(statusText, colWidthStatus), colWidthStatus)

	cpuRaw, memRaw := formatContainerMetrics(c)
	cpuPadded := components.PadLeft(cpuRaw, colWidthCPU)
	memPadded := components.PadRight(components.Truncate(memRaw, colWidthMem), colWidthMem)
	portsPreview := components.Truncate(c.Ports, max(minPortsWidth, termWidth-105))

	if isSelected {
		plainRow := fmt.Sprintf("%-8s  %-20s  %-22s  %-20s  %-7s  %-14s  %s",
			cID, cName, cImg, statusPadded, cpuPadded, memPadded, portsPreview,
		)
		return fmt.Sprintf("%s%s\n", prefix, theme.SelectedRowStyle.Render(plainRow))
	}

	row := fmt.Sprintf("%-8s  %-20s  %-22s  %s  %s  %-14s  %s",
		cID,
		theme.BoldStyle.Render(cName),
		theme.DimStyle.Render(cImg),
		statusStyle.Render(statusPadded),
		styleCPU(c.CPUPerc, c.State, cpuPadded),
		theme.DimStyle.Render(memPadded),
		portsPreview,
	)
	return fmt.Sprintf("%s%s\n", prefix, row)
}

func renderDetailRow(label, value string) string {
	var formattedLabel string
	if label == "" {
		formattedLabel = strings.Repeat(" ", detailColWidth)
	} else {
		formattedLabel = theme.BoldStyle.Render(components.PadRight(label, detailColWidth))
	}
	return fmt.Sprintf("  %s %s %s\n",
		theme.CyanStyle.Render("│"),
		formattedLabel,
		value,
	)
}

// FormatPortDisplay converts a raw comma-separated ports string into clean,
// deduplicated mappings with unicode arrows (e.g. "0.0.0.0:8080 → 8080/tcp").
func FormatPortDisplay(portsStr string) []string {
	if strings.TrimSpace(portsStr) == "" {
		return nil
	}
	rawParts := strings.Split(portsStr, ",")
	seenMap := make(map[string]bool)
	var result []string

	for _, p := range rawParts {
		tr := strings.TrimSpace(p)
		if tr == "" {
			continue
		}

		clean := strings.ReplaceAll(tr, "->", " → ")

		// Deduplicate IPv6 if identical IPv4 mapping exists
		dedupKey := clean
		if strings.HasPrefix(clean, "[::]:") {
			dedupKey = "0.0.0.0:" + strings.TrimPrefix(clean, "[::]:")
		}

		if seenMap[dedupKey] || seenMap[clean] {
			continue
		}
		seenMap[dedupKey] = true
		seenMap[clean] = true
		result = append(result, clean)
	}
	return result
}

func renderContainerDetailsCard(ctx *DockerContext, dict *i18n.Dictionary, s *strings.Builder) {
	if ctx.CursorIdx < 0 || ctx.CursorIdx >= len(ctx.Containers) {
		return
	}

	sel := ctx.Containers[ctx.CursorIdx]
	s.WriteString("\n")
	cardWidth := max(ctx.Width-4, minCardWidth)
	title := fmt.Sprintf(" %s: %s (%s) ", dict.SelectedContainerDetails, sel.Name, sel.ID)
	topDashLen := cardWidth - lipgloss.Width(title) - 3
	if topDashLen < 2 {
		topDashLen = 2
	}
	s.WriteString("  " + theme.CyanStyle.Render("┌─") + theme.BoldStyle.Render(title) + theme.CyanStyle.Render(strings.Repeat("─", topDashLen)+"┐") + "\n")

	// 1. IP / Network
	ipDisplay := sel.IPAddress
	if ipDisplay == "" {
		if sel.NetworkName != "" {
			ipDisplay = fmt.Sprintf("%s (%s)", theme.DimStyle.Render(dict.DetailNone), sel.NetworkName)
		} else {
			ipDisplay = theme.DimStyle.Render(dict.DetailNone)
		}
	} else if sel.NetworkName != "" {
		ipDisplay = fmt.Sprintf("%s  %s", theme.BoldStyle.Render(ipDisplay), theme.DimStyle.Render(fmt.Sprintf("(%s)", sel.NetworkName)))
	} else {
		ipDisplay = theme.BoldStyle.Render(ipDisplay)
	}
	s.WriteString(renderDetailRow(dict.DetailIPLabel, ipDisplay))

	// 2. Uptime / Started
	uptimeDisplay := sel.Status
	if uptimeDisplay == "" {
		uptimeDisplay = sel.State
	}
	if sel.Created != "" {
		uptimeDisplay = fmt.Sprintf("%s  %s %s",
			theme.GreenStyle.Render(uptimeDisplay),
			theme.DimStyle.Render("• "+dict.CreatedLabel),
			theme.DimStyle.Render(sel.Created),
		)
	} else {
		uptimeDisplay = theme.GreenStyle.Render(uptimeDisplay)
	}
	s.WriteString(renderDetailRow(dict.DetailUptimeLabel, uptimeDisplay))

	// 3. Restart Policy
	restartDisplay := sel.RestartPolicy
	if restartDisplay == "" {
		restartDisplay = theme.DimStyle.Render(dict.DetailNone)
	} else {
		switch strings.ToLower(restartDisplay) {
		case "always", "unless-stopped":
			restartDisplay = theme.GreenStyle.Render(restartDisplay)
		case "on-failure":
			restartDisplay = theme.YellowStyle.Render(restartDisplay)
		case "no":
			restartDisplay = theme.DimStyle.Render(restartDisplay)
		default:
			restartDisplay = theme.CyanStyle.Render(restartDisplay)
		}
	}
	s.WriteString(renderDetailRow(dict.DetailRestartPolicyLabel, restartDisplay))

	// 4. Ports
	portItems := FormatPortDisplay(sel.Ports)
	if len(portItems) == 0 {
		s.WriteString(renderDetailRow(dict.DetailPortsLabel, theme.DimStyle.Render(dict.DetailNone)))
	} else if len(portItems) == 1 {
		s.WriteString(renderDetailRow(dict.DetailPortsLabel, theme.CyanStyle.Render(portItems[0])))
	} else {
		for idx, port := range portItems {
			label := ""
			if idx == 0 {
				label = dict.DetailPortsLabel
			}
			s.WriteString(renderDetailRow(label, theme.CyanStyle.Render("• ")+theme.CyanStyle.Render(port)))
		}
	}

	// 5. Volumes / Mounts
	if len(sel.Mounts) == 0 {
		s.WriteString(renderDetailRow(dict.DetailMountsLabel, theme.DimStyle.Render(dict.DetailNone)))
	} else {
		for idx, mnt := range sel.Mounts {
			label := ""
			if idx == 0 {
				label = dict.DetailMountsLabel
			}
			s.WriteString(renderDetailRow(label, theme.YellowStyle.Render("• ")+mnt))
		}
	}

	// 6. Network I/O
	netIODisplay := theme.DimStyle.Render("0 B / 0 B")
	if sel.State == "running" {
		netIODisplay = fmt.Sprintf("%s / %s",
			theme.GreenStyle.Render(monitor.FormatBytes(int64(sel.NetRxBytes))),
			theme.CyanStyle.Render(monitor.FormatBytes(int64(sel.NetTxBytes))),
		)
	}
	s.WriteString(renderDetailRow(dict.DetailNetIOLabel, netIODisplay))

	// 7. Block I/O
	blockIODisplay := theme.DimStyle.Render("0 B / 0 B")
	if sel.State == "running" {
		blockIODisplay = fmt.Sprintf("%s / %s",
			theme.GreenStyle.Render(monitor.FormatBytes(int64(sel.BlockReadBytes))),
			theme.YellowStyle.Render(monitor.FormatBytes(int64(sel.BlockWriteBytes))),
		)
	}
	s.WriteString(renderDetailRow(dict.DetailBlockIOLabel, blockIODisplay))

	// 8. Compose Info
	if sel.ComposeProject != "" {
		composeDisplay := fmt.Sprintf("%s / %s", theme.BoldStyle.Render(sel.ComposeProject), sel.ComposeService)
		s.WriteString(renderDetailRow(dict.DetailComposeLabel, composeDisplay))
	}

	s.WriteString("  " + theme.CyanStyle.Render("└"+strings.Repeat("─", cardWidth-2)+"┘") + "\n")
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
