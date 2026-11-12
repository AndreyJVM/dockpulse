package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func TestEnglishLanguageModel(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.6.1")

	// Host tab.
	m.activeTab = 1
	m.dockerInfo = monitor.DockerSummary{
		Status:        "running",
		StatusDisplay: "Running",
	}
	m.sysInfo = monitor.SystemInfo{
		Hostname:  "test-host",
		UptimeSec: 90000,
	}
	vHost := m.View()
	if !strings.Contains(vHost, "SYSTEM INFORMATION") {
		t.Errorf("expected English title 'SYSTEM INFORMATION' in View(), got: %s", vHost)
	}
	if !strings.Contains(vHost, "HOST PERFORMANCE") {
		t.Errorf("expected 'HOST PERFORMANCE' in View(), got: %s", vHost)
	}
	if !strings.Contains(vHost, "Tabs") {
		t.Errorf("expected 'Tabs' in footer, got: %s", vHost)
	}
	if !strings.Contains(vHost, "Status: Running") {
		t.Errorf("expected English 'Status: Running', got: %s", vHost)
	}
	if !strings.Contains(vHost, "1d 01:00") {
		t.Errorf("expected English uptime format '1d 01:00', got: %s", vHost)
	}

	// Docker tab.
	m.activeTab = 2
	vDocker := m.View()
	if !strings.Contains(vDocker, "DOCKER CONTAINERS") {
		t.Errorf("expected 'DOCKER CONTAINERS' in View(), got: %s", vDocker)
	}
	if !strings.Contains(vDocker, "Daemon: Running") {
		t.Errorf("expected English 'Daemon: Running', got: %s", vDocker)
	}

	// Compose tab.
	m.activeTab = 3
	m.compose = []monitor.ComposeProject{
		{
			Name:            "demo-stack",
			Status:          "running",
			StatusDisplay:   "Running",
			TotalServices:   1,
			RunningServices: 1,
		},
	}
	vCompose := m.View()
	if !strings.Contains(vCompose, "DOCKER COMPOSE PROJECTS") {
		t.Errorf("expected 'DOCKER COMPOSE PROJECTS' in View(), got: %s", vCompose)
	}
	if !strings.Contains(vCompose, "[Running]") {
		t.Errorf("expected English '[Running]' status in Compose tab, got: %s", vCompose)
	}

	// Swarm tab.
	m.activeTab = 4
	m.swarm = monitor.SwarmSummary{
		IsActive:      false,
		StatusDisplay: "Inactive",
	}
	vSwarm := m.View()
	if !strings.Contains(vSwarm, "DOCKER SWARM CLUSTER") {
		t.Errorf("expected 'DOCKER SWARM CLUSTER' in View(), got: %s", vSwarm)
	}
	if !strings.Contains(vSwarm, "Status: Inactive") {
		t.Errorf("expected English 'Status: Inactive', got: %s", vSwarm)
	}

	// Help modal.
	m.showHelp = true
	vHelp := m.View()
	if !strings.Contains(vHelp, "HOTKEYS:") {
		t.Errorf("expected 'HOTKEYS:' in View(), got: %s", vHelp)
	}
	m.showHelp = false

	// Compose down modal.
	m.composeDownTarget = "my-stack"
	m.activeModal = "compose_down"
	vDown := m.View()
	if !strings.Contains(vDown, "CONFIRM DOCKER COMPOSE DOWN") {
		t.Errorf("expected 'CONFIRM DOCKER COMPOSE DOWN' in View(), got: %s", vDown)
	}
	if !strings.Contains(vDown, "Confirm") {
		t.Errorf("expected 'Confirm' in View(), got: %s", vDown)
	}
}
