package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func TestComposeTabNavigationAndActions(t *testing.T) {
	m := NewModel(2*time.Second, 3, "v0.4.1")
	m.compose = []monitor.ComposeProject{
		{
			Name:            "proj-alpha",
			Status:          "running",
			StatusDisplay:   "Running",
			TotalServices:   2,
			RunningServices: 2,
			Services: []monitor.ComposeService{
				{Name: "web", State: "running"},
				{Name: "api", State: "running"},
			},
		},
		{
			Name:            "proj-beta",
			Status:          "stopped",
			StatusDisplay:   "Stopped",
			TotalServices:   1,
			RunningServices: 0,
			Services: []monitor.ComposeService{
				{Name: "worker", State: "exited"},
			},
		},
	}
	m.composeCursorIdx = 0

	// Down moves to second project
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.composeCursorIdx != 1 {
		t.Errorf("expected composeCursorIdx=1, got %d", m.composeCursorIdx)
	}

	// Down wraps back to 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.composeCursorIdx != 0 {
		t.Errorf("expected composeCursorIdx=0 after wrap, got %d", m.composeCursorIdx)
	}

	// Up wraps to 1
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.composeCursorIdx != 1 {
		t.Errorf("expected composeCursorIdx=1 after up wrap, got %d", m.composeCursorIdx)
	}

	// Key 'k' moves to 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = res.(Model)
	if m.composeCursorIdx != 0 {
		t.Errorf("expected composeCursorIdx=0 after 'k', got %d", m.composeCursorIdx)
	}

	// Key 'j' moves to 1
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = res.(Model)
	if m.composeCursorIdx != 1 {
		t.Errorf("expected composeCursorIdx=1 after 'j', got %d", m.composeCursorIdx)
	}

	// Footer on tab 3 contains compose hints
	view := m.View()
	if !strings.Contains(view, "Logs") || !strings.Contains(view, "Restart") || !strings.Contains(view, "Stop") || !strings.Contains(view, "Start") || !strings.Contains(view, "Down") {
		t.Errorf("expected compose hints in footer view, got: %s", view)
	}

	// Selection marker rendered
	if !strings.Contains(view, "► ") {
		t.Errorf("expected cursor marker '► ' in compose tab view")
	}

	// Test 'l' opens logs modal or triggers error toast
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = res.(Model)
	if m.activeModal == "logs" && !strings.Contains(m.logsTargetName, "proj-beta") {
		t.Errorf("expected logsTargetName to contain 'proj-beta', got %q", m.logsTargetName)
	}

	// Test 'p' restarts compose project
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = res.(Model)
	if m.actionMessage == "" {
		t.Errorf("expected actionMessage after pressing 'p'")
	}

	// Test 's' stops compose project
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = res.(Model)
	if m.actionMessage == "" {
		t.Errorf("expected actionMessage after pressing 's'")
	}

	// Test 'u' starts compose project
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	m = res.(Model)
	if m.actionMessage == "" {
		t.Errorf("expected actionMessage after pressing 'u'")
	}

	// Test 'd' opens compose down confirmation modal
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = res.(Model)
	if m.activeModal != "compose_down" {
		t.Fatalf("expected activeModal 'compose_down', got %q", m.activeModal)
	}
	if m.composeDownTarget != "proj-beta" {
		t.Errorf("expected composeDownTarget 'proj-beta', got %q", m.composeDownTarget)
	}

	modalView := m.View()
	if !strings.Contains(modalView, "CONFIRM DOCKER COMPOSE DOWN") || !strings.Contains(modalView, "proj-beta") {
		t.Errorf("expected compose down modal content, got: %s", modalView)
	}

	// Test 'n' cancels compose down modal
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected modal to be closed after 'n', got %q", m.activeModal)
	}

	// Re-open and confirm with 'y'
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = res.(Model)
	if m.activeModal != "compose_down" {
		t.Fatalf("expected activeModal 'compose_down', got %q", m.activeModal)
	}

	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to be cleared after 'y', got %q", m.activeModal)
	}
	if m.actionMessage == "" {
		t.Errorf("expected actionMessage after confirming down")
	}
}
