package monitor

import (
	"os"
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
