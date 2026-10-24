package monitor

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	netUtil "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
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

// PartitionInfo содержит данные о разделе накопителя.
type PartitionInfo struct {
	Mountpoint string
	TotalGB    float64
	UsedGB     float64
	Percent    float64
}

// NetworkInfo содержит данные о сети и текущей скорости.
type NetworkInfo struct {
	IP        string
	DNS       string
	RXSpeed   string
	TXSpeed   string
	SpeedText string
}

// ProcessInfo содержит информацию о запущенном процессе.
type ProcessInfo struct {
	PID        int32
	Name       string
	CPUPercent float64
	MemPercent float32
	RSSMB      float64
}

// NetworkTracker отслеживает счетчики байт сети для вычисления скорости.
type NetworkTracker struct {
	lastTime time.Time
	lastRX   uint64
	lastTX   uint64
}

var netTracker = &NetworkTracker{}

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

// GetDisksInfo возвращает информацию о смонтированных накопителях.
func GetDisksInfo() []PartitionInfo {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	var res []PartitionInfo
	for _, p := range parts {
		if strings.HasPrefix(p.Fstype, "squashfs") || strings.HasPrefix(p.Fstype, "tmpfs") || strings.HasPrefix(p.Fstype, "devtmpfs") {
			continue
		}
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		res = append(res, PartitionInfo{
			Mountpoint: p.Mountpoint,
			TotalGB:    float64(usage.Total) / (1024 * 1024 * 1024),
			UsedGB:     float64(usage.Used) / (1024 * 1024 * 1024),
			Percent:    usage.UsedPercent,
		})
	}
	return res
}

// GetNetworkInfo рассчитывает скорость сети и возвращает сетевые реквизиты.
func GetNetworkInfo() NetworkInfo {
	ip := getDefaultIP()
	dns := getDNSServers()

	rxSpeed, txSpeed := netTracker.calculateSpeed()

	return NetworkInfo{
		IP:        ip,
		DNS:       dns,
		RXSpeed:   formatSpeed(rxSpeed),
		TXSpeed:   formatSpeed(txSpeed),
		SpeedText: fmt.Sprintf("↓ %s  ↑ %s", formatSpeed(rxSpeed), formatSpeed(txSpeed)),
	}
}

// GetTopProcesses возвращает список процессов с наибольшей нагрузкой.
func GetTopProcesses(limit int) []ProcessInfo {
	procs, err := process.Processes()
	if err != nil || len(procs) == 0 {
		return nil
	}

	var list []ProcessInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}

		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		rssMB := 0.0
		if memInfo != nil {
			rssMB = float64(memInfo.RSS) / (1024 * 1024)
		}

		list = append(list, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPUPercent: cpuPct,
			MemPercent: memPct,
			RSSMB:      rssMB,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CPUPercent > list[j].CPUPercent
	})

	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

func getDefaultIP() string {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 800*time.Millisecond)
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}
	return "127.0.0.1"
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

func parseDNSServers(content string) string {
	var servers []string
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "nameserver") && !strings.HasPrefix(l, "#") {
			parts := strings.Fields(l)
			if len(parts) > 1 {
				servers = append(servers, parts[1])
			}
		}
	}

	if len(servers) > 0 {
		maxCount := len(servers)
		if maxCount > 3 {
			maxCount = 3
		}
		return strings.Join(servers[:maxCount], ", ")
	}
	return "System defaults"
}

func getDNSServers() string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return "System defaults"
	}
	return parseDNSServers(string(data))
}

func (t *NetworkTracker) calculateSpeed() (float64, float64) {
	counters, err := netUtil.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return 0, 0
	}

	now := time.Now()
	totalRX := counters[0].BytesRecv
	totalTX := counters[0].BytesSent

	if t.lastTime.IsZero() {
		t.lastTime = now
		t.lastRX = totalRX
		t.lastTX = totalTX
		return 0, 0
	}

	dt := now.Sub(t.lastTime).Seconds()
	if dt <= 0 {
		return 0, 0
	}

	rxDiff := float64(0)
	if totalRX >= t.lastRX {
		rxDiff = float64(totalRX - t.lastRX)
	}

	txDiff := float64(0)
	if totalTX >= t.lastTX {
		txDiff = float64(totalTX - t.lastTX)
	}

	t.lastTime = now
	t.lastRX = totalRX
	t.lastTX = totalTX

	return rxDiff / dt, txDiff / dt
}

func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
