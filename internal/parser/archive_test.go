package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testdataMilestonesDir = "testdata/milestones"
const testdataMilestonesFile = "testdata/milestones/MILESTONES.md"

func TestParseArchivedMilestones(t *testing.T) {
	result := parseArchivedMilestones(testdataMilestonesDir, testdataMilestonesFile)
	if len(result) != 2 {
		t.Fatalf("expected 2 ArchivedMilestones, got %d", len(result))
	}
}

func TestParseArchivedMilestones_Name(t *testing.T) {
	result := parseArchivedMilestones(testdataMilestonesDir, testdataMilestonesFile)
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(result))
	}
	if result[0].Name != "v1.1" {
		t.Errorf("expected result[0].Name=%q (newest-first), got %q", "v1.1", result[0].Name)
	}
	if result[1].Name != "v1.0" {
		t.Errorf("expected result[1].Name=%q, got %q", "v1.0", result[1].Name)
	}
}

func TestParseArchivedMilestones_PhaseCount(t *testing.T) {
	result := parseArchivedMilestones(testdataMilestonesDir, testdataMilestonesFile)
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(result))
	}
	// v1.1-phases has 1 subdir (07-parser)
	if result[0].PhaseCount != 1 {
		t.Errorf("expected result[0].PhaseCount=1 (v1.1 has 1 subdir), got %d", result[0].PhaseCount)
	}
	// v1.0-phases has 2 subdirs (01-core, 02-data)
	if result[1].PhaseCount != 2 {
		t.Errorf("expected result[1].PhaseCount=2 (v1.0 has 2 subdirs), got %d", result[1].PhaseCount)
	}
}

func TestParseArchivedMilestones_CompletionDate(t *testing.T) {
	result := parseArchivedMilestones(testdataMilestonesDir, testdataMilestonesFile)
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(result))
	}
	if result[0].CompletionDate != "2026-03-25" {
		t.Errorf("expected result[0].CompletionDate=%q, got %q", "2026-03-25", result[0].CompletionDate)
	}
	if result[1].CompletionDate != "2026-03-23" {
		t.Errorf("expected result[1].CompletionDate=%q, got %q", "2026-03-23", result[1].CompletionDate)
	}
}

func TestParseArchivedMilestones_MissingMilestonesFile(t *testing.T) {
	result := parseArchivedMilestones(testdataMilestonesDir, "/nonexistent/MILESTONES.md")
	if result == nil {
		t.Fatal("expected non-nil slice even when MILESTONES.md is missing")
	}
	// Entries should still be returned, but with empty CompletionDate
	for _, m := range result {
		if m.CompletionDate != "" {
			t.Errorf("expected empty CompletionDate when MILESTONES.md missing, got %q for %s", m.CompletionDate, m.Name)
		}
	}
}

func TestParseArchivedMilestones_SkipsMalformed(t *testing.T) {
	dir := t.TempDir()
	// Create a dir that doesn't match the vX.Y-phases pattern
	if err := os.MkdirAll(filepath.Join(dir, "not-a-version"), 0755); err != nil {
		t.Fatal(err)
	}
	result := parseArchivedMilestones(dir, "/nonexistent/MILESTONES.md")
	if len(result) != 0 {
		t.Errorf("expected 0 results for non-matching dirs, got %d", len(result))
	}
}

func TestParseArchivedMilestones_MissingDir(t *testing.T) {
	result := parseArchivedMilestones("/nonexistent/milestones/dir", "/nonexistent/MILESTONES.md")
	if result == nil {
		t.Fatal("expected non-nil slice for nonexistent dir, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected len=0 for nonexistent dir, got %d", len(result))
	}
}

func TestParseArchivedMilestones_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	result := parseArchivedMilestones(dir, "/nonexistent/MILESTONES.md")
	if result == nil {
		t.Fatal("expected non-nil slice for empty dir, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected len=0 for empty dir, got %d", len(result))
	}
}

func TestParseArchivedMilestones_Sort(t *testing.T) {
	result := parseArchivedMilestones(testdataMilestonesDir, testdataMilestonesFile)
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(result))
	}
	// Newest first: v1.1 > v1.0 lexicographically
	if result[0].Name <= result[1].Name {
		t.Errorf("expected newest-first sort: result[0].Name=%q should be > result[1].Name=%q", result[0].Name, result[1].Name)
	}
}

func TestParseArchivedMilestones_Debug(t *testing.T) {
	orig := DebugOut
	defer func() { DebugOut = orig }()

	var buf bytes.Buffer
	DebugOut = &buf

	parseArchivedMilestones(testdataMilestonesDir, testdataMilestonesFile)

	output := buf.String()
	if !strings.Contains(output, "archive_dir") {
		t.Errorf("expected debug output to contain %q, got: %q", "archive_dir", output)
	}
}
