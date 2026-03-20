package parser

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestPlanning creates a minimal .planning/ fixture in a temp dir and returns root path.
func setupTestPlanning(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// config.json
	writeFile(t, filepath.Join(root, "config.json"), `{"model_profile":"balanced","mode":"yolo"}`)

	// STATE.md with frontmatter + prose
	writeFile(t, filepath.Join(root, "STATE.md"), `---
milestone_name: v1.0
stopped_at: Phase 1 plan 1 complete
progress:
  percent: 25
---

# Project State

## Current Position

Phase: 1 of 2
Plan: 1 of 2
`)

	// ROADMAP.md (minimal)
	writeFile(t, filepath.Join(root, "ROADMAP.md"), `---
---

# Roadmap

| Phase | Name |
|-------|------|
| 1 | Test Phase |
`)

	// phases/01-test/01-01-PLAN.md
	phaseDir := filepath.Join(root, "phases", "01-test")
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		t.Fatalf("mkdir phases: %v", err)
	}
	writeFile(t, filepath.Join(phaseDir, "01-01-PLAN.md"), `---
status: in_progress
wave: 1
---
<objective>
Test plan one
</objective>
`)

	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// TestNewCache verifies NewCache returns a non-nil cache with empty mtimes.
func TestNewCache(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	if c == nil {
		t.Fatal("NewCache returned nil")
	}
	if c.root != root {
		t.Errorf("root: got %q, want %q", c.root, root)
	}
	if c.mtimes == nil {
		t.Error("mtimes map is nil, want empty map")
	}
	if len(c.mtimes) != 0 {
		t.Errorf("expected empty mtimes, got %d entries", len(c.mtimes))
	}
}

// TestParseFull verifies ParseFull() returns same key fields as ParseProject().
func TestParseFull(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)

	got := c.ParseFull()
	want := ParseProject(root)

	// Name, CurrentAction, ModelProfile, Mode, ProgressPercent must match.
	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}
	if got.CurrentAction != want.CurrentAction {
		t.Errorf("CurrentAction: got %q, want %q", got.CurrentAction, want.CurrentAction)
	}
	if got.ModelProfile != want.ModelProfile {
		t.Errorf("ModelProfile: got %q, want %q", got.ModelProfile, want.ModelProfile)
	}
	if got.Mode != want.Mode {
		t.Errorf("Mode: got %q, want %q", got.Mode, want.Mode)
	}
	if got.ProgressPercent != want.ProgressPercent {
		t.Errorf("ProgressPercent: got %v, want %v", got.ProgressPercent, want.ProgressPercent)
	}
	if len(got.Phases) != len(want.Phases) {
		t.Errorf("Phases count: got %d, want %d", len(got.Phases), len(want.Phases))
	}
	// mtimes map should be populated after ParseFull.
	if len(c.mtimes) == 0 {
		t.Error("mtimes not populated after ParseFull()")
	}
}

// TestCacheUpdateStateMd verifies Update(statePath) re-parses only STATE.md fields.
func TestCacheUpdateStateMd(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	before := c.ParseFull()

	statePath := filepath.Join(root, "STATE.md")

	// Ensure mtime will change by sleeping briefly.
	time.Sleep(5 * time.Millisecond)

	// Modify STATE.md stopped_at.
	writeFile(t, statePath, `---
milestone_name: v1.0
stopped_at: Phase 2 plan 2 complete
progress:
  percent: 50
---

# Project State

## Current Position

Phase: 1 of 2
Plan: 1 of 2
`)

	after := c.Update(statePath)

	if after.CurrentAction == before.CurrentAction {
		t.Errorf("CurrentAction unchanged after STATE.md update: %q", after.CurrentAction)
	}
	if after.CurrentAction != "Phase 2 plan 2 complete" {
		t.Errorf("CurrentAction: got %q, want %q", after.CurrentAction, "Phase 2 plan 2 complete")
	}
	// ProgressPercent is now computed from phase completion, not STATE.md percent.
	// Fixture has 1 phase with 1 in_progress plan → 0/1 complete → 0.0.
	if after.ProgressPercent != 0.0 {
		t.Errorf("ProgressPercent: got %v, want 0.0 (computed from phases, not STATE.md percent)", after.ProgressPercent)
	}
	// Phases should still be present (not wiped).
	if len(after.Phases) == 0 {
		t.Error("Phases empty after STATE.md update")
	}
}

// TestCacheUpdateConfigJson verifies Update(configPath) re-parses model/mode.
func TestCacheUpdateConfigJson(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	before := c.ParseFull()

	configPath := filepath.Join(root, "config.json")
	time.Sleep(5 * time.Millisecond)

	writeFile(t, configPath, `{"model_profile":"fast","mode":"strict"}`)

	after := c.Update(configPath)

	if after.ModelProfile == before.ModelProfile {
		t.Errorf("ModelProfile unchanged after config.json update")
	}
	if after.ModelProfile != "fast" {
		t.Errorf("ModelProfile: got %q, want %q", after.ModelProfile, "fast")
	}
	if after.Mode != "strict" {
		t.Errorf("Mode: got %q, want %q", after.Mode, "strict")
	}
	// Other fields untouched — CurrentAction should remain.
	if after.CurrentAction != before.CurrentAction {
		t.Errorf("CurrentAction changed unexpectedly: got %q, want %q", after.CurrentAction, before.CurrentAction)
	}
}

// TestCacheUpdatePlanMd verifies Update(planPath) only re-parses the affected plan.
func TestCacheUpdatePlanMd(t *testing.T) {
	root := setupTestPlanning(t)

	// Add a second plan so we can verify it's untouched.
	phaseDir := filepath.Join(root, "phases", "01-test")
	writeFile(t, filepath.Join(phaseDir, "01-02-PLAN.md"), `---
status: pending
wave: 1
---
<objective>
Test plan two
</objective>
`)

	c := NewCache(root)
	before := c.ParseFull()

	planPath := filepath.Join(phaseDir, "01-01-PLAN.md")
	time.Sleep(5 * time.Millisecond)

	// Change plan 01-01 status to complete.
	writeFile(t, planPath, `---
status: complete
wave: 1
---
<objective>
Test plan one
</objective>
`)

	after := c.Update(planPath)

	// Find plan 01-01 in the result.
	var found bool
	for _, ph := range after.Phases {
		for _, pl := range ph.Plans {
			if pl.Filename == "01-01-PLAN.md" {
				found = true
				if pl.Status != StatusComplete {
					t.Errorf("01-01-PLAN.md status: got %q, want %q", pl.Status, StatusComplete)
				}
			}
		}
	}
	if !found {
		t.Error("01-01-PLAN.md not found in updated phases")
	}

	// Other fields untouched.
	if after.ModelProfile != before.ModelProfile {
		t.Errorf("ModelProfile changed unexpectedly after PLAN.md update")
	}
}

// TestCacheUpdateBadgeFile verifies Update(badgePath) re-detects badges for that phase.
func TestCacheUpdateBadgeFile(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	before := c.ParseFull()

	// Verify no "discussed" badge initially.
	for _, ph := range before.Phases {
		for _, b := range ph.Badges {
			if b == BadgeDiscussed {
				t.Fatalf("unexpected discussed badge before CONTEXT.md created")
			}
		}
	}

	// Create badge file with new mtime.
	time.Sleep(5 * time.Millisecond)
	phaseDir := filepath.Join(root, "phases", "01-test")
	badgePath := filepath.Join(phaseDir, "01-CONTEXT.md")
	writeFile(t, badgePath, "# Context")

	after := c.Update(badgePath)

	var found bool
	for _, ph := range after.Phases {
		if ph.DirName == "01-test" {
			for _, b := range ph.Badges {
				if b == BadgeDiscussed {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("discussed badge not found after CONTEXT.md created and Update() called")
	}
}

// TestCacheUpdateUnknownFile verifies Update() with unrecognized path falls back to full re-parse.
func TestCacheUpdateUnknownFile(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	c.ParseFull()

	// Write an unrecognized file.
	unknownPath := filepath.Join(root, "some-random-file.txt")
	time.Sleep(5 * time.Millisecond)
	writeFile(t, unknownPath, "hello")

	after := c.Update(unknownPath)

	// Should still return valid data (full re-parse fallback).
	if after.Name == "" {
		t.Error("Name empty after unknown file Update()")
	}
	if len(after.Phases) == 0 {
		t.Error("Phases empty after unknown file Update()")
	}
}

// TestCacheUpdateMtimeSkip verifies that Update() skips re-parse when mtime unchanged.
func TestCacheUpdateMtimeSkip(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	c.ParseFull()

	configPath := filepath.Join(root, "config.json")

	// First Update() — mtime hasn't changed since ParseFull walked the file.
	first := c.Update(configPath)

	// Overwrite file with new content but DON'T sleep — mtime may or may not change.
	// For a reliable mtime-skip test, we call Update() a second time on a file
	// that we already updated in the first call (mtime is now recorded).
	// Write new content and call Update once to record the new mtime.
	time.Sleep(5 * time.Millisecond)
	writeFile(t, configPath, `{"model_profile":"updated","mode":"strict"}`)
	second := c.Update(configPath) // should re-parse (new mtime)

	if second.ModelProfile != "updated" {
		t.Errorf("ModelProfile: got %q, want %q (second update didn't re-parse)", second.ModelProfile, "updated")
	}

	// Third call without any file modification — mtime guard should skip re-parse.
	// Stale in memory: we'll overwrite the field to simulate stale cache.
	c.data.ModelProfile = "stale-sentinel"
	third := c.Update(configPath) // same mtime => should skip re-parse

	if third.ModelProfile != "stale-sentinel" {
		t.Errorf("mtime guard failed: expected stale-sentinel (skipped re-parse), got %q", third.ModelProfile)
	}

	_ = first
}
