package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// NetworksContext contains data needed to render the Docker Networks tab.
type NetworksContext struct {
	Networks  []monitor.DockerNetwork
	CursorIdx int
	Error     string
	I18n      *i18n.Dictionary
}

// RenderNetworks renders Docker networks and details for the selected network.
func RenderNetworks(ctx NetworksContext) string {
	d := getDictionary(ctx.I18n)
	var s strings.Builder
	s.WriteString(theme.TitleStyle.Render(fmt.Sprintf(d.NetworksTabTitle, len(ctx.Networks))) + "\n\n")

	if ctx.Error != "" {
		s.WriteString("  " + theme.YellowStyle.Render(fmt.Sprintf(d.NetworksUnavailable, ctx.Error)) + "\n")
		return s.String()
	}
	if len(ctx.Networks) == 0 {
		s.WriteString("  " + theme.DimStyle.Render(d.NoNetworks) + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %-12s  %-20s  %-10s  %-10s  %s\n",
		d.ColID, d.ColName, d.ColDriver, d.ColContainers, d.ColSubnet))
	s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", 76)) + "\n")
	for index, network := range ctx.Networks {
		row := fmt.Sprintf("%-12s  %-20s  %-10s  %-10d  %s",
			components.Truncate(network.ID, 12),
			components.Truncate(network.Name, 20),
			components.Truncate(network.Driver, 10),
			network.Containers,
			network.Subnet,
		)
		if index == ctx.CursorIdx {
			s.WriteString(theme.GreenStyle.Render("► ") + theme.SelectedRowStyle.Render(row) + "\n")
		} else {
			s.WriteString("  " + row + "\n")
		}
	}

	if ctx.CursorIdx >= 0 && ctx.CursorIdx < len(ctx.Networks) {
		network := ctx.Networks[ctx.CursorIdx]
		s.WriteString("\n" + theme.TitleStyle.Render(d.NetworkDetailsTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-18s %s\n", d.NetworkGatewayLabel, valueOrDash(network.Gateway)))
		s.WriteString(fmt.Sprintf("  %-18s %s\n", d.NetworkScopeLabel, valueOrDash(network.Scope)))
		s.WriteString(fmt.Sprintf("  %-18s %t\n", d.NetworkInternalLabel, network.Internal))
		s.WriteString(fmt.Sprintf("  %-18s %t\n", d.NetworkAttachableLabel, network.Attachable))
		s.WriteString(fmt.Sprintf("  %-18s %s\n", d.ComposeProjectLabel, valueOrDash(network.ComposeProject)))
	}
	return s.String()
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
