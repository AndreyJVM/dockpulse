package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

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
	if !strings.Contains(view, m.i18n.ColStatus) {
		t.Errorf("expected %s column in header, got: %s", m.i18n.ColStatus, view)
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

func TestDockerTabSelectedContainerDetails(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.6.2", "en")
	m.activeTab = 2
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{
			ID:             "abcdef123456",
			Name:           "web-server",
			Image:          "nginx:alpine",
			State:          "running",
			Ports:          "8080->80/tcp, 8443->443/tcp",
			Mounts:         []string{"/var/www -> /usr/share/nginx/html (ro)", "nginx-logs -> /var/log/nginx (rw)"},
			ComposeProject: "my-stack",
			ComposeService: "web",
			IPAddress:      "172.18.0.5",
			NetworkName:    "my-network",
			RestartPolicy:  "unless-stopped",
			Created:        "2026-11-07 15:30:00",
			NetRxBytes:     12 * 1024 * 1024,
			NetTxBytes:     45 * 1024 * 1024,
			BlockReadBytes:  10 * 1024 * 1024,
			BlockWriteBytes: 20 * 1024 * 1024,
		},
	}

	view := m.View()
	if !strings.Contains(view, "SELECTED CONTAINER DETAILS") {
		t.Errorf("expected details panel header in View(), got: %s", view)
	}
	if !strings.Contains(view, "8080 → 80/tcp") || !strings.Contains(view, "8443 → 443/tcp") {
		t.Errorf("expected formatted ports in details panel, got: %s", view)
	}
	if !strings.Contains(view, "172.18.0.5") || !strings.Contains(view, "my-network") {
		t.Errorf("expected IP and network in details panel, got: %s", view)
	}
	if !strings.Contains(view, "unless-stopped") {
		t.Errorf("expected restart policy in details panel, got: %s", view)
	}
	if !strings.Contains(view, "2026-11-07 15:30:00") {
		t.Errorf("expected created timestamp in details panel, got: %s", view)
	}
	if !strings.Contains(view, "/var/www -> /usr/share/nginx/html (ro)") {
		t.Errorf("expected volume mount in details panel, got: %s", view)
	}
	if !strings.Contains(view, "my-stack") {
		t.Errorf("expected compose project in details panel, got: %s", view)
	}
	if !strings.Contains(view, "Net I/O:") || !strings.Contains(view, "12.0 MB / 45.0 MB") {
		t.Errorf("expected Net I/O in details panel, got: %s", view)
	}
	if !strings.Contains(view, "Block I/O:") || !strings.Contains(view, "10.0 MB / 20.0 MB") {
		t.Errorf("expected Block I/O in details panel, got: %s", view)
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

func TestContainerExecHotkeysAndMessages(t *testing.T) {
	// Test when Docker CLI is not available
	restoreUnavailable := monitor.SetDockerCLIAvailableCheckerForTesting(func() bool { return false })
	mUnavailable := NewModel(2*time.Second, 2, "v0.5.0")
	mUnavailable.dockerInfo = monitor.DockerSummary{
		Status:        "running",
		StatusDisplay: "[Запущен]",
		Containers: []monitor.ContainerItem{
			{ID: "c1", Name: "running-container", State: "running", Status: "Up 2 hours"},
		},
	}
	mUnavailable.cursorIdx = 0
	resUnavail, _ := mUnavailable.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	mUnavailResult := resUnavail.(Model)
	if !strings.Contains(mUnavailResult.actionMessage, "утилита docker не найдена") {
		t.Errorf("expected docker not found message, got %q", mUnavailResult.actionMessage)
	}
	restoreUnavailable()

	// Test when Docker CLI is available
	restoreAvailable := monitor.SetDockerCLIAvailableCheckerForTesting(func() bool { return true })
	defer restoreAvailable()

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

	// 1. Press 'e' on running container opens exec modal
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = res.(Model)
	if m.activeModal != "exec" {
		t.Fatalf("expected activeModal 'exec', got %q", m.activeModal)
	}
	if m.execTargetID != "c1" {
		t.Errorf("expected execTargetID 'c1', got %q", m.execTargetID)
	}
	// Select first option (auto bash) via enter
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd from tea.ExecProcess on selecting option")
	}

	// 2. Press 'E' on stopped container
	m.activeModal = ""
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
	m.activeModal = ""
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

func TestContainerSorting(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.6.3", "en")
	m.activeTab = 2
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "zeta-app", State: "running", CPUPerc: 10.0, MemUsageMB: 200.0, Health: "healthy"},
		{ID: "c2", Name: "alpha-app", State: "running", CPUPerc: 50.0, MemUsageMB: 50.0, Health: "unhealthy"},
		{ID: "c3", Name: "beta-app", State: "stopped", CPUPerc: 0.0, MemUsageMB: 0.0},
		{ID: "c4", Name: "gamma-app", State: "running", CPUPerc: 30.0, MemUsageMB: 500.0, Health: "starting"},
	}

	// 1. Default order
	list := m.getFilteredContainers()
	if list[0].Name != "zeta-app" || list[1].Name != "alpha-app" {
		t.Errorf("expected default order preserved, got: %s, %s", list[0].Name, list[1].Name)
	}

	// 2. Press O -> CPU desc
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newModel.(Model)
	if m.containerSortMode != SortCPU {
		t.Fatalf("expected SortCPU, got %d", m.containerSortMode)
	}
	list = m.getFilteredContainers()
	if list[0].Name != "alpha-app" || list[0].CPUPerc != 50.0 {
		t.Errorf("expected top CPU alpha-app, got %s", list[0].Name)
	}
	if list[1].Name != "gamma-app" || list[1].CPUPerc != 30.0 {
		t.Errorf("expected 2nd CPU gamma-app, got %s", list[1].Name)
	}
	view := m.View()
	if !strings.Contains(view, "Sort: CPU ↓") {
		t.Errorf("expected Sort: CPU ↓ in View(), got: %s", view)
	}

	// 3. Press O -> MEM desc
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newModel.(Model)
	if m.containerSortMode != SortMem {
		t.Fatalf("expected SortMem, got %d", m.containerSortMode)
	}
	list = m.getFilteredContainers()
	if list[0].Name != "gamma-app" || list[0].MemUsageMB != 500.0 {
		t.Errorf("expected top MEM gamma-app, got %s", list[0].Name)
	}

	// 4. Press O -> Name asc
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newModel.(Model)
	if m.containerSortMode != SortName {
		t.Fatalf("expected SortName, got %d", m.containerSortMode)
	}
	list = m.getFilteredContainers()
	if list[0].Name != "alpha-app" || list[1].Name != "beta-app" || list[2].Name != "gamma-app" || list[3].Name != "zeta-app" {
		t.Errorf("unexpected name sorting order: %v", list)
	}

	// 5. Press O -> Status
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newModel.(Model)
	if m.containerSortMode != SortStatus {
		t.Fatalf("expected SortStatus, got %d", m.containerSortMode)
	}
	list = m.getFilteredContainers()
	if list[0].Name != "alpha-app" {
		t.Errorf("expected unhealthy alpha-app first, got %s", list[0].Name)
	}
	if list[1].Name != "gamma-app" {
		t.Errorf("expected starting gamma-app second, got %s", list[1].Name)
	}

	// 6. Press O -> Wrap around to Default
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newModel.(Model)
	if m.containerSortMode != SortDefault {
		t.Fatalf("expected wrap around to SortDefault, got %d", m.containerSortMode)
	}
}

func TestContainerKillAndRemoveWorkflow(t *testing.T) {
	m := NewModel(2*time.Second, 2, "v0.6.3", "en")
	m.activeTab = 2
	m.dockerInfo.Containers = []monitor.ContainerItem{
		{ID: "c1", Name: "app-run", State: "running"},
		{ID: "c2", Name: "app-stop", State: "stopped"},
	}
	m.cursorIdx = 0

	// 1. Press K on running container triggers kill toast & cmd
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	m = res.(Model)
	if cmd == nil {
		t.Errorf("expected non-nil cmd for killContainerCmd")
	}
	if !strings.Contains(m.actionMessage, "Force killing app-run") {
		t.Errorf("expected toast 'Force killing app-run', got: %s", m.actionMessage)
	}

	// 2. Press K on stopped container shows cannot connect toast
	m.cursorIdx = 1
	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	m = res.(Model)
	if cmd != nil {
		t.Errorf("expected nil cmd when killing stopped container")
	}
	if !strings.Contains(m.actionMessage, "Cannot connect") {
		t.Errorf("expected 'Cannot connect' toast, got: %s", m.actionMessage)
	}

	// 3. Press x on container opens remove confirmation modal
	m.cursorIdx = 1
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = res.(Model)
	if m.activeModal != "remove_container" {
		t.Fatalf("expected activeModal 'remove_container', got: %s", m.activeModal)
	}
	if m.removeContainerTargetID != "c2" || m.removeContainerTargetName != "app-stop" {
		t.Errorf("unexpected remove targets: id=%s name=%s", m.removeContainerTargetID, m.removeContainerTargetName)
	}

	// 4. Modal view verification
	view := m.View()
	if !strings.Contains(view, "CONFIRM CONTAINER REMOVAL") || !strings.Contains(view, "app-stop") {
		t.Errorf("expected remove modal content, got: %s", view)
	}

	// 5. Cancel removal with n
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal closed after 'n', got: %s", m.activeModal)
	}

	// 6. Press d on container opens remove modal and confirm with y
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = res.(Model)
	if m.activeModal != "remove_container" {
		t.Fatalf("expected activeModal 'remove_container' after 'd', got: %s", m.activeModal)
	}

	res, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = res.(Model)
	if m.activeModal != "" {
		t.Errorf("expected activeModal closed after 'y', got: %s", m.activeModal)
	}
	if cmd == nil {
		t.Errorf("expected non-nil cmd for removeContainerCmd")
	}
	if !strings.Contains(m.actionMessage, "Removing container app-stop") {
		t.Errorf("expected remove toast, got: %s", m.actionMessage)
	}
}
