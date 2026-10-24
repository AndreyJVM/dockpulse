package monitor

import (
	"strings"
	"testing"
)

func TestGetComposeProjects(t *testing.T) {
	containers := []ContainerItem{
		{
			ID:             "1",
			Name:           "web",
			Image:          "nginx:alpine",
			State:          "running",
			ComposeProject: "my-app",
			ComposeService: "web",
		},
		{
			ID:             "2",
			Name:           "db",
			Image:          "postgres:15",
			State:          "running",
			ComposeProject: "my-app",
			ComposeService: "db",
		},
		{
			ID:             "3",
			Name:           "redis",
			Image:          "redis:alpine",
			State:          "exited",
			ComposeProject: "cache-stack",
			ComposeService: "redis",
		},
		{
			ID:    "4",
			Name:  "standalone-job",
			Image: "busybox",
			State: "running",
		},
	}

	projects := GetComposeProjects(containers)

	if len(projects) != 2 {
		t.Fatalf("expected 2 compose projects, got %d", len(projects))
	}

	// sorted alphabetically: cache-stack, my-app
	if projects[0].Name != "cache-stack" {
		t.Errorf("expected first project to be 'cache-stack', got '%s'", projects[0].Name)
	}
	if projects[0].TotalServices != 1 || projects[0].RunningServices != 0 || projects[0].Status != "stopped" {
		t.Errorf("unexpected status for cache-stack: %+v", projects[0])
	}

	if projects[1].Name != "my-app" {
		t.Errorf("expected second project to be 'my-app', got '%s'", projects[1].Name)
	}
	if projects[1].TotalServices != 2 || projects[1].RunningServices != 2 || projects[1].Status != "running" {
		t.Errorf("unexpected status for my-app: %+v", projects[1])
	}
}

func TestGetComposeProjectsPartialAndEmpty(t *testing.T) {
	// Empty list
	if len(GetComposeProjects(nil)) != 0 {
		t.Errorf("expected empty list for nil containers")
	}

	// Partial running project
	containers := []ContainerItem{
		{
			ID:             "1",
			Name:           "api",
			Image:          "api:latest",
			State:          "running",
			ComposeProject: "backend",
			ComposeService: "api",
		},
		{
			ID:             "2",
			Name:           "db",
			Image:          "postgres:15",
			State:          "exited",
			ComposeProject: "backend",
			ComposeService: "db",
		},
	}

	projects := GetComposeProjects(containers)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	p := projects[0]
	if p.Status != "partial" {
		t.Errorf("expected status 'partial', got '%s'", p.Status)
	}
	if p.RunningServices != 1 || p.TotalServices != 2 {
		t.Errorf("expected 1/2 running services, got %d/%d", p.RunningServices, p.TotalServices)
	}
	if p.StatusDisplay != "Частично" {
		t.Errorf("expected StatusDisplay 'Частично', got '%s'", p.StatusDisplay)
	}
}

func TestGetComposeProjectLogsNotFound(t *testing.T) {
	_, err := GetComposeProjectLogs("non-existent-proj-xyz-99", nil, 50)
	if err == nil {
		t.Errorf("expected error for non-existent compose project with empty containers")
	}
}

func TestRestartComposeProjectNotFound(t *testing.T) {
	err := RestartComposeProject("non-existent-proj-xyz-99", nil)
	if err == nil {
		t.Errorf("expected error for non-existent compose project with empty containers")
	}
}

func TestGetComposeProjectLogsFallback(t *testing.T) {
	// Containers exist for project, but have invalid IDs so socket reading returns empty logs
	containers := []ContainerItem{
		{
			ID:             "mock-c1-invalid",
			Name:           "web-1",
			ComposeProject: "mock-stack",
		},
	}

	logs, err := GetComposeProjectLogs("mock-stack", containers, 0) // tailLines <= 0 defaults to 100
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(logs, "Логи сервисов проекта пусты") && !strings.Contains(logs, "web-1") {
		t.Errorf("expected fallback log message, got: %q", logs)
	}
}
