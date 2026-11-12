package views

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// HeaderContext contains the application and Docker status shown above every tab.
type HeaderContext struct {
	Version               string
	LastUpdate            string
	DockerInfo            monitor.DockerSummary
	FormattedDockerStatus string
	I18n                  *i18n.Dictionary
}

// RenderHeader renders the application banner and persistent Docker status.
func RenderHeader(ctx HeaderContext) string {
	verStr := ctx.Version
	if verStr == "" {
		verStr = "v0.4.0"
	}
	d := ctx.I18n
	if d == nil {
		dict := i18n.Get()
		d = &dict
	}
	appLine := theme.HeaderStyle.Render("Dockpulse") +
		theme.DimStyle.Render(fmt.Sprintf("  %s  ", verStr)) +
		theme.GreenStyle.Render("["+ctx.LastUpdate+"]")
	statusLine := fmt.Sprintf("%s %s  |  %s  |  %s %d  |  %s %s",
		theme.DimStyle.Render(d.ServiceLabel), ctx.FormattedDockerStatus,
		theme.BoldStyle.Render(fmt.Sprintf(d.ContainersSummary, ctx.DockerInfo.RunningContainers, ctx.DockerInfo.StoppedContainers)),
		theme.DimStyle.Render(d.ImagesCountLabel), ctx.DockerInfo.ImagesCount,
		theme.DimStyle.Render(d.VersionLabel), ctx.DockerInfo.Version,
	)
	return appLine + "\n" + statusLine
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
	ShellActive  bool
	ShellFocused bool
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
		} else if ctx.ActiveModal == "remove_container" {
			content = theme.RedStyle.Render("[Y]") + " " + d.FooterConfirm + " (Remove)  |  " +
				theme.BoldStyle.Render("[N / Esc]") + " " + d.FooterCancel
		} else if ctx.ActiveModal == "exec" {
			content = theme.BoldStyle.Render("[1-7 / Enter]") + " " + d.FooterSelectAction + "  |  " +
				theme.BoldStyle.Render("[Esc / Q]") + " " + d.FooterCancel
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
	} else if ctx.ShellActive {
		if ctx.ShellFocused {
			content = theme.BoldStyle.Render("[Ctrl+]]") + " " + d.FooterShellReleaseFocus + "  |  " +
				theme.BoldStyle.Render("[Ctrl+D]") + " " + d.FooterShellExit
		} else {
			content = theme.BoldStyle.Render("[E]") + " " + d.FooterShellFocus + "  |  " +
				theme.BoldStyle.Render("[Ctrl+X]") + " " + d.FooterShellClose + "  |  " +
				theme.BoldStyle.Render("[Q]") + " " + d.FooterQuit
		}
	} else if ctx.ActiveTab == 1 {
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
				theme.BoldStyle.Render("[P]") + " " + d.FooterRestart + "  |  " +
				theme.BoldStyle.Render("[K]") + " " + d.FooterKill + "  |  " +
				theme.BoldStyle.Render("[x]") + " " + d.FooterRemove
		} else {
			content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelect + "  |  " +
				theme.BoldStyle.Render("[/]") + " " + d.FooterSearch + "  |  " +
				theme.BoldStyle.Render("[O]") + " " + d.FooterSort + "  |  " +
				theme.BoldStyle.Render("[L]") + " " + d.FooterLogs + "  |  " +
				theme.BoldStyle.Render("[Enter/I]") + " " + d.FooterInfo + "  |  " +
				theme.BoldStyle.Render("[E]") + " " + d.FooterShell + "  |  " +
				theme.BoldStyle.Render("[S]") + " " + d.FooterStartStop + "  |  " +
				theme.BoldStyle.Render("[P]") + " " + d.FooterRestart + "  |  " +
				theme.BoldStyle.Render("[K]") + " " + d.FooterKill + "  |  " +
				theme.BoldStyle.Render("[x]") + " " + d.FooterRemove + "  |  " +
				theme.BoldStyle.Render("[C]") + " " + d.FooterPrune + "  |  " +
				theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
		}
	} else if ctx.ActiveTab == 2 {
		content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelectStack + "  |  " +
			theme.BoldStyle.Render("[L]") + " " + d.FooterLogs + "  |  " +
			theme.BoldStyle.Render("[S]") + " " + d.FooterStop + "  |  " +
			theme.BoldStyle.Render("[U]") + " " + d.FooterStart + "  |  " +
			theme.BoldStyle.Render("[P]") + " " + d.FooterRestart + "  |  " +
			theme.BoldStyle.Render("[D]") + " " + d.FooterDown + "  |  " +
			theme.BoldStyle.Render("[C]") + " " + d.FooterPrune + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
	} else if ctx.ActiveTab == 3 || ctx.ActiveTab == 4 {
		content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelect + "  |  " +
			theme.BoldStyle.Render("[R]") + " " + d.FooterRefresh + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
	} else if ctx.ActiveTab == 5 {
		content = theme.BoldStyle.Render("[↑ / ↓]") + " " + d.FooterSelectService + "  |  " +
			theme.BoldStyle.Render("[L]") + " " + d.FooterServiceLogs + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp
	} else {
		content = theme.BoldStyle.Render("[1-5 / Tab / ← →]") + " " + d.FooterTabs + "  |  " +
			theme.BoldStyle.Render("[R]") + " " + d.FooterRefresh + "  |  " +
			theme.BoldStyle.Render("[?]") + " " + d.FooterHelp + "  |  " +
			theme.BoldStyle.Render("[Q]") + " " + d.FooterQuit
	}

	return "\n" + theme.DimStyle.Render(strings.Repeat("─", max(ctx.Width, 70))) + "\n" +
		theme.FooterStyle.Render(content)
}
