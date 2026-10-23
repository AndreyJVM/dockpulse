package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AndreyJVM/system-monitor/internal/tui"
)

var (
	version = "v0.3.0"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("system-monitor %s (Go Edition)\n", version)
		return
	}

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running system-monitor: %v\n", err)
		os.Exit(1)
	}
}
