package tui

import (
	"github.com/AndreyJVM/dockpulse/internal/monitor"
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
		case monitor.DockerStatusDisplayRunning, "running":
			return m.i18n.StatusRunning
		case monitor.DockerStatusDisplayStopped, "stopped":
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
		case monitor.StatusDisplayRunning, "running":
			return m.i18n.StatusRunning
		case monitor.StatusDisplayPartial, "partial":
			return m.i18n.StatusPartial
		case monitor.StatusDisplayStopped, "stopped":
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
