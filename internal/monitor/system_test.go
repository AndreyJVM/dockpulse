package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
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

func TestGetSystemInfo_Fallbacks(t *testing.T) {
	origHostInfo := hostInfoGetter
	origOSHostname := osHostnameGetter
	t.Cleanup(func() {
		hostInfoGetter = origHostInfo
		osHostnameGetter = origOSHostname
	})

	t.Run("host info error and hostname success", func(t *testing.T) {
		hostInfoGetter = func() (*host.InfoStat, error) {
			return nil, errors.New("host info error")
		}
		osHostnameGetter = func() (string, error) {
			return "fallback-host", nil
		}

		info := GetSystemInfo()
		if info.Hostname != "fallback-host" {
			t.Errorf("expected hostname 'fallback-host', got %q", info.Hostname)
		}
		if info.OS != defaultUnknownOS {
			t.Errorf("expected OS %q, got %q", defaultUnknownOS, info.OS)
		}
		if info.Kernel != defaultEmptyKernel {
			t.Errorf("expected Kernel %q, got %q", defaultEmptyKernel, info.Kernel)
		}
		if info.Uptime != defaultEmptyUptime {
			t.Errorf("expected Uptime %q, got %q", defaultEmptyUptime, info.Uptime)
		}
	})

	t.Run("host info error and hostname error", func(t *testing.T) {
		hostInfoGetter = func() (*host.InfoStat, error) {
			return nil, errors.New("host info error")
		}
		osHostnameGetter = func() (string, error) {
			return "", errors.New("hostname error")
		}

		info := GetSystemInfo()
		if info.Hostname != defaultUnknownHost {
			t.Errorf("expected hostname %q, got %q", defaultUnknownHost, info.Hostname)
		}
	})

	t.Run("host info with empty fields fallback to OS and defaults", func(t *testing.T) {
		hostInfoGetter = func() (*host.InfoStat, error) {
			return &host.InfoStat{
				Hostname:        "",
				Platform:        "",
				PlatformVersion: "",
				OS:              "linux",
				KernelVersion:   "",
				Uptime:          3600,
			}, nil
		}
		osHostnameGetter = func() (string, error) {
			return "", nil
		}

		info := GetSystemInfo()
		if info.Hostname != defaultUnknownHost {
			t.Errorf("expected hostname %q, got %q", defaultUnknownHost, info.Hostname)
		}
		if info.OS != "linux" {
			t.Errorf("expected OS 'linux', got %q", info.OS)
		}
		if info.Kernel != defaultEmptyKernel {
			t.Errorf("expected Kernel %q, got %q", defaultEmptyKernel, info.Kernel)
		}
		if info.UptimeSec != 3600 {
			t.Errorf("expected UptimeSec 3600, got %d", info.UptimeSec)
		}
	})
}

func TestGetCPUInfo(t *testing.T) {
	cpu := GetCPUInfo()
	if cpu.Cores <= 0 {
		t.Errorf("expected CPU cores > 0, got %d", cpu.Cores)
	}
}

func TestGetCPUInfo_Fallbacks(t *testing.T) {
	origCPUPercent := cpuPercentGetter
	origCPUCounts := cpuCountsGetter
	t.Cleanup(func() {
		cpuPercentGetter = origCPUPercent
		cpuCountsGetter = origCPUCounts
	})

	t.Run("cpu percent and counts error", func(t *testing.T) {
		cpuPercentGetter = func(_ time.Duration, _ bool) ([]float64, error) {
			return nil, errors.New("cpu percent error")
		}
		cpuCountsGetter = func(_ bool) (int, error) {
			return 0, errors.New("cpu counts error")
		}

		cpu := GetCPUInfo()
		if cpu.Percent != 0.0 {
			t.Errorf("expected Percent 0.0, got %f", cpu.Percent)
		}
		if cpu.Cores != defaultCoreCount {
			t.Errorf("expected Cores %d, got %d", defaultCoreCount, cpu.Cores)
		}
	})

	t.Run("cpu percent negative normalized to zero", func(t *testing.T) {
		cpuPercentGetter = func(_ time.Duration, _ bool) ([]float64, error) {
			return []float64{-5.0}, nil
		}
		cpuCountsGetter = func(_ bool) (int, error) {
			return 8, nil
		}

		cpu := GetCPUInfo()
		if cpu.Percent != 0.0 {
			t.Errorf("expected Percent 0.0, got %f", cpu.Percent)
		}
		if cpu.Cores != 8 {
			t.Errorf("expected Cores 8, got %d", cpu.Cores)
		}
	})
}

func TestGetMemoryInfo(t *testing.T) {
	mem := GetMemoryInfo()
	if mem.TotalGB <= 0 {
		t.Errorf("expected total memory > 0, got %f", mem.TotalGB)
	}
}

func TestGetMemoryInfo_Fallbacks(t *testing.T) {
	origVM := virtualMemoryGetter
	t.Cleanup(func() {
		virtualMemoryGetter = origVM
	})

	t.Run("virtual memory error", func(t *testing.T) {
		virtualMemoryGetter = func() (*mem.VirtualMemoryStat, error) {
			return nil, errors.New("mem error")
		}

		info := GetMemoryInfo()
		if info.TotalGB != 0 || info.UsedGB != 0 || info.Percent != 0 {
			t.Errorf("expected zero MemoryInfo, got %+v", info)
		}
	})

	t.Run("virtual memory values converted properly", func(t *testing.T) {
		virtualMemoryGetter = func() (*mem.VirtualMemoryStat, error) {
			return &mem.VirtualMemoryStat{
				Total:       2 * bytesPerGB,
				Used:        1 * bytesPerGB,
				Available:   1 * bytesPerGB,
				UsedPercent: 50.0,
			}, nil
		}

		info := GetMemoryInfo()
		if info.TotalGB != 2.0 || info.UsedGB != 1.0 || info.AvailableGB != 1.0 || info.Percent != 50.0 {
			t.Errorf("unexpected MemoryInfo: %+v", info)
		}
	})
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
