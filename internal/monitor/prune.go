package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	pruneTimeout              = 30 * time.Second
	totalReclaimedSpacePrefix = "Total reclaimed space:"
)

type pruneResourceConfig struct {
	socketPath     string
	cliResource    string
	label          string
	parseSocketRes func(r io.Reader) (count int, reclaimed int64, ok bool)
}

func executePruneResource(cfg pruneResourceConfig) (string, error) {
	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
		ctx, cancel := context.WithTimeout(context.Background(), pruneTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/"+cfg.socketPath, nil)
		if err == nil {
			resp, doErr := client.Do(req)
			if doErr == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				if count, reclaimed, ok := cfg.parseSocketRes(resp.Body); ok {
					return fmt.Sprintf("Удалено %s: %d, освобождено: %s", cfg.label, count, FormatBytes(reclaimed)), nil
				}
			}
		}
	}

	// CLI fallback
	ctx, cancel := context.WithTimeout(context.Background(), pruneTimeout)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, cfg.cliResource, "prune", "-f")
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return formatPruneCLIOutput(string(out), cfg.label), nil
}

// PruneContainers removes all stopped containers.
func PruneContainers() (string, error) {
	return executePruneResource(pruneResourceConfig{
		socketPath:  "containers/prune",
		cliResource: "container",
		label:       "контейнеров",
		parseSocketRes: func(r io.Reader) (int, int64, bool) {
			var res struct {
				ContainersDeleted []string `json:"ContainersDeleted"`
				SpaceReclaimed    int64    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(r).Decode(&res) == nil {
				return len(res.ContainersDeleted), res.SpaceReclaimed, true
			}
			return 0, 0, false
		},
	})
}

// PruneImages removes dangling Docker images.
func PruneImages() (string, error) {
	return executePruneResource(pruneResourceConfig{
		socketPath:  "images/prune",
		cliResource: "image",
		label:       "образов",
		parseSocketRes: func(r io.Reader) (int, int64, bool) {
			var res struct {
				ImagesDeleted  []map[string]interface{} `json:"ImagesDeleted"`
				SpaceReclaimed int64                    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(r).Decode(&res) == nil {
				return len(res.ImagesDeleted), res.SpaceReclaimed, true
			}
			return 0, 0, false
		},
	})
}

// PruneVolumes removes unused Docker volumes.
func PruneVolumes() (string, error) {
	return executePruneResource(pruneResourceConfig{
		socketPath:  "volumes/prune",
		cliResource: "volume",
		label:       "томов",
		parseSocketRes: func(r io.Reader) (int, int64, bool) {
			var res struct {
				VolumesDeleted []string `json:"VolumesDeleted"`
				SpaceReclaimed int64    `json:"SpaceReclaimed"`
			}
			if json.NewDecoder(r).Decode(&res) == nil {
				return len(res.VolumesDeleted), res.SpaceReclaimed, true
			}
			return 0, 0, false
		},
	})
}

func pruneBuildCacheSocket(sockPath string) string {
	client := socketClientFactory(sockPath)
	ctx, cancel := context.WithTimeout(context.Background(), pruneTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/build/prune", nil)
	if err != nil {
		return ""
	}
	resp, doErr := client.Do(req)
	if doErr != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var res struct {
		CachesDeleted  []string `json:"CachesDeleted"`
		SpaceReclaimed int64    `json:"SpaceReclaimed"`
	}
	if json.NewDecoder(resp.Body).Decode(&res) == nil && len(res.CachesDeleted) > 0 {
		return fmt.Sprintf("Кэш сборки: %d записей (%s)", len(res.CachesDeleted), FormatBytes(res.SpaceReclaimed))
	}
	return ""
}

// PruneAll performs cleanup of unused containers, images, volumes, and build cache.
func PruneAll() (string, error) {
	sockPath := dockerSocketFinder()
	if sockPath != "" {
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
		if bcRes := pruneBuildCacheSocket(sockPath); bcRes != "" {
			details = append(details, bcRes)
		}

		if len(details) > 0 {
			return strings.Join(details, " | "), nil
		}
	}

	// CLI fallback
	ctx, cancel := context.WithTimeout(context.Background(), pruneTimeout)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "system", "prune", "-f", "--volumes")
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return formatPruneCLIOutput(string(out), "ресурсов"), nil
}

func formatPruneCLIOutput(out string, label string) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, l := range lines {
		if strings.Contains(l, totalReclaimedSpacePrefix) {
			space := strings.TrimPrefix(l, totalReclaimedSpacePrefix)
			return fmt.Sprintf("Очистка %s завершена. Освобождено:%s", label, space)
		}
	}
	return fmt.Sprintf("Очистка %s успешно выполнена", label)
}
