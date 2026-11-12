package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	containerStatsTimeout = 2 * time.Second
)

// ContainerStatsRaw represents the Docker Engine API container stats response.
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

// ContainerStats contains calculated CPU, memory, network, and block I/O metrics.
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

	// Calculate CPU usage as a percentage.
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(s.CPUStats.SystemCPUUsage) - float64(s.PreCPUStats.SystemCPUUsage)
	onlineCPUs := s.CPUStats.OnlineCPUs
	if onlineCPUs <= 0 {
		onlineCPUs = 1
	}

	if systemDelta > 0 && cpuDelta > 0 {
		stats.CPUPerc = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
	}

	// Calculate memory usage.
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

	// Calculate network I/O.
	if len(s.Networks) > 0 {
		for _, net := range s.Networks {
			stats.NetRxBytes += net.RxBytes
			stats.NetTxBytes += net.TxBytes
		}
	} else {
		stats.NetRxBytes = s.Network.RxBytes
		stats.NetTxBytes = s.Network.TxBytes
	}

	// Calculate block I/O.
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

// fetchContainerInspect reads a compact set of container metadata through the socket.
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
