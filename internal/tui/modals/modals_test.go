package modals

import (
	"strings"
	"testing"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

func TestRenderLogs(t *testing.T) {
	dict := i18n.Get("en")
	
	// Empty logs
	emptyView := RenderLogs(nil, 0, "test-container", "c12345", 100, &dict)
	if !strings.Contains(emptyView, "test-container") || !strings.Contains(emptyView, dict.NoLogsOutput) {
		t.Errorf("expected container name and no logs message, got: %s", emptyView)
	}

	// Logs with scrolling
	lines := []string{"log line 1", "log line 2", "log line 3", "log line 4"}
	view := RenderLogs(lines, 1, "test-container", "c12345", 80, &dict)
	if !strings.Contains(view, "log line 2") {
		t.Errorf("expected 'log line 2' in view, got: %s", view)
	}
	if !strings.Contains(view, "4 lines") && !strings.Contains(view, "4") {
		t.Errorf("expected total lines count in view, got: %s", view)
	}
}

func TestRenderInspect(t *testing.T) {
	dict := i18n.Get("ru")

	detail := monitor.ContainerDetail{
		ID:        "c123456789",
		Name:      "web-service",
		Image:     "nginx:alpine",
		State:     "running",
		Status:    "Up 3 hours",
		Health:    "healthy",
		CPUPerc:   4.5,
		MemUsageMB: 120.0,
		MemPerc:   3.2,
		Created:   "2026-11-06",
		IP:        "172.17.0.2",
		Gateway:   "172.17.0.1",
		Command:   "nginx -g daemon off;",
		Mounts:          []string{"/host/data -> /data (ro)"},
		NetRxBytes:      5 * 1024 * 1024,
		NetTxBytes:      10 * 1024 * 1024,
		BlockReadBytes:  2 * 1024 * 1024,
		BlockWriteBytes: 4 * 1024 * 1024,
	}

	view := RenderInspect(detail, &dict)
	if !strings.Contains(view, "web-service") || !strings.Contains(view, "[healthy]") {
		t.Errorf("expected container name and healthy badge, got: %s", view)
	}
	if !strings.Contains(view, "172.17.0.2") || !strings.Contains(view, "/host/data") {
		t.Errorf("expected IP address and mounts in inspect view, got: %s", view)
	}
	if !strings.Contains(view, "Net I/O:") || !strings.Contains(view, "5.0 MB / 10.0 MB") {
		t.Errorf("expected Net I/O in inspect view, got: %s", view)
	}
	if !strings.Contains(view, "Block I/O:") || !strings.Contains(view, "2.0 MB / 4.0 MB") {
		t.Errorf("expected Block I/O in inspect view, got: %s", view)
	}
}

func TestRenderPrune(t *testing.T) {
	dict := i18n.Get("en")

	// Error view
	errView := RenderPrune(monitor.DockerDiskUsage{}, "daemon error", 0, &dict)
	if !strings.Contains(errView, "daemon error") {
		t.Errorf("expected error message in prune view, got: %s", errView)
	}

	// Menu view
	usage := monitor.DockerDiskUsage{
		Images:     monitor.DockerDiskItem{TotalCount: 5, ActiveCount: 2, TotalSize: 1024 * 1024 * 500, Reclaimable: 1024 * 1024 * 200},
		Containers: monitor.DockerDiskItem{TotalCount: 3, ActiveCount: 1, TotalSize: 1024 * 1024 * 50, Reclaimable: 1024 * 1024 * 30},
		TotalSize:  1024 * 1024 * 550,
		TotalRecl:  1024 * 1024 * 230,
	}
	menuView := RenderPrune(usage, "", 0, &dict)
	if !strings.Contains(menuView, dict.PruneTitle) || !strings.Contains(menuView, "[1]") {
		t.Errorf("expected prune menu items, got: %s", menuView)
	}

	// Confirmation step view
	confirmView := RenderPrune(usage, "", 1, &dict)
	if !strings.Contains(confirmView, dict.PruneWarnTitle) || !strings.Contains(confirmView, "[Y]") {
		t.Errorf("expected prune confirmation view, got: %s", confirmView)
	}
}

func TestRenderComposeDown(t *testing.T) {
	dict := i18n.Get("en")
	view := RenderComposeDown("my-stack", &dict)
	if !strings.Contains(view, "my-stack") || !strings.Contains(view, dict.ComposeDownTitle) {
		t.Errorf("expected project name and title in compose down modal, got: %s", view)
	}
}

func TestRenderRemoveContainer(t *testing.T) {
	dict := i18n.Get("ru")
	view := RenderRemoveContainer("my-db", "c98765", &dict)
	if !strings.Contains(view, "my-db") || !strings.Contains(view, "c98765") || !strings.Contains(view, dict.RemoveContainerTitle) {
		t.Errorf("expected container name, id and title in remove modal, got: %s", view)
	}
}

func TestRenderHelp(t *testing.T) {
	dict := i18n.Get("en")
	view := RenderHelp(2*time.Second, &dict)
	if !strings.Contains(view, dict.HelpTitle) || !strings.Contains(view, "[K]") || !strings.Contains(view, "[x]") {
		t.Errorf("expected help title and new shortcut keys [K] and [x], got: %s", view)
	}
}
