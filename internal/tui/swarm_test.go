package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSwarmTabNavigationAndLogs(t *testing.T) {
	m := NewModel(2*time.Second, 4, "v0.4.1")
	m.swarm = monitor.SwarmSummary{
		IsActive:      true,
		StatusDisplay: "Активен",
		NodesCount:    3,
		ManagersCount: 1,
		Services: []monitor.SwarmService{
			{ID: "svc1", Name: "web-svc", Image: "nginx:alpine", Replicas: "3/3"},
			{ID: "svc2", Name: "api-svc", Image: "golang:alpine", Replicas: "1/2"},
		},
	}
	m.swarmCursorIdx = 0

	// Down moves to second service
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.swarmCursorIdx != 1 {
		t.Errorf("expected swarmCursorIdx=1, got %d", m.swarmCursorIdx)
	}

	// Down wraps to 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.swarmCursorIdx != 0 {
		t.Errorf("expected swarmCursorIdx=0 after wrap, got %d", m.swarmCursorIdx)
	}

	// Up wraps to 1
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.swarmCursorIdx != 1 {
		t.Errorf("expected swarmCursorIdx=1 after up wrap, got %d", m.swarmCursorIdx)
	}

	// Key 'k' moves to 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = res.(Model)
	if m.swarmCursorIdx != 0 {
		t.Errorf("expected swarmCursorIdx=0 after 'k', got %d", m.swarmCursorIdx)
	}

	// Key 'j' moves to 1
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = res.(Model)
	if m.swarmCursorIdx != 1 {
		t.Errorf("expected swarmCursorIdx=1 after 'j', got %d", m.swarmCursorIdx)
	}

	// View shows cursor marker '► ' and footer hints
	view := m.View()
	if !strings.Contains(view, "► ") {
		t.Errorf("expected cursor marker '► ' in swarm tab view")
	}
	if !strings.Contains(view, "Выбрать службу") || !strings.Contains(view, "Логи службы") {
		t.Errorf("expected swarm footer hints in View(), got: %s", view)
	}

	// Pressing 'L' attempts to fetch logs
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = res.(Model)
	if m.activeModal == "logs" {
		if !strings.Contains(m.logsTargetName, "api-svc") {
			t.Errorf("expected logsTargetName to contain 'api-svc', got %q", m.logsTargetName)
		}
	} else if m.actionMessage == "" {
		t.Errorf("expected logs modal or actionMessage on pressing 'l'")
	}
}
