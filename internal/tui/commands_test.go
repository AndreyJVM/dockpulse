package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(m.actionMessage, "Остановка контейнера web") {
		t.Errorf("expected toast 'Остановка контейнера web', got: %s", m.actionMessage)
	}

	// 2. Press 's' on stopped container -> starts asynchronously
	m.cursorIdx = 1
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 's' for stopped container")
	}
	if !strings.Contains(m.actionMessage, "Запуск контейнера db") {
		t.Errorf("expected toast 'Запуск контейнера db', got: %s", m.actionMessage)
	}

	// 3. Press 'p' on container -> restarts asynchronously
	m.cursorIdx = 0
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 'p'")
	}
	if !strings.Contains(m.actionMessage, "Перезапуск web") {
		t.Errorf("expected toast 'Перезапуск web', got: %s", m.actionMessage)
	}

	// 4. Press 'r' -> refresh command
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd on 'r'")
	}
	if !strings.Contains(m.actionMessage, "Обновление данных") {
		t.Errorf("expected refresh toast, got: %s", m.actionMessage)
	}
}

func TestActionFinishedMsg(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")

	// Success message without refresh
	res, cmd := m.Update(actionFinishedMsg{message: "Операция завершена успешно", refresh: false})
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd when refresh=false")
	}
	if !strings.Contains(m.actionMessage, "Операция завершена успешно") {
		t.Errorf("expected toast message, got: %s", m.actionMessage)
	}

	// Error message
	res, cmd = m.Update(actionFinishedMsg{err: errors.New("ошибка выполнения команды"), refresh: false})
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd on error")
	}
	if !strings.Contains(m.actionMessage, "ошибка выполнения команды") {
		t.Errorf("expected error toast, got: %s", m.actionMessage)
	}

	// Success message with refresh -> triggers fetchDataCmd
	res, cmd = m.Update(actionFinishedMsg{message: "Контейнер удален", refresh: true})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd when refresh=true")
	}
	if !strings.Contains(m.actionMessage, "Контейнер удален") {
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
