package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/AndreyJVM/dockpulse/internal/tui"
)

var (
	version = "v0.4.1"
)

func main() {
	var (
		interval     time.Duration
		initialTab   int
		showVersion  bool
		showHelp     bool
	)

	flag.DurationVar(&interval, "interval", 2*time.Second, "Интервал обновления метрик (например, 1s, 2s, 500ms)")
	flag.DurationVar(&interval, "i", 2*time.Second, "Интервал обновления (короткий флаг)")
	flag.IntVar(&initialTab, "tab", 1, "Начальная вкладка (1: Host, 2: Docker, 3: Compose, 4: Swarm)")
	flag.IntVar(&initialTab, "t", 1, "Начальная вкладка (короткий флаг)")
	flag.BoolVar(&showVersion, "version", false, "Показать версию программы")
	flag.BoolVar(&showVersion, "v", false, "Показать версию (короткий флаг)")
	flag.BoolVar(&showHelp, "help", false, "Показать справку")
	flag.BoolVar(&showHelp, "h", false, "Показать справку (короткий флаг)")

	flag.Usage = func() {
		fmt.Printf("Dockpulse %s (Go Edition)\n\n", version)
		fmt.Printf("Использование:\n")
		fmt.Printf("  dockpulse [опции]\n\n")
		fmt.Printf("Опции:\n")
		fmt.Printf("  -i, --interval <duration>  Интервал автообновления (по умолчанию: 2s)\n")
		fmt.Printf("  -t, --tab <1-4>            Начальная вкладка (по умолчанию: 1)\n")
		fmt.Printf("  -v, --version              Показать версию\n")
		fmt.Printf("  -h, --help                 Показать эту справку\n\n")
		fmt.Printf("Управление в TUI:\n")
		fmt.Printf("  1..4, Tab, ←/→             Переключение вкладок\n")
		fmt.Printf("  r, R                       Принудительное обновление данных\n")
		fmt.Printf("  ?                          Окно подсказок горячих клавиш\n")
		fmt.Printf("  q, Esc, Ctrl+C             Выход из приложения\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("dockpulse %s (Go Edition)\n", version)
		return
	}

	if showHelp {
		flag.Usage()
		return
	}

	model := tui.NewModel(interval, initialTab, version)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка запуска dockpulse: %v\n", err)
		os.Exit(1)
	}
}
