package tui

import (
	"fmt"
	"strings"
	"time"
)

func (m Model) View() string {
	var b strings.Builder

	// 1. Панель заголовка
	verStr := m.version
	if verStr == "" {
		verStr = "v0.4.0"
	}
	header := headerStyle.Render("Dockpulse") +
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
	} else if m.activeModal == "compose_down" {
		b.WriteString(m.renderComposeDownModal())
	} else if m.activeModal == "prune" {
		b.WriteString(m.renderPruneModal())
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
		if m.activeModal == "compose_down" {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[Y]") + " Подтвердить (Down)  |  " +
					boldStyle.Render("[N / Esc]") + " Отмена",
			))
		} else if m.activeModal == "prune" {
			if m.pruneConfirm > 0 {
				b.WriteString(footerStyle.Render(
					boldStyle.Render("[Y]") + " Подтвердить очистку  |  " +
						boldStyle.Render("[N / Esc]") + " Отмена",
				))
			} else {
				b.WriteString(footerStyle.Render(
					boldStyle.Render("[1-4]") + " Выбрать действие  |  " +
						boldStyle.Render("[Esc / Q]") + " Закрыть окно",
				))
			}
		} else {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[Esc / Q]") + " Закрыть окно  |  " +
					boldStyle.Render("[↑ / ↓ / PgUp / PgDn]") + " Прокрутка",
			))
		}
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
					boldStyle.Render("[E]") + " Shell  |  " +
					boldStyle.Render("[S]") + " Старт/Стоп  |  " +
					boldStyle.Render("[P]") + " Рестарт",
			))
		} else {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[↑ / ↓]") + " Выбрать  |  " +
					boldStyle.Render("[/]") + " Поиск  |  " +
					boldStyle.Render("[L]") + " Логи  |  " +
					boldStyle.Render("[Enter/I]") + " Инфо  |  " +
					boldStyle.Render("[E]") + " Shell  |  " +
					boldStyle.Render("[S]") + " Старт/Стоп  |  " +
					boldStyle.Render("[P]") + " Рестарт  |  " +
					boldStyle.Render("[C]") + " Очистка  |  " +
					boldStyle.Render("[?]") + " Справка",
			))
		}
	} else if m.activeTab == 3 {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[↑ / ↓]") + " Выбрать стек  |  " +
				boldStyle.Render("[L]") + " Логи  |  " +
				boldStyle.Render("[S]") + " Стоп  |  " +
				boldStyle.Render("[U]") + " Старт  |  " +
				boldStyle.Render("[P]") + " Рестарт  |  " +
				boldStyle.Render("[D]") + " Down  |  " +
				boldStyle.Render("[C]") + " Очистка  |  " +
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
