package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// ExecModalContext holds information required to render the interactive exec/attach modal.
type ExecModalContext struct {
	ContainerName string
	ContainerID   string
	CursorIdx     int
	IsCustomInput bool
	CustomCmd     string
	I18n          *i18n.Dictionary
}

// ExecMenuItem represents a selectable execution target or shell.
type ExecMenuItem struct {
	Key         string
	Title       string
	Description string
}

// GetExecMenuItems returns the list of available shell/exec modes.
func GetExecMenuItems(d *i18n.Dictionary) []ExecMenuItem {
	return []ExecMenuItem{
		{
			Key:         "1",
			Title:       "Auto Bash / Sh",
			Description: d.ExecModeAutoBash,
		},
		{
			Key:         "2",
			Title:       "POSIX Shell (/bin/sh)",
			Description: d.ExecModeSh,
		},
		{
			Key:         "3",
			Title:       "Zsh Shell (/bin/zsh)",
			Description: d.ExecModeZsh,
		},
		{
			Key:         "4",
			Title:       "Docker Attach",
			Description: d.ExecModeAttach,
		},
		{
			Key:         "5",
			Title:       "Root Shell (-u 0)",
			Description: d.ExecModeRootBash,
		},
		{
			Key:         "6",
			Title:       "Custom Command...",
			Description: d.ExecModeCustom,
		},
	}
}

// RenderExec renders the interactive exec / attach selection modal.
func RenderExec(ctx ExecModalContext) string {
	d := getModalDictionary(ctx.I18n)
	var s strings.Builder

	s.WriteString(theme.TitleStyle.Render(d.ExecModalTitle) + "\n\n")
	s.WriteString(fmt.Sprintf("  "+d.ContainerLabel+" (%s)\n\n",
		theme.BoldStyle.Render(ctx.ContainerName),
		theme.DimStyle.Render(ctx.ContainerID),
	))

	if ctx.IsCustomInput {
		s.WriteString(fmt.Sprintf("  %s\n", theme.YellowStyle.Render(fmt.Sprintf(d.ExecCustomPrompt, ctx.CustomCmd+"█"))))
		s.WriteString(fmt.Sprintf("  %s\n\n", theme.DimStyle.Render(d.ExecCustomHint)))
		return theme.BoxStyle.Render(s.String())
	}

	items := GetExecMenuItems(d)
	for i, item := range items {
		cursor := "  "
		itemStyle := theme.BoldStyle
		if i == ctx.CursorIdx {
			cursor = theme.GreenStyle.Render("▸ ")
			itemStyle = theme.ActiveTabStyle
		}

		shortcut := theme.BoldStyle.Render(fmt.Sprintf("[%s]", item.Key))
		s.WriteString(fmt.Sprintf("%s%s %-24s %s\n",
			cursor,
			shortcut,
			itemStyle.Render(item.Title),
			theme.DimStyle.Render(item.Description),
		))
	}

	s.WriteString("\n  " +
		theme.BoldStyle.Render("[1-6 / Enter]") + " " + d.FooterSelectAction + "  |  " +
		theme.BoldStyle.Render("[Esc / Q]") + " " + d.FooterCancel + "\n")

	return theme.BoxStyle.Render(s.String())
}
