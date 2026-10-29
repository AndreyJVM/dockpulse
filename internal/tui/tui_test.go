package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
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

func TestComposeTabNavigationAndActions(t *testing.T) {
	m := NewModel(2*time.Second, 3, "v0.4.1")
	m.compose = []monitor.ComposeProject{
		{
			Name:            "proj-alpha",
			Status:          "running",
			StatusDisplay:   "Запущен",
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
			StatusDisplay:   "Остановлен",
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
	if !strings.Contains(view, "Логи стека") || !strings.Contains(view, "Перезапуск стека") {
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
}

func TestDockerTabHealthBadges(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")
	m.activeTab = 2
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "healthy-app", Image: "nginx", State: "running", Health: "healthy"},
		{ID: "c2", Name: "broken-app", Image: "postgres", State: "running", Health: "unhealthy"},
		{ID: "c3", Name: "booting-app", Image: "redis", State: "running", Health: "starting"},
		{ID: "c4", Name: "plain-app", Image: "busybox", State: "running", Health: ""},
	}

	view := m.View()
	if !strings.Contains(view, "HEALTH") {
		t.Errorf("expected HEALTH column in header, got: %s", view)
	}
	if !strings.Contains(view, "[healthy]") {
		t.Errorf("expected [healthy] badge in View(), got: %s", view)
	}
	if !strings.Contains(view, "[unhealthy]") {
		t.Errorf("expected [unhealthy] badge in View(), got: %s", view)
	}
	if !strings.Contains(view, "[starting]") {
		t.Errorf("expected [starting] badge in View(), got: %s", view)
	}
}

func TestFilterContainersByHealth(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "web", Image: "nginx", Health: "healthy"},
		{ID: "c2", Name: "db", Image: "postgres", Health: "unhealthy"},
		{ID: "c3", Name: "cache", Image: "redis", Health: "starting"},
	}

	m.filterQuery = "unhealthy"
	filtered := m.getFilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "db" {
		t.Errorf("expected 1 result matching health 'unhealthy', got %v", filtered)
	}

	m.filterQuery = "healthy"
	filtered = m.getFilteredContainers()
	// "healthy" is substring of "unhealthy" as well, so matches c1 and c2
	if len(filtered) != 2 {
		t.Errorf("expected 2 results matching 'healthy', got %d", len(filtered))
	}
}

func TestInspectModalHealthBadge(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.4.1")
	m.activeModal = "inspect"
	m.inspectTarget = monitor.ContainerDetail{
		ID:     "c1",
		Name:   "app-prod",
		Image:  "my-app:v1",
		Status: "Up 2 hours",
		Health: "healthy",
	}

	view := m.View()
	if !strings.Contains(view, "Здоровье:") || !strings.Contains(view, "[healthy]") {
		t.Errorf("expected health badge in inspect modal view, got: %s", view)
	}

	// Test unhealthy
	m.inspectTarget.Health = "unhealthy"
	view = m.View()
	if !strings.Contains(view, "[unhealthy]") {
		t.Errorf("expected [unhealthy] badge in inspect modal view, got: %s", view)
	}

	// Test starting
	m.inspectTarget.Health = "starting"
	view = m.View()
	if !strings.Contains(view, "[starting]") {
		t.Errorf("expected [starting] badge in inspect modal view, got: %s", view)
	}

	// Test no health configured
	m.inspectTarget.Health = ""
	view = m.View()
	if !strings.Contains(view, "Здоровье:") {
		t.Errorf("expected 'Здоровье:' line even if health is empty, got: %s", view)
	}
}

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

func TestContainerExecHotkeysAndMessages(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.5.0")
	m.dockerInfo = monitor.DockerSummary{
		Status:        "running",
		StatusDisplay: "[Запущен]",
		Containers: []monitor.ContainerItem{
			{ID: "c1", Name: "running-container", State: "running", Status: "Up 2 hours"},
			{ID: "c2", Name: "stopped-container", State: "exited", Status: "Exited (0) 5 mins ago"},
		},
	}
	m.cursorIdx = 0

	// 1. Press 'e' on running container
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = res.(Model)
	if monitor.IsDockerCLIAvailable() {
		if cmd == nil {
			t.Errorf("expected non-nil tea.Cmd from tea.ExecProcess on running container")
		}
	} else {
		if !strings.Contains(m.actionMessage, "утилита docker не найдена") {
			t.Errorf("expected docker not found message, got %q", m.actionMessage)
		}
	}

	// 2. Press 'E' on stopped container
	m.cursorIdx = 1
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd for stopped container, got non-nil")
	}
	if !strings.Contains(m.actionMessage, "Нельзя подключиться") {
		t.Errorf("expected cannot connect message, got %q", m.actionMessage)
	}

	// 3. ExecFinishedMsg with nil error
	res, _ = m.Update(execFinishedMsg{containerName: "running-container", err: nil})
	m = res.(Model)
	if !strings.Contains(m.actionMessage, "Сессия shell в running-container завершена") {
		t.Errorf("expected shell session finished message, got %q", m.actionMessage)
	}

	// 4. ExecFinishedMsg with error
	res, _ = m.Update(execFinishedMsg{containerName: "running-container", err: fmt.Errorf("shell not found")})
	m = res.(Model)
	if !strings.Contains(m.actionMessage, "Ошибка shell в running-container") {
		t.Errorf("expected shell error message, got %q", m.actionMessage)
	}

	// 5. Check footer and help hints
	view := m.View()
	if !strings.Contains(view, "[E]") || !strings.Contains(view, "Shell") {
		t.Errorf("expected '[E] Shell' in tab 2 footer view, got: %s", view)
	}

	m.showHelp = true
	helpView := m.View()
	if !strings.Contains(helpView, "[E]") || !strings.Contains(helpView, "Интерактивный shell") {
		t.Errorf("expected '[E]' and 'Интерактивный shell' in help modal, got: %s", helpView)
	}
}

func TestPruneModalAndWorkflow(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.5.0")
	m.width = 100
	m.height = 30

	// 1. Open prune modal on tab 2 with "c"
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = res.(Model)
	if m.activeModal != "prune" {
		t.Fatalf("expected activeModal to be 'prune', got %q", m.activeModal)
	}

	// 2. View in menu state
	view := m.View()
	if !strings.Contains(view, "ОЧИСТКА РЕСУРСОВ DOCKER") {
		t.Errorf("expected prune modal header in view, got: %s", view)
	}
	if !strings.Contains(view, "Очистить остановленные контейнеры") {
		t.Errorf("expected menu item 1 in view, got: %s", view)
	}
	if !strings.Contains(view, "[1-4]") {
		t.Errorf("expected [1-4] footer hint, got: %s", view)
	}

	// 3. Select option 1 (prune stopped containers)
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	m = res.(Model)
	if m.pruneConfirm != 1 {
		t.Errorf("expected pruneConfirm to be 1, got %d", m.pruneConfirm)
	}
	confirmView := m.View()
	if !strings.Contains(confirmView, "Удаление всех остановленных контейнеров") {
		t.Errorf("expected confirmation prompt in view, got: %s", confirmView)
	}
	if !strings.Contains(confirmView, "[Y]") {
		t.Errorf("expected [Y] confirmation hint, got: %s", confirmView)
	}

	// 4. Cancel confirmation with "n"
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = res.(Model)
	if m.pruneConfirm != 0 {
		t.Errorf("expected pruneConfirm to be 0 after 'n', got %d", m.pruneConfirm)
	}
	if m.activeModal != "prune" {
		t.Errorf("expected activeModal to remain 'prune', got %q", m.activeModal)
	}

	// 5. Select option 4 and confirm with "Y"
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	m = res.(Model)
	if m.pruneConfirm != 4 {
		t.Errorf("expected pruneConfirm to be 4, got %d", m.pruneConfirm)
	}
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to close after confirmation, got %q", m.activeModal)
	}
	if m.pruneConfirm != 0 {
		t.Errorf("expected pruneConfirm to reset to 0, got %d", m.pruneConfirm)
	}

	// 6. Test opening prune modal on Tab 3 (Compose)
	m.activeTab = 3
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")})
	m = res.(Model)
	if m.activeModal != "prune" {
		t.Errorf("expected activeModal 'prune' on tab 3 with 'C', got %q", m.activeModal)
	}

	// 7. Close modal with Esc
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal to close on Esc, got %q", m.activeModal)
	}

	// 8. Check footer hints on tab 2 and tab 3
	m.activeTab = 2
	m.activeModal = ""
	tab2View := m.View()
	if !strings.Contains(tab2View, "[C]") || !strings.Contains(tab2View, "Очистка") {
		t.Errorf("expected [C] Очистка in tab 2 footer, got: %s", tab2View)
	}

	m.activeTab = 3
	tab3View := m.View()
	if !strings.Contains(tab3View, "[C]") || !strings.Contains(tab3View, "Очистка") {
		t.Errorf("expected [C] Очистка in tab 3 footer, got: %s", tab3View)
	}

	// 9. Check help box hint
	m.showHelp = true
	helpView := m.View()
	if !strings.Contains(helpView, "[C]") || !strings.Contains(helpView, "Очистка неиспользуемых ресурсов Docker") {
		t.Errorf("expected [C] hint in help view, got: %s", helpView)
	}
}

func TestRenderSparkline(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		if res := renderSparkline(nil); res != "" {
			t.Errorf("expected empty string for nil history, got %q", res)
		}
		if res := renderSparkline([]float64{}); res != "" {
			t.Errorf("expected empty string for empty history, got %q", res)
		}
	})

	t.Run("boundary and clamped values", func(t *testing.T) {
		history := []float64{-10.0, 0.0, 50.0, 100.0, 150.0}
		res := renderSparkline(history)
		if res == "" {
			t.Fatal("expected non-empty rendered sparkline")
		}
	})

	t.Run("color grading checks", func(t *testing.T) {
		low := renderSparkline([]float64{10.0})
		if !strings.Contains(low, string(sparkChars[0])) {
			t.Errorf("expected low sparkline character, got %q", low)
		}

		med := renderSparkline([]float64{70.0})
		if !strings.Contains(med, string(sparkChars[5])) {
			t.Errorf("expected medium sparkline character, got %q", med)
		}

		high := renderSparkline([]float64{95.0})
		if !strings.Contains(high, string(sparkChars[7])) {
			t.Errorf("expected high sparkline character, got %q", high)
		}
	})
}

func TestRecordMetricHistory(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")
	initialCPULen := len(m.cpuHistory)
	initialMemLen := len(m.memHistory)

	if initialCPULen == 0 || initialMemLen == 0 {
		t.Errorf("expected initial history to have at least 1 point from fetchAllData")
	}

	// Push 40 entries to verify maxHistory truncation (limit = 30)
	for i := 0; i < 40; i++ {
		m.recordMetricHistory(float64(i), float64(i*2))
	}

	if len(m.cpuHistory) != 30 {
		t.Errorf("expected cpuHistory to be capped at 30, got %d", len(m.cpuHistory))
	}
	if len(m.memHistory) != 30 {
		t.Errorf("expected memHistory to be capped at 30, got %d", len(m.memHistory))
	}

	// Verify latest value is at the end
	if m.cpuHistory[29] != 39.0 {
		t.Errorf("expected latest cpu value 39.0, got %f", m.cpuHistory[29])
	}
	if m.memHistory[29] != 78.0 {
		t.Errorf("expected latest mem value 78.0, got %f", m.memHistory[29])
	}
}

func TestHostTabSparklineDisplay(t *testing.T) {
	m := NewModel(2*time.Second, 1, "v0.5.0")
	m.width = 100
	m.height = 30
	m.cpuHistory = []float64{10, 20, 30, 70, 90}
	m.memHistory = []float64{15, 25, 35, 65, 85}

	view := m.View()
	if !strings.Contains(view, "CPU (") || !strings.Contains(view, "RAM (") {
		t.Errorf("expected CPU and RAM in Host Overview view, got: %s", view)
	}
	var hasSparkChar bool
	for _, ch := range sparkChars {
		if strings.ContainsRune(view, ch) {
			hasSparkChar = true
			break
		}
	}
	if !hasSparkChar {
		t.Errorf("expected sparkline characters in view output, got: %s", view)
	}
}
