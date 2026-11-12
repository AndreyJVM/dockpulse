package tui

import (
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeEmbeddedShell struct {
	writes       [][]byte
	feed         []byte
	resizeWidth  int
	resizeHeight int
	closeCount   int
	render       string
}

func (s *fakeEmbeddedShell) Read() ([]byte, error) { return nil, io.EOF }
func (s *fakeEmbeddedShell) Write(data []byte) error {
	s.writes = append(s.writes, append([]byte(nil), data...))
	return nil
}
func (s *fakeEmbeddedShell) Feed(data []byte) error {
	s.feed = append(s.feed, data...)
	return nil
}
func (s *fakeEmbeddedShell) Render() string { return s.render }
func (s *fakeEmbeddedShell) Resize(width, height int) error {
	s.resizeWidth, s.resizeHeight = width, height
	return nil
}
func (s *fakeEmbeddedShell) Wait() error { return nil }
func (s *fakeEmbeddedShell) Close() error {
	s.closeCount++
	return nil
}

func TestEmbeddedShellLifecycleAndFocus(t *testing.T) {
	originalStart := startEmbeddedShell
	defer func() { startEmbeddedShell = originalStart }()

	fake := &fakeEmbeddedShell{render: "root@api:/app#"}
	var startedCommand *exec.Cmd
	startEmbeddedShell = func(cmd *exec.Cmd, _, _ int) (embeddedShellSession, error) {
		startedCommand = cmd
		return fake, nil
	}

	m := NewModel(2*time.Second, int(TabDocker), "v0.9.0")
	m.width, m.height = 140, 40
	m.activeModal = ModalExec
	m.execTargetID = "container-id"
	m.execTargetName = "api"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected shell read and wait commands")
	}
	if m.embeddedShell != fake || !m.shellFocused {
		t.Fatalf("expected an active focused embedded shell")
	}
	if startedCommand == nil || !strings.Contains(strings.Join(startedCommand.Args, " "), "TERM=xterm-256color") {
		t.Fatalf("expected an xterm-compatible Docker exec command, got %#v", startedCommand)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("pwd")})
	m = updated.(Model)
	if len(fake.writes) != 1 || string(fake.writes[0]) != "pwd" {
		t.Fatalf("expected keyboard input to reach the shell, got %#v", fake.writes)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlCloseBracket})
	m = updated.(Model)
	if m.shellFocused {
		t.Fatal("expected Ctrl+] to return focus to Dockpulse")
	}
	if len(fake.writes) != 1 {
		t.Fatal("expected Ctrl+] not to be forwarded to the shell")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = updated.(Model)
	if !m.shellFocused {
		t.Fatal("expected E to focus the existing shell")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlCloseBracket})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m = updated.(Model)
	if m.embeddedShell != nil || fake.closeCount != 1 {
		t.Fatalf("expected Ctrl+X to close the unfocused shell once")
	}
}

func TestEmbeddedShellStartFailureDoesNotLaunchFallback(t *testing.T) {
	originalStart := startEmbeddedShell
	defer func() { startEmbeddedShell = originalStart }()

	startEmbeddedShell = func(_ *exec.Cmd, _, _ int) (embeddedShellSession, error) {
		return nil, errors.New("PTY unavailable")
	}
	m := NewModel(2*time.Second, int(TabDocker), "v0.9.0")
	m.activeModal = ModalExec
	m.execTargetID = "container-id"
	m.execTargetName = "api"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("expected no external fallback command after PTY failure")
	}
	if m.embeddedShell != nil || !strings.Contains(m.actionMessage, "PTY unavailable") {
		t.Fatalf("expected a visible PTY error without an active session, message=%q", m.actionMessage)
	}
}

func TestShellKeyBytes(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
		want string
	}{
		{name: "runes", key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ls")}, want: "ls"},
		{name: "enter", key: tea.KeyMsg{Type: tea.KeyEnter}, want: "\r"},
		{name: "ctrl c", key: tea.KeyMsg{Type: tea.KeyCtrlC}, want: "\x03"},
		{name: "up", key: tea.KeyMsg{Type: tea.KeyUp}, want: "\x1b[A"},
		{name: "shift tab", key: tea.KeyMsg{Type: tea.KeyShiftTab}, want: "\x1b[Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(shellKeyBytes(tt.key)); got != tt.want {
				t.Fatalf("shellKeyBytes(%s) = %q, want %q", tt.key.String(), got, tt.want)
			}
		})
	}
}

func TestEmbeddedShellOutputAndExitMessagesIgnoreStaleSessions(t *testing.T) {
	active := &fakeEmbeddedShell{}
	stale := &fakeEmbeddedShell{}
	m := NewModel(2*time.Second, int(TabDocker), "v0.9.0")
	m.embeddedShell = active
	m.shellContainerName = "api"

	updated, cmd := m.Update(embeddedShellOutputMsg{session: stale, data: []byte("stale")})
	m = updated.(Model)
	if cmd != nil || len(active.feed) != 0 {
		t.Fatal("expected stale shell output to be ignored")
	}

	updated, cmd = m.Update(embeddedShellOutputMsg{session: active, data: []byte("ready")})
	m = updated.(Model)
	if cmd == nil || string(active.feed) != "ready" {
		t.Fatalf("expected active output to be rendered and reading to continue")
	}

	updated, _ = m.Update(embeddedShellExitedMsg{session: stale, err: errors.New("stale")})
	m = updated.(Model)
	if m.embeddedShell != active {
		t.Fatal("expected stale exit message not to close the active shell")
	}

	updated, _ = m.Update(embeddedShellExitedMsg{session: active, err: nil})
	m = updated.(Model)
	if m.embeddedShell != nil || active.closeCount != 1 {
		t.Fatal("expected active shell exit to close and clear the session")
	}
}

func TestViewRendersEmbeddedShellAndFocusHelp(t *testing.T) {
	m := NewModel(2*time.Second, int(TabDocker), "v0.9.0")
	m.width, m.height = 140, 40
	m.embeddedShell = &fakeEmbeddedShell{render: "root@api:/app# pwd"}
	m.shellContainerName = "api"
	m.shellFocused = true

	view := m.View()
	for _, want := range []string{
		"TERMINAL: api",
		"root@api:/app# pwd",
		"Ctrl+]",
		"Return to Dockpulse",
		"Ctrl+D",
		"Exit shell",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected embedded shell view to contain %q, got:\n%s", want, view)
		}
	}
}
