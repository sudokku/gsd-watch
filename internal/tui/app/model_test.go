package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/radu/gsd-watch/internal/config"
	"github.com/radu/gsd-watch/internal/parser"
	"github.com/radu/gsd-watch/internal/tui"
)

func strPtr(s string) *string { return &s }

// TestNew_WithColorOverrides verifies that New() applies color overrides from
// Config.Colors without panicking. This is the integration test for the
// ApplyColorOverrides wiring added in model.go.
func TestNew_WithColorOverrides(t *testing.T) {
	events := make(chan tea.Msg, 1)
	cfg := config.Config{
		Emoji:  true,
		Preset: "",
		Colors: config.ThemeColors{
			Complete: strPtr("#ff0000"),
		},
	}
	// Must not panic — exercises ThemeByName + ApplyColorOverrides path
	_ = New(events, cfg)
}

// TestKeyDown_WrappedRowFullyVisible verifies that pressing 'down' onto a
// multi-line wrapped quick-task row scrolls the viewport so the row's LAST
// rendered line sits at YOffset+Height-1 (and firstLine is also in view).
// Locks the bottom-bound scroll fix that uses lastLine instead of firstLine.
func TestKeyDown_WrappedRowFullyVisible(t *testing.T) {
	events := make(chan tea.Msg, 10)
	m := New(events, config.Defaults())

	// Build a project with a small phase set so the cursor reaches the wrapped
	// quick task quickly, and a long-DisplayName quick task to force wrapping.
	longName := "wrap me across three lines at width thirty two please"
	data := parser.ProjectData{
		Name: "test",
		Phases: []parser.Phase{
			{DirName: "01-p", Name: "Phase 1", Status: "in_progress", Plans: nil},
		},
		QuickTasks: []parser.QuickTask{
			{DirName: "260509-aaa-short", DisplayName: "short task", Date: "260509", Status: "in_progress"},
			{DirName: "260509-bbb-long", DisplayName: longName, Date: "260509", Status: "in_progress"},
		},
	}

	// Size the window narrow enough to force wrapping and short enough that
	// the last quick task overflows the viewport.
	width, height := 32, 14
	m2, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = m2.(Model)
	m2, _ = m.Update(tui.ParsedMsg{Project: data})
	m = m2.(Model)

	// Visible rows: [0]=Phase 1, [1]=Quick Tasks (collapsed). Expand quick section.
	// Navigate cursor to row 1 (Quick Tasks header), then press 'l' to expand.
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = m2.(Model)
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight}) // expand
	m = m2.(Model)

	// Now navigate down twice to land on the long-name quick task (row 3).
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = m2.(Model)
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = m2.(Model)

	if m.tree.Cursor() != 3 {
		t.Fatalf("expected cursor at row 3 (long quick task), got %d", m.tree.Cursor())
	}

	firstLine, lastLine := m.tree.RenderedCursorLineSpan(m.width)
	if !(lastLine > firstLine) {
		t.Fatalf("expected wrapped row (lastLine > firstLine); got firstLine=%d lastLine=%d (width=%d)", firstLine, lastLine, m.width)
	}

	yoff := m.viewport.YOffset
	vh := m.viewport.Height
	if vh <= 0 {
		t.Fatalf("viewport height non-positive: %d (windowH=%d)", vh, height)
	}

	// Bottom-bound post-condition: lastLine == YOffset + Height - 1.
	if lastLine != yoff+vh-1 {
		t.Errorf("expected lastLine == YOffset+Height-1; got lastLine=%d, YOffset=%d, Height=%d (=> %d)",
			lastLine, yoff, vh, yoff+vh-1)
	}
	// Top of cursor row is also visible.
	if firstLine < yoff {
		t.Errorf("expected firstLine >= YOffset; got firstLine=%d, YOffset=%d", firstLine, yoff)
	}
}
