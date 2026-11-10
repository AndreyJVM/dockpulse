package tui

import (
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

// tickMsg signals a periodic refresh timer tick.
type tickMsg time.Time

// execFinishedMsg is returned when an interactive container exec/attach session ends.
type execFinishedMsg struct {
	containerName string
	err           error
}

// dataFetchedMsg delivers freshly gathered metrics and container lists across all domains.
type dataFetchedMsg struct {
	sysInfo    monitor.SystemInfo
	cpuInfo    monitor.CPUInfo
	memInfo    monitor.MemoryInfo
	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	swarm      monitor.SwarmSummary
	timestamp  string
}

// actionFinishedMsg is returned upon completion of an async container or compose operation.
type actionFinishedMsg struct {
	message string
	err     error
	refresh bool
}
