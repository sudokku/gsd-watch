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
	// viewport height should be 24 - header(3) - footer(2) = 19
	got := m.ViewportHeight()
	if got != 19 {
		t.Errorf("expected viewport height 19, got %d", got)
	}
}

func TestWindowSizeTiny(t *testing.T) {
	m := newTestModel()
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 3})
	// 3 - 3 - 2 = -2; clamped to 0
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

func TestQuitQ(t *testing.T) {
	m := newTestModel()
	_, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	// tea.Quit returns a special QuitMsg; verify by executing the cmd and checking the result.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected cmd() to return tea.QuitMsg, got %T", msg)
	}
}

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
