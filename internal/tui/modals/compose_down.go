package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderComposeDown renders the compose stack down confirmation modal.
func RenderComposeDown(target string, dict *i18n.Dictionary) string {
	var s strings.Builder
	s.WriteString(theme.TitleStyle.Render(dict.ComposeDownTitle) + "\n\n")
	s.WriteString(fmt.Sprintf("  "+dict.StackLabel+"\n\n", theme.BoldStyle.Render(target)))
	s.WriteString(fmt.Sprintf("  %s %s\n\n", theme.YellowStyle.Render(dict.WarningLabel), dict.ComposeDownNotice))
	s.WriteString(fmt.Sprintf("  "+dict.ComposeDownConfirm+"\n\n", theme.BoldStyle.Render(target)))
	s.WriteString(fmt.Sprintf("  %s %s (Down)    %s %s\n", theme.GreenStyle.Render("[Y]"), dict.FooterConfirm, theme.DimStyle.Render("[N / Esc]"), dict.FooterCancel))
	return theme.BoxStyle.Render(s.String())
}
