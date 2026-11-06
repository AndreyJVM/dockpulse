package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AndreyJVM/dockpulse/internal/i18n"
	"github.com/AndreyJVM/dockpulse/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

var (
	version = "v0.7.1"
)

const (
	defaultInterval = 2 * time.Second
	defaultTab      = 1
	minInterval     = 100 * time.Millisecond
)

type config struct {
	interval    time.Duration
	initialTab  int
	lang        string
	showVersion bool
}

var (
	runProgram = func(model tea.Model) error {
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}
	isTerminal = func(fd uintptr) bool {
		return term.IsTerminal(fd)
	}
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Dockpulse %s\n\n", version)
	fmt.Fprintf(w, "Использование:\n")
	fmt.Fprintf(w, "  dockpulse [опции]\n\n")
	fmt.Fprintf(w, "Опции:\n")
	fmt.Fprintf(w, "  -i, --interval <duration>  Интервал автообновления (по умолчанию: %v, env: DOCKPULSE_INTERVAL)\n", defaultInterval)
	fmt.Fprintf(w, "  -t, --tab <1-4|name>       Начальная вкладка: 1|host, 2|docker, 3|compose, 4|swarm (по умолчанию: %d, env: DOCKPULSE_TAB)\n", defaultTab)
	fmt.Fprintf(w, "  -l, --lang <en|ru>         Язык интерфейса: en или ru (по умолчанию: автоопределение, env: DOCKPULSE_LANG)\n")
	fmt.Fprintf(w, "  -v, --version              Показать версию\n")
	fmt.Fprintf(w, "  -h, --help                 Показать эту справку\n\n")
	fmt.Fprintf(w, "Управление в TUI:\n")
	fmt.Fprintf(w, "  1..4, Tab, ←/→             Переключение вкладок\n")
	fmt.Fprintf(w, "  r, R                       Принудительное обновление данных\n")
	fmt.Fprintf(w, "  ?                          Окно подсказок горячих клавиш\n")
	fmt.Fprintf(w, "  q, Esc, Ctrl+C             Выход из приложения\n")
}

func parseFlags(args []string, stdout, stderr io.Writer) (*config, error) {
	intervalDefault := defaultInterval
	if envInterval := os.Getenv("DOCKPULSE_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			intervalDefault = d
		}
	}

	tabDefault := "1"
	if envTab := os.Getenv("DOCKPULSE_TAB"); envTab != "" {
		tabDefault = envTab
	}

	langDefault := os.Getenv("DOCKPULSE_LANG")

	var tabFlag string
	cfg := &config{
		interval:   intervalDefault,
		initialTab: defaultTab,
	}

	flags := flag.NewFlagSet("dockpulse", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		printUsage(stdout)
	}

	flags.DurationVar(&cfg.interval, "interval", intervalDefault, "Интервал обновления метрик (например, 1s, 2s, 500ms)")
	flags.DurationVar(&cfg.interval, "i", intervalDefault, "Интервал обновления (короткий флаг)")
	flags.StringVar(&tabFlag, "tab", tabDefault, "Начальная вкладка: 1|host, 2|docker, 3|compose, 4|swarm")
	flags.StringVar(&tabFlag, "t", tabDefault, "Начальная вкладка (короткий флаг)")
	flags.StringVar(&cfg.lang, "lang", langDefault, "Язык интерфейса: en или ru (по умолчанию: автоопределение)")
	flags.StringVar(&cfg.lang, "l", langDefault, "Язык интерфейса (короткий флаг)")
	flags.BoolVar(&cfg.showVersion, "version", false, "Показать версию программы")
	flags.BoolVar(&cfg.showVersion, "v", false, "Показать версию (короткий флаг)")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if len(flags.Args()) > 0 {
		return nil, fmt.Errorf("неизвестные позиционные аргументы: %s", strings.Join(flags.Args(), " "))
	}

	tab, err := parseTab(tabFlag)
	if err != nil {
		return nil, err
	}
	cfg.initialTab = tab

	return cfg, nil
}

func parseTab(val string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "host":
		return 1, nil
	case "2", "docker":
		return 2, nil
	case "3", "compose":
		return 3, nil
	case "4", "swarm":
		return 4, nil
	default:
		return 0, fmt.Errorf("недопустимая начальная вкладка: %q (допустимы: 1|host, 2|docker, 3|compose, 4|swarm)", val)
	}
}

func validateConfig(cfg *config) error {
	if cfg.interval < minInterval {
		return fmt.Errorf("интервал обновления должен быть не менее %v (передано: %v)", minInterval, cfg.interval)
	}
	if cfg.initialTab < 1 || cfg.initialTab > 4 {
		return fmt.Errorf("начальная вкладка должна быть в диапазоне от 1 до 4 (передано: %d)", cfg.initialTab)
	}
	if cfg.lang != "" && cfg.lang != "en" && cfg.lang != "ru" {
		return fmt.Errorf("неподдерживаемый язык: %s (допустимы: en, ru)", cfg.lang)
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
		fmt.Fprintf(stdout, "dockpulse %s\n", version)
		return nil
	}

	if err := validateConfig(cfg); err != nil {
		return err
	}

	if !isTerminal(os.Stdin.Fd()) || !isTerminal(os.Stdout.Fd()) {
		return errors.New("dockpulse требует интерактивного терминала (TTY)")
	}

	if cfg.lang == "" {
		cfg.lang = i18n.DetectLanguage()
	}

	model := tui.NewModel(cfg.interval, cfg.initialTab, version, cfg.lang)
	return runProgram(model)
}
