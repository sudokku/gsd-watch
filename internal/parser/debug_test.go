package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestDebugSilentByDefault verifies DebugOut is nil at package init and
// calling debugf with nil DebugOut produces no panic and no output.
func TestDebugSilentByDefault(t *testing.T) {
	if DebugOut != nil {
		t.Fatal("DebugOut must be nil by default (package zero value)")
	}
	// Should not panic when DebugOut is nil.
	debugf("test", "hello %s", "world")
}

// TestDebugPhaseDir verifies parsePhases emits phase_dir events with num and name.
func TestDebugPhaseDir(t *testing.T) {
	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	dir := t.TempDir()
	phaseDir := filepath.Join(dir, "01-test-phase")
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(phaseDir, "01-01-PLAN.md"), "---\nstatus: pending\n---\n")

	parsePhases(dir, map[int]string{1: "test phase"}, 0, 0)

	out := buf.String()
	if !strings.Contains(out, "phase_dir:") {
		t.Errorf("expected 'phase_dir:' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "num=1") {
		t.Errorf("expected 'num=1' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, `name="Phase 1: test phase"`) {
		t.Errorf(`expected 'name="Phase 1: test phase"' in debug output, got:\n%s`, out)
	}
}

// TestDebugPlan verifies parsePhases emits plan events with status, title, wave.
func TestDebugPlan(t *testing.T) {
	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	dir := t.TempDir()
	phaseDir := filepath.Join(dir, "01-test-phase")
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(phaseDir, "01-01-PLAN.md"), "---\nstatus: pending\nwave: 2\n---\n<objective>\nTest objective\n</objective>\n")

	parsePhases(dir, map[int]string{1: "test phase"}, 0, 0)

	out := buf.String()
	if !strings.Contains(out, "plan:") {
		t.Errorf("expected 'plan:' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "status=") {
		t.Errorf("expected 'status=' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "title=") {
		t.Errorf("expected 'title=' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "wave=") {
		t.Errorf("expected 'wave=' in debug output, got:\n%s", out)
	}
}

// TestDebugPlanError verifies parsePhases emits plan_error events for malformed PLAN.md.
func TestDebugPlanError(t *testing.T) {
	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	dir := t.TempDir()
	phaseDir := filepath.Join(dir, "01-test-phase")
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Invalid YAML: unclosed bracket causes yaml.Unmarshal to fail.
	writeFile(t, filepath.Join(phaseDir, "01-01-PLAN.md"), "---\nstatus: [invalid yaml\n---\n")

	parsePhases(dir, map[int]string{1: "test phase"}, 0, 0)

	out := buf.String()
	if !strings.Contains(out, "plan_error:") {
		t.Errorf("expected 'plan_error:' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "err=") {
		t.Errorf("expected 'err=' in debug output, got:\n%s", out)
	}
}

// TestDebugBadge verifies parsePhases emits badge events for badge files.
func TestDebugBadge(t *testing.T) {
	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	dir := t.TempDir()
	phaseDir := filepath.Join(dir, "01-test")
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a badge file: 01-CONTEXT.md maps to BadgeDiscussed ("discussed").
	writeFile(t, filepath.Join(phaseDir, "01-CONTEXT.md"), "")

	parsePhases(dir, map[int]string{1: "test"}, 0, 0)

	out := buf.String()
	if !strings.Contains(out, "badge:") {
		t.Errorf("expected 'badge:' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "discussed") {
		t.Errorf("expected 'discussed' in debug output, got:\n%s", out)
	}
}

// TestDebugCacheHIT verifies Update() emits cache HIT when mtime unchanged.
func TestDebugCacheHIT(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	c.ParseFull()

	statePath := filepath.Join(root, "STATE.md")

	// First Update() call — mtime recorded by ParseFull, should be a HIT.
	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	c.Update(statePath)

	out := buf.String()
	if !strings.Contains(out, "cache:") {
		t.Errorf("expected 'cache:' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "HIT") {
		t.Errorf("expected 'HIT' in debug output, got:\n%s", out)
	}
}

// TestDebugCacheMISS verifies Update() emits cache MISS when file mtime changed.
func TestDebugCacheMISS(t *testing.T) {
	root := setupTestPlanning(t)
	c := NewCache(root)
	c.ParseFull()

	statePath := filepath.Join(root, "STATE.md")

	// Modify the file so mtime changes.
	time.Sleep(5 * time.Millisecond)
	writeFile(t, statePath, `---
milestone_name: v1.1
stopped_at: Phase 2 plan 1 complete
progress:
  percent: 50
---

# Project State
`)

	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	c.Update(statePath)

	out := buf.String()
	if !strings.Contains(out, "cache:") {
		t.Errorf("expected 'cache:' in debug output, got:\n%s", out)
	}
	if !strings.Contains(out, "MISS") {
		t.Errorf("expected 'MISS' in debug output, got:\n%s", out)
	}
}

// TestDebugFormat verifies the debug output format matches [debug HH:MM:SS] event: details.
func TestDebugFormat(t *testing.T) {
	var buf bytes.Buffer
	DebugOut = &buf
	defer func() { DebugOut = nil }()

	debugf("test_event", "key=%d", 42)

	out := buf.String()
	re := regexp.MustCompile(`\[debug \d{2}:\d{2}:\d{2}\] test_event: key=42\n`)
	if !re.MatchString(out) {
		t.Errorf("debug output does not match expected format, got: %q", out)
	}
}
