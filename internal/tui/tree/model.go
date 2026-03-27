package tree

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

// RowKind identifies whether a Row represents a phase, a plan, or a quick task.
type RowKind int

const (
	RowPhase        RowKind = iota
	RowPlan
	RowQuickSection // "Quick tasks" collapsible header
	RowQuickTask    // individual quick task row
)

// quickSectionKey is the fixed expanded-map key for the Quick Tasks section header.
const quickSectionKey = "__quick_tasks__"

// Row is a single renderable line in the tree.
type Row struct {
	Key          string            // phase DirName for RowPhase; plan Filename for RowPlan; quickSectionKey or DirName for quick rows
	Kind         RowKind
	Phase        parser.Phase      // populated for RowPhase rows
	Plan         parser.Plan       // populated for RowPlan rows
	QuickTask    parser.QuickTask  // populated for RowQuickTask rows
	QuickTaskIdx int               // index into data.QuickTasks (for RowQuickTask rows)
	PhaseIdx     int               // index into data.Phases (for both RowPhase and RowPlan rows)
	Expanded     bool              // true when this phase/section row is currently expanded
}

// Options configures optional rendering behaviors for TreeModel.
type Options struct {
	NoEmoji bool
	Theme   tui.Theme // color theme; zero value resolves to ThemeDefault() at render time
}

// themeFor returns opts.Theme if it has a non-zero Highlight style, else ThemeDefault().
// A zero Theme occurs when Options is constructed without a Theme field (e.g. in tests).
func themeFor(opts Options) tui.Theme {
	// Highlight is always set in every ThemeX() constructor; use it as a zero-check.
	if opts.Theme.Highlight.GetBold() || opts.Theme.Pending.GetForeground() != nil {
		return opts.Theme
	}
	return tui.ThemeDefault()
}

// TreeModel manages collapsible tree state: data, expanded map, and cursor.
type TreeModel struct {
	data     parser.ProjectData
	expanded map[string]bool // key: phase.DirName
	cursor   int
	keys     tui.KeyMap
	opts     Options
}

// SetOptions returns a copy of TreeModel with the given options applied.
func (t TreeModel) SetOptions(o Options) TreeModel {
	t.opts = o
	return t
}

// New returns a TreeModel initialized with an empty expanded map and cursor 0.
func New() TreeModel {
	return TreeModel{
		expanded: make(map[string]bool),
		keys:     tui.DefaultKeyMap(),
	}
}

// SetData replaces the model's data and preserves existing expanded state.
// The cursor is clamped to the new row count.
func (t TreeModel) SetData(d parser.ProjectData) TreeModel {
	t.data = d
	// preserve expanded map; just clamp cursor
	t.cursor = clamp(t.cursor, 0, max(0, len(t.visibleRows())-1))
	return t
}

// visibleRows returns the ordered list of rows currently visible in the tree.
// Returns empty slice when no project data is loaded (no phases).
func (t TreeModel) visibleRows() []Row {
	if len(t.data.Phases) == 0 {
		return nil
	}
	var rows []Row
	for i, phase := range t.data.Phases {
		expanded := t.expanded[phase.DirName]
		rows = append(rows, Row{
			Key:      phase.DirName,
			Kind:     RowPhase,
			Phase:    phase,
			PhaseIdx: i,
			Expanded: expanded,
		})
		if expanded {
			for _, plan := range phase.Plans {
				rows = append(rows, Row{
					Key:      plan.Filename,
					Kind:     RowPlan,
					Plan:     plan,
					PhaseIdx: i,
				})
			}
		}
	}

	// Quick Tasks section — always present below phases
	quickExpanded := t.expanded[quickSectionKey]
	rows = append(rows, Row{
		Key:      quickSectionKey,
		Kind:     RowQuickSection,
		Expanded: quickExpanded,
	})
	if quickExpanded {
		for i, qt := range t.data.QuickTasks {
			rows = append(rows, Row{
				Key:          qt.DirName,
				Kind:         RowQuickTask,
				QuickTask:    qt,
				QuickTaskIdx: i,
			})
		}
	}

	return rows
}

// ExpandAll returns a TreeModel with all phases and quick tasks section expanded.
func (t TreeModel) ExpandAll() TreeModel {
	for _, phase := range t.data.Phases {
		t.expanded[phase.DirName] = true
	}
	t.expanded[quickSectionKey] = true
	return t
}

// CollapseAll returns a TreeModel with all phases collapsed and cursor reset to 0.
func (t TreeModel) CollapseAll() TreeModel {
	t.expanded = make(map[string]bool)
	t.cursor = 0
	return t
}

// VisibleRows is the public version of visibleRows.
func (t TreeModel) VisibleRows() []Row {
	return t.visibleRows()
}

// Cursor returns the current cursor position.
func (t TreeModel) Cursor() int {
	return t.cursor
}

// Update handles key messages for navigation, expand, and collapse.
func (t TreeModel) Update(msg tea.Msg) (TreeModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return t, nil
	}

	rows := t.visibleRows()
	if len(rows) == 0 {
		return t, nil
	}

	switch {
	case key.Matches(keyMsg, t.keys.Down):
		t.cursor = clamp(t.cursor+1, 0, len(rows)-1)

	case key.Matches(keyMsg, t.keys.Up):
		t.cursor = clamp(t.cursor-1, 0, len(rows)-1)

	case key.Matches(keyMsg, t.keys.Expand):
		row := rows[t.cursor]
		switch row.Kind {
		case RowPhase:
			t.expanded[row.Phase.DirName] = true
		case RowQuickSection:
			t.expanded[quickSectionKey] = true
		}
		// no-op for RowPlan, RowQuickTask

	case key.Matches(keyMsg, t.keys.ExpandAll):
		return t.ExpandAll(), nil

	case key.Matches(keyMsg, t.keys.CollapseAll):
		return t.CollapseAll(), nil

	case key.Matches(keyMsg, t.keys.Collapse):
		row := rows[t.cursor]
		switch row.Kind {
		case RowPhase:
			t.expanded[row.Phase.DirName] = false
			// clamp cursor after collapsing
			newRows := t.visibleRows()
			t.cursor = clamp(t.cursor, 0, len(newRows)-1)

		case RowPlan:
			// collapse parent phase and jump cursor to the phase row
			phaseIdx := row.PhaseIdx
			phaseDirName := t.data.Phases[phaseIdx].DirName
			t.expanded[phaseDirName] = false
			// find the phase row index in the new visible rows
			newRows := t.visibleRows()
			phaseRowIdx := 0
			for i, r := range newRows {
				if r.Kind == RowPhase && r.PhaseIdx == phaseIdx {
					phaseRowIdx = i
					break
				}
			}
			t.cursor = clamp(phaseRowIdx, 0, len(newRows)-1)

		case RowQuickSection:
			t.expanded[quickSectionKey] = false
			newRows := t.visibleRows()
			t.cursor = clamp(t.cursor, 0, len(newRows)-1)

		case RowQuickTask:
			// collapse parent section and jump cursor to section header
			t.expanded[quickSectionKey] = false
			newRows := t.visibleRows()
			for i, r := range newRows {
				if r.Kind == RowQuickSection {
					t.cursor = i
					break
				}
			}
		}
	}

	return t, nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
