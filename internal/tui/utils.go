package tui

import (
	"fmt"
	"strings"
)

var sparkChars = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func formatPercent(val float64) string {
	str := fmt.Sprintf("%.1f%%", val)
	if val >= 80 {
		return redStyle.Render(str)
	} else if val >= 50 {
		return yellowStyle.Render(str)
	}
	return greenStyle.Render(str)
}

func formatBar(percent float64, width int) string {
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
		coloredFilled = redStyle.Render(barFilled)
	} else if percent >= 60 {
		coloredFilled = yellowStyle.Render(barFilled)
	} else {
		coloredFilled = greenStyle.Render(barFilled)
	}

	pctStr := fmt.Sprintf("%5.1f%%", percent)
	return fmt.Sprintf("[%s%s] %s", coloredFilled, dimStyle.Render(barEmpty), pctStr)
}

func truncate(str string, maxLen int) string {
	runes := []rune(str)
	if len(runes) <= maxLen {
		return str
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderSparkline(history []float64) string {
	if len(history) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, val := range history {
		v := val
		if v < 0 {
			v = 0
		} else if v > 100 {
			v = 100
		}

		idx := int((v / 100.0) * float64(len(sparkChars)))
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		ch := string(sparkChars[idx])

		if v >= 80 {
			sb.WriteString(redStyle.Render(ch))
		} else if v >= 60 {
			sb.WriteString(yellowStyle.Render(ch))
		} else {
			sb.WriteString(greenStyle.Render(ch))
		}
	}
	return sb.String()
}
