package monitor

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestPruneContainers_Socket(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
	}()

	dockerSocketFinder = func() string { return "/tmp/mock-docker.sock" }
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "/containers/prune") {
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"ContainersDeleted": []string{"c1", "c2"},
					"SpaceReclaimed":    2097152,
				}), nil
			}
			return mockStringResponse(http.StatusNotFound, "not found"), nil
		})
	}

	res, err := PruneContainers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "2") || !strings.Contains(res, "2.0 MB") {
		t.Errorf("unexpected prune containers result: %s", res)
	}
}

func TestPruneContainers_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("Deleted Containers:\nabc\ndef\nTotal reclaimed space: 15.5MB"), nil
	}

	res, err := PruneContainers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "15.5MB") {
		t.Errorf("unexpected CLI prune result: %s", res)
	}
}

func TestPruneContainers_CLIError(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("docker daemon not running"), errors.New("exit status 1")
	}

	_, err := PruneContainers()
	if err == nil {
		t.Errorf("expected error when CLI prune fails")
	}
}

func TestPruneImages_Socket(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
	}()

	dockerSocketFinder = func() string { return "/tmp/mock-docker.sock" }
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "/images/prune") {
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"ImagesDeleted":  []map[string]interface{}{{"Deleted": "img1"}},
					"SpaceReclaimed": 5242880,
				}), nil
			}
			return mockStringResponse(http.StatusNotFound, "not found"), nil
		})
	}

	res, err := PruneImages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "1") || !strings.Contains(res, "5.0 MB") {
		t.Errorf("unexpected prune images result: %s", res)
	}
}

func TestPruneImages_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("Total reclaimed space: 120MB"), nil
	}

	res, err := PruneImages()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "120MB") {
		t.Errorf("unexpected CLI prune images result: %s", res)
	}
}

func TestPruneImages_CLIError(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("error pruning images"), errors.New("exit status 1")
	}

	_, err := PruneImages()
	if err == nil {
		t.Errorf("expected error when CLI image prune fails")
	}
}

func TestPruneVolumes_Socket(t *testing.T) {
	origFinder := dockerSocketFinder
	origFactory := socketClientFactory
	defer func() {
		dockerSocketFinder = origFinder
		socketClientFactory = origFactory
	}()

	dockerSocketFinder = func() string { return "/tmp/mock-docker.sock" }
	socketClientFactory = func(_ string) *http.Client {
		return newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "/volumes/prune") {
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"VolumesDeleted": []string{"v1", "v2"},
					"SpaceReclaimed": 10485760,
				}), nil
			}
			return mockStringResponse(http.StatusNotFound, "not found"), nil
		})
	}

	res, err := PruneVolumes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "2") || !strings.Contains(res, "10.0 MB") {
		t.Errorf("unexpected prune volumes result: %s", res)
	}
}

func TestPruneVolumes_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("Total reclaimed space: 50MB"), nil
	}

	res, err := PruneVolumes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "50MB") {
		t.Errorf("unexpected CLI prune volumes result: %s", res)
	}
}

func TestPruneVolumes_CLIError(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("error pruning volumes"), errors.New("exit status 1")
	}

	_, err := PruneVolumes()
	if err == nil {
		t.Errorf("expected error when CLI volume prune fails")
	}
}

func TestPruneAll_Socket(t *testing.T) {
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
			case strings.Contains(path, "/containers/prune"):
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"ContainersDeleted": []string{"c1"},
					"SpaceReclaimed":    1048576,
				}), nil
			case strings.Contains(path, "/images/prune"):
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"ImagesDeleted":  []map[string]interface{}{{"Deleted": "i1"}},
					"SpaceReclaimed": 2097152,
				}), nil
			case strings.Contains(path, "/volumes/prune"):
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"VolumesDeleted": []string{"v1"},
					"SpaceReclaimed": 3145728,
				}), nil
			case strings.Contains(path, "/build/prune"):
				return mockJSONResponse(http.StatusOK, map[string]interface{}{
					"CachesDeleted":  []string{"bc1", "bc2"},
					"SpaceReclaimed": 4194304,
				}), nil
			default:
				return mockStringResponse(http.StatusNotFound, "not found"), nil
			}
		})
	}

	res, err := PruneAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "контейнеров") || !strings.Contains(res, "образов") ||
		!strings.Contains(res, "томов") || !strings.Contains(res, "Кэш сборки") {
		t.Errorf("unexpected PruneAll result: %s", res)
	}
}

func TestPruneAll_CLIFallback(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("Total reclaimed space: 250MB"), nil
	}

	res, err := PruneAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "250MB") {
		t.Errorf("unexpected CLI prune all result: %s", res)
	}
}

func TestPruneAll_CLIError(t *testing.T) {
	origFinder := dockerSocketFinder
	origRunner := dockerCLICommandRunner
	defer func() {
		dockerSocketFinder = origFinder
		dockerCLICommandRunner = origRunner
	}()

	dockerSocketFinder = func() string { return "" }
	dockerCLICommandRunner = func(_ context.Context, _ ...string) ([]byte, error) {
		return []byte("error pruning all"), errors.New("exit status 1")
	}

	_, err := PruneAll()
	if err == nil {
		t.Errorf("expected error when CLI prune all fails")
	}
}

func TestFormatPruneCLIOutput(t *testing.T) {
	outputWithReclaimed := "Deleted Containers:\nabc123\nTotal reclaimed space: 14.5MB"
	res := formatPruneCLIOutput(outputWithReclaimed, "контейнеров")
	if !strings.Contains(res, "14.5MB") || !strings.Contains(res, "контейнеров") {
		t.Errorf("unexpected output: %s", res)
	}

	outputWithoutReclaimed := "Deleted Images: 0"
	res2 := formatPruneCLIOutput(outputWithoutReclaimed, "образов")
	if !strings.Contains(res2, "успешно выполнена") {
		t.Errorf("unexpected fallback output: %s", res2)
	}
}
