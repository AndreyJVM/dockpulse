package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	socketDialTimeout   = 500 * time.Millisecond
	socketClientTimeout = 5 * time.Second
)

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
