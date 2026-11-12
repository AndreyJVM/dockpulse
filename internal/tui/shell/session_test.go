package shell

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
)

type fakePty struct {
	bytes.Buffer
	startErr    error
	started     bool
	closedCount int
	width       int
	height      int
}

func (p *fakePty) Read(_ []byte) (int, error) { return 0, io.EOF }
func (p *fakePty) Close() error {
	p.closedCount++
	return nil
}
func (p *fakePty) Resize(width, height int) error {
	p.width, p.height = width, height
	return nil
}
func (p *fakePty) Start(_ *exec.Cmd) error {
	p.started = true
	return p.startErr
}

type fakeTerminal struct {
	output      bytes.Buffer
	closedCount int
	width       int
	height      int
}

func (t *fakeTerminal) Read(_ []byte) (int, error) { return 0, io.EOF }
func (t *fakeTerminal) Write(data []byte) (int, error) {
	return t.output.Write(data)
}
func (t *fakeTerminal) Close() error {
	t.closedCount++
	return nil
}
func (t *fakeTerminal) Render() string { return t.output.String() }
func (t *fakeTerminal) Resize(width, height int) {
	t.width, t.height = width, height
}

func TestStartRejectsNilCommandBeforeOpeningPty(t *testing.T) {
	ptyCreated := false
	_, err := startWithFactories(nil, 80, 20,
		func(_, _ int) (ptyBackend, error) {
			ptyCreated = true
			return &fakePty{}, nil
		},
		func(_, _ int) terminalBackend { return &fakeTerminal{} },
	)

	if err == nil {
		t.Fatal("expected nil command to be rejected")
	}
	if ptyCreated {
		t.Fatal("expected no PTY allocation for a nil command")
	}
}

func TestStartFailureClosesAllocatedResources(t *testing.T) {
	fakeErr := errors.New("start failed")
	pty := &fakePty{startErr: fakeErr}
	terminal := &fakeTerminal{}

	_, err := startWithFactories(exec.Command("ignored"), 80, 20,
		func(_, _ int) (ptyBackend, error) { return pty, nil },
		func(_, _ int) terminalBackend { return terminal },
	)

	if !errors.Is(err, fakeErr) {
		t.Fatalf("expected start error, got %v", err)
	}
	if pty.closedCount != 1 || terminal.closedCount != 1 {
		t.Fatalf("expected allocated resources to close once, PTY=%d terminal=%d", pty.closedCount, terminal.closedCount)
	}
}

func TestSessionInputOutputResizeAndIdempotentClose(t *testing.T) {
	pty := &fakePty{}
	terminal := &fakeTerminal{}
	session, err := startWithFactories(exec.Command("ignored"), 80, 20,
		func(_, _ int) (ptyBackend, error) { return pty, nil },
		func(_, _ int) terminalBackend { return terminal },
	)
	if err != nil {
		t.Fatalf("startWithFactories() returned an error: %v", err)
	}

	if err := session.Write([]byte("pwd\r")); err != nil {
		t.Fatalf("Write() returned an error: %v", err)
	}
	if got := pty.String(); got != "pwd\r" {
		t.Fatalf("unexpected PTY input: %q", got)
	}
	if err := session.Feed([]byte("/app\r\n")); err != nil {
		t.Fatalf("Feed() returned an error: %v", err)
	}
	if got := session.Render(); got != "/app\r\n" {
		t.Fatalf("unexpected terminal render: %q", got)
	}

	if err := session.Resize(100, 30); err != nil {
		t.Fatalf("Resize() returned an error: %v", err)
	}
	if pty.width != 100 || pty.height != 30 || terminal.width != 100 || terminal.height != 30 {
		t.Fatalf("expected both PTY and terminal to resize, PTY=%dx%d terminal=%dx%d", pty.width, pty.height, terminal.width, terminal.height)
	}

	if err := session.Close(); err != nil {
		t.Fatalf("Close() returned an error: %v", err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("second Close() returned an error: %v", err)
	}
	if pty.closedCount != 1 || terminal.closedCount != 1 {
		t.Fatalf("expected idempotent close, PTY=%d terminal=%d", pty.closedCount, terminal.closedCount)
	}
}

func TestSessionRunsLocalPtyProcess(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not available")
	}

	session, err := Start(exec.Command(sh, "-c", "printf 'pty-ready\\r\\n'"), 40, 8)
	if err != nil {
		t.Fatalf("Start() returned an error: %v", err)
	}
	defer session.Close()

	data, readErr := session.Read()
	if len(data) == 0 {
		t.Fatalf("expected PTY output, read error: %v", readErr)
	}
	if err := session.Feed(data); err != nil {
		t.Fatalf("Feed() returned an error: %v", err)
	}
	if got := session.Render(); !strings.Contains(got, "pty-ready") {
		t.Fatalf("expected rendered PTY output, got %q", got)
	}
	if err := session.Wait(); err != nil {
		t.Fatalf("Wait() returned an error: %v", err)
	}
}
