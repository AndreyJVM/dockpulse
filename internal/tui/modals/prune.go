package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderPrune renders the Docker resource pruning menu and confirmation modal.
func RenderPrune(usage monitor.DockerDiskUsage, usageErr string, confirmStep int, dict *i18n.Dictionary) string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render(dict.PruneTitle) + "\n\n")

	if usageErr != "" {
		s.WriteString(fmt.Sprintf("  %s %s\n\n", theme.DimStyle.Render(dict.StatusLabel), theme.RedStyle.Render(usageErr)))
	} else {
		s.WriteString("  " + theme.BoldStyle.Render(dict.CurrentDiskUsage) + "\n")
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			theme.DimStyle.Render(dict.TypeCol), theme.DimStyle.Render(dict.SizeCol), theme.DimStyle.Render(dict.ActiveTotalCol), theme.DimStyle.Render(dict.ReclaimableCol),
		))
		s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", 58)) + "\n")

		u := usage
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			dict.TypeImages, monitor.FormatBytes(u.Images.TotalSize),
			fmt.Sprintf("%d / %d", u.Images.ActiveCount, u.Images.TotalCount),
			theme.YellowStyle.Render(monitor.FormatBytes(u.Images.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			dict.TypeContainers, monitor.FormatBytes(u.Containers.TotalSize),
			fmt.Sprintf("%d / %d", u.Containers.ActiveCount, u.Containers.TotalCount),
			theme.YellowStyle.Render(monitor.FormatBytes(u.Containers.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			dict.TypeVolumes, monitor.FormatBytes(u.Volumes.TotalSize),
			fmt.Sprintf("%d / %d", u.Volumes.ActiveCount, u.Volumes.TotalCount),
			theme.YellowStyle.Render(monitor.FormatBytes(u.Volumes.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			dict.TypeBuildCache, monitor.FormatBytes(u.BuildCache.TotalSize),
			fmt.Sprintf("%d / %d", u.BuildCache.ActiveCount, u.BuildCache.TotalCount),
			theme.YellowStyle.Render(monitor.FormatBytes(u.BuildCache.Reclaimable)),
		))
		s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", 58)) + "\n")
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n\n",
			theme.BoldStyle.Render(dict.TypeTotal), theme.BoldStyle.Render(monitor.FormatBytes(u.TotalSize)),
			"", theme.GreenStyle.Render(monitor.FormatBytes(u.TotalRecl)),
		))
	}

	if confirmStep > 0 {
		var actionName string
		switch confirmStep {
		case 1:
			actionName = dict.PruneWarnAction1
		case 2:
			actionName = dict.PruneWarnAction2
		case 3:
			actionName = dict.PruneWarnAction3
		case 4:
			actionName = dict.PruneWarnAction4
		}

		s.WriteString("  " + theme.RedStyle.Render(dict.PruneWarnTitle) + theme.BoldStyle.Render(actionName) + "\n")
		s.WriteString("  " + dict.PruneWarnIrrev + "\n\n")
		s.WriteString(fmt.Sprintf("  %s %s  |  %s %s\n",
			theme.GreenStyle.Render("[Y]"), dict.FooterConfirm,
			theme.DimStyle.Render("[N / Esc]"), dict.FooterCancel,
		))
	} else {
		s.WriteString("  " + theme.BoldStyle.Render(dict.PruneChooseAction) + "\n")
		s.WriteString(fmt.Sprintf("  %s %s\n", theme.GreenStyle.Render("[1]"), dict.PruneOpt1))
		s.WriteString(fmt.Sprintf("  %s %s\n", theme.GreenStyle.Render("[2]"), dict.PruneOpt2))
		s.WriteString(fmt.Sprintf("  %s %s\n", theme.GreenStyle.Render("[3]"), dict.PruneOpt3))
		s.WriteString(fmt.Sprintf("  %s %s\n\n", theme.YellowStyle.Render("[4]"), dict.PruneOpt4))
		s.WriteString(fmt.Sprintf("  %s %s\n", theme.DimStyle.Render("[Esc / Q]"), dict.FooterClose))
	}

	return theme.BoxStyle.Render(s.String())
}
