package monitor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
