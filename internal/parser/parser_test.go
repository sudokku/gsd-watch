package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// testdataProjectDir is a fully-structured .planning/ fixture directory.
const testdataProjectDir = "testdata/project"

func TestParseProject_FullFixture(t *testing.T) {
	data := ParseProject(testdataProjectDir)

	// From testdata/PROJECT.md: # Test Project
	if data.Name != "Test Project" {
		t.Errorf("expected Name=%q from PROJECT.md H1, got %q", "Test Project", data.Name)
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
	// Should not panic; name falls back to project directory basename.
	data := ParseProject("/nonexistent/path/to/.planning")

	if data.Name != "to" {
		t.Errorf("expected Name=%q (dir basename fallback) for missing root, got %q", "to", data.Name)
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

func TestParseProject_ProjectMDH1(t *testing.T) {
	// Fixture layout mirrors real usage: PROJECT.md at project root, STATE.md inside .planning/
	data := ParseProject("testdata/project-fallback/.planning")
	if data.Name != "My Test Project" {
		t.Errorf("expected Name=%q from PROJECT.md H1, got %q", "My Test Project", data.Name)
	}
}

func TestParseProject_NoProjectMDUsesDir(t *testing.T) {
	dir := t.TempDir()
	// .planning/ subdir with STATE.md, no PROJECT.md at project root
	planningDir := filepath.Join(dir, ".planning")
	os.MkdirAll(filepath.Join(planningDir, "phases"), 0755)
	stateContent := "---\nmilestone_name: some-milestone\n---\n\n# Project State\n\nPhase: 0\nPlan: 0\n"
	os.WriteFile(filepath.Join(planningDir, "STATE.md"), []byte(stateContent), 0644)
	data := ParseProject(planningDir)
	// No PROJECT.md → name is the project directory basename (temp dir name)
	if data.Name != filepath.Base(dir) {
		t.Errorf("expected Name=%q (dir basename), got %q", filepath.Base(dir), data.Name)
	}
}

func TestParseProject_EmptyRoot(t *testing.T) {
	dir := t.TempDir()
	data := ParseProject(dir)

	// No PROJECT.md → name is the parent directory basename; just check it's non-empty.
	if data.Name == "" {
		t.Errorf("expected non-empty Name for empty root, got %q", data.Name)
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
