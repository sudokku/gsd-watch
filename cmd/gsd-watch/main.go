package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/tui/app"
)

func main() {
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
