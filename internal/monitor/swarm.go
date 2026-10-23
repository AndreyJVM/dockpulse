package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// SwarmNode представляет узел кластера Swarm.
type SwarmNode struct {
	ID       string
	Hostname string
	Role     string
	State    string
	IsLeader bool
}

// SwarmService представляет службу кластера Swarm.
type SwarmService struct {
	ID       string
	Name     string
	Image    string
	Replicas string
}

// SwarmSummary содержит статус кластера Swarm.
type SwarmSummary struct {
	IsActive       bool
	StatusDisplay  string
	NodeID         string
	IsManager      bool
	NodesCount     int
	ManagersCount  int
	Nodes          []SwarmNode
	Services       []SwarmService
	Text           string
}

// GetSwarmSummary собирает информацию о кластере Swarm.
func GetSwarmSummary() SwarmSummary {
	sockPath := FindDockerSocket()
	summary := SwarmSummary{
		IsActive:      false,
		StatusDisplay: "⚪ Не активен",
		Text:          "Docker Swarm не активирован",
	}

	if sockPath != "" {
		client := getSocketHTTPClient(sockPath)
		resp, err := client.Get("http://localhost/info")
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var infoData struct {
				Swarm struct {
					LocalNodeState   string `json:"LocalNodeState"`
					ControlAvailable bool   `json:"ControlAvailable"`
					NodeID           string `json:"NodeID"`
					Nodes            int    `json:"Nodes"`
					Managers         int    `json:"Managers"`
				} `json:"Swarm"`
			}

			if json.Unmarshal(body, &infoData) == nil && strings.ToLower(infoData.Swarm.LocalNodeState) == "active" {
				summary.IsActive = true
				summary.StatusDisplay = "🟢 Активен"
				summary.NodeID = infoData.Swarm.NodeID
				if len(summary.NodeID) > 12 {
					summary.NodeID = summary.NodeID[:12]
				}
				summary.IsManager = infoData.Swarm.ControlAvailable
				summary.NodesCount = infoData.Swarm.Nodes
				summary.ManagersCount = infoData.Swarm.Managers
				summary.Nodes = fetchSwarmNodes(client)
				summary.Services = fetchSwarmServices(client)
				summary.Text = fmt.Sprintf("Кластер Swarm: %d нод (%d менеджеров), %d служб", summary.NodesCount, summary.ManagersCount, len(summary.Services))
				return summary
			}
		}
	}

	// Fallback CLI
	if _, err := exec.LookPath("docker"); err == nil {
		out, err := exec.Command("docker", "info", "--format", "{{.Swarm.LocalNodeState}} {{.Swarm.ControlAvailable}} {{.Swarm.Nodes}} {{.Swarm.Managers}}").Output()
		if err == nil {
			parts := strings.Fields(string(out))
			if len(parts) > 0 && strings.ToLower(parts[0]) == "active" {
				summary.IsActive = true
				summary.StatusDisplay = "🟢 Активен"
				summary.Text = "Docker Swarm активен на этом хосте"
				return summary
			}
		}
	}

	return summary
}

func fetchSwarmNodes(client *http.Client) []SwarmNode {
	resp, err := client.Get("http://localhost/nodes")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rawNodes []struct {
		ID   string `json:"ID"`
		Spec struct {
			Role string `json:"Role"`
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
		nid := n.ID
		if len(nid) > 12 {
			nid = nid[:12]
		}
		nodes = append(nodes, SwarmNode{
			ID:       nid,
			Hostname: n.Description.Hostname,
			Role:     n.Spec.Role,
			State:    n.Status.State,
			IsLeader: n.ManagerStatus.Leader,
		})
	}
	return nodes
}

func fetchSwarmServices(client *http.Client) []SwarmService {
	resp, err := client.Get("http://localhost/services")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
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
				Replicated struct {
					Replicas int `json:"Replicas"`
				} `json:"Replicated"`
			} `json:"Mode"`
		} `json:"Spec"`
	}

	if json.Unmarshal(body, &rawSvcs) != nil {
		return nil
	}

	var svcs []SwarmService
	for _, s := range rawSvcs {
		sid := s.ID
		if len(sid) > 12 {
			sid = sid[:12]
		}
		img := strings.Split(s.Spec.TaskTemplate.ContainerSpec.Image, "@")[0]
		rep := fmt.Sprintf("%d replicas", s.Spec.Mode.Replicated.Replicas)

		svcs = append(svcs, SwarmService{
			ID:       sid,
			Name:     s.Spec.Name,
			Image:    img,
			Replicas: rep,
		})
	}
	return svcs
}
