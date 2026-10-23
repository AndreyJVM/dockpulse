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

// FindDockerSocket ищет доступный UNIX-сокет Docker или Podman.
func FindDockerSocket() string {
	if dockerHost := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dockerHost, "unix://") {
		p := strings.TrimPrefix(dockerHost, "unix://")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	home, _ := os.UserHomeDir()
	uid := os.Getuid()

	candidates := []string{
		"/var/run/docker.sock",
		"/run/docker.sock",
		fmt.Sprintf("/run/user/%d/docker.sock", uid),
		filepath.Join(home, ".docker/run/docker.sock"),
		filepath.Join(home, ".docker/desktop/docker.sock"),
		fmt.Sprintf("/run/user/%d/podman/podman.sock", uid),
		"/run/podman/podman.sock",
		filepath.Join(home, ".local/share/containers/podman/machine/podman.sock"),
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
		StatusDisplay: "⚪ Остановлен",
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
				ContainersStopped int    `json:"ContainersStopped"`
				Images            int    `json:"Images"`
				ServerVersion     string `json:"ServerVersion"`
			}
			if json.Unmarshal(body, &infoData) == nil {
				summary.Status = "running"
				summary.StatusDisplay = "🟢 Запущен"
				summary.TotalContainers = infoData.Containers
				summary.RunningContainers = infoData.ContainersRunning
				summary.StoppedContainers = infoData.ContainersStopped
				summary.ImagesCount = infoData.Images
				summary.Version = infoData.ServerVersion
			}
		}

		// 2. Опрос /containers/json?all=1
		if summary.Status == "running" {
			respC, err := client.Get("http://localhost/containers/json?all=1")
			if err == nil && respC.StatusCode == 200 {
				defer respC.Body.Close()
				bodyC, _ := io.ReadAll(respC.Body)

				var rawContainers []struct {
					ID     string   `json:"Id"`
					Names  []string `json:"Names"`
					Image  string   `json:"Image"`
					State  string   `json:"State"`
					Status string   `json:"Status"`
					Ports  []struct {
						IP          string `json:"IP"`
						PrivatePort int    `json:"PrivatePort"`
						PublicPort  int    `json:"PublicPort"`
						Type        string `json:"Type"`
					} `json:"Ports"`
					Labels map[string]string `json:"Labels"`
				}

				if json.Unmarshal(bodyC, &rawContainers) == nil {
					for _, c := range rawContainers {
						cID := c.ID
						if len(cID) > 12 {
							cID = cID[:12]
						}

						name := ""
						if len(c.Names) > 0 {
							name = strings.TrimPrefix(c.Names[0], "/")
						}

						var portsList []string
						for _, p := range c.Ports {
							if p.PublicPort > 0 {
								portsList = append(portsList, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
							} else if p.PrivatePort > 0 {
								portsList = append(portsList, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
							}
						}

						composeProject := c.Labels["com.docker.compose.project"]
						composeService := c.Labels["com.docker.compose.service"]

						summary.Containers = append(summary.Containers, ContainerItem{
							ID:             cID,
							Name:           name,
							Image:          c.Image,
							State:          c.State,
							Status:         c.Status,
							Ports:          strings.Join(portsList, ", "),
							ComposeProject: composeProject,
							ComposeService: composeService,
							Labels:         c.Labels,
						})
					}
				}
			}
			return summary
		}
	}

	// Fallback CLI
	out, err := exec.Command("docker", "info", "--format", "{{.Containers}} {{.ContainersRunning}} {{.ContainersStopped}} {{.Images}} {{.ServerVersion}}").Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 5 {
			summary.Status = "running"
			summary.StatusDisplay = "🟢 Запущен"
			summary.Version = parts[4]
		}
	}

	return summary
}
