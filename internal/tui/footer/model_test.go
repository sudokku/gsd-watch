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

func TestFooterView_ContainsCurrentAction(t *testing.T) {
	data := mock.MockProject()
	f := footer.New(data, tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "Phase 1 — building TUI scaffold") {
		t.Errorf("expected View(80) to contain current action, got:\n%s", out)
	}
}

func TestFooterView_ContainsTimeSince(t *testing.T) {
	data := mock.MockProject()
	// MockProject sets LastUpdated to time.Now(), so we expect "0s ago" or "1s ago".
	f := footer.New(data, tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "ago") {
		t.Errorf("expected View(80) to contain time-since string like '0s ago', got:\n%s", out)
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

func TestFooterSetData_UpdatesAction(t *testing.T) {
	data := mock.MockProject()
	f := footer.New(data, tui.DefaultKeyMap())

	newData := parser.ProjectData{
		Name:          data.Name,
		ModelProfile:  data.ModelProfile,
		Mode:          data.Mode,
		CurrentAction: "Phase 2 — parsing files",
		LastUpdated:   time.Now(),
		Phases:        data.Phases,
	}
	f = f.SetData(newData)
	out := f.View(80)
	if !strings.Contains(out, "Phase 2 — parsing files") {
		t.Errorf("expected View(80) after SetData to contain new action, got:\n%s", out)
	}
}

func TestFooter_RefreshIdle(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "↺") {
		t.Errorf("expected idle refresh icon ↺, got:\n%s", out)
	}
}

func TestFooter_RefreshFlash(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetRefreshFlash(true)
	out := f.View(80)
	if !strings.Contains(out, "⟳") {
		t.Errorf("expected flash refresh icon ⟳, got:\n%s", out)
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

func TestFooterSetRefreshFlash(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	f = f.SetRefreshFlash(true)
	out1 := f.View(80)
	if !strings.Contains(out1, "⟳") {
		t.Errorf("expected flash icon after SetRefreshFlash(true)")
	}
	f = f.SetRefreshFlash(false)
	out2 := f.View(80)
	if !strings.Contains(out2, "↺") {
		t.Errorf("expected idle icon after SetRefreshFlash(false)")
	}
}

func TestFooterView_ContainsHelpHint(t *testing.T) {
	f := footer.New(mock.MockProject(), tui.DefaultKeyMap())
	out := f.View(80)
	if !strings.Contains(out, "? help") {
		t.Errorf("expected footer to contain '? help' hint, got:\n%s", out)
	}
}
