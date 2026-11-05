package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// ComposeContext contains data needed to render the Compose Stacks tab.
type ComposeContext struct {
	Projects     []monitor.ComposeProject
	CursorIdx    int
	FormatStatus func(status, display string) string
	I18n         *i18n.Dictionary
}

// RenderCompose renders the Compose Stacks tab.
func RenderCompose(ctx ComposeContext) string {
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render(fmt.Sprintf(ctx.I18n.ComposeTabTitle, len(ctx.Projects))) + "\n\n")

	if len(ctx.Projects) == 0 {
		s.WriteString("  " + theme.DimStyle.Render(ctx.I18n.NoComposeProjects) + "\n")
		return s.String()
	}

	for i, p := range ctx.Projects {
		prefix := "  "
		statusDisp := ctx.FormatStatus(p.Status, p.StatusDisplay)
		headerText := fmt.Sprintf("┌─ %s  [%s]  "+ctx.I18n.ComposeServicesUp,
			theme.BoldStyle.Render(p.Name), statusDisp, p.RunningServices, p.TotalServices,
		)
		if i == ctx.CursorIdx {
			prefix = theme.GreenStyle.Render("► ")
			headerText = theme.SelectedRowStyle.Render(fmt.Sprintf("┌─ %s  [%s]  "+ctx.I18n.ComposeServicesUp,
				p.Name, statusDisp, p.RunningServices, p.TotalServices,
			))
		}
		s.WriteString(fmt.Sprintf("%s%s\n", prefix, headerText))
		for _, svc := range p.Services {
			st := theme.DimStyle.Render("○ " + svc.State)
			if svc.State == "running" {
				st = theme.GreenStyle.Render("● running")
			}
			s.WriteString(fmt.Sprintf("│   • %-16s %-20s %-22s %s\n",
				theme.BoldStyle.Render(svc.Name), st, components.Truncate(svc.Image, 20), svc.Ports,
			))
		}
		s.WriteString("└" + strings.Repeat("─", 65) + "\n\n")
	}

	return s.String()
}
