package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	// Defaults when invalid arguments are passed
	m := NewModel(0, 99, "v1.0.0")
	if m.activeTab != 1 {
		t.Errorf("expected activeTab to default to 1, got %d", m.activeTab)
	}
	if m.refreshInterval != 2*time.Second {
		t.Errorf("expected refreshInterval to default to 2s, got %v", m.refreshInterval)
	}
	if m.version != "v1.0.0" {
		t.Errorf("expected version 'v1.0.0', got %q", m.version)
	}

	// Init command
	cmd := m.Init()
	if cmd == nil {
		t.Errorf("expected non-nil Init cmd")
	}
}

func TestModelNavigationAndUpdates(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.4.1")

	// Resize
	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd != nil {
		t.Errorf("expected nil cmd on window resize")
	}
	m = updatedModel.(Model)
	if m.width != 120 || m.height != 40 {
		t.Errorf("expected dimensions 120x40, got %dx%d", m.width, m.height)
	}

	// Tab switching
	keyTests := []struct {
		key         string
		expectedTab int
	}{
		{"tab", 2},
		{"tab", 3},
		{"right", 4},
		{"tab", 1},
		{"left", 4},
		{"shift+tab", 3},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
	}

	for _, kt := range keyTests {
		res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(kt.key)})
		m = res.(Model)
		if m.activeTab != kt.expectedTab {
			t.Errorf("after key %q: expected tab %d, got %d", kt.key, kt.expectedTab, m.activeTab)
		}
	}

	// Help modal toggle
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = res.(Model)
	if !m.showHelp {
		t.Errorf("expected showHelp=true after '?'")
	}

	// Close help
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)
	if m.showHelp {
		t.Errorf("expected showHelp=false after 'esc'")
	}

	// Quit keys
	_, quitCmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quitCmd == nil {
		t.Errorf("expected quitCmd on ctrl+c")
	}
}

func TestRenderAllTabs(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.4.1")

	// Render Tab 1: Host
	m.activeTab = 1
	v1 := m.View()
	if !strings.Contains(v1, "СИСТЕМНАЯ ИНФОРМАЦИЯ") {
		t.Errorf("expected Host tab header in View()")
	}

	// Render Tab 2: Docker
	m.activeTab = 2
	v2 := m.View()
	if !strings.Contains(v2, "DOCKER КОНТЕЙНЕРЫ") {
		t.Errorf("expected Docker tab header in View()")
	}

	// Render Tab 3: Compose
	m.activeTab = 3
	v3 := m.View()
	if !strings.Contains(v3, "DOCKER COMPOSE") {
		t.Errorf("expected Compose tab header in View()")
	}

	// Render Tab 4: Swarm
	m.activeTab = 4
	v4 := m.View()
	if !strings.Contains(v4, "DOCKER SWARM") {
		t.Errorf("expected Swarm tab header in View()")
	}

	// Render Help modal
	m.showHelp = true
	vHelp := m.View()
	if !strings.Contains(vHelp, "ГОРЯЧИЕ КЛАВИШИ") {
		t.Errorf("expected Help modal in View()")
	}
}

func TestRecordMetricHistory(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")
	initialCPULen := len(m.cpuHistory)
	initialMemLen := len(m.memHistory)

	if initialCPULen == 0 || initialMemLen == 0 {
		t.Errorf("expected initial history to have at least 1 point from fetchAllData")
	}

	// Push 40 entries to verify maxHistory truncation (limit = 30)
	for i := 0; i < 40; i++ {
		m.recordMetricHistory(float64(i), float64(i*2))
	}

	if len(m.cpuHistory) != 30 {
		t.Errorf("expected cpuHistory to be capped at 30, got %d", len(m.cpuHistory))
	}
	if len(m.memHistory) != 30 {
		t.Errorf("expected memHistory to be capped at 30, got %d", len(m.memHistory))
	}

	// Verify latest value is at the end
	if m.cpuHistory[29] != 39.0 {
		t.Errorf("expected latest cpu value 39.0, got %f", m.cpuHistory[29])
	}
	if m.memHistory[29] != 78.0 {
		t.Errorf("expected latest mem value 78.0, got %f", m.memHistory[29])
	}
}

func TestHostTabSparklineDisplay(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")
	m.width = 100
	m.height = 30
	m.cpuHistory = []float64{10, 20, 30, 70, 90}
	m.memHistory = []float64{15, 25, 35, 65, 85}

	view := m.View()
	if !strings.Contains(view, "CPU (") || !strings.Contains(view, "RAM (") {
		t.Errorf("expected CPU and RAM in Host Overview view, got: %s", view)
	}
	var hasSparkChar bool
	for _, ch := range sparkChars {
		if strings.ContainsRune(view, ch) {
			hasSparkChar = true
			break
		}
	}
	if !hasSparkChar {
		t.Errorf("expected sparkline characters in view output, got: %s", view)
	}
}
