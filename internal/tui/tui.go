package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/AndreyJVM/system-monitor/internal/monitor"
)

// Стили оформления Lip Gloss
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2563EB")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#06B6D4")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38BDF8")).
			MarginTop(1).
			MarginBottom(0)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	boldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F3F4F6"))

	greenStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22C55E"))

	yellowStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EAB308"))

	redStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EF4444"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginTop(1)
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model представляет состояние TUI.
type Model struct {
	activeTab  int
	width      int
	height     int
	sysInfo    monitor.SystemInfo
	cpuInfo    monitor.CPUInfo
	memInfo    monitor.MemoryInfo
	disksInfo  []monitor.PartitionInfo
	netInfo    monitor.NetworkInfo
	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	swarm      monitor.SwarmSummary
	lastUpdate string
}

// NewModel создает новую TUI модель.
func NewModel() Model {
	m := Model{activeTab: 1}
	m.fetchAllData()
	return m
}

func (m *Model) fetchAllData() {
	m.sysInfo = monitor.GetSystemInfo()
	m.cpuInfo = monitor.GetCPUInfo()
	m.memInfo = monitor.GetMemoryInfo()
	m.disksInfo = monitor.GetDisksInfo()
	m.netInfo = monitor.GetNetworkInfo()
	m.dockerInfo = monitor.GetDockerSummary()
	m.compose = monitor.GetComposeProjects(m.dockerInfo.Containers)
	m.swarm = monitor.GetSwarmSummary()
	m.lastUpdate = time.Now().Format("15:04:05")
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.activeTab = 1
		case "2":
			m.activeTab = 2
		case "3":
			m.activeTab = 3
		case "4":
			m.activeTab = 4
		case "r", "R":
			m.fetchAllData()
		}

	case tickMsg:
		m.fetchAllData()
		return m, tickCmd()
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// 1. Верхняя панель заголовка
	header := headerStyle.Render("⚡ System, Docker, Compose & Swarm Monitor") +
		dimStyle.Render("  v0.3.0 (Go Edition)  ") +
		greenStyle.Render("["+m.lastUpdate+"]")
	b.WriteString(header + "\n\n")

	// 2. Вкладки
	tabs := []string{
		"1: Host Overview",
		"2: Docker Containers",
		"3: Compose Stacks",
		"4: Swarm Cluster",
	}

	var renderedTabs []string
	for i, tab := range tabs {
		tabIdx := i + 1
		if m.activeTab == tabIdx {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(fmt.Sprintf("[%s]", tab)))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(fmt.Sprintf(" %s ", tab)))
		}
	}
	b.WriteString(strings.Join(renderedTabs, " ") + "\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", max(m.width, 70))) + "\n")

	// 3. Содержимое активной вкладки
	switch m.activeTab {
	case 1:
		b.WriteString(m.renderHostTab())
	case 2:
		b.WriteString(m.renderDockerTab())
	case 3:
		b.WriteString(m.renderComposeTab())
	case 4:
		b.WriteString(m.renderSwarmTab())
	}

	// 4. Подвал
	b.WriteString("\n" + dimStyle.Render(strings.Repeat("─", max(m.width, 70))) + "\n")
	b.WriteString(footerStyle.Render(
		boldStyle.Render("[1-4]") + " Вкладки  |  " +
			boldStyle.Render("[R]") + " Обновить  |  " +
			boldStyle.Render("[Q]") + " Выход",
	))

	return b.String()
}

func (m Model) renderHostTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🖥️  СИСТЕМНАЯ ИНФОРМАЦИЯ") + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		dimStyle.Render("Хост:"), boldStyle.Render(m.sysInfo.Hostname),
		dimStyle.Render("Uptime:"), greenStyle.Render(m.sysInfo.Uptime),
	))
	s.WriteString(fmt.Sprintf("  %s %s (Kernel: %s)\n\n",
		dimStyle.Render("ОС:"), m.sysInfo.OS, m.sysInfo.Kernel,
	))

	s.WriteString(titleStyle.Render("⚡  ПРОИЗВОДИТЕЛЬНОСТЬ ХОСТА") + "\n")
	s.WriteString(fmt.Sprintf("  CPU (%d cores): %s\n",
		m.cpuInfo.Cores, formatBar(m.cpuInfo.Percent, 28),
	))
	s.WriteString(fmt.Sprintf("  RAM (%.1f/%.1f GB): %s\n\n",
		m.memInfo.UsedGB, m.memInfo.TotalGB, formatBar(m.memInfo.Percent, 28),
	))

	s.WriteString(titleStyle.Render("🌐  СЕТЬ И ТРАФИК") + "\n")
	s.WriteString(fmt.Sprintf("  IP: %s    DNS: %s\n",
		boldStyle.Render(m.netInfo.IP), m.netInfo.DNS,
	))
	s.WriteString(fmt.Sprintf("  Скорость: %s    %s\n\n",
		greenStyle.Render("↓ RX: "+m.netInfo.RXSpeed),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Render("↑ TX: "+m.netInfo.TXSpeed),
	))

	s.WriteString(titleStyle.Render("💾  ДИСКОВЫЕ РАЗДЕЛЫ") + "\n")
	if len(m.disksInfo) > 0 {
		for _, d := range m.disksInfo {
			s.WriteString(fmt.Sprintf("  %-12s %s  (%.1f/%.1f GB)\n",
				d.Mountpoint, formatBar(d.Percent, 20), d.UsedGB, d.TotalGB,
			))
		}
	} else {
		s.WriteString("  " + dimStyle.Render("Дисковые разделы недоступны") + "\n")
	}

	s.WriteString("\n" + titleStyle.Render("🐳  DOCKER СТАТУС") + "\n")
	s.WriteString(fmt.Sprintf("  Статус: %s    Контейнеры: %s\n",
		m.dockerInfo.StatusDisplay,
		boldStyle.Render(fmt.Sprintf("%d активно, %d остановлено", m.dockerInfo.RunningContainers, m.dockerInfo.StoppedContainers)),
	))

	return s.String()
}

func (m Model) renderDockerTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🐳  DOCKER КОНТЕЙНЕРЫ") + "\n")
	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %s\n\n",
		dimStyle.Render("Служба:"), m.dockerInfo.StatusDisplay,
		dimStyle.Render("Образов:"), m.dockerInfo.ImagesCount,
		dimStyle.Render("Версия:"), m.dockerInfo.Version,
	))

	if len(m.dockerInfo.Containers) == 0 {
		if m.dockerInfo.Status == "running" {
			s.WriteString("  " + dimStyle.Render("Контейнеры не запущены.") + "\n")
		} else {
			s.WriteString("  " + yellowStyle.Render("Демон Docker не запущен или недоступен.") + "\n")
		}
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %-12s %-20s %-24s %-12s %-20s\n",
		dimStyle.Render("ID"), dimStyle.Render("NAME"), dimStyle.Render("IMAGE"),
		dimStyle.Render("STATE"), dimStyle.Render("PORTS"),
	))
	s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 88)) + "\n")

	for _, c := range m.dockerInfo.Containers {
		stateColored := dimStyle.Render(c.State)
		if c.State == "running" {
			stateColored = greenStyle.Render("running")
		} else if c.State == "paused" {
			stateColored = yellowStyle.Render("paused")
		}

		s.WriteString(fmt.Sprintf("  %-12s %-20s %-24s %-20s %-20s\n",
			c.ID, boldStyle.Render(truncate(c.Name, 18)), truncate(c.Image, 22),
			stateColored, truncate(c.Ports, 20),
		))
	}

	return s.String()
}

func (m Model) renderComposeTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("📦  DOCKER COMPOSE ПРОЕКТЫ (%d)", len(m.compose))) + "\n\n")

	if len(m.compose) == 0 {
		s.WriteString("  " + dimStyle.Render("Проекты Docker Compose не обнаружены (нет меток com.docker.compose.project).") + "\n")
		return s.String()
	}

	for _, p := range m.compose {
		s.WriteString(fmt.Sprintf("┌─ %s  [%s]  (Сервисов поднято: %d/%d)\n",
			boldStyle.Render(p.Name), p.StatusDisplay, p.RunningServices, p.TotalServices,
		))
		for _, svc := range p.Services {
			st := dimStyle.Render("○ " + svc.State)
			if svc.State == "running" {
				st = greenStyle.Render("● running")
			}
			s.WriteString(fmt.Sprintf("│   • %-16s %-20s %-22s %s\n",
				boldStyle.Render(svc.Name), st, truncate(svc.Image, 20), svc.Ports,
			))
		}
		s.WriteString("└" + strings.Repeat("─", 65) + "\n\n")
	}

	return s.String()
}

func (m Model) renderSwarmTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🌐  DOCKER SWARM КЛАСТЕР") + "\n\n")

	if !m.swarm.IsActive {
		s.WriteString(fmt.Sprintf("  Статус: %s\n\n", m.swarm.StatusDisplay))
		s.WriteString("  " + dimStyle.Render("Хост не входит в активный кластер Swarm.") + "\n")
		s.WriteString("  Для создания кластера выполните: " + boldStyle.Render("docker swarm init") + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %d    %s %d\n\n",
		dimStyle.Render("Статус:"), m.swarm.StatusDisplay,
		dimStyle.Render("Нод в кластере:"), m.swarm.NodesCount,
		dimStyle.Render("Менеджеров:"), m.swarm.ManagersCount,
	))

	if len(m.swarm.Nodes) > 0 {
		s.WriteString(boldStyle.Render("  УЗЛЫ КЛАСТЕРА (NODES):") + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-10s\n",
			dimStyle.Render("NODE ID"), dimStyle.Render("HOSTNAME"), dimStyle.Render("ROLE"),
			dimStyle.Render("STATE"), dimStyle.Render("LEADER"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 66)) + "\n")
		for _, n := range m.swarm.Nodes {
			leaderMark := ""
			if n.IsLeader {
				leaderMark = yellowStyle.Render("★ Leader")
			}
			s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-10s\n",
				n.ID, boldStyle.Render(truncate(n.Hostname, 18)), n.Role, n.State, leaderMark,
			))
		}
		s.WriteString("\n")
	}

	if len(m.swarm.Services) > 0 {
		s.WriteString(boldStyle.Render("  СЛУЖБЫ SWARM (SERVICES):") + "\n")
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-24s %-14s\n",
			dimStyle.Render("SERVICE ID"), dimStyle.Render("NAME"), dimStyle.Render("IMAGE"),
			dimStyle.Render("REPLICAS"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 72)) + "\n")
		for _, svc := range m.swarm.Services {
			s.WriteString(fmt.Sprintf("  %-14s %-20s %-24s %-14s\n",
				svc.ID, boldStyle.Render(truncate(svc.Name, 18)), truncate(svc.Image, 22), greenStyle.Render(svc.Replicas),
			))
		}
	}

	return s.String()
}

func formatBar(percent float64, width int) string {
	pct := maxFloat(0.0, minFloat(100.0, percent))
	filled := int(float64(width) * pct / 100.0)
	empty := width - filled

	color := greenStyle
	if pct >= 90.0 {
		color = redStyle
	} else if pct >= 70.0 {
		color = yellowStyle
	}

	bar := color.Render(strings.Repeat("█", filled)) + dimStyle.Render(strings.Repeat("░", empty))
	return fmt.Sprintf("[%s] %s", bar, color.Render(fmt.Sprintf("%5.1f%%", pct)))
}

func truncate(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-2] + ".."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
