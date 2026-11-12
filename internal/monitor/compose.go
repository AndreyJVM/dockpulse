package monitor

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	// Project statuses
	StatusRunning = "running"
	StatusPartial = "partial"
	StatusStopped = "stopped"

	// Localized status displays
	StatusDisplayRunning = "Running"
	StatusDisplayPartial = "Partial"
	StatusDisplayStopped = "Stopped"

	// Action names for Compose CLI commands
	ActionRestart = "restart"
	ActionStop    = "stop"
	ActionStart   = "start"
	ActionDown    = "down"

	// Tail and timeout defaults
	defaultTailLines       = 100
	minPerContainerTail    = 20
	composeLogsTimeout     = 5 * time.Second
	composeRestartTimeout  = 30 * time.Second
	emptyProjectLogsNotice = "Project service logs are empty or the services produced no output."
	containerStateRunning  = "running"
)

var (
	execLookPath         = exec.LookPath
	composeCommandRunner = runComposeCommand
)

// ComposeService represents a service container inside a Compose project.
type ComposeService struct {
	Name          string
	ContainerName string
	Image         string
	State         string
	Status        string
	Ports         string
}

// ComposeProject represents a Docker Compose stack/project with aggregated health.
type ComposeProject struct {
	Name            string
	Status          string // "running", "partial", "stopped"
	StatusDisplay   string
	TotalServices   int
	RunningServices int
	Services        []ComposeService
}

// GetComposeProjects groups containers into Docker Compose projects and computes aggregate statuses.
func GetComposeProjects(containers []ContainerItem) []ComposeProject {
	projectsMap := make(map[string]*ComposeProject)

	for _, c := range containers {
		if c.ComposeProject == "" {
			continue
		}

		proj, ok := projectsMap[c.ComposeProject]
		if !ok {
			proj = &ComposeProject{
				Name: c.ComposeProject,
			}
			projectsMap[c.ComposeProject] = proj
		}

		svcName := c.ComposeService
		if svcName == "" {
			svcName = c.Name
		}

		proj.Services = append(proj.Services, ComposeService{
			Name:          svcName,
			ContainerName: c.Name,
			Image:         c.Image,
			State:         c.State,
			Status:        c.Status,
			Ports:         c.Ports,
		})
	}

	res := make([]ComposeProject, 0, len(projectsMap))
	for _, proj := range projectsMap {
		proj.TotalServices = len(proj.Services)
		running := 0
		for _, s := range proj.Services {
			if s.State == containerStateRunning {
				running++
			}
		}
		proj.RunningServices = running

		// Deterministic sort of services within each project
		slices.SortFunc(proj.Services, func(a, b ComposeService) int {
			if n := strings.Compare(a.Name, b.Name); n != 0 {
				return n
			}
			return strings.Compare(a.ContainerName, b.ContainerName)
		})

		switch {
		case running == proj.TotalServices && proj.TotalServices > 0:
			proj.Status = StatusRunning
			proj.StatusDisplay = StatusDisplayRunning
		case running > 0:
			proj.Status = StatusPartial
			proj.StatusDisplay = StatusDisplayPartial
		default:
			proj.Status = StatusStopped
			proj.StatusDisplay = StatusDisplayStopped
		}

		res = append(res, *proj)
	}

	slices.SortFunc(res, func(a, b ComposeProject) int {
		return strings.Compare(a.Name, b.Name)
	})

	return res
}

// filterContainersByProject filters containers belonging to the specified Compose project.
func filterContainersByProject(containers []ContainerItem, projectName string) []ContainerItem {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		return nil
	}
	var res []ContainerItem
	for _, c := range containers {
		if c.ComposeProject == projectName {
			res = append(res, c)
		}
	}
	return res
}

// runComposeCommand executes a Docker Compose CLI command with fallback to standalone docker-compose.
func runComposeCommand(ctx context.Context, args ...string) ([]byte, error) {
	if _, err := execLookPath("docker"); err == nil {
		cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
		out, cErr := cmd.CombinedOutput()
		if cErr == nil {
			return out, nil
		}
		// If 'docker compose' returned an error, fallback to standalone 'docker-compose'
		if _, dcErr := execLookPath("docker-compose"); dcErr == nil {
			cmdDC := exec.CommandContext(ctx, "docker-compose", args...)
			return cmdDC.CombinedOutput()
		}
		return out, cErr
	}

	if _, dcErr := execLookPath("docker-compose"); dcErr == nil {
		cmdDC := exec.CommandContext(ctx, "docker-compose", args...)
		return cmdDC.CombinedOutput()
	}

	return nil, exec.ErrNotFound
}

// GetComposeProjectLogs retrieves aggregated logs for all services in a Compose project.
func GetComposeProjectLogs(projectName string, containers []ContainerItem, tailLines int) (string, error) {
	if tailLines <= 0 {
		tailLines = defaultTailLines
	}

	// 1. Attempt via Docker Compose CLI
	if strings.TrimSpace(projectName) != "" {
		ctx, cancel := context.WithTimeout(context.Background(), composeLogsTimeout)
		defer cancel()
		out, err := composeCommandRunner(ctx, "-p", projectName, "logs", "--tail", fmt.Sprintf("%d", tailLines))
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return string(out), nil
		}
	}

	// 2. Fallback: fetch logs concurrently across identified project containers
	projectContainers := filterContainersByProject(containers, projectName)
	if len(projectContainers) == 0 {
		return "", fmt.Errorf("containers for project %q were not found", projectName)
	}

	perContainerTail := tailLines / len(projectContainers)
	if perContainerTail < minPerContainerTail {
		perContainerTail = minPerContainerTail
	}

	type containerLogEntry struct {
		name string
		logs string
		err  error
	}

	entries := make([]containerLogEntry, len(projectContainers))
	var wg sync.WaitGroup
	for i, c := range projectContainers {
		wg.Add(1)
		go func(idx int, cnt ContainerItem) {
			defer wg.Done()
			logs, err := GetContainerLogs(cnt.ID, perContainerTail)
			entries[idx] = containerLogEntry{
				name: cnt.Name,
				logs: logs,
				err:  err,
			}
		}(i, c)
	}
	wg.Wait()

	var sb strings.Builder
	for _, entry := range entries {
		if entry.err == nil && strings.TrimSpace(entry.logs) != "" {
			for _, line := range strings.Split(strings.TrimSpace(entry.logs), "\n") {
				sb.WriteString(entry.name)
				sb.WriteString(" | ")
				sb.WriteString(line)
				sb.WriteByte('\n')
			}
		}
	}

	if sb.Len() == 0 {
		return emptyProjectLogsNotice, nil
	}
	return sb.String(), nil
}

// executeComposeProjectAction executes a project-level lifecycle action via CLI with fallback to per-container handler.
func executeComposeProjectAction(projectName, actionName string, containers []ContainerItem, fallbackFn func(id string) error) error {
	projectContainers := filterContainersByProject(containers, projectName)
	if len(projectContainers) == 0 {
		return fmt.Errorf("no containers found for project %q", projectName)
	}

	// 1. Attempt via Docker Compose CLI
	if strings.TrimSpace(projectName) != "" {
		ctx, cancel := context.WithTimeout(context.Background(), composeRestartTimeout)
		defer cancel()
		if _, err := composeCommandRunner(ctx, "-p", projectName, actionName); err == nil {
			return nil
		}
	}

	// 2. Fallback: execute action concurrently for each container in the project
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []string
	)
	for _, c := range projectContainers {
		wg.Add(1)
		go func(cnt ContainerItem) {
			defer wg.Done()
			if err := fallbackFn(cnt.ID); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s: %v", cnt.Name, err))
				mu.Unlock()
			}
		}(c)
	}
	wg.Wait()

	if len(errs) > 0 {
		slices.Sort(errs)
		var actionTitle string
		switch actionName {
		case ActionRestart:
			actionTitle = "restart"
		case ActionStop:
			actionTitle = "stop"
		case ActionStart:
			actionTitle = "start"
		case ActionDown:
			actionTitle = "down"
		default:
			actionTitle = actionName
		}
		return fmt.Errorf("%s errors: %s", actionTitle, strings.Join(errs, "; "))
	}
	return nil
}

// RestartComposeProject restarts all containers in a Compose project.
func RestartComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, ActionRestart, containers, RestartContainer)
}

// StopComposeProject stops all containers in a Compose project.
func StopComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, ActionStop, containers, StopContainer)
}

// StartComposeProject starts all containers in a Compose project.
func StartComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, ActionStart, containers, StartContainer)
}

// DownComposeProject tears down a Compose project (docker compose down),
// with a concurrent fallback to stop and remove each container.
func DownComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, ActionDown, containers, func(id string) error {
		_ = StopContainer(id)
		return RemoveContainer(id)
	})
}
