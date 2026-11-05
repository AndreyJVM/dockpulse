package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// SwarmContext contains data needed to render the Swarm Cluster tab.
type SwarmContext struct {
	Swarm        monitor.SwarmSummary
	CursorIdx    int
	FormatStatus func(isActive bool) string
	I18n         *i18n.Dictionary
}

// RenderSwarm renders the Swarm Cluster tab.
func RenderSwarm(ctx SwarmContext) string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render(ctx.I18n.SwarmTabTitle) + "\n\n")

	swarmStatus := ctx.FormatStatus(ctx.Swarm.IsActive)
	if !ctx.Swarm.IsActive {
		s.WriteString(fmt.Sprintf("  %s %s\n\n", ctx.I18n.DaemonStatus, swarmStatus))
		s.WriteString("  " + theme.DimStyle.Render(ctx.I18n.SwarmInactive) + "\n")
		s.WriteString("  " + fmt.Sprintf(ctx.I18n.SwarmInitHint, theme.BoldStyle.Render("docker swarm init")) + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %d\n\n",
		theme.DimStyle.Render(ctx.I18n.DaemonStatus), swarmStatus,
		theme.DimStyle.Render(ctx.I18n.SwarmNodesCount), ctx.Swarm.NodesCount,
		theme.DimStyle.Render(ctx.I18n.SwarmManagers), ctx.Swarm.ManagersCount,
	))

	if len(ctx.Swarm.Nodes) > 0 {
		s.WriteString(theme.BoldStyle.Render("  "+ctx.I18n.SwarmNodesTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8s %-10s\n",
			theme.DimStyle.Render(ctx.I18n.ColNodeID), theme.DimStyle.Render(ctx.I18n.ColHostname), theme.DimStyle.Render(ctx.I18n.ColRole),
			theme.DimStyle.Render(ctx.I18n.ColState), theme.DimStyle.Render(ctx.I18n.ColAvailability), theme.DimStyle.Render(ctx.I18n.ColTasks), theme.DimStyle.Render(ctx.I18n.ColLeader),
		))
		s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", 88)) + "\n")
		for _, n := range ctx.Swarm.Nodes {
			leaderMark := ""
			if n.IsLeader {
				leaderMark = theme.YellowStyle.Render(ctx.I18n.LeaderBadge)
			}
			avail := n.Availability
			if avail == "" {
				avail = "active"
			}
			s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8d %-10s\n",
				n.ID, theme.BoldStyle.Render(components.Truncate(n.Hostname, 18)), n.Role, n.State, avail, n.ActiveTasks, leaderMark,
			))
		}
		s.WriteString("\n")
	}

	if len(ctx.Swarm.Services) > 0 {
		s.WriteString(theme.BoldStyle.Render("  "+ctx.I18n.SwarmSvcsTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-20s %-10s %-12s %s\n",
			theme.DimStyle.Render(ctx.I18n.ColServiceID), theme.DimStyle.Render(ctx.I18n.ColName), theme.DimStyle.Render(ctx.I18n.ColImage),
			theme.DimStyle.Render("MODE"), theme.DimStyle.Render(ctx.I18n.ColReplicas), theme.DimStyle.Render(ctx.I18n.ColPorts),
		))
		s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", 88)) + "\n")
		for i, svc := range ctx.Swarm.Services {
			prefix := "  "
			if i == ctx.CursorIdx {
				prefix = theme.GreenStyle.Render("► ")
			}

			repStr := svc.Replicas
			if svc.Mode == "replicated" && svc.TargetReplicas > 0 {
				if svc.RunningReplicas == 0 {
					repStr = theme.RedStyle.Render(svc.Replicas)
				} else if svc.RunningReplicas < svc.TargetReplicas {
					repStr = theme.YellowStyle.Render(svc.Replicas)
				} else {
					repStr = theme.GreenStyle.Render(svc.Replicas)
				}
			}

			row := fmt.Sprintf("%-14s %-20s %-20s %-10s %-12s %s",
				svc.ID, theme.BoldStyle.Render(components.Truncate(svc.Name, 18)), components.Truncate(svc.Image, 18),
				svc.Mode, repStr, svc.Ports,
			)

			if i == ctx.CursorIdx {
				s.WriteString(fmt.Sprintf("%s%s\n", prefix, theme.SelectedRowStyle.Render(row)))
			} else {
				s.WriteString(fmt.Sprintf("%s%s\n", prefix, row))
			}
		}
	}

	return s.String()
}
