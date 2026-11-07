package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderInspect renders the detailed container inspection modal.
func RenderInspect(d monitor.ContainerDetail, dict *i18n.Dictionary) string {
	var s strings.Builder

	title := fmt.Sprintf(dict.InspectTitle, theme.BoldStyle.Render(d.Name))
	s.WriteString(theme.TitleStyle.Render(title) + "\n\n")

	s.WriteString(fmt.Sprintf("  %s %s\n", theme.DimStyle.Render("ID:"), d.ID))
	s.WriteString(fmt.Sprintf("  %s %s\n", theme.DimStyle.Render(dict.ImageLabel), theme.BoldStyle.Render(d.Image)))

	healthBadge := theme.DimStyle.Render("-")
	switch d.Health {
	case "healthy":
		healthBadge = theme.GreenStyle.Render("[healthy]")
	case "unhealthy":
		healthBadge = theme.RedStyle.Render("[unhealthy]")
	case "starting":
		healthBadge = theme.YellowStyle.Render("[starting]")
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		theme.DimStyle.Render(dict.StatusLabel), theme.GreenStyle.Render(d.Status),
		theme.DimStyle.Render(dict.HealthLabel), healthBadge,
	))
	if d.State == "running" {
		if d.CPUPerc > 0 || d.MemUsageMB > 0 {
			s.WriteString(fmt.Sprintf("  %s %.1f%%    %s %.1f MB (%.1f%%)\n",
				theme.DimStyle.Render("CPU:"), d.CPUPerc,
				theme.DimStyle.Render(dict.MemLabel), d.MemUsageMB, d.MemPerc,
			))
		}
		if d.NetRxBytes > 0 || d.NetTxBytes > 0 || d.BlockReadBytes > 0 || d.BlockWriteBytes > 0 {
			s.WriteString(fmt.Sprintf("  %s %s / %s    %s %s / %s\n",
				theme.DimStyle.Render("Net I/O:"),
				monitor.FormatBytes(int64(d.NetRxBytes)),
				monitor.FormatBytes(int64(d.NetTxBytes)),
				theme.DimStyle.Render("Block I/O:"),
				monitor.FormatBytes(int64(d.BlockReadBytes)),
				monitor.FormatBytes(int64(d.BlockWriteBytes)),
			))
		}
	}
	restartVal := d.RestartPolicy
	if restartVal == "" {
		restartVal = dict.DetailNone
	}
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		theme.DimStyle.Render(dict.CreatedLabel), d.Created,
		theme.DimStyle.Render(dict.DetailRestartPolicyLabel), theme.CyanStyle.Render(restartVal),
	))
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n", theme.DimStyle.Render(dict.IPAddrLabel), theme.BoldStyle.Render(d.IP), theme.DimStyle.Render(dict.GatewayLabel), d.Gateway))
	s.WriteString(fmt.Sprintf("  %s %s\n", theme.DimStyle.Render(dict.CommandLabel), components.Truncate(d.Command, 60)))

	if len(d.Mounts) > 0 {
		s.WriteString("\n  " + theme.BoldStyle.Render(dict.MountsTitle) + "\n")
		for _, mnt := range d.Mounts {
			s.WriteString(fmt.Sprintf("   • %s\n", mnt))
		}
	}

	return theme.BoxStyle.Render(s.String())
}
