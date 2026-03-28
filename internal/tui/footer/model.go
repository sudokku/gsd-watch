package footer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

// spinFrames is the braille spinner sequence cycled during active file changes.
var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// FooterModel renders the bottom bar with last changed file, time since last
// update, and keybinding hints.
type FooterModel struct {
	lastFile      string    // base name of the last changed file, e.g. "STATE.md"
	lastUpdated   time.Time // when the last file change was parsed
	activeChanges bool      // true while within the 3-second activity window
	spinFrame     int       // current index into spinFrames
	keys          tui.KeyMap
	width         int
	quitPending   bool
	theme         tui.Theme
}

// New creates a FooterModel populated from the given ProjectData and KeyMap.
// Uses the default theme; call SetTheme to apply a different preset.
func New(data parser.ProjectData, keys tui.KeyMap) FooterModel {
	return FooterModel{
		lastUpdated: data.LastUpdated,
		keys:        keys,
		theme:       tui.ThemeDefault(),
	}
}

// SetData returns a new FooterModel with lastUpdated refreshed from data.
func (f FooterModel) SetData(data parser.ProjectData) FooterModel {
	f.lastUpdated = data.LastUpdated
	return f
}

// SetWidth returns a new FooterModel with the terminal width stored for height calculation.
func (f FooterModel) SetWidth(width int) FooterModel {
	f.width = width
	return f
}

// SetLastFile returns a new FooterModel with lastFile updated to the base name of path.
func (f FooterModel) SetLastFile(file string) FooterModel {
	f.lastFile = file
	return f
}

// SetActiveChanges returns a new FooterModel with the activity state set.
// Passing false also resets the spinner frame to 0.
func (f FooterModel) SetActiveChanges(active bool) FooterModel {
	f.activeChanges = active
	if !active {
		f.spinFrame = 0
	}
	return f
}

// AdvanceSpinFrame returns a new FooterModel with spinFrame incremented by one,
// wrapping around the spinFrames slice.
func (f FooterModel) AdvanceSpinFrame() FooterModel {
	f.spinFrame = (f.spinFrame + 1) % len(spinFrames)
	return f
}

// ActiveChanges reports whether the footer is currently in the activity window.
func (f FooterModel) ActiveChanges() bool {
	return f.activeChanges
}

// SetQuitPending returns a new FooterModel with the quit-confirm state set.
func (f FooterModel) SetQuitPending(pending bool) FooterModel {
	f.quitPending = pending
	return f
}

// SetTheme returns a new FooterModel with the given theme applied.
func (f FooterModel) SetTheme(th tui.Theme) FooterModel {
	f.theme = th
	return f
}

// Height returns the number of lines the footer occupies.
// Dynamic: extra lines are added when the left label wraps.
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

// actionLines word-wraps the left label to fit beside the right-side indicator.
func (f FooterModel) actionLines() []string {
	// icon (1 cell) + space (1) + time string
	rightW := 2 + lipgloss.Width(timeSince(f.lastUpdated))
	availWidth := f.width - 2 - rightW - 1 // 2 for L/R padding, 1 for gap
	if availWidth < 10 {
		availWidth = f.width - 2
		if availWidth < 1 {
			availWidth = 1
		}
	}
	return tui.WordWrap(f.leftLabel(), availWidth)
}

// leftLabel returns the text shown on the left side of the footer action row.
func (f FooterModel) leftLabel() string {
	if f.lastFile == "" {
		return "watching…"
	}
	return "Last change: " + f.lastFile
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
	// Use theme.SeparatorFg for theme-aware coloring.
	sepStyle := lipgloss.NewStyle().Foreground(f.theme.SeparatorFg)
	sepLine := sepStyle.Render(strings.Repeat("─", width))

	// Build right-side indicator: spinner or checkmark + time string.
	ts := timeSince(f.lastUpdated)
	var rightStr string
	if f.activeChanges {
		frame := spinFrames[f.spinFrame%len(spinFrames)]
		rightStr = f.theme.RefreshFlash.Render(frame + " " + ts)
	} else {
		rightStr = grayStyle.Render("✓ " + ts)
	}
	rightWidth := lipgloss.Width(rightStr)

	// Use stored width for wrapping; fall back to the passed width if not set.
	wrapWidth := f.width
	if wrapWidth == 0 {
		wrapWidth = width
	}
	actionParts := tui.WordWrap(f.leftLabel(), wrapWidth-2-rightWidth-1)

	var allLines []string
	labelStyle := grayStyle
	if f.lastFile != "" {
		labelStyle = lipgloss.NewStyle()
	}
	for i, part := range actionParts {
		rendered := labelStyle.Render(part)
		if i == 0 {
			// Line 1: label on left, indicator on right, 1-char L/R padding.
			actionW := lipgloss.Width(rendered)
			padding := contentWidth - actionW - rightWidth
			if padding < 0 {
				padding = 0
			}
			allLines = append(allLines, strings.Repeat(" ", pad)+rendered+strings.Repeat(" ", padding)+rightStr)
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
	if t.IsZero() {
		return "–"
	}
	d := time.Since(t)
	switch {
	case d < 3*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
}
