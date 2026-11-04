package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func (m Model) renderLogsModal() string {
	var s strings.Builder
	title := fmt.Sprintf(m.i18n.LogsTitle, boldStyle.Render(m.logsTargetName), dimStyle.Render(m.logsTargetID))
	s.WriteString(titleStyle.Render(title) + "\n\n")

	maxVisible := 18
	if len(m.logsLines) == 0 {
		s.WriteString("  " + dimStyle.Render(m.i18n.NoLogsOutput) + "\n")
	} else {
		start := min(m.logsScroll, max(0, len(m.logsLines)-1))
		end := min(start+maxVisible, len(m.logsLines))

		for i := start; i < end; i++ {
			line := m.logsLines[i]
			s.WriteString(fmt.Sprintf(" %s %s\n", dimStyle.Render("│"), truncate(line, max(40, m.width-8))))
		}

		s.WriteString(fmt.Sprintf("\n  %s\n", dimStyle.Render(fmt.Sprintf(m.i18n.LogsRangeCount, start+1, end, len(m.logsLines)))))
	}

	return boxStyle.Render(s.String())
}

func (m Model) renderInspectModal() string {
	var s strings.Builder
	d := m.inspectTarget

	title := fmt.Sprintf(m.i18n.InspectTitle, boldStyle.Render(d.Name))
	s.WriteString(titleStyle.Render(title) + "\n\n")

	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("ID:"), d.ID))
	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(m.i18n.ImageLabel), boldStyle.Render(d.Image)))

	healthBadge := dimStyle.Render("-")
	switch d.Health {
	case "healthy":
		healthBadge = greenStyle.Render("[healthy]")
	case "unhealthy":
		healthBadge = redStyle.Render("[unhealthy]")
	case "starting":
		healthBadge = yellowStyle.Render("[starting]")
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		dimStyle.Render(m.i18n.StatusLabel), greenStyle.Render(d.Status),
		dimStyle.Render(m.i18n.HealthLabel), healthBadge,
	))
	if d.State == "running" && (d.CPUPerc > 0 || d.MemUsageMB > 0) {
		s.WriteString(fmt.Sprintf("  %s %.1f%%    %s %.1f MB (%.1f%%)\n",
			dimStyle.Render("CPU:"), d.CPUPerc,
			dimStyle.Render(m.i18n.MemLabel), d.MemUsageMB, d.MemPerc,
		))
	}
	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(m.i18n.CreatedLabel), d.Created))
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n", dimStyle.Render(m.i18n.IPAddrLabel), boldStyle.Render(d.IP), dimStyle.Render(m.i18n.GatewayLabel), d.Gateway))
	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render(m.i18n.CommandLabel), truncate(d.Command, 60)))

	if len(d.Mounts) > 0 {
		s.WriteString("\n  " + boldStyle.Render(m.i18n.MountsTitle) + "\n")
		for _, mnt := range d.Mounts {
			s.WriteString(fmt.Sprintf("   • %s\n", mnt))
		}
	}

	return boxStyle.Render(s.String())
}

func (m Model) renderPruneModal() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(m.i18n.PruneTitle) + "\n\n")

	if m.pruneUsageErr != "" {
		s.WriteString(fmt.Sprintf("  %s %s\n\n", dimStyle.Render(m.i18n.StatusLabel), redStyle.Render(m.pruneUsageErr)))
	} else {
		s.WriteString("  " + boldStyle.Render(m.i18n.CurrentDiskUsage) + "\n")
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			dimStyle.Render(m.i18n.TypeCol), dimStyle.Render(m.i18n.SizeCol), dimStyle.Render(m.i18n.ActiveTotalCol), dimStyle.Render(m.i18n.ReclaimableCol),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 58)) + "\n")

		u := m.pruneUsage
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			m.i18n.TypeImages, monitor.FormatBytes(u.Images.TotalSize),
			fmt.Sprintf("%d / %d", u.Images.ActiveCount, u.Images.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.Images.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			m.i18n.TypeContainers, monitor.FormatBytes(u.Containers.TotalSize),
			fmt.Sprintf("%d / %d", u.Containers.ActiveCount, u.Containers.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.Containers.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			m.i18n.TypeVolumes, monitor.FormatBytes(u.Volumes.TotalSize),
			fmt.Sprintf("%d / %d", u.Volumes.ActiveCount, u.Volumes.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.Volumes.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			m.i18n.TypeBuildCache, monitor.FormatBytes(u.BuildCache.TotalSize),
			fmt.Sprintf("%d / %d", u.BuildCache.ActiveCount, u.BuildCache.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.BuildCache.Reclaimable)),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 58)) + "\n")
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n\n",
			boldStyle.Render(m.i18n.TypeTotal), boldStyle.Render(monitor.FormatBytes(u.TotalSize)),
			"", greenStyle.Render(monitor.FormatBytes(u.TotalRecl)),
		))
	}

	if m.pruneConfirm > 0 {
		var actionName string
		switch m.pruneConfirm {
		case 1:
			actionName = m.i18n.PruneWarnAction1
		case 2:
			actionName = m.i18n.PruneWarnAction2
		case 3:
			actionName = m.i18n.PruneWarnAction3
		case 4:
			actionName = m.i18n.PruneWarnAction4
		}

		s.WriteString("  " + redStyle.Render(m.i18n.PruneWarnTitle) + boldStyle.Render(actionName) + "\n")
		s.WriteString("  " + m.i18n.PruneWarnIrrev + "\n\n")
		s.WriteString(fmt.Sprintf("  %s %s  |  %s %s\n",
			greenStyle.Render("[Y]"), m.i18n.FooterConfirm,
			dimStyle.Render("[N / Esc]"), m.i18n.FooterCancel,
		))
	} else {
		s.WriteString("  " + boldStyle.Render(m.i18n.PruneChooseAction) + "\n")
		s.WriteString(fmt.Sprintf("  %s %s\n", greenStyle.Render("[1]"), m.i18n.PruneOpt1))
		s.WriteString(fmt.Sprintf("  %s %s\n", greenStyle.Render("[2]"), m.i18n.PruneOpt2))
		s.WriteString(fmt.Sprintf("  %s %s\n", greenStyle.Render("[3]"), m.i18n.PruneOpt3))
		s.WriteString(fmt.Sprintf("  %s %s\n\n", yellowStyle.Render("[4]"), m.i18n.PruneOpt4))
		s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("[Esc / Q]"), m.i18n.FooterClose))
	}

	return boxStyle.Render(s.String())
}

func (m Model) renderComposeDownModal() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render(m.i18n.ComposeDownTitle) + "\n\n")
	s.WriteString(fmt.Sprintf("  "+m.i18n.StackLabel+"\n\n", boldStyle.Render(m.composeDownTarget)))
	s.WriteString(fmt.Sprintf("  %s %s\n\n", yellowStyle.Render(m.i18n.WarningLabel), m.i18n.ComposeDownNotice))
	s.WriteString(fmt.Sprintf("  "+m.i18n.ComposeDownConfirm+"\n\n", boldStyle.Render(m.composeDownTarget)))
	s.WriteString(fmt.Sprintf("  %s %s (Down)    %s %s\n", greenStyle.Render("[Y]"), m.i18n.FooterConfirm, dimStyle.Render("[N / Esc]"), m.i18n.FooterCancel))
	return boxStyle.Render(s.String())
}

func (m Model) renderHelpBox() string {
	content := fmt.Sprintf(`%s

  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s

  `+m.i18n.HelpAutoInt,
		boldStyle.Render(m.i18n.HelpTitle),
		greenStyle.Render("[1] - [4]"), m.i18n.HelpKeyTabs,
		greenStyle.Render("[Tab] / [Shift+Tab] / [←] / [→]"), m.i18n.HelpKeyPrev,
		greenStyle.Render("[↑] / [↓] / [k] / [j]"), m.i18n.HelpKeyNav,
		greenStyle.Render("[/]"), m.i18n.HelpKeySearch,
		greenStyle.Render("[O]"), m.i18n.HelpKeySort,
		greenStyle.Render("[L]"), m.i18n.HelpKeyLogs,
		greenStyle.Render("[Enter] / [I]"), m.i18n.HelpKeyInfo,
		greenStyle.Render("[E]"), m.i18n.HelpKeyShell,
		greenStyle.Render("[S]"), m.i18n.HelpKeySS,
		greenStyle.Render("[U]"), m.i18n.HelpKeyStart,
		greenStyle.Render("[P]"), m.i18n.HelpKeyRest,
		greenStyle.Render("[D]"), m.i18n.HelpKeyDown,
		greenStyle.Render("[C]"), m.i18n.HelpKeyPrune,
		greenStyle.Render("[R]"), m.i18n.HelpKeyRef,
		greenStyle.Render("[?]"), m.i18n.HelpKeyHelp,
		greenStyle.Render("[Q] / [Esc]"), m.i18n.HelpKeyQuit,
		yellowStyle.Render(m.refreshInterval.String()),
	)
	return boxStyle.Render(content)
}
