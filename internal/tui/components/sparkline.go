package components

import (
	"strings"
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
		v := Clamp(val, MinPercent, MaxPercent)

		idx := int((v / MaxPercent) * float64(len(SparkChars)))
		if idx >= len(SparkChars) {
			idx = len(SparkChars) - 1
		}
		ch := string(SparkChars[idx])
		sb.WriteString(PercentStyle(v).Render(ch))
	}
	return sb.String()
}
