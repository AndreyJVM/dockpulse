package monitor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// containerAction performs a container lifecycle action through the socket with a CLI fallback.
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

// StartContainer starts a container by ID.
func StartContainer(containerID string) error {
	return containerAction(containerID, "start", http.StatusNoContent, http.StatusNotModified)
}

// StopContainer stops a container by ID.
func StopContainer(containerID string) error {
	return containerAction(containerID, "stop", http.StatusNoContent, http.StatusNotModified)
}

// RestartContainer restarts a container by ID.
func RestartContainer(containerID string) error {
	return containerAction(containerID, "restart", http.StatusNoContent)
}

// KillContainer terminates a container by ID with SIGKILL.
func KillContainer(containerID string) error {
	return containerAction(containerID, "kill", http.StatusNoContent)
}

// RemoveContainer force-removes a container by ID.
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
