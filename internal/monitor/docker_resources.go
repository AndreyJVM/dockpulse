package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// DockerNetwork describes a Docker network shown in the dashboard.
type DockerNetwork struct {
	ID             string
	Name           string
	Driver         string
	Scope          string
	Internal       bool
	Attachable     bool
	Containers     int
	Subnet         string
	Gateway        string
	ComposeProject string
}

// DockerVolume describes a Docker volume shown in the dashboard.
type DockerVolume struct {
	Name           string
	Driver         string
	Scope          string
	Mountpoint     string
	Size           int64
	RefCount       int
	ComposeProject string
}

// GetDockerNetworks returns Docker networks without changing daemon state.
func GetDockerNetworks() ([]DockerNetwork, error) {
	if sockPath := dockerSocketFinder(); sockPath != "" {
		var raw []struct {
			ID         string                 `json:"Id"`
			Name       string                 `json:"Name"`
			Driver     string                 `json:"Driver"`
			Scope      string                 `json:"Scope"`
			Internal   bool                   `json:"Internal"`
			Attachable bool                   `json:"Attachable"`
			Containers map[string]interface{} `json:"Containers"`
			Labels     map[string]string      `json:"Labels"`
			IPAM       struct {
				Config []struct {
					Subnet  string `json:"Subnet"`
					Gateway string `json:"Gateway"`
				} `json:"Config"`
			} `json:"IPAM"`
		}
		if err := fetchDockerResourceJSON(sockPath, "/networks", &raw); err == nil {
			networks := make([]DockerNetwork, 0, len(raw))
			for _, item := range raw {
				subnet, gateway := "", ""
				if len(item.IPAM.Config) > 0 {
					subnet = item.IPAM.Config[0].Subnet
					gateway = item.IPAM.Config[0].Gateway
				}
				networks = append(networks, DockerNetwork{
					ID:             truncateResourceID(item.ID),
					Name:           item.Name,
					Driver:         item.Driver,
					Scope:          item.Scope,
					Internal:       item.Internal,
					Attachable:     item.Attachable,
					Containers:     len(item.Containers),
					Subnet:         subnet,
					Gateway:        gateway,
					ComposeProject: item.Labels[composeProjectLabel],
				})
			}
			return networks, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
	defer cancel()
	format := "{{.ID}}|{{.Name}}|{{.Driver}}|{{.Scope}}|{{.Internal}}|{{.Labels}}"
	out, err := dockerCLICommandRunner(ctx, "network", "ls", "--no-trunc", "--format", format)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker networks: %w", err)
	}

	var networks []DockerNetwork
	for _, line := range nonEmptyLines(string(out)) {
		parts := strings.SplitN(line, "|", 6)
		if len(parts) != 6 {
			continue
		}
		networks = append(networks, DockerNetwork{
			ID:             truncateResourceID(parts[0]),
			Name:           parts[1],
			Driver:         parts[2],
			Scope:          parts[3],
			Internal:       parseBool(parts[4]),
			ComposeProject: labelValue(parts[5], composeProjectLabel),
		})
	}
	return networks, nil
}

// GetDockerVolumes returns Docker volumes without changing daemon state.
func GetDockerVolumes() ([]DockerVolume, error) {
	if sockPath := dockerSocketFinder(); sockPath != "" {
		var raw struct {
			Volumes []struct {
				Name       string            `json:"Name"`
				Driver     string            `json:"Driver"`
				Scope      string            `json:"Scope"`
				Mountpoint string            `json:"Mountpoint"`
				Labels     map[string]string `json:"Labels"`
				UsageData  *struct {
					Size     int64 `json:"Size"`
					RefCount int   `json:"RefCount"`
				} `json:"UsageData"`
			} `json:"Volumes"`
		}
		if err := fetchDockerResourceJSON(sockPath, "/volumes", &raw); err == nil {
			volumes := make([]DockerVolume, 0, len(raw.Volumes))
			for _, item := range raw.Volumes {
				size, refCount := int64(-1), -1
				if item.UsageData != nil {
					size = item.UsageData.Size
					refCount = item.UsageData.RefCount
				}
				volumes = append(volumes, DockerVolume{
					Name:           item.Name,
					Driver:         item.Driver,
					Scope:          item.Scope,
					Mountpoint:     item.Mountpoint,
					Size:           size,
					RefCount:       refCount,
					ComposeProject: item.Labels[composeProjectLabel],
				})
			}
			return volumes, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
	defer cancel()
	format := "{{.Name}}|{{.Driver}}|{{.Scope}}|{{.Mountpoint}}|{{.Labels}}"
	out, err := dockerCLICommandRunner(ctx, "volume", "ls", "--format", format)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker volumes: %w", err)
	}

	var volumes []DockerVolume
	for _, line := range nonEmptyLines(string(out)) {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) != 5 {
			continue
		}
		volumes = append(volumes, DockerVolume{
			Name:           parts[0],
			Driver:         parts[1],
			Scope:          parts[2],
			Mountpoint:     parts[3],
			Size:           -1,
			RefCount:       -1,
			ComposeProject: labelValue(parts[4], composeProjectLabel),
		})
	}
	return volumes, nil
}

func fetchDockerResourceJSON(sockPath, path string, target interface{}) error {
	client := socketClientFactory(sockPath)
	ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost"+path, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Docker API %s returned status %d", path, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func truncateResourceID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func parseBool(value string) bool {
	parsed, _ := strconv.ParseBool(strings.TrimSpace(value))
	return parsed
}

func nonEmptyLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(value), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func labelValue(labels, key string) string {
	for _, label := range strings.Split(labels, ",") {
		name, value, found := strings.Cut(strings.TrimSpace(label), "=")
		if found && name == key {
			return value
		}
	}
	return ""
}
