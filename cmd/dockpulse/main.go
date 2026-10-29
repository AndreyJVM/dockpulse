package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "v0.5.1"
)

const (
	defaultInterval = 2 * time.Second
	defaultTab      = 1
	minInterval     = 100 * time.Millisecond
)

type config struct {
	interval    time.Duration
	initialTab  int
	showVersion bool
}

var runProgram = func(model tea.Model) error {
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Dockpulse %s (Go Edition)\n\n", version)
	fmt.Fprintf(w, "Использование:\n")
	fmt.Fprintf(w, "  dockpulse [опции]\n\n")
	fmt.Fprintf(w, "Опции:\n")
	fmt.Fprintf(w, "  -i, --interval <duration>  Интервал автообновления (по умолчанию: %v)\n", defaultInterval)
	fmt.Fprintf(w, "  -t, --tab <1-4>            Начальная вкладка (по умолчанию: %d)\n", defaultTab)
	fmt.Fprintf(w, "  -v, --version              Показать версию\n")
	fmt.Fprintf(w, "  -h, --help                 Показать эту справку\n\n")
	fmt.Fprintf(w, "Управление в TUI:\n")
	fmt.Fprintf(w, "  1..4, Tab, ←/→             Переключение вкладок\n")
	fmt.Fprintf(w, "  r, R                       Принудительное обновление данных\n")
	fmt.Fprintf(w, "  ?                          Окно подсказок горячих клавиш\n")
	fmt.Fprintf(w, "  q, Esc, Ctrl+C             Выход из приложения\n")
}

func parseFlags(args []string, stdout, stderr io.Writer) (*config, error) {
	cfg := &config{
		interval:   defaultInterval,
		initialTab: defaultTab,
	}

	flags := flag.NewFlagSet("dockpulse", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		printUsage(stdout)
	}

	flags.DurationVar(&cfg.interval, "interval", defaultInterval, "Интервал обновления метрик (например, 1s, 2s, 500ms)")
	flags.DurationVar(&cfg.interval, "i", defaultInterval, "Интервал обновления (короткий флаг)")
	flags.IntVar(&cfg.initialTab, "tab", defaultTab, "Начальная вкладка (1: Host, 2: Docker, 3: Compose, 4: Swarm)")
	flags.IntVar(&cfg.initialTab, "t", defaultTab, "Начальная вкладка (короткий флаг)")
	flags.BoolVar(&cfg.showVersion, "version", false, "Показать версию программы")
	flags.BoolVar(&cfg.showVersion, "v", false, "Показать версию (короткий флаг)")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if len(flags.Args()) > 0 {
		return nil, fmt.Errorf("неизвестные позиционные аргументы: %s", strings.Join(flags.Args(), " "))
	}

	return cfg, nil
}

func validateConfig(cfg *config) error {
	if cfg.interval < minInterval {
		return fmt.Errorf("интервал обновления должен быть не менее %v (передано: %v)", minInterval, cfg.interval)
	}
	if cfg.initialTab < 1 || cfg.initialTab > 4 {
		return fmt.Errorf("начальная вкладка должна быть в диапазоне от 1 до 4 (передано: %d)", cfg.initialTab)
	}
	return nil
}

func run(args []string, stdout, stderr io.Writer) error {
	cfg, err := parseFlags(args, stdout, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if cfg.showVersion {
		fmt.Fprintf(stdout, "dockpulse %s (Go Edition)\n", version)
		return nil
	}

	if err := validateConfig(cfg); err != nil {
		return err
	}

	model := tui.NewModel(cfg.interval, cfg.initialTab, version)
	return runProgram(model)
}
