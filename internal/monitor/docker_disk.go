package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DockerDiskItem represents disk space information for a single Docker resource category.
type DockerDiskItem struct {
	TotalCount  int
	ActiveCount int
	TotalSize   int64
	Reclaimable int64
}

// DockerDiskUsage describes the total disk space usage across Docker images, containers, volumes, and build cache.
type DockerDiskUsage struct {
	Images     DockerDiskItem
	Containers DockerDiskItem
	Volumes    DockerDiskItem
	BuildCache DockerDiskItem
	TotalSize  int64
	TotalRecl  int64
}

// FormatBytes formats a byte count into a human-readable string (B, KB, MB, GB, TB).
func FormatBytes(b int64) string {
	const unit = 1024
	if b < 0 {
		b = 0
	}
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB"}
	if exp < len(suffixes) {
		return fmt.Sprintf("%.1f %s", float64(b)/float64(div), suffixes[exp])
	}
	return fmt.Sprintf("%.1f TB", float64(b)/float64(div))
}

type systemDFRaw struct {
	LayersSize int64 `json:"LayersSize"`
	Images     []struct {
		Size       int64 `json:"Size"`
		SharedSize int64 `json:"SharedSize"`
		Containers int64 `json:"Containers"`
	} `json:"Images"`
	Containers []struct {
		SizeRw     int64  `json:"SizeRw"`
		SizeRootFs int64  `json:"SizeRootFs"`
		State      string `json:"State"`
	} `json:"Containers"`
	Volumes []struct {
		UsageData struct {
			Size     int64 `json:"Size"`
			RefCount int64 `json:"RefCount"`
		} `json:"UsageData"`
	} `json:"Volumes"`
	BuildCache []struct {
		Size        int64 `json:"Size"`
		Reclaimable bool  `json:"Reclaimable"`
	} `json:"BuildCache"`
}

// ParseSystemDFRaw calculates aggregated disk usage metrics from Docker Engine API system/df response.
func ParseSystemDFRaw(raw systemDFRaw) DockerDiskUsage {
	var usage DockerDiskUsage

	// Images
	usage.Images.TotalCount = len(raw.Images)
	for _, img := range raw.Images {
		usage.Images.TotalSize += img.Size
		if img.Containers == 0 {
			usage.Images.Reclaimable += img.Size
		} else {
			usage.Images.ActiveCount++
		}
	}

	// Containers
	usage.Containers.TotalCount = len(raw.Containers)
	for _, c := range raw.Containers {
		usage.Containers.TotalSize += c.SizeRw
		if c.State != "running" {
			usage.Containers.Reclaimable += c.SizeRw
		} else {
			usage.Containers.ActiveCount++
		}
	}

	// Volumes
	usage.Volumes.TotalCount = len(raw.Volumes)
	for _, v := range raw.Volumes {
		usage.Volumes.TotalSize += v.UsageData.Size
		if v.UsageData.RefCount <= 0 {
			usage.Volumes.Reclaimable += v.UsageData.Size
		} else {
			usage.Volumes.ActiveCount++
		}
	}

	// BuildCache
	usage.BuildCache.TotalCount = len(raw.BuildCache)
	for _, bc := range raw.BuildCache {
		usage.BuildCache.TotalSize += bc.Size
		if bc.Reclaimable {
			usage.BuildCache.Reclaimable += bc.Size
		} else {
			usage.BuildCache.ActiveCount++
		}
	}

	usage.TotalSize = usage.Images.TotalSize + usage.Containers.TotalSize + usage.Volumes.TotalSize + usage.BuildCache.TotalSize
	usage.TotalRecl = usage.Images.Reclaimable + usage.Containers.Reclaimable + usage.Volumes.Reclaimable + usage.BuildCache.Reclaimable

	return usage
}

// GetDockerDiskUsage collects Docker disk usage statistics (system df) via socket or CLI fallback.
func GetDockerDiskUsage() (DockerDiskUsage, error) {
	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
		ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/system/df", nil)
		if err == nil {
			resp, doErr := client.Do(req)
			if doErr == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				body, readErr := io.ReadAll(resp.Body)
				if readErr == nil {
					var raw systemDFRaw
					if unmarshalErr := json.Unmarshal(body, &raw); unmarshalErr == nil {
						return ParseSystemDFRaw(raw), nil
					}
				}
			}
		}
	}

	// CLI fallback
	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "system", "df", "--format", "{{json .}}")
	if err == nil && len(out) > 0 {
		var usage DockerDiskUsage
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			var item struct {
				Type        string `json:"Type"`
				TotalCount  string `json:"TotalCount"`
				Active      string `json:"Active"`
				Size        string `json:"Size"`
				Reclaimable string `json:"Reclaimable"`
			}
			if json.Unmarshal([]byte(line), &item) == nil {
				switch strings.ToLower(item.Type) {
				case "images":
					usage.Images.TotalCount = parseCLIInt(item.TotalCount)
					usage.Images.ActiveCount = parseCLIInt(item.Active)
				case "containers":
					usage.Containers.TotalCount = parseCLIInt(item.TotalCount)
					usage.Containers.ActiveCount = parseCLIInt(item.Active)
				case "local volumes", "volumes":
					usage.Volumes.TotalCount = parseCLIInt(item.TotalCount)
					usage.Volumes.ActiveCount = parseCLIInt(item.Active)
				case "build cache":
					usage.BuildCache.TotalCount = parseCLIInt(item.TotalCount)
					usage.BuildCache.ActiveCount = parseCLIInt(item.Active)
				}
			}
		}
		return usage, nil
	}

	return DockerDiskUsage{}, fmt.Errorf("служба Docker недоступна")
}

func parseCLIInt(s string) int {
	var n int
	fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n
}
