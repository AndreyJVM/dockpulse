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
	"github.com/charmbracelet/x/term"
)

var (
	version = "v0.9.0"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Dockpulse %s\n\n", version)
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  dockpulse [options]\n\n")
	fmt.Fprintf(w, "Options:\n")
	fmt.Fprintf(w, "  -i, --interval <duration>  Auto-refresh interval (default: %v, env: DOCKPULSE_INTERVAL)\n", defaultInterval)
	fmt.Fprintf(w, "  -t, --tab <1-5|name>       Initial tab: 1|docker, 2|compose, 3|networks, 4|volumes, 5|swarm (default: %d, env: DOCKPULSE_TAB)\n", defaultTab)
	fmt.Fprintf(w, "  -v, --version              Show version\n")
	fmt.Fprintf(w, "  -h, --help                 Show this help\n\n")
	fmt.Fprintf(w, "TUI controls:\n")
	fmt.Fprintf(w, "  1..5, Tab, ←/→             Switch tabs\n")
	fmt.Fprintf(w, "  r, R                       Refresh data immediately\n")
	fmt.Fprintf(w, "  ?                          Show keyboard shortcuts\n")
	fmt.Fprintf(w, "  q, Esc, Ctrl+C             Quit\n")
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

	flags.DurationVar(&cfg.interval, "interval", intervalDefault, "Metrics refresh interval (for example: 1s, 2s, 500ms)")
	flags.DurationVar(&cfg.interval, "i", intervalDefault, "Metrics refresh interval (short flag)")
	flags.StringVar(&tabFlag, "tab", tabDefault, "Initial tab: 1|docker, 2|compose, 3|networks, 4|volumes, 5|swarm")
	flags.StringVar(&tabFlag, "t", tabDefault, "Initial tab (short flag)")
	flags.BoolVar(&cfg.showVersion, "version", false, "Show version")
	flags.BoolVar(&cfg.showVersion, "v", false, "Show version (short flag)")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if len(flags.Args()) > 0 {
		return nil, fmt.Errorf("unknown positional arguments: %s", strings.Join(flags.Args(), " "))
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
	case "1", "docker":
		return 1, nil
	case "2", "compose":
		return 2, nil
	case "3", "networks", "network":
		return 3, nil
	case "4", "volumes", "volume":
		return 4, nil
	case "5", "swarm":
		return 5, nil
	default:
		return 0, fmt.Errorf("invalid initial tab: %q (allowed: 1|docker, 2|compose, 3|networks, 4|volumes, 5|swarm)", val)
	}
}

func validateConfig(cfg *config) error {
	if cfg.interval < minInterval {
		return fmt.Errorf("refresh interval must be at least %v (got: %v)", minInterval, cfg.interval)
	}
	if cfg.initialTab < 1 || cfg.initialTab > 5 {
		return fmt.Errorf("initial tab must be between 1 and 5 (got: %d)", cfg.initialTab)
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
		return errors.New("dockpulse requires an interactive terminal (TTY)")
	}

	model := tui.NewModel(cfg.interval, cfg.initialTab, version)
	return runProgram(model)
}
