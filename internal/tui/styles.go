package tui

import (
	"github.com/charmbracelet/lipgloss"
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

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#38BDF8")).
			Padding(1, 2)
)
