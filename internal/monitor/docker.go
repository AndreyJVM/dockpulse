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

const (
	// Docker statuses
	DockerStatusRunning      = "running"
	DockerStatusStopped      = "stopped"
	DockerStatusNotInstalled = "not_installed"

	// Localized status displays
	DockerStatusDisplayRunning      = "Запущен"
	DockerStatusDisplayStopped      = "Остановлен"
	DockerStatusDisplayNotInstalled = "Не установлен"

	// Defaults and timeouts
	maxConcurrentStatsFetch = 10
	socketDialTimeout       = 500 * time.Millisecond
	socketClientTimeout     = 5 * time.Second
	containerStatsTimeout   = 2 * time.Second
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
	ctx, cancel := context.WithTimeout(context.Background(), containerStatsTimeout)
	defer cancel()

	url := fmt.Sprintf("http://localhost/containers/%s/stats?stream=false&one-shot=true", containerID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0, 0, 0
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
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

	candidates := []string{
		"/var/run/docker.sock",
		"/run/docker.sock",
	}

	uid := os.Getuid()
	candidates = append(candidates,
		fmt.Sprintf("/run/user/%d/docker.sock", uid),
		fmt.Sprintf("/run/user/%d/podman/podman.sock", uid),
		"/run/podman/podman.sock",
	)

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".docker/run/docker.sock"),
			filepath.Join(home, ".docker/desktop/docker.sock"),
			filepath.Join(home, ".local/share/containers/podman/machine/podman.sock"),
		)
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
				var d net.Dialer
				d.Timeout = socketDialTimeout
				return d.DialContext(ctx, "unix", sockPath)
			},
		},
		Timeout: socketClientTimeout,
	}
}

type rawPort struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

func formatPorts(ports []rawPort) string {
	if len(ports) == 0 {
		return ""
	}
	portsList := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.PublicPort > 0 {
			portsList = append(portsList, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
		} else if p.PrivatePort > 0 {
			portsList = append(portsList, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
		}
	}
	return strings.Join(portsList, ", ")
}

// GetDockerSummary собирает статус службы Docker и список контейнеров с живыми метриками.
func GetDockerSummary() DockerSummary {
	_, cliErr := exec.LookPath("docker")
	sockPath := FindDockerSocket()

	if cliErr != nil && sockPath == "" {
		return DockerSummary{
			Status:        DockerStatusNotInstalled,
			StatusDisplay: DockerStatusDisplayNotInstalled,
		}
	}

	summary := DockerSummary{
		Status:        DockerStatusStopped,
		StatusDisplay: DockerStatusDisplayStopped,
	}

	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)

		// 1. Опрос /info
		resp, err := client.Get("http://localhost/info")
		if err == nil && resp.StatusCode == http.StatusOK {
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
				summary.Status = DockerStatusRunning
				summary.StatusDisplay = DockerStatusDisplayRunning
				summary.TotalContainers = infoData.Containers
				summary.RunningContainers = infoData.ContainersRunning
				summary.StoppedContainers = infoData.ContainersStopped
				summary.ImagesCount = infoData.Images
				summary.Version = infoData.ServerVersion
			}
		}

		// 2. Опрос /containers/json?all=1
		respC, errC := client.Get("http://localhost/containers/json?all=1")
		if errC == nil && respC.StatusCode == http.StatusOK {
			defer respC.Body.Close()
			bodyC, _ := io.ReadAll(respC.Body)

			var rawContainers []struct {
				ID     string            `json:"Id"`
				Names  []string          `json:"Names"`
				Image  string            `json:"Image"`
				State  string            `json:"State"`
				Status string            `json:"Status"`
				Ports  []rawPort         `json:"Ports"`
				Labels map[string]string `json:"Labels"`
			}

			if json.Unmarshal(bodyC, &rawContainers) == nil {
				summary.Containers = make([]ContainerItem, 0, len(rawContainers))
				for _, c := range rawContainers {
					cID := c.ID
					if len(cID) > 12 {
						cID = cID[:12]
					}

					name := ""
					if len(c.Names) > 0 {
						name = strings.TrimPrefix(c.Names[0], "/")
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
						Ports:          formatPorts(c.Ports),
						ComposeProject: composeProject,
						ComposeService: composeService,
						Labels:         c.Labels,
					})
				}

				// Параллельный опрос метрик CPU и памяти с ограниченным числом горутин
				if len(summary.Containers) > 0 {
					var wg sync.WaitGroup
					sem := make(chan struct{}, maxConcurrentStatsFetch)
					for i := range summary.Containers {
						if summary.Containers[i].State == "running" {
							wg.Add(1)
							sem <- struct{}{}
							go func(idx int, cID string) {
								defer wg.Done()
								defer func() { <-sem }()
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
			summary.Status = DockerStatusRunning
			summary.StatusDisplay = DockerStatusDisplayRunning
			summary.Version = parts[4]
		}
	}

	return summary
}

// containerAction выполняет стандартное действие жизненного цикла контейнера через сокет с фоллбэком на CLI.
func containerAction(containerID, action string, validCodes ...int) error {
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/%s", containerID, action)
		resp, err := client.Post(url, "application/json", nil)
		if err == nil {
			defer resp.Body.Close()
			for _, code := range validCodes {
				if resp.StatusCode == code {
					return nil
				}
			}
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	out, err := exec.Command("docker", action, containerID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// StartContainer запускает контейнер по ID.
func StartContainer(containerID string) error {
	return containerAction(containerID, "start", http.StatusNoContent, http.StatusNotModified)
}

// StopContainer останавливает контейнер по ID.
func StopContainer(containerID string) error {
	return containerAction(containerID, "stop", http.StatusNoContent, http.StatusNotModified)
}

// RestartContainer перезапускает контейнер по ID.
func RestartContainer(containerID string) error {
	return containerAction(containerID, "restart", http.StatusNoContent)
}

// GetContainerLogs читает последние N строк логов контейнера.
func GetContainerLogs(containerID string, tailLines int) (string, error) {
	if strings.TrimSpace(containerID) == "" {
		return "", fmt.Errorf("container ID cannot be empty")
	}
	if tailLines <= 0 {
		tailLines = defaultTailLines
	}

	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/logs?stdout=1&stderr=1&tail=%d", containerID, tailLines)
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
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
		if size < 0 || idx+8+size > len(all) {
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
	if strings.TrimSpace(containerID) == "" {
		return detail, fmt.Errorf("container ID cannot be empty")
	}
	detail.ID = containerID

	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/json", containerID)
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
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
	if strings.TrimSpace(containerID) == "" {
		return nil
	}
	return exec.Command("docker", "exec", "-it", containerID, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi")
}
