package parser

import (
	"path/filepath"
	"testing"
)

func TestParseState_Full(t *testing.T) {
	sd, err := parseState(filepath.Join("testdata", "state.md"))
	if err != nil {
		t.Fatalf("parseState returned unexpected error: %v", err)
	}

	if sd.MilestoneName != "milestone" {
		t.Errorf("MilestoneName: got %q, want %q", sd.MilestoneName, "milestone")
	}
	if sd.StoppedAt != "Phase 2 context gathered" {
		t.Errorf("StoppedAt: got %q, want %q", sd.StoppedAt, "Phase 2 context gathered")
	}
	if sd.ProgressPercent != 25 {
		t.Errorf("ProgressPercent: got %d, want %d", sd.ProgressPercent, 25)
	}
	if sd.ActivePhase != 1 {
		t.Errorf("ActivePhase: got %d, want %d", sd.ActivePhase, 1)
	}
	if sd.ActivePlan != 4 {
		t.Errorf("ActivePlan: got %d, want %d", sd.ActivePlan, 4)
	}
}

func TestParseState_NoFrontmatter(t *testing.T) {
	sd, err := parseState(filepath.Join("testdata", "state-no-frontmatter.md"))
	if err != nil {
		t.Fatalf("parseState returned unexpected error: %v", err)
	}

	if sd.MilestoneName != "" {
		t.Errorf("MilestoneName: got %q, want empty string", sd.MilestoneName)
	}
	if sd.StoppedAt != "" {
		t.Errorf("StoppedAt: got %q, want empty string", sd.StoppedAt)
	}
	if sd.ProgressPercent != 0 {
		t.Errorf("ProgressPercent: got %d, want 0", sd.ProgressPercent)
	}
	if sd.ActivePhase != 0 {
		t.Errorf("ActivePhase: got %d, want 0", sd.ActivePhase)
	}
	if sd.ActivePlan != 0 {
		t.Errorf("ActivePlan: got %d, want 0", sd.ActivePlan)
	}
}

func TestParseState_MinimalFrontmatter(t *testing.T) {
	sd, err := parseState(filepath.Join("testdata", "state-minimal.md"))
	if err != nil {
		t.Fatalf("parseState returned unexpected error: %v", err)
	}

	if sd.MilestoneName != "test-project" {
		t.Errorf("MilestoneName: got %q, want %q", sd.MilestoneName, "test-project")
	}
	if sd.StoppedAt != "" {
		t.Errorf("StoppedAt: got %q, want empty string", sd.StoppedAt)
	}
	if sd.ProgressPercent != 0 {
		t.Errorf("ProgressPercent: got %d, want 0", sd.ProgressPercent)
	}
	if sd.ActivePhase != 0 {
		t.Errorf("ActivePhase: got %d, want 0", sd.ActivePhase)
	}
	if sd.ActivePlan != 0 {
		t.Errorf("ActivePlan: got %d, want 0", sd.ActivePlan)
	}
}

func TestParseState_MissingFile(t *testing.T) {
	_, err := parseState(filepath.Join("testdata", "nonexistent-state.md"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseState_ActivePlanRegex(t *testing.T) {
	sd, err := parseState(filepath.Join("testdata", "state-active-plan.md"))
	if err != nil {
		t.Fatalf("parseState returned unexpected error: %v", err)
	}

	if sd.ActivePhase != 2 {
		t.Errorf("ActivePhase: got %d, want 2", sd.ActivePhase)
	}
	if sd.ActivePlan != 1 {
		t.Errorf("ActivePlan: got %d, want 1", sd.ActivePlan)
	}
}
