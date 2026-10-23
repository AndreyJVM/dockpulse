package monitor

import (
	"os"
	"strings"
	"testing"
)

func TestFindDockerSocket(t *testing.T) {
	// Should not crash
	sock := FindDockerSocket()
	t.Logf("Detected Docker socket: '%s'", sock)

	// Test override with DOCKER_HOST
	tmpFile, err := os.CreateTemp("", "docker-test-*.sock")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	origHost := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", origHost)

	os.Setenv("DOCKER_HOST", "unix://"+tmpFile.Name())
	detected := FindDockerSocket()
	if detected != tmpFile.Name() {
		t.Errorf("expected '%s', got '%s'", tmpFile.Name(), detected)
	}
}

func TestGetDockerSummary(t *testing.T) {
	// Call should safely complete whether Docker is installed/running or not
	summary := GetDockerSummary()
	if summary.Status != "running" && summary.Status != "stopped" && summary.Status != "not_installed" {
		t.Errorf("unexpected docker status: %s", summary.Status)
	}
}

func TestContainerOperationsInvalidID(t *testing.T) {
	invalidID := "non-existent-container-99999"

	// All operations on invalid ID should safely return an error without panic
	if err := StartContainer(invalidID); err == nil {
		t.Errorf("expected error when starting invalid container ID")
	}

	if err := StopContainer(invalidID); err == nil {
		t.Errorf("expected error when stopping invalid container ID")
	}

	if err := RestartContainer(invalidID); err == nil {
		t.Errorf("expected error when restarting invalid container ID")
	}

	if _, err := GetContainerLogs(invalidID, 10); err == nil {
		t.Errorf("expected error when fetching logs for invalid container ID")
	}

	if _, err := GetContainerDetail(invalidID); err == nil {
		t.Errorf("expected error when fetching details for invalid container ID")
	}
}

func TestDemuxDockerLogs(t *testing.T) {
	// Plain string stream without demux headers
	plain := "Hello world log output"
	res := demuxDockerLogs(strings.NewReader(plain))
	if res != plain {
		t.Errorf("expected '%s', got '%s'", plain, res)
	}
}
