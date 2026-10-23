package monitor

import (
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	info := GetSystemInfo()
	if info.Hostname == "" {
		t.Errorf("expected non-empty hostname")
	}
	if info.OS == "" {
		t.Errorf("expected non-empty OS")
	}
}

func TestGetCPUInfo(t *testing.T) {
	cpu := GetCPUInfo()
	if cpu.Cores <= 0 {
		t.Errorf("expected CPU cores > 0, got %d", cpu.Cores)
	}
}

func TestGetMemoryInfo(t *testing.T) {
	mem := GetMemoryInfo()
	if mem.TotalGB <= 0 {
		t.Errorf("expected total memory > 0, got %f", mem.TotalGB)
	}
}

func TestFormatSpeed(t *testing.T) {
	cases := []struct {
		bytesPerSec float64
		expected    string
	}{
		{500, "500 B/s"},
		{1024, "1.0 KB/s"},
		{1048576, "1.00 MB/s"},
		{1073741824, "1.00 GB/s"},
	}

	for _, c := range cases {
		got := formatSpeed(c.bytesPerSec)
		if got != c.expected {
			t.Errorf("formatSpeed(%f) = %s; want %s", c.bytesPerSec, got, c.expected)
		}
	}
}
