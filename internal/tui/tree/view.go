package tree

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tui "github.com/radu/gsd-watch/internal/tui"
)

var highlightStyle = lipgloss.NewStyle().Bold(true)

// View renders the tree as a string for the given terminal width.
// If width < tui.MinWidth, it returns a "too narrow" placeholder.
func (t TreeModel) View(width int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	rows := t.VisibleRows()
	var lines []string

	for i, row := range rows {
		switch row.Kind {
		case RowPhase:
			expandIndicator := "▶ "
			if t.expanded[row.Phase.DirName] {
				expandIndicator = "▼ "
			}
			icon := tui.StatusIcon(row.Phase.Status)
			// Apply highlight to the name text only — the icon contains ANSI reset
			// codes that would kill bold if the entire line were wrapped.
			name := row.Phase.Name
			if i == t.cursor {
				name = highlightStyle.Render(name)
			}
			lines = append(lines, expandIndicator+icon+" "+name)

			// Render badges on a separate line below the phase header.
			if len(row.Phase.Badges) > 0 {
				var badgeParts []string
				for _, badge := range row.Phase.Badges {
					b := tui.BadgeString(badge)
					if b != "" {
						badgeParts = append(badgeParts, b)
					}
				}
				if len(badgeParts) > 0 {
					lines = append(lines, "    "+strings.Join(badgeParts, " "))
				}
			}

		case RowPlan:
			// Determine connector: last plan in phase gets └──, others get ├──
			phase := t.data.Phases[row.PhaseIdx]
			isLast := row.Plan.Filename == phase.Plans[len(phase.Plans)-1].Filename
			connector := "    ├── "
			if isLast {
				connector = "    └── "
			}

			icon := tui.StatusIcon(row.Plan.Status)
			nowMarker := ""
			if row.Plan.IsActive {
				nowMarker = " " + tui.NowMarkerStyle.Render("← now")
			}

			// Word-wrap the title to fit within the available width.
			prefixWidth := lipgloss.Width(connector) + lipgloss.Width(icon) + 1
			nowWidth := lipgloss.Width(nowMarker)
			wrapWidth := width - prefixWidth - nowWidth
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			continuation := strings.Repeat(" ", prefixWidth)
			titleParts := tui.WordWrap(row.Plan.Title, wrapWidth)

			// Apply highlight to each text part individually so the icon's ANSI
			// reset code on line 1 does not interfere with subsequent lines.
			var itemLines []string
			for j, part := range titleParts {
				suffix := ""
				if j == len(titleParts)-1 {
					suffix = nowMarker
				}
				text := part + suffix
				if i == t.cursor {
					text = highlightStyle.Render(text)
				}
				var l string
				if j == 0 {
					l = connector + icon + " " + text
				} else {
					l = continuation + text
				}
				itemLines = append(itemLines, l)
			}
			lines = append(lines, strings.Join(itemLines, "\n"))
		}
	}

	return strings.Join(lines, "\n")
}

// RenderedCursorLine returns the line index (0-based) of the cursor row's
// first rendered line within the full tree output. Used by the app model to
// scroll the viewport so the cursor is always visible.
func (t TreeModel) RenderedCursorLine(width int) int {
	rows := t.VisibleRows()
	line := 0
	for i, row := range rows {
		if i == t.cursor {
			return line
		}
		line += renderedRowLines(row, width)
	}
	return line
}

// renderedRowLines returns the number of output lines a single row occupies.
func renderedRowLines(row Row, width int) int {
	switch row.Kind {
	case RowPhase:
		n := 1 // phase header line
		if len(row.Phase.Badges) > 0 {
			for _, b := range row.Phase.Badges {
				if tui.BadgeString(b) != "" {
					n++ // badge line
					break
				}
			}
		}
		return n

	case RowPlan:
		icon := tui.StatusIcon(row.Plan.Status)
		nowWidth := 0
		if row.Plan.IsActive {
			nowWidth = lipgloss.Width(" " + tui.NowMarkerStyle.Render("← now"))
		}
		const connectorWidth = 8 // "    ├── " or "    └── " = 8 cells
		prefixWidth := connectorWidth + lipgloss.Width(icon) + 1
		wrapWidth := width - prefixWidth - nowWidth
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		parts := tui.WordWrap(row.Plan.Title, wrapWidth)
		return len(parts)
	}
	return 1
}
