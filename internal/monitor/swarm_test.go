package monitor

import (
	"context"
	"errors"
	"net/http"
	"strings"
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
		"node1234567890abcdef":   4,
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
		"svc1234567890abcdef":  3,
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

func TestFetchSwarmTasks(t *testing.T) {
	// Success case
	clientOK := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "/tasks") {
			return mockStringResponse(http.StatusOK, `[
				{"ID":"t1","ServiceID":"s1","NodeID":"n1","Status":{"State":"running"}},
				{"ID":"t2","ServiceID":"s1","NodeID":"n2","Status":{"State":"running"}}
			]`), nil
		}
		return mockStringResponse(http.StatusNotFound, "not found"), nil
	})

	svc, node := fetchSwarmTasks(clientOK)
	if svc["s1"] != 2 || node["n1"] != 1 || node["n2"] != 1 {
		t.Errorf("unexpected fetchSwarmTasks result: svc=%+v node=%+v", svc, node)
	}

	// Failure case
	clientErr := newMockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return mockStringResponse(http.StatusInternalServerError, "error"), nil
	})
	svcErr, nodeErr := fetchSwarmTasks(clientErr)
	if len(svcErr) != 0 || len(nodeErr) != 0 {
		t.Errorf("expected empty maps on HTTP error")
	}
}

func TestFetchSwarmNodes(t *testing.T) {
	// Success case
	clientOK := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "/nodes") {
			return mockStringResponse(http.StatusOK, `[
				{"ID":"n1","Spec":{"Role":"manager"},"Description":{"Hostname":"node-1"},"Status":{"State":"ready"}}
			]`), nil
		}
		return mockStringResponse(http.StatusNotFound, "not found"), nil
	})

	nodes := fetchSwarmNodes(clientOK, map[string]int{"n1": 3})
	if len(nodes) != 1 || nodes[0].Hostname != "node-1" || nodes[0].ActiveTasks != 3 {
		t.Errorf("unexpected fetchSwarmNodes result: %+v", nodes)
	}

	// Failure case
	clientErr := newMockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("network failure")
	})
	nodesErr := fetchSwarmNodes(clientErr, nil)
	if nodesErr != nil {
		t.Errorf("expected nil on error")
	}
}

func TestFetchSwarmServices(t *testing.T) {
	// Success case
	clientOK := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "/services") {
			return mockStringResponse(http.StatusOK, `[
				{"ID":"s1","Spec":{"Name":"api-service","TaskTemplate":{"ContainerSpec":{"Image":"api:latest"}},"Mode":{"Replicated":{"Replicas":2}}}}
			]`), nil
		}
		return mockStringResponse(http.StatusNotFound, "not found"), nil
	})

	svcs := fetchSwarmServices(clientOK, map[string]int{"s1": 2})
	if len(svcs) != 1 || svcs[0].Name != "api-service" || svcs[0].Replicas != "2/2" {
		t.Errorf("unexpected fetchSwarmServices result: %+v", svcs)
	}

	// Failure case
	clientErr := newMockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return mockStringResponse(http.StatusInternalServerError, "server error"), nil
	})
	svcsErr := fetchSwarmServices(clientErr, nil)
	if svcsErr != nil {
		t.Errorf("expected nil on error")
	}
}

func TestGetSwarmSummary_SocketActive(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
	}()

	dockerSocketFinder = func() string { return "/tmp/mock-docker.sock" }
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			path := req.URL.Path
			switch {
			case strings.Contains(path, "/info"):
				return mockStringResponse(http.StatusOK, `{
					"Swarm": {
						"LocalNodeState": "active",
						"ControlAvailable": true,
						"NodeID": "mgrnode12345678",
						"Nodes": 3,
						"Managers": 1
					}
				}`), nil
			case strings.Contains(path, "/tasks"):
				return mockStringResponse(http.StatusOK, `[
					{"ID":"t1","ServiceID":"s1","NodeID":"mgrnode12345678","Status":{"State":"running"}}
				]`), nil
			case strings.Contains(path, "/nodes"):
				return mockStringResponse(http.StatusOK, `[
					{"ID":"mgrnode12345678","Spec":{"Role":"manager"},"Description":{"Hostname":"master-1"},"Status":{"State":"ready"},"ManagerStatus":{"Leader":true}}
				]`), nil
			case strings.Contains(path, "/services"):
				return mockStringResponse(http.StatusOK, `[
					{"ID":"s1","Spec":{"Name":"web","TaskTemplate":{"ContainerSpec":{"Image":"nginx"}},"Mode":{"Replicated":{"Replicas":1}}}}
				]`), nil
			default:
				return mockStringResponse(http.StatusNotFound, "not found"), nil
			}
		})
	}

	summary := GetSwarmSummary()
	if !summary.IsActive || !summary.IsManager {
		t.Errorf("expected active manager Swarm summary, got %+v", summary)
	}
	if summary.NodesCount != 3 || len(summary.Nodes) != 1 || len(summary.Services) != 1 {
		t.Errorf("unexpected counts: nodes=%d, services=%d", len(summary.Nodes), len(summary.Services))
	}
	if !strings.Contains(summary.Text, "3 нод") {
		t.Errorf("unexpected summary text: %s", summary.Text)
	}
}

func TestGetSwarmSummary_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("active true 4 2"), nil
	}

	summary := GetSwarmSummary()
	if !summary.IsActive || !summary.IsManager {
		t.Errorf("expected active manager Swarm summary from CLI: %+v", summary)
	}
	if summary.NodesCount != 4 || summary.ManagersCount != 2 {
		t.Errorf("expected 4 nodes, 2 managers, got %d/%d", summary.NodesCount, summary.ManagersCount)
	}
}

func TestGetSwarmSummary_CLIFallback_Inactive(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("inactive false 0 0"), nil
	}

	summary := GetSwarmSummary()
	if summary.IsActive {
		t.Errorf("expected inactive Swarm summary from CLI")
	}
	if !strings.Contains(summary.Text, "не активирован") {
		t.Errorf("expected inactive text, got %q", summary.Text)
	}
}

func TestGetSwarmServiceLogs_Socket(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
	}()

	dockerSocketFinder = func() string { return "/tmp/mock-docker.sock" }
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "/services/svc-1/logs") {
				return mockStringResponse(http.StatusOK, "Swarm service live logs output\nLine 2"), nil
			}
			return mockStringResponse(http.StatusNotFound, "not found"), nil
		})
	}

	logs, err := GetSwarmServiceLogs("svc-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(logs, "Swarm service live logs output") {
		t.Errorf("unexpected logs: %s", logs)
	}
}

func TestGetSwarmServiceLogs_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("CLI Swarm log line 1\nCLI Swarm log line 2"), nil
	}

	logs, err := GetSwarmServiceLogs("svc-2", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(logs, "CLI Swarm log line 1") {
		t.Errorf("unexpected CLI logs: %s", logs)
	}
}

func TestGetSwarmServiceLogs_CLIError(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("Error: no such service: svc-3"), errors.New("exit status 1")
	}

	_, err := GetSwarmServiceLogs("svc-3", 10)
	if err == nil {
		t.Errorf("expected error when fetching logs for non-existent service")
	}
}

func TestGetSwarmServiceLogs_DefaultTail(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	var capturedArgs []string
	dockerCLICommandRunner = func(_ context.Context, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte("log output"), nil
	}

	_, err := GetSwarmServiceLogs("svc-test", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Check that tail defaulted to 100
	foundTail := false
	for i, arg := range capturedArgs {
		if arg == "--tail" && i+1 < len(capturedArgs) && capturedArgs[i+1] == "100" {
			foundTail = true
			break
		}
	}
	if !foundTail {
		t.Errorf("expected default tail 100 in args: %v", capturedArgs)
	}
}

func TestTruncateSwarmID(t *testing.T) {
	if got := truncateSwarmID("12345"); got != "12345" {
		t.Errorf("expected 12345, got %s", got)
	}
	if got := truncateSwarmID("1234567890abcdef123"); got != "1234567890ab" {
		t.Errorf("expected 1234567890ab, got %s", got)
	}
}

func TestLookupTaskCount(t *testing.T) {
	if got := lookupTaskCount(nil, "full", "short"); got != 0 {
		t.Errorf("expected 0 for nil map, got %d", got)
	}
	taskMap := map[string]int{"fullID": 5, "shortID": 2}
	if got := lookupTaskCount(taskMap, "fullID", "shortID"); got != 5 {
		t.Errorf("expected 5 for fullID, got %d", got)
	}
	if got := lookupTaskCount(taskMap, "unknown", "shortID"); got != 2 {
		t.Errorf("expected 2 for shortID, got %d", got)
	}
	if got := lookupTaskCount(taskMap, "unknown", "unknown"); got != 0 {
		t.Errorf("expected 0 for unknown, got %d", got)
	}
}

func TestCleanSwarmImage(t *testing.T) {
	if got := cleanSwarmImage("nginx:alpine@sha256:123456"); got != "nginx:alpine" {
		t.Errorf("expected nginx:alpine, got %s", got)
	}
	if got := cleanSwarmImage("redis:latest"); got != "redis:latest" {
		t.Errorf("expected redis:latest, got %s", got)
	}
}

func TestFormatSwarmServicePorts(t *testing.T) {
	if got := formatSwarmServicePorts(nil); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	ports := []struct {
		Protocol      string `json:"Protocol"`
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
	}{
		{Protocol: "tcp", TargetPort: 80, PublishedPort: 8080},
		{Protocol: "udp", TargetPort: 53, PublishedPort: 0},
	}
	if got := formatSwarmServicePorts(ports); got != "8080:80/tcp, 53/udp" {
		t.Errorf("expected 8080:80/tcp, 53/udp, got %q", got)
	}
}

func TestParseSwarmCLINodes(t *testing.T) {
	output := "node1234567890\thost-1\tmanager\tready\tactive\tLeader\nworker123456789\thost-2\tworker\tdown\tdrain\t\n\n"
	nodes := parseSwarmCLINodes(output)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].ID != "node12345678" || nodes[0].Hostname != "host-1" || !nodes[0].IsLeader {
		t.Errorf("unexpected leader node: %+v", nodes[0])
	}
	if nodes[1].ID != "worker123456" || nodes[1].Role != "worker" || nodes[1].IsLeader {
		t.Errorf("unexpected worker node: %+v", nodes[1])
	}
}

func TestParseSwarmCLIServices(t *testing.T) {
	output := "svc1234567890\tapi-service\treplicated\t2/2\tapi:v1@sha256:abc\t*:8080->80/tcp\nsvcGlobal12345\tlogger\tglobal\t1/1 (global)\tfluentd:latest\t\n"
	svcs := parseSwarmCLIServices(output)
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
	if svcs[0].Name != "api-service" || svcs[0].RunningReplicas != 2 || svcs[0].TargetReplicas != 2 || svcs[0].Image != "api:v1" {
		t.Errorf("unexpected replicated service: %+v", svcs[0])
	}
	if svcs[1].Name != "logger" || svcs[1].Mode != "global" || svcs[1].RunningReplicas != 1 || svcs[1].TargetReplicas != 1 {
		t.Errorf("unexpected global service: %+v", svcs[1])
	}
}

func TestFetchSocketAPI_Errors(t *testing.T) {
	// Nil client
	if _, err := fetchSocketAPI(nil, "/test", swarmTimeout); err == nil {
		t.Errorf("expected error for nil client")
	}

	// Non-200 status code
	client404 := newMockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return mockStringResponse(http.StatusNotFound, "not found"), nil
	})
	if _, err := fetchSocketAPI(client404, "/test", swarmTimeout); err == nil {
		t.Errorf("expected error for 404 response")
	}
}

func TestGetSwarmServiceLogs_EmptyID(t *testing.T) {
	if _, err := GetSwarmServiceLogs("   ", 100); err == nil {
		t.Errorf("expected error for empty service ID")
	}
}
