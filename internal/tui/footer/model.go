package footer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

// FooterModel renders the bottom bar with current action, time since last update,
// and keybinding hints.
type FooterModel struct {
	currentAction string
	lastUpdated   time.Time
	keys          tui.KeyMap
}

// New creates a FooterModel populated from the given ProjectData and KeyMap.
func New(data parser.ProjectData, keys tui.KeyMap) FooterModel {
	return FooterModel{
		currentAction: data.CurrentAction,
		lastUpdated:   data.LastUpdated,
		keys:          keys,
	}
}

// SetData returns a new FooterModel with currentAction and lastUpdated updated from data.
func (f FooterModel) SetData(data parser.ProjectData) FooterModel {
	f.currentAction = data.CurrentAction
	f.lastUpdated = data.LastUpdated
	return f
}

// Height returns the fixed number of lines the footer occupies (2).
// This is used by the root model for viewport height calculation.
func (f FooterModel) Height() int {
	return 2
}

// View renders the footer for the given terminal width.
// If width is below tui.MinWidth, a "too narrow" placeholder is returned.
func (f FooterModel) View(width int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	grayStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)

	// Line 1: current action on left, time-since on right.
	actionStr := grayStyle.Render(f.currentAction)
	timeSinceStr := timeSince(f.lastUpdated)

	actionWidth := lipgloss.Width(actionStr)
	rightWidth := len(timeSinceStr)
	padding := width - actionWidth - rightWidth
	if padding < 0 {
		padding = 0
	}
	line1 := actionStr + strings.Repeat(" ", padding) + timeSinceStr

	// Line 2: keybinding hints.
	bindings := f.keys.ShortHelp()
	hints := make([]string, 0, len(bindings))
	for _, b := range bindings {
		help := b.Help()
		if help.Key != "" {
			hints = append(hints, help.Key+" "+help.Desc)
		}
	}
	line2 := grayStyle.Render(strings.Join(hints, " · "))

	return strings.Join([]string{line1, line2}, "\n")
}

// timeSince returns a human-readable duration string for how long ago t was.
func timeSince(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
}
