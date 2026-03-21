package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/tui/app"
)

func main() {
	// --help flag: print usage and exit.
	showHelp := flag.Bool("help", false, "Show usage information")
	flag.Parse()
	if *showHelp {
		fmt.Println(`gsd-watch — live GSD project status sidebar for tmux

Start via /gsd-watch slash command in Claude Code.

Keybindings:
  ←/h    collapse        →/l    expand
  ↓/j    move down       ↑/k    move up
  e      expand all      w      collapse all
  ?      help overlay    qq     quit

https://github.com/radu/gsd-watch`)
		os.Exit(0)
	}

	// Outside-tmux detection: require tmux session.
	if os.Getenv("TMUX") == "" {
		fmt.Fprintln(os.Stderr, `gsd-watch requires tmux.
Install: brew install tmux
Then start a session: tmux new-session`)
		os.Exit(1)
	}

	// Set tmux pane title for duplicate detection
	cwd, _ := os.Getwd()
	fmt.Printf("\033]2;gsd-watch:%s\007", filepath.Base(cwd))

	events := make(chan tea.Msg, 10)
	p := tea.NewProgram(
		app.New(events),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
