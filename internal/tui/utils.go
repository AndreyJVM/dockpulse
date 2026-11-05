package tui

import (
	"fmt"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	"github.com/AndreyJVM/dockpulse/internal/tui/components"
)

var (
	sparkChars      = components.SparkChars
	formatPercent   = components.FormatPercent
	formatBar       = components.FormatBar
	truncate        = components.Truncate
	padRight        = components.PadRight
	padLeft         = components.PadLeft
	renderSparkline = components.RenderSparkline
)

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
