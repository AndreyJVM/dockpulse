package components

import (
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// SparkChars are the Unicode block characters used for sparklines.
var SparkChars = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline renders a mini graphical sparkline for historical metric slices.
func RenderSparkline(history []float64) string {
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

		idx := int((v / 100.0) * float64(len(SparkChars)))
		if idx >= len(SparkChars) {
			idx = len(SparkChars) - 1
		}
		ch := string(SparkChars[idx])

		if v >= 80 {
			sb.WriteString(theme.RedStyle.Render(ch))
		} else if v >= 60 {
			sb.WriteString(theme.YellowStyle.Render(ch))
		} else {
			sb.WriteString(theme.GreenStyle.Render(ch))
		}
	}
	return sb.String()
}
