package monitor

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
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
			// All running -> "running"
			{ID: "1", Name: "app-web", ComposeProject: "alpha-stack", ComposeService: "web", State: "running"},
			{ID: "2", Name: "app-db", ComposeProject: "alpha-stack", ComposeService: "db", State: "running"},

			// Partial -> "partial"
			{ID: "3", Name: "beta-api", ComposeProject: "beta-stack", ComposeService: "api", State: "running"},
			{ID: "4", Name: "beta-worker", ComposeProject: "beta-stack", ComposeService: "worker", State: "exited"},

			// All stopped -> "stopped"
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
			"alpha-stack": {status: "running", statusDisplay: "Running", total: 2, running: 2},
			"beta-stack":  {status: "partial", statusDisplay: "Partial", total: 2, running: 1},
			"gamma-stack": {status: "stopped", statusDisplay: "Stopped", total: 2, running: 0},
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
	t.Run("cli logs success", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return []byte("web-1 | Listening on 8080\napi-1 | Server ready\n"), nil
		}
		defer func() { composeCommandRunner = oldRunner }()

		logs, err := GetComposeProjectLogs("my-stack", nil, 50)
		if err != nil {
			t.Fatalf("expected nil error on CLI success, got: %v", err)
		}
		if !strings.Contains(logs, "Listening on 8080") {
			t.Errorf("expected cli logs output, got: %q", logs)
		}
	})

	t.Run("non-existent project returns error", func(t *testing.T) {
		_, err := GetComposeProjectLogs("non-existent-proj", nil, 50)
		if err == nil {
			t.Fatal("expected error for non-existent project, got nil")
		}
		if !strings.Contains(err.Error(), "were not found") {
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
		if !strings.Contains(logs, "Project service logs are empty") && !strings.Contains(logs, "web-1") {
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
		if !strings.Contains(err.Error(), "no containers found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("cli restart succeeds", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return []byte("Restarted"), nil
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{ID: "c1", Name: "web", ComposeProject: "my-stack"},
		}
		err := RestartComposeProject("my-stack", containers)
		if err != nil {
			t.Fatalf("expected nil error on CLI success, got: %v", err)
		}
	})

	t.Run("fallback restarts invalid containers with error aggregation", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return nil, errors.New("compose cli unavailable")
		}
		defer func() { composeCommandRunner = oldRunner }()

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
		if !strings.Contains(err.Error(), "restart errors") {
			t.Errorf("expected restart errors in error message, got: %v", err)
		}
	})
}

func TestStopComposeProject(t *testing.T) {
	t.Run("non-existent project returns error", func(t *testing.T) {
		err := StopComposeProject("missing-proj", nil)
		if err == nil {
			t.Fatal("expected error for non-existent project, got nil")
		}
	})

	t.Run("cli stop succeeds", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return []byte("Stopped"), nil
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{ID: "c1", Name: "web", ComposeProject: "my-stack"},
		}
		err := StopComposeProject("my-stack", containers)
		if err != nil {
			t.Fatalf("expected nil error on CLI success, got: %v", err)
		}
	})

	t.Run("fallback stops invalid containers with error aggregation", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return nil, errors.New("compose cli unavailable")
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{
				ID:             "non-existent-id-001",
				Name:           "broken-web",
				ComposeProject: "failing-stack",
			},
		}

		err := StopComposeProject("failing-stack", containers)
		if err == nil {
			t.Fatal("expected error when stopping non-existent container ID, got nil")
		}
		if !strings.Contains(err.Error(), "stop errors") {
			t.Errorf("expected stop errors in error message, got: %v", err)
		}
	})
}

func TestStartComposeProject(t *testing.T) {
	t.Run("non-existent project returns error", func(t *testing.T) {
		err := StartComposeProject("missing-proj", nil)
		if err == nil {
			t.Fatal("expected error for non-existent project, got nil")
		}
	})

	t.Run("cli start succeeds", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return []byte("Started"), nil
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{ID: "c1", Name: "web", ComposeProject: "my-stack"},
		}
		err := StartComposeProject("my-stack", containers)
		if err != nil {
			t.Fatalf("expected nil error on CLI success, got: %v", err)
		}
	})

	t.Run("fallback starts invalid containers with error aggregation", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return nil, errors.New("compose cli unavailable")
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{
				ID:             "non-existent-id-001",
				Name:           "broken-web",
				ComposeProject: "failing-stack",
			},
		}

		err := StartComposeProject("failing-stack", containers)
		if err == nil {
			t.Fatal("expected error when starting non-existent container ID, got nil")
		}
		if !strings.Contains(err.Error(), "start errors") {
			t.Errorf("expected start errors in error message, got: %v", err)
		}
	})
}

func TestDownComposeProject(t *testing.T) {
	t.Run("non-existent project returns error", func(t *testing.T) {
		err := DownComposeProject("missing-proj", nil)
		if err == nil {
			t.Fatal("expected error for non-existent project, got nil")
		}
	})

	t.Run("cli down succeeds", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return []byte("Removed"), nil
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{ID: "c1", Name: "web", ComposeProject: "my-stack"},
		}
		err := DownComposeProject("my-stack", containers)
		if err != nil {
			t.Fatalf("expected nil error on CLI success, got: %v", err)
		}
	})

	t.Run("fallback down invalid containers with error aggregation", func(t *testing.T) {
		oldRunner := composeCommandRunner
		composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
			return nil, errors.New("compose cli unavailable")
		}
		defer func() { composeCommandRunner = oldRunner }()

		containers := []ContainerItem{
			{
				ID:             "non-existent-id-001",
				Name:           "broken-web",
				ComposeProject: "failing-stack",
			},
		}

		err := DownComposeProject("failing-stack", containers)
		if err == nil {
			t.Fatal("expected error when downing non-existent container ID, got nil")
		}
		if !strings.Contains(err.Error(), "down errors") {
			t.Errorf("expected down errors in error message, got: %v", err)
		}
	})
}

func TestExecuteComposeProjectAction_FallbackSuccess(t *testing.T) {
	oldRunner := composeCommandRunner
	composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return nil, errors.New("compose cli unavailable")
	}
	defer func() { composeCommandRunner = oldRunner }()

	containers := []ContainerItem{
		{ID: "c1", Name: "web-1", ComposeProject: "success-stack"},
		{ID: "c2", Name: "web-2", ComposeProject: "success-stack"},
	}

	var mu sync.Mutex
	callCount := 0
	mockFallback := func(_ string) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	err := executeComposeProjectAction("success-stack", "restart", containers, mockFallback)
	if err != nil {
		t.Fatalf("expected nil error when fallback succeeds, got: %v", err)
	}
	mu.Lock()
	count := callCount
	mu.Unlock()
	if count != 2 {
		t.Errorf("expected 2 calls, got %d", count)
	}
}

func TestExecuteComposeProjectAction_CustomActionTitle(t *testing.T) {
	oldRunner := composeCommandRunner
	composeCommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return nil, errors.New("compose cli unavailable")
	}
	defer func() { composeCommandRunner = oldRunner }()

	containers := []ContainerItem{
		{ID: "c1", Name: "web-1", ComposeProject: "custom-stack"},
	}

	failingFallback := func(_ string) error {
		return errors.New("custom failure")
	}

	err := executeComposeProjectAction("custom-stack", "pause", containers, failingFallback)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "pause errors") {
		t.Errorf("expected pause errors in error message, got: %v", err)
	}
}

func TestDownComposeProject_EmptyProjectName(t *testing.T) {
	err := DownComposeProject("   ", nil)
	if err == nil {
		t.Fatal("expected error for empty project name, got nil")
	}
}

func TestFilterContainersByProject(t *testing.T) {
	containers := []ContainerItem{
		{ID: "c1", Name: "w1", ComposeProject: "my-project"},
		{ID: "c2", Name: "w2", ComposeProject: "other-project"},
	}

	if res := filterContainersByProject(containers, ""); len(res) != 0 {
		t.Errorf("expected 0 containers for empty project name, got %d", len(res))
	}
	if res := filterContainersByProject(containers, "   "); len(res) != 0 {
		t.Errorf("expected 0 containers for whitespace project name, got %d", len(res))
	}
	if res := filterContainersByProject(containers, "my-project"); len(res) != 1 || res[0].ID != "c1" {
		t.Errorf("expected 1 matching container c1, got %v", res)
	}
}

func TestGetComposeProjects_DeterministicTieBreaker(t *testing.T) {
	containers := []ContainerItem{
		{ID: "c2", Name: "worker-2", ComposeProject: "scale-stack", ComposeService: "worker"},
		{ID: "c1", Name: "worker-1", ComposeProject: "scale-stack", ComposeService: "worker"},
		{ID: "c3", Name: "worker-3", ComposeProject: "scale-stack", ComposeService: "worker"},
	}

	projects := GetComposeProjects(containers)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	svcs := projects[0].Services
	if len(svcs) != 3 {
		t.Fatalf("expected 3 services, got %d", len(svcs))
	}
	if svcs[0].ContainerName != "worker-1" || svcs[1].ContainerName != "worker-2" || svcs[2].ContainerName != "worker-3" {
		t.Errorf("expected deterministic tie-breaker sorting [worker-1, worker-2, worker-3], got [%s, %s, %s]",
			svcs[0].ContainerName, svcs[1].ContainerName, svcs[2].ContainerName)
	}
}

func TestComposeCommandRunner_Execution(_ *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Should not crash when executed
	_, _ = composeCommandRunner(ctx, "version")
}
