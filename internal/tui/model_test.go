// model_test.go is a thin adapter in the tui package that exercises
// the root model via the app sub-package.
// It lives here so `go test ./internal/tui/` covers the integration.
package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/tui"
	"github.com/radu/gsd-watch/internal/tui/app"
	"github.com/radu/gsd-watch/internal/tui/mock"
)

// helper: call Update and cast back to app.Model
func updateModel(m app.Model, msg tea.Msg) (app.Model, tea.Cmd) {
	newModel, cmd := m.Update(msg)
	return newModel.(app.Model), cmd
}

// newTestModel creates a Model with a buffered events channel for testing.
// Tests never send to the channel, so it just needs to be non-nil.
func newTestModel() app.Model {
	return app.New(make(chan tea.Msg, 10))
}

func TestWindowSizeNormal(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// viewport height should be 24 - header(3) - footer(3) = 18
	got := m.ViewportHeight()
	if got != 18 {
		t.Errorf("expected viewport height 18, got %d", got)
	}
}

func TestWindowSizeTiny(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 3})
	// 3 - 3 - 3 = -3; clamped to 0
	got := m.ViewportHeight()
	if got != 0 {
		t.Errorf("expected viewport height 0 for tiny window, got %d", got)
	}
}

func TestWindowSizeNarrow(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 10, Height: 24})
	// Width should be stored; View handles "too narrow"
	got := m.Width()
	if got != 10 {
		t.Errorf("expected width 10, got %d", got)
	}
}

// TestQuit_DoubleQ: single q does not quit; second q does.
func TestQuit_DoubleQ(t *testing.T) {
	m := newTestModel()
	// First q: no quit
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatal("expected nil cmd on first q, got non-nil")
	}
	// Second q: quits
	_, cmd = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command on second q, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// TestQuit_DoubleEsc: single Esc does not quit; second Esc does.
func TestQuit_DoubleEsc(t *testing.T) {
	m := newTestModel()
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		t.Fatal("expected nil cmd on first Esc, got non-nil")
	}
	_, cmd = updateModel(m, tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected quit command on second Esc, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// TestQuit_QResetByOtherKey: q then j resets quitPending, so next q does not quit.
func TestQuit_QResetByOtherKey(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatal("expected nil cmd after reset by 'j', got non-nil (quitPending should have been cleared)")
	}
}

// TestQuit_CtrlCAlwaysQuits: Ctrl+C quits immediately without double-press.
func TestQuit_CtrlCAlwaysQuits(t *testing.T) {
	m := newTestModel()
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected cmd() to return tea.QuitMsg, got %T", msg)
	}
}

// TestQuitCtrlC kept for backward compat (same as TestQuit_CtrlCAlwaysQuits).
func TestQuitCtrlC(t *testing.T) {
	m := newTestModel()
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected cmd() to return tea.QuitMsg, got %T", msg)
	}
}

// TestHelpOverlay_OpenClose: '?' opens overlay; 'q' dismisses without quitting.
func TestHelpOverlay_OpenClose(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Open help
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view := m.View()
	if !strings.Contains(view, "gsd-watch help") {
		t.Error("expected help overlay to contain 'gsd-watch help'")
	}
	// Close with q (single q closes overlay, does NOT quit)
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Error("expected nil cmd when closing help overlay with q")
	}
	view = m.View()
	if strings.Contains(view, "gsd-watch help") {
		t.Error("expected help overlay to be dismissed after q")
	}
}

// TestHelpOverlay_CtrlCQuits: Ctrl+C quits even when help overlay is open.
func TestHelpOverlay_CtrlCQuits(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command from Ctrl+C with overlay open, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// TestHelpOverlay_EscCloses: Esc closes overlay without quitting.
func TestHelpOverlay_EscCloses(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		t.Error("expected nil cmd when closing overlay with Esc")
	}
	view := m.View()
	if strings.Contains(view, "gsd-watch help") {
		t.Error("expected help overlay to be dismissed after Esc")
	}
}

// TestExpandAllKey: pressing 'e' expands all phases revealing plan titles.
func TestExpandAllKey(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 40})
	// Press 'e' to expand all
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	// Check View contains plan titles that would only appear when expanded
	view := m.View()
	if !strings.Contains(view, "Foundation") {
		t.Error("expected expanded tree to contain plan title 'Foundation'")
	}
}

// TestCollapseAllKey: pressing 'w' collapses all phases hiding plan titles.
func TestCollapseAllKey(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 40})
	// Expand first, then collapse
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	// Plan titles should no longer appear
	view := m.View()
	if strings.Contains(view, "Foundation") {
		t.Error("expected collapsed tree to NOT contain plan title 'Foundation'")
	}
}

func TestKeyDelegationMovesTreeCursor(t *testing.T) {
	m := newTestModel()
	// Inject project data via ParsedMsg (new live-data path) so tree has items to navigate.
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	// Must resize first to make model ready
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	cursorBefore := m.TreeCursor()
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	cursorAfter := m.TreeCursor()
	if cursorAfter == cursorBefore {
		t.Error("expected tree cursor to move after pressing j, but it did not change")
	}
}

func TestViewContainsHeaderAndFooter(t *testing.T) {
	m := newTestModel()
	// Inject project data via ParsedMsg (new live-data path) so header has project name.
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View()
	if !strings.Contains(view, "gsd-watch") {
		t.Error("expected View() to contain project name 'gsd-watch'")
	}
	// Footer should contain key hints
	if !strings.Contains(view, "quit") {
		t.Error("expected View() to contain key hint 'quit'")
	}
}
