package tree

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

var highlightStyle = lipgloss.NewStyle().Bold(true)

// View renders the tree as a string for the given terminal width.
// If width < tui.MinWidth, it returns a "too narrow" placeholder.
func (t TreeModel) View(width int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	// Empty state (D-01): no phases in project data.
	if len(t.data.Phases) == 0 {
		msg := "No GSD project found.\n\nTo get started, open\nClaude Code and run:\n/gsd:new-project"
		centered := lipgloss.NewStyle().
			Width(width - 2).
			Align(lipgloss.Center).
			Foreground(tui.ColorGray).
			Render(msg)
		var padded []string
		for _, line := range strings.Split(centered, "\n") {
			padded = append(padded, " "+line)
		}
		return strings.Join(padded, "\n")
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
			isDimmedPhase := row.Phase.Status == parser.StatusComplete

			// Calculate prefix width and available wrap width for phase name.
			prefixStr := expandIndicator + icon + " "
			prefixWidth := lipgloss.Width(prefixStr)
			// -1 mirrors D-10 left-padding added at bottom of View().
			wrapWidth := width - 1 - prefixWidth
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			nameParts := tui.WordWrap(row.Phase.Name, wrapWidth)

			continuation := strings.Repeat(" ", prefixWidth)

			for j, part := range nameParts {
				var text string
				switch {
				case i == t.cursor:
					text = highlightStyle.Render(part)
				case isDimmedPhase:
					text = tui.PendingStyle.Render(part)
				default:
					text = part
				}

				var phaseLine string
				if j == 0 {
					if isDimmedPhase {
						phaseLine = tui.PendingStyle.Render(prefixStr) + text
					} else {
						phaseLine = prefixStr + text
					}
				} else {
					cont := continuation
					if isDimmedPhase {
						cont = tui.PendingStyle.Render(continuation)
					}
					phaseLine = cont + text
				}
				lines = append(lines, phaseLine)
			}

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

			// D-02: show "(no plans yet)" for expanded phases with no plans.
			if t.expanded[row.Phase.DirName] && len(row.Phase.Plans) == 0 {
				lines = append(lines, "    "+tui.PendingStyle.Render("(no plans yet)"))
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

			// Bug-2 fix: subtract 1 for the D-10 left-padding so the assembled line
			// is exactly `width` cells wide after the pad is prepended.
			prefixWidth := lipgloss.Width(connector) + lipgloss.Width(icon) + 1
			nowWidth := lipgloss.Width(nowMarker)
			wrapWidth := width - 1 - prefixWidth - nowWidth
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			// Bug-1 fix: use │ (U+2502) which aligns with ├/└ on the right cell edge.
			// Bug-2 fix: same -1 adjustment so continuation column matches wrapWidth.
			var continuation string
			if isLast {
				continuation = strings.Repeat(" ", prefixWidth)
			} else {
				continuation = "    │" + strings.Repeat(" ", prefixWidth-5)
			}
			titleParts := tui.WordWrap(row.Plan.Title, wrapWidth)

			// D-03: dim plan rows belonging to a completed phase.
			// Bug-3 fix: apply dim independently to the connector/continuation and
			// the text rather than to the whole assembled line. Wrapping the entire
			// line with PendingStyle.Render() causes the icon's inner \033[0m reset
			// to kill the gray mid-string, so line-1 text ends up white while
			// line-2 continuation (no inner reset) stays gray.
			isDimmed := phase.Status == parser.StatusComplete

			var itemLines []string
			for j, part := range titleParts {
				suffix := ""
				if j == len(titleParts)-1 {
					suffix = nowMarker
				}
				rawText := part + suffix
				var text string
				switch {
				case i == t.cursor:
					text = highlightStyle.Render(rawText)
				case isDimmed:
					text = tui.PendingStyle.Render(rawText)
				default:
					text = rawText
				}

				var l string
				if j == 0 {
					c := connector
					if isDimmed {
						c = tui.PendingStyle.Render(connector)
					}
					l = c + icon + " " + text
				} else {
					cont := continuation
					if isDimmed {
						cont = tui.PendingStyle.Render(continuation)
					}
					l = cont + text
				}
				itemLines = append(itemLines, l)
			}
			lines = append(lines, strings.Join(itemLines, "\n"))
		}
	}

	// D-10: add 1-char left padding to every line.
	var padded []string
	for _, line := range strings.Split(strings.Join(lines, "\n"), "\n") {
		padded = append(padded, " "+line)
	}
	return strings.Join(padded, "\n")
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
		// Calculate wrapped phase name line count.
		// expandIndicator ("▶ " or "▼ ") is always 2 chars wide. icon + " " is prefix.
		// We use a fixed prefix string for width calculation.
		expandIndicatorWidth := 2 // "▶ " or "▼ " — both 2 display cells
		icon := tui.StatusIcon(row.Phase.Status)
		prefixWidth := expandIndicatorWidth + lipgloss.Width(icon) + 1
		wrapWidth := width - 1 - prefixWidth
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		n := len(tui.WordWrap(row.Phase.Name, wrapWidth))
		if len(row.Phase.Badges) > 0 {
			for _, b := range row.Phase.Badges {
				if tui.BadgeString(b) != "" {
					n++ // badge line
					break
				}
			}
		}
		// D-02: "(no plans yet)" placeholder line when expanded and no plans.
		if row.Expanded && len(row.Phase.Plans) == 0 {
			n++ // placeholder line
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
		// -1 mirrors the D-10 left-padding adjustment in View() so line counts match.
		wrapWidth := width - 1 - prefixWidth - nowWidth
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		parts := tui.WordWrap(row.Plan.Title, wrapWidth)
		return len(parts)
	}
	return 1
}
