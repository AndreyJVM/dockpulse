package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
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
