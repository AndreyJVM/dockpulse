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

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#0284C7"))

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

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#38BDF8")).
			Padding(1, 2)
)

type tickMsg time.Time

func tickCmd(d time.Duration) tea.Cmd {
	if d <= 0 {
		d = 2 * time.Second
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Model представляет состояние TUI.
type Model struct {
	version         string
	activeTab       int
	refreshInterval time.Duration
	showHelp        bool
	width           int
	height          int

	// Навигация и модальные окна
	cursorIdx           int
	composeCursorIdx    int
	swarmCursorIdx      int
	activeModal         string // "", "logs", "inspect"
	logsTargetName      string
	logsTargetID        string
	logsLines           []string
	logsScroll          int
	inspectTarget       monitor.ContainerDetail
	actionMessage       string
	actionMessageExpiry time.Time

	// Поиск и фильтрация контейнеров
	isFiltering bool
	filterQuery string

	sysInfo    monitor.SystemInfo
	cpuInfo    monitor.CPUInfo
	memInfo    monitor.MemoryInfo
	disksInfo  []monitor.PartitionInfo
	netInfo    monitor.NetworkInfo
	topProcs   []monitor.ProcessInfo
	dockerInfo monitor.DockerSummary
	compose    []monitor.ComposeProject
	swarm      monitor.SwarmSummary
	lastUpdate string
}

// NewModel создает новую TUI модель с настраиваемым интервалом обновления и начальной вкладкой.
func NewModel(refreshInterval time.Duration, initialTab int, version string) Model {
	if initialTab < 1 || initialTab > 4 {
		initialTab = 1
	}
	if refreshInterval <= 0 {
		refreshInterval = 2 * time.Second
	}
	m := Model{
		version:         version,
		activeTab:       initialTab,
		refreshInterval: refreshInterval,
	}
	m.fetchAllData()
	return m
}

func (m *Model) fetchAllData() {
	m.sysInfo = monitor.GetSystemInfo()
	m.cpuInfo = monitor.GetCPUInfo()
	m.memInfo = monitor.GetMemoryInfo()
	m.disksInfo = monitor.GetDisksInfo()
	m.netInfo = monitor.GetNetworkInfo()
	m.topProcs = monitor.GetTopProcesses(5)
	m.dockerInfo = monitor.GetDockerSummary()
	m.compose = monitor.GetComposeProjects(m.dockerInfo.Containers)
	m.swarm = monitor.GetSwarmSummary()
	m.lastUpdate = time.Now().Format("15:04:05")

	m.clampCursor()
	if m.composeCursorIdx >= len(m.compose) {
		m.composeCursorIdx = max(0, len(m.compose)-1)
	}
	if m.swarmCursorIdx >= len(m.swarm.Services) {
		m.swarmCursorIdx = max(0, len(m.swarm.Services)-1)
	}
}

func (m Model) getFilteredContainers() []monitor.ContainerItem {
	if strings.TrimSpace(m.filterQuery) == "" {
		return m.dockerInfo.Containers
	}
	q := strings.ToLower(strings.TrimSpace(m.filterQuery))
	var res []monitor.ContainerItem
	for _, c := range m.dockerInfo.Containers {
		if strings.Contains(strings.ToLower(c.Name), q) ||
			strings.Contains(strings.ToLower(c.Image), q) ||
			strings.Contains(strings.ToLower(c.ID), q) ||
			strings.Contains(strings.ToLower(c.ComposeProject), q) ||
			strings.Contains(strings.ToLower(c.Health), q) {
			res = append(res, c)
		}
	}
	return res
}

func (m *Model) clampCursor() {
	containers := m.getFilteredContainers()
	if len(containers) == 0 {
		m.cursorIdx = 0
	} else if m.cursorIdx >= len(containers) {
		m.cursorIdx = len(containers) - 1
	} else if m.cursorIdx < 0 {
		m.cursorIdx = 0
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd(m.refreshInterval)
}

func (m *Model) showToast(msg string) {
	m.actionMessage = msg
	m.actionMessageExpiry = time.Now().Add(4 * time.Second)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Если активен ввод фильтра
		if m.isFiltering {
			switch msg.Type {
			case tea.KeyEsc:
				m.isFiltering = false
				m.filterQuery = ""
				m.clampCursor()
				return m, nil
			case tea.KeyEnter:
				m.isFiltering = false
				return m, nil
			case tea.KeyBackspace:
				if len(m.filterQuery) > 0 {
					runes := []rune(m.filterQuery)
					m.filterQuery = string(runes[:len(runes)-1])
					m.clampCursor()
				}
				return m, nil
			case tea.KeyRunes, tea.KeySpace:
				m.filterQuery += msg.String()
				m.clampCursor()
				return m, nil
			}
			return m, nil
		}

		// Если активно модальное окно
		if m.activeModal != "" {
			switch msg.String() {
			case "q", "esc":
				m.activeModal = ""
				return m, nil

			case "up", "k":
				if m.activeModal == "logs" && m.logsScroll > 0 {
					m.logsScroll--
				}
				return m, nil

			case "down", "j":
				if m.activeModal == "logs" && m.logsScroll < len(m.logsLines)-10 {
					m.logsScroll++
				}
				return m, nil

			case "pgup":
				if m.activeModal == "logs" {
					m.logsScroll = max(0, m.logsScroll-10)
				}
				return m, nil

			case "pgdown":
				if m.activeModal == "logs" {
					m.logsScroll = min(max(0, len(m.logsLines)-10), m.logsScroll+10)
				}
				return m, nil
			}
			return m, nil
		}

		// Если открыта справка
		if m.showHelp {
			switch msg.String() {
			case "q", "esc", "?":
				m.showHelp = false
				return m, nil
			}
			return m, nil
		}

		// Основной режим
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		// Навигация по вкладкам
		case "tab", "right":
			m.activeTab++
			if m.activeTab > 4 {
				m.activeTab = 1
			}
		case "shift+tab", "left":
			m.activeTab--
			if m.activeTab < 1 {
				m.activeTab = 4
			}

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
			m.showToast("Данные успешно обновлены")

		// Поиск по списку контейнеров
		case "/":
			if m.activeTab == 2 {
				m.isFiltering = true
				return m, nil
			}

		case "esc":
			if m.filterQuery != "" {
				m.filterQuery = ""
				m.clampCursor()
				return m, nil
			}

		// Навигация и действия в списке контейнеров (Вкладка 2), Compose (Вкладка 3), Swarm (Вкладка 4)
		case "up", "k":
			if m.activeTab == 2 {
				containers := m.getFilteredContainers()
				if len(containers) > 0 {
					m.cursorIdx--
					if m.cursorIdx < 0 {
						m.cursorIdx = len(containers) - 1
					}
				}
			} else if m.activeTab == 3 && len(m.compose) > 0 {
				m.composeCursorIdx--
				if m.composeCursorIdx < 0 {
					m.composeCursorIdx = len(m.compose) - 1
				}
			} else if m.activeTab == 4 && len(m.swarm.Services) > 0 {
				m.swarmCursorIdx--
				if m.swarmCursorIdx < 0 {
					m.swarmCursorIdx = len(m.swarm.Services) - 1
				}
			}
		case "down", "j":
			if m.activeTab == 2 {
				containers := m.getFilteredContainers()
				if len(containers) > 0 {
					m.cursorIdx++
					if m.cursorIdx >= len(containers) {
						m.cursorIdx = 0
					}
				}
			} else if m.activeTab == 3 && len(m.compose) > 0 {
				m.composeCursorIdx++
				if m.composeCursorIdx >= len(m.compose) {
					m.composeCursorIdx = 0
				}
			} else if m.activeTab == 4 && len(m.swarm.Services) > 0 {
				m.swarmCursorIdx++
				if m.swarmCursorIdx >= len(m.swarm.Services) {
					m.swarmCursorIdx = 0
				}
			}

		// Просмотр логов контейнера, проекта Compose или службы Swarm
		case "l", "L":
			if m.activeTab == 2 {
				containers := m.getFilteredContainers()
				if len(containers) > 0 && m.cursorIdx < len(containers) {
					c := containers[m.cursorIdx]
					m.logsTargetName = c.Name
					m.logsTargetID = c.ID
					logs, err := monitor.GetContainerLogs(c.ID, 120)
					if err != nil {
						m.showToast("Ошибка чтения логов: " + err.Error())
					} else {
						m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
						m.logsScroll = max(0, len(m.logsLines)-18)
						m.activeModal = "logs"
					}
				}
			} else if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
				proj := m.compose[m.composeCursorIdx]
				m.logsTargetName = "Compose: " + proj.Name
				m.logsTargetID = proj.Name
				logs, err := monitor.GetComposeProjectLogs(proj.Name, m.dockerInfo.Containers, 150)
				if err != nil {
					m.showToast("Ошибка чтения логов проекта: " + err.Error())
				} else {
					m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
					m.logsScroll = max(0, len(m.logsLines)-18)
					m.activeModal = "logs"
				}
			} else if m.activeTab == 4 && len(m.swarm.Services) > 0 && m.swarmCursorIdx < len(m.swarm.Services) {
				svc := m.swarm.Services[m.swarmCursorIdx]
				m.logsTargetName = "Swarm: " + svc.Name
				m.logsTargetID = svc.ID
				logs, err := monitor.GetSwarmServiceLogs(svc.ID, 150)
				if err != nil {
					m.showToast("Ошибка чтения логов службы: " + err.Error())
				} else {
					m.logsLines = strings.Split(strings.TrimSpace(logs), "\n")
					m.logsScroll = max(0, len(m.logsLines)-18)
					m.activeModal = "logs"
				}
			}

		// Детали контейнера (Inspect)
		case "enter", "i", "I":
			containers := m.getFilteredContainers()
			if m.activeTab == 2 && len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				detail, err := monitor.GetContainerDetail(c.ID)
				if err != nil {
					m.showToast("Ошибка inspect: " + err.Error())
				} else {
					m.inspectTarget = detail
					m.activeModal = "inspect"
				}
			}

		// Старт / Стоп контейнера
		case "s", "S":
			containers := m.getFilteredContainers()
			if m.activeTab == 2 && len(containers) > 0 && m.cursorIdx < len(containers) {
				c := containers[m.cursorIdx]
				if c.State == "running" {
					m.showToast(fmt.Sprintf("Остановка контейнера %s...", c.Name))
					if err := monitor.StopContainer(c.ID); err != nil {
						m.showToast("Ошибка остановки: " + err.Error())
					} else {
						m.fetchAllData()
						m.showToast(fmt.Sprintf("Контейнер %s остановлен", c.Name))
					}
				} else {
					m.showToast(fmt.Sprintf("Запуск контейнера %s...", c.Name))
					if err := monitor.StartContainer(c.ID); err != nil {
						m.showToast("Ошибка запуска: " + err.Error())
					} else {
						m.fetchAllData()
						m.showToast(fmt.Sprintf("Контейнер %s запущен", c.Name))
					}
				}
			}

		// Перезапуск контейнера или проекта Compose
		case "p", "P":
			if m.activeTab == 2 {
				containers := m.getFilteredContainers()
				if len(containers) > 0 && m.cursorIdx < len(containers) {
					c := containers[m.cursorIdx]
					m.showToast(fmt.Sprintf("Перезапуск %s...", c.Name))
					if err := monitor.RestartContainer(c.ID); err != nil {
						m.showToast("Ошибка перезапуска: " + err.Error())
					} else {
						m.fetchAllData()
						m.showToast(fmt.Sprintf("Контейнер %s успешно перезапущен", c.Name))
					}
				}
			} else if m.activeTab == 3 && len(m.compose) > 0 && m.composeCursorIdx < len(m.compose) {
				proj := m.compose[m.composeCursorIdx]
				m.showToast(fmt.Sprintf("Перезапуск Compose проекта %s...", proj.Name))
				if err := monitor.RestartComposeProject(proj.Name, m.dockerInfo.Containers); err != nil {
					m.showToast("Ошибка перезапуска проекта: " + err.Error())
				} else {
					m.fetchAllData()
					m.showToast(fmt.Sprintf("Проект %s успешно перезапущен", proj.Name))
				}
			}
		}

	case tickMsg:
		m.fetchAllData()
		return m, tickCmd(m.refreshInterval)
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// 1. Панель заголовка
	verStr := m.version
	if verStr == "" {
		verStr = "v0.4.0"
	}
	header := headerStyle.Render("System & Docker Monitor") +
		dimStyle.Render(fmt.Sprintf("  %s (Go Edition)  ", verStr)) +
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

	// Всплывающее уведомление о действии
	if m.actionMessage != "" && time.Now().Before(m.actionMessageExpiry) {
		b.WriteString(yellowStyle.Render("[!] "+m.actionMessage) + "\n\n")
	}

	// 3. Модальные окна или содержимое вкладки
	if m.activeModal == "logs" {
		b.WriteString(m.renderLogsModal())
	} else if m.activeModal == "inspect" {
		b.WriteString(m.renderInspectModal())
	} else if m.showHelp {
		b.WriteString("\n" + m.renderHelpBox() + "\n")
	} else {
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
	}

	// 4. Подвал
	b.WriteString("\n" + dimStyle.Render(strings.Repeat("─", max(m.width, 70))) + "\n")
	if m.activeModal != "" {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[Esc / Q]") + " Закрыть окно  |  " +
				boldStyle.Render("[↑ / ↓ / PgUp / PgDn]") + " Прокрутка",
		))
	} else if m.activeTab == 2 {
		if m.isFiltering {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[Enter]") + " Применить фильтр  |  " +
					boldStyle.Render("[Esc]") + " Отмена",
			))
		} else if m.filterQuery != "" {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[↑ / ↓]") + " Выбрать  |  " +
					boldStyle.Render("[Esc]") + " Сбросить фильтр  |  " +
					boldStyle.Render("[L]") + " Логи  |  " +
					boldStyle.Render("[Enter/I]") + " Инфо  |  " +
					boldStyle.Render("[S]") + " Старт/Стоп  |  " +
					boldStyle.Render("[P]") + " Рестарт",
			))
		} else {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[↑ / ↓]") + " Выбрать  |  " +
					boldStyle.Render("[/]") + " Поиск  |  " +
					boldStyle.Render("[L]") + " Логи  |  " +
					boldStyle.Render("[Enter/I]") + " Инфо  |  " +
					boldStyle.Render("[S]") + " Старт/Стоп  |  " +
					boldStyle.Render("[P]") + " Рестарт  |  " +
					boldStyle.Render("[?]") + " Справка",
			))
		}
	} else if m.activeTab == 3 {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[↑ / ↓]") + " Выбрать стек  |  " +
				boldStyle.Render("[L]") + " Логи стека  |  " +
				boldStyle.Render("[P]") + " Перезапуск стека  |  " +
				boldStyle.Render("[?]") + " Справка",
		))
	} else if m.activeTab == 4 {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[↑ / ↓]") + " Выбрать службу  |  " +
				boldStyle.Render("[L]") + " Логи службы  |  " +
				boldStyle.Render("[?]") + " Справка",
		))
	} else {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[1-4 / Tab / ← →]") + " Вкладки  |  " +
				boldStyle.Render("[R]") + " Обновить  |  " +
				boldStyle.Render("[?]") + " Справка  |  " +
				boldStyle.Render("[Q]") + " Выход",
		))
	}

	return b.String()
}

func (m Model) renderLogsModal() string {
	var s strings.Builder
	title := fmt.Sprintf("ЛОГИ: %s (ID: %s)", boldStyle.Render(m.logsTargetName), dimStyle.Render(m.logsTargetID))
	s.WriteString(titleStyle.Render(title) + "\n\n")

	maxVisible := 18
	if len(m.logsLines) == 0 {
		s.WriteString("  " + dimStyle.Render("Логи отсутствуют или служба не выводила сообщений.") + "\n")
	} else {
		start := min(m.logsScroll, max(0, len(m.logsLines)-1))
		end := min(start+maxVisible, len(m.logsLines))

		for i := start; i < end; i++ {
			line := m.logsLines[i]
			s.WriteString(fmt.Sprintf(" %s %s\n", dimStyle.Render("│"), truncate(line, max(40, m.width-8))))
		}

		s.WriteString(fmt.Sprintf("\n  %s\n", dimStyle.Render(fmt.Sprintf("[Строки %d-%d из %d] (прокрутка ↑/↓/PgUp/PgDn)", start+1, end, len(m.logsLines)))))
	}

	return boxStyle.Render(s.String())
}

func (m Model) renderInspectModal() string {
	var s strings.Builder
	d := m.inspectTarget

	title := fmt.Sprintf("ИНСПЕКЦИЯ КОНТЕЙНЕРА: %s", boldStyle.Render(d.Name))
	s.WriteString(titleStyle.Render(title) + "\n\n")

	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("ID:"), d.ID))
	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("Образ:"), boldStyle.Render(d.Image)))

	healthBadge := dimStyle.Render("-")
	switch d.Health {
	case "healthy":
		healthBadge = greenStyle.Render("[healthy]")
	case "unhealthy":
		healthBadge = redStyle.Render("[unhealthy]")
	case "starting":
		healthBadge = yellowStyle.Render("[starting]")
	}

	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n",
		dimStyle.Render("Статус:"), greenStyle.Render(d.Status),
		dimStyle.Render("Здоровье:"), healthBadge,
	))
	if d.State == "running" && (d.CPUPerc > 0 || d.MemUsageMB > 0) {
		s.WriteString(fmt.Sprintf("  %s %.1f%%    %s %.1f MB (%.1f%%)\n",
			dimStyle.Render("CPU:"), d.CPUPerc,
			dimStyle.Render("Память:"), d.MemUsageMB, d.MemPerc,
		))
	}
	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("Создан:"), d.Created))
	s.WriteString(fmt.Sprintf("  %s %s    %s %s\n", dimStyle.Render("IP-адрес:"), boldStyle.Render(d.IP), dimStyle.Render("Шлюз:"), d.Gateway))
	s.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("Команда:"), truncate(d.Command, 60)))

	if len(d.Mounts) > 0 {
		s.WriteString("\n  " + boldStyle.Render("Тома и монтирования (Mounts):") + "\n")
		for _, mnt := range d.Mounts {
			s.WriteString(fmt.Sprintf("   • %s\n", mnt))
		}
	}

	return boxStyle.Render(s.String())
}

func (m Model) renderHelpBox() string {
	content := fmt.Sprintf(`%s

  %s   Переключение вкладок (1: Host, 2: Docker, 3: Compose, 4: Swarm)
  %s   След. / пред. вкладка
  %s   Перемещение курсора (контейнеры, стеки Compose, службы Swarm)
  %s   Поиск и фильтрация контейнеров по имени, образу или ID
  %s   Просмотр логов контейнера, стека Compose или службы Swarm
  %s   Подробная инспекция выбранного контейнера (IP, тома, сеть)
  %s   Запуск / остановка контейнера
  %s   Перезапуск (restart) контейнера или проекта Compose
  %s   Принудительное обновление всех данных
  %s   Показать / скрыть эту справку
  %s   Выход из приложения

  Интервал автообновления: %s`,
		boldStyle.Render("ГОРЯЧИЕ КЛАВИШИ:"),
		greenStyle.Render("[1] - [4]"),
		greenStyle.Render("[Tab] / [Shift+Tab] / [←] / [→]"),
		greenStyle.Render("[↑] / [↓] / [k] / [j]"),
		greenStyle.Render("[/]"),
		greenStyle.Render("[L]"),
		greenStyle.Render("[Enter] / [I]"),
		greenStyle.Render("[S]"),
		greenStyle.Render("[P]"),
		greenStyle.Render("[R]"),
		greenStyle.Render("[?]"),
		greenStyle.Render("[Q] / [Esc]"),
		yellowStyle.Render(m.refreshInterval.String()),
	)
	return boxStyle.Render(content)
}

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
	s.WriteString(fmt.Sprintf("  CPU (%d cores): %s\n",
		m.cpuInfo.Cores, formatBar(m.cpuInfo.Percent, 28),
	))
	s.WriteString(fmt.Sprintf("  RAM (%.1f/%.1f GB): %s\n\n",
		m.memInfo.UsedGB, m.memInfo.TotalGB, formatBar(m.memInfo.Percent, 28),
	))

	s.WriteString(titleStyle.Render("СЕТЬ И ТРАФИК") + "\n")
	s.WriteString(fmt.Sprintf("  IP: %s    DNS: %s\n",
		boldStyle.Render(m.netInfo.IP), m.netInfo.DNS,
	))
	s.WriteString(fmt.Sprintf("  Скорость: %s    %s\n\n",
		greenStyle.Render("↓ RX: "+m.netInfo.RXSpeed),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Render("↑ TX: "+m.netInfo.TXSpeed),
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

func (m Model) renderDockerTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("DOCKER КОНТЕЙНЕРЫ (ИНТЕРАКТИВНОЕ УПРАВЛЕНИЕ)") + "\n")
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

	containers := m.getFilteredContainers()

	if m.isFiltering {
		s.WriteString("  " + yellowStyle.Render(fmt.Sprintf("[Поиск: %s█]", m.filterQuery)) + "  " + dimStyle.Render("(Enter: применить, Esc: отмена)") + "\n\n")
	} else if m.filterQuery != "" {
		s.WriteString("  " + activeTabStyle.Render(fmt.Sprintf("Фильтр: %s", m.filterQuery)) + "  " + dimStyle.Render(fmt.Sprintf("(Найдено: %d из %d, Esc: сбросить)", len(containers), len(m.dockerInfo.Containers))) + "\n\n")
	}

	if len(containers) == 0 && m.filterQuery != "" {
		s.WriteString("  " + dimStyle.Render(fmt.Sprintf("По запросу %q контейнеры не найдены.", m.filterQuery)) + "\n")
		s.WriteString("  " + dimStyle.Render("Нажмите Esc для сброса поискового фильтра.") + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("   %-12s %-18s %-18s %-8s %-15s %-9s %-11s %-18s\n",
		dimStyle.Render("ID"), dimStyle.Render("NAME"), dimStyle.Render("IMAGE"),
		dimStyle.Render("CPU%"), dimStyle.Render("MEM USAGE"),
		dimStyle.Render("STATE"), dimStyle.Render("HEALTH"), dimStyle.Render("PORTS"),
	))
	s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 114)) + "\n")

	for i, c := range containers {
		stateFormatted := fmt.Sprintf("%-9s", c.State)
		var stateColored string
		if c.State == "running" {
			stateColored = greenStyle.Render(stateFormatted)
		} else if c.State == "paused" {
			stateColored = yellowStyle.Render(stateFormatted)
		} else {
			stateColored = dimStyle.Render(stateFormatted)
		}

		healthText := "-"
		switch c.Health {
		case "healthy":
			healthText = "[healthy]"
		case "unhealthy":
			healthText = "[unhealthy]"
		case "starting":
			healthText = "[starting]"
		}

		healthFormatted := fmt.Sprintf("%-11s", healthText)
		var healthColored string
		switch c.Health {
		case "healthy":
			healthColored = greenStyle.Render(healthFormatted)
		case "unhealthy":
			healthColored = redStyle.Render(healthFormatted)
		case "starting":
			healthColored = yellowStyle.Render(healthFormatted)
		default:
			healthColored = dimStyle.Render(healthFormatted)
		}

		cpuStr := "-"
		memStr := "-"
		if c.State == "running" {
			cpuStr = fmt.Sprintf("%.1f%%", c.CPUPerc)
			if c.MemUsageMB > 0 {
				if c.MemPerc > 0 {
					memStr = fmt.Sprintf("%.1fM (%.1f%%)", c.MemUsageMB, c.MemPerc)
				} else {
					memStr = fmt.Sprintf("%.1f MB", c.MemUsageMB)
				}
			}
		}

		cpuFormatted := fmt.Sprintf("%-8s", cpuStr)
		var cpuColored string
		if c.CPUPerc > 80.0 {
			cpuColored = redStyle.Render(cpuFormatted)
		} else if c.CPUPerc > 50.0 {
			cpuColored = yellowStyle.Render(cpuFormatted)
		} else if c.CPUPerc > 0.0 {
			cpuColored = cpuFormatted
		} else {
			cpuColored = dimStyle.Render(cpuFormatted)
		}

		memFormatted := fmt.Sprintf("%-15s", truncate(memStr, 15))
		var memColored string
		if c.MemPerc > 85.0 {
			memColored = redStyle.Render(memFormatted)
		} else if c.MemPerc > 70.0 {
			memColored = yellowStyle.Render(memFormatted)
		} else if c.MemUsageMB > 0 {
			memColored = memFormatted
		} else {
			memColored = dimStyle.Render(memFormatted)
		}

		prefix := "  "
		rowContent := fmt.Sprintf("%-12s %-18s %-18s %s %s %s %s %s",
			c.ID, boldStyle.Render(truncate(c.Name, 16)), truncate(c.Image, 16),
			cpuColored, memColored,
			stateColored, healthColored, truncate(c.Ports, 18),
		)

		if i == m.cursorIdx {
			prefix = greenStyle.Render("► ")
			rowContent = selectedRowStyle.Render(fmt.Sprintf("%-12s %-18s %-18s %-8s %-15s %-9s %-11s %-18s",
				c.ID, truncate(c.Name, 16), truncate(c.Image, 16),
				cpuStr, truncate(memStr, 15),
				c.State, healthText, truncate(c.Ports, 18),
			))
		}

		s.WriteString(fmt.Sprintf("%s%s\n", prefix, rowContent))
	}

	return s.String()
}

func (m Model) renderComposeTab() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("DOCKER COMPOSE ПРОЕКТЫ (%d)", len(m.compose))) + "\n\n")

	if len(m.compose) == 0 {
		s.WriteString("  " + dimStyle.Render("Проекты Docker Compose не обнаружены (нет меток com.docker.compose.project).") + "\n")
		return s.String()
	}

	for i, p := range m.compose {
		prefix := "  "
		headerText := fmt.Sprintf("┌─ %s  [%s]  (Сервисов поднято: %d/%d)",
			boldStyle.Render(p.Name), p.StatusDisplay, p.RunningServices, p.TotalServices,
		)
		if i == m.composeCursorIdx {
			prefix = greenStyle.Render("► ")
			headerText = selectedRowStyle.Render(fmt.Sprintf("┌─ %s  [%s]  (Сервисов поднято: %d/%d)",
				p.Name, p.StatusDisplay, p.RunningServices, p.TotalServices,
			))
		}
		s.WriteString(fmt.Sprintf("%s%s\n", prefix, headerText))
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

	s.WriteString(titleStyle.Render("DOCKER SWARM КЛАСТЕР") + "\n\n")

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
		s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8s %-10s\n",
			dimStyle.Render("NODE ID"), dimStyle.Render("HOSTNAME"), dimStyle.Render("ROLE"),
			dimStyle.Render("STATE"), dimStyle.Render("AVAILABILITY"), dimStyle.Render("TASKS"), dimStyle.Render("LEADER"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 88)) + "\n")
		for _, n := range m.swarm.Nodes {
			leaderMark := ""
			if n.IsLeader {
				leaderMark = yellowStyle.Render("[Leader]")
			}
			avail := n.Availability
			if avail == "" {
				avail = "active"
			}
			s.WriteString(fmt.Sprintf("  %-14s %-20s %-12s %-10s %-14s %-8d %-10s\n",
				n.ID, boldStyle.Render(truncate(n.Hostname, 18)), n.Role, n.State, avail, n.ActiveTasks, leaderMark,
			))
		}
		s.WriteString("\n")
	}

	if len(m.swarm.Services) > 0 {
		s.WriteString(boldStyle.Render("  СЛУЖБЫ SWARM (SERVICES):") + "\n")
		s.WriteString(fmt.Sprintf("   %-14s %-20s %-24s %-16s %-18s\n",
			dimStyle.Render("SERVICE ID"), dimStyle.Render("NAME"), dimStyle.Render("IMAGE"),
			dimStyle.Render("REPLICAS"), dimStyle.Render("PORTS"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 94)) + "\n")
		for i, svc := range m.swarm.Services {
			ports := svc.Ports
			if ports == "" {
				ports = "-"
			}
			repColored := greenStyle.Render(svc.Replicas)
			if svc.Mode == "replicated" && svc.TargetReplicas > 0 && svc.RunningReplicas < svc.TargetReplicas {
				if svc.RunningReplicas == 0 {
					repColored = redStyle.Render(svc.Replicas)
				} else {
					repColored = yellowStyle.Render(svc.Replicas)
				}
			}

			prefix := "  "
			rowContent := fmt.Sprintf("%-14s %-20s %-24s %-16s %-18s",
				svc.ID, boldStyle.Render(truncate(svc.Name, 18)), truncate(svc.Image, 22), repColored, truncate(ports, 18),
			)
			if i == m.swarmCursorIdx {
				prefix = greenStyle.Render("► ")
				rowContent = selectedRowStyle.Render(fmt.Sprintf("%-14s %-20s %-24s %-16s %-18s",
					svc.ID, truncate(svc.Name, 18), truncate(svc.Image, 22), svc.Replicas, truncate(ports, 18),
				))
			}

			s.WriteString(fmt.Sprintf("%s%s\n", prefix, rowContent))
		}
	}

	return s.String()
}

func formatPercent(val float64) string {
	str := fmt.Sprintf("%.1f%%", val)
	if val >= 80 {
		return redStyle.Render(str)
	} else if val >= 50 {
		return yellowStyle.Render(str)
	}
	return greenStyle.Render(str)
}

func formatBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	barFilled := strings.Repeat("█", filled)
	barEmpty := strings.Repeat("░", empty)

	var coloredFilled string
	if percent >= 85 {
		coloredFilled = redStyle.Render(barFilled)
	} else if percent >= 60 {
		coloredFilled = yellowStyle.Render(barFilled)
	} else {
		coloredFilled = greenStyle.Render(barFilled)
	}

	pctStr := fmt.Sprintf("%5.1f%%", percent)
	return fmt.Sprintf("[%s%s] %s", coloredFilled, dimStyle.Render(barEmpty), pctStr)
}

func truncate(str string, maxLen int) string {
	runes := []rune(str)
	if len(runes) <= maxLen {
		return str
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
