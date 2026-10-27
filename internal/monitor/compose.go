package monitor

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
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

	var res []ComposeProject
	for _, proj := range projectsMap {
		proj.TotalServices = len(proj.Services)
		running := 0
		for _, s := range proj.Services {
			if s.State == "running" {
				running++
			}
		}
		proj.RunningServices = running

		if running == proj.TotalServices && proj.TotalServices > 0 {
			proj.Status = "running"
			proj.StatusDisplay = "Запущен"
		} else if running > 0 {
			proj.Status = "partial"
			proj.StatusDisplay = "Частично"
		} else {
			proj.Status = "stopped"
			proj.StatusDisplay = "Остановлен"
		}

		res = append(res, *proj)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res
}

// GetComposeProjectLogs читает логи всех сервисов проекта Compose.
func GetComposeProjectLogs(projectName string, containers []ContainerItem, tailLines int) (string, error) {
	if tailLines <= 0 {
		tailLines = 100
	}

	// 1. Попытка через docker compose CLI
	if _, err := exec.LookPath("docker"); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "docker", "compose", "-p", projectName, "logs", "--tail", fmt.Sprintf("%d", tailLines))
		out, err := cmd.CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return string(out), nil
		}
	}

	// 2. Fallback: сбор логов по найденным контейнерам проекта
	var projectContainers []ContainerItem
	for _, c := range containers {
		if c.ComposeProject == projectName {
			projectContainers = append(projectContainers, c)
		}
	}

	if len(projectContainers) == 0 {
		return "", fmt.Errorf("контейнеры проекта %q не найдены", projectName)
	}

	perContainerTail := tailLines / len(projectContainers)
	if perContainerTail < 20 {
		perContainerTail = 20
	}

	var sb strings.Builder
	for _, c := range projectContainers {
		logs, err := GetContainerLogs(c.ID, perContainerTail)
		if err == nil && strings.TrimSpace(logs) != "" {
			for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
				sb.WriteString(fmt.Sprintf("%s | %s\n", c.Name, line))
			}
		}
	}

	if sb.Len() == 0 {
		return "Логи сервисов проекта пусты или сервисы не выводили сообщений.", nil
	}
	return sb.String(), nil
}

// RestartComposeProject перезапускает все контейнеры проекта Compose.
func RestartComposeProject(projectName string, containers []ContainerItem) error {
	var projectContainers []ContainerItem
	for _, c := range containers {
		if c.ComposeProject == projectName {
			projectContainers = append(projectContainers, c)
		}
	}

	if len(projectContainers) == 0 {
		return fmt.Errorf("не найдено контейнеров для проекта %q", projectName)
	}

	// 1. Попытка через docker compose CLI
	if _, err := exec.LookPath("docker"); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "docker", "compose", "-p", projectName, "restart")
		if _, err := cmd.CombinedOutput(); err == nil {
			return nil
		}
	}

	// 2. Fallback: перезапуск через сокет / список контейнеров проекта
	var errs []string
	for _, c := range projectContainers {
		if err := RestartContainer(c.ID); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", c.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("ошибки перезапуска: %s", strings.Join(errs, "; "))
	}
	return nil
}
