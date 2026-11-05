package monitor

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemInfo содержит общие данные о хосте.
type SystemInfo struct {
	Hostname  string
	OS        string
	Kernel    string
	Uptime    string
	UptimeSec uint64
}

// CPUInfo содержит метрики процессора.
type CPUInfo struct {
	Percent float64
	Cores   int
}

// MemoryInfo содержит метрики оперативной памяти.
type MemoryInfo struct {
	TotalGB     float64
	UsedGB      float64
	AvailableGB float64
	Percent     float64
}

// GetSystemInfo возвращает базовую информацию о хосте.
func GetSystemInfo() SystemInfo {
	hInfo, err := host.Info()
	hostname, _ := os.Hostname()

	if err != nil {
		return SystemInfo{
			Hostname: hostname,
			OS:       "Unknown",
			Kernel:   "-",
			Uptime:   "-",
		}
	}

	uptimeStr := formatUptime(hInfo.Uptime)

	return SystemInfo{
		Hostname:  hInfo.Hostname,
		OS:        fmt.Sprintf("%s %s", hInfo.Platform, hInfo.PlatformVersion),
		Kernel:    hInfo.KernelVersion,
		Uptime:    uptimeStr,
		UptimeSec: hInfo.Uptime,
	}
}

// GetCPUInfo возвращает нагрузку на CPU и число ядер.
func GetCPUInfo() CPUInfo {
	percentages, err := cpu.Percent(0, false)
	pct := 0.0
	if err == nil && len(percentages) > 0 {
		pct = percentages[0]
	}

	cores, err := cpu.Counts(true)
	if err != nil || cores == 0 {
		cores = 1
	}

	return CPUInfo{
		Percent: pct,
		Cores:   cores,
	}
}

// GetMemoryInfo возвращает статистику RAM.
func GetMemoryInfo() MemoryInfo {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}
	}

	return MemoryInfo{
		TotalGB:     float64(vm.Total) / (1024 * 1024 * 1024),
		UsedGB:      float64(vm.Used) / (1024 * 1024 * 1024),
		AvailableGB: float64(vm.Available) / (1024 * 1024 * 1024),
		Percent:     vm.UsedPercent,
	}
}

func formatUptime(uptimeSec uint64) string {
	days := uptimeSec / (60 * 60 * 24)
	hours := (uptimeSec % (60 * 60 * 24)) / 3600
	minutes := (uptimeSec % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%d дн %02d:%02d", days, hours, minutes)
	}
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}
