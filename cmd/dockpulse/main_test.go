package main

import (
	"bytes"
	"errors"
	"os"
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
			name:         "short flags numeric",
			args:         []string{"-i", "1s", "-t", "3"},
			wantInterval: 1 * time.Second,
			wantTab:      3,
		},
		{
			name:         "long flags numeric",
			args:         []string{"--interval", "500ms", "--tab", "2"},
			wantInterval: 500 * time.Millisecond,
			wantTab:      2,
		},
		{
			name:         "named tab host",
			args:         []string{"--tab", "host"},
			wantInterval: defaultInterval,
			wantTab:      1,
		},
		{
			name:         "named tab docker",
			args:         []string{"-t", "docker"},
			wantInterval: defaultInterval,
			wantTab:      2,
		},
		{
			name:         "named tab compose",
			args:         []string{"--tab", "compose"},
			wantInterval: defaultInterval,
			wantTab:      3,
		},
		{
			name:         "named tab swarm",
			args:         []string{"-t", "swarm"},
			wantInterval: defaultInterval,
			wantTab:      4,
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
				t.Errorf("expected interval %v, got %v", defaultInterval, cfg.interval)
			}
			if cfg.initialTab != tt.wantTab {
				t.Errorf("expected initialTab %d, got %d", tt.wantTab, cfg.initialTab)
			}
		})
	}
}

func TestParseFlags_EnvVariables(t *testing.T) {
	origInterval := os.Getenv("DOCKPULSE_INTERVAL")
	origTab := os.Getenv("DOCKPULSE_TAB")
	defer func() {
		os.Setenv("DOCKPULSE_INTERVAL", origInterval)
		os.Setenv("DOCKPULSE_TAB", origTab)
	}()

	os.Setenv("DOCKPULSE_INTERVAL", "3s")
	os.Setenv("DOCKPULSE_TAB", "compose")

	var stdout, stderr bytes.Buffer
	cfg, err := parseFlags([]string{}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.interval != 3*time.Second {
		t.Errorf("expected interval from env 3s, got %v", cfg.interval)
	}
	if cfg.initialTab != 3 {
		t.Errorf("expected tab from env 3 (compose), got %d", cfg.initialTab)
	}
	// CLI flags must override env variables
	cfgOverridden, err := parseFlags([]string{"-i", "500ms", "-t", "swarm"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfgOverridden.interval != 500*time.Millisecond {
		t.Errorf("expected overridden interval 500ms, got %v", cfgOverridden.interval)
	}
	if cfgOverridden.initialTab != 4 {
		t.Errorf("expected overridden tab 4 (swarm), got %d", cfgOverridden.initialTab)
	}
}

func TestParseTab(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1", 1, false},
		{"host", 1, false},
		{"Host", 1, false},
		{"2", 2, false},
		{"docker", 2, false},
		{"3", 3, false},
		{"compose", 3, false},
		{"4", 4, false},
		{"swarm", 4, false},
		{"5", 0, true},
		{"unknown", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTab(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTab(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseTab(%q) = %d, want = %d", tt.input, got, tt.want)
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
	if !strings.Contains(err.Error(), "unknown positional arguments") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseFlags_RejectsLanguageFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	_, err := parseFlags([]string{"--lang", "fr"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected the removed --lang flag to be rejected")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined: -lang") {
		t.Fatalf("unexpected error: %v", err)
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
		if !strings.Contains(output, "Dockpulse") || !strings.Contains(output, "Usage:") {
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
		if !strings.Contains(output, "dockpulse") || !strings.Contains(output, version) {
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
	if !strings.Contains(err.Error(), "invalid initial tab") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_TerminalCheck(t *testing.T) {
	origIsTerminal := isTerminal
	defer func() { isTerminal = origIsTerminal }()

	// Non-terminal check.
	isTerminal = func(_ uintptr) bool {
		return false
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when running outside of interactive terminal, got nil")
	}
	if !strings.Contains(err.Error(), "requires an interactive terminal (TTY)") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_ProgramExecution(t *testing.T) {
	origRunProgram := runProgram
	origIsTerminal := isTerminal
	defer func() {
		runProgram = origRunProgram
		isTerminal = origIsTerminal
	}()

	isTerminal = func(_ uintptr) bool {
		return true
	}

	var executed bool
	runProgram = func(_ tea.Model) error {
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
	runProgram = func(_ tea.Model) error {
		return expectedErr
	}

	err = run([]string{}, &stdout, &stderr)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
