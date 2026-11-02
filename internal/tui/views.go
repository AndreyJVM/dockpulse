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
		m.i18n.TabHost,
		m.i18n.TabDocker,
		m.i18n.TabCompose,
		m.i18n.TabSwarm,
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
				boldStyle.Render("[Y]") + " " + m.i18n.FooterConfirm + " (Down)  |  " +
					boldStyle.Render("[N / Esc]") + " " + m.i18n.FooterCancel,
			))
		} else if m.activeModal == "prune" {
			if m.pruneConfirm > 0 {
				b.WriteString(footerStyle.Render(
					boldStyle.Render("[Y]") + " " + m.i18n.FooterConfirmPrune + "  |  " +
						boldStyle.Render("[N / Esc]") + " " + m.i18n.FooterCancel,
				))
			} else {
				b.WriteString(footerStyle.Render(
					boldStyle.Render("[1-4]") + " " + m.i18n.FooterSelectAction + "  |  " +
						boldStyle.Render("[Esc / Q]") + " " + m.i18n.FooterClose,
				))
			}
		} else {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[Esc / Q]") + " " + m.i18n.FooterClose + "  |  " +
					boldStyle.Render("[↑ / ↓ / PgUp / PgDn]") + " " + m.i18n.FooterScroll,
			))
		}
	} else if m.activeTab == 2 {
		if m.isFiltering {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[Enter]") + " " + m.i18n.FooterApplyFilter + "  |  " +
					boldStyle.Render("[Esc]") + " " + m.i18n.FooterCancel,
			))
		} else if m.filterQuery != "" {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[↑ / ↓]") + " " + m.i18n.FooterSelect + "  |  " +
					boldStyle.Render("[Esc]") + " " + m.i18n.FooterResetFilter + "  |  " +
					boldStyle.Render("[L]") + " " + m.i18n.FooterLogs + "  |  " +
					boldStyle.Render("[Enter/I]") + " " + m.i18n.FooterInfo + "  |  " +
					boldStyle.Render("[E]") + " " + m.i18n.FooterShell + "  |  " +
					boldStyle.Render("[S]") + " " + m.i18n.FooterStartStop + "  |  " +
					boldStyle.Render("[P]") + " " + m.i18n.FooterRestart,
			))
		} else {
			b.WriteString(footerStyle.Render(
				boldStyle.Render("[↑ / ↓]") + " " + m.i18n.FooterSelect + "  |  " +
					boldStyle.Render("[/]") + " " + m.i18n.FooterSearch + "  |  " +
					boldStyle.Render("[L]") + " " + m.i18n.FooterLogs + "  |  " +
					boldStyle.Render("[Enter/I]") + " " + m.i18n.FooterInfo + "  |  " +
					boldStyle.Render("[E]") + " " + m.i18n.FooterShell + "  |  " +
					boldStyle.Render("[S]") + " " + m.i18n.FooterStartStop + "  |  " +
					boldStyle.Render("[P]") + " " + m.i18n.FooterRestart + "  |  " +
					boldStyle.Render("[C]") + " " + m.i18n.FooterPrune + "  |  " +
					boldStyle.Render("[?]") + " " + m.i18n.FooterHelp,
			))
		}
	} else if m.activeTab == 3 {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[↑ / ↓]") + " " + m.i18n.FooterSelectStack + "  |  " +
				boldStyle.Render("[L]") + " " + m.i18n.FooterLogs + "  |  " +
				boldStyle.Render("[S]") + " " + m.i18n.FooterStop + "  |  " +
				boldStyle.Render("[U]") + " " + m.i18n.FooterStart + "  |  " +
				boldStyle.Render("[P]") + " " + m.i18n.FooterRestart + "  |  " +
				boldStyle.Render("[D]") + " " + m.i18n.FooterDown + "  |  " +
				boldStyle.Render("[C]") + " " + m.i18n.FooterPrune + "  |  " +
				boldStyle.Render("[?]") + " " + m.i18n.FooterHelp,
		))
	} else if m.activeTab == 4 {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[↑ / ↓]") + " " + m.i18n.FooterSelectService + "  |  " +
				boldStyle.Render("[L]") + " " + m.i18n.FooterServiceLogs + "  |  " +
				boldStyle.Render("[?]") + " " + m.i18n.FooterHelp,
		))
	} else {
		b.WriteString(footerStyle.Render(
			boldStyle.Render("[1-4 / Tab / ← →]") + " " + m.i18n.FooterTabs + "  |  " +
				boldStyle.Render("[R]") + " " + m.i18n.FooterRefresh + "  |  " +
				boldStyle.Render("[?]") + " " + m.i18n.FooterHelp + "  |  " +
				boldStyle.Render("[Q]") + " " + m.i18n.FooterQuit,
		))
	}

	return b.String()
}
