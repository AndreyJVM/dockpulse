package tui

import (
	"fmt"
	"strings"
)

func (m Model) renderHostTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("СИСТЕМНАЯ ИНФОРМАЦИЯ") + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		dimStyle.Render("Хост:"), boldStyle.Render(m.sysInfo.Hostname),
		dimStyle.Render("Uptime:"), greenStyle.Render(m.sysInfo.Uptime),
	))
	s.WriteString(fmt.Sprintf("  %s %s (Kernel: %s)\n\n",
		dimStyle.Render("ОС:"), m.sysInfo.OS, m.sysInfo.Kernel,
	))

	s.WriteString(titleStyle.Render("ПРОИЗВОДИТЕЛЬНОСТЬ ХОСТА") + "\n")
	s.WriteString(fmt.Sprintf("  CPU (%d cores): %s  %s\n",
		m.cpuInfo.Cores, formatBar(m.cpuInfo.Percent, 22), renderSparkline(m.cpuHistory),
	))
	s.WriteString(fmt.Sprintf("  RAM (%.1f/%.1f GB): %s  %s\n\n",
		m.memInfo.UsedGB, m.memInfo.TotalGB, formatBar(m.memInfo.Percent, 22), renderSparkline(m.memHistory),
	))

	s.WriteString(titleStyle.Render("СЕТЬ И ТРАФИК") + "\n")
	s.WriteString(fmt.Sprintf("  IP: %s    DNS: %s\n",
		boldStyle.Render(m.netInfo.IP), m.netInfo.DNS,
	))
	s.WriteString(fmt.Sprintf("  Скорость: %s    %s\n\n",
		greenStyle.Render("↓ RX: "+m.netInfo.RXSpeed),
		cyanStyle.Render("↑ TX: "+m.netInfo.TXSpeed),
	))

	s.WriteString(titleStyle.Render("ДИСКОВЫЕ РАЗДЕЛЫ") + "\n")
	if len(m.disksInfo) > 0 {
		for _, d := range m.disksInfo {
			s.WriteString(fmt.Sprintf("  %-12s %s  (%.1f/%.1f GB)\n",
				d.Mountpoint, formatBar(d.Percent, 20), d.UsedGB, d.TotalGB,
			))
		}
	} else {
		s.WriteString("  " + dimStyle.Render("Дисковые разделы недоступны") + "\n")
	}

	if len(m.topProcs) > 0 {
		s.WriteString("\n" + titleStyle.Render("ТОП ПРОЦЕССОВ (ПО CPU)") + "\n")
		s.WriteString(fmt.Sprintf("  %-8s %-24s %-10s %-12s\n",
			dimStyle.Render("PID"), dimStyle.Render("ИМЯ"),
			dimStyle.Render("CPU %"), dimStyle.Render("RAM (MB)"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 56)) + "\n")
		for _, p := range m.topProcs {
			s.WriteString(fmt.Sprintf("  %-8d %-24s %-10s %-12s\n",
				p.PID, boldStyle.Render(truncate(p.Name, 22)),
				formatPercent(p.CPUPercent), fmt.Sprintf("%.1f MB", p.RSSMB),
			))
		}
	}

	s.WriteString("\n" + titleStyle.Render("DOCKER СТАТУС") + "\n")
	s.WriteString(fmt.Sprintf("  Статус: %s    Контейнеры: %s\n",
		m.dockerInfo.StatusDisplay,
		boldStyle.Render(fmt.Sprintf("%d активно, %d остановлено", m.dockerInfo.RunningContainers, m.dockerInfo.StoppedContainers)),
	))

	return s.String()
}
