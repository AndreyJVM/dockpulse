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
	ID              string
	Name            string
	Image           string
	State           string
	Status          string
	Health          string
	Ports           string
	ComposeProject  string
	ComposeService  string
	Labels          map[string]string
	Mounts          []string
	IPAddress       string
	NetworkName     string
	Created         string
	StartedAt       string
	RestartPolicy   string
	CPUPerc         float64
	MemUsageMB      float64
	MemLimitMB      float64
	MemPerc         float64
	NetRxBytes      uint64
	NetTxBytes      uint64
	BlockReadBytes  uint64
	BlockWriteBytes uint64
}

// ContainerDetail содержит детальную информацию об одном контейнере.
type ContainerDetail struct {
	ID              string
	Name            string
	Image           string
	Created         string
	State           string
	Status          string
	Health          string
	IP              string
	Gateway         string
	MACAddress      string
	RestartPolicy   string
	Command         string
	Mounts          []string
	Env             []string
	CPUPerc         float64
	MemUsageMB      float64
	MemPerc         float64
	NetRxBytes      uint64
	NetTxBytes      uint64
	BlockReadBytes  uint64
	BlockWriteBytes uint64
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
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
	Network struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"network"`
	BlkioStats struct {
		IOServiceBytesRecursive []struct {
			Major uint64 `json:"major"`
			Minor uint64 `json:"minor"`
			Op    string `json:"op"`
			Value uint64 `json:"value"`
		} `json:"io_service_bytes_recursive"`
	} `json:"blkio_stats"`
}

// ContainerStats содержит вычисленные метрики контейнера (CPU, RAM, Net I/O, Block I/O).
type ContainerStats struct {
	CPUPerc         float64
	MemUsageMB      float64
	MemLimitMB      float64
	MemPerc         float64
	NetRxBytes      uint64
	NetTxBytes      uint64
	BlockReadBytes  uint64
	BlockWriteBytes uint64
}

func calculateContainerStats(s ContainerStatsRaw) ContainerStats {
	var stats ContainerStats
	const mebiByte = 1024 * 1024

	// Расчет загрузки CPU в процентах
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(s.CPUStats.SystemCPUUsage) - float64(s.PreCPUStats.SystemCPUUsage)
	onlineCPUs := s.CPUStats.OnlineCPUs
	if onlineCPUs <= 0 {
		onlineCPUs = 1
	}

	if systemDelta > 0 && cpuDelta > 0 {
		stats.CPUPerc = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
	}

	// Расчет потребления оперативной памяти
	memUsage := s.MemoryStats.Usage
	if s.MemoryStats.Stats.TotalInactiveFile > 0 && memUsage > s.MemoryStats.Stats.TotalInactiveFile {
		memUsage -= s.MemoryStats.Stats.TotalInactiveFile
	} else if s.MemoryStats.Stats.InactiveFile > 0 && memUsage > s.MemoryStats.Stats.InactiveFile {
		memUsage -= s.MemoryStats.Stats.InactiveFile
	}

	stats.MemUsageMB = float64(memUsage) / mebiByte
	stats.MemLimitMB = float64(s.MemoryStats.Limit) / mebiByte
	if s.MemoryStats.Limit > 0 {
		stats.MemPerc = (float64(memUsage) / float64(s.MemoryStats.Limit)) * 100.0
	}

	// Расчет Network I/O
	if len(s.Networks) > 0 {
		for _, net := range s.Networks {
			stats.NetRxBytes += net.RxBytes
			stats.NetTxBytes += net.TxBytes
		}
	} else {
		stats.NetRxBytes = s.Network.RxBytes
		stats.NetTxBytes = s.Network.TxBytes
	}

	// Расчет Block I/O
	for _, bio := range s.BlkioStats.IOServiceBytesRecursive {
		switch strings.ToLower(bio.Op) {
		case "read":
			stats.BlockReadBytes += bio.Value
		case "write":
			stats.BlockWriteBytes += bio.Value
		}
	}

	return stats
}

func fetchContainerStats(client *http.Client, containerID string) ContainerStats {
	ctx, cancel := context.WithTimeout(context.Background(), containerStatsTimeout)
	defer cancel()

	url := fmt.Sprintf("http://localhost/containers/%s/stats?stream=false&one-shot=true", containerID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ContainerStats{}
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ContainerStats{}
	}
	defer resp.Body.Close()

	var s ContainerStatsRaw
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return ContainerStats{}
	}

	return calculateContainerStats(s)
}

// fetchContainerInspect запрашивает краткие метаданные инспекции контейнера через сокет.
func fetchContainerInspect(client *http.Client, containerID string) (restartPolicy, startedAt string) {
	ctx, cancel := context.WithTimeout(context.Background(), containerStatsTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost/containers/%s/json", containerID), nil)
	if err != nil {
		return "", ""
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return "", ""
	}
	defer resp.Body.Close()

	var data struct {
		HostConfig struct {
			RestartPolicy struct {
				Name string `json:"Name"`
			} `json:"RestartPolicy"`
		} `json:"HostConfig"`
		State struct {
			StartedAt string `json:"StartedAt"`
		} `json:"State"`
	}

	if json.NewDecoder(resp.Body).Decode(&data) == nil {
		started := data.State.StartedAt
		if t, err := time.Parse(time.RFC3339Nano, started); err == nil && !t.IsZero() {
			started = t.Local().Format("2006-01-02 15:04:05")
		}
		return data.HostConfig.RestartPolicy.Name, started
	}
	return "", ""
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

var (
	dockerSocketFinder   = FindDockerSocket
	socketClientFactory  = getSocketHTTPClient
	dockerCLICommandRunner = func(ctx context.Context, args ...string) ([]byte, error) {
		if ctx == nil {
			return exec.Command("docker", args...).CombinedOutput()
		}
		return exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	}
)

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
	seen := make(map[string]bool)
	for _, p := range ports {
		var entry string
		if p.PublicPort > 0 {
			entry = fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type)
		} else if p.PrivatePort > 0 {
			entry = fmt.Sprintf("%d/%s", p.PrivatePort, p.Type)
		}
		if entry != "" && !seen[entry] {
			seen[entry] = true
			portsList = append(portsList, entry)
		}
	}
	return strings.Join(portsList, ", ")
}

// GetDockerSummary собирает статус службы Docker и список контейнеров с живыми метриками.
func GetDockerSummary() DockerSummary {
	_, cliErr := exec.LookPath("docker")
	sockPath := dockerSocketFinder()

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
		client := socketClientFactory(sockPath)

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
				Created int64            `json:"Created"`
				NetworkSettings struct {
					Networks map[string]struct {
						IPAddress string `json:"IPAddress"`
						Gateway   string `json:"Gateway"`
					} `json:"Networks"`
				} `json:"NetworkSettings"`
				Mounts []struct {
					Type        string `json:"Type"`
					Name        string `json:"Name"`
					Source      string `json:"Source"`
					Destination string `json:"Destination"`
					Driver      string `json:"Driver"`
					Mode        string `json:"Mode"`
					RW          bool   `json:"RW"`
				} `json:"Mounts"`
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

					var mountsList []string
					for _, m := range c.Mounts {
						src := m.Source
						if src == "" {
							src = m.Name
						}
						mode := m.Mode
						if mode == "" {
							if m.RW {
								mode = "rw"
							} else {
								mode = "ro"
							}
						}
						if src != "" && m.Destination != "" {
							mountsList = append(mountsList, fmt.Sprintf("%s -> %s (%s)", src, m.Destination, mode))
						} else if m.Destination != "" {
							mountsList = append(mountsList, m.Destination)
						}
					}

					var ipAddr, netName string
					if len(c.NetworkSettings.Networks) > 0 {
						var ips []string
						var nets []string
						for nName, nData := range c.NetworkSettings.Networks {
							nets = append(nets, nName)
							if nData.IPAddress != "" {
								ips = append(ips, nData.IPAddress)
							}
						}
						ipAddr = strings.Join(ips, ", ")
						netName = strings.Join(nets, ", ")
					}

					createdStr := ""
					if c.Created > 0 {
						createdStr = time.Unix(c.Created, 0).Format("2006-01-02 15:04:05")
					}

					summary.Containers = append(summary.Containers, ContainerItem{
						ID:             cID,
						Name:           name,
						Image:          c.Image,
						State:          c.State,
						Status:         c.Status,
						Health:         ParseHealthStatus(c.Status),
						Ports:          formatPorts(c.Ports),
						Mounts:         mountsList,
						ComposeProject: composeProject,
						ComposeService: composeService,
						Labels:         c.Labels,
						IPAddress:      ipAddr,
						NetworkName:    netName,
						Created:        createdStr,
					})
				}

				// Параллельный опрос метрик и инспекции с ограниченным числом горутин
				if len(summary.Containers) > 0 {
					var wg sync.WaitGroup
					sem := make(chan struct{}, maxConcurrentStatsFetch)
					for i := range summary.Containers {
						wg.Add(1)
						sem <- struct{}{}
						go func(idx int, cID, state string) {
							defer wg.Done()
							defer func() { <-sem }()
							if state == "running" {
								stats := fetchContainerStats(client, cID)
								summary.Containers[idx].CPUPerc = stats.CPUPerc
								summary.Containers[idx].MemUsageMB = stats.MemUsageMB
								summary.Containers[idx].MemLimitMB = stats.MemLimitMB
								summary.Containers[idx].MemPerc = stats.MemPerc
								summary.Containers[idx].NetRxBytes = stats.NetRxBytes
								summary.Containers[idx].NetTxBytes = stats.NetTxBytes
								summary.Containers[idx].BlockReadBytes = stats.BlockReadBytes
								summary.Containers[idx].BlockWriteBytes = stats.BlockWriteBytes
							}
							rp, started := fetchContainerInspect(client, cID)
							if rp != "" {
								summary.Containers[idx].RestartPolicy = rp
							}
							if started != "" {
								summary.Containers[idx].StartedAt = started
							}
						}(i, summary.Containers[i].ID, summary.Containers[i].State)
					}
					wg.Wait()
				}
			}
			return summary
		}
	}

	// Fallback CLI
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "info", "--format", "{{.Containers}} {{.ContainersRunning}} {{.ContainersStopped}} {{.Images}} {{.ServerVersion}}")
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 5 {
			summary.Status = DockerStatusRunning
			summary.StatusDisplay = DockerStatusDisplayRunning
			summary.Version = parts[4]
		}
	}

	if summary.Status == DockerStatusRunning && len(summary.Containers) == 0 {
		psCtx, psCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer psCancel()
		psOut, psErr := dockerCLICommandRunner(psCtx, "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}\t{{.Mounts}}\t{{.Labels}}\t{{.Networks}}\t{{.CreatedAt}}\t{{.RunningFor}}")
		if psErr == nil {
			lines := strings.Split(strings.TrimSpace(string(psOut)), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				parts := strings.Split(line, "\t")
				if len(parts) >= 6 {
					cID := parts[0]
					cName := parts[1]
					cImg := parts[2]
					cState := parts[3]
					cStatus := parts[4]
					cPorts := parts[5]
					var cMounts []string
					if len(parts) >= 7 && parts[6] != "" {
						for _, m := range strings.Split(parts[6], ",") {
							if tr := strings.TrimSpace(m); tr != "" {
								cMounts = append(cMounts, tr)
							}
						}
					}
					var pProj, pSvc string
					if len(parts) >= 8 {
						for _, lbl := range strings.Split(parts[7], ",") {
							kv := strings.SplitN(lbl, "=", 2)
							if len(kv) == 2 {
								if kv[0] == "com.docker.compose.project" {
									pProj = kv[1]
								} else if kv[0] == "com.docker.compose.service" {
									pSvc = kv[1]
								}
							}
						}
					}
					var cNet, cCreated, cStarted string
					if len(parts) >= 9 {
						cNet = parts[8]
					}
					if len(parts) >= 10 {
						cCreated = parts[9]
					}
					if len(parts) >= 11 {
						cStarted = parts[10]
					}
					summary.Containers = append(summary.Containers, ContainerItem{
						ID:             cID,
						Name:           cName,
						Image:          cImg,
						State:          cState,
						Status:         cStatus,
						Health:         ParseHealthStatus(cStatus),
						Ports:          cPorts,
						Mounts:         cMounts,
						ComposeProject: pProj,
						ComposeService: pSvc,
						NetworkName:    cNet,
						Created:        cCreated,
						StartedAt:      cStarted,
					})
				}
			}
		}
		if len(summary.Containers) > 0 {
			inspectCtx, inspectCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer inspectCancel()
			inspectArgs := []string{"inspect", "--format", "{{slice .Id 0 12}}|{{.HostConfig.RestartPolicy.Name}}|{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}"}
			for _, c := range summary.Containers {
				inspectArgs = append(inspectArgs, c.ID)
			}
			if insOut, insErr := dockerCLICommandRunner(inspectCtx, inspectArgs...); insErr == nil {
				for _, line := range strings.Split(strings.TrimSpace(string(insOut)), "\n") {
					f := strings.Split(line, "|")
					if len(f) >= 3 {
						cID := strings.TrimSpace(f[0])
						rPolicy := strings.TrimSpace(f[1])
						ips := strings.Fields(f[2])
						for idx := range summary.Containers {
							if summary.Containers[idx].ID == cID {
								summary.Containers[idx].RestartPolicy = rPolicy
								if len(ips) > 0 && summary.Containers[idx].IPAddress == "" {
									summary.Containers[idx].IPAddress = strings.Join(ips, ", ")
								}
								break
							}
						}
					}
				}
			}
		}
	}

	return summary
}

// containerAction выполняет стандартное действие жизненного цикла контейнера через сокет с фоллбэком на CLI.
func containerAction(containerID, action string, validCodes ...int) error {
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, action, containerID)
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

// KillContainer принудительно завершает контейнер по ID (SIGKILL).
func KillContainer(containerID string) error {
	return containerAction(containerID, "kill", http.StatusNoContent)
}

// RemoveContainer удаляет контейнер по ID (с флагом force=true).
func RemoveContainer(containerID string) error {
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s?v=1&force=true", containerID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
					return nil
				}
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "rm", "-f", containerID)
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

// GetContainerLogs читает последние N строк логов контейнера.
func GetContainerLogs(containerID string, tailLines int) (string, error) {
	if strings.TrimSpace(containerID) == "" {
		return "", fmt.Errorf("container ID cannot be empty")
	}
	if tailLines <= 0 {
		tailLines = defaultTailLines
	}

	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
		url := fmt.Sprintf("http://localhost/containers/%s/logs?stdout=1&stderr=1&tail=%d", containerID, tailLines)
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return demuxDockerLogs(resp.Body), nil
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "logs", "--tail", fmt.Sprintf("%d", tailLines), containerID)
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

	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
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
				HostConfig struct {
					RestartPolicy struct {
						Name string `json:"Name"`
					} `json:"RestartPolicy"`
				} `json:"HostConfig"`
				NetworkSettings struct {
					IPAddress  string `json:"IPAddress"`
					Gateway    string `json:"Gateway"`
					MacAddress string `json:"MacAddress"`
					Networks   map[string]struct {
						IPAddress  string `json:"IPAddress"`
						Gateway    string `json:"Gateway"`
						MacAddress string `json:"MacAddress"`
					} `json:"Networks"`
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
				detail.RestartPolicy = data.HostConfig.RestartPolicy.Name
				detail.IP = data.NetworkSettings.IPAddress
				detail.Gateway = data.NetworkSettings.Gateway
				detail.MACAddress = data.NetworkSettings.MacAddress
				if detail.IP == "" {
					for _, net := range data.NetworkSettings.Networks {
						if net.IPAddress != "" {
							detail.IP = net.IPAddress
							if detail.Gateway == "" {
								detail.Gateway = net.Gateway
							}
							if detail.MACAddress == "" {
								detail.MACAddress = net.MacAddress
							}
							break
						}
					}
				}
				detail.Command = strings.Join(data.Config.Cmd, " ")

				for _, m := range data.Mounts {
					detail.Mounts = append(detail.Mounts, fmt.Sprintf("%s -> %s (%s)", m.Source, m.Destination, m.Mode))
				}
				detail.Env = data.Config.Env

				// Опрос метрик если контейнер запущен
				if detail.State == "running" {
					stats := fetchContainerStats(client, containerID)
					detail.CPUPerc = stats.CPUPerc
					detail.MemUsageMB = stats.MemUsageMB
					detail.MemPerc = stats.MemPerc
					detail.NetRxBytes = stats.NetRxBytes
					detail.NetTxBytes = stats.NetTxBytes
					detail.BlockReadBytes = stats.BlockReadBytes
					detail.BlockWriteBytes = stats.BlockWriteBytes
				}

				return detail, nil
			}
		}
	}

	// CLI inspect fallback
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "inspect", containerID)
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
			if hostConfig, ok := first["HostConfig"].(map[string]interface{}); ok {
				if rp, ok := hostConfig["RestartPolicy"].(map[string]interface{}); ok {
					if name, ok := rp["Name"].(string); ok {
						detail.RestartPolicy = name
					}
				}
			}
			if detail.IP == "" {
				if netSettings, ok := first["NetworkSettings"].(map[string]interface{}); ok {
					if nets, ok := netSettings["Networks"].(map[string]interface{}); ok {
						for _, n := range nets {
							if nMap, ok := n.(map[string]interface{}); ok {
								if ip, ok := nMap["IPAddress"].(string); ok && ip != "" {
									detail.IP = ip
									if gw, ok := nMap["Gateway"].(string); ok && detail.Gateway == "" {
										detail.Gateway = gw
									}
									if mac, ok := nMap["MacAddress"].(string); ok && detail.MACAddress == "" {
										detail.MACAddress = mac
									}
									break
								}
							}
						}
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

// ExecOptions specifies options for launching an interactive session in a container.
type ExecOptions struct {
	Mode       string // "auto", "sh", "zsh", "root", "attach", "custom"
	User       string // optional user (e.g., "0" or "root")
	CustomCmd  string // optional custom command when Mode == "custom"
}

// GetContainerExecCommandWithOptions creates an exec.Cmd based on specified ExecOptions.
func GetContainerExecCommandWithOptions(containerID string, opts ExecOptions) *exec.Cmd {
	cleanID := strings.TrimSpace(containerID)
	if cleanID == "" {
		return nil
	}

	if opts.Mode == "attach" {
		return exec.Command("docker", "attach", "--sig-proxy=false", cleanID)
	}

	args := []string{"exec", "-it"}
	if opts.User != "" {
		args = append(args, "-u", opts.User)
	}
	args = append(args, cleanID)

	switch opts.Mode {
	case "sh":
		args = append(args, "sh")
	case "zsh":
		args = append(args, "zsh")
	case "root":
		if opts.User == "" {
			// Insert -u 0 before container ID if not already present
			args = []string{"exec", "-it", "-u", "0", cleanID, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi"}
		} else {
			args = append(args, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi")
		}
	case "custom":
		cmdStr := strings.TrimSpace(opts.CustomCmd)
		if cmdStr == "" {
			cmdStr = "sh"
		}
		args = append(args, "sh", "-c", cmdStr)
	case "auto":
		fallthrough
	default:
		args = append(args, "sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi")
	}

	return exec.Command("docker", args...)
}

// GetContainerExecCommand возвращает команду exec.Cmd для запуска интерактивного шелла в контейнере.
func GetContainerExecCommand(containerID string) *exec.Cmd {
	return GetContainerExecCommandWithOptions(containerID, ExecOptions{Mode: "auto"})
}
