package footer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/radu/gsd-watch/internal/parser"
	"github.com/radu/gsd-watch/internal/tui/footer"
	"github.com/radu/gsd-watch/internal/tui/mock"
	tui "github.com/radu/gsd-watch/internal/tui"
)

func TestFooterView_ContainsWatchingLabel(t *testing.T) {
	// No file set yet — footer should show the "watching…" fallback label.
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetWidth(80)
	out := f.View(80)
	if !strings.Contains(out, "watching") {
		t.Errorf("expected View(80) to contain 'watching…' label before any file change, got:\n%s", out)
	}
}

func TestFooterView_ContainsLastChangeLabel(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetWidth(80)
	f = f.SetLastFile("STATE.md")
	out := f.View(80)
	if !strings.Contains(out, "Last change: STATE.md") {
		t.Errorf("expected View(80) to contain 'Last change: STATE.md', got:\n%s", out)
	}
}

func TestFooterView_ContainsTimeSince_JustNow(t *testing.T) {
	// MockProject sets LastUpdated to time.Now() — should show "just now".
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetWidth(80)
	out := f.View(80)
	if !strings.Contains(out, "just now") {
		t.Errorf("expected View(80) to contain 'just now' for a fresh timestamp, got:\n%s", out)
	}
}

func TestFooterView_ContainsTimeSince_SecondsAgo(t *testing.T) {
	// Set LastUpdated to 10 seconds ago — should show "Ns ago".
	data := mock.MockProject()
	data.LastUpdated = time.Now().Add(-10 * time.Second)
	f := footer.New(data, tui.DefaultKeyMap())
	f = f.SetWidth(80)
	out := f.View(80)
	if !strings.Contains(out, "ago") {
		t.Errorf("expected View(80) to contain 'ago' for a 10s-old timestamp, got:\n%s", out)
	}
}

func TestFooterView_ContainsKeyHints(t *testing.T) {
	data := mock.MockProject()
	f := footer.New(data, tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "←h") {
		t.Errorf("expected View(80) to contain nav hint '←h', got:\n%s", out)
	}
	if !strings.Contains(out, "qq esc quit") {
		t.Errorf("expected View(80) to contain 'qq esc quit', got:\n%s", out)
	}
}

func TestFooterView_TooNarrow(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	out := f.View(20)
	if !strings.Contains(out, "too narrow") {
		t.Errorf("expected View(20) to contain 'too narrow', got:\n%s", out)
	}
}

func TestFooterHeight(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetWidth(80)
	if f.Height() != 5 {
		t.Errorf("expected Height() to return 5, got %d", f.Height())
	}
}

func TestFooterSetData_UpdatesLastUpdated(t *testing.T) {
	data := mock.MockProject()
	f := footer.New(data, tui.DefaultKeyMap())
	f = f.SetWidth(80)

	// Set lastUpdated to something old so timeSince shows "ago".
	oldData := parser.ProjectData{
		LastUpdated: time.Now().Add(-30 * time.Second),
	}
	f = f.SetData(oldData)
	out := f.View(80)
	if !strings.Contains(out, "ago") {
		t.Errorf("expected View(80) after SetData to show stale timestamp ('ago'), got:\n%s", out)
	}
}

func TestFooter_IdleCheckmark(t *testing.T) {
	// Default state (no active changes) should show the ✓ checkmark.
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "✓") {
		t.Errorf("expected idle checkmark ✓ in footer, got:\n%s", out)
	}
}

func TestFooter_ActiveChanges_ShowsSpinner(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetActiveChanges(true)
	out := f.View(80)
	// The idle ✓ should not appear when active.
	if strings.Contains(out, "✓") {
		t.Errorf("expected no ✓ while active, got:\n%s", out)
	}
	// A braille spinner frame should be present.
	spinFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	found := false
	for _, frame := range spinFrames {
		if strings.Contains(out, frame) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a braille spinner frame while active, got:\n%s", out)
	}
}

func TestFooter_AdvanceSpinFrame(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetActiveChanges(true)
	// Advance through all 10 frames and verify each appears in the view.
	spinFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for i, want := range spinFrames {
		out := f.View(80)
		if !strings.Contains(out, want) {
			t.Errorf("frame %d: expected %q in view, got:\n%s", i, want, out)
		}
		f = f.AdvanceSpinFrame()
	}
}

func TestFooter_SetActiveChanges_False_ResetsFrame(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetActiveChanges(true)
	f = f.AdvanceSpinFrame()
	f = f.AdvanceSpinFrame()
	f = f.SetActiveChanges(false)
	// After clearing active state, view should show ✓ and first braille frame should
	// appear again if re-activated (spinFrame reset to 0).
	out := f.View(80)
	if !strings.Contains(out, "✓") {
		t.Errorf("expected ✓ after SetActiveChanges(false), got:\n%s", out)
	}
	f = f.SetActiveChanges(true)
	out2 := f.View(80)
	if !strings.Contains(out2, "⠋") {
		t.Errorf("expected spinner to restart at frame 0 (⠋) after re-activation, got:\n%s", out2)
	}
}

func TestFooterHeight_FiveLines(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetWidth(80)
	got := f.Height()
	if got != 5 {
		t.Errorf("expected Height() == 5 after layout changes, got %d", got)
	}
}

func TestFooter_QuitPending(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetQuitPending(true)
	out := f.View(80)
	if !strings.Contains(out, "press q or esc again to exit") {
		t.Errorf("expected quit-pending message in footer, got:\n%s", out)
	}
	// Normal hints should not appear while pending.
	if strings.Contains(out, "←h") {
		t.Errorf("expected nav hints hidden while quit-pending, got:\n%s", out)
	}
}

func TestFooterView_ContainsHelpHint(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "? help") {
		t.Errorf("expected footer to contain '? help' hint, got:\n%s", out)
	}
}

// TestFooterSetTheme_AllPresets verifies that SetTheme() can be called on all three
// presets and that View() still renders the ─ separator line.
func TestFooterSetTheme_AllPresets(t *testing.T) {
	presets := []struct {
		name string
		th   tui.Theme
	}{
		{"default", tui.ThemeDefault()},
		{"minimal", tui.ThemeMinimal()},
		{"high-contrast", tui.ThemeHighContrast()},
	}
	for _, tt := range presets {
		f := footer.New(mock.MockProject(), tui.DefaultKeyMap()).SetTheme(tt.th)
		f = f.SetWidth(80)
		out := f.View(80)
		// Separator line should always be present regardless of theme.
		if !strings.Contains(out, "─") {
			t.Errorf("preset %q: expected ─ separator in footer output, got:\n%s", tt.name, out)
		}
	}
}
