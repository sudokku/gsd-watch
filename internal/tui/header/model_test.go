package header_test

import (
	"strings"
	"testing"

	"github.com/radu/gsd-watch/internal/parser"
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
	// Build data with 5 of 10 plans complete for precisely 50%.
	data := parser.ProjectData{
		Name:         "test-project",
		ModelProfile: "fast",
		Mode:         "auto",
		Phases: []parser.Phase{
			{
				Plans: []parser.Plan{
					{Status: "complete"},
					{Status: "complete"},
					{Status: "complete"},
					{Status: "complete"},
					{Status: "complete"},
					{Status: "pending"},
					{Status: "pending"},
					{Status: "pending"},
					{Status: "pending"},
					{Status: "pending"},
				},
			},
		},
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
	data := mock.MockProject()
	for i := range data.Phases {
		for j := range data.Phases[i].Plans {
			data.Phases[i].Plans[j].Status = "pending"
		}
	}
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
	data := mock.MockProject()
	for i := range data.Phases {
		for j := range data.Phases[i].Plans {
			data.Phases[i].Plans[j].Status = "complete"
		}
	}
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
	if h.Height() != 3 {
		t.Errorf("expected Height() to return 3, got %d", h.Height())
	}
}
