package tree

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

// FormatArchiveDate converts an ISO date string ("2006-01-02") to "Jan 2006" format.
// Returns empty string for empty input or invalid dates.
func FormatArchiveDate(iso string) string {
	if iso == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return ""
	}
	return t.Format("Jan 2006")
}

// RenderArchiveRow renders a single archived milestone row in emoji or noEmoji mode.
// Archive rows are styled with th.Pending — non-interactive and dimmed.
func RenderArchiveRow(am parser.ArchivedMilestone, noEmoji bool, th tui.Theme) string {
	indicator := "▸ "
	checkmark := "✓"
	if noEmoji {
		indicator = "> "
		checkmark = "[done]"
	}
	dateStr := FormatArchiveDate(am.CompletionDate)
	var row string
	if dateStr != "" {
		row = fmt.Sprintf("%s%s — %d phases %s  %s", indicator, am.Name, am.PhaseCount, checkmark, dateStr)
	} else {
		row = fmt.Sprintf("%s%s — %d phases %s", indicator, am.Name, am.PhaseCount, checkmark)
	}
	return th.Pending.Render(row)
}

// RenderArchiveSeparator renders the "- - Archived Milestones - - -..." separator line
// at full width (no D-10 offset — the caller must not add left-padding to this line).
func RenderArchiveSeparator(width int) string {
	label := " Archived Milestones "
	prefix := "- -"
	body := prefix + label
	remaining := width - len(body)
	if remaining < 0 {
		remaining = 0
	}
	dashes := strings.Repeat(" -", (remaining/2)+1)
	result := body + dashes
	if len(result) > width {
		result = result[:width]
	}
	return lipgloss.NewStyle().Render(result)
}

// RenderArchiveZone renders the full pinned archive zone: separator + one row per milestone.
// Returns empty string when archives is nil or empty (D-04).
func RenderArchiveZone(archives []parser.ArchivedMilestone, width int, noEmoji bool, th tui.Theme) string {
	if len(archives) == 0 {
		return ""
	}
	lines := []string{RenderArchiveSeparator(width)}
	for _, am := range archives {
		lines = append(lines, RenderArchiveRow(am, noEmoji, th))
	}
	return strings.Join(lines, "\n")
}

// IsPhaseActive returns true if the cursor is on the phase row at phaseRowIdx,
// or on any of its consecutive child RowPlan rows immediately after it.
func (t TreeModel) IsPhaseActive(rows []Row, phaseRowIdx int) bool {
	if t.cursor == phaseRowIdx {
		return true
	}
	// Check if cursor is on a child plan row of this phase
	if t.cursor > phaseRowIdx {
		for ci := phaseRowIdx + 1; ci < len(rows); ci++ {
			if rows[ci].Kind != RowPlan {
				break
			}
			if ci == t.cursor {
				return true
			}
		}
	}
	return false
}

// View renders the tree as a string for the given terminal width and height.
// The available height is split between a scrollable zone (phases + quick tasks)
// and a pinned archive zone at the bottom (separator + archive rows).
// If width < tui.MinWidth, it returns a "too narrow" placeholder.
func (t TreeModel) View(width, height int) string {
	if width < tui.MinWidth {
		return "◀ too narrow"
	}

	th := themeFor(t.opts)

	// Empty state (D-01): no phases in project data.
	if len(t.data.Phases) == 0 {
		var msg string
		if len(t.data.ArchivedMilestones) > 0 {
			msg = "All milestones archived.\n\nStart a new milestone:\n/gsd:new-milestone\n\nor run a quick task:\n/gsd:quick"
		} else {
			msg = "No GSD project found.\n\nTo get started, open\nClaude Code and run:\n/gsd:new-project"
		}
		centered := lipgloss.NewStyle().
			Width(width - 2).
			Align(lipgloss.Center).
			Foreground(th.EmptyFg).
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
			icon := tui.StatusIcon(row.Phase.Status, t.opts.NoEmoji, th)
			isDimmedPhase := row.Phase.Status == parser.StatusComplete
			phaseActive := t.IsPhaseActive(rows, i)

			// Calculate prefix width and available wrap width for phase name.
			prefixStr := expandIndicator + icon + " "
			prefixWidth := lipgloss.Width(prefixStr)
			// -2 mirrors D-10 left-padding (1 char) + implicit right-padding (1 char).
			wrapWidth := width - 2 - prefixWidth
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			nameParts := tui.WordWrap(row.Phase.Name, wrapWidth)

			continuation := strings.Repeat(" ", prefixWidth)

			for j, part := range nameParts {
				var text string
				switch {
				case phaseActive:
					text = th.Highlight.Render(part)
				case isDimmedPhase:
					text = th.Pending.Render(part)
				default:
					text = part
				}

				var phaseLine string
				if j == 0 {
					if phaseActive {
						phaseLine = th.Highlight.Render(prefixStr) + text
					} else if isDimmedPhase {
						phaseLine = th.Pending.Render(prefixStr) + text
					} else {
						phaseLine = prefixStr + text
					}
				} else {
					cont := continuation
					if phaseActive {
						cont = th.Highlight.Render(continuation)
					} else if isDimmedPhase {
						cont = th.Pending.Render(continuation)
					}
					phaseLine = cont + text
				}
				lines = append(lines, phaseLine)
			}

			// Render badges on a separate line below the phase header.
			if len(row.Phase.Badges) > 0 {
				var badgeParts []string
				for _, badge := range row.Phase.Badges {
					b := tui.BadgeString(badge, t.opts.NoEmoji)
					if b != "" {
						badgeParts = append(badgeParts, b)
					}
				}
				if len(badgeParts) > 0 {
					badgeLine := "    " + strings.Join(badgeParts, " ")
					switch {
					case phaseActive:
						lines = append(lines, th.Highlight.Render(badgeLine))
					case isDimmedPhase:
						lines = append(lines, th.Pending.Render(badgeLine))
					default:
						lines = append(lines, badgeLine)
					}
				}
			}

			// D-02: show "(no plans yet)" for expanded phases with no plans.
			if t.expanded[row.Phase.DirName] && len(row.Phase.Plans) == 0 {
				lines = append(lines, "    "+th.Pending.Render("(no plans yet)"))
			}

		case RowPlan:
			// Determine connector: last plan in phase gets └──, others get ├──
			phase := t.data.Phases[row.PhaseIdx]
			isLast := row.Plan.Filename == phase.Plans[len(phase.Plans)-1].Filename
			connector := "    ├── "
			if isLast {
				connector = "    └── "
			}

			icon := tui.StatusIcon(row.Plan.Status, t.opts.NoEmoji, th)
			nowMarker := ""
			if row.Plan.IsActive {
				nowMarker = " " + th.NowMarker.Render("← now")
			}

			// Bug-2 fix: subtract 2 for D-10 left-padding (1) + implicit right-padding (1)
			// so content never reaches the terminal's rightmost column.
			prefixWidth := lipgloss.Width(connector) + lipgloss.Width(icon) + 1
			nowWidth := lipgloss.Width(nowMarker)
			wrapWidth := width - 2 - prefixWidth - nowWidth
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			// Bug-1 fix: use │ (U+2502) which aligns with ├/└ on the right cell edge.
			// Bug-2 fix: same -2 adjustment so continuation column matches wrapWidth.
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
			// line with Pending.Render() causes the icon's inner \033[0m reset
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
					text = th.Highlight.Render(rawText)
				case isDimmed:
					text = th.Pending.Render(rawText)
				default:
					text = rawText
				}

				var l string
				if j == 0 {
					c := connector
					if isDimmed {
						c = th.Pending.Render(connector)
					}
					l = c + icon + " " + text
				} else {
					cont := continuation
					if isDimmed {
						cont = th.Pending.Render(continuation)
					}
					l = cont + text
				}
				itemLines = append(itemLines, l)
			}
			lines = append(lines, strings.Join(itemLines, "\n"))

		case RowQuickSection:
			indicator := "▶ "
			if t.expanded[quickSectionKey] {
				indicator = "▼ "
			}
			label := "Quick tasks"
			if i == t.cursor {
				lines = append(lines, th.Highlight.Render(indicator)+th.Highlight.Render(label))
			} else {
				lines = append(lines, indicator+label)
			}
			// D-02: empty state placeholder when expanded with no quick tasks
			if t.expanded[quickSectionKey] && len(t.data.QuickTasks) == 0 {
				lines = append(lines, "    "+th.Pending.Render("(no quick tasks)"))
			}

		case RowQuickTask:
			qt := row.QuickTask
			isLast := row.QuickTaskIdx == len(t.data.QuickTasks)-1
			connector := "    ├── "
			if isLast {
				connector = "    └── "
			}
			icon := tui.StatusIcon(qt.Status, t.opts.NoEmoji, th)
			isDimmed := qt.Status == parser.StatusComplete

			// Calculate wrap width — same pattern as RowPlan
			prefixWidth := lipgloss.Width(connector) + lipgloss.Width(icon) + 1
			wrapWidth := width - 2 - prefixWidth
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			nameParts := tui.WordWrap(qt.DisplayName, wrapWidth)

			var continuation string
			if isLast {
				continuation = strings.Repeat(" ", prefixWidth)
			} else {
				continuation = "    │" + strings.Repeat(" ", prefixWidth-5)
			}

			var itemLines []string
			for j, part := range nameParts {
				rawText := part
				var text string
				switch {
				case i == t.cursor:
					text = th.Highlight.Render(rawText)
				case isDimmed:
					text = th.Pending.Render(rawText)
				default:
					text = rawText
				}

				var l string
				if j == 0 {
					c := connector
					if isDimmed {
						c = th.Pending.Render(connector)
					}
					l = c + icon + " " + text
				} else {
					cont := continuation
					if isDimmed {
						cont = th.Pending.Render(continuation)
					}
					l = cont + text
				}
				itemLines = append(itemLines, l)
			}
			lines = append(lines, strings.Join(itemLines, "\n"))
		}
	}

	// D-10: add 1-char left padding to every line.
	// Right padding is achieved by reducing wrapWidth by 2 (left + right)
	// so content never reaches the terminal's rightmost column.
	var padded []string
	for _, line := range strings.Split(strings.Join(lines, "\n"), "\n") {
		padded = append(padded, " "+line)
	}

	// View() renders only the scrollable zone (phases + quick tasks).
	// The archive zone is rendered separately via ArchiveZone() and
	// pinned outside the viewport by the app model.
	return strings.Join(padded, "\n")
}

// ArchiveZoneHeight returns the number of lines the pinned archive zone occupies.
// Returns 0 when no archived milestones exist.
func (t TreeModel) ArchiveZoneHeight() int {
	if len(t.data.ArchivedMilestones) == 0 {
		return 0
	}
	return len(t.data.ArchivedMilestones) + 1 // separator + rows
}

// ArchiveZone renders the pinned archive zone: separator at full width (no D-10 padding),
// archive rows with D-10 left-padding. Returns empty string when no archives exist.
func (t TreeModel) ArchiveZone(width int) string {
	th := themeFor(t.opts)
	content := RenderArchiveZone(t.data.ArchivedMilestones, width, t.opts.NoEmoji, th)
	if content == "" {
		return ""
	}
	lines := strings.Split(content, "\n")
	padded := make([]string, 0, len(lines))
	for i, line := range lines {
		if i == 0 {
			// Separator spans full width — no D-10 left-padding.
			padded = append(padded, line)
		} else {
			// Archive rows get D-10 left-padding (1 char), matching the main tree content.
			padded = append(padded, " "+line)
		}
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
		line += t.renderedRowLines(row, width, t.opts.NoEmoji)
	}
	return line
}

// renderedRowLines returns the number of output lines a single row occupies.
func (t TreeModel) renderedRowLines(row Row, width int, noEmoji bool) int {
	th := themeFor(t.opts)
	switch row.Kind {
	case RowPhase:
		// Calculate wrapped phase name line count.
		// expandIndicator ("▶ " or "▼ ") is always 2 chars wide. icon + " " is prefix.
		// We use a fixed prefix string for width calculation.
		expandIndicatorWidth := 2 // "▶ " or "▼ " — both 2 display cells
		icon := tui.StatusIcon(row.Phase.Status, noEmoji, th)
		prefixWidth := expandIndicatorWidth + lipgloss.Width(icon) + 1
		wrapWidth := width - 2 - prefixWidth
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		n := len(tui.WordWrap(row.Phase.Name, wrapWidth))
		if len(row.Phase.Badges) > 0 {
			for _, b := range row.Phase.Badges {
				if tui.BadgeString(b, noEmoji) != "" {
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
		icon := tui.StatusIcon(row.Plan.Status, noEmoji, th)
		nowWidth := 0
		if row.Plan.IsActive {
			// Use theme NowMarker to get correct width (theme may affect ANSI sequence length).
			th := themeFor(t.opts)
			nowWidth = lipgloss.Width(" " + th.NowMarker.Render("← now"))
		}
		const connectorWidth = 8 // "    ├── " or "    └── " = 8 cells
		prefixWidth := connectorWidth + lipgloss.Width(icon) + 1
		// -2 mirrors D-10 left-padding (1) + implicit right-padding (1) in View().
		wrapWidth := width - 2 - prefixWidth - nowWidth
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		parts := tui.WordWrap(row.Plan.Title, wrapWidth)
		return len(parts)

	case RowQuickSection:
		n := 1 // the header line
		if row.Expanded && len(t.data.QuickTasks) == 0 {
			n++ // "(no quick tasks)" placeholder
		}
		return n

	case RowQuickTask:
		icon := tui.StatusIcon(row.QuickTask.Status, noEmoji, th)
		const connectorWidth = 8
		prefixWidth := connectorWidth + lipgloss.Width(icon) + 1
		wrapWidth := width - 2 - prefixWidth
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		return len(tui.WordWrap(row.QuickTask.DisplayName, wrapWidth))
	}
	return 1
}
