package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderLogs renders the interactive container/service log viewer modal.
func RenderLogs(lines []string, scroll int, targetName, targetID string, width int, dict *i18n.Dictionary) string {
	var s strings.Builder
	title := fmt.Sprintf(dict.LogsTitle, theme.BoldStyle.Render(targetName), theme.DimStyle.Render(targetID))
	s.WriteString(theme.TitleStyle.Render(title) + "\n\n")

	maxVisible := 18
	if len(lines) == 0 {
		s.WriteString("  " + theme.DimStyle.Render(dict.NoLogsOutput) + "\n")
	} else {
		start := min(scroll, max(0, len(lines)-1))
		end := min(start+maxVisible, len(lines))

		for i := start; i < end; i++ {
			line := lines[i]
			s.WriteString(fmt.Sprintf(" %s %s\n", theme.DimStyle.Render("│"), components.Truncate(line, max(40, width-8))))
		}

		s.WriteString(fmt.Sprintf("\n  %s\n", theme.DimStyle.Render(fmt.Sprintf(dict.LogsRangeCount, start+1, end, len(lines)))))
	}

	return theme.BoxStyle.Render(s.String())
}
