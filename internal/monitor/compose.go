package monitor

import "sort"

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
