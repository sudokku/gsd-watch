package parser

import (
	"os"
	"testing"
)

// testdataProjectDir is a fully-structured .planning/ fixture directory.
const testdataProjectDir = "testdata/project"

func TestParseProject_FullFixture(t *testing.T) {
	data := ParseProject(testdataProjectDir)

	// From STATE.md: milestone_name: milestone
	if data.Name != "milestone" {
		t.Errorf("expected Name=%q, got %q", "milestone", data.Name)
	}

	// From config.json: model_profile: balanced
	if data.ModelProfile != "balanced" {
		t.Errorf("expected ModelProfile=%q, got %q", "balanced", data.ModelProfile)
	}

	// From config.json: mode: yolo
	if data.Mode != "yolo" {
		t.Errorf("expected Mode=%q, got %q", "yolo", data.Mode)
	}

	// From STATE.md: stopped_at: "Phase 2 context gathered"
	if data.CurrentAction != "Phase 2 context gathered" {
		t.Errorf("expected CurrentAction=%q, got %q", "Phase 2 context gathered", data.CurrentAction)
	}

	// ProgressPercent is computed from actual phase completion, not STATE.md percent.
	// Fixture phases: phase 01 is in_progress (plan 01-02 not complete), others pending.
	// Expected: 0/4 phases complete → 0.0.
	if data.ProgressPercent != 0.0 {
		t.Errorf("expected ProgressPercent=0.0 (computed from phases), got %f", data.ProgressPercent)
	}

	// Four phases total: 2 from directories + 2 stubs from ROADMAP.md
	if len(data.Phases) != 4 {
		t.Fatalf("expected 4 phases, got %d", len(data.Phases))
	}

	// Phase 1 should have badges (CONTEXT.md + RESEARCH.md)
	phase1 := data.Phases[0]
	hasBadge := func(badges []string, badge string) bool {
		for _, b := range badges {
			if b == badge {
				return true
			}
		}
		return false
	}
	if !hasBadge(phase1.Badges, BadgeDiscussed) {
		t.Errorf("expected phase 1 to have badge %q, got %v", BadgeDiscussed, phase1.Badges)
	}
	if !hasBadge(phase1.Badges, BadgeResearched) {
		t.Errorf("expected phase 1 to have badge %q, got %v", BadgeResearched, phase1.Badges)
	}

	// Phase names should come from ROADMAP.md
	if phase1.Name != "Phase 1: Core TUI Scaffold" {
		t.Errorf("expected phase1.Name=%q, got %q", "Phase 1: Core TUI Scaffold", phase1.Name)
	}
}

func TestParseProject_MissingRoot(t *testing.T) {
	// Should not panic; should return "unknown" defaults.
	data := ParseProject("/nonexistent/path/to/.planning")

	if data.Name != "unknown" {
		t.Errorf("expected Name=%q for missing root, got %q", "unknown", data.Name)
	}
	if data.ModelProfile != "unknown" {
		t.Errorf("expected ModelProfile=%q for missing root, got %q", "unknown", data.ModelProfile)
	}
	if data.Mode != "unknown" {
		t.Errorf("expected Mode=%q for missing root, got %q", "unknown", data.Mode)
	}
	if data.CurrentAction != "unknown" {
		t.Errorf("expected CurrentAction=%q for missing root, got %q", "unknown", data.CurrentAction)
	}
	if len(data.Phases) != 0 {
		t.Errorf("expected empty Phases for missing root, got %d phases", len(data.Phases))
	}
}

func TestParseProject_EmptyRoot(t *testing.T) {
	dir := t.TempDir()
	data := ParseProject(dir)

	if data.Name != "unknown" {
		t.Errorf("expected Name=%q for empty root, got %q", "unknown", data.Name)
	}
	if data.ModelProfile != "unknown" {
		t.Errorf("expected ModelProfile=%q for empty root, got %q", "unknown", data.ModelProfile)
	}
	if len(data.Phases) != 0 {
		t.Errorf("expected empty Phases for empty root, got %d phases", len(data.Phases))
	}

	// Empty root: create phases dir but no phase subdirs
	if err := os.Mkdir(dir+"/phases", 0755); err != nil {
		t.Fatal(err)
	}
	data2 := ParseProject(dir)
	if len(data2.Phases) != 0 {
		t.Errorf("expected empty Phases for root with empty phases dir, got %d phases", len(data2.Phases))
	}
}
