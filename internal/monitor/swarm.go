package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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

	nodeID := infoData.Swarm.NodeID
	if len(nodeID) > 12 {
		nodeID = nodeID[:12]
	}

	return SwarmSummary{
		IsActive:      true,
		StatusDisplay: "Активен",
		NodeID:        nodeID,
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

	var nodes []SwarmNode
	for _, n := range rawNodes {
		fullID := n.ID
		nid := fullID
		if len(nid) > 12 {
			nid = nid[:12]
		}
		availability := n.Spec.Availability
		if availability == "" {
			availability = "active"
		}

		activeTasks := 0
		if nodeTasks != nil {
			activeTasks = nodeTasks[fullID]
			if activeTasks == 0 && nodeTasks[nid] > 0 {
				activeTasks = nodeTasks[nid]
			}
		}

		nodes = append(nodes, SwarmNode{
			ID:           nid,
			Hostname:     n.Description.Hostname,
			Role:         n.Spec.Role,
			State:        n.Status.State,
			Availability: availability,
			IsLeader:     n.ManagerStatus.Leader,
			ActiveTasks:  activeTasks,
		})
	}
	return nodes
}

// parseSwarmServices разбирает JSON ответа Docker Engine API /services.
func parseSwarmServices(body []byte, runningTasks map[string]int) []SwarmService {
	var rawSvcs []struct {
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

	if json.Unmarshal(body, &rawSvcs) != nil {
		return nil
	}

	var svcs []SwarmService
	for _, s := range rawSvcs {
		fullID := s.ID
		sid := fullID
		if len(sid) > 12 {
			sid = sid[:12]
		}
		img := strings.Split(s.Spec.TaskTemplate.ContainerSpec.Image, "@")[0]

		mode := "replicated"
		targetReplicas := 0
		running := 0
		if runningTasks != nil {
			running = runningTasks[fullID]
			if running == 0 && runningTasks[sid] > 0 {
				running = runningTasks[sid]
			}
		}

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

		var portList []string
		for _, p := range s.Endpoint.Ports {
			if p.PublishedPort > 0 {
				portList = append(portList, fmt.Sprintf("%d:%d/%s", p.PublishedPort, p.TargetPort, p.Protocol))
			} else if p.TargetPort > 0 {
				portList = append(portList, fmt.Sprintf("%d/%s", p.TargetPort, p.Protocol))
			}
		}

		svcs = append(svcs, SwarmService{
			ID:              sid,
			Name:            s.Spec.Name,
			Image:           img,
			Mode:            mode,
			Replicas:        rep,
			RunningReplicas: running,
			TargetReplicas:  targetReplicas,
			Ports:           strings.Join(portList, ", "),
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

// GetSwarmSummary собирает информацию о кластере Swarm.
func GetSwarmSummary() SwarmSummary {
	sockPath := FindDockerSocket()
	summary := SwarmSummary{
		IsActive:      false,
		StatusDisplay: "Не активен",
		Text:          "Docker Swarm не активирован",
	}

	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		resp, err := client.Get("http://localhost/info")
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			if parsed, ok := parseSwarmInfo(body); ok {
				summary = parsed
				svcTasks, nodeTasks := fetchSwarmTasks(client)
				summary.Nodes = fetchSwarmNodes(client, nodeTasks)
				summary.Services = fetchSwarmServices(client, svcTasks)
				summary.Text = fmt.Sprintf("Кластер Swarm: %d нод (%d менеджеров), %d служб", summary.NodesCount, summary.ManagersCount, len(summary.Services))
				return summary
			}
		}
	}

	// Fallback CLI
	if _, err := exec.LookPath("docker"); err == nil {
		out, err := exec.Command("docker", "info", "--format", "{{.Swarm.LocalNodeState}} {{.Swarm.ControlAvailable}} {{.Swarm.Nodes}} {{.Swarm.Managers}}").Output()
		if err == nil {
			active, isManager, nodes, managers := parseSwarmCLIInfo(string(out))
			if active {
				summary.IsActive = true
				summary.StatusDisplay = "Активен"
				summary.IsManager = isManager
				summary.NodesCount = nodes
				summary.ManagersCount = managers
				summary.Text = fmt.Sprintf("Кластер Swarm: %d нод (%d менеджеров)", nodes, managers)
				return summary
			}
		}
	}

	return summary
}

func fetchSwarmTasks(client *http.Client) (map[string]int, map[string]int) {
	resp, err := client.Get("http://localhost/tasks")
	if err != nil || resp.StatusCode != 200 {
		return make(map[string]int), make(map[string]int)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return parseSwarmTasks(body)
}

func fetchSwarmNodes(client *http.Client, nodeTasks map[string]int) []SwarmNode {
	resp, err := client.Get("http://localhost/nodes")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return parseSwarmNodes(body, nodeTasks)
}

func fetchSwarmServices(client *http.Client, runningTasks map[string]int) []SwarmService {
	resp, err := client.Get("http://localhost/services")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return parseSwarmServices(body, runningTasks)
}
