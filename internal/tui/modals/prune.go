package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

const (
	pruneColTypeWidth   = 16
	pruneColSizeWidth   = 12
	pruneColActiveWidth = 14
	pruneTableWidth     = 58
)

// RenderPrune renders the Docker resource pruning menu and confirmation modal.
func RenderPrune(usage monitor.DockerDiskUsage, usageErr string, confirmStep int, dict *i18n.Dictionary) string {
	var s strings.Builder
	d := getModalDictionary(dict)

	s.WriteString(theme.TitleStyle.Render(d.PruneTitle) + "\n\n")

	if usageErr != "" {
		s.WriteString(fmt.Sprintf("  %s %s\n\n", theme.DimStyle.Render(d.StatusLabel), theme.RedStyle.Render(usageErr)))
	} else {
		s.WriteString(renderPruneUsageTable(usage, d))
	}

	if confirmStep > 0 {
		s.WriteString(renderPruneConfirmStep(confirmStep, d))
	} else {
		s.WriteString(renderPruneActionMenu(d))
	}

	return theme.BoxStyle.Render(s.String())
}

func renderPruneRow(name string, item monitor.DockerDiskItem) string {
	return fmt.Sprintf("  %-16s %-12s %-14s %s\n",
		name,
		monitor.FormatBytes(item.TotalSize),
		fmt.Sprintf("%d / %d", item.ActiveCount, item.TotalCount),
		theme.YellowStyle.Render(monitor.FormatBytes(item.Reclaimable)),
	)
}

func renderPruneUsageTable(u monitor.DockerDiskUsage, d *i18n.Dictionary) string {
	var sb strings.Builder
	sb.WriteString("  " + theme.BoldStyle.Render(d.CurrentDiskUsage) + "\n")
	sb.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
		theme.DimStyle.Render(d.TypeCol),
		theme.DimStyle.Render(d.SizeCol),
		theme.DimStyle.Render(d.ActiveTotalCol),
		theme.DimStyle.Render(d.ReclaimableCol),
	))
	sb.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", pruneTableWidth)) + "\n")

	categories := []struct {
		name string
		item monitor.DockerDiskItem
	}{
		{d.TypeImages, u.Images},
		{d.TypeContainers, u.Containers},
		{d.TypeVolumes, u.Volumes},
		{d.TypeBuildCache, u.BuildCache},
	}

	for _, c := range categories {
		sb.WriteString(renderPruneRow(c.name, c.item))
	}

	totalSize := u.TotalSize
	totalRecl := u.TotalRecl
	if totalSize == 0 && totalRecl == 0 {
		totalSize = u.Images.TotalSize + u.Containers.TotalSize + u.Volumes.TotalSize + u.BuildCache.TotalSize
		totalRecl = u.Images.Reclaimable + u.Containers.Reclaimable + u.Volumes.Reclaimable + u.BuildCache.Reclaimable
	}

	sb.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", pruneTableWidth)) + "\n")
	sb.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n\n",
		theme.BoldStyle.Render(d.TypeTotal),
		theme.BoldStyle.Render(monitor.FormatBytes(totalSize)),
		"",
		theme.GreenStyle.Render(monitor.FormatBytes(totalRecl)),
	))

	return sb.String()
}

func getPruneActionName(step int, d *i18n.Dictionary) string {
	switch step {
	case 1:
		return d.PruneWarnAction1
	case 2:
		return d.PruneWarnAction2
	case 3:
		return d.PruneWarnAction3
	case 4:
		return d.PruneWarnAction4
	default:
		return ""
	}
}

func renderPruneConfirmStep(confirmStep int, d *i18n.Dictionary) string {
	actionName := getPruneActionName(confirmStep, d)
	var sb strings.Builder
	sb.WriteString("  " + theme.RedStyle.Render(d.PruneWarnTitle) + theme.BoldStyle.Render(actionName) + "\n")
	sb.WriteString("  " + d.PruneWarnIrrev + "\n\n")
	sb.WriteString(fmt.Sprintf("  %s %s  |  %s %s\n",
		theme.GreenStyle.Render("[Y]"), d.FooterConfirm,
		theme.DimStyle.Render("[N / Esc]"), d.FooterCancel,
	))
	return sb.String()
}

func renderPruneActionMenu(d *i18n.Dictionary) string {
	var sb strings.Builder
	sb.WriteString("  " + theme.BoldStyle.Render(d.PruneChooseAction) + "\n")
	sb.WriteString(fmt.Sprintf("  %s %s\n", theme.GreenStyle.Render("[1]"), d.PruneOpt1))
	sb.WriteString(fmt.Sprintf("  %s %s\n", theme.GreenStyle.Render("[2]"), d.PruneOpt2))
	sb.WriteString(fmt.Sprintf("  %s %s\n", theme.GreenStyle.Render("[3]"), d.PruneOpt3))
	sb.WriteString(fmt.Sprintf("  %s %s\n\n", theme.YellowStyle.Render("[4]"), d.PruneOpt4))
	sb.WriteString(fmt.Sprintf("  %s %s\n", theme.DimStyle.Render("[Esc / Q]"), d.FooterClose))
	return sb.String()
}
