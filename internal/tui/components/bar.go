package components

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// FormatPercent formats a percentage value with color grading.
func FormatPercent(val float64) string {
	str := fmt.Sprintf("%.1f%%", val)
	if val >= 80 {
		return theme.RedStyle.Render(str)
	} else if val >= 50 {
		return theme.YellowStyle.Render(str)
	}
	return theme.GreenStyle.Render(str)
}

// FormatBar formats a horizontal progress bar with percentage.
func FormatBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	barFilled := strings.Repeat("█", filled)
	barEmpty := strings.Repeat("░", empty)

	var coloredFilled string
	if percent >= 85 {
		coloredFilled = theme.RedStyle.Render(barFilled)
	} else if percent >= 60 {
		coloredFilled = theme.YellowStyle.Render(barFilled)
	} else {
		coloredFilled = theme.GreenStyle.Render(barFilled)
	}

	pctStr := fmt.Sprintf("%5.1f%%", percent)
	return fmt.Sprintf("[%s%s] %s", coloredFilled, theme.DimStyle.Render(barEmpty), pctStr)
}
