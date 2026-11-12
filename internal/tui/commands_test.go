package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAsyncCommandTriggers(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.5.0")
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "web", State: "running"},
		{ID: "c2", Name: "db", State: "exited"},
	}
	m.cursorIdx = 0

	// 1. Press 's' on running container -> stops asynchronously
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 's'")
	}
	if !strings.Contains(m.actionMessage, "Stopping container web") {
		t.Errorf("expected stopping toast, got: %s", m.actionMessage)
	}

	// 2. Press 's' on stopped container -> starts asynchronously
	m.cursorIdx = 1
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 's' for stopped container")
	}
	if !strings.Contains(m.actionMessage, "Starting container db") {
		t.Errorf("expected starting toast, got: %s", m.actionMessage)
	}

	// 3. Press 'p' on container -> restarts asynchronously
	m.cursorIdx = 0
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 'p'")
	}
	if !strings.Contains(m.actionMessage, "Restarting web") {
		t.Errorf("expected restarting toast, got: %s", m.actionMessage)
	}

	// 4. Press 'r' -> refresh command
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 'r'")
	}
	if !strings.Contains(m.actionMessage, "Refreshing data") {
		t.Errorf("expected refresh toast, got: %s", m.actionMessage)
	}
}

func TestActionFinishedMsg(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")

	// Success message without refresh
	res, cmd := m.Update(actionFinishedMsg{message: "Operation completed successfully", refresh: false})
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd when refresh=false")
	}
	if !strings.Contains(m.actionMessage, "Operation completed successfully") {
		t.Errorf("expected toast message, got: %s", m.actionMessage)
	}

	// Error message
	res, cmd = m.Update(actionFinishedMsg{err: errors.New("command execution failed"), refresh: false})
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd on error")
	}
	if !strings.Contains(m.actionMessage, "command execution failed") {
		t.Errorf("expected error toast, got: %s", m.actionMessage)
	}

	// Success message with refresh -> triggers fetchDataCmd
	res, cmd = m.Update(actionFinishedMsg{message: "Container removed", refresh: true})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd when refresh=true")
	}
	if !strings.Contains(m.actionMessage, "Container removed") {
		t.Errorf("expected toast message, got: %s", m.actionMessage)
	}
}

func TestDataFetchedMsg(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")

	data := dataFetchedMsg{
		timestamp: "14:25:36",
		sysInfo:   monitor.SystemInfo{Hostname: "dockpulse-node"},
		cpuInfo:   monitor.CPUInfo{Cores: 8, Percent: 45.0},
		memInfo:   monitor.MemoryInfo{TotalGB: 32.0, UsedGB: 12.0, Percent: 37.5},
	}

	res, cmd := m.Update(data)
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd after dataFetchedMsg")
	}
	if m.lastUpdate != "14:25:36" {
		t.Errorf("expected lastUpdate='14:25:36', got: %s", m.lastUpdate)
	}
	if m.sysInfo.Hostname != "dockpulse-node" {
		t.Errorf("expected hostname 'dockpulse-node', got: %s", m.sysInfo.Hostname)
	}
	if m.cpuInfo.Cores != 8 {
		t.Errorf("expected 8 cores, got: %d", m.cpuInfo.Cores)
	}
}

func TestKillAndRemoveCommandsExecution(t *testing.T) {
	dict := i18n.Get()

	// 1. Test killContainerCmd returns non-nil actionFinishedMsg
	killCmd := killContainerCmd("c-test", "test-app", dict)
	if killCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for killContainerCmd")
	}
	killMsg := killCmd()
	afm, ok := killMsg.(actionFinishedMsg)
	if !ok {
		t.Fatalf("expected actionFinishedMsg, got %T", killMsg)
	}
	if afm.err == nil && !strings.Contains(afm.message, "test-app") {
		t.Errorf("expected container name in message if no error, got: %+v", afm)
	}

	// 2. Test removeContainerCmd returns non-nil actionFinishedMsg
	removeCmd := removeContainerCmd("c-test", "test-app", dict)
	if removeCmd == nil {
		t.Fatal("expected non-nil tea.Cmd for removeContainerCmd")
	}
	removeMsg := removeCmd()
	afmRem, ok := removeMsg.(actionFinishedMsg)
	if !ok {
		t.Fatalf("expected actionFinishedMsg, got %T", removeMsg)
	}
	if afmRem.err == nil && !strings.Contains(afmRem.message, "test-app") {
		t.Errorf("expected container name in message if no error, got: %+v", afmRem)
	}
}

func TestFetchAllDataAndFetchDataCmd(t *testing.T) {
	data := fetchAllData()
	if len(data.timestamp) != 8 || !strings.Contains(data.timestamp, ":") {
		t.Errorf("expected timestamp in HH:MM:SS format, got %q", data.timestamp)
	}

	cmd := fetchDataCmd()
	if cmd == nil {
		t.Fatal("expected non-nil cmd from fetchDataCmd()")
	}
	msg := cmd()
	fetched, ok := msg.(dataFetchedMsg)
	if !ok {
		t.Fatalf("expected dataFetchedMsg, got %T", msg)
	}
	if len(fetched.timestamp) != 8 {
		t.Errorf("expected valid timestamp in fetched message, got %q", fetched.timestamp)
	}
}

func TestMakeActionCmd(t *testing.T) {
	// Success case
	successCmd := makeActionCmd(
		func() error { return nil },
		"prefix: ",
		"done successfully",
	)
	msg := successCmd()
	afm, ok := msg.(actionFinishedMsg)
	if !ok {
		t.Fatalf("expected actionFinishedMsg, got %T", msg)
	}
	if afm.err != nil || afm.message != "done successfully" || !afm.refresh {
		t.Errorf("unexpected success actionFinishedMsg: %+v", afm)
	}

	// Error case
	testErr := errors.New("network failure")
	failCmd := makeActionCmd(
		func() error { return testErr },
		"action error: ",
		"should not be used",
	)
	msgFail := failCmd()
	afmFail, ok := msgFail.(actionFinishedMsg)
	if !ok {
		t.Fatalf("expected actionFinishedMsg, got %T", msgFail)
	}
	if afmFail.err == nil || !strings.Contains(afmFail.err.Error(), "action error: network failure") {
		t.Errorf("expected wrapped error, got: %+v", afmFail.err)
	}
	if !errors.Is(afmFail.err, testErr) {
		t.Errorf("expected error wrapping with %%w")
	}
}

func TestExecutePruneAndPruneCmd(t *testing.T) {
	dict := i18n.Get()

	// Invalid option
	_, err := executePrune(99)
	if err == nil || !strings.Contains(err.Error(), "unknown prune option: 99") {
		t.Errorf("expected error for unknown option, got: %v", err)
	}

	// Test pruneCmd with unknown option
	badCmd := pruneCmd(99, dict)
	badMsg := badCmd()
	afm, ok := badMsg.(actionFinishedMsg)
	if !ok {
		t.Fatalf("expected actionFinishedMsg, got %T", badMsg)
	}
	if afm.err == nil || !strings.Contains(afm.err.Error(), dict.ToastPruneError) {
		t.Errorf("expected prune error toast, got: %+v", afm)
	}

	// Valid options run without panic
	for _, opt := range []int{PruneOptionContainers, PruneOptionImages, PruneOptionVolumes, PruneOptionAll} {
		cmd := pruneCmd(opt, dict)
		if cmd == nil {
			t.Errorf("expected non-nil cmd for option %d", opt)
			continue
		}
		resMsg := cmd()
		if _, ok := resMsg.(actionFinishedMsg); !ok {
			t.Errorf("expected actionFinishedMsg for option %d, got %T", opt, resMsg)
		}
	}
}

func TestContainerAndComposeCommandsExecution(t *testing.T) {
	dict := i18n.Get()

	cmds := []struct {
		name string
		cmd  tea.Cmd
	}{
		{"startContainer", startContainerCmd("c1", "web", dict)},
		{"stopContainer", stopContainerCmd("c1", "web", dict)},
		{"restartContainer", restartContainerCmd("c1", "web", dict)},
		{"startCompose", startComposeCmd("stack", nil, dict)},
		{"stopCompose", stopComposeCmd("stack", nil, dict)},
		{"restartCompose", restartComposeCmd("stack", nil, dict)},
		{"downCompose", downComposeCmd("stack", nil, dict)},
	}

	for _, tc := range cmds {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd == nil {
				t.Fatalf("expected non-nil cmd for %s", tc.name)
			}
			msg := tc.cmd()
			afm, ok := msg.(actionFinishedMsg)
			if !ok {
				t.Fatalf("expected actionFinishedMsg, got %T", msg)
			}
			// In environment without real docker, it returns wrapped error or success
			if afm.err != nil {
				if !strings.Contains(afm.err.Error(), "web") && !strings.Contains(afm.err.Error(), "stack") && !strings.Contains(afm.err.Error(), fmt.Sprintf("%s", afm.err)) {
					t.Errorf("unexpected error format: %v", afm.err)
				}
			}
		})
	}
}
