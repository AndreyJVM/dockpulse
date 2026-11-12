package views

import (
	"strings"
	"testing"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func TestRenderHeaderAndTabs(t *testing.T) {
	dict := i18n.Get()

	// Header
	hdr := RenderHeader(HeaderContext{
		Version:               "v0.6.3",
		LastUpdate:            "12:00:00",
		DockerInfo:            monitor.DockerSummary{RunningContainers: 2, StoppedContainers: 1, ImagesCount: 7, Version: "27.4.0"},
		FormattedDockerStatus: "Running",
		I18n:                  &dict,
	})
	if !strings.Contains(hdr, "Dockpulse") || !strings.Contains(hdr, "v0.6.3") || !strings.Contains(hdr, "12:00:00") {
		t.Errorf("unexpected header output: %s", hdr)
	}
	if !strings.Contains(hdr, "Containers: 2 active, 1 stopped") || !strings.Contains(hdr, "27.4.0") {
		t.Errorf("expected persistent Docker summary, got: %s", hdr)
	}

	// Tabs
	tabs := []string{dict.TabDocker, dict.TabCompose, dict.TabNetworks, dict.TabVolumes, dict.TabSwarm}
	tabBar := RenderTabs(tabs, 1, 80)
	if !strings.Contains(tabBar, "[1: Containers]") {
		t.Errorf("expected active tab bracketed, got: %s", tabBar)
	}
}

func TestRenderFooter(t *testing.T) {
	dict := i18n.Get()

	// Tab 1 footer
	f1 := RenderFooter(FooterContext{ActiveTab: 1, I18n: &dict, Width: 80})
	if !strings.Contains(f1, dict.FooterKill) || !strings.Contains(f1, dict.FooterRemove) {
		t.Errorf("expected Kill and Remove hints in footer 1, got: %s", f1)
	}

	// Tab 1 footer filtering
	f1Filter := RenderFooter(FooterContext{ActiveTab: 1, IsFiltering: true, I18n: &dict, Width: 80})
	if !strings.Contains(f1Filter, dict.FooterApplyFilter) {
		t.Errorf("expected apply filter hint in footer, got: %s", f1Filter)
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

func TestRenderComposeView(t *testing.T) {
	dict := i18n.Get()

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

func TestRenderNetworksView(t *testing.T) {
	dict := i18n.Get()
	view := RenderNetworks(NetworksContext{
		Networks: []monitor.DockerNetwork{{
			ID:             "abc123",
			Name:           "frontend",
			Driver:         "bridge",
			Scope:          "local",
			Containers:     2,
			Subnet:         "172.20.0.0/16",
			Gateway:        "172.20.0.1",
			Attachable:     true,
			ComposeProject: "shop",
		}},
		I18n: &dict,
	})

	for _, expected := range []string{"frontend", "bridge", "172.20.0.0/16", "172.20.0.1", "shop"} {
		if !strings.Contains(view, expected) {
			t.Errorf("expected %q in networks view, got: %s", expected, view)
		}
	}

	errorView := RenderNetworks(NetworksContext{Error: "daemon unavailable", I18n: &dict})
	if !strings.Contains(errorView, "daemon unavailable") {
		t.Errorf("expected discovery error in networks view, got: %s", errorView)
	}
}

func TestRenderVolumesView(t *testing.T) {
	dict := i18n.Get()
	view := RenderVolumes(VolumesContext{
		Volumes: []monitor.DockerVolume{{
			Name:           "shop_data",
			Driver:         "local",
			Scope:          "local",
			Mountpoint:     "/var/lib/docker/volumes/shop_data/_data",
			Size:           4096,
			RefCount:       2,
			ComposeProject: "shop",
		}},
		I18n: &dict,
	})

	for _, expected := range []string{"shop_data", "4.0 KB", "/var/lib/docker/volumes/shop_data/_data", "shop"} {
		if !strings.Contains(view, expected) {
			t.Errorf("expected %q in volumes view, got: %s", expected, view)
		}
	}

	errorView := RenderVolumes(VolumesContext{Error: "daemon unavailable", I18n: &dict})
	if !strings.Contains(errorView, "daemon unavailable") {
		t.Errorf("expected discovery error in volumes view, got: %s", errorView)
	}
}

func TestRenderSwarmView(t *testing.T) {
	dict := i18n.Get()

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
	dict := i18n.Get()

	// Empty
	emptyView := RenderDocker(DockerContext{
		DockerInfo: monitor.DockerSummary{Status: "running"},
		I18n:       &dict,
	})
	if !strings.Contains(emptyView, dict.ContainersNotRunning) {
		t.Errorf("expected no containers running in docker view, got: %s", emptyView)
	}

	// Containers table with details panel
	containers := []monitor.ContainerItem{
		{
			ID:              "c12345678901",
			Name:            "api-server",
			Image:           "golang:alpine",
			State:           "running",
			Health:          "healthy",
			CPUPerc:         12.5,
			MemUsageMB:      250.0,
			MemLimitMB:      1024.0,
			MemPerc:         24.4,
			Ports:           "8080->8080/tcp",
			Mounts:          []string{"/vol -> /app (rw)"},
			ComposeProject:  "my-app",
			ComposeService:  "api",
			NetRxBytes:      1024 * 1024,
			NetTxBytes:      2048 * 1024,
			BlockReadBytes:  512 * 1024,
			BlockWriteBytes: 1536 * 1024,
		},
	}

	view := RenderDocker(DockerContext{
		DockerInfo:           monitor.DockerSummary{Status: "running", Containers: containers, ImagesCount: 5, Version: "27.1.0"},
		Containers:           containers,
		TotalContainersCount: 1,
		CursorIdx:            0,
		Width:                100,
		I18n:                 &dict,
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
	dict := i18n.Get()
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

func TestFormatPortDisplay(t *testing.T) {
	// Empty input
	if got := FormatPortDisplay(""); got != nil {
		t.Errorf("expected nil for empty string, got: %v", got)
	}
	if got := FormatPortDisplay("   "); got != nil {
		t.Errorf("expected nil for whitespace string, got: %v", got)
	}

	// Single exposed port
	if got := FormatPortDisplay("80/tcp"); len(got) != 1 || got[0] != "80/tcp" {
		t.Errorf("expected [80/tcp], got: %v", got)
	}

	// Arrow conversion
	if got := FormatPortDisplay("8080->80/tcp"); len(got) != 1 || got[0] != "8080 → 80/tcp" {
		t.Errorf("expected arrow converted, got: %v", got)
	}

	// IPv6 deduplication
	ipv46 := "0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp, 0.0.0.0:443->443/tcp"
	got := FormatPortDisplay(ipv46)
	if len(got) != 2 {
		t.Fatalf("expected 2 deduped ports, got %d: %v", len(got), got)
	}
	if got[0] != "0.0.0.0:8080 → 8080/tcp" || got[1] != "0.0.0.0:443 → 443/tcp" {
		t.Errorf("unexpected deduped ports: %v", got)
	}
}

func TestRenderContainerDetailsCard_Extended(t *testing.T) {
	dict := i18n.Get()

	// 1. Container with complete extended fields (IP, Network, RestartPolicy, Created)
	cFull := monitor.ContainerItem{
		ID:            "c123456789ab",
		Name:          "prod-api",
		Image:         "api:v2",
		State:         "running",
		Status:        "Up 4 hours",
		IPAddress:     "10.0.1.25",
		NetworkName:   "backend-net",
		RestartPolicy: "always",
		Created:       "2026-11-07 12:00:00",
		Ports:         "0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp",
	}

	viewFull := RenderDocker(DockerContext{
		DockerInfo: monitor.DockerSummary{Status: "running", Containers: []monitor.ContainerItem{cFull}},
		Containers: []monitor.ContainerItem{cFull},
		CursorIdx:  0,
		I18n:       &dict,
		Width:      100,
	})

	if !strings.Contains(viewFull, "10.0.1.25") || !strings.Contains(viewFull, "backend-net") {
		t.Errorf("expected IP and network name in details card, got: %s", viewFull)
	}
	if !strings.Contains(viewFull, "always") {
		t.Errorf("expected restart policy always in details card, got: %s", viewFull)
	}
	if !strings.Contains(viewFull, "2026-11-07 12:00:00") {
		t.Errorf("expected created timestamp in details card, got: %s", viewFull)
	}
	if !strings.Contains(viewFull, "0.0.0.0:8080 → 8080/tcp") {
		t.Errorf("expected arrow formatted port in details card, got: %s", viewFull)
	}

	// 2. Container with on-failure restart policy and no IP
	cOnFailure := monitor.ContainerItem{
		ID:            "c98765432100",
		Name:          "worker-job",
		Image:         "job:latest",
		State:         "stopped",
		Status:        "Exited (1)",
		RestartPolicy: "on-failure",
	}

	viewOnFail := RenderDocker(DockerContext{
		DockerInfo: monitor.DockerSummary{Status: "running", Containers: []monitor.ContainerItem{cOnFailure}},
		Containers: []monitor.ContainerItem{cOnFailure},
		CursorIdx:  0,
		I18n:       &dict,
		Width:      100,
	})

	if !strings.Contains(viewOnFail, "on-failure") {
		t.Errorf("expected on-failure restart policy, got: %s", viewOnFail)
	}
	if !strings.Contains(viewOnFail, "none") {
		t.Errorf("expected none for missing IP, got: %s", viewOnFail)
	}
}
