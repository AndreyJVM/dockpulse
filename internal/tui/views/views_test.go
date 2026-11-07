package views

import (
	"strings"
	"testing"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func TestRenderHeaderAndTabs(t *testing.T) {
	dict := i18n.Get("en")

	// Header
	hdr := RenderHeader("v0.6.3", "12:00:00")
	if !strings.Contains(hdr, "Dockpulse") || !strings.Contains(hdr, "v0.6.3") || !strings.Contains(hdr, "12:00:00") {
		t.Errorf("unexpected header output: %s", hdr)
	}

	// Tabs
	tabs := []string{dict.TabHost, dict.TabDocker, dict.TabCompose, dict.TabSwarm}
	tabBar := RenderTabs(tabs, 2, 80)
	if !strings.Contains(tabBar, "[2: Docker Containers]") {
		t.Errorf("expected active tab bracketed, got: %s", tabBar)
	}
}

func TestRenderFooter(t *testing.T) {
	dict := i18n.Get("en")

	// Tab 1 footer
	f1 := RenderFooter(FooterContext{ActiveTab: 1, I18n: &dict, Width: 80})
	if !strings.Contains(f1, dict.FooterTabs) {
		t.Errorf("expected tabs hint in footer 1, got: %s", f1)
	}

	// Tab 2 footer normal
	f2 := RenderFooter(FooterContext{ActiveTab: 2, I18n: &dict, Width: 80})
	if !strings.Contains(f2, dict.FooterKill) || !strings.Contains(f2, dict.FooterRemove) {
		t.Errorf("expected Kill and Remove hints in footer 2, got: %s", f2)
	}

	// Tab 2 footer filtering
	f2Filter := RenderFooter(FooterContext{ActiveTab: 2, IsFiltering: true, I18n: &dict, Width: 80})
	if !strings.Contains(f2Filter, dict.FooterApplyFilter) {
		t.Errorf("expected apply filter hint in footer, got: %s", f2Filter)
	}

	// Modal footers
	fDown := RenderFooter(FooterContext{ActiveModal: "compose_down", I18n: &dict, Width: 80})
	if !strings.Contains(fDown, "(Down)") {
		t.Errorf("expected compose down modal footer, got: %s", fDown)
	}

	fRemove := RenderFooter(FooterContext{ActiveModal: "remove_container", I18n: &dict, Width: 80})
	if !strings.Contains(fRemove, "(Remove)") {
		t.Errorf("expected remove container modal footer, got: %s", fRemove)
	}
}

func TestRenderHostView(t *testing.T) {
	dict := i18n.Get("en")
	view := RenderHost(HostContext{
		SysInfo:               monitor.SystemInfo{Hostname: "my-host", OS: "Linux", Kernel: "6.8.0"},
		CPUInfo:               monitor.CPUInfo{Cores: 4, Percent: 25.0},
		MemInfo:               monitor.MemoryInfo{TotalGB: 16.0, UsedGB: 4.0, Percent: 25.0},
		DockerInfo:            monitor.DockerSummary{Status: "running", RunningContainers: 2, StoppedContainers: 1, ImagesCount: 10, Version: "27.0"},
		CPUHistory:            []float64{10.0, 20.0, 25.0},
		MemHistory:            []float64{20.0, 25.0, 25.0},
		FormattedUptime:       "02:45",
		FormattedDockerStatus: "Running",
		I18n:                  &dict,
	})

	if !strings.Contains(view, "my-host") || !strings.Contains(view, "Running") || !strings.Contains(view, "Containers: 2 active, 1 stopped") {
		t.Errorf("expected host view details, got: %s", view)
	}
}

func TestRenderComposeView(t *testing.T) {
	dict := i18n.Get("en")

	// Empty
	emptyView := RenderCompose(ComposeContext{Projects: nil, I18n: &dict})
	if !strings.Contains(emptyView, dict.NoComposeProjects) {
		t.Errorf("expected empty compose message, got: %s", emptyView)
	}

	// With projects
	projs := []monitor.ComposeProject{
		{
			Name:            "web-stack",
			Status:          "running",
			TotalServices:   1,
			RunningServices: 1,
			Services: []monitor.ComposeService{
				{Name: "web", State: "running", Image: "nginx:alpine", Ports: "80:80"},
			},
		},
	}
	view := RenderCompose(ComposeContext{
		Projects:     projs,
		CursorIdx:    0,
		FormatStatus: func(s, _ string) string { return s },
		I18n:         &dict,
	})
	if !strings.Contains(view, "web-stack") || !strings.Contains(view, "nginx:alpine") {
		t.Errorf("expected project details in compose view, got: %s", view)
	}
}

func TestRenderSwarmView(t *testing.T) {
	dict := i18n.Get("en")

	// Inactive
	inactView := RenderSwarm(SwarmContext{
		Swarm:        monitor.SwarmSummary{IsActive: false},
		FormatStatus: func(_ bool) string { return "Inactive" },
		I18n:         &dict,
	})
	if !strings.Contains(inactView, dict.SwarmInactive) {
		t.Errorf("expected swarm inactive notice, got: %s", inactView)
	}

	// Active
	actView := RenderSwarm(SwarmContext{
		Swarm: monitor.SwarmSummary{
			IsActive:      true,
			NodesCount:    3,
			ManagersCount: 1,
			Nodes: []monitor.SwarmNode{
				{ID: "node1", Hostname: "worker-1", Role: "worker", State: "ready", Availability: "active"},
			},
			Services: []monitor.SwarmService{
				{ID: "svc1", Name: "db-svc", Image: "postgres:16", Mode: "replicated", Replicas: "1/1", RunningReplicas: 1, TargetReplicas: 1},
			},
		},
		CursorIdx:    0,
		FormatStatus: func(_ bool) string { return "Active" },
		I18n:         &dict,
	})
	if !strings.Contains(actView, "worker-1") || !strings.Contains(actView, "db-svc") {
		t.Errorf("expected swarm nodes and services in view, got: %s", actView)
	}
}

func TestRenderDockerView(t *testing.T) {
	dict := i18n.Get("en")

	// Empty
	emptyView := RenderDocker(DockerContext{
		DockerInfo:            monitor.DockerSummary{Status: "running"},
		FormattedDockerStatus: "Running",
		I18n:                  &dict,
	})
	if !strings.Contains(emptyView, dict.ContainersNotRunning) {
		t.Errorf("expected no containers running in docker view, got: %s", emptyView)
	}

	// Containers table with details panel
	containers := []monitor.ContainerItem{
		{
			ID:             "c12345678901",
			Name:           "api-server",
			Image:          "golang:alpine",
			State:          "running",
			Health:         "healthy",
			CPUPerc:        12.5,
			MemUsageMB:     250.0,
			MemLimitMB:     1024.0,
			MemPerc:        24.4,
			Ports:          "8080->8080/tcp",
			Mounts:         []string{"/vol -> /app (rw)"},
			ComposeProject: "my-app",
			ComposeService: "api",
			NetRxBytes:     1024 * 1024,
			NetTxBytes:     2048 * 1024,
			BlockReadBytes:  512 * 1024,
			BlockWriteBytes: 1536 * 1024,
		},
	}

	view := RenderDocker(DockerContext{
		DockerInfo:            monitor.DockerSummary{Status: "running", Containers: containers, ImagesCount: 5, Version: "27.1.0"},
		Containers:            containers,
		TotalContainersCount:  1,
		CursorIdx:             0,
		Width:                 100,
		FormattedDockerStatus: "Running",
		I18n:                  &dict,
	})

	if !strings.Contains(view, "api-server") || !strings.Contains(view, "[healthy]") || !strings.Contains(view, "8080->8080/tcp") {
		t.Errorf("expected container row and details in docker view, got: %s", view)
	}
	if !strings.Contains(view, "/vol -> /app (rw)") || !strings.Contains(view, "my-app") {
		t.Errorf("expected mounts and compose info in details panel, got: %s", view)
	}
	if !strings.Contains(view, "Net I/O:") || !strings.Contains(view, "1.0 MB / 2.0 MB") {
		t.Errorf("expected Net I/O in details panel, got: %s", view)
	}
	if !strings.Contains(view, "Block I/O:") || !strings.Contains(view, "512.0 KB / 1.5 MB") {
		t.Errorf("expected Block I/O in details panel, got: %s", view)
	}
}

func TestFormatContainerStatus(t *testing.T) {
	tests := []struct {
		state    string
		health   string
		expected string
	}{
		{"running", "healthy", "● [healthy]"},
		{"running", "unhealthy", "▲ [unhealthy]"},
		{"running", "starting", "⏳ [starting]"},
		{"running", "", "● running"},
		{"paused", "", "⏸ paused"},
		{"restarting", "", "↻ restarting"},
		{"exited", "", "○ exited"},
		{"", "", "○ stopped"},
	}

	for _, tc := range tests {
		text, _ := formatContainerStatus(tc.state, tc.health)
		if text != tc.expected {
			t.Errorf("formatContainerStatus(%q, %q) = %q, expected %q", tc.state, tc.health, text, tc.expected)
		}
	}
}

func TestFormatContainerMetrics(t *testing.T) {
	// Stopped container
	cpu, mem := formatContainerMetrics(monitor.ContainerItem{State: "exited"})
	if cpu != "-" || mem != "-" {
		t.Errorf("expected - and - for stopped container, got %s and %s", cpu, mem)
	}

	// Running with limit
	cpu, mem = formatContainerMetrics(monitor.ContainerItem{
		State:      "running",
		CPUPerc:    15.5,
		MemUsageMB: 120.0,
		MemLimitMB: 512.0,
		MemPerc:    23.4,
	})
	if cpu != "15.5%" || !strings.Contains(mem, "120.0M (23.4%)") {
		t.Errorf("unexpected metrics format: %s, %s", cpu, mem)
	}

	// Running without limit (MemLimitMB == 0)
	_, memNoLimit := formatContainerMetrics(monitor.ContainerItem{
		State:      "running",
		MemUsageMB: 64.0,
		MemLimitMB: 0,
	})
	if memNoLimit != "64.0M" {
		t.Errorf("expected 64.0M, got: %s", memNoLimit)
	}
}

func TestStyleCPU(t *testing.T) {
	// Stopped container returns DimStyle
	stopped := styleCPU(99.0, "exited", "  99.0%")
	if !strings.Contains(stopped, "99.0%") {
		t.Errorf("expected formatted string in output, got: %s", stopped)
	}

	// Running >= 80
	high := styleCPU(85.0, "running", "  85.0%")
	if !strings.Contains(high, "85.0%") {
		t.Errorf("expected formatted string in output, got: %s", high)
	}

	// Running >= 50
	mid := styleCPU(60.0, "running", "  60.0%")
	if !strings.Contains(mid, "60.0%") {
		t.Errorf("expected formatted string in output, got: %s", mid)
	}

	// Running normal
	low := styleCPU(20.0, "running", "  20.0%")
	if !strings.Contains(low, "20.0%") {
		t.Errorf("expected formatted string in output, got: %s", low)
	}
}

func TestRenderDockerView_EdgeCases(t *testing.T) {
	// 1. Daemon down
	downView := RenderDocker(DockerContext{
		DockerInfo: monitor.DockerSummary{Status: "stopped"},
	})
	if !strings.Contains(downView, "Docker") {
		t.Errorf("expected daemon down notice, got: %s", downView)
	}

	// 2. Filter active with results and sorting badge
	dict := i18n.Get("en")
	containers := []monitor.ContainerItem{
		{ID: "c1", Name: "app", State: "running", Ports: "80:80"},
	}
	filteredView := RenderDocker(DockerContext{
		DockerInfo:           monitor.DockerSummary{Status: "running", Containers: containers},
		Containers:           containers,
		TotalContainersCount: 1,
		CursorIdx:            -1, // Edge case: invalid negative cursor should not panic
		FilterQuery:          "app",
		SortMode:             1,
		SortNames:            []string{"Default", "CPU"},
		I18n:                 &dict,
	})
	if !strings.Contains(filteredView, "Filter:") || !strings.Contains(filteredView, "Sort:") {
		t.Errorf("expected filter and sort badges, got: %s", filteredView)
	}

	// 3. Filter mode active (IsFiltering == true)
	filteringView := RenderDocker(DockerContext{
		DockerInfo:  monitor.DockerSummary{Status: "running", Containers: containers},
		Containers:  containers,
		IsFiltering: true,
		FilterQuery: "ap",
		I18n:        &dict,
	})
	if !strings.Contains(filteringView, "[Search: ap") {
		t.Errorf("expected search prompt in filtering view, got: %s", filteringView)
	}

	// 4. No containers matched filter
	noMatchView := RenderDocker(DockerContext{
		DockerInfo:  monitor.DockerSummary{Status: "running", Containers: containers},
		Containers:  nil,
		FilterQuery: "xyz",
		I18n:        &dict,
	})
	if !strings.Contains(noMatchView, "No containers found matching") {
		t.Errorf("expected no match notice, got: %s", noMatchView)
	}

	// 5. Cursor out of bounds should not panic, and nil I18n safely handled
	outOfBoundsView := RenderDocker(DockerContext{
		DockerInfo: monitor.DockerSummary{Status: "running", Containers: containers},
		Containers: containers,
		CursorIdx:  999,
		I18n:       nil,
	})
	if !strings.Contains(outOfBoundsView, "app") {
		t.Errorf("expected container row in view, got: %s", outOfBoundsView)
	}
}
