package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func (m Model) renderSwarmTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(m.i18n.SwarmTabTitle) + "\n\n")

	swarmStatus := m.formatSwarmStatus(m.swarm.IsActive)
	if !m.swarm.IsActive {
		s.WriteString(fmt.Sprintf("  %s %s\n\n", m.i18n.DaemonStatus, swarmStatus))
		s.WriteString("  " + dimStyle.Render(m.i18n.SwarmInactive) + "\n")
		s.WriteString("  " + fmt.Sprintf(m.i18n.SwarmInitHint, boldStyle.Render("docker swarm init")) + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %d\n\n",
		dimStyle.Render(m.i18n.DaemonStatus), swarmStatus,
		dimStyle.Render(m.i18n.SwarmNodesCount), m.swarm.NodesCount,
		dimStyle.Render(m.i18n.SwarmManagers), m.swarm.ManagersCount,
	))

	if len(m.swarm.Nodes) > 0 {
		s.WriteString(boldStyle.Render("  "+m.i18n.SwarmNodesTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8s %-10s\n",
			dimStyle.Render(m.i18n.ColNodeID), dimStyle.Render(m.i18n.ColHostname), dimStyle.Render(m.i18n.ColRole),
			dimStyle.Render(m.i18n.ColState), dimStyle.Render(m.i18n.ColAvailability), dimStyle.Render(m.i18n.ColTasks), dimStyle.Render(m.i18n.ColLeader),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 88)) + "\n")
		for _, n := range m.swarm.Nodes {
			leaderMark := ""
			if n.IsLeader {
				leaderMark = yellowStyle.Render(m.i18n.LeaderBadge)
			}
			avail := n.Availability
			if avail == "" {
				avail = "active"
			}
			s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8d %-10s\n",
				n.ID, boldStyle.Render(truncate(n.Hostname, 18)), n.Role, n.State, avail, n.ActiveTasks, leaderMark,
			))
		}
		s.WriteString("\n")
	}

	if len(m.swarm.Services) > 0 {
		s.WriteString(boldStyle.Render("  "+m.i18n.SwarmSvcsTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-20s %-10s %-12s %s\n",
			dimStyle.Render(m.i18n.ColServiceID), dimStyle.Render(m.i18n.ColName), dimStyle.Render(m.i18n.ColImage),
			dimStyle.Render("MODE"), dimStyle.Render(m.i18n.ColReplicas), dimStyle.Render(m.i18n.ColPorts),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 88)) + "\n")
		for i, svc := range m.swarm.Services {
			prefix := "  "
			if i == m.swarmCursorIdx {
				prefix = greenStyle.Render("► ")
			}

			repStr := svc.Replicas
			if svc.Mode == "replicated" && svc.TargetReplicas > 0 {
				if svc.RunningReplicas == 0 {
					repStr = redStyle.Render(svc.Replicas)
				} else if svc.RunningReplicas < svc.TargetReplicas {
					repStr = yellowStyle.Render(svc.Replicas)
				} else {
					repStr = greenStyle.Render(svc.Replicas)
				}
			}

			row := fmt.Sprintf("%-14s %-20s %-20s %-10s %-12s %s",
				svc.ID, boldStyle.Render(truncate(svc.Name, 18)), truncate(svc.Image, 18),
				svc.Mode, repStr, svc.Ports,
			)

			if i == m.swarmCursorIdx {
				s.WriteString(fmt.Sprintf("%s%s\n", prefix, selectedRowStyle.Render(row)))
			} else {
				s.WriteString(fmt.Sprintf("%s%s\n", prefix, row))
			}
		}
	}

	return s.String()
}
