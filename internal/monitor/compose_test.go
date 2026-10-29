package monitor

import (
	"strings"
	"testing"
)

func TestGetComposeProjects(t *testing.T) {
	t.Run("empty and nil container list", func(t *testing.T) {
		if len(GetComposeProjects(nil)) != 0 {
			t.Errorf("expected empty result for nil containers")
		}
		if len(GetComposeProjects([]ContainerItem{})) != 0 {
			t.Errorf("expected empty result for empty container slice")
		}
	})

	t.Run("skip non-compose containers", func(t *testing.T) {
		containers := []ContainerItem{
			{ID: "1", Name: "standalone-1", ComposeProject: ""},
			{ID: "2", Name: "standalone-2", ComposeProject: ""},
		}
		projects := GetComposeProjects(containers)
		if len(projects) != 0 {
			t.Fatalf("expected 0 compose projects, got %d", len(projects))
		}
	})

	t.Run("service name fallback to container name", func(t *testing.T) {
		containers := []ContainerItem{
			{
				ID:             "c1",
				Name:           "fallback-container-name",
				ComposeProject: "proj-fallback",
				ComposeService: "", // empty service name
				State:          "running",
			},
		}
		projects := GetComposeProjects(containers)
		if len(projects) != 1 {
			t.Fatalf("expected 1 project, got %d", len(projects))
		}
		if len(projects[0].Services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(projects[0].Services))
		}
		if projects[0].Services[0].Name != "fallback-container-name" {
			t.Errorf("expected service name to fallback to container name, got %q", projects[0].Services[0].Name)
		}
	})

	t.Run("project statuses: running, partial, stopped", func(t *testing.T) {
		containers := []ContainerItem{
			// All running -> "running" / "Запущен"
			{ID: "1", Name: "app-web", ComposeProject: "alpha-stack", ComposeService: "web", State: "running"},
			{ID: "2", Name: "app-db", ComposeProject: "alpha-stack", ComposeService: "db", State: "running"},

			// Partial -> "partial" / "Частично"
			{ID: "3", Name: "beta-api", ComposeProject: "beta-stack", ComposeService: "api", State: "running"},
			{ID: "4", Name: "beta-worker", ComposeProject: "beta-stack", ComposeService: "worker", State: "exited"},

			// All stopped -> "stopped" / "Остановлен"
			{ID: "5", Name: "gamma-redis", ComposeProject: "gamma-stack", ComposeService: "redis", State: "exited"},
			{ID: "6", Name: "gamma-worker", ComposeProject: "gamma-stack", ComposeService: "worker", State: "created"},
		}

		projects := GetComposeProjects(containers)
		if len(projects) != 3 {
			t.Fatalf("expected 3 projects, got %d", len(projects))
		}

		expected := map[string]struct {
			status        string
			statusDisplay string
			total         int
			running       int
		}{
			"alpha-stack": {status: "running", statusDisplay: "Запущен", total: 2, running: 2},
			"beta-stack":  {status: "partial", statusDisplay: "Частично", total: 2, running: 1},
			"gamma-stack": {status: "stopped", statusDisplay: "Остановлен", total: 2, running: 0},
		}

		for _, p := range projects {
			exp, ok := expected[p.Name]
			if !ok {
				t.Errorf("unexpected project name %q", p.Name)
				continue
			}
			if p.Status != exp.status {
				t.Errorf("project %q: expected status %q, got %q", p.Name, exp.status, p.Status)
			}
			if p.StatusDisplay != exp.statusDisplay {
				t.Errorf("project %q: expected statusDisplay %q, got %q", p.Name, exp.statusDisplay, p.StatusDisplay)
			}
			if p.TotalServices != exp.total {
				t.Errorf("project %q: expected total %d, got %d", p.Name, exp.total, p.TotalServices)
			}
			if p.RunningServices != exp.running {
				t.Errorf("project %q: expected running %d, got %d", p.Name, exp.running, p.RunningServices)
			}
		}
	})

	t.Run("alphabetical sorting of projects", func(t *testing.T) {
		containers := []ContainerItem{
			{ID: "1", Name: "z", ComposeProject: "zebra", ComposeService: "svc"},
			{ID: "2", Name: "a", ComposeProject: "alpha", ComposeService: "svc"},
			{ID: "3", Name: "m", ComposeProject: "middle", ComposeService: "svc"},
		}

		projects := GetComposeProjects(containers)
		if len(projects) != 3 {
			t.Fatalf("expected 3 projects, got %d", len(projects))
		}
		if projects[0].Name != "alpha" || projects[1].Name != "middle" || projects[2].Name != "zebra" {
			t.Errorf("expected sorted [alpha, middle, zebra], got [%s, %s, %s]",
				projects[0].Name, projects[1].Name, projects[2].Name)
		}
	})

	t.Run("preserves all service attributes", func(t *testing.T) {
		containers := []ContainerItem{
			{
				ID:             "cid1",
				Name:           "my-container",
				Image:          "custom-image:v1",
				State:          "running",
				Status:         "Up 2 hours",
				Ports:          "0.0.0.0:80->80/tcp",
				ComposeProject: "my-stack",
				ComposeService: "web-svc",
			},
		}

		projects := GetComposeProjects(containers)
		if len(projects) != 1 || len(projects[0].Services) != 1 {
			t.Fatalf("expected 1 project with 1 service")
		}

		svc := projects[0].Services[0]
		if svc.Name != "web-svc" {
			t.Errorf("expected service name 'web-svc', got %q", svc.Name)
		}
		if svc.ContainerName != "my-container" {
			t.Errorf("expected container name 'my-container', got %q", svc.ContainerName)
		}
		if svc.Image != "custom-image:v1" {
			t.Errorf("expected image 'custom-image:v1', got %q", svc.Image)
		}
		if svc.State != "running" {
			t.Errorf("expected state 'running', got %q", svc.State)
		}
		if svc.Status != "Up 2 hours" {
			t.Errorf("expected status 'Up 2 hours', got %q", svc.Status)
		}
		if svc.Ports != "0.0.0.0:80->80/tcp" {
			t.Errorf("expected ports '0.0.0.0:80->80/tcp', got %q", svc.Ports)
		}
	})
}

func TestGetComposeProjectLogs(t *testing.T) {
	t.Run("non-existent project returns error", func(t *testing.T) {
		_, err := GetComposeProjectLogs("non-existent-proj", nil, 50)
		if err == nil {
			t.Fatal("expected error for non-existent project, got nil")
		}
		if !strings.Contains(err.Error(), "не найдены") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("fallback with empty logs message", func(t *testing.T) {
		containers := []ContainerItem{
			{
				ID:             "mock-invalid-cid-1",
				Name:           "web-1",
				ComposeProject: "mock-proj",
			},
		}

		logs, err := GetComposeProjectLogs("mock-proj", containers, 0) // tailLines <= 0 defaults to 100
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(logs, "Логи сервисов проекта пусты") && !strings.Contains(logs, "web-1") {
			t.Errorf("expected fallback log message, got: %q", logs)
		}
	})

	t.Run("many containers per-container tail clamp", func(t *testing.T) {
		var containers []ContainerItem
		for i := 0; i < 10; i++ {
			containers = append(containers, ContainerItem{
				ID:             "mock-c",
				Name:           "worker",
				ComposeProject: "large-stack",
			})
		}

		logs, err := GetComposeProjectLogs("large-stack", containers, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if logs == "" {
			t.Error("expected non-empty response string")
		}
	})
}

func TestRestartComposeProject(t *testing.T) {
	t.Run("non-existent project returns error", func(t *testing.T) {
		err := RestartComposeProject("non-existent-proj", nil)
		if err == nil {
			t.Fatal("expected error for non-existent project, got nil")
		}
		if !strings.Contains(err.Error(), "не найдено контейнеров") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("fallback restarts invalid containers with error aggregation", func(t *testing.T) {
		containers := []ContainerItem{
			{
				ID:             "non-existent-id-001",
				Name:           "broken-web",
				ComposeProject: "failing-stack",
			},
		}

		err := RestartComposeProject("failing-stack", containers)
		if err == nil {
			t.Fatal("expected error when restarting non-existent container ID, got nil")
		}
		if !strings.Contains(err.Error(), "ошибки перезапуска") {
			t.Errorf("expected 'ошибки перезапуска' in error message, got: %v", err)
		}
	})
}
