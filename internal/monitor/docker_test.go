package monitor

import (
	"bytes"
	"encoding/binary"
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
	if summary.Status != DockerStatusRunning && summary.Status != DockerStatusStopped && summary.Status != DockerStatusNotInstalled {
		t.Errorf("unexpected docker status: %s", summary.Status)
	}
}

func TestContainerOperationsEmptyID(t *testing.T) {
	emptyID := "   "

	if err := StartContainer(emptyID); err == nil {
		t.Errorf("expected error when starting empty container ID")
	}

	if err := StopContainer(emptyID); err == nil {
		t.Errorf("expected error when stopping empty container ID")
	}

	if err := RestartContainer(emptyID); err == nil {
		t.Errorf("expected error when restarting empty container ID")
	}

	if err := KillContainer(emptyID); err == nil {
		t.Errorf("expected error when killing empty container ID")
	}

	if err := RemoveContainer(emptyID); err == nil {
		t.Errorf("expected error when removing empty container ID")
	}

	if _, err := GetContainerLogs(emptyID, 10); err == nil {
		t.Errorf("expected error when fetching logs for empty container ID")
	}

	if _, err := GetContainerDetail(emptyID); err == nil {
		t.Errorf("expected error when fetching details for empty container ID")
	}

	if cmd := GetContainerExecCommand(emptyID); cmd != nil {
		t.Errorf("expected nil exec command for empty container ID")
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

	if err := KillContainer(invalidID); err == nil {
		t.Errorf("expected error when killing invalid container ID")
	}

	if err := RemoveContainer(invalidID); err == nil {
		t.Errorf("expected error when removing invalid container ID")
	}

	if _, err := GetContainerLogs(invalidID, 10); err == nil {
		t.Errorf("expected error when fetching logs for invalid container ID")
	}

	if _, err := GetContainerDetail(invalidID); err == nil {
		t.Errorf("expected error when fetching details for invalid container ID")
	}
}

func TestDemuxDockerLogs(t *testing.T) {
	t.Run("plain text without headers", func(t *testing.T) {
		plain := "Hello world log output"
		res := demuxDockerLogs(strings.NewReader(plain))
		if res != plain {
			t.Errorf("expected '%s', got '%s'", plain, res)
		}
	})

	t.Run("multiplexed stdout and stderr frames", func(t *testing.T) {
		var buf bytes.Buffer

		// Frame 1: stdout (streamType 1)
		stdoutMsg := "Line from stdout\n"
		var hdr1 [8]byte
		hdr1[0] = 1 // stdout
		binary.BigEndian.PutUint32(hdr1[4:8], uint32(len(stdoutMsg)))
		buf.Write(hdr1[:])
		buf.WriteString(stdoutMsg)

		// Frame 2: stderr (streamType 2)
		stderrMsg := "Line from stderr\n"
		var hdr2 [8]byte
		hdr2[0] = 2 // stderr
		binary.BigEndian.PutUint32(hdr2[4:8], uint32(len(stderrMsg)))
		buf.Write(hdr2[:])
		buf.WriteString(stderrMsg)

		expected := stdoutMsg + stderrMsg
		res := demuxDockerLogs(&buf)
		if res != expected {
			t.Errorf("expected demuxed %q, got %q", expected, res)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		if res := demuxDockerLogs(strings.NewReader("")); res != "" {
			t.Errorf("expected empty string for empty input, got %q", res)
		}
	})
}

func TestFormatPorts(t *testing.T) {
	tests := []struct {
		name     string
		ports    []rawPort
		expected string
	}{
		{
			name:     "empty",
			ports:    nil,
			expected: "",
		},
		{
			name: "public and private",
			ports: []rawPort{
				{PublicPort: 8080, PrivatePort: 80, Type: "tcp"},
			},
			expected: "8080->80/tcp",
		},
		{
			name: "private only",
			ports: []rawPort{
				{PrivatePort: 5432, Type: "tcp"},
			},
			expected: "5432/tcp",
		},
		{
			name: "multiple ports",
			ports: []rawPort{
				{PublicPort: 80, PrivatePort: 80, Type: "tcp"},
				{PublicPort: 443, PrivatePort: 443, Type: "tcp"},
			},
			expected: "80->80/tcp, 443->443/tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPorts(tt.ports)
			if got != tt.expected {
				t.Errorf("formatPorts() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCalculateContainerStats(t *testing.T) {
	t.Run("standard stats calculation", func(t *testing.T) {
		var raw ContainerStatsRaw
		raw.CPUStats.CPUUsage.TotalUsage = 200000000
		raw.PreCPUStats.CPUUsage.TotalUsage = 100000000
		raw.CPUStats.SystemCPUUsage = 1000000000
		raw.PreCPUStats.SystemCPUUsage = 900000000
		raw.CPUStats.OnlineCPUs = 2

		// 50 MB usage out of 100 MB limit
		raw.MemoryStats.Usage = 50 * 1024 * 1024
		raw.MemoryStats.Limit = 100 * 1024 * 1024

		// Networks map
		raw.Networks = map[string]struct {
			RxBytes uint64 `json:"rx_bytes"`
			TxBytes uint64 `json:"tx_bytes"`
		}{
			"eth0": {RxBytes: 15 * 1024 * 1024, TxBytes: 25 * 1024 * 1024},
			"eth1": {RxBytes: 5 * 1024 * 1024, TxBytes: 10 * 1024 * 1024},
		}

		// Block I/O
		raw.BlkioStats.IOServiceBytesRecursive = []struct {
			Major uint64 `json:"major"`
			Minor uint64 `json:"minor"`
			Op    string `json:"op"`
			Value uint64 `json:"value"`
		}{
			{Major: 8, Minor: 0, Op: "Read", Value: 12 * 1024 * 1024},
			{Major: 8, Minor: 0, Op: "Write", Value: 34 * 1024 * 1024},
			{Major: 8, Minor: 0, Op: "Sync", Value: 1000},
		}

		stats := calculateContainerStats(raw)

		// CPU delta = 100,000,000; System delta = 100,000,000; 2 CPUs -> 1.0 * 2 * 100 = 200%
		if stats.CPUPerc != 200.0 {
			t.Errorf("expected CPU 200.0%%, got %.2f%%", stats.CPUPerc)
		}

		if stats.MemUsageMB != 50.0 {
			t.Errorf("expected MemUsageMB 50.0, got %.2f", stats.MemUsageMB)
		}

		if stats.MemLimitMB != 100.0 {
			t.Errorf("expected MemLimitMB 100.0, got %.2f", stats.MemLimitMB)
		}

		if stats.MemPerc != 50.0 {
			t.Errorf("expected MemPerc 50.0%%, got %.2f%%", stats.MemPerc)
		}

		// Net I/O sum: 15MB + 5MB = 20MB, 25MB + 10MB = 35MB
		expectedRx := uint64(20 * 1024 * 1024)
		expectedTx := uint64(35 * 1024 * 1024)
		if stats.NetRxBytes != expectedRx {
			t.Errorf("expected NetRxBytes %d, got %d", expectedRx, stats.NetRxBytes)
		}
		if stats.NetTxBytes != expectedTx {
			t.Errorf("expected NetTxBytes %d, got %d", expectedTx, stats.NetTxBytes)
		}

		// Block I/O: 12MB read, 34MB write
		expectedRead := uint64(12 * 1024 * 1024)
		expectedWrite := uint64(34 * 1024 * 1024)
		if stats.BlockReadBytes != expectedRead {
			t.Errorf("expected BlockReadBytes %d, got %d", expectedRead, stats.BlockReadBytes)
		}
		if stats.BlockWriteBytes != expectedWrite {
			t.Errorf("expected BlockWriteBytes %d, got %d", expectedWrite, stats.BlockWriteBytes)
		}
	})

	t.Run("inactive file cache deduction", func(t *testing.T) {
		var raw ContainerStatsRaw
		raw.MemoryStats.Usage = 100 * 1024 * 1024
		raw.MemoryStats.Limit = 200 * 1024 * 1024
		raw.MemoryStats.Stats.TotalInactiveFile = 20 * 1024 * 1024

		stats := calculateContainerStats(raw)
		// 100 MB - 20 MB = 80 MB
		if stats.MemUsageMB != 80.0 {
			t.Errorf("expected MemUsageMB 80.0 after cache deduction, got %.2f", stats.MemUsageMB)
		}
	})

	t.Run("network fallback to single network struct", func(t *testing.T) {
		var raw ContainerStatsRaw
		raw.Network.RxBytes = 1000
		raw.Network.TxBytes = 2000

		stats := calculateContainerStats(raw)
		if stats.NetRxBytes != 1000 || stats.NetTxBytes != 2000 {
			t.Errorf("expected NetRx=1000, NetTx=2000, got NetRx=%d, NetTx=%d", stats.NetRxBytes, stats.NetTxBytes)
		}
	})
}

func TestParseHealthStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"exact healthy", "healthy", "healthy"},
		{"exact unhealthy", "unhealthy", "unhealthy"},
		{"exact starting", "starting", "starting"},
		{"status with (healthy)", "Up 2 hours (healthy)", "healthy"},
		{"status with (unhealthy)", "Up 35 minutes (unhealthy)", "unhealthy"},
		{"status with (health: starting)", "Up 12 seconds (health: starting)", "starting"},
		{"status without healthcheck", "Up 4 days", ""},
		{"empty string", "", ""},
		{"whitespace string", "   ", ""},
		{"case insensitive HEALTHY", "HEALTHY", "healthy"},
		{"case insensitive (HEALTHY)", "Up 1 hour (HEALTHY)", "healthy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseHealthStatus(tt.input)
			if got != tt.expected {
				t.Errorf("ParseHealthStatus(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetContainerExecCommand(t *testing.T) {
	containerID := "abc123def456"
	cmd := GetContainerExecCommand(containerID)
	if cmd == nil {
		t.Fatal("expected non-nil *exec.Cmd")
	}

	expectedArgs := []string{
		"docker", "exec", "-it", containerID, "sh", "-c",
		"if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi",
	}

	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d (%v)", len(expectedArgs), len(cmd.Args), cmd.Args)
	}

	for i, arg := range cmd.Args {
		if arg != expectedArgs[i] {
			t.Errorf("arg[%d] = %q, want %q", i, arg, expectedArgs[i])
		}
	}
}

func TestIsDockerCLIAvailable(_ *testing.T) {
	// Function must execute safely without panics
	_ = IsDockerCLIAvailable()
}
