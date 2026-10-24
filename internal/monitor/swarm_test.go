package monitor

import (
	"testing"
)

func TestGetSwarmSummary(t *testing.T) {
	summary := GetSwarmSummary()
	// Should safely return summary without panic
	if summary.Text == "" {
		t.Errorf("expected non-empty Swarm summary text")
	}
}

func TestParseSwarmInfo(t *testing.T) {
	activeJSON := []byte(`{
		"Swarm": {
			"LocalNodeState": "active",
			"ControlAvailable": true,
			"NodeID": "2468135790abcdef123456",
			"Nodes": 5,
			"Managers": 3
		}
	}`)

	summary, ok := parseSwarmInfo(activeJSON)
	if !ok {
		t.Fatalf("expected parseSwarmInfo to succeed for active JSON")
	}
	if !summary.IsActive {
		t.Errorf("expected IsActive=true")
	}
	if summary.StatusDisplay != "Активен" {
		t.Errorf("expected StatusDisplay 'Активен', got '%s'", summary.StatusDisplay)
	}
	if summary.NodeID != "2468135790ab" {
		t.Errorf("expected truncated NodeID '2468135790ab', got '%s'", summary.NodeID)
	}
	if !summary.IsManager {
		t.Errorf("expected IsManager=true")
	}
	if summary.NodesCount != 5 || summary.ManagersCount != 3 {
		t.Errorf("expected 5 nodes and 3 managers, got %d / %d", summary.NodesCount, summary.ManagersCount)
	}

	// Inactive
	inactiveJSON := []byte(`{
		"Swarm": {
			"LocalNodeState": "inactive"
		}
	}`)
	_, okInactive := parseSwarmInfo(inactiveJSON)
	if okInactive {
		t.Errorf("expected parseSwarmInfo to return false for inactive swarm")
	}

	// Invalid JSON
	_, okInvalid := parseSwarmInfo([]byte(`{invalid`))
	if okInvalid {
		t.Errorf("expected parseSwarmInfo to return false for malformed JSON")
	}
}

func TestParseSwarmTasks(t *testing.T) {
	tasksJSON := []byte(`[
		{
			"ID": "t1",
			"ServiceID": "svc1",
			"NodeID": "node1",
			"Status": { "State": "running" },
			"DesiredState": "running"
		},
		{
			"ID": "t2",
			"ServiceID": "svc1",
			"NodeID": "node2",
			"Status": { "State": "running" },
			"DesiredState": "running"
		},
		{
			"ID": "t3",
			"ServiceID": "svc1",
			"NodeID": "node1",
			"Status": { "State": "failed" },
			"DesiredState": "shutdown"
		},
		{
			"ID": "t4",
			"ServiceID": "svc2",
			"NodeID": "node1",
			"Status": { "State": "running" },
			"DesiredState": "running"
		}
	]`)

	svcTasks, nodeTasks := parseSwarmTasks(tasksJSON)
	if svcTasks["svc1"] != 2 {
		t.Errorf("expected 2 running tasks for svc1, got %d", svcTasks["svc1"])
	}
	if svcTasks["svc2"] != 1 {
		t.Errorf("expected 1 running task for svc2, got %d", svcTasks["svc2"])
	}
	if nodeTasks["node1"] != 2 {
		t.Errorf("expected 2 running tasks on node1, got %d", nodeTasks["node1"])
	}
	if nodeTasks["node2"] != 1 {
		t.Errorf("expected 1 running task on node2, got %d", nodeTasks["node2"])
	}

	// Malformed JSON
	badSvc, badNode := parseSwarmTasks([]byte(`malformed`))
	if len(badSvc) != 0 || len(badNode) != 0 {
		t.Errorf("expected empty maps for malformed JSON")
	}
}

func TestParseSwarmNodes(t *testing.T) {
	nodesJSON := []byte(`[
		{
			"ID": "node1234567890abcdef",
			"Spec": {
				"Role": "manager",
				"Availability": "active"
			},
			"Description": {
				"Hostname": "swarm-mgr-01"
			},
			"Status": {
				"State": "ready"
			},
			"ManagerStatus": {
				"Leader": true
			}
		},
		{
			"ID": "worker9876543210fedcba",
			"Spec": {
				"Role": "worker",
				"Availability": "drain"
			},
			"Description": {
				"Hostname": "swarm-wrk-01"
			},
			"Status": {
				"State": "ready"
			},
			"ManagerStatus": {
				"Leader": false
			}
		}
	]`)

	nodeTasks := map[string]int{
		"node1234567890abcdef": 4,
		"worker9876543210fedcba": 1,
	}

	nodes := parseSwarmNodes(nodesJSON, nodeTasks)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	n1 := nodes[0]
	if n1.ID != "node12345678" {
		t.Errorf("expected truncated ID 'node12345678', got '%s'", n1.ID)
	}
	if n1.Hostname != "swarm-mgr-01" {
		t.Errorf("expected Hostname 'swarm-mgr-01', got '%s'", n1.Hostname)
	}
	if n1.Role != "manager" || !n1.IsLeader {
		t.Errorf("expected manager leader node, got role=%s leader=%v", n1.Role, n1.IsLeader)
	}
	if n1.Availability != "active" {
		t.Errorf("expected Availability 'active', got '%s'", n1.Availability)
	}
	if n1.ActiveTasks != 4 {
		t.Errorf("expected 4 ActiveTasks, got %d", n1.ActiveTasks)
	}

	n2 := nodes[1]
	if n2.Role != "worker" || n2.IsLeader {
		t.Errorf("expected worker non-leader node, got role=%s leader=%v", n2.Role, n2.IsLeader)
	}
	if n2.Availability != "drain" {
		t.Errorf("expected Availability 'drain', got '%s'", n2.Availability)
	}
	if n2.ActiveTasks != 1 {
		t.Errorf("expected 1 ActiveTask, got %d", n2.ActiveTasks)
	}

	// Malformed JSON
	if parseSwarmNodes([]byte(`malformed`), nil) != nil {
		t.Errorf("expected nil for malformed nodes JSON")
	}
}

func TestParseSwarmServices(t *testing.T) {
	svcsJSON := []byte(`[
		{
			"ID": "svc1234567890abcdef",
			"Spec": {
				"Name": "web-frontend",
				"TaskTemplate": {
					"ContainerSpec": {
						"Image": "nginx:alpine@sha256:73984729384729384"
					}
				},
				"Mode": {
					"Replicated": {
						"Replicas": 3
					}
				}
			},
			"Endpoint": {
				"Ports": [
					{
						"Protocol": "tcp",
						"TargetPort": 80,
						"PublishedPort": 8080
					}
				]
			}
		},
		{
			"ID": "svcGlobal9876543210fe",
			"Spec": {
				"Name": "log-collector",
				"TaskTemplate": {
					"ContainerSpec": {
						"Image": "fluentd:v1"
					}
				},
				"Mode": {
					"Global": {}
				}
			}
		}
	]`)

	svcTasks := map[string]int{
		"svc1234567890abcdef": 3,
		"svcGlobal9876543210fe": 2,
	}

	svcs := parseSwarmServices(svcsJSON, svcTasks)
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}

	s1 := svcs[0]
	if s1.Name != "web-frontend" {
		t.Errorf("expected Name 'web-frontend', got '%s'", s1.Name)
	}
	if s1.Image != "nginx:alpine" {
		t.Errorf("expected stripped image 'nginx:alpine', got '%s'", s1.Image)
	}
	if s1.Replicas != "3/3" {
		t.Errorf("expected '3/3', got '%s'", s1.Replicas)
	}
	if s1.RunningReplicas != 3 || s1.TargetReplicas != 3 {
		t.Errorf("expected 3/3 replicas, got running=%d target=%d", s1.RunningReplicas, s1.TargetReplicas)
	}
	if s1.Ports != "8080:80/tcp" {
		t.Errorf("expected Ports '8080:80/tcp', got '%s'", s1.Ports)
	}

	s2 := svcs[1]
	if s2.Name != "log-collector" {
		t.Errorf("expected Name 'log-collector', got '%s'", s2.Name)
	}
	if s2.Mode != "global" || s2.Replicas != "2 (global)" {
		t.Errorf("expected global mode and replica string '2 (global)', got mode=%s replicas=%s", s2.Mode, s2.Replicas)
	}

	// Malformed JSON
	if parseSwarmServices([]byte(`invalid`), nil) != nil {
		t.Errorf("expected nil for malformed services JSON")
	}
}

func TestParseSwarmCLIInfo(t *testing.T) {
	activeOutput := "active true 3 1"
	active, isMgr, nodes, mgrs := parseSwarmCLIInfo(activeOutput)
	if !active || !isMgr || nodes != 3 || mgrs != 1 {
		t.Errorf("parseSwarmCLIInfo(%q) = (%v, %v, %d, %d); want (true, true, 3, 1)", activeOutput, active, isMgr, nodes, mgrs)
	}

	inactiveOutput := "inactive false 0 0"
	active, _, _, _ = parseSwarmCLIInfo(inactiveOutput)
	if active {
		t.Errorf("expected active=false for inactive output")
	}

	emptyOutput := ""
	active, _, _, _ = parseSwarmCLIInfo(emptyOutput)
	if active {
		t.Errorf("expected active=false for empty output")
	}
}
