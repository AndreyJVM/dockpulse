package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	swarmTimeout        = 5 * time.Second
	swarmLogsTimeout    = 10 * time.Second
	defaultSwarmTail    = 100
	swarmIDTruncateLen  = 12
	swarmStatusActive   = "Активен"
	swarmStatusInactive = "Не активен"
	swarmTextInactive   = "Docker Swarm не активирован"
)

// SwarmNode представляет узел кластера Swarm.
type SwarmNode struct {
	ID           string
	Hostname     string
	Role         string
	State        string
	Availability string
	IsLeader     bool
	ActiveTasks  int
}

// SwarmService представляет службу кластера Swarm.
type SwarmService struct {
	ID              string
	Name            string
	Image           string
	Mode            string
	Replicas        string
	RunningReplicas int
	TargetReplicas  int
	Ports           string
}

// SwarmSummary содержит статус кластера Swarm.
type SwarmSummary struct {
	IsActive      bool
	StatusDisplay string
	NodeID        string
	IsManager     bool
	NodesCount    int
	ManagersCount int
	Nodes         []SwarmNode
	Services      []SwarmService
	Text          string
}

func truncateSwarmID(id string) string {
	if len(id) > swarmIDTruncateLen {
		return id[:swarmIDTruncateLen]
	}
	return id
}

func lookupTaskCount(taskMap map[string]int, fullID, shortID string) int {
	if taskMap == nil {
		return 0
	}
	if count, ok := taskMap[fullID]; ok {
		return count
	}
	return taskMap[shortID]
}

func cleanSwarmImage(rawImage string) string {
	return strings.Split(rawImage, "@")[0]
}

func fetchSocketAPI(client *http.Client, path string, timeout time.Duration) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost"+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, path)
	}

	return io.ReadAll(resp.Body)
}

// parseSwarmInfo разбирает JSON ответа Docker Engine API /info для секции Swarm.
func parseSwarmInfo(body []byte) (SwarmSummary, bool) {
	var infoData struct {
		Swarm struct {
			LocalNodeState   string `json:"LocalNodeState"`
			ControlAvailable bool   `json:"ControlAvailable"`
			NodeID           string `json:"NodeID"`
			Nodes            int    `json:"Nodes"`
			Managers         int    `json:"Managers"`
		} `json:"Swarm"`
	}

	if json.Unmarshal(body, &infoData) != nil || strings.ToLower(infoData.Swarm.LocalNodeState) != "active" {
		return SwarmSummary{}, false
	}

	return SwarmSummary{
		IsActive:      true,
		StatusDisplay: swarmStatusActive,
		NodeID:        truncateSwarmID(infoData.Swarm.NodeID),
		IsManager:     infoData.Swarm.ControlAvailable,
		NodesCount:    infoData.Swarm.Nodes,
		ManagersCount: infoData.Swarm.Managers,
	}, true
}

// parseSwarmTasks разбирает JSON ответа Docker Engine API /tasks и возвращает число работающих тасок по serviceID и nodeID.
func parseSwarmTasks(body []byte) (map[string]int, map[string]int) {
	var rawTasks []struct {
		ID        string `json:"ID"`
		ServiceID string `json:"ServiceID"`
		NodeID    string `json:"NodeID"`
		Status    struct {
			State string `json:"State"`
		} `json:"Status"`
		DesiredState string `json:"DesiredState"`
	}

	svcTasks := make(map[string]int)
	nodeTasks := make(map[string]int)

	if json.Unmarshal(body, &rawTasks) != nil {
		return svcTasks, nodeTasks
	}

	for _, t := range rawTasks {
		if strings.ToLower(t.Status.State) == "running" {
			svcTasks[t.ServiceID]++
			nodeTasks[t.NodeID]++
		}
	}
	return svcTasks, nodeTasks
}

// parseSwarmNodes разбирает JSON ответа Docker Engine API /nodes.
func parseSwarmNodes(body []byte, nodeTasks map[string]int) []SwarmNode {
	var rawNodes []struct {
		ID   string `json:"ID"`
		Spec struct {
			Role         string `json:"Role"`
			Availability string `json:"Availability"`
		} `json:"Spec"`
		Description struct {
			Hostname string `json:"Hostname"`
		} `json:"Description"`
		Status struct {
			State string `json:"State"`
		} `json:"Status"`
		ManagerStatus struct {
			Leader bool `json:"Leader"`
		} `json:"ManagerStatus"`
	}

	if json.Unmarshal(body, &rawNodes) != nil {
		return nil
	}

	nodes := make([]SwarmNode, 0, len(rawNodes))
	for _, n := range rawNodes {
		nid := truncateSwarmID(n.ID)
		availability := n.Spec.Availability
		if availability == "" {
			availability = "active"
		}

		nodes = append(nodes, SwarmNode{
			ID:           nid,
			Hostname:     n.Description.Hostname,
			Role:         n.Spec.Role,
			State:        n.Status.State,
			Availability: availability,
			IsLeader:     n.ManagerStatus.Leader,
			ActiveTasks:  lookupTaskCount(nodeTasks, n.ID, nid),
		})
	}
	return nodes
}

type swarmServiceRaw struct {
	ID   string `json:"ID"`
	Spec struct {
		Name         string `json:"Name"`
		TaskTemplate struct {
			ContainerSpec struct {
				Image string `json:"Image"`
			} `json:"ContainerSpec"`
		} `json:"TaskTemplate"`
		Mode struct {
			Replicated *struct {
				Replicas int `json:"Replicas"`
			} `json:"Replicated"`
			Global *struct{} `json:"Global"`
		} `json:"Mode"`
	} `json:"Spec"`
	Endpoint struct {
		Ports []struct {
			Protocol      string `json:"Protocol"`
			TargetPort    int    `json:"TargetPort"`
			PublishedPort int    `json:"PublishedPort"`
		} `json:"Ports"`
	} `json:"Endpoint"`
}

func formatSwarmServicePorts(ports []struct {
	Protocol      string `json:"Protocol"`
	TargetPort    int    `json:"TargetPort"`
	PublishedPort int    `json:"PublishedPort"`
}) string {
	if len(ports) == 0 {
		return ""
	}
	portList := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.PublishedPort > 0 {
			portList = append(portList, fmt.Sprintf("%d:%d/%s", p.PublishedPort, p.TargetPort, p.Protocol))
		} else if p.TargetPort > 0 {
			portList = append(portList, fmt.Sprintf("%d/%s", p.TargetPort, p.Protocol))
		}
	}
	return strings.Join(portList, ", ")
}

// parseSwarmServices разбирает JSON ответа Docker Engine API /services.
func parseSwarmServices(body []byte, runningTasks map[string]int) []SwarmService {
	var rawSvcs []swarmServiceRaw
	if json.Unmarshal(body, &rawSvcs) != nil {
		return nil
	}

	svcs := make([]SwarmService, 0, len(rawSvcs))
	for _, s := range rawSvcs {
		sid := truncateSwarmID(s.ID)
		running := lookupTaskCount(runningTasks, s.ID, sid)

		mode := "replicated"
		targetReplicas := 0
		var rep string

		if s.Spec.Mode.Global != nil {
			mode = "global"
			targetReplicas = running
			rep = fmt.Sprintf("%d (global)", running)
		} else if s.Spec.Mode.Replicated != nil {
			targetReplicas = s.Spec.Mode.Replicated.Replicas
			rep = fmt.Sprintf("%d/%d", running, targetReplicas)
		} else {
			rep = fmt.Sprintf("%d", running)
		}

		svcs = append(svcs, SwarmService{
			ID:              sid,
			Name:            s.Spec.Name,
			Image:           cleanSwarmImage(s.Spec.TaskTemplate.ContainerSpec.Image),
			Mode:            mode,
			Replicas:        rep,
			RunningReplicas: running,
			TargetReplicas:  targetReplicas,
			Ports:           formatSwarmServicePorts(s.Endpoint.Ports),
		})
	}
	return svcs
}

// parseSwarmCLIInfo разбирает вывод команды docker info для Swarm.
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

// parseSwarmCLINodes разбирает вывод команды docker node ls.
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

// parseSwarmCLIServices разбирает вывод команды docker service ls.
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

// GetSwarmSummary собирает информацию о кластере Swarm.
func GetSwarmSummary() SwarmSummary {
	sockPath := dockerSocketFinder()
	summary := SwarmSummary{
		IsActive:      false,
		StatusDisplay: swarmStatusInactive,
		Text:          swarmTextInactive,
	}

	if sockPath != "" {
		client := socketClientFactory(sockPath)
		body, err := fetchSocketAPI(client, "/info", swarmTimeout)
		if err == nil {
			if parsed, ok := parseSwarmInfo(body); ok {
				summary = parsed

				// Параллельный опрос /tasks, /nodes и /services для минимизации задержки
				var (
					tasksBody    []byte
					nodesBody    []byte
					servicesBody []byte
					wg           sync.WaitGroup
				)
				wg.Add(3)
				go func() {
					defer wg.Done()
					tasksBody, _ = fetchSocketAPI(client, "/tasks", swarmTimeout)
				}()
				go func() {
					defer wg.Done()
					nodesBody, _ = fetchSocketAPI(client, "/nodes", swarmTimeout)
				}()
				go func() {
					defer wg.Done()
					servicesBody, _ = fetchSocketAPI(client, "/services", swarmTimeout)
				}()
				wg.Wait()

				svcTasks, nodeTasks := parseSwarmTasks(tasksBody)
				summary.Nodes = parseSwarmNodes(nodesBody, nodeTasks)
				summary.Services = parseSwarmServices(servicesBody, svcTasks)
				summary.Text = fmt.Sprintf("Кластер Swarm: %d нод (%d менеджеров), %d служб", summary.NodesCount, summary.ManagersCount, len(summary.Services))
				return summary
			}
		}
	}

	// Fallback CLI
	ctx, cancel := context.WithTimeout(context.Background(), swarmTimeout)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "info", "--format", "{{.Swarm.LocalNodeState}} {{.Swarm.ControlAvailable}} {{.Swarm.Nodes}} {{.Swarm.Managers}}")
	if err == nil {
		active, isManager, nodes, managers := parseSwarmCLIInfo(string(out))
		if active {
			summary.IsActive = true
			summary.StatusDisplay = swarmStatusActive
			summary.IsManager = isManager
			summary.NodesCount = nodes
			summary.ManagersCount = managers

			if isManager {
				if nodeOut, nErr := dockerCLICommandRunner(ctx, "node", "ls", "--format", "{{.ID}}\t{{.Hostname}}\t{{.Role}}\t{{.Status}}\t{{.Availability}}\t{{.ManagerStatus}}"); nErr == nil {
					summary.Nodes = parseSwarmCLINodes(string(nodeOut))
				}
				if svcOut, sErr := dockerCLICommandRunner(ctx, "service", "ls", "--format", "{{.ID}}\t{{.Name}}\t{{.Mode}}\t{{.Replicas}}\t{{.Image}}\t{{.Ports}}"); sErr == nil {
					summary.Services = parseSwarmCLIServices(string(svcOut))
				}
			}

			if len(summary.Services) > 0 {
				summary.Text = fmt.Sprintf("Кластер Swarm: %d нод (%d менеджеров), %d служб", summary.NodesCount, summary.ManagersCount, len(summary.Services))
			} else {
				summary.Text = fmt.Sprintf("Кластер Swarm: %d нод (%d менеджеров)", nodes, managers)
			}
			return summary
		}
	}

	return summary
}

func fetchSwarmTasks(client *http.Client) (map[string]int, map[string]int) {
	body, err := fetchSocketAPI(client, "/tasks", swarmTimeout)
	if err != nil {
		return make(map[string]int), make(map[string]int)
	}
	return parseSwarmTasks(body)
}

func fetchSwarmNodes(client *http.Client, nodeTasks map[string]int) []SwarmNode {
	body, err := fetchSocketAPI(client, "/nodes", swarmTimeout)
	if err != nil {
		return nil
	}
	return parseSwarmNodes(body, nodeTasks)
}

func fetchSwarmServices(client *http.Client, runningTasks map[string]int) []SwarmService {
	body, err := fetchSocketAPI(client, "/services", swarmTimeout)
	if err != nil {
		return nil
	}
	return parseSwarmServices(body, runningTasks)
}

// GetSwarmServiceLogs читает логи службы Swarm по её ID или имени.
func GetSwarmServiceLogs(serviceID string, tailLines int) (string, error) {
	cleanID := strings.TrimSpace(serviceID)
	if cleanID == "" {
		return "", fmt.Errorf("service ID cannot be empty")
	}
	if tailLines <= 0 {
		tailLines = defaultSwarmTail
	}

	sockPath := dockerSocketFinder()
	if sockPath != "" {
		client := socketClientFactory(sockPath)
		url := fmt.Sprintf("/services/%s/logs?stdout=1&stderr=1&tail=%d", cleanID, tailLines)
		ctx, cancel := context.WithTimeout(context.Background(), swarmLogsTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost"+url, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return demuxDockerLogs(resp.Body), nil
				}
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), swarmLogsTimeout)
	defer cancel()
	out, err := dockerCLICommandRunner(ctx, "service", "logs", "--tail", strconv.Itoa(tailLines), cleanID)
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
