package monitor

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestGetDockerNetworksAndVolumesViaSocket(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "/tmp/mock.sock" }
	dockerCLICommandRunner = func(_ context.Context, args ...string) ([]byte, error) {
		t.Fatalf("CLI must not run after a successful socket request: %v", args)
		return nil, errors.New("unexpected CLI call")
	}
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/networks":
				return mockStringResponse(http.StatusOK, `[
					{"Id":"network1234567890","Name":"frontend","Driver":"bridge","Scope":"local","Internal":false,"Attachable":true,"Containers":{"c1":{}},"Labels":{"com.docker.compose.project":"shop"},"IPAM":{"Config":[{"Subnet":"172.20.0.0/16","Gateway":"172.20.0.1"}]}}
				]`), nil
			case "/volumes":
				return mockStringResponse(http.StatusOK, `{"Volumes":[
					{"Name":"shop_data","Driver":"local","Scope":"local","Mountpoint":"/var/lib/docker/volumes/shop_data/_data","Labels":{"com.docker.compose.project":"shop"},"UsageData":{"Size":4096,"RefCount":2}}
				]}`), nil
			default:
				return mockStringResponse(http.StatusNotFound, "not found"), nil
			}
		})
	}

	networks, err := GetDockerNetworks()
	if err != nil {
		t.Fatalf("GetDockerNetworks() returned an error: %v", err)
	}
	wantNetwork := DockerNetwork{ID: "network12345", Name: "frontend", Driver: "bridge", Scope: "local", Attachable: true, Containers: 1, Subnet: "172.20.0.0/16", Gateway: "172.20.0.1", ComposeProject: "shop"}
	if !reflect.DeepEqual(networks, []DockerNetwork{wantNetwork}) {
		t.Fatalf("unexpected networks: %#v", networks)
	}

	volumes, err := GetDockerVolumes()
	if err != nil {
		t.Fatalf("GetDockerVolumes() returned an error: %v", err)
	}
	wantVolume := DockerVolume{Name: "shop_data", Driver: "local", Scope: "local", Mountpoint: "/var/lib/docker/volumes/shop_data/_data", Size: 4096, RefCount: 2, ComposeProject: "shop"}
	if !reflect.DeepEqual(volumes, []DockerVolume{wantVolume}) {
		t.Fatalf("unexpected volumes: %#v", volumes)
	}
}

func TestGetDockerNetworksAndVolumesCLIFallbackIsReadOnly(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	var calls [][]string
	dockerCLICommandRunner = func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		switch args[0] {
		case "network":
			return []byte("network1234567890|frontend|bridge|local|false|com.docker.compose.project=shop\n"), nil
		case "volume":
			return []byte("shop_data|local|local|/data|com.docker.compose.project=shop\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}

	networks, err := GetDockerNetworks()
	if err != nil {
		t.Fatalf("GetDockerNetworks() returned an error: %v", err)
	}
	volumes, err := GetDockerVolumes()
	if err != nil {
		t.Fatalf("GetDockerVolumes() returned an error: %v", err)
	}
	if len(networks) != 1 || networks[0].ComposeProject != "shop" {
		t.Fatalf("unexpected CLI network data: %#v", networks)
	}
	if len(volumes) != 1 || volumes[0].ComposeProject != "shop" || volumes[0].Size != -1 {
		t.Fatalf("unexpected CLI volume data: %#v", volumes)
	}

	wantCalls := [][]string{
		{"network", "ls", "--no-trunc", "--format", "{{.ID}}|{{.Name}}|{{.Driver}}|{{.Scope}}|{{.Internal}}|{{.Labels}}"},
		{"volume", "ls", "--format", "{{.Name}}|{{.Driver}}|{{.Scope}}|{{.Mountpoint}}|{{.Labels}}"},
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("unexpected Docker CLI calls: %#v", calls)
	}
}

func TestGetDockerNetworksAndVolumesReportUnavailable(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return nil, errors.New("Docker unavailable")
	}

	if _, err := GetDockerNetworks(); err == nil {
		t.Fatal("expected a network discovery error")
	}
	if _, err := GetDockerVolumes(); err == nil {
		t.Fatal("expected a volume discovery error")
	}
}
