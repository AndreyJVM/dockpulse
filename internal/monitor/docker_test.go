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

func TestCalculateContainerStats(t *testing.T) {
	var raw ContainerStatsRaw
	raw.CPUStats.CPUUsage.TotalUsage = 200000000
	raw.PreCPUStats.CPUUsage.TotalUsage = 100000000
	raw.CPUStats.SystemCPUUsage = 1000000000
	raw.PreCPUStats.SystemCPUUsage = 900000000
	raw.CPUStats.OnlineCPUs = 2

	// 50 MB usage out of 100 MB limit
	raw.MemoryStats.Usage = 50 * 1024 * 1024
	raw.MemoryStats.Limit = 100 * 1024 * 1024

	cpu, memUsageMB, memLimitMB, memPerc := calculateContainerStats(raw)

	// CPU delta = 100,000,000; System delta = 100,000,000; 2 CPUs -> 1.0 * 2 * 100 = 200%
	if cpu != 200.0 {
		t.Errorf("expected CPU 200.0%%, got %.2f%%", cpu)
	}

	if memUsageMB != 50.0 {
		t.Errorf("expected MemUsageMB 50.0, got %.2f", memUsageMB)
	}

	if memLimitMB != 100.0 {
		t.Errorf("expected MemLimitMB 100.0, got %.2f", memLimitMB)
	}

	if memPerc != 50.0 {
		t.Errorf("expected MemPerc 50.0%%, got %.2f%%", memPerc)
	}
}
