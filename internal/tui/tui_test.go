package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AndreyJVM/system-monitor/internal/monitor"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"привет мир", 9, "привет..."},
		{"абвгдеж", 3, "абв"},
		{"абв", 2, "аб"},
		{"", 5, ""},
		{"short", 0, ""},
	}

	for _, c := range cases {
		got := truncate(c.input, c.maxLen)
		if got != c.expected {
			t.Errorf("truncate(%q, %d) = %q; want %q", c.input, c.maxLen, got, c.expected)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	p1 := formatPercent(25.0)
	if !strings.Contains(p1, "25.0%") {
		t.Errorf("expected '25.0%%' in %q", p1)
	}

	p2 := formatPercent(65.0)
	if !strings.Contains(p2, "65.0%") {
		t.Errorf("expected '65.0%%' in %q", p2)
	}

	p3 := formatPercent(92.0)
	if !strings.Contains(p3, "92.0%") {
		t.Errorf("expected '92.0%%' in %q", p3)
	}
}

func TestFormatBar(t *testing.T) {
	cases := []struct {
		pct   float64
		width int
	}{
		{-5.0, 10},
		{0.0, 10},
		{50.0, 20},
		{75.0, 15},
		{90.0, 15},
		{110.0, 10},
	}

	for _, c := range cases {
		bar := formatBar(c.pct, c.width)
		if !strings.HasPrefix(bar, "[") || !strings.Contains(bar, "]") || !strings.Contains(bar, "%") {
			t.Errorf("formatBar(%f, %d) returned invalid format: %q", c.pct, c.width, bar)
		}
	}
}

func TestNewModel(t *testing.T) {
	// Defaults when invalid arguments are passed
	m := NewModel(0, 99, "v1.0.0")
	if m.activeTab != 1 {
		t.Errorf("expected activeTab to default to 1, got %d", m.activeTab)
	}
	if m.refreshInterval != 2*time.Second {
		t.Errorf("expected refreshInterval to default to 2s, got %v", m.refreshInterval)
	}
	if m.version != "v1.0.0" {
		t.Errorf("expected version 'v1.0.0', got %q", m.version)
	}

	// Init command
	cmd := m.Init()
	if cmd == nil {
		t.Errorf("expected non-nil Init cmd")
	}
}

func TestModelNavigationAndUpdates(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.4.1")

	// Resize
	updatedModel, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd != nil {
		t.Errorf("expected nil cmd on window resize")
	}
	m = updatedModel.(Model)
	if m.width != 120 || m.height != 40 {
		t.Errorf("expected dimensions 120x40, got %dx%d", m.width, m.height)
	}

	// Tab switching
	keyTests := []struct {
		key         string
		expectedTab int
	}{
		{"tab", 2},
		{"tab", 3},
		{"right", 4},
		{"tab", 1},
		{"left", 4},
		{"shift+tab", 3},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
	}

	for _, kt := range keyTests {
		res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(kt.key)})
		m = res.(Model)
		if m.activeTab != kt.expectedTab {
			t.Errorf("after key %q: expected tab %d, got %d", kt.key, kt.expectedTab, m.activeTab)
		}
	}

	// Help modal toggle
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = res.(Model)
	if !m.showHelp {
		t.Errorf("expected showHelp=true after '?'")
	}

	// Close help
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)
	if m.showHelp {
		t.Errorf("expected showHelp=false after 'esc'")
	}

	// Quit keys
	_, quitCmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quitCmd == nil {
		t.Errorf("expected quitCmd on ctrl+c")
	}
}

func TestContainerCursorAndModals(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")

	// Populate dummy containers
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{
			ID:         "c1",
			Name:       "web-server",
			Image:      "nginx:latest",
			State:      "running",
			CPUPerc:    5.2,
			MemUsageMB: 45.0,
			MemPerc:    12.0,
		},
		{
			ID:    "c2",
			Name:  "database",
			Image: "postgres:16",
			State: "stopped",
		},
	}
	m.cursorIdx = 0

	// Cursor navigation
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.cursorIdx != 1 {
		t.Errorf("expected cursorIdx=1 after down, got %d", m.cursorIdx)
	}

	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.cursorIdx != 0 {
		t.Errorf("expected cursorIdx=0 after up, got %d", m.cursorIdx)
	}

	// Open inspect modal
	m.inspectTarget = monitor.ContainerDetail{
		ID:         "c1",
		Name:       "web-server",
		Image:      "nginx:latest",
		State:      "running",
		CPUPerc:    5.2,
		MemUsageMB: 45.0,
	}
	m.activeModal = "inspect"

	view := m.View()
	if !strings.Contains(view, "ИНСПЕКЦИЯ КОНТЕЙНЕРА") {
		t.Errorf("expected inspect modal title in View(), got %q", view)
	}

	// Close inspect modal with 'q'
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to close on 'q', got %q", m.activeModal)
	}

	// Open logs modal and test scrolling
	m.activeModal = "logs"
	m.logsLines = []string{
		"line 1", "line 2", "line 3", "line 4", "line 5",
		"line 6", "line 7", "line 8", "line 9", "line 10",
		"line 11", "line 12", "line 13", "line 14", "line 15",
	}
	m.logsScroll = 0

	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.logsScroll != 1 {
		t.Errorf("expected logsScroll=1 after down, got %d", m.logsScroll)
	}

	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.logsScroll != 0 {
		t.Errorf("expected logsScroll=0 after up, got %d", m.logsScroll)
	}

	// Page down
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = res.(Model)
	if m.logsScroll == 0 {
		t.Errorf("expected logsScroll to increase on PgDown")
	}

	// Close logs modal
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to close on Esc, got %q", m.activeModal)
	}
}

func TestRenderAllTabs(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.4.1")

	// Render Tab 1: Host
	m.activeTab = 1
	v1 := m.View()
	if !strings.Contains(v1, "СИСТЕМНАЯ ИНФОРМАЦИЯ") {
		t.Errorf("expected Host tab header in View()")
	}

	// Render Tab 2: Docker
	m.activeTab = 2
	v2 := m.View()
	if !strings.Contains(v2, "DOCKER КОНТЕЙНЕРЫ") {
		t.Errorf("expected Docker tab header in View()")
	}

	// Render Tab 3: Compose
	m.activeTab = 3
	v3 := m.View()
	if !strings.Contains(v3, "DOCKER COMPOSE") {
		t.Errorf("expected Compose tab header in View()")
	}

	// Render Tab 4: Swarm
	m.activeTab = 4
	v4 := m.View()
	if !strings.Contains(v4, "DOCKER SWARM") {
		t.Errorf("expected Swarm tab header in View()")
	}

	// Render Help modal
	m.showHelp = true
	vHelp := m.View()
	if !strings.Contains(vHelp, "ГОРЯЧИЕ КЛАВИШИ") {
		t.Errorf("expected Help modal in View()")
	}
}

func TestFilterContainers(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "web-gateway", Image: "nginx:alpine", ComposeProject: "infra"},
		{ID: "c2", Name: "db-postgres", Image: "postgres:16", ComposeProject: "backend"},
		{ID: "c3", Name: "redis-cache", Image: "redis:7-alpine", ComposeProject: "cache"},
	}

	// Empty query
	m.filterQuery = ""
	if len(m.getFilteredContainers()) != 3 {
		t.Errorf("expected all 3 containers for empty filter query")
	}

	// Filter by Name
	m.filterQuery = "web"
	filtered := m.getFilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "web-gateway" {
		t.Errorf("expected 1 result 'web-gateway', got %v", filtered)
	}

	// Case-insensitive match on Image
	m.filterQuery = "POSTGRES"
	filtered = m.getFilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "db-postgres" {
		t.Errorf("expected 1 result 'db-postgres', got %v", filtered)
	}

	// Match on ID
	m.filterQuery = "c3"
	filtered = m.getFilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "redis-cache" {
		t.Errorf("expected 1 result 'redis-cache', got %v", filtered)
	}

	// Match on ComposeProject
	m.filterQuery = "infra"
	filtered = m.getFilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "web-gateway" {
		t.Errorf("expected 1 result for compose project 'infra', got %v", filtered)
	}

	// No match
	m.filterQuery = "non-existent"
	if len(m.getFilteredContainers()) != 0 {
		t.Errorf("expected 0 results for non-matching query")
	}
}

func TestFilterModeKeys(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "web", Image: "nginx"},
		{ID: "c2", Name: "worker", Image: "app"},
	}

	// 1. Activate search mode with '/'
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = res.(Model)
	if !m.isFiltering {
		t.Fatalf("expected isFiltering=true after '/'")
	}

	// 2. Type "w", "e", "b"
	for _, r := range []rune("web") {
		res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = res.(Model)
	}
	if m.filterQuery != "web" {
		t.Errorf("expected filterQuery='web', got %q", m.filterQuery)
	}
	if len(m.getFilteredContainers()) != 1 {
		t.Errorf("expected 1 container matching 'web', got %d", len(m.getFilteredContainers()))
	}

	// 3. Backspace
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = res.(Model)
	if m.filterQuery != "we" {
		t.Errorf("expected filterQuery='we' after backspace, got %q", m.filterQuery)
	}

	// 4. Enter to apply filter and exit typing mode
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)
	if m.isFiltering {
		t.Errorf("expected isFiltering=false after Enter")
	}
	if m.filterQuery != "we" {
		t.Errorf("expected filterQuery to persist after Enter")
	}

	// 5. Esc in normal mode clears filter
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)
	if m.filterQuery != "" {
		t.Errorf("expected filterQuery='' after Esc, got %q", m.filterQuery)
	}
}

func TestRenderDockerTabFilteringView(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")
	m.activeTab = 2
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "alpha", Image: "alpine"},
	}

	// Active search input
	m.isFiltering = true
	m.filterQuery = "alp"
	vSearch := m.View()
	if !strings.Contains(vSearch, "[Поиск: alp█]") {
		t.Errorf("expected search prompt in View(), got: %s", vSearch)
	}

	// Applied filter view
	m.isFiltering = false
	vFilter := m.View()
	if !strings.Contains(vFilter, "Фильтр: alp") {
		t.Errorf("expected filter badge in View(), got: %s", vFilter)
	}

	// Empty search results
	m.filterQuery = "nonexistent"
	vEmpty := m.View()
	if !strings.Contains(vEmpty, "не найдены") {
		t.Errorf("expected 'не найдены' in View(), got: %s", vEmpty)
	}
}
