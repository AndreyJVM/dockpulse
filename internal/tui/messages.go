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

type embeddedShellOutputMsg struct {
	session embeddedShellSession
	data    []byte
	err     error
}

type embeddedShellExitedMsg struct {
	session embeddedShellSession
	err     error
}

// dataFetchedMsg delivers freshly gathered metrics and container lists across all domains.
type dataFetchedMsg struct {
	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	networks   []monitor.DockerNetwork
	volumes    []monitor.DockerVolume
	swarm      monitor.SwarmSummary
	networkErr string
	volumeErr  string
	timestamp  string
}

// actionFinishedMsg is returned upon completion of an async container or compose operation.
type actionFinishedMsg struct {
	message string
	err     error
	refresh bool
}
