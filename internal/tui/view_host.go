package tui

import (
	"fmt"
	"strings"
)

func (m Model) renderHostTab() string {
	var s strings.Builder

	// 1. Host System Information
	s.WriteString(titleStyle.Render(m.i18n.HostSysInfoTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		dimStyle.Render(m.i18n.HostLabel), boldStyle.Render(m.sysInfo.Hostname),
		dimStyle.Render(m.i18n.UptimeLabel), greenStyle.Render(m.formatUptime(m.sysInfo.UptimeSec)),
	))
	s.WriteString(fmt.Sprintf("  %s %s (%s %s)\n\n",
		dimStyle.Render(m.i18n.OSLabel), m.sysInfo.OS, m.i18n.KernelLabel, m.sysInfo.Kernel,
	))

	// 2. Host Resources (CPU & RAM)
	s.WriteString(titleStyle.Render(m.i18n.HostPerfTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s  %s\n",
		fmt.Sprintf(m.i18n.CPUCoresLabel, m.cpuInfo.Cores), formatBar(m.cpuInfo.Percent, 22), renderSparkline(m.cpuHistory),
	))
	s.WriteString(fmt.Sprintf("  %s %s  %s\n\n",
		fmt.Sprintf(m.i18n.RAMLabel, m.memInfo.UsedGB, m.memInfo.TotalGB), formatBar(m.memInfo.Percent, 22), renderSparkline(m.memHistory),
	))

	// 3. Docker Service Status
	s.WriteString(titleStyle.Render(m.i18n.DockerStatusTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s\n",
		m.i18n.DaemonStatus, m.formatDockerStatus(m.dockerInfo.Status),
		boldStyle.Render(fmt.Sprintf(m.i18n.ContainersSummary, m.dockerInfo.RunningContainers, m.dockerInfo.StoppedContainers)),
	))
	s.WriteString(fmt.Sprintf("  %s %d    %s %s\n",
		dimStyle.Render(m.i18n.ImagesCountLabel), m.dockerInfo.ImagesCount,
		dimStyle.Render(m.i18n.VersionLabel), m.dockerInfo.Version,
	))

	return s.String()
}
