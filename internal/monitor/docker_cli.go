package monitor

import (
	"context"
	"strconv"
	"strings"
)

func parseCLIPsOutput(output string) []ContainerItem {
	var containers []ContainerItem
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 6 {
			continue
		}
		cID := parts[0]
		cName := parts[1]
		cImg := parts[2]
		cState := parts[3]
		cStatus := parts[4]
		cPorts := parts[5]
		var cMounts []string
		if len(parts) >= 7 && parts[6] != "" {
			for _, m := range strings.Split(parts[6], ",") {
				if tr := strings.TrimSpace(m); tr != "" {
					cMounts = append(cMounts, tr)
				}
			}
		}
		var pProj, pSvc string
		if len(parts) >= 8 {
			for _, lbl := range strings.Split(parts[7], ",") {
				kv := strings.SplitN(lbl, "=", 2)
				if len(kv) == 2 {
					if kv[0] == composeProjectLabel {
						pProj = kv[1]
					} else if kv[0] == composeServiceLabel {
						pSvc = kv[1]
					}
				}
			}
		}
		var cNet, cCreated, cStarted string
		if len(parts) >= 9 {
			cNet = parts[8]
		}
		if len(parts) >= 10 {
			cCreated = parts[9]
		}
		if len(parts) >= 11 {
			cStarted = parts[10]
		}
		containers = append(containers, ContainerItem{
			ID:             cID,
			Name:           cName,
			Image:          cImg,
			State:          cState,
			Status:         cStatus,
			Health:         ParseHealthStatus(cStatus),
			Ports:          cPorts,
			Mounts:         cMounts,
			ComposeProject: pProj,
			ComposeService: pSvc,
			NetworkName:    cNet,
			Created:        cCreated,
			StartedAt:      cStarted,
		})
	}
	return containers
}

func enrichContainersWithCLIInspect(containers []ContainerItem) {
	if len(containers) == 0 {
		return
	}
	inspectCtx, inspectCancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
	defer inspectCancel()
	inspectArgs := []string{"inspect", "--format", "{{slice .Id 0 12}}|{{.HostConfig.RestartPolicy.Name}}|{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}"}
	for _, c := range containers {
		inspectArgs = append(inspectArgs, c.ID)
	}
	insOut, insErr := dockerCLICommandRunner(inspectCtx, inspectArgs...)
	if insErr != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(string(insOut)), "\n") {
		f := strings.Split(line, "|")
		if len(f) >= 3 {
			cID := strings.TrimSpace(f[0])
			rPolicy := strings.TrimSpace(f[1])
			ips := strings.Fields(f[2])
			for idx := range containers {
				if containers[idx].ID == cID {
					containers[idx].RestartPolicy = rPolicy
					if len(ips) > 0 && containers[idx].IPAddress == "" {
						containers[idx].IPAddress = strings.Join(ips, ", ")
					}
					break
				}
			}
		}
	}
}

func fetchDockerSummaryViaCLI(summary *DockerSummary) {
	if summary.Status != DockerStatusRunning {
		ctx, cancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
		defer cancel()
		out, err := dockerCLICommandRunner(ctx, "info", "--format", "{{.Containers}} {{.ContainersRunning}} {{.ContainersStopped}} {{.Images}} {{.ServerVersion}}")
		if err == nil {
			parts := strings.Fields(string(out))
			if len(parts) >= 5 {
				summary.Status = DockerStatusRunning
				summary.StatusDisplay = DockerStatusDisplayRunning
				if val, convErr := strconv.Atoi(parts[0]); convErr == nil {
					summary.TotalContainers = val
				}
				if val, convErr := strconv.Atoi(parts[1]); convErr == nil {
					summary.RunningContainers = val
				}
				if val, convErr := strconv.Atoi(parts[2]); convErr == nil {
					summary.StoppedContainers = val
				}
				if val, convErr := strconv.Atoi(parts[3]); convErr == nil {
					summary.ImagesCount = val
				}
				summary.Version = parts[4]
			}
		}
	}

	if summary.Status == DockerStatusRunning && len(summary.Containers) == 0 {
		psCtx, psCancel := context.WithTimeout(context.Background(), defaultDockerTimeout)
		defer psCancel()
		psOut, psErr := dockerCLICommandRunner(psCtx, "ps", "-a", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}\t{{.Mounts}}\t{{.Labels}}\t{{.Networks}}\t{{.CreatedAt}}\t{{.RunningFor}}")
		if psErr == nil {
			summary.Containers = parseCLIPsOutput(string(psOut))
		}
		if len(summary.Containers) > 0 {
			enrichContainersWithCLIInspect(summary.Containers)
		}
	}
}
