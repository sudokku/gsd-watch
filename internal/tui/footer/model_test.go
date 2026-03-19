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
	if !strings.Contains(out, "↑/k") {
		t.Errorf("expected View(80) to contain key hint '↑/k', got:\n%s", out)
	}
	if !strings.Contains(out, "q") {
		t.Errorf("expected View(80) to contain key hint 'q', got:\n%s", out)
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
	if f.Height() != 2 {
		t.Errorf("expected Height() to return 2, got %d", f.Height())
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
