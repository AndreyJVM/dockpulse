package modals

import (
	"fmt"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderHelp renders the keyboard shortcuts help dialog box.
func RenderHelp(refreshInterval time.Duration, dict *i18n.Dictionary) string {
	content := fmt.Sprintf(`%s

  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s
  %s   %s

  `+dict.HelpAutoInt,
		theme.BoldStyle.Render(dict.HelpTitle),
		theme.GreenStyle.Render("[1] - [4]"), dict.HelpKeyTabs,
		theme.GreenStyle.Render("[Tab] / [Shift+Tab] / [←] / [→]"), dict.HelpKeyPrev,
		theme.GreenStyle.Render("[↑] / [↓] / [k] / [j]"), dict.HelpKeyNav,
		theme.GreenStyle.Render("[/]"), dict.HelpKeySearch,
		theme.GreenStyle.Render("[O]"), dict.HelpKeySort,
		theme.GreenStyle.Render("[L]"), dict.HelpKeyLogs,
		theme.GreenStyle.Render("[Enter] / [I]"), dict.HelpKeyInfo,
		theme.GreenStyle.Render("[E]"), dict.HelpKeyShell,
		theme.GreenStyle.Render("[S]"), dict.HelpKeySS,
		theme.GreenStyle.Render("[U]"), dict.HelpKeyStart,
		theme.GreenStyle.Render("[P]"), dict.HelpKeyRest,
		theme.GreenStyle.Render("[K]"), dict.HelpKeyKill,
		theme.GreenStyle.Render("[x]"), dict.HelpKeyRemove,
		theme.GreenStyle.Render("[D]"), dict.HelpKeyDown,
		theme.GreenStyle.Render("[C]"), dict.HelpKeyPrune,
		theme.GreenStyle.Render("[R]"), dict.HelpKeyRef,
		theme.GreenStyle.Render("[?]"), dict.HelpKeyHelp,
		theme.GreenStyle.Render("[Q] / [Esc]"), dict.HelpKeyQuit,
		theme.YellowStyle.Render(refreshInterval.String()),
	)
	return theme.BoxStyle.Render(content)
}
