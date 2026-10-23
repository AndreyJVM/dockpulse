package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ContainerItem содержит описание отдельного контейнера.
type ContainerItem struct {
	ID             string
	Name           string
	Image          string
	State          string
	Status         string
	Ports          string
	ComposeProject string
	ComposeService string
	Labels         map[string]string
}

// DockerSummary содержит сводку службы Docker.
type DockerSummary struct {
	Status            string // "running", "stopped", "not_installed"
	StatusDisplay     string
	RunningContainers int
	StoppedContainers int
	TotalContainers   int
	ImagesCount       int
	Version           string
	Containers        []ContainerItem
}

// FindDockerSocket ищет UNIX-сокет Docker.
func FindDockerSocket() string {
	if dockerHost := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dockerHost, "unix://") {
		p := strings.TrimPrefix(dockerHost, "unix://")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	home, _ := os.UserHomeDir()
	candidates := []string{
		"/var/run/docker.sock",
		"/run/docker.sock",
		filepath.Join(home, ".docker/run/docker.sock"),
		filepath.Join(home, ".docker/desktop/docker.sock"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func getSocketHTTPClient(sockPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.DialTimeout("unix", sockPath, 500*time.Millisecond)
			},
		},
		Timeout: 1 * time.Second,
	}
}

// GetDockerSummary собирает статус службы Docker и список контейнеров.
func GetDockerSummary() DockerSummary {
	_, cliErr := exec.LookPath("docker")
	sockPath := FindDockerSocket()

	if cliErr != nil && sockPath == "" {
		return DockerSummary{
			Status:        "not_installed",
			StatusDisplay: "🔴 Не установлен",
		}
	}

	summary := DockerSummary{
		Status:        "stopped",
		StatusDisplay: "⚪ Остановлен (Inactive)",
	}

	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)

		// 1. Опрос /info
		resp, err := client.Get("http://localhost/info")
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var infoData struct {
				Containers        int    `json:"Containers"`
				ContainersRunning int    `json:"ContainersRunning"`
				ContainersPaused  int    `json:"ContainersPaused"`
				ContainersStopped int    `json:"ContainersStopped"`
				Images            int    `json:"Images"`
				ServerVersion     string `json:"ServerVersion"`
			}

			if json.Unmarshal(body, &infoData) == nil {
				summary.Status = "running"
				summary.StatusDisplay = "🟢 Запущен (Active)"
				summary.RunningContainers = infoData.ContainersRunning
				summary.StoppedContainers = infoData.ContainersStopped + infoData.ContainersPaused
				summary.TotalContainers = infoData.Containers
				summary.ImagesCount = infoData.Images
				summary.Version = infoData.ServerVersion
			}
		}

		// 2. Опрос /containers/json?all=1 если демон запущен
		if summary.Status == "running" {
			summary.Containers = fetchContainersViaSocket(client)
			return summary
		}
	}

	// Fallback через CLI
	if cliErr == nil {
		out, err := exec.Command("docker", "info", "--format", "{{.ContainersRunning}} {{.ContainersStopped}} {{.Containers}} {{.Images}} {{.ServerVersion}}").Output()
		if err == nil {
			parts := strings.Fields(string(out))
			if len(parts) >= 3 {
				summary.Status = "running"
				summary.StatusDisplay = "🟢 Запущен (Active)"
				fmt.Sscanf(parts[0], "%d", &summary.RunningContainers)
				fmt.Sscanf(parts[1], "%d", &summary.StoppedContainers)
				fmt.Sscanf(parts[2], "%d", &summary.TotalContainers)
				if len(parts) > 3 {
					fmt.Sscanf(parts[3], "%d", &summary.ImagesCount)
				}
				if len(parts) > 4 {
					summary.Version = parts[4]
				}
				summary.Containers = fetchContainersViaCLI()
				return summary
			}
		}
	}

	return summary
}

func fetchContainersViaSocket(client *http.Client) []ContainerItem {
	resp, err := client.Get("http://localhost/containers/json?all=1")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rawList []struct {
		ID     string            `json:"Id"`
		Names  []string          `json:"Names"`
		Image  string            `json:"Image"`
		State  string            `json:"State"`
		Status string            `json:"Status"`
		Labels map[string]string `json:"Labels"`
		Ports  []struct {
			PublicPort  int    `json:"PublicPort"`
			PrivatePort int    `json:"PrivatePort"`
			Type        string `json:"Type"`
		} `json:"Ports"`
	}

	if json.Unmarshal(body, &rawList) != nil {
		return nil
	}

	var res []ContainerItem
	for _, c := range rawList {
		name := "unknown"
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		cid := c.ID
		if len(cid) > 12 {
			cid = cid[:12]
		}

		var portStrs []string
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				portStrs = append(portStrs, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else if p.PrivatePort > 0 {
				portStrs = append(portStrs, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}
		portsVal := "—"
		if len(portStrs) > 0 {
			portsVal = strings.Join(portStrs, ", ")
		}

		proj := ""
		svc := ""
		if c.Labels != nil {
			proj = c.Labels["com.docker.compose.project"]
			svc = c.Labels["com.docker.compose.service"]
		}

		res = append(res, ContainerItem{
			ID:             cid,
			Name:           name,
			Image:          c.Image,
			State:          c.State,
			Status:         c.Status,
			Ports:          portsVal,
			ComposeProject: proj,
			ComposeService: svc,
			Labels:         c.Labels,
		})
	}
	return res
}

func fetchContainersViaCLI() []ContainerItem {
	out, err := exec.Command("docker", "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}").Output()
	if err != nil {
		return nil
	}

	var res []ContainerItem
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) >= 5 {
			ports := "—"
			if len(parts) > 5 && parts[5] != "" {
				ports = parts[5]
			}
			res = append(res, ContainerItem{
				ID:     parts[0],
				Name:   parts[1],
				Image:  parts[2],
				State:  strings.ToLower(parts[3]),
				Status: parts[4],
				Ports:  ports,
			})
		}
	}
	return res
}
