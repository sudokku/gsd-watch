package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/tui/app"
)

func main() {
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
