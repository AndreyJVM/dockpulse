package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
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

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return s + strings.Repeat(" ", width-len(r))
}

func padLeft(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return string(r[:width])
	}
	return strings.Repeat(" ", width-len(r)) + s
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

func (m Model) formatDockerStatus(status string) string {
	switch status {
	case monitor.DockerStatusRunning:
		return m.i18n.StatusRunning
	case monitor.DockerStatusStopped:
		return m.i18n.StatusStopped
	case monitor.DockerStatusNotInstalled:
		return m.i18n.StatusNotInstalled
	default:
		switch m.dockerInfo.StatusDisplay {
		case monitor.DockerStatusDisplayRunning, "[Запущен]", "running", "Running":
			return m.i18n.StatusRunning
		case monitor.DockerStatusDisplayStopped, "stopped", "Stopped":
			return m.i18n.StatusStopped
		case monitor.DockerStatusDisplayNotInstalled, "not_installed":
			return m.i18n.StatusNotInstalled
		default:
			if m.dockerInfo.StatusDisplay != "" {
				return m.dockerInfo.StatusDisplay
			}
			if status == "" {
				return m.i18n.StatusStopped
			}
			return status
		}
	}
}

func (m Model) formatComposeStatus(status, statusDisplay string) string {
	switch status {
	case monitor.StatusRunning:
		return m.i18n.StatusRunning
	case monitor.StatusPartial:
		return m.i18n.StatusPartial
	case monitor.StatusStopped:
		return m.i18n.StatusStopped
	default:
		switch statusDisplay {
		case monitor.StatusDisplayRunning, "[Запущен]", "running", "Running":
			return m.i18n.StatusRunning
		case monitor.StatusDisplayPartial, "partial", "Partial":
			return m.i18n.StatusPartial
		case monitor.StatusDisplayStopped, "stopped", "Stopped":
			return m.i18n.StatusStopped
		default:
			if statusDisplay != "" {
				return statusDisplay
			}
			return status
		}
	}
}

func (m Model) formatSwarmStatus(isActive bool) string {
	if isActive {
		return m.i18n.StatusActive
	}
	return m.i18n.StatusInactive
}

func (m Model) formatUptime(uptimeSec uint64) string {
	if uptimeSec == 0 {
		if m.sysInfo.Uptime != "" {
			return m.sysInfo.Uptime
		}
		return "00:00"
	}
	days := uptimeSec / 86400
	hours := (uptimeSec % 86400) / 3600
	minutes := (uptimeSec % 3600) / 60
	if days > 0 {
		return fmt.Sprintf(m.i18n.UptimeDaysFormat, days, hours, minutes)
	}
	return fmt.Sprintf(m.i18n.UptimeHoursFormat, hours, minutes)
}
