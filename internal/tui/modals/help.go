package modals

import (
	"fmt"
	"strings"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

const helpKeyColumnWidth = 34

type helpItem struct {
	key  string
	desc string
}

type helpSection struct {
	title string
	items []helpItem
}

// RenderHelp renders the categorized and aligned keyboard shortcuts help dialog box.
func RenderHelp(refreshInterval time.Duration, dict *i18n.Dictionary) string {
	d := getModalDictionary(dict)
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render(d.HelpTitle) + "\n\n")

	sections := []helpSection{
		{
			title: d.HelpSecNav,
			items: []helpItem{
				{key: "[1] - [4]", desc: d.HelpKeyTabs},
				{key: "[Tab] / [Shift+Tab] / [←] / [→]", desc: d.HelpKeyPrev},
				{key: "[↑] / [↓] / [k] / [j]", desc: d.HelpKeyNav},
			},
		},
		{
			title: d.HelpSecContainers,
			items: []helpItem{
				{key: "[/]", desc: d.HelpKeySearch},
				{key: "[O]", desc: d.HelpKeySort},
				{key: "[L]", desc: d.HelpKeyLogs},
				{key: "[Enter] / [I]", desc: d.HelpKeyInfo},
				{key: "[E]", desc: d.HelpKeyShell},
				{key: "[S]", desc: d.HelpKeySS},
				{key: "[P]", desc: d.HelpKeyRest},
				{key: "[K]", desc: d.HelpKeyKill},
				{key: "[x]", desc: d.HelpKeyRemove},
			},
		},
		{
			title: d.HelpSecCompose,
			items: []helpItem{
				{key: "[U]", desc: d.HelpKeyStart},
				{key: "[D]", desc: d.HelpKeyDown},
			},
		},
		{
			title: d.HelpSecGeneral,
			items: []helpItem{
				{key: "[C]", desc: d.HelpKeyPrune},
				{key: "[R]", desc: d.HelpKeyRef},
				{key: "[?]", desc: d.HelpKeyHelp},
				{key: "[Q] / [Esc]", desc: d.HelpKeyQuit},
			},
		},
	}

	for _, sec := range sections {
		s.WriteString(fmt.Sprintf("  %s\n", theme.CyanStyle.Render("• "+sec.title)))
		for _, item := range sec.items {
			padCount := helpKeyColumnWidth - len([]rune(item.key))
			padding := ""
			if padCount > 0 {
				padding = strings.Repeat(" ", padCount)
			}
			s.WriteString(fmt.Sprintf("    %s%s %s\n",
				theme.GreenStyle.Render(item.key),
				padding,
				item.desc,
			))
		}
		s.WriteString("\n")
	}

	s.WriteString(fmt.Sprintf("  "+d.HelpAutoInt, theme.YellowStyle.Render(refreshInterval.String())))

	return theme.BoxStyle.Render(s.String())
}
