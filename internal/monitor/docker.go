package monitor

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ContainerItem содержит описание отдельного контейнера.
type ContainerItem struct {
	ID             string
	Name           string
	Image          string
	State          string
	Status         string
	Health         string
	Ports          string
	ComposeProject string
	ComposeService string
	Labels         map[string]string
	CPUPerc        float64
	MemUsageMB     float64
	MemLimitMB     float64
	MemPerc        float64
}

// ContainerDetail содержит детальную информацию об одном контейнере.
type ContainerDetail struct {
	ID         string
	Name       string
	Image      string
	Created    string
	State      string
	Status     string
	Health     string
	IP         string
	Gateway    string
	MACAddress string
	Command    string
	Mounts     []string
	Env        []string
	CPUPerc    float64
	MemUsageMB float64
	MemPerc    float64
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

// ParseHealthStatus извлекает стандартизированный статус healthcheck ("healthy", "unhealthy", "starting" или "")
// из поля State.Health.Status или строки Status контейнера Docker.
func ParseHealthStatus(val string) string {
	s := strings.ToLower(strings.TrimSpace(val))
	if s == "" {
		return ""
	}
	switch s {
	case "healthy":
		return "healthy"
	case "unhealthy":
		return "unhealthy"
	case "starting":
		return "starting"
	}

	if strings.Contains(s, "(healthy)") {
		return "healthy"
	}
	if strings.Contains(s, "(unhealthy)") {
		return "unhealthy"
	}
	if strings.Contains(s, "(health: starting)") || (strings.Contains(s, "starting") && strings.Contains(s, "health")) {
		return "starting"
	}

	return ""
}

// ContainerStatsRaw описывает структуру ответа /containers/{id}/stats?stream=false от Docker Engine API.
type ContainerStatsRaw struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
		Stats struct {
			InactiveFile      uint64 `json:"inactive_file"`
			TotalInactiveFile uint64 `json:"total_inactive_file"`
		} `json:"stats"`
	} `json:"memory_stats"`
}

func calculateContainerStats(s ContainerStatsRaw) (cpuPerc float64, memUsageMB float64, memLimitMB float64, memPerc float64) {
	// Расчет загрузки CPU в процентах
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(s.CPUStats.SystemCPUUsage) - float64(s.PreCPUStats.SystemCPUUsage)
	onlineCPUs := s.CPUStats.OnlineCPUs
	if onlineCPUs <= 0 {
		onlineCPUs = 1
	}

	if systemDelta > 0 && cpuDelta > 0 {
		cpuPerc = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
	}

	// Расчет потребления оперативной памяти
	memUsage := s.MemoryStats.Usage
	if s.MemoryStats.Stats.TotalInactiveFile > 0 && memUsage > s.MemoryStats.Stats.TotalInactiveFile {
		memUsage -= s.MemoryStats.Stats.TotalInactiveFile
	} else if s.MemoryStats.Stats.InactiveFile > 0 && memUsage > s.MemoryStats.Stats.InactiveFile {
		memUsage -= s.MemoryStats.Stats.InactiveFile
	}

	memUsageMB = float64(memUsage) / (1024 * 1024)
	memLimitMB = float64(s.MemoryStats.Limit) / (1024 * 1024)
	if s.MemoryStats.Limit > 0 {
		memPerc = (float64(memUsage) / float64(s.MemoryStats.Limit)) * 100.0
	}

	return cpuPerc, memUsageMB, memLimitMB, memPerc
}

func fetchContainerStats(client *http.Client, containerID string) (cpuPerc float64, memUsageMB float64, memLimitMB float64, memPerc float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost/containers/%s/stats?stream=false&one-shot=true", containerID), nil)
	if err != nil {
		return 0, 0, 0, 0
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return 0, 0, 0, 0
	}
	defer resp.Body.Close()

	var s ContainerStatsRaw
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return 0, 0, 0, 0
	}

	return calculateContainerStats(s)
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
		Timeout: 5 * time.Second,
	}
}

// GetDockerSummary собирает статус службы Docker и список контейнеров с живыми метриками.
func GetDockerSummary() DockerSummary {
	_, cliErr := exec.LookPath("docker")
	sockPath := FindDockerSocket()

	if cliErr != nil && sockPath == "" {
		return DockerSummary{
			Status:        "not_installed",
			StatusDisplay: "Не установлен",
		}
	}

	summary := DockerSummary{
		Status:        "stopped",
		StatusDisplay: "Остановлен",
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
				summary.StatusDisplay = "Запущен"
				summary.TotalContainers = infoData.Containers
				summary.RunningContainers = infoData.ContainersRunning
				summary.StoppedContainers = infoData.ContainersStopped
				summary.ImagesCount = infoData.Images
				summary.Version = infoData.ServerVersion
			}
		}

		// 2. Опрос /containers/json?all=1
		respC, errC := client.Get("http://localhost/containers/json?all=1")
		if errC == nil && respC.StatusCode == 200 {
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
						Health:         ParseHealthStatus(c.Status),
						Ports:          strings.Join(portsList, ", "),
						ComposeProject: composeProject,
						ComposeService: composeService,
						Labels:         c.Labels,
					})
				}

				// Параллельный опрос метрик CPU и памяти для работающих контейнеров
				if len(summary.Containers) > 0 {
					var wg sync.WaitGroup
					for i := range summary.Containers {
						if summary.Containers[i].State == "running" {
							wg.Add(1)
							go func(idx int, cID string) {
								defer wg.Done()
								cpu, memUsage, memLimit, memPerc := fetchContainerStats(client, cID)
								summary.Containers[idx].CPUPerc = cpu
								summary.Containers[idx].MemUsageMB = memUsage
								summary.Containers[idx].MemLimitMB = memLimit
								summary.Containers[idx].MemPerc = memPerc
							}(i, summary.Containers[i].ID)
						}
					}
					wg.Wait()
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
			summary.StatusDisplay = "Запущен"
			summary.Version = parts[4]
		}
	}

	return summary
}

// StartContainer запускает контейнер по ID.
func StartContainer(containerID string) error {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/start", containerID)
		resp, err := client.Post(url, "application/json", nil)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 204 || resp.StatusCode == 304 {
				return nil
			}
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	out, err := exec.Command("docker", "start", containerID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StopContainer останавливает контейнер по ID.
func StopContainer(containerID string) error {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/stop", containerID)
		resp, err := client.Post(url, "application/json", nil)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 204 || resp.StatusCode == 304 {
				return nil
			}
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	out, err := exec.Command("docker", "stop", containerID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// RestartContainer перезапускает контейнер по ID.
func RestartContainer(containerID string) error {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/restart", containerID)
		resp, err := client.Post(url, "application/json", nil)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 204 {
				return nil
			}
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	out, err := exec.Command("docker", "restart", containerID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// GetContainerLogs читает последние N строк логов контейнера.
func GetContainerLogs(containerID string, tailLines int) (string, error) {
	if tailLines <= 0 {
		tailLines = 100
	}

	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/logs?stdout=1&stderr=1&tail=%d", containerID, tailLines)
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				return demuxDockerLogs(resp.Body), nil
			}
		}
	}

	out, err := exec.Command("docker", "logs", "--tail", fmt.Sprintf("%d", tailLines), containerID).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// demuxDockerLogs разбирает мультиплексированный стрим логов Docker Engine API (8-байтный заголовок каждого фрейма).
func demuxDockerLogs(r io.Reader) string {
	all, err := io.ReadAll(r)
	if err != nil || len(all) == 0 {
		return ""
	}

	var sb strings.Builder
	idx := 0
	for idx+8 <= len(all) {
		streamType := all[idx]
		if streamType > 2 { // Not a multiplexed frame (0=stdin, 1=stdout, 2=stderr)
			return string(all)
		}
		size := int(binary.BigEndian.Uint32(all[idx+4 : idx+8]))
		if idx+8+size > len(all) {
			return string(all)
		}
		sb.Write(all[idx+8 : idx+8+size])
		idx += 8 + size
	}

	if sb.Len() > 0 && idx == len(all) {
		return sb.String()
	}
	return string(all)
}

// GetContainerDetail возвращает подробную инспекцию контейнера.
func GetContainerDetail(containerID string) (ContainerDetail, error) {
	var detail ContainerDetail
	detail.ID = containerID

	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/json", containerID)
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var data struct {
				Name    string `json:"Name"`
				Created string `json:"Created"`
				State   struct {
					Status   string `json:"Status"`
					ExitCode int    `json:"ExitCode"`
					Health   *struct {
						Status        string `json:"Status"`
						FailingStreak int    `json:"FailingStreak"`
					} `json:"Health"`
				} `json:"State"`
				Config struct {
					Image string   `json:"Image"`
					Cmd   []string `json:"Cmd"`
					Env   []string `json:"Env"`
				} `json:"Config"`
				NetworkSettings struct {
					IPAddress  string `json:"IPAddress"`
					Gateway    string `json:"Gateway"`
					MacAddress string `json:"MacAddress"`
				} `json:"NetworkSettings"`
				Mounts []struct {
					Source      string `json:"Source"`
					Destination string `json:"Destination"`
					Mode        string `json:"Mode"`
				} `json:"Mounts"`
			}

			if json.Unmarshal(body, &data) == nil {
				detail.Name = strings.TrimPrefix(data.Name, "/")
				detail.Image = data.Config.Image
				detail.Created = data.Created
				detail.State = data.State.Status
				detail.Status = fmt.Sprintf("%s (ExitCode: %d)", data.State.Status, data.State.ExitCode)
				if data.State.Health != nil && data.State.Health.Status != "" {
					detail.Health = ParseHealthStatus(data.State.Health.Status)
				} else {
					detail.Health = ParseHealthStatus(detail.Status)
				}
				detail.IP = data.NetworkSettings.IPAddress
				detail.Gateway = data.NetworkSettings.Gateway
				detail.MACAddress = data.NetworkSettings.MacAddress
				detail.Command = strings.Join(data.Config.Cmd, " ")

				for _, m := range data.Mounts {
					detail.Mounts = append(detail.Mounts, fmt.Sprintf("%s -> %s (%s)", m.Source, m.Destination, m.Mode))
				}
				detail.Env = data.Config.Env

				// Опрос метрик если контейнер запущен
				if detail.State == "running" {
					cpu, memUsage, _, memPerc := fetchContainerStats(client, containerID)
					detail.CPUPerc = cpu
					detail.MemUsageMB = memUsage
					detail.MemPerc = memPerc
				}

				return detail, nil
			}
		}
	}

	// CLI inspect fallback
	out, err := exec.Command("docker", "inspect", containerID).Output()
	if err == nil {
		var raw []map[string]interface{}
		if json.Unmarshal(out, &raw) == nil && len(raw) > 0 {
			first := raw[0]
			if name, ok := first["Name"].(string); ok {
				detail.Name = strings.TrimPrefix(name, "/")
			}
			if state, ok := first["State"].(map[string]interface{}); ok {
				if status, ok := state["Status"].(string); ok {
					detail.State = status
					detail.Status = status
				}
				if health, ok := state["Health"].(map[string]interface{}); ok {
					if hStatus, ok := health["Status"].(string); ok {
						detail.Health = ParseHealthStatus(hStatus)
					}
				}
			}
			if detail.Health == "" {
				detail.Health = ParseHealthStatus(detail.Status)
			}
			return detail, nil
		}
	}

	return detail, fmt.Errorf("container not found: %s", containerID)
}

// IsDockerCLIAvailable проверяет доступность утилиты docker в PATH.
func IsDockerCLIAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// GetContainerExecCommand возвращает команду exec.Cmd для запуска интерактивного шелла в контейнере.
func GetContainerExecCommand(containerID string) *exec.Cmd {
	return exec.Command("docker", "exec", "-it", containerID, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi")
}
