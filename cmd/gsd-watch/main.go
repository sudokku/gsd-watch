package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/config"
	"github.com/radu/gsd-watch/internal/parser"
	"github.com/radu/gsd-watch/internal/tui"
	"github.com/radu/gsd-watch/internal/tui/app"
)

func main() {
	// --help and --debug flags.
	showHelp := flag.Bool("help", false, "Show usage information")
	debugMode := flag.Bool("debug", false, "Print parser decisions to stderr")
	_ = flag.Bool("no-emoji", false, "Use ASCII status icons and badges (for SSH and minimal terminals)")
	themeFlag := flag.String("theme", "", "Color theme name")
	flag.Parse()
	if *showHelp {
		fmt.Println(`gsd-watch — live GSD project status sidebar for tmux and cmux

Start via /gsd-watch slash command in Claude Code.

Keybindings:
  ←/h    collapse        →/l    expand
  ↓/j    move down       ↑/k    move up
  e      expand all      w      collapse all
  ?      help overlay    qq     quit

Flags:
  --help     Show this help
  --debug    Print parser decisions to stderr
  --no-emoji  Use ASCII status icons and badges (for SSH and minimal terminals)
  --theme    Color theme name (overrides config file)

https://github.com/radu/gsd-watch`)
		os.Exit(0)
	}

	if *debugMode {
		parser.DebugOut = os.Stderr
	}

	// Multiplexer detection: require tmux or cmux.
	inTmux := os.Getenv("TMUX") != ""
	inCmux := os.Getenv("CMUX_WORKSPACE_ID") != ""
	if !inTmux && !inCmux {
		installHint := "brew install tmux"
		if runtime.GOOS == "linux" {
			installHint = "sudo apt install tmux"
		}
		fmt.Fprintf(os.Stderr, "gsd-watch requires tmux or cmux.\ntmux:  %s\n       then: tmux new-session\ncmux:  open cmux — gsd-watch will work inside it automatically\n", installHint)
		os.Exit(1)
	}

	// Load config from ~/.config/gsd-watch/config.toml.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	cfgPath := filepath.Join(homeDir, config.ConfigPath)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		var ukErr *config.UnknownKeysError
		if errors.As(err, &ukErr) {
			// CFG-03: warn on stderr, continue with partial config
			for _, k := range ukErr.Keys {
				fmt.Fprintf(os.Stderr, "gsd-watch: unknown config key %q (ignored)\n", k)
			}
		} else {
			// CFG-02: malformed TOML — fatal with path
			fmt.Fprintf(os.Stderr, "gsd-watch: error reading config %s: %v\n", cfgPath, err)
			os.Exit(1)
		}
	}

	// Apply CLI flag overrides (D-05: flag.Visit after flag.Parse).
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "no-emoji":
			cfg.Emoji = false
		case "theme":
			cfg.Preset = *themeFlag
		}
	})

	// THEME-04: validate theme name; warn and fall back to default on unknown names.
	if cfg.Preset != "" {
		if _, ok := tui.ThemeByName(cfg.Preset); !ok {
			fmt.Fprintf(os.Stderr, "gsd-watch: unknown theme %q, using default\n", cfg.Preset)
			cfg.Preset = ""
		}
	}

	// Set tmux pane title for duplicate detection.
	// Deferred reset clears the window-level title on exit so the stale
	// title does not block a future /gsd-watch invocation.
	cwd, _ := os.Getwd()
	fmt.Printf("\033]0;gsd-watch:%s\007", filepath.Base(cwd))
	defer fmt.Printf("\033]0;\007")

	events := make(chan tea.Msg, 10)
	p := tea.NewProgram(
		app.New(events, cfg),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
