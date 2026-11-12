package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Constants for log modal dimensions and constraints.
const (
	DefaultMaxVisibleLogs = 18
	MinLogLineWidth       = 40
	LogModalBorderMargin  = 8
)

// RenderLogs renders the interactive container/service log viewer modal.
func RenderLogs(lines []string, scroll int, targetName, targetID string, width int, dict *i18n.Dictionary) string {
	var s strings.Builder
	d := getModalDictionary(dict)

	title := fmt.Sprintf(d.LogsTitle, theme.BoldStyle.Render(targetName), theme.DimStyle.Render(targetID))
	s.WriteString(theme.TitleStyle.Render(title) + "\n\n")

	if len(lines) == 0 {
		s.WriteString("  " + theme.DimStyle.Render(d.NoLogsOutput) + "\n")
		return theme.BoxStyle.Render(s.String())
	}

	start, end := calculateLogRange(len(lines), scroll, DefaultMaxVisibleLogs)
	maxLineWidth := max(MinLogLineWidth, width-LogModalBorderMargin)

	for i := start; i < end; i++ {
		line := lines[i]
		gutter := getLogGutterStyle(line).Render("│")
		s.WriteString(fmt.Sprintf(" %s %s\n", gutter, components.Truncate(line, maxLineWidth)))
	}

	s.WriteString(fmt.Sprintf("\n  %s\n", theme.DimStyle.Render(fmt.Sprintf(d.LogsRangeCount, start+1, end, len(lines)))))

	return theme.BoxStyle.Render(s.String())
}

func getModalDictionary(dict *i18n.Dictionary) *i18n.Dictionary {
	if dict != nil {
		return dict
	}
	d := i18n.Get()
	return &d
}

func calculateLogRange(totalLines, scroll, maxVisible int) (int, int) {
	if totalLines <= 0 {
		return 0, 0
	}
	maxScroll := max(0, totalLines-1)
	start := scroll
	if start < 0 {
		start = 0
	}
	if start > maxScroll {
		start = maxScroll
	}
	end := min(start+maxVisible, totalLines)
	return start, end
}

func getLogGutterStyle(line string) lipgloss.Style {
	upper := strings.ToUpper(line)
	if strings.Contains(upper, "ERROR") || strings.Contains(upper, "FATAL") || strings.Contains(upper, "PANIC") {
		return theme.RedStyle
	}
	if strings.Contains(upper, "WARN") {
		return theme.YellowStyle
	}
	return theme.DimStyle
}
