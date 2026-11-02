package tui

import (
	"fmt"
	"strings"
)

func (m Model) renderHostTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(m.i18n.HostSysInfoTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		dimStyle.Render(m.i18n.HostLabel), boldStyle.Render(m.sysInfo.Hostname),
		dimStyle.Render(m.i18n.UptimeLabel), greenStyle.Render(m.formatUptime(m.sysInfo.UptimeSec)),
	))
	s.WriteString(fmt.Sprintf("  %s %s (%s %s)\n\n",
		dimStyle.Render(m.i18n.OSLabel), m.sysInfo.OS, m.i18n.KernelLabel, m.sysInfo.Kernel,
	))

	s.WriteString(titleStyle.Render(m.i18n.HostPerfTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s  %s\n",
		fmt.Sprintf(m.i18n.CPUCoresLabel, m.cpuInfo.Cores), formatBar(m.cpuInfo.Percent, 22), renderSparkline(m.cpuHistory),
	))
	s.WriteString(fmt.Sprintf("  %s %s  %s\n\n",
		fmt.Sprintf(m.i18n.RAMLabel, m.memInfo.UsedGB, m.memInfo.TotalGB), formatBar(m.memInfo.Percent, 22), renderSparkline(m.memHistory),
	))

	s.WriteString(titleStyle.Render(m.i18n.NetTrafficTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		m.i18n.IPLabel, boldStyle.Render(m.netInfo.IP), m.i18n.DNSLabel, m.netInfo.DNS,
	))
	s.WriteString(fmt.Sprintf("  %s %s    %s\n\n",
		m.i18n.SpeedLabel,
		greenStyle.Render("↓ RX: "+m.netInfo.RXSpeed),
		cyanStyle.Render("↑ TX: "+m.netInfo.TXSpeed),
	))

	s.WriteString(titleStyle.Render(m.i18n.DisksTitle) + "\n")
	if len(m.disksInfo) > 0 {
		for _, d := range m.disksInfo {
			s.WriteString(fmt.Sprintf("  %-12s %s  (%.1f/%.1f GB)\n",
				d.Mountpoint, formatBar(d.Percent, 20), d.UsedGB, d.TotalGB,
			))
		}
	} else {
		s.WriteString("  " + dimStyle.Render(m.i18n.DisksUnavailable) + "\n")
	}

	if len(m.topProcs) > 0 {
		s.WriteString("\n" + titleStyle.Render(m.i18n.TopProcsTitle) + "\n")
		s.WriteString(fmt.Sprintf("  %-8s %-24s %-10s %-12s\n",
			dimStyle.Render(m.i18n.ProcPID), dimStyle.Render(m.i18n.ProcName),
			dimStyle.Render(m.i18n.ProcCPU), dimStyle.Render(m.i18n.ProcRAM),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 56)) + "\n")
		for _, p := range m.topProcs {
			s.WriteString(fmt.Sprintf("  %-8d %-24s %-10s %-12s\n",
				p.PID, boldStyle.Render(truncate(p.Name, 22)),
				formatPercent(p.CPUPercent), fmt.Sprintf("%.1f MB", p.RSSMB),
			))
		}
	}

	s.WriteString("\n" + titleStyle.Render(m.i18n.DockerStatusTitle) + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s\n",
		m.i18n.DaemonStatus, m.formatDockerStatus(m.dockerInfo.Status),
		boldStyle.Render(fmt.Sprintf(m.i18n.ContainersSummary, m.dockerInfo.RunningContainers, m.dockerInfo.StoppedContainers)),
	))

	return s.String()
}
