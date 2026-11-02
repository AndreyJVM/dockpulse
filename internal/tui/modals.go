package tui

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/monitor"
)

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

func (m Model) renderPruneModal() string {
	var s strings.Builder

	title := "ОЧИСТКА РЕСУРСОВ DOCKER (SYSTEM PRUNE)"
	s.WriteString(titleStyle.Render(title) + "\n\n")

	if m.pruneUsageErr != "" {
		s.WriteString(fmt.Sprintf("  %s %s\n\n", dimStyle.Render("Статус:"), redStyle.Render(m.pruneUsageErr)))
	} else {
		s.WriteString("  " + boldStyle.Render("ТЕКУЩЕЕ ИСПОЛЬЗОВАНИЕ ДИСКА:") + "\n")
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			dimStyle.Render("ТИП"), dimStyle.Render("ОБЪЕМ"), dimStyle.Render("АКТИВНО/ВСЕГО"), dimStyle.Render("ОСВОБОЖДАЕМО"),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 58)) + "\n")

		u := m.pruneUsage
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			"Образы", monitor.FormatBytes(u.Images.TotalSize),
			fmt.Sprintf("%d / %d", u.Images.ActiveCount, u.Images.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.Images.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			"Контейнеры", monitor.FormatBytes(u.Containers.TotalSize),
			fmt.Sprintf("%d / %d", u.Containers.ActiveCount, u.Containers.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.Containers.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			"Тома", monitor.FormatBytes(u.Volumes.TotalSize),
			fmt.Sprintf("%d / %d", u.Volumes.ActiveCount, u.Volumes.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.Volumes.Reclaimable)),
		))
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n",
			"Кэш сборки", monitor.FormatBytes(u.BuildCache.TotalSize),
			fmt.Sprintf("%d / %d", u.BuildCache.ActiveCount, u.BuildCache.TotalCount),
			yellowStyle.Render(monitor.FormatBytes(u.BuildCache.Reclaimable)),
		))
		s.WriteString("  " + dimStyle.Render(strings.Repeat("─", 58)) + "\n")
		s.WriteString(fmt.Sprintf("  %-16s %-12s %-14s %s\n\n",
			boldStyle.Render("Всего"), boldStyle.Render(monitor.FormatBytes(u.TotalSize)),
			"", greenStyle.Render(monitor.FormatBytes(u.TotalRecl)),
		))
	}

	if m.pruneConfirm > 0 {
		var actionName string
		switch m.pruneConfirm {
		case 1:
			actionName = "Удаление всех остановленных контейнеров"
		case 2:
			actionName = "Удаление всех висячих образов (dangling images)"
		case 3:
			actionName = "Удаление всех неиспользуемых томов (volumes)"
		case 4:
			actionName = "ПОЛНАЯ очистка: остановленные контейнеры, образы, тома и кэш сборки"
		}

		s.WriteString("  " + redStyle.Render("[ВНИМАНИЕ] ") + boldStyle.Render(actionName) + "\n")
		s.WriteString("  Это действие необратимо. Подтвердить выполнение?\n\n")
		s.WriteString(fmt.Sprintf("  %s Подтвердить  |  %s Отмена\n",
			greenStyle.Render("[Y]"),
			dimStyle.Render("[N / Esc]"),
		))
	} else {
		s.WriteString("  " + boldStyle.Render("ВЫБЕРИТЕ ДЕЙСТВИЕ ДЛЯ ОЧИСТКИ:") + "\n")
		s.WriteString(fmt.Sprintf("  %s Очистить остановленные контейнеры (Prune stopped containers)\n", greenStyle.Render("[1]")))
		s.WriteString(fmt.Sprintf("  %s Очистить висячие образы без тегов (Prune dangling images)\n", greenStyle.Render("[2]")))
		s.WriteString(fmt.Sprintf("  %s Очистить неиспользуемые тома данных (Prune unused volumes)\n", greenStyle.Render("[3]")))
		s.WriteString(fmt.Sprintf("  %s Полная очистка всех неиспользуемых ресурсов (Prune all)\n\n", yellowStyle.Render("[4]")))
		s.WriteString(fmt.Sprintf("  %s Закрыть окно\n", dimStyle.Render("[Esc / Q]")))
	}

	return boxStyle.Render(s.String())
}

func (m Model) renderComposeDownModal() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("ПОДТВЕРЖДЕНИЕ СВОРАЧИВАНИЯ COMPOSE СТЕКА") + "\n\n")
	s.WriteString(fmt.Sprintf("  Стек: %s\n\n", boldStyle.Render(m.composeDownTarget)))
	s.WriteString(fmt.Sprintf("  %s Остановка и удаление всех контейнеров проекта (docker compose down)\n\n", yellowStyle.Render("Внимание:"))))
	s.WriteString(fmt.Sprintf("  Вы уверены, что хотите выполнить down для проекта %s?\n\n", boldStyle.Render(m.composeDownTarget)))
	s.WriteString(fmt.Sprintf("  %s Подтвердить (Down)    %s Отмена\n", greenStyle.Render("[Y]"), dimStyle.Render("[N / Esc]")))
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
  %s   Интерактивный shell в контейнере (bash / sh)
  %s   Запуск / остановка контейнера, остановка Compose проекта
  %s   Запуск остановленного проекта Compose (Start)
  %s   Перезапуск (restart) контейнера или проекта Compose
  %s   Сворачивание и удаление проекта Compose (Down)
  %s   Очистка неиспользуемых ресурсов Docker (Prune)
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
		greenStyle.Render("[E]"),
		greenStyle.Render("[S]"),
		greenStyle.Render("[U]"),
		greenStyle.Render("[P]"),
		greenStyle.Render("[D]"),
		greenStyle.Render("[C]"),
		greenStyle.Render("[R]"),
		greenStyle.Render("[?]"),
		greenStyle.Render("[Q] / [Esc]"),
		yellowStyle.Render(m.refreshInterval.String()),
	)
	return boxStyle.Render(content)
}
