package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
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
		expectedTab TabID
	}{
		{"tab", TabDocker},
		{"tab", TabCompose},
		{"right", TabSwarm},
		{"tab", TabHost},
		{"left", TabSwarm},
		{"shift+tab", TabCompose},
		{"1", TabHost},
		{"2", TabDocker},
		{"3", TabCompose},
		{"4", TabSwarm},
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
	if !strings.Contains(v1, "SYSTEM INFORMATION") {
		t.Errorf("expected Host tab header in View()")
	}

	// Render Tab 2: Docker
	m.activeTab = 2
	v2 := m.View()
	if !strings.Contains(v2, "DOCKER CONTAINERS") {
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
	if !strings.Contains(vHelp, "HOTKEYS") {
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

func TestNavigationMethods(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.7.4")

	// nextTab wrapping
	m.setTab(TabSwarm)
	m.nextTab()
	if m.activeTab != TabHost {
		t.Errorf("expected wrap to TabHost, got %v", m.activeTab)
	}

	// prevTab wrapping
	m.prevTab()
	if m.activeTab != TabSwarm {
		t.Errorf("expected wrap to TabSwarm, got %v", m.activeTab)
	}

	// setTab valid and invalid
	m.setTab(TabDocker)
	if m.activeTab != TabDocker {
		t.Errorf("expected TabDocker, got %v", m.activeTab)
	}
	m.setTab(99) // Invalid, should remain unchanged
	if m.activeTab != TabDocker {
		t.Errorf("expected TabDocker after invalid tab, got %v", m.activeTab)
	}

	// moveCursor in Docker tab
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "app-1"},
		{ID: "c2", Name: "app-2"},
		{ID: "c3", Name: "app-3"},
	}
	m.cursorIdx = 0
	m.moveCursor(1)
	if m.cursorIdx != 1 {
		t.Errorf("expected cursorIdx 1, got %d", m.cursorIdx)
	}
	m.moveCursor(-1)
	if m.cursorIdx != 0 {
		t.Errorf("expected cursorIdx 0, got %d", m.cursorIdx)
	}
	m.moveCursor(-1) // Wrap to end
	if m.cursorIdx != 2 {
		t.Errorf("expected cursorIdx wrapped to 2, got %d", m.cursorIdx)
	}
	m.moveCursor(1) // Wrap back to 0
	if m.cursorIdx != 0 {
		t.Errorf("expected cursorIdx wrapped to 0, got %d", m.cursorIdx)
	}

	// moveCursor in Compose tab
	m.setTab(TabCompose)
	m.compose = []monitor.ComposeProject{
		{Name: "p1"},
		{Name: "p2"},
	}
	m.composeCursorIdx = 0
	m.moveCursor(1)
	if m.composeCursorIdx != 1 {
		t.Errorf("expected composeCursorIdx 1, got %d", m.composeCursorIdx)
	}
	m.moveCursor(1) // Wrap
	if m.composeCursorIdx != 0 {
		t.Errorf("expected composeCursorIdx 0, got %d", m.composeCursorIdx)
	}

	// moveCursor in Swarm tab
	m.setTab(TabSwarm)
	m.swarm.Services = []monitor.SwarmService{
		{ID: "s1", Name: "svc-1"},
		{ID: "s2", Name: "svc-2"},
	}
	m.swarmCursorIdx = 0
	m.moveCursor(1)
	if m.swarmCursorIdx != 1 {
		t.Errorf("expected swarmCursorIdx 1, got %d", m.swarmCursorIdx)
	}
	m.moveCursor(1) // Wrap
	if m.swarmCursorIdx != 0 {
		t.Errorf("expected swarmCursorIdx 0, got %d", m.swarmCursorIdx)
	}
}

func TestModalLifecycle(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.7.4")
	if m.activeModal != ModalNone {
		t.Errorf("expected initial activeModal to be ModalNone, got %v", m.activeModal)
	}

	m.openModal(ModalExec)
	m.execIsCustomInput = true
	m.execCustomCmd = "top"
	if m.activeModal != ModalExec {
		t.Errorf("expected ModalExec, got %v", m.activeModal)
	}

	m.closeModal()
	if m.activeModal != ModalNone {
		t.Errorf("expected ModalNone after closeModal, got %v", m.activeModal)
	}
	if m.execIsCustomInput {
		t.Errorf("expected execIsCustomInput to be reset to false")
	}
	if m.execCustomCmd != "" {
		t.Errorf("expected execCustomCmd to be reset to empty")
	}
}

func TestContainerStatusScore(t *testing.T) {
	tests := []struct {
		state  string
		health string
		score  int
	}{
		{"running", "unhealthy", 1},
		{"running", "starting", 2},
		{"running", "healthy", 3},
		{"running", "", 3},
		{"restarting", "", 4},
		{"paused", "", 5},
		{"exited", "", 6},
		{"unknown", "", 6},
	}

	for _, tt := range tests {
		s := containerStatusScore(monitor.ContainerItem{
			State:  tt.state,
			Health: tt.health,
		})
		if s != tt.score {
			t.Errorf("for state=%q health=%q expected score %d, got %d", tt.state, tt.health, tt.score, s)
		}
	}
}
