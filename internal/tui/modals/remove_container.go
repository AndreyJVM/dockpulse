package modals

import (
	"fmt"
	"strings"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui/theme"
)

// RenderRemoveContainer renders the container deletion confirmation modal.
func RenderRemoveContainer(containerName, containerID string, dict *i18n.Dictionary) string {
	var s strings.Builder
	s.WriteString(theme.TitleStyle.Render(dict.RemoveContainerTitle) + "\n\n")
	label := fmt.Sprintf("%s (%s)", containerName, containerID)
	s.WriteString(fmt.Sprintf("  "+dict.ContainerLabel+"\n\n", theme.BoldStyle.Render(label)))
	s.WriteString(fmt.Sprintf("  %s %s\n\n", theme.YellowStyle.Render(dict.WarningLabel), dict.RemoveContainerNotice))
	s.WriteString(fmt.Sprintf("  "+dict.RemoveContainerConfirm+"\n\n", theme.BoldStyle.Render(containerName)))
	s.WriteString(fmt.Sprintf("  %s %s (Remove)    %s %s\n", theme.RedStyle.Render("[Y]"), dict.FooterConfirm, theme.DimStyle.Render("[N / Esc]"), dict.FooterCancel))
	return theme.BoxStyle.Render(s.String())
}
