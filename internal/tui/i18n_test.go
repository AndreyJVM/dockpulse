package tui

import (
	"strings"
	"testing"
	"time"
)

func TestEnglishLanguageModel(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0", "en")
	if m.lang != "en" {
		t.Fatalf("expected lang='en', got %q", m.lang)
	}

	// 1. Host Tab
	m.activeTab = 1
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

	// 2. Docker Tab
	m.activeTab = 2
	vDocker := m.View()
	if !strings.Contains(vDocker, "DOCKER CONTAINERS") {
		t.Errorf("expected 'DOCKER CONTAINERS' in View(), got: %s", vDocker)
	}

	// 3. Compose Tab
	m.activeTab = 3
	vCompose := m.View()
	if !strings.Contains(vCompose, "DOCKER COMPOSE PROJECTS") {
		t.Errorf("expected 'DOCKER COMPOSE PROJECTS' in View(), got: %s", vCompose)
	}

	// 4. Swarm Tab
	m.activeTab = 4
	vSwarm := m.View()
	if !strings.Contains(vSwarm, "DOCKER SWARM CLUSTER") {
		t.Errorf("expected 'DOCKER SWARM CLUSTER' in View(), got: %s", vSwarm)
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
