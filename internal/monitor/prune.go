package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// DockerDiskItem содержит информацию об одной категории дисковых ресурсов Docker.
type DockerDiskItem struct {
	TotalCount  int
	ActiveCount int
	TotalSize   int64
	Reclaimable int64
}

// DockerDiskUsage описывает общую сводку использования дискового пространства Docker.
type DockerDiskUsage struct {
	Images     DockerDiskItem
	Containers DockerDiskItem
	Volumes    DockerDiskItem
	BuildCache DockerDiskItem
	TotalSize  int64
	TotalRecl  int64
}

// FormatBytes форматирует размер в байтах в человекочитаемую строку (B, KB, MB, GB).
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

// ParseSystemDFRaw рассчитывает агрегированные дисковые показатели из структуры ответа Docker API.
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

// GetDockerDiskUsage собирает статистику использования диска Docker (system df).
func GetDockerDiskUsage() (DockerDiskUsage, error) {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost/system/df", nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				var raw systemDFRaw
				if err := json.Unmarshal(body, &raw); err == nil {
					return ParseSystemDFRaw(raw), nil
				}
			}
		}
	}

	// CLI fallback
	if IsDockerCLIAvailable() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "docker", "system", "df", "--format", "{{json .}}")
		out, err := cmd.CombinedOutput()
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
	}

	return DockerDiskUsage{}, fmt.Errorf("служба Docker недоступна")
}

func parseCLIInt(s string) int {
	var n int
	fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n
}

// PruneContainers удаляет остановленные контейнеры.
func PruneContainers() (string, error) {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		resp, err := client.Post("http://localhost/containers/prune", "application/json", nil)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var res struct {
				ContainersDeleted []string `json:"ContainersDeleted"`
				SpaceReclaimed    int64    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(resp.Body).Decode(&res) == nil {
				count := len(res.ContainersDeleted)
				return fmt.Sprintf("Удалено контейнеров: %d, освобождено: %s", count, FormatBytes(res.SpaceReclaimed)), nil
			}
		}
	}

	// CLI fallback
	out, err := exec.Command("docker", "container", "prune", "-f").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return formatPruneCLIOutput(string(out), "контейнеров"), nil
}

// PruneImages удаляет висячие (dangling) образы Docker.
func PruneImages() (string, error) {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		resp, err := client.Post("http://localhost/images/prune", "application/json", nil)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var res struct {
				ImagesDeleted  []map[string]interface{} `json:"ImagesDeleted"`
				SpaceReclaimed int64                    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(resp.Body).Decode(&res) == nil {
				count := len(res.ImagesDeleted)
				return fmt.Sprintf("Удалено образов: %d, освобождено: %s", count, FormatBytes(res.SpaceReclaimed)), nil
			}
		}
	}

	// CLI fallback
	out, err := exec.Command("docker", "image", "prune", "-f").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return formatPruneCLIOutput(string(out), "образов"), nil
}

// PruneVolumes удаляет неиспользуемые тома Docker.
func PruneVolumes() (string, error) {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		resp, err := client.Post("http://localhost/volumes/prune", "application/json", nil)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var res struct {
				VolumesDeleted []string `json:"VolumesDeleted"`
				SpaceReclaimed int64    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(resp.Body).Decode(&res) == nil {
				count := len(res.VolumesDeleted)
				return fmt.Sprintf("Удалено томов: %d, освобождено: %s", count, FormatBytes(res.SpaceReclaimed)), nil
			}
		}
	}

	// CLI fallback
	out, err := exec.Command("docker", "volume", "prune", "-f").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return formatPruneCLIOutput(string(out), "томов"), nil
}

// PruneAll выполняет полную очистку неиспользуемых контейнеров, образов, томов и кэша сборки.
func PruneAll() (string, error) {
	sockPath := FindDockerSocket()
	if sockPath != "" {
		var totalReclaimed int64
		var details []string

		// 1. Containers
		if res, err := PruneContainers(); err == nil {
			details = append(details, res)
		}
		// 2. Images
		if res, err := PruneImages(); err == nil {
			details = append(details, res)
		}
		// 3. Volumes
		if res, err := PruneVolumes(); err == nil {
			details = append(details, res)
		}
		// 4. Build Cache
		client := getSocketHTTPClient(sockPath)
		if resp, err := client.Post("http://localhost/build/prune", "application/json", nil); err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var res struct {
				CachesDeleted  []string `json:"CachesDeleted"`
				SpaceReclaimed int64    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(resp.Body).Decode(&res) == nil && len(res.CachesDeleted) > 0 {
				totalReclaimed += res.SpaceReclaimed
				details = append(details, fmt.Sprintf("Кэш сборки: %d записей (%s)", len(res.CachesDeleted), FormatBytes(res.SpaceReclaimed)))
			}
		}

		if len(details) > 0 {
			return strings.Join(details, " | "), nil
		}
	}

	// CLI fallback
	out, err := exec.Command("docker", "system", "prune", "-f", "--volumes").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return formatPruneCLIOutput(string(out), "ресурсов"), nil
}

func formatPruneCLIOutput(out string, label string) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, l := range lines {
		if strings.Contains(l, "Total reclaimed space:") {
			space := strings.TrimPrefix(l, "Total reclaimed space:")
			return fmt.Sprintf("Очистка %s завершена. Освобождено:%s", label, space)
		}
	}
	return fmt.Sprintf("Очистка %s успешно выполнена", label)
}
