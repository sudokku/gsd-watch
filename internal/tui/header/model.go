package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

// HeaderModel renders the top bar with project info and progress bar.
type HeaderModel struct {
	projectName  string
	modelProfile string
	mode         string
	completion   float64 // 0.0 to 1.0
}

// New creates a HeaderModel populated from the given ProjectData.
func New(data parser.ProjectData) HeaderModel {
	return HeaderModel{
		projectName:  data.Name,
		modelProfile: data.ModelProfile,
		mode:         data.Mode,
		completion:   data.CompletionPercent(),
	}
}

// SetData returns a new HeaderModel with fields updated from data.
func (h HeaderModel) SetData(data parser.ProjectData) HeaderModel {
	h.projectName = data.Name
	h.modelProfile = data.ModelProfile
	h.mode = data.Mode
	h.completion = data.CompletionPercent()
	return h
}

// Height returns the fixed number of lines the header occupies (3).
// This is used by the root model for viewport height calculation.
func (h HeaderModel) Height() int {
	return 3
}

// View renders the header for the given terminal width.
// If width is below tui.MinWidth, a "too narrow" placeholder is returned.
func (h HeaderModel) View(width int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	// Line 1: project name on left, profile·mode on right.
	nameStr := lipgloss.NewStyle().Bold(true).Render(h.projectName)
	profileModeStr := h.modelProfile + " · " + h.mode

	// Calculate visible length of nameStr (lipgloss Bold adds ANSI escapes,
	// so we use lipgloss.Width for correct padding math).
	nameWidth := lipgloss.Width(nameStr)
	rightWidth := len(profileModeStr)
	padding := width - nameWidth - rightWidth
	if padding < 0 {
		padding = 0
	}
	line1 := nameStr + strings.Repeat(" ", padding) + profileModeStr

	// Line 2: progress bar spanning full width.
	line2 := progressBar(h.completion, width)

	// Line 3: separator line in gray.
	separatorStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)
	line3 := separatorStyle.Render(strings.Repeat("─", width))

	return strings.Join([]string{line1, line2, line3}, "\n")
}

// progressBar renders a filled/empty block bar for the given percentage and width.
func progressBar(pct float64, width int) string {
	barWidth := width - 2
	if barWidth < 0 {
		barWidth = 0
	}
	filled := int(float64(barWidth) * pct)
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}

	filledStr := lipgloss.NewStyle().Foreground(tui.ColorGreen).Render(strings.Repeat("▓", filled))
	emptyStr := lipgloss.NewStyle().Foreground(tui.ColorGray).Render(strings.Repeat("░", barWidth-filled))
	return filledStr + emptyStr
}
