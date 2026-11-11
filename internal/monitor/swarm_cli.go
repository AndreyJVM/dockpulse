package monitor

import (
	"strconv"
	"strings"
)

// parseSwarmCLIInfo parses the output of `docker info` for Docker Swarm section.
func parseSwarmCLIInfo(output string) (isActive bool, isManager bool, nodes int, managers int) {
	parts := strings.Fields(output)
	if len(parts) == 0 || strings.ToLower(parts[0]) != "active" {
		return false, false, 0, 0
	}
	isActive = true
	if len(parts) > 1 {
		isManager = strings.ToLower(parts[1]) == "true"
	}
	if len(parts) > 2 {
		nodes, _ = strconv.Atoi(parts[2])
	}
	if len(parts) > 3 {
		managers, _ = strconv.Atoi(parts[3])
	}
	return isActive, isManager, nodes, managers
}

// parseSwarmCLINodes parses the output of `docker node ls`.
func parseSwarmCLINodes(output string) []SwarmNode {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	nodes := make([]SwarmNode, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 5 {
			isLeader := false
			if len(parts) >= 6 && strings.Contains(strings.ToLower(parts[5]), "leader") {
				isLeader = true
			}
			nodes = append(nodes, SwarmNode{
				ID:           truncateSwarmID(parts[0]),
				Hostname:     parts[1],
				Role:         strings.ToLower(parts[2]),
				State:        strings.ToLower(parts[3]),
				Availability: strings.ToLower(parts[4]),
				IsLeader:     isLeader,
			})
		}
	}
	return nodes
}

// parseSwarmCLIServices parses the output of `docker service ls`.
func parseSwarmCLIServices(output string) []SwarmService {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	svcs := make([]SwarmService, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 5 {
			mode := strings.ToLower(parts[2])
			repStr := parts[3]
			running := 0
			target := 0

			repParts := strings.Split(repStr, "/")
			if len(repParts) == 2 {
				running, _ = strconv.Atoi(strings.TrimSpace(repParts[0]))
				tStr := strings.Fields(repParts[1])
				if len(tStr) > 0 {
					target, _ = strconv.Atoi(tStr[0])
				}
			}

			ports := ""
			if len(parts) >= 6 {
				ports = parts[5]
			}

			svcs = append(svcs, SwarmService{
				ID:              truncateSwarmID(parts[0]),
				Name:            parts[1],
				Mode:            mode,
				Replicas:        repStr,
				RunningReplicas: running,
				TargetReplicas:  target,
				Image:           cleanSwarmImage(parts[4]),
				Ports:           ports,
			})
		}
	}
	return svcs
}
