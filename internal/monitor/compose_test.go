package monitor

import (
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
