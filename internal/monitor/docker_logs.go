package monitor

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
