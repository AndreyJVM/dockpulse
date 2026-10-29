package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseFlags_Defaults(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cfg, err := parseFlags([]string{}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.interval != defaultInterval {
		t.Errorf("expected interval %v, got %v", defaultInterval, cfg.interval)
	}
	if cfg.initialTab != defaultTab {
		t.Errorf("expected initialTab %d, got %d", defaultTab, cfg.initialTab)
	}
	if cfg.showVersion {
		t.Errorf("expected showVersion to be false")
	}
}

func TestParseFlags_CustomValues(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantInterval time.Duration
		wantTab      int
	}{
		{
			name:         "short flags",
			args:         []string{"-i", "1s", "-t", "3"},
			wantInterval: 1 * time.Second,
			wantTab:      3,
		},
		{
			name:         "long flags",
			args:         []string{"--interval", "500ms", "--tab", "2"},
			wantInterval: 500 * time.Millisecond,
			wantTab:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cfg, err := parseFlags(tt.args, &stdout, &stderr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.interval != tt.wantInterval {
				t.Errorf("expected interval %v, got %v", tt.wantInterval, cfg.interval)
			}
			if cfg.initialTab != tt.wantTab {
				t.Errorf("expected initialTab %d, got %d", tt.wantTab, cfg.initialTab)
			}
		})
	}
}

func TestParseFlags_PositionalArgsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	_, err := parseFlags([]string{"extra-arg"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for positional arguments, got nil")
	}
	if !strings.Contains(err.Error(), "неизвестные позиционные аргументы") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config
		wantErr bool
	}{
		{
			name:    "valid defaults",
			cfg:     config{interval: 2 * time.Second, initialTab: 1},
			wantErr: false,
		},
		{
			name:    "valid boundary values",
			cfg:     config{interval: 100 * time.Millisecond, initialTab: 4},
			wantErr: false,
		},
		{
			name:    "interval too small",
			cfg:     config{interval: 50 * time.Millisecond, initialTab: 1},
			wantErr: true,
		},
		{
			name:    "interval negative",
			cfg:     config{interval: -1 * time.Second, initialTab: 1},
			wantErr: true,
		},
		{
			name:    "tab too low",
			cfg:     config{interval: 1 * time.Second, initialTab: 0},
			wantErr: true,
		},
		{
			name:    "tab too high",
			cfg:     config{interval: 1 * time.Second, initialTab: 5},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_Help(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		var stdout, stderr bytes.Buffer
		err := run([]string{flag}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("expected nil error for %s, got: %v", flag, err)
		}
		output := stdout.String()
		if !strings.Contains(output, "Dockpulse") || !strings.Contains(output, "Использование:") {
			t.Errorf("expected usage info in stdout for %s, got: %s", flag, output)
		}
	}
}

func TestRun_Version(t *testing.T) {
	for _, flag := range []string{"-v", "--version"} {
		var stdout, stderr bytes.Buffer
		err := run([]string{flag}, &stdout, &stderr)
		if err != nil {
			t.Fatalf("expected nil error for %s, got: %v", flag, err)
		}
		output := stdout.String()
		if !strings.Contains(output, "dockpulse") || !strings.Contains(output, "(Go Edition)") {
			t.Errorf("expected version info in stdout for %s, got: %s", flag, output)
		}
	}
}

func TestRun_ValidationFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--tab", "99"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid tab, got nil")
	}
	if !strings.Contains(err.Error(), "начальная вкладка должна быть в диапазоне от 1 до 4") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_ProgramExecution(t *testing.T) {
	origRunProgram := runProgram
	defer func() { runProgram = origRunProgram }()

	var executed bool
	runProgram = func(m tea.Model) error {
		executed = true
		return nil
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{"-i", "1s", "-t", "2"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("expected runProgram to be called")
	}

	// Test program error propagation
	expectedErr := errors.New("tui failed")
	runProgram = func(m tea.Model) error {
		return expectedErr
	}

	err = run([]string{}, &stdout, &stderr)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
