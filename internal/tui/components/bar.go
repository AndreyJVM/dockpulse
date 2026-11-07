package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Standard percentage boundaries and visual thresholds.
const (
	MinPercent           = 0.0
	MaxPercent           = 100.0
	ThresholdWarnPercent = 60.0
	ThresholdHighPercent = 80.0

	barBlockFilled = "█"
	barBlockEmpty  = "░"
)

// Clamp restricts a float64 value to the range [minVal, maxVal].
func Clamp(val, minVal, maxVal float64) float64 {
	if math.IsNaN(val) {
		return minVal
	}
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// PercentStyle returns the lipgloss style corresponding to the metric percentage.
// RedStyle for >= ThresholdHighPercent (80%),
// YellowStyle for >= ThresholdWarnPercent (60%),
// GreenStyle otherwise.
func PercentStyle(val float64) lipgloss.Style {
	if val >= ThresholdHighPercent {
		return theme.RedStyle
	}
	if val >= ThresholdWarnPercent {
		return theme.YellowStyle
	}
	return theme.GreenStyle
}

// FormatPercent formats a percentage value with color grading.
func FormatPercent(val float64) string {
	str := fmt.Sprintf("%.1f%%", val)
	return PercentStyle(val).Render(str)
}

// RenderBar formats a horizontal progress bar with optional percentage label.
func RenderBar(percent float64, width int, showPercent bool) string {
	clampedPct := Clamp(percent, MinPercent, MaxPercent)

	if width <= 0 {
		if showPercent {
			return fmt.Sprintf("[%s] %5.1f%%", "", clampedPct)
		}
		return "[]"
	}

	filled := int((clampedPct / MaxPercent) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	if empty < 0 {
		empty = 0
	}

	barFilled := strings.Repeat(barBlockFilled, filled)
	barEmpty := strings.Repeat(barBlockEmpty, empty)

	style := PercentStyle(clampedPct)
	coloredFilled := style.Render(barFilled)
	dimmedEmpty := theme.DimStyle.Render(barEmpty)

	if !showPercent {
		return fmt.Sprintf("[%s%s]", coloredFilled, dimmedEmpty)
	}

	pctStr := fmt.Sprintf("%5.1f%%", clampedPct)
	return fmt.Sprintf("[%s%s] %s", coloredFilled, dimmedEmpty, pctStr)
}

// FormatBar formats a horizontal progress bar with percentage.
func FormatBar(percent float64, width int) string {
	return RenderBar(percent, width, true)
}
