package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func (m Model) renderSwarmTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("DOCKER SWARM КЛАСТЕР") + "\n\n")

	if !m.swarm.IsActive {
		s.WriteString(fmt.Sprintf("  Статус: %s\n\n", m.swarm.StatusDisplay))
		s.WriteString("  " + dimStyle.Render("Хост не входит в активный кластер Swarm.") + "\n")
		s.WriteString("  Для создания кластера выполните: " + boldStyle.Render("docker swarm init") + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %d\n\n",
		dimStyle.Render("Статус:"), m.swarm.StatusDisplay,
		dimStyle.Render("Нод в кластере:"), m.swarm.NodesCount,
		dimStyle.Render("Менеджеров:"), m.swarm.ManagersCount,
	))

	if len(m.swarm.Nodes) > 0 {
		s.WriteString(boldStyle.Render("  УЗЛЫ КЛАСТЕРА (NODES):") + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8s %-10s\n",
			dimStyle.Render("NODE ID"), dimStyle.Render("HOSTNAME"), dimStyle.Render("ROLE"),
			dimStyle.Render("STATE"), dimStyle.Render("AVAILABILITY"), dimStyle.Render("TASKS"), dimStyle.Render("LEADER"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 88)) + "\n")
		for _, n := range m.swarm.Nodes {
			leaderMark := ""
			if n.IsLeader {
				leaderMark = yellowStyle.Render("[Leader]")
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
		s.WriteString(boldStyle.Render("  СЛУЖБЫ SWARM (SERVICES):") + "\n")
		s.WriteString(fmt.Sprintf("   %-14s %-20s %-24s %-16s %-18s\n",
			dimStyle.Render("SERVICE ID"), dimStyle.Render("NAME"), dimStyle.Render("IMAGE"),
			dimStyle.Render("REPLICAS"), dimStyle.Render("PORTS"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 94)) + "\n")
		for i, svc := range m.swarm.Services {
			s.WriteString(renderSwarmServiceRow(svc, i == m.swarmCursorIdx))
		}
	}

	return s.String()
}

func renderSwarmServiceRow(svc monitor.SwarmService, isSelected bool) string {
	ports := svc.Ports
	if ports == "" {
		ports = "-"
	}
	repColored := formatSwarmReplicas(svc)

	prefix := "  "
	rowContent := fmt.Sprintf("%-14s %-20s %-24s %-16s %-18s",
		svc.ID, boldStyle.Render(truncate(svc.Name, 18)), truncate(svc.Image, 22), repColored, truncate(ports, 18),
	)
	if isSelected {
		prefix = greenStyle.Render("► ")
		rowContent = selectedRowStyle.Render(fmt.Sprintf("%-14s %-20s %-24s %-16s %-18s",
			svc.ID, truncate(svc.Name, 18), truncate(svc.Image, 22), svc.Replicas, truncate(ports, 18),
		))
	}

	return fmt.Sprintf("%s%s\n", prefix, rowContent)
}

func formatSwarmReplicas(svc monitor.SwarmService) string {
	if svc.Mode == "replicated" && svc.TargetReplicas > 0 && svc.RunningReplicas < svc.TargetReplicas {
		if svc.RunningReplicas == 0 {
			return redStyle.Render(svc.Replicas)
		}
		return yellowStyle.Render(svc.Replicas)
	}
	return greenStyle.Render(svc.Replicas)
}
