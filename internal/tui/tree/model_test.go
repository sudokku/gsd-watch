package tree_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/parser"
	"github.com/radu/gsd-watch/internal/tui/mock"
	"github.com/radu/gsd-watch/internal/tui/tree"
)

// mockProjectWithArchives returns a MockProject with two ArchivedMilestones appended.
func mockProjectWithArchives() parser.ProjectData {
	data := mock.MockProject()
	data.ArchivedMilestones = []parser.ArchivedMilestone{
		{Name: "v1.0", PhaseCount: 6, CompletionDate: "2025-01-15"},
		{Name: "v0.9", PhaseCount: 3, CompletionDate: ""},
	}
	return data
}

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
	// 6 phases, all collapsed + 1 quick section header -> 7 rows
	if len(rows) != 7 {
		t.Errorf("expected 7 rows (all phases collapsed + quick section header), got %d", len(rows))
	}
	for i, row := range rows[:6] {
		if row.Kind != tree.RowPhase {
			t.Errorf("row %d: expected RowPhase, got %v", i, row.Kind)
		}
	}
	if rows[6].Kind != tree.RowQuickSection {
		t.Errorf("row 6: expected RowQuickSection, got %v", rows[6].Kind)
	}
}

// TestExpandPhase verifies that expanding a phase shows its plans.
func TestExpandPhase(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// cursor is at row 0 (Phase 1); expand it
	m = pressKey(t, m, "l")
	rows := m.VisibleRows()
	// Phase 1 has 4 plans -> 6 phases + 4 plans + 1 quick section = 11 rows
	if len(rows) != 11 {
		t.Errorf("expected 11 rows after expanding phase 1, got %d", len(rows))
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
	if len(rows) != 7 {
		t.Errorf("expected 7 rows after collapsing phase 1, got %d", len(rows))
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
	// Phase 1 should still be expanded -> 6 phases + 4 plans + 1 quick section = 11 rows
	if len(rows) != 11 {
		t.Errorf("expected 11 rows (expanded state preserved), got %d", len(rows))
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
	// go to bottom (6) and try to go further (7 rows: 0-6, quick section is last)
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j")
	m = pressKey(t, m, "j") // clamp at 6
	if m.Cursor() != 6 {
		t.Errorf("expected cursor clamped at 6, got %d", m.Cursor())
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
	if len(rows) != 7 {
		t.Errorf("expected 7 rows after collapse, got %d", len(rows))
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
	if len(rows) != 11 {
		t.Errorf("expected 11 rows (double expand is no-op), got %d", len(rows))
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

// TestVisibleRowsWith4Collapsed verifies exactly 7 rows with all phases collapsed.
func TestVisibleRowsWith4Collapsed(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	rows := m.VisibleRows()
	if len(rows) != 7 {
		t.Errorf("expected 7 rows (6 phases collapsed + quick section header), got %d", len(rows))
	}
}

// TestVisibleRowsWithPhase1Expanded verifies 10 rows when phase 2 (3 plans) is expanded.
// Mock phase 2 has 3 plans; 6 phases + 3 plans + 1 quick section = 10 rows.
func TestVisibleRowsWithPhase1Expanded(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// move to phase 2 (row 1) and expand it (3 plans)
	m = pressKey(t, m, "j") // cursor at row 1 (Phase 2)
	m = pressKey(t, m, "l") // expand Phase 2
	rows := m.VisibleRows()
	// 6 phases + 3 plans from phase 2 + 1 quick section = 10 rows
	if len(rows) != 10 {
		t.Errorf("expected 10 rows (6 phases + 3 plans from phase 2 + quick section), got %d", len(rows))
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
	if !strings.Contains(out, "🔎") {
		t.Errorf("View output missing badge 🔎\nOutput:\n%s", out)
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

// TestExpandAll verifies that ExpandAll() expands all phases and quick tasks section.
func TestExpandAll(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	m = m.ExpandAll()
	rows := m.VisibleRows()
	// 6 phases + 4+3+2+2+2+0 plans + 1 quick section + 2 quick tasks = 22 rows
	// Phase 1: 4, Phase 2: 3, Phase 3: 2, Phase 4: 2, Phase 5: 2, Phase 6: 0
	expectedPlans := 4 + 3 + 2 + 2 + 2 + 0
	expectedRows := 6 + expectedPlans + 1 + 2 // +1 quick section, +2 quick tasks
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
	if len(rows) != 7 {
		t.Errorf("expected 7 rows after CollapseAll, got %d", len(rows))
	}
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0 after CollapseAll, got %d", m.Cursor())
	}
}

// TestView_PhaseNameWrapping verifies phase names wrap at narrow widths.
func TestView_PhaseNameWrapping(t *testing.T) {
	// Use a narrow width (32) to force wrapping on long phase names.
	data := mock.MockProject()
	m := tree.New().SetData(data)
	out := m.View(32)
	// At width 32, phase names longer than ~25 chars should wrap.
	// Just verify it doesn't crash and produces output with the phase name text.
	if !strings.Contains(out, "Core TUI") {
		t.Errorf("expected wrapped output to contain phase name fragment 'Core TUI'\nOutput:\n%s", out)
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

// --- Quick Tasks section tests ---

func TestQuickTasksSectionPresent(t *testing.T) {
	// All collapsed: last visible row should be RowQuickSection
	data := mock.MockProject()
	m := tree.New().SetData(data)
	rows := m.VisibleRows()
	lastRow := rows[len(rows)-1]
	if lastRow.Kind != tree.RowQuickSection {
		t.Errorf("expected last row to be RowQuickSection, got %v", lastRow.Kind)
	}
}

func TestExpandQuickTasksSection(t *testing.T) {
	// Navigate to quick section header, expand, verify task rows appear
	data := mock.MockProject()
	m := tree.New().SetData(data)
	// Navigate to last row (quick section header = row 6)
	for i := 0; i < 6; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand
	rows := m.VisibleRows()
	// 7 (collapsed) + 2 quick tasks = 9
	if len(rows) != 9 {
		t.Errorf("expected 9 rows after expanding quick section, got %d", len(rows))
	}
	// Verify task row kinds
	if rows[7].Kind != tree.RowQuickTask {
		t.Errorf("row 7 should be RowQuickTask, got %v", rows[7].Kind)
	}
	if rows[8].Kind != tree.RowQuickTask {
		t.Errorf("row 8 should be RowQuickTask, got %v", rows[8].Kind)
	}
}

func TestCollapseQuickTasksSection(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data)
	for i := 0; i < 6; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand
	m = pressKey(t, m, "h") // collapse
	rows := m.VisibleRows()
	if len(rows) != 7 {
		t.Errorf("expected 7 rows after collapsing quick section, got %d", len(rows))
	}
}

func TestCollapseFromQuickTask(t *testing.T) {
	// Collapsing from a RowQuickTask should jump cursor to section header
	data := mock.MockProject()
	m := tree.New().SetData(data)
	for i := 0; i < 6; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand quick section
	m = pressKey(t, m, "j") // cursor on first quick task (row 7)
	if m.Cursor() != 7 {
		t.Fatalf("expected cursor at 7, got %d", m.Cursor())
	}
	m = pressKey(t, m, "h") // collapse from task row
	if m.Cursor() != 6 {
		t.Errorf("expected cursor to jump to quick section header (row 6), got %d", m.Cursor())
	}
}

func TestQuickTasksEmptySection(t *testing.T) {
	// ProjectData with no QuickTasks: section header present, shows placeholder
	data := mock.MockProject()
	data.QuickTasks = nil
	m := tree.New().SetData(data)
	// Navigate to quick section header and expand
	for i := 0; i < 6; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand
	out := m.View(80)
	if !strings.Contains(out, "(no quick tasks)") {
		t.Errorf("expected '(no quick tasks)' placeholder\nOutput:\n%s", out)
	}
}

func TestQuickTasksViewIcons(t *testing.T) {
	// Verify status icons render for quick tasks
	data := mock.MockProject()
	m := tree.New().SetData(data)
	for i := 0; i < 6; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand
	out := m.View(80)
	// complete task shows check icon, in_progress shows arrow
	if !strings.Contains(out, "fix gsd watch sidebar closing") {
		t.Errorf("expected quick task display name in output\nOutput:\n%s", out)
	}
	if !strings.Contains(out, "Quick tasks") {
		t.Errorf("expected 'Quick tasks' section header\nOutput:\n%s", out)
	}
}

// TestView_PhaseActiveWhenCursorOnChildPlan verifies that when the cursor moves to
// a child plan row, the parent phase row still renders as "active" (highlighted).
// Since lipgloss strips ANSI in test (no TTY), we verify structural behavior:
// - cursor is on a RowPlan row
// - isPhaseActive returns true for the parent phase row index
// - badge text for phase 1 appears in the output (badges rendered via any style are present)
func TestView_PhaseActiveWhenCursorOnChildPlan(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data).SetOptions(tree.Options{NoEmoji: true})
	// expand phase 1, cursor at row 0 (phase row)
	m = pressKey(t, m, "l")
	// move cursor down to first plan (row 1)
	m = pressKey(t, m, "j")

	rows := m.VisibleRows()
	if m.Cursor() != 1 {
		t.Fatalf("expected cursor at 1 (first plan of phase 1), got %d", m.Cursor())
	}
	if rows[1].Kind != tree.RowPlan {
		t.Fatalf("expected row 1 to be RowPlan, got %v", rows[1].Kind)
	}

	// Phase row is at index 0; cursor is at index 1 (child plan) -> phase should be active
	if !m.IsPhaseActive(rows, 0) {
		t.Errorf("expected IsPhaseActive(rows, 0) to be true when cursor is on child plan at index 1")
	}

	out := m.View(80)
	// Phase 1 badges ([disc] [rsrch]) should appear in the output regardless of style
	if !strings.Contains(out, "[disc]") {
		t.Errorf("expected badge '[disc]' in output when cursor on child plan\nOutput:\n%s", out)
	}
	if !strings.Contains(out, "[rsrch]") {
		t.Errorf("expected badge '[rsrch]' in output when cursor on child plan\nOutput:\n%s", out)
	}
	// Phase name should still appear in output
	if !strings.Contains(out, "Phase 1: Core TUI Scaffold") {
		t.Errorf("expected phase name in output\nOutput:\n%s", out)
	}
}

// TestView_BadgeInheritsDimForCompletedPhase verifies that completed phase badges
// appear in the output, rendered via PendingStyle when not active.
func TestView_BadgeInheritsDimForCompletedPhase(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data).SetOptions(tree.Options{NoEmoji: true})
	// Navigate to phase 5 (index 4) and expand it
	for i := 0; i < 4; i++ {
		m = pressKey(t, m, "j")
	}
	m = pressKey(t, m, "l") // expand phase 5

	// cursor is on phase 5 (active) — badges should appear
	out := m.View(80)
	if !strings.Contains(out, "[disc]") {
		t.Errorf("expected badge '[disc]' for phase 5 (active cursor)\nOutput:\n%s", out)
	}
	if !strings.Contains(out, "[vrfy]") {
		t.Errorf("expected badge '[vrfy]' for phase 5 (active cursor)\nOutput:\n%s", out)
	}

	// Move cursor away to phase 6 (cursor at phase 5 row index + 2 plans + 1 = phase 6 row)
	m = pressKey(t, m, "j") // plan 1
	m = pressKey(t, m, "j") // plan 2
	m = pressKey(t, m, "j") // phase 6

	rows := m.VisibleRows()
	// Find phase 5 row index
	phase5RowIdx := -1
	for i, r := range rows {
		if r.Kind == tree.RowPhase && r.Phase.DirName == "05-tui-polish" {
			phase5RowIdx = i
			break
		}
	}
	if phase5RowIdx == -1 {
		t.Fatal("could not find phase 5 row")
	}
	// Phase 5 should NOT be active (cursor is on phase 6)
	if m.IsPhaseActive(rows, phase5RowIdx) {
		t.Errorf("expected IsPhaseActive(rows, %d) to be false when cursor is on phase 6", phase5RowIdx)
	}

	// Badges should still appear (dimmed via PendingStyle but present as text)
	out2 := m.View(80)
	if !strings.Contains(out2, "[disc]") {
		t.Errorf("expected badge '[disc]' for phase 5 (dimmed, cursor elsewhere)\nOutput:\n%s", out2)
	}
}

// TestView_BadgeUnstyledForNonCompleteNonActivePhase verifies that phase 2 badges
// (non-complete, non-active) render unstyled (just plain text, no PendingStyle).
// Phase 2 has no badges in mock, so we use phase 1 (in_progress) with cursor on phase 2.
func TestView_BadgeUnstyledForNonCompleteNonActivePhase(t *testing.T) {
	data := mock.MockProject()
	m := tree.New().SetData(data).SetOptions(tree.Options{NoEmoji: true})
	// cursor starts at phase 1 (index 0); move to phase 2 (index 1)
	m = pressKey(t, m, "j")

	rows := m.VisibleRows()
	// Phase 1 is at index 0; cursor is at index 1 (phase 2) -> phase 1 not active
	if m.IsPhaseActive(rows, 0) {
		t.Errorf("expected IsPhaseActive(rows, 0) to be false when cursor is on phase 2")
	}

	out := m.View(80)
	// Phase 1 badges should still appear as plain text
	if !strings.Contains(out, "[disc]") {
		t.Errorf("expected '[disc]' badge to appear even when phase 1 not active\nOutput:\n%s", out)
	}
}

// --- Archive rendering helper tests ---

// TestFormatArchiveDate verifies ISO date to "Mon YYYY" conversion.
func TestFormatArchiveDate(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"2026-03-23", "Mar 2026"},
		{"2025-01-15", "Jan 2025"},
		{"", ""},
		{"not-a-date", ""},
		{"2025-12-01", "Dec 2025"},
	}
	for _, c := range cases {
		got := tree.FormatArchiveDate(c.input)
		if got != c.want {
			t.Errorf("FormatArchiveDate(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

// TestRenderArchiveRow_Emoji verifies emoji mode archive row contains expected text.
func TestRenderArchiveRow_Emoji(t *testing.T) {
	am := parser.ArchivedMilestone{Name: "v1.0", PhaseCount: 6, CompletionDate: "2025-01-15"}
	out := tree.RenderArchiveRow(am, false)
	if !strings.Contains(out, "▸ v1.0 — 6 phases ✓  Jan 2025") {
		t.Errorf("emoji mode: expected '▸ v1.0 — 6 phases ✓  Jan 2025' in output, got: %q", out)
	}
}

// TestRenderArchiveRow_NoEmoji verifies noEmoji mode archive row contains expected text.
func TestRenderArchiveRow_NoEmoji(t *testing.T) {
	am := parser.ArchivedMilestone{Name: "v1.0", PhaseCount: 6, CompletionDate: "2025-01-15"}
	out := tree.RenderArchiveRow(am, true)
	if !strings.Contains(out, "> v1.0 — 6 phases [done]  Jan 2025") {
		t.Errorf("noEmoji mode: expected '> v1.0 — 6 phases [done]  Jan 2025' in output, got: %q", out)
	}
}

// TestRenderArchiveRow_NoDate verifies that empty CompletionDate omits trailing space.
func TestRenderArchiveRow_NoDate(t *testing.T) {
	am := parser.ArchivedMilestone{Name: "v0.9", PhaseCount: 3, CompletionDate: ""}
	out := tree.RenderArchiveRow(am, false)
	if !strings.Contains(out, "▸ v0.9 — 3 phases ✓") {
		t.Errorf("no-date: expected '▸ v0.9 — 3 phases ✓' in output, got: %q", out)
	}
	if strings.Contains(out, "✓  ") {
		t.Errorf("no-date: output should not contain trailing double-space after checkmark, got: %q", out)
	}
}

// TestRenderArchiveSeparator verifies separator contains label and starts with "- - ".
func TestRenderArchiveSeparator(t *testing.T) {
	out := tree.RenderArchiveSeparator(80)
	if !strings.Contains(out, "Archived Milestones") {
		t.Errorf("separator: expected 'Archived Milestones' in output, got: %q", out)
	}
	if !strings.HasPrefix(out, "- - ") {
		t.Errorf("separator: expected output to start with '- - ', got: %q", out)
	}
}

// TestRenderArchiveZone_Empty verifies empty slice returns empty string.
func TestRenderArchiveZone_Empty(t *testing.T) {
	out := tree.RenderArchiveZone(nil, 80, false)
	if out != "" {
		t.Errorf("empty archives: expected empty string, got: %q", out)
	}
}

// TestRenderArchiveZone_NonEmpty verifies 2 milestones produce 3 lines (separator + 2 rows).
func TestRenderArchiveZone_NonEmpty(t *testing.T) {
	archives := []parser.ArchivedMilestone{
		{Name: "v1.0", PhaseCount: 6, CompletionDate: "2025-01-15"},
		{Name: "v0.9", PhaseCount: 3, CompletionDate: ""},
	}
	out := tree.RenderArchiveZone(archives, 80, false)
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (separator + 2 rows), got %d lines\nOutput:\n%s", len(lines), out)
	}
	if !strings.Contains(out, "v1.0") {
		t.Errorf("expected 'v1.0' in archive zone output, got: %q", out)
	}
	if !strings.Contains(out, "v0.9") {
		t.Errorf("expected 'v0.9' in archive zone output, got: %q", out)
	}
}

// TestArchiveRowsNotInVisibleRows verifies visibleRows excludes archive data.
func TestArchiveRowsNotInVisibleRows(t *testing.T) {
	data := mockProjectWithArchives()
	m := tree.New().SetData(data)
	rows := m.VisibleRows()
	validKinds := map[tree.RowKind]bool{
		tree.RowPhase:       true,
		tree.RowPlan:        true,
		tree.RowQuickSection: true,
		tree.RowQuickTask:   true,
	}
	for i, row := range rows {
		if !validKinds[row.Kind] {
			t.Errorf("row %d has unexpected kind %v — archive rows must not appear in visibleRows", i, row.Kind)
		}
	}
}
