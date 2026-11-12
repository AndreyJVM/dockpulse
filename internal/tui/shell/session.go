package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
)

const readBufferSize = 32 * 1024

type ptyBackend interface {
	io.ReadWriteCloser
	Resize(width, height int) error
	Start(cmd *exec.Cmd) error
}

type terminalBackend interface {
	io.WriteCloser
	Render() string
	Resize(width, height int)
}

type ptyFactory func(width, height int) (ptyBackend, error)
type terminalFactory func(width, height int) terminalBackend

// Session owns one interactive process, its PTY, and the virtual terminal
// used to render process output inside Dockpulse.
type Session struct {
	mu       sync.Mutex
	pty      ptyBackend
	terminal terminalBackend
	cmd      *exec.Cmd
	closed   bool
}

// Start creates a PTY-backed interactive session for cmd.
func Start(cmd *exec.Cmd, width, height int) (*Session, error) {
	return startWithFactories(
		cmd,
		width,
		height,
		func(width, height int) (ptyBackend, error) {
			return xpty.NewPty(width, height)
		},
		func(width, height int) terminalBackend {
			terminal := vt.NewEmulator(width, height)
			terminal.SetScrollbackSize(1000)
			return terminal
		},
	)
}

func startWithFactories(cmd *exec.Cmd, width, height int, newPty ptyFactory, newTerminal terminalFactory) (*Session, error) {
	if cmd == nil {
		return nil, errors.New("interactive shell command is required")
	}
	width, height = normalizeSize(width, height)

	pty, err := newPty(width, height)
	if err != nil {
		return nil, fmt.Errorf("create shell PTY: %w", err)
	}
	terminal := newTerminal(width, height)
	if terminal == nil {
		_ = pty.Close()
		return nil, errors.New("create virtual terminal: empty terminal")
	}

	if err := pty.Start(cmd); err != nil {
		closeErr := errors.Join(pty.Close(), terminal.Close())
		return nil, errors.Join(fmt.Errorf("start interactive shell: %w", err), closeErr)
	}

	session := &Session{pty: pty, terminal: terminal, cmd: cmd}
	return session, nil
}

// Read waits for the next block of PTY output.
func (s *Session) Read() ([]byte, error) {
	buffer := make([]byte, readBufferSize)
	n, err := s.pty.Read(buffer)
	if n == 0 {
		return nil, err
	}
	data := make([]byte, n)
	copy(data, buffer[:n])
	return data, err
}

// Write sends keyboard input to the interactive process.
func (s *Session) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	_, err := s.pty.Write(data)
	return err
}

// Feed applies PTY output to the virtual terminal emulator.
func (s *Session) Feed(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	_, err := s.terminal.Write(data)
	return err
}

// Render returns the current virtual terminal screen.
func (s *Session) Render() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.terminal.Render()
}

// Resize updates both the child PTY and the virtual terminal screen.
func (s *Session) Resize(width, height int) error {
	width, height = normalizeSize(width, height)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	if err := s.pty.Resize(width, height); err != nil {
		return fmt.Errorf("resize shell PTY: %w", err)
	}
	s.terminal.Resize(width, height)
	return nil
}

// Wait waits until the interactive process exits.
func (s *Session) Wait() error {
	return xpty.WaitProcess(context.Background(), s.cmd)
}

// Close terminates the client process and releases terminal resources.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	pty := s.pty
	terminal := s.terminal
	cmd := s.cmd
	s.mu.Unlock()

	var killErr error
	if cmd != nil && cmd.Process != nil && cmd.ProcessState == nil {
		killErr = cmd.Process.Kill()
		if errors.Is(killErr, os.ErrProcessDone) {
			killErr = nil
		}
	}
	return errors.Join(killErr, pty.Close(), terminal.Close())
}

func normalizeSize(width, height int) (int, int) {
	return max(1, width), max(1, height)
}
