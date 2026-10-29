package monitor

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"
)

const (
	// Project statuses
	StatusRunning = "running"
	StatusPartial = "partial"
	StatusStopped = "stopped"

	// Localized status displays
	StatusDisplayRunning = "Запущен"
	StatusDisplayPartial = "Частично"
	StatusDisplayStopped = "Остановлен"

	// Tail and timeout defaults
	defaultTailLines      = 100
	minPerContainerTail   = 20
	composeLogsTimeout    = 5 * time.Second
	composeRestartTimeout = 30 * time.Second
)

// ComposeService представляет сервис внутри проекта Compose.
type ComposeService struct {
	Name          string
	ContainerName string
	Image         string
	State         string
	Status        string
	Ports         string
}

// ComposeProject представляет стек/проект Docker Compose.
type ComposeProject struct {
	Name            string
	Status          string // "running", "partial", "stopped"
	StatusDisplay   string
	TotalServices   int
	RunningServices int
	Services        []ComposeService
}

// GetComposeProjects группирует контейнеры по проектам Compose.
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
			if s.State == "running" {
				running++
			}
		}
		proj.RunningServices = running

		// Стабильная сортировка сервисов внутри каждого проекта
		slices.SortFunc(proj.Services, func(a, b ComposeService) int {
			return strings.Compare(a.Name, b.Name)
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

// filterContainersByProject возвращает контейнеры, принадлежащие указанному проекту Compose.
func filterContainersByProject(containers []ContainerItem, projectName string) []ContainerItem {
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

var composeCommandRunner = func(ctx context.Context, args ...string) ([]byte, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	return cmd.CombinedOutput()
}

// GetComposeProjectLogs читает логи всех сервисов проекта Compose.
func GetComposeProjectLogs(projectName string, containers []ContainerItem, tailLines int) (string, error) {
	if tailLines <= 0 {
		tailLines = defaultTailLines
	}

	// 1. Попытка через docker compose CLI
	if projectName != "" {
		ctx, cancel := context.WithTimeout(context.Background(), composeLogsTimeout)
		defer cancel()
		out, err := composeCommandRunner(ctx, "-p", projectName, "logs", "--tail", fmt.Sprintf("%d", tailLines))
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return string(out), nil
		}
	}

	// 2. Fallback: сбор логов по найденным контейнерам проекта
	projectContainers := filterContainersByProject(containers, projectName)
	if len(projectContainers) == 0 {
		return "", fmt.Errorf("контейнеры проекта %q не найдены", projectName)
	}

	perContainerTail := tailLines / len(projectContainers)
	if perContainerTail < minPerContainerTail {
		perContainerTail = minPerContainerTail
	}

	var sb strings.Builder
	for _, c := range projectContainers {
		logs, err := GetContainerLogs(c.ID, perContainerTail)
		if err == nil && strings.TrimSpace(logs) != "" {
			for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
				sb.WriteString(c.Name)
				sb.WriteString(" | ")
				sb.WriteString(line)
				sb.WriteByte('\n')
			}
		}
	}

	if sb.Len() == 0 {
		return "Логи сервисов проекта пусты или сервисы не выводили сообщений.", nil
	}
	return sb.String(), nil
}

// executeComposeProjectAction выполняет действие над проектом Compose через CLI с fallback на функцию контейнера.
func executeComposeProjectAction(projectName, actionName string, containers []ContainerItem, fallbackFn func(id string) error) error {
	projectContainers := filterContainersByProject(containers, projectName)
	if len(projectContainers) == 0 {
		return fmt.Errorf("не найдено контейнеров для проекта %q", projectName)
	}

	// 1. Попытка через docker compose CLI
	if projectName != "" {
		ctx, cancel := context.WithTimeout(context.Background(), composeRestartTimeout)
		defer cancel()
		if _, err := composeCommandRunner(ctx, "-p", projectName, actionName); err == nil {
			return nil
		}
	}

	// 2. Fallback: выполнение действия для каждого контейнера проекта
	var errs []string
	for _, c := range projectContainers {
		if err := fallbackFn(c.ID); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", c.Name, err))
		}
	}

	if len(errs) > 0 {
		var actionTitle string
		switch actionName {
		case "restart":
			actionTitle = "перезапуска"
		case "stop":
			actionTitle = "остановки"
		case "start":
			actionTitle = "запуска"
		default:
			actionTitle = actionName
		}
		return fmt.Errorf("ошибки %s: %s", actionTitle, strings.Join(errs, "; "))
	}
	return nil
}

// RestartComposeProject перезапускает все контейнеры проекта Compose.
func RestartComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, "restart", containers, RestartContainer)
}

// StopComposeProject останавливает все контейнеры проекта Compose.
func StopComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, "stop", containers, StopContainer)
}

// StartComposeProject запускает все контейнеры проекта Compose.
func StartComposeProject(projectName string, containers []ContainerItem) error {
	return executeComposeProjectAction(projectName, "start", containers, StartContainer)
}

// DownComposeProject сворачивает проект Compose (docker compose down), с fallback на остановку и удаление контейнеров.
func DownComposeProject(projectName string, containers []ContainerItem) error {
	projectContainers := filterContainersByProject(containers, projectName)
	if len(projectContainers) == 0 {
		return fmt.Errorf("не найдено контейнеров для проекта %q", projectName)
	}

	// 1. Попытка через docker compose CLI
	if projectName != "" {
		ctx, cancel := context.WithTimeout(context.Background(), composeRestartTimeout)
		defer cancel()
		if _, err := composeCommandRunner(ctx, "-p", projectName, "down"); err == nil {
			return nil
		}
	}

	// 2. Fallback: остановка и удаление каждого контейнера проекта
	var errs []string
	for _, c := range projectContainers {
		_ = StopContainer(c.ID)
		if err := RemoveContainer(c.ID); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", c.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("ошибки сворачивания (down): %s", strings.Join(errs, "; "))
	}
	return nil
}
