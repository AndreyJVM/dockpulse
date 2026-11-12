package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInspectModalHealthBadge(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.4.1")
	m.activeModal = "inspect"
	m.inspectTarget = monitor.ContainerDetail{
		ID:     "c1",
		Name:   "app-prod",
		Image:  "my-app:v1",
		Status: "Up 2 hours",
		Health: "healthy",
	}

	view := m.View()
	if !strings.Contains(view, "Health:") || !strings.Contains(view, "[healthy]") {
		t.Errorf("expected health badge in inspect modal view, got: %s", view)
	}

	// Test unhealthy
	m.inspectTarget.Health = "unhealthy"
	view = m.View()
	if !strings.Contains(view, "[unhealthy]") {
		t.Errorf("expected [unhealthy] badge in inspect modal view, got: %s", view)
	}

	// Test starting
	m.inspectTarget.Health = "starting"
	view = m.View()
	if !strings.Contains(view, "[starting]") {
		t.Errorf("expected [starting] badge in inspect modal view, got: %s", view)
	}

	// Test no health configured
	m.inspectTarget.Health = ""
	view = m.View()
	if !strings.Contains(view, "Health:") {
		t.Errorf("expected Health line even if health is empty, got: %s", view)
	}
}

func TestPruneModalAndWorkflow(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")
	m.width = 100
	m.height = 30

	// 1. Open prune modal on tab 2 with "c"
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = res.(Model)
	if m.activeModal != "prune" {
		t.Fatalf("expected activeModal to be 'prune', got %q", m.activeModal)
	}

	// 2. View in menu state
	view := m.View()
	if !strings.Contains(view, "DOCKER RESOURCE PRUNE") {
		t.Errorf("expected prune modal header in view, got: %s", view)
	}
	if !strings.Contains(view, "Prune stopped containers") {
		t.Errorf("expected menu item 1 in view, got: %s", view)
	}
	if !strings.Contains(view, "[1-4]") {
		t.Errorf("expected [1-4] footer hint, got: %s", view)
	}

	// 3. Select option 1 (prune stopped containers)
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	m = res.(Model)
	if m.pruneConfirm != 1 {
		t.Errorf("expected pruneConfirm to be 1, got %d", m.pruneConfirm)
	}
	confirmView := m.View()
	if !strings.Contains(confirmView, "Remove all stopped containers") {
		t.Errorf("expected confirmation prompt in view, got: %s", confirmView)
	}
	if !strings.Contains(confirmView, "[Y]") {
		t.Errorf("expected [Y] confirmation hint, got: %s", confirmView)
	}

	// 4. Cancel confirmation with "n"
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = res.(Model)
	if m.pruneConfirm != 0 {
		t.Errorf("expected pruneConfirm to be 0 after 'n', got %d", m.pruneConfirm)
	}
	if m.activeModal != "prune" {
		t.Errorf("expected activeModal to remain 'prune', got %q", m.activeModal)
	}

	// 5. Select option 4 and confirm with "Y"
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	m = res.(Model)
	if m.pruneConfirm != 4 {
		t.Errorf("expected pruneConfirm to be 4, got %d", m.pruneConfirm)
	}
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to close after confirmation, got %q", m.activeModal)
	}
	if m.pruneConfirm != 0 {
		t.Errorf("expected pruneConfirm to reset to 0, got %d", m.pruneConfirm)
	}

	// 6. Test opening prune modal on Tab 2 (Compose)
	m.activeTab = 2
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})
	m = res.(Model)
	if m.activeModal != "prune" {
		t.Errorf("expected activeModal 'prune' on tab 2 with 'C', got %q", m.activeModal)
	}

	// 7. Close modal with Esc
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to close on Esc, got %q", m.activeModal)
	}

	// 8. Check footer hints on Docker and Compose tabs.
	m.activeTab = 1
	m.activeModal = ""
	dockerView := m.View()
	if !strings.Contains(dockerView, "[C]") || !strings.Contains(dockerView, "Prune") {
		t.Errorf("expected [C] Prune in Docker footer, got: %s", dockerView)
	}

	m.activeTab = 2
	composeView := m.View()
	if !strings.Contains(composeView, "[C]") || !strings.Contains(composeView, "Prune") {
		t.Errorf("expected [C] Prune in Compose footer, got: %s", composeView)
	}

	// 9. Check help box hint
	m.showHelp = true
	helpView := m.View()
	if !strings.Contains(helpView, "[C]") || !strings.Contains(helpView, "Prune unused Docker resources") {
		t.Errorf("expected [C] hint in help view, got: %s", helpView)
	}
}

func TestExecModalWorkflow(t *testing.T) {
	restore := monitor.SetDockerCLIAvailableCheckerForTesting(func() bool { return true })
	defer restore()

	m := NewModel(2*time.Second, 1, "v0.7.4")
	m.dockerInfo = monitor.DockerSummary{
		Status: "running",
		Containers: []monitor.ContainerItem{
			{ID: "c123", Name: "app-test", State: "running"},
		},
	}
	m.cursorIdx = 0

	// 1. Open exec modal with "e"
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = res.(Model)
	if m.activeModal != "exec" {
		t.Fatalf("expected activeModal to be 'exec', got %q", m.activeModal)
	}

	// 2. Navigate down and up in menu
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.execCursorIdx != 1 {
		t.Errorf("expected execCursorIdx 1 after down, got %d", m.execCursorIdx)
	}

	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.execCursorIdx != 0 {
		t.Errorf("expected execCursorIdx 0 after up, got %d", m.execCursorIdx)
	}

	// Wrap around up
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.execCursorIdx != 6 {
		t.Errorf("expected execCursorIdx 6 after wrap up, got %d", m.execCursorIdx)
	}

	// 3. Switch to custom command mode
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = res.(Model)
	if !m.execIsCustomInput {
		t.Errorf("expected execIsCustomInput to be true")
	}

	// Type custom command "python3"
	for _, r := range "python3" {
		res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = res.(Model)
	}
	if m.execCustomCmd != "python3" {
		t.Errorf("expected execCustomCmd 'python3', got %q", m.execCustomCmd)
	}

	// Backspace
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = res.(Model)
	if m.execCustomCmd != "python" {
		t.Errorf("expected execCustomCmd 'python', got %q", m.execCustomCmd)
	}

	// Esc in custom command goes back to menu
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = res.(Model)
	if m.execIsCustomInput {
		t.Errorf("expected execIsCustomInput false after Esc")
	}
	if m.activeModal != "exec" {
		t.Errorf("expected activeModal still 'exec'")
	}

	// Direct shortcut "4" launches attach
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected modal to close on launching exec, got %q", m.activeModal)
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd for launching attach")
	}
}
