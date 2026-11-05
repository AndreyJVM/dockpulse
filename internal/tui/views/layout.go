package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderHeader renders the top status banner with app name, version, and clock.
func RenderHeader(version, lastUpdate string) string {
	verStr := version
	if verStr == "" {
		verStr = "v0.4.0"
	}
	return theme.HeaderStyle.Render("Dockpulse") +
		theme.DimStyle.Render(fmt.Sprintf("  %s  ", verStr)) +
		theme.GreenStyle.Render("["+lastUpdate+"]")
}

// RenderTabs renders the tab navigation bar.
func RenderTabs(tabs []string, activeTab int, width int) string {
	var renderedTabs []string
	for i, tab := range tabs {
		tabIdx := i + 1
		if activeTab == tabIdx {
			renderedTabs = append(renderedTabs, theme.ActiveTabStyle.Render(fmt.Sprintf("[%s]", tab)))
		} else {
			renderedTabs = append(renderedTabs, theme.InactiveTabStyle.Render(fmt.Sprintf(" %s ", tab)))
		}
	}
	bar := strings.Join(renderedTabs, " ") + "\n" +
		theme.DimStyle.Render(strings.Repeat("─", max(width, 70)))
	return bar
}

// FooterContext contains parameters needed to render context-aware bottom hotkey hints.
type FooterContext struct {
	ActiveTab    int
	ActiveModal  string
	PruneConfirm int
	IsFiltering  bool
	FilterQuery  string
	Width        int
	I18n         *i18n.Dictionary
}

// RenderFooter renders the bottom hotkeys guide.
func RenderFooter(ctx FooterContext) string {
	d := ctx.I18n
	var content string

	if ctx.ActiveModal != "" {
		if ctx.ActiveModal == "compose_down" {
			content = theme.BoldStyle.Render("[Y]") + " " + d.FooterConfirm + " (Down)  |  " +
				theme.BoldStyle.Render("[N / Esc]") + " " + d.FooterCancel
		} else if ctx.ActiveModal == "prune" {
			if ctx.PruneConfirm > 0 {
				content = theme.BoldStyle.Render("[Y]") + " " + d.FooterConfirmPrune + "  |  " +
					theme.BoldStyle.Render("[N / Esc]") + " " + d.FooterCancel
			} else {
				content = theme.BoldStyle.Render("[1-4]") + " " + d.FooterSelectAction + "  |  " +
					theme.BoldStyle.Render("[Esc / Q]") + " " + d.FooterClose
			}
		} else {
			content = theme.BoldStyle.Render("[Esc / Q]") + " " + d.FooterClose + "  |  " +
				theme.BoldStyle.Render("[↑ / ↓ / PgUp / PgDn]") + " " + d.FooterScroll
		}
	} else if ctx.ActiveTab == 2 {
		if ctx.IsFiltering {
			content = theme.BoldStyle.Render("[Enter]") + " " + d.FooterApplyFilter + "  |  " +
				theme.BoldStyle.Render("[Esc]") + " " + d.FooterCancel
		} else if ctx.FilterQuery != "" {
			content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelect + "  |  " +
				theme.BoldStyle.Render("[Esc]") + " " + d.FooterResetFilter + "  |  " +
				theme.BoldStyle.Render("[O]") + " " + d.FooterSort + "  |  " +
				theme.BoldStyle.Render("[L]") + " " + d.FooterLogs + "  |  " +
				theme.BoldStyle.Render("[Enter/I]") + " " + d.FooterInfo + "  |  " +
				theme.BoldStyle.Render("[E]") + " " + d.FooterShell + "  |  " +
				theme.BoldStyle.Render("[S]") + " " + d.FooterStartStop + "  |  " +
				theme.BoldStyle.Render("[P]") + " " + d.FooterRestart
		} else {
			content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelect + "  |  " +
				theme.BoldStyle.Render("[/]") + " " + d.FooterSearch + "  |  " +
				theme.BoldStyle.Render("[O]") + " " + d.FooterSort + "  |  " +
				theme.BoldStyle.Render("[L]") + " " + d.FooterLogs + "  |  " +
				theme.BoldStyle.Render("[Enter/I]") + " " + d.FooterInfo + "  |  " +
				theme.BoldStyle.Render("[E]") + " " + d.FooterShell + "  |  " +
				theme.BoldStyle.Render("[S]") + " " + d.FooterStartStop + "  |  " +
				theme.BoldStyle.Render("[P]") + " " + d.FooterRestart + "  |  " +
				theme.BoldStyle.Render("[C]") + " " + d.FooterPrune + "  |  " +
				theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
		}
	} else if ctx.ActiveTab == 3 {
		content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelectStack + "  |  " +
			theme.BoldStyle.Render("[L]") + " " + d.FooterLogs + "  |  " +
			theme.BoldStyle.Render("[S]") + " " + d.FooterStop + "  |  " +
			theme.BoldStyle.Render("[U]") + " " + d.FooterStart + "  |  " +
			theme.BoldStyle.Render("[P]") + " " + d.FooterRestart + "  |  " +
			theme.BoldStyle.Render("[D]") + " " + d.FooterDown + "  |  " +
			theme.BoldStyle.Render("[C]") + " " + d.FooterPrune + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
	} else if ctx.ActiveTab == 4 {
		content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelectService + "  |  " +
			theme.BoldStyle.Render("[L]") + " " + d.FooterServiceLogs + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
	} else {
		content = theme.BoldStyle.Render("[1-4 / Tab / ← →]") + " " + d.FooterTabs + "  |  " +
			theme.BoldStyle.Render("[R]") + " " + d.FooterRefresh + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp + "  |  " +
			theme.BoldStyle.Render("[Q]") + " " + d.FooterQuit
	}

	return "\n" + theme.DimStyle.Render(strings.Repeat("─", max(ctx.Width, 70))) + "\n" +
		theme.FooterStyle.Render(content)
}
