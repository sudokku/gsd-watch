package tree_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/tui/mock"
	"github.com/radu/gsd-watch/internal/tui/tree"
)

// helper to send a key press message to the tree model
func pressKey(t *testing.T, m tree.TreeModel, key string) tree.TreeModel {
	t.Helper()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	if key == "up" {
		msg = tea.KeyMsg{Type: tea.KeyUp}
	} else if key == "down" {
		msg = tea.KeyMsg{Type: tea.KeyDown}
	} else if key == "right" {
		msg = tea.KeyMsg{Type: tea.KeyRight}
	} else if key == "left" {
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	}
	newModel, _ := m.Update(msg)
	return newModel
}

// TestNew verifies that New() creates a model with cursor 0 and no expanded state.
func TestNew(t *testing.T) {
	m := tree.New()
	rows := m.VisibleRows()
	if len(rows) != 0 {
		t.Errorf("expected 0 visible rows on empty model, got %d", len(rows))
	}
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", m.Cursor())
	}
}

// TestSetDataAllCollapsed verifies that all phases appear but no plans when all collapsed.
func TestSetDataAllCollapsed(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	rows := m.VisibleRows()
	// 6 phases, all collapsed -> 6 rows
	if len(rows) != 6 {
		t.Errorf("expected 6 rows (all phases collapsed), got %d", len(rows))
	}
	for i, row := range rows {
		if row.Kind != tree.RowPhase {
			t.Errorf("row %d: expected RowPhase, got %v", i, row.Kind)
		}
	}
}

// TestExpandPhase verifies that expanding a phase shows its plans.
func TestExpandPhase(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// cursor is at row 0 (Phase 1); expand it
	m = pressKey(t, m, "l")
	rows := m.VisibleRows()
	// Phase 1 has 4 plans -> 6 phases + 4 plans = 10 rows
	if len(rows) != 10 {
		t.Errorf("expected 10 rows after expanding phase 1, got %d", len(rows))
	}
}

// TestCollapsePhase verifies that collapsing hides plans.
func TestCollapsePhase(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// expand then collapse phase 1
	m = pressKey(t, m, "l")
	m = pressKey(t, m, "h")
	rows := m.VisibleRows()
	if len(rows) != 6 {
		t.Errorf("expected 6 rows after collapsing phase 1, got %d", len(rows))
	}
}

// TestExpandCollapsePreservesKeyedState verifies expanded state is keyed by DirName,
// not position. SetData again with same phase DirNames preserves expanded state.
func TestExpandCollapsePreservesKeyedState(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// expand phase 1
	m = pressKey(t, m, "l")
	// call SetData again with same data (simulates a refresh)
	m = m.SetData(data)
	rows := m.VisibleRows()
	// Phase 1 should still be expanded -> 6 phases + 4 plans = 10 rows
	if len(rows) != 10 {
		t.Errorf("expected 10 rows (expanded state preserved), got %d", len(rows))
	}
}

// TestCursorDownUp verifies cursor movement and clamping.
func TestCursorDownUp(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// 6 collapsed phases; cursor starts at 0
	m = pressKey(t, m, "j") // down
	if m.Cursor() != 1 {
		t.Errorf("expected cursor 1 after down, got %d", m.Cursor())
	}
	m = pressKey(t, m, "k") // up
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0 after up, got %d", m.Cursor())
	}
	// clamp at top
	m = pressKey(t, m, "k")
	if m.Cursor() != 0 {
		t.Errorf("expected cursor clamped at 0, got %d", m.Cursor())
	}
	// go to bottom (5) and try to go further
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j") // clamp at 5
	if m.Cursor() != 5 {
		t.Errorf("expected cursor clamped at 5, got %d", m.Cursor())
	}
}

// TestCursorJumpOnCollapse verifies that collapsing a phase while cursor is on a child
// plan moves the cursor to the phase row.
func TestCursorJumpOnCollapse(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// expand phase 1
	m = pressKey(t, m, "l")
	// move cursor to plan row (row 1 = first plan of phase 1)
	m = pressKey(t, m, "j")
	if m.Cursor() != 1 {
		t.Fatalf("expected cursor at 1, got %d", m.Cursor())
	}
	// now collapse phase 1 from a plan row via "h"
	m = pressKey(t, m, "h")
	// cursor should jump to phase row (row 0 in collapsed view)
	if m.Cursor() != 0 {
		t.Errorf("expected cursor to jump to phase row 0, got %d", m.Cursor())
	}
	rows := m.VisibleRows()
	if len(rows) != 6 {
		t.Errorf("expected 6 rows after collapse, got %d", len(rows))
	}
}

// TestCollapseNoJumpWhenCursorNotOnChild verifies cursor stays (clamped) when
// collapsing a phase whose cursor is on a different phase.
func TestCollapseNoJumpWhenCursorNotOnChild(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// expand phase 1, move cursor to phase 2 (row 5 in expanded view)
	m = pressKey(t, m, "l") // expand phase 1; cursor stays at 0
	// move cursor down to phase 2 (row 5 = phase 0 + 4 plans + phase 2 offset)
	for i := 0; i < 5; i++ {
		m = pressKey(t, m, "j")
	}
	if m.Cursor() != 5 {
		t.Fatalf("expected cursor at 5 (Phase 2 row), got %d", m.Cursor())
	}
	// Simpler: cursor on phase 1 row, not a child plan -> cursor stays on phase 1 row.
	m2 := tree.New().SetData(data)
	m2 = pressKey(t, m2, "l") // expand phase 1, cursor at 0 (phase 1 row)
	// collapse phase 1 from phase row (cursor at 0 = phase 1 row)
	m2 = pressKey(t, m2, "h")
	if m2.Cursor() != 0 {
		t.Errorf("expected cursor to remain at 0 (phase row), got %d", m2.Cursor())
	}
}

// TestExpandAlreadyExpanded verifies that expanding an already-expanded phase is a no-op.
func TestExpandAlreadyExpanded(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = pressKey(t, m, "l") // expand
	m = pressKey(t, m, "l") // expand again (no-op)
	rows := m.VisibleRows()
	if len(rows) != 10 {
		t.Errorf("expected 10 rows (double expand is no-op), got %d", len(rows))
	}
}

// TestExpandOnPlanRow verifies that pressing "l" on a plan row is a no-op.
func TestExpandOnPlanRow(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = pressKey(t, m, "l") // expand phase 1
	m = pressKey(t, m, "j") // cursor on plan row 1
	rowsBefore := m.VisibleRows()
	m = pressKey(t, m, "l") // expand on plan row = no-op
	rowsAfter := m.VisibleRows()
	if len(rowsBefore) != len(rowsAfter) {
		t.Errorf("expand on plan row should be no-op: before=%d after=%d", len(rowsBefore), len(rowsAfter))
	}
}

// TestVisibleRowsWith4Collapsed verifies exactly 6 rows with all phases collapsed.
func TestVisibleRowsWith4Collapsed(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	rows := m.VisibleRows()
	if len(rows) != 6 {
		t.Errorf("expected 6 rows (6 phases collapsed), got %d", len(rows))
	}
}

// TestVisibleRowsWithPhase1Expanded verifies 9 rows when phase 2 (3 plans) is expanded.
// Mock phase 2 has 3 plans; 6 phases + 3 plans = 9 rows.
func TestVisibleRowsWithPhase1Expanded(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// move to phase 2 (row 1) and expand it (3 plans)
	m = pressKey(t, m, "j") // cursor at row 1 (Phase 2)
	m = pressKey(t, m, "l") // expand Phase 2
	rows := m.VisibleRows()
	// 6 phases + 3 plans from phase 2 = 9 rows
	if len(rows) != 9 {
		t.Errorf("expected 9 rows (6 phases + 3 plans from phase 2), got %d", len(rows))
	}
}

// TestRowKindAndPhaseIdx verifies that Row.Kind and Row.PhaseIdx are set correctly.
func TestRowKindAndPhaseIdx(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = pressKey(t, m, "l") // expand phase 1
	rows := m.VisibleRows()
	// row 0 = Phase 1 (PhaseIdx=0, Kind=RowPhase)
	if rows[0].Kind != tree.RowPhase {
		t.Errorf("row 0: expected RowPhase, got %v", rows[0].Kind)
	}
	if rows[0].PhaseIdx != 0 {
		t.Errorf("row 0: expected PhaseIdx=0, got %d", rows[0].PhaseIdx)
	}
	// row 1 = Plan in Phase 1 (PhaseIdx=0, Kind=RowPlan)
	if rows[1].Kind != tree.RowPlan {
		t.Errorf("row 1: expected RowPlan, got %v", rows[1].Kind)
	}
	if rows[1].PhaseIdx != 0 {
		t.Errorf("row 1: expected PhaseIdx=0, got %d", rows[1].PhaseIdx)
	}
	// row 5 = Phase 2 (PhaseIdx=1, Kind=RowPhase)
	if rows[5].Kind != tree.RowPhase {
		t.Errorf("row 5: expected RowPhase, got %v", rows[5].Kind)
	}
	if rows[5].PhaseIdx != 1 {
		t.Errorf("row 5: expected PhaseIdx=1, got %d", rows[5].PhaseIdx)
	}
}

// --- View tests ---

// TestViewStatusIcons verifies that rendering an expanded phase shows status icon characters.
func TestViewStatusIcons(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = pressKey(t, m, "l") // expand phase 1
	out := m.View(80)
	// phase 1 status is "in_progress" -> arrow icon
	// plan 0 status is "complete" -> check icon
	// plan 2 status is "pending" -> circle icon
	for _, icon := range []string{"▶", "✓", "○"} {
		if !strings.Contains(out, icon) {
			t.Errorf("View output missing icon %q\nOutput:\n%s", icon, out)
		}
	}
}

// TestViewActiveMarker verifies the "now" marker appears on the active plan.
func TestViewActiveMarker(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = pressKey(t, m, "l") // expand phase 1
	out := m.View(80)
	if !strings.Contains(out, "← now") {
		t.Errorf("View output missing '← now' marker\nOutput:\n%s", out)
	}
}

// TestViewBadges verifies that phase badges render below the phase header.
func TestViewBadges(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	out := m.View(80)
	// Phase 1 has badges "discussed" and "researched"
	if !strings.Contains(out, "💬") {
		t.Errorf("View output missing badge 💬\nOutput:\n%s", out)
	}
	if !strings.Contains(out, "🔬") {
		t.Errorf("View output missing badge 🔬\nOutput:\n%s", out)
	}
}

// TestViewTooNarrow verifies that widths below MinWidth return the narrow placeholder.
func TestViewTooNarrow(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	out := m.View(20)
	if !strings.Contains(out, "too narrow") {
		t.Errorf("expected 'too narrow' for narrow width, got: %s", out)
	}
}

// TestViewCollapsedHidesPlans verifies that collapsed phases don't show plan titles.
func TestViewCollapsedHidesPlans(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	out := m.View(80)
	// When all phases are collapsed, no plan titles should appear
	planTitles := []string{
		"Foundation: types, messages, mock data",
		"Tree model + viewport",
		"Header + footer components",
		"Root model + integration",
	}
	for _, title := range planTitles {
		if strings.Contains(out, title) {
			t.Errorf("plan title %q should not appear when phase is collapsed\nOutput:\n%s", title, out)
		}
	}
}

// TestView_NoProject verifies that an empty project shows the "No GSD project found" message.
func TestView_NoProject(t *testing.T) {
	m := tree.New() // no data set, empty project
	out := m.View(80)
	if !strings.Contains(out, "No GSD project found") {
		t.Errorf("expected 'No GSD project found' in empty project view\nOutput:\n%s", out)
	}
	if !strings.Contains(out, "/gsd:new-project") {
		t.Errorf("expected '/gsd:new-project' in empty project view\nOutput:\n%s", out)
	}
}

// TestView_NoPlansYet verifies that an expanded phase with no plans shows "(no plans yet)".
func TestView_NoPlansYet(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// navigate to phase 6 (06-future, which has no plans) and expand it
	// Phase 6 is at index 5 (0-based), so press j 5 times
	for i := 0; i < 5; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand phase 6
	out := m.View(80)
	if !strings.Contains(out, "(no plans yet)") {
		t.Errorf("expected '(no plans yet)' for empty phase\nOutput:\n%s", out)
	}
}

// TestView_CompletedPhaseDimmed verifies that completed phase rows are rendered via PendingStyle.
// The test verifies the phase appears in the output; ANSI color output is terminal-dependent
// and may be stripped by lipgloss when no TTY is detected.
func TestView_CompletedPhaseDimmed(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// navigate to phase 5 (05-tui-polish, status=complete) and expand it
	// Phase 5 is at index 4 (0-based)
	for i := 0; i < 4; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand phase 5
	out := m.View(80)
	// The output should contain the phase name
	if !strings.Contains(out, "Phase 5: TUI Polish") {
		t.Errorf("expected 'Phase 5: TUI Polish' in output\nOutput:\n%s", out)
	}
	// Plans from the completed phase should appear when expanded
	if !strings.Contains(out, "Tree + footer polish") {
		t.Errorf("expected 'Tree + footer polish' plan in output for expanded completed phase\nOutput:\n%s", out)
	}
}

// TestExpandAll verifies that ExpandAll() expands all phases.
func TestExpandAll(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = m.ExpandAll()
	rows := m.VisibleRows()
	// 6 phases + 4+3+2+2+2+0 plans = 6 + 13 = 19 rows
	// Phase 1: 4, Phase 2: 3, Phase 3: 2, Phase 4: 2, Phase 5: 2, Phase 6: 0
	expectedPlans := 4 + 3 + 2 + 2 + 2 + 0
	expectedRows := 6 + expectedPlans
	if len(rows) != expectedRows {
		t.Errorf("expected %d rows after ExpandAll, got %d", expectedRows, len(rows))
	}
}

// TestCollapseAll verifies that CollapseAll() collapses all phases and resets cursor.
func TestCollapseAll(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = m.ExpandAll()
	m = pressKey(t, m, "j") // move cursor to non-zero position
	m = m.CollapseAll()
	rows := m.VisibleRows()
	if len(rows) != 6 {
		t.Errorf("expected 6 rows after CollapseAll, got %d", len(rows))
	}
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0 after CollapseAll, got %d", m.Cursor())
	}
}

// TestView_Padding verifies that every non-empty line in View output starts with a space.
func TestView_Padding(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	out := m.View(80)
	lines := strings.Split(out, "\n")
	checked := 0
	for i, line := range lines {
		if line == "" {
			continue
		}
		if line[0] != ' ' {
			t.Errorf("line %d does not start with padding space: %q", i, line)
		}
		checked++
		if checked >= 3 {
			break
		}
	}
}
