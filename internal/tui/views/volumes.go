package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// VolumesContext contains data needed to render the Docker Volumes tab.
type VolumesContext struct {
	Volumes   []monitor.DockerVolume
	CursorIdx int
	Error     string
	I18n      *i18n.Dictionary
}

// RenderVolumes renders Docker volumes and details for the selected volume.
func RenderVolumes(ctx VolumesContext) string {
	d := getDictionary(ctx.I18n)
	var s strings.Builder
	s.WriteString(theme.TitleStyle.Render(fmt.Sprintf(d.VolumesTabTitle, len(ctx.Volumes))) + "\n\n")

	if ctx.Error != "" {
		s.WriteString("  " + theme.YellowStyle.Render(fmt.Sprintf(d.VolumesUnavailable, ctx.Error)) + "\n")
		return s.String()
	}
	if len(ctx.Volumes) == 0 {
		s.WriteString("  " + theme.DimStyle.Render(d.NoVolumes) + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %-24s  %-10s  %-8s  %-6s  %s\n",
		d.ColName, d.ColDriver, d.ColScope, d.ColReferences, d.ColSize))
	s.WriteString("  " + theme.DimStyle.Render(strings.Repeat("─", 68)) + "\n")
	for index, volume := range ctx.Volumes {
		size := "-"
		if volume.Size >= 0 {
			size = monitor.FormatBytes(volume.Size)
		}
		references := "-"
		if volume.RefCount >= 0 {
			references = fmt.Sprintf("%d", volume.RefCount)
		}
		row := fmt.Sprintf("%-24s  %-10s  %-8s  %-6s  %s",
			components.Truncate(volume.Name, 24),
			components.Truncate(volume.Driver, 10),
			components.Truncate(volume.Scope, 8),
			references,
			size,
		)
		if index == ctx.CursorIdx {
			s.WriteString(theme.GreenStyle.Render("► ") + theme.SelectedRowStyle.Render(row) + "\n")
		} else {
			s.WriteString("  " + row + "\n")
		}
	}

	if ctx.CursorIdx >= 0 && ctx.CursorIdx < len(ctx.Volumes) {
		volume := ctx.Volumes[ctx.CursorIdx]
		s.WriteString("\n" + theme.TitleStyle.Render(d.VolumeDetailsTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-18s %s\n", d.VolumeMountLabel, valueOrDash(volume.Mountpoint)))
		s.WriteString(fmt.Sprintf("  %-18s %s\n", d.ComposeProjectLabel, valueOrDash(volume.ComposeProject)))
	}
	return s.String()
}
