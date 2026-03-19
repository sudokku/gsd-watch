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
			line := expandIndicator + icon + " " + row.Phase.Name
			if i == t.cursor {
				line = highlightStyle.Render(line)
			}
			lines = append(lines, line)

			// Render badges on a separate line below the phase header
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

			line := connector + icon + " " + row.Plan.Title + nowMarker
			if i == t.cursor {
				line = highlightStyle.Render(line)
			}
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}
