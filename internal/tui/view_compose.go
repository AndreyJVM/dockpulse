package tui

import (
	"fmt"
	"strings"
)

func (m Model) renderComposeTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf(m.i18n.ComposeTabTitle, len(m.compose))) + "\n\n")

	if len(m.compose) == 0 {
		s.WriteString("  " + dimStyle.Render(m.i18n.NoComposeProjects) + "\n")
		return s.String()
	}

	for i, p := range m.compose {
		prefix := "  "
		headerText := fmt.Sprintf("┌─ %s  [%s]  "+m.i18n.ComposeServicesUp,
			boldStyle.Render(p.Name), p.StatusDisplay, p.RunningServices, p.TotalServices,
		)
		if i == m.composeCursorIdx {
			prefix = greenStyle.Render("► ")
			headerText = selectedRowStyle.Render(fmt.Sprintf("┌─ %s  [%s]  "+m.i18n.ComposeServicesUp,
				p.Name, p.StatusDisplay, p.RunningServices, p.TotalServices,
			))
		}
		s.WriteString(fmt.Sprintf("%s%s\n", prefix, headerText))
		for _, svc := range p.Services {
			st := dimStyle.Render("○ " + svc.State)
			if svc.State == "running" {
				st = greenStyle.Render("● running")
			}
			s.WriteString(fmt.Sprintf("│   • %-16s %-20s %-22s %s\n",
				boldStyle.Render(svc.Name), st, truncate(svc.Image, 20), svc.Ports,
			))
		}
		s.WriteString("└" + strings.Repeat("─", 65) + "\n\n")
	}

	return s.String()
}
