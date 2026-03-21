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
	quitPending   bool
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

// SetQuitPending returns a new FooterModel with the quit-confirm state set.
func (f FooterModel) SetQuitPending(pending bool) FooterModel {
	f.quitPending = pending
	return f
}

// Height returns the number of lines the footer occupies.
// Dynamic: extra lines are added when currentAction wraps.
// Returns 5 when width is not yet set (before first WindowSizeMsg).
func (f FooterModel) Height() int {
	if f.width == 0 {
		return 5 // default: separator + action + 2 hint lines + blank
	}
	if f.width < tui.MinWidth {
		return 1 // single "too narrow" line
	}
	return len(f.actionLines()) + 4 // separator + action lines + 2 hint lines + blank
}

// actionLines word-wraps currentAction to fit beside timeSince on line 1.
func (f FooterModel) actionLines() []string {
	timeSinceW := lipgloss.Width(timeSince(f.lastUpdated))
	availWidth := f.width - 2 - timeSinceW - 1 // 2 for L/R padding, 1 for space
	if availWidth < 10 {
		// Terminal too narrow to share the line — give action the content width.
		availWidth = f.width - 2
		if availWidth < 1 {
			availWidth = 1
		}
	}
	return tui.WordWrap(f.currentAction, availWidth)
}

// View renders the footer for the given terminal width.
// If width is below tui.MinWidth, a "too narrow" placeholder is returned.
func (f FooterModel) View(width int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	const pad = 1
	contentWidth := width - 2*pad

	grayStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)

	// First line: light-horizontal separator spanning full width.
	sepLine := grayStyle.Render(strings.Repeat("─", width))

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
	actionParts := tui.WordWrap(f.currentAction, wrapWidth-2-rightWidth-1)

	var allLines []string
	for i, part := range actionParts {
		rendered := grayStyle.Render(part)
		if i == 0 {
			// Line 1: action on left, time-since (with icon) on right, 1-char L/R padding.
			actionW := lipgloss.Width(rendered)
			padding := contentWidth - actionW - rightWidth
			if padding < 0 {
				padding = 0
			}
			allLines = append(allLines, strings.Repeat(" ", pad)+rendered+strings.Repeat(" ", padding)+timeSinceStr)
		} else {
			allLines = append(allLines, strings.Repeat(" ", pad)+rendered)
		}
	}

	if f.quitPending {
		// Replace both hint lines with a centered confirmation prompt.
		msg := "press q or esc again to exit"
		msgW := lipgloss.Width(msg)
		leftPad := (width - msgW) / 2
		if leftPad < 0 {
			leftPad = 0
		}
		allLines = append(allLines, strings.Repeat(" ", leftPad)+tui.QuitPendingStyle.Render(msg))
		allLines = append(allLines, "") // keep line count equal to normal (2 hint lines)
	} else {
		// Navigation hints with 1-char left padding.
		navLine := strings.Repeat(" ", pad) + grayStyle.Render("←h · ↓j · ↑k · →l")
		allLines = append(allLines, navLine)

		// Actions left, quit right-aligned, 1-char L/R padding.
		leftActions := "w collapse · e expand · ? help"
		rightQuit := "qq esc quit"
		leftW := lipgloss.Width(leftActions)
		rightW := lipgloss.Width(rightQuit)
		actionPad := contentWidth - leftW - rightW
		if actionPad < 1 {
			actionPad = 1
		}
		actionsLine := strings.Repeat(" ", pad) + grayStyle.Render(leftActions+strings.Repeat(" ", actionPad)+rightQuit)
		allLines = append(allLines, actionsLine)
	}

	// Trailing blank line for bottom breathing room.
	allLines = append(allLines, "")

	return strings.Join(append([]string{sepLine}, allLines...), "\n")
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
