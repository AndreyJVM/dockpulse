package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func TestEnglishLanguageModel(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.6.1", "en")
	if m.lang != "en" {
		t.Fatalf("expected lang='en', got %q", m.lang)
	}

	// 1. Host Tab
	m.activeTab = 1
	m.dockerInfo = monitor.DockerSummary{
		Status:        "running",
		StatusDisplay: "Запущен", // legacy monitor status display
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
	if strings.Contains(vHost, "Запущен") {
		t.Errorf("found unwanted Russian 'Запущен' in English Host Tab: %s", vHost)
	}
	if !strings.Contains(vHost, "1d 01:00") {
		t.Errorf("expected English uptime format '1d 01:00', got: %s", vHost)
	}
	if strings.Contains(vHost, "дн") {
		t.Errorf("found unwanted Russian 'дн' in English Host Tab: %s", vHost)
	}

	// 2. Docker Tab
	m.activeTab = 2
	vDocker := m.View()
	if !strings.Contains(vDocker, "DOCKER CONTAINERS") {
		t.Errorf("expected 'DOCKER CONTAINERS' in View(), got: %s", vDocker)
	}
	if !strings.Contains(vDocker, "Daemon: Running") {
		t.Errorf("expected English 'Daemon: Running', got: %s", vDocker)
	}
	if strings.Contains(vDocker, "Запущен") {
		t.Errorf("found unwanted Russian 'Запущен' in English Docker Tab: %s", vDocker)
	}

	// 3. Compose Tab
	m.activeTab = 3
	m.compose = []monitor.ComposeProject{
		{
			Name:            "demo-stack",
			Status:          "running",
			StatusDisplay:   "Запущен",
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
	if strings.Contains(vCompose, "[Запущен]") {
		t.Errorf("found unwanted Russian '[Запущен]' in English Compose Tab: %s", vCompose)
	}

	// 4. Swarm Tab
	m.activeTab = 4
	m.swarm = monitor.SwarmSummary{
		IsActive:      false,
		StatusDisplay: "Не активен",
	}
	vSwarm := m.View()
	if !strings.Contains(vSwarm, "DOCKER SWARM CLUSTER") {
		t.Errorf("expected 'DOCKER SWARM CLUSTER' in View(), got: %s", vSwarm)
	}
	if !strings.Contains(vSwarm, "Status: Inactive") {
		t.Errorf("expected English 'Status: Inactive', got: %s", vSwarm)
	}
	if strings.Contains(vSwarm, "Не активен") {
		t.Errorf("found unwanted Russian 'Не активен' in English Swarm Tab: %s", vSwarm)
	}

	// 5. Help Modal
	m.showHelp = true
	vHelp := m.View()
	if !strings.Contains(vHelp, "HOTKEYS:") {
		t.Errorf("expected 'HOTKEYS:' in View(), got: %s", vHelp)
	}
	m.showHelp = false

	// 6. Compose down modal
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
