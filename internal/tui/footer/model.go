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
	width         int
	refreshFlash  bool
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

// SetWidth returns a new FooterModel with the terminal width stored for height calculation.
func (f FooterModel) SetWidth(width int) FooterModel {
	f.width = width
	return f
}

// SetRefreshFlash returns a new FooterModel with the refresh flash state toggled.
func (f FooterModel) SetRefreshFlash(flash bool) FooterModel {
	f.refreshFlash = flash
	return f
}

// Height returns the number of lines the footer occupies.
// Dynamic: extra lines are added when currentAction wraps.
// Returns 3 when width is not yet set (before first WindowSizeMsg).
func (f FooterModel) Height() int {
	if f.width == 0 {
		return 3 // default before width is known (action + 2 hint lines)
	}
	if f.width < tui.MinWidth {
		return 1 // single "too narrow" line
	}
	return len(f.actionLines()) + 2 // action lines + 2 hint lines
}

// actionLines word-wraps currentAction to fit beside timeSince on line 1.
func (f FooterModel) actionLines() []string {
	timeSinceW := lipgloss.Width(timeSince(f.lastUpdated))
	availWidth := f.width - timeSinceW - 1
	if availWidth < 10 {
		// Terminal too narrow to share the line — give action the full width.
		availWidth = f.width
	}
	return tui.WordWrap(f.currentAction, availWidth)
}

// View renders the footer for the given terminal width.
// If width is below tui.MinWidth, a "too narrow" placeholder is returned.
func (f FooterModel) View(width int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	grayStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)

	// Build the time-since string with refresh icon.
	ts := timeSince(f.lastUpdated)
	var timeSinceStr string
	if f.refreshFlash {
		timeSinceStr = tui.RefreshFlashStyle.Render("⟳ " + ts)
	} else {
		timeSinceStr = grayStyle.Render("↺ " + ts)
	}
	rightWidth := lipgloss.Width(timeSinceStr)

	// Use stored width for wrapping; fall back to the passed width if not set.
	wrapWidth := f.width
	if wrapWidth == 0 {
		wrapWidth = width
	}
	actionParts := tui.WordWrap(f.currentAction, wrapWidth-rightWidth-1)

	var allLines []string
	for i, part := range actionParts {
		rendered := grayStyle.Render(part)
		if i == 0 {
			// Line 1: action on left, time-since (with icon) on right.
			actionW := lipgloss.Width(rendered)
			padding := width - actionW - rightWidth
			if padding < 0 {
				padding = 0
			}
			allLines = append(allLines, rendered+strings.Repeat(" ", padding)+timeSinceStr)
		} else {
			allLines = append(allLines, rendered)
		}
	}

	// Line 2: navigation hints
	navLine := grayStyle.Render("←h · ↓j · ↑k · →l")
	allLines = append(allLines, navLine)

	// Line 3: actions left, quit right-aligned
	leftActions := "w collapse · e expand"
	rightQuit := "qq esc quit"
	leftW := lipgloss.Width(leftActions)
	rightW := lipgloss.Width(rightQuit)
	actionPad := width - leftW - rightW
	if actionPad < 1 {
		actionPad = 1
	}
	actionsLine := grayStyle.Render(leftActions + strings.Repeat(" ", actionPad) + rightQuit)
	allLines = append(allLines, actionsLine)

	return strings.Join(allLines, "\n")
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
