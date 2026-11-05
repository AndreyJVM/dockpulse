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

func TestFormatUptime(t *testing.T) {
	cases := []struct {
		sec      uint64
		expected string
	}{
		{0, "00:00"},
		{59, "00:00"},
		{65, "00:01"},
		{3665, "01:01"},
		{90000, "1 дн 01:00"},
		{172800, "2 дн 00:00"},
	}

	for _, c := range cases {
		got := formatUptime(c.sec)
		if got != c.expected {
			t.Errorf("formatUptime(%d) = %q; want %q", c.sec, got, c.expected)
		}
	}
}
