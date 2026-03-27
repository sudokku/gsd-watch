// model_test.go is a thin adapter in the tui package that exercises
// the root model via the app sub-package.
// It lives here so `go test ./internal/tui/` covers the integration.
package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radu/gsd-watch/internal/config"
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
// Uses config.Defaults() which has Emoji=true (emoji rendering behavior).
func newTestModel() app.Model {
	return app.New(make(chan tea.Msg, 10), config.Defaults())
}

// newTestModelNoEmoji creates a Model with Emoji=false for ASCII rendering tests.
func newTestModelNoEmoji() app.Model {
	cfg := config.Defaults()
	cfg.Emoji = false
	return app.New(make(chan tea.Msg, 10), cfg)
}

func TestWindowSizeNormal(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// viewport height should be 24 - header(4) - footer(5) = 15
	got := m.ViewportHeight()
	if got != 15 {
		t.Errorf("expected viewport height 15, got %d", got)
	}
}

func TestWindowSizeTiny(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 3})
	// 3 - 4 - 5 = -6; clamped to 0
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
	// First q: schedules timeout tick (non-nil cmd), does not quit.
	m, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected tick cmd on first q (quit-pending), got nil")
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
	if cmd == nil {
		t.Fatal("expected tick cmd on first Esc (quit-pending), got nil")
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

// TestQuit_TimeoutResets: QuitTimeoutMsg clears pending so next q starts fresh.
func TestQuit_TimeoutResets(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}) // pending=true
	m, _ = updateModel(m, tui.QuitTimeoutMsg{})                               // timeout fires
	// Next q should start a new pending (non-nil tick cmd), not quit.
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected tick cmd after timeout reset, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Error("expected timeout reset to prevent immediate quit on next q")
	}
}

// TestQuit_QResetByOtherKey: q then j resets quitPending, so next q restarts the
// confirm window (tick cmd) rather than quitting immediately.
func TestQuit_QResetByOtherKey(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}) // pending=true
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}) // reset
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}) // new pending
	if cmd == nil {
		t.Fatal("expected tick cmd on q after reset, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Error("expected restart of confirm window (not quit) after j reset")
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

// TestHelpOverlay_ContainsPhaseStages: help overlay includes Phase stages section.
func TestHelpOverlay_ContainsPhaseStages(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view := m.View()
	if !strings.Contains(view, "Phase stages") {
		t.Error("expected help overlay to contain 'Phase stages' section")
	}
	if !strings.Contains(view, "💬") {
		t.Error("expected help overlay to contain discussed badge 💬")
	}
	if !strings.Contains(view, "🧪") {
		t.Error("expected help overlay to contain UAT badge 🧪")
	}
}

// TestNoEmoji_TreeRenders_ASCIIIcons: noEmoji=true renders [x] instead of checkmark emoji.
func TestNoEmoji_TreeRenders_ASCIIIcons(t *testing.T) {
	m := newTestModelNoEmoji()
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 40})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	view := m.View()
	if !strings.Contains(view, "[x]") {
		t.Error("expected noEmoji tree to contain '[x]' for complete status")
	}
	// The footer always renders ✓ as its idle indicator, so we only verify that
	// the tree uses ASCII [x] for complete phases — not that ✓ is absent everywhere.
}

// TestNoEmoji_TreeRenders_ASCIIBadges: noEmoji=true renders [disc]/[plan]/etc. instead of emoji badges.
func TestNoEmoji_TreeRenders_ASCIIBadges(t *testing.T) {
	m := newTestModelNoEmoji()
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 40})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	view := m.View()
	hasBadge := strings.Contains(view, "[disc]") ||
		strings.Contains(view, "[plan]") ||
		strings.Contains(view, "[exec]") ||
		strings.Contains(view, "[vrfy]")
	if !hasBadge {
		t.Error("expected noEmoji tree to contain ASCII badge like [disc], [plan], [exec], or [vrfy]")
	}
	// clipboard emoji (📋 planned badge) should not appear
	if strings.Contains(view, "📋") {
		t.Error("expected noEmoji tree to NOT contain clipboard emoji '📋'")
	}
}

// TestNoEmoji_HelpOverlay_ASCIIBadges: noEmoji=true shows ASCII brackets in help overlay.
func TestNoEmoji_HelpOverlay_ASCIIBadges(t *testing.T) {
	m := newTestModelNoEmoji()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 40})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view := m.View()
	if !strings.Contains(view, "[disc]") {
		t.Error("expected noEmoji help overlay to contain '[disc]'")
	}
	if !strings.Contains(view, "[uat]") {
		t.Error("expected noEmoji help overlay to contain '[uat]'")
	}
	// speech balloon emoji should not appear
	if strings.Contains(view, "💬") {
		t.Error("expected noEmoji help overlay to NOT contain speech balloon '💬'")
	}
}

// TestNoEmoji_False_RendersEmoji: default model (noEmoji=false) renders emoji icons not ASCII.
func TestNoEmoji_False_RendersEmoji(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tui.ParsedMsg{Project: mock.MockProject()})
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 40})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	view := m.View()
	// Should NOT contain ASCII "[x]" for complete status; should use "✓"
	if strings.Contains(view, "[x]") {
		t.Error("expected default (emoji) tree to NOT contain '[x]'; should use emoji checkmark")
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
