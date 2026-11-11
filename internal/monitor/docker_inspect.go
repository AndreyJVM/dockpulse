package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type socketInspectData struct {
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

func getContainerDetailViaSocket(sockPath, containerID string) (ContainerDetail, error) {
	var detail ContainerDetail
	detail.ID = containerID

	client := socketClientFactory(sockPath)
	url := fmt.Sprintf("http://localhost/containers/%s/json", containerID)
	resp, err := client.Get(url)
	if err != nil {
		return detail, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return detail, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return detail, err
	}

	var data socketInspectData
	if err := json.Unmarshal(body, &data); err != nil {
		return detail, err
	}

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

func getContainerDetailViaCLI(containerID string) (ContainerDetail, error) {
	var detail ContainerDetail
	detail.ID = containerID

	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "inspect", containerID)
	if err != nil {
		return detail, fmt.Errorf("container not found: %s", containerID)
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(out, &raw); err != nil || len(raw) == 0 {
		return detail, fmt.Errorf("container not found: %s", containerID)
	}

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

// GetContainerDetail returns detailed inspection data for a container.
func GetContainerDetail(containerID string) (ContainerDetail, error) {
	var detail ContainerDetail
	if strings.TrimSpace(containerID) == "" {
		return detail, fmt.Errorf("container ID cannot be empty")
	}
	detail.ID = containerID

	sockPath := dockerSocketFinder()
	if sockPath != "" {
		if d, err := getContainerDetailViaSocket(sockPath, containerID); err == nil {
			return d, nil
		}
	}

	return getContainerDetailViaCLI(containerID)
}
