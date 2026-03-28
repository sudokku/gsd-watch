package header_test

import (
	"strings"
	"testing"

	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
	"github.com/radu/gsd-watch/internal/tui/header"
	"github.com/radu/gsd-watch/internal/tui/mock"
)

func TestHeaderView_ContainsProjectName(t *testing.T) {
	h := header.New(mock.MockProject())
	out := h.View(80)
	if !strings.Contains(out, "gsd-watch") {
		t.Errorf("expected View(80) to contain project name %q, got:\n%s", "gsd-watch", out)
	}
}

func TestHeaderView_ContainsModelProfile(t *testing.T) {
	h := header.New(mock.MockProject())
	out := h.View(80)
	if !strings.Contains(out, "balanced") {
		t.Errorf("expected View(80) to contain model profile %q, got:\n%s", "balanced", out)
	}
}

func TestHeaderView_ContainsMode(t *testing.T) {
	h := header.New(mock.MockProject())
	out := h.View(80)
	if !strings.Contains(out, "yolo") {
		t.Errorf("expected View(80) to contain mode %q, got:\n%s", "yolo", out)
	}
}

func TestHeaderView_ProgressBar50Percent(t *testing.T) {
	// Build data with ProgressPercent=0.5 for precisely 50%.
	data := parser.ProjectData{
		Name:            "test-project",
		ModelProfile:    "fast",
		Mode:            "auto",
		ProgressPercent: 0.5,
	}
	h := header.New(data)
	out := h.View(80)
	if !strings.Contains(out, "▓") {
		t.Errorf("expected View(80) at 50%% to contain filled bar chars '▓', got:\n%s", out)
	}
	if !strings.Contains(out, "░") {
		t.Errorf("expected View(80) at 50%% to contain empty bar chars '░', got:\n%s", out)
	}
}

func TestHeaderView_ZeroPercent(t *testing.T) {
	// ProgressPercent=0 (default zero value) → all empty bar.
	data := mock.MockProject()
	data.ProgressPercent = 0.0
	h := header.New(data)
	out := h.View(80)
	if strings.Contains(out, "▓") {
		t.Errorf("expected View(80) at 0%% to have no filled bar chars '▓', got:\n%s", out)
	}
	if !strings.Contains(out, "░") {
		t.Errorf("expected View(80) at 0%% to have empty bar chars '░', got:\n%s", out)
	}
}

func TestHeaderView_HundredPercent(t *testing.T) {
	// ProgressPercent=1.0 → all filled bar.
	data := mock.MockProject()
	data.ProgressPercent = 1.0
	h := header.New(data)
	out := h.View(80)
	if strings.Contains(out, "░") {
		t.Errorf("expected View(80) at 100%% to have no empty bar chars '░', got:\n%s", out)
	}
	if !strings.Contains(out, "▓") {
		t.Errorf("expected View(80) at 100%% to have filled bar chars '▓', got:\n%s", out)
	}
}

func TestHeaderView_TooNarrow(t *testing.T) {
	h := header.New(mock.MockProject())
	out := h.View(20)
	if !strings.Contains(out, "too narrow") {
		t.Errorf("expected View(20) to contain 'too narrow', got:\n%s", out)
	}
}

func TestHeaderHeight(t *testing.T) {
	h := header.New(mock.MockProject())
	if h.Height() != 4 {
		t.Errorf("expected Height() to return 4, got %d", h.Height())
	}
}

// TestHeaderSetTheme_ContainsSeparator verifies that View() renders the ═ separator line
// and that SetTheme() can be called without panic on all three presets.
func TestHeaderSetTheme_AllPresets(t *testing.T) {
	presets := []struct {
		name string
		th   tui.Theme
	}{
		{"default", tui.ThemeDefault()},
		{"minimal", tui.ThemeMinimal()},
		{"high-contrast", tui.ThemeHighContrast()},
	}
	for _, tt := range presets {
		h := header.New(mock.MockProject()).SetTheme(tt.th)
		out := h.View(80)
		// Separator line should always be present regardless of theme.
		if !strings.Contains(out, "═") {
			t.Errorf("preset %q: expected ═ separator in header output, got:\n%s", tt.name, out)
		}
		// Project name should always be present.
		if !strings.Contains(out, "gsd-watch") {
			t.Errorf("preset %q: expected project name in header output, got:\n%s", tt.name, out)
		}
	}
}

// TestHeader_ProgressBar_ThemeColors verifies that at 50% completion the progress bar
// contains both filled (▓) and empty (░) blocks with each theme applied.
func TestHeader_ProgressBar_ThemeColors(t *testing.T) {
	data := parser.ProjectData{
		Name:            "test-project",
		ModelProfile:    "fast",
		Mode:            "auto",
		ProgressPercent: 0.5,
	}
	presets := []tui.Theme{tui.ThemeDefault(), tui.ThemeMinimal(), tui.ThemeHighContrast()}
	for _, th := range presets {
		h := header.New(data).SetTheme(th)
		out := h.View(80)
		if !strings.Contains(out, "▓") {
			t.Errorf("theme progress bar at 50%%: expected '▓' in output, got:\n%s", out)
		}
		if !strings.Contains(out, "░") {
			t.Errorf("theme progress bar at 50%%: expected '░' in output, got:\n%s", out)
		}
	}
}
