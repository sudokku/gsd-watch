package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/parser"
	"github.com/radu/gsd-watch/internal/tui/app"
)

func main() {
	// --help and --debug flags.
	showHelp := flag.Bool("help", false, "Show usage information")
	debugMode := flag.Bool("debug", false, "Print parser decisions to stderr")
	noEmoji := flag.Bool("no-emoji", false, "Use ASCII status icons and badges (for SSH and minimal terminals)")
	flag.Parse()
	if *showHelp {
		fmt.Println(`gsd-watch — live GSD project status sidebar for tmux

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

https://github.com/radu/gsd-watch`)
		os.Exit(0)
	}

	if *debugMode {
		parser.DebugOut = os.Stderr
	}

	// Outside-tmux detection: require tmux session.
	if os.Getenv("TMUX") == "" {
		fmt.Fprintln(os.Stderr, `gsd-watch requires tmux.
Install: brew install tmux
Then start a session: tmux new-session`)
		os.Exit(1)
	}

	// Set tmux pane title for duplicate detection.
	// Deferred reset clears the window-level title on exit so the stale
	// title does not block a future /gsd-watch invocation.
	cwd, _ := os.Getwd()
	fmt.Printf("\033]2;gsd-watch:%s\007", filepath.Base(cwd))
	defer fmt.Printf("\033]2;\007")

	events := make(chan tea.Msg, 10)
	p := tea.NewProgram(
		app.New(events, *noEmoji),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
