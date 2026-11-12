package views

import (
	"fmt"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// EmbeddedShellContext contains the terminal pane state.
type EmbeddedShellContext struct {
	ContainerName string
	Output        string
	Width         int
	Height        int
	Focused       bool
	I18n          *i18n.Dictionary
}

// RenderEmbeddedShell renders an interactive terminal pane inside Dockpulse.
func RenderEmbeddedShell(ctx EmbeddedShellContext) string {
	d := getDictionary(ctx.I18n)
	width := max(20, ctx.Width)
	height := max(6, ctx.Height)
	innerWidth := max(1, width-2)
	innerHeight := max(1, height-2)

	status := d.EmbeddedShellUnfocused
	statusStyle := theme.DimStyle
	borderColor := lipgloss.Color("#6B7280")
	if ctx.Focused {
		status = d.EmbeddedShellFocused
		statusStyle = theme.GreenStyle
		borderColor = lipgloss.Color("#22C55E")
	}

	header := theme.BoldStyle.Render(fmt.Sprintf(d.EmbeddedShellTitle, ctx.ContainerName)) +
		"  " + statusStyle.Render("["+status+"]")
	output := ctx.Output
	if output == "" {
		output = theme.DimStyle.Render(d.EmbeddedShellStarting)
	}

	content := lipgloss.NewStyle().
		Width(innerWidth).
		Height(max(1, innerHeight-1)).
		MaxWidth(innerWidth).
		MaxHeight(max(1, innerHeight-1)).
		Render(header + "\n" + output)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight).
		MaxWidth(width).
		MaxHeight(height).
		Render(content)
}
