package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testdataQuickDir = "testdata/quick"

func TestParseQuickTasks_Complete(t *testing.T) {
	tasks := parseQuickTasks(testdataQuickDir)
	var completeTask *QuickTask
	for i := range tasks {
		if tasks[i].DirName == "260101-ab1-sample-complete-task" {
			completeTask = &tasks[i]
			break
		}
	}
	if completeTask == nil {
		t.Fatal("260101-ab1-sample-complete-task not found in tasks")
	}
	if completeTask.Status != StatusComplete {
		t.Errorf("expected Status=%q for dir with PLAN+SUMMARY, got %q", StatusComplete, completeTask.Status)
	}
}

func TestParseQuickTasks_InProgress(t *testing.T) {
	tasks := parseQuickTasks(testdataQuickDir)
	var inProgressTask *QuickTask
	for i := range tasks {
		if tasks[i].DirName == "260215-cd2-another-in-progress-task" {
			inProgressTask = &tasks[i]
			break
		}
	}
	if inProgressTask == nil {
		t.Fatal("260215-cd2-another-in-progress-task not found in tasks")
	}
	if inProgressTask.Status != StatusInProgress {
		t.Errorf("expected Status=%q for dir with PLAN only, got %q", StatusInProgress, inProgressTask.Status)
	}
}

func TestParseQuickTasks_Pending(t *testing.T) {
	dir := t.TempDir()
	// Create a task dir with no PLAN.md or SUMMARY.md
	taskDir := filepath.Join(dir, "260301-ab3-pending-task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	tasks := parseQuickTasks(dir)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Status != StatusPending {
		t.Errorf("expected Status=%q for dir with no files, got %q", StatusPending, tasks[0].Status)
	}
}

func TestParseQuickTasks_MissingDir(t *testing.T) {
	tasks := parseQuickTasks("/nonexistent/path/to/quick")
	if tasks != nil {
		t.Errorf("expected nil for nonexistent dir, got %v", tasks)
	}
}

func TestParseQuickTasks_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	tasks := parseQuickTasks(dir)
	if tasks != nil {
		t.Errorf("expected nil for empty dir, got %v", tasks)
	}
}

func TestParseQuickTasks_Sort(t *testing.T) {
	tasks := parseQuickTasks(testdataQuickDir)
	if len(tasks) < 2 {
		t.Fatalf("expected at least 2 tasks, got %d", len(tasks))
	}
	// Newest date first: 260215 > 260101
	if tasks[0].Date != "260215" {
		t.Errorf("expected first task Date=%q (newest), got %q", "260215", tasks[0].Date)
	}
	if tasks[1].Date != "260101" {
		t.Errorf("expected second task Date=%q (older), got %q", "260101", tasks[1].Date)
	}
}

func TestParseQuickTasks_DisplayName(t *testing.T) {
	dir := t.TempDir()
	taskDir := filepath.Join(dir, "260323-re2-fix-gsd-watch-sidebar")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	tasks := parseQuickTasks(dir)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].DisplayName != "fix gsd watch sidebar" {
		t.Errorf("expected DisplayName=%q, got %q", "fix gsd watch sidebar", tasks[0].DisplayName)
	}
}

func TestParseQuickTasks_SkipsNonDirs(t *testing.T) {
	dir := t.TempDir()
	// Create a flat file that should be ignored
	flatFile := filepath.Join(dir, "260323-re2-this-is-a-file.md")
	if err := os.WriteFile(flatFile, []byte("flat file"), 0644); err != nil {
		t.Fatal(err)
	}
	tasks := parseQuickTasks(dir)
	if tasks != nil {
		t.Errorf("expected nil (non-dirs skipped), got %v", tasks)
	}
}

func TestParseQuickTasks_SkipsMalformedDirs(t *testing.T) {
	dir := t.TempDir()
	// Create dirs that don't match the YYMMDD-ID-slug pattern
	malformedDirs := []string{
		"not-a-valid-dir",
		"12345-ab1-only-five-digits",
		"1234567-ab1-seven-digits",
		"260101",
	}
	for _, d := range malformedDirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}
	tasks := parseQuickTasks(dir)
	if tasks != nil {
		t.Errorf("expected nil (malformed dirs skipped), got %v", tasks)
	}
}

func TestParseQuickTasks_Debug(t *testing.T) {
	orig := DebugOut
	defer func() { DebugOut = orig }()

	var buf bytes.Buffer
	DebugOut = &buf

	parseQuickTasks(testdataQuickDir)

	output := buf.String()
	if !strings.Contains(output, "quick_task_dir") {
		t.Errorf("expected debug output to contain %q, got: %q", "quick_task_dir", output)
	}
}

func TestParseProject_QuickTasks(t *testing.T) {
	data := ParseProject(testdataProjectDir)

	if len(data.QuickTasks) != 2 {
		t.Fatalf("expected 2 QuickTasks, got %d", len(data.QuickTasks))
	}
	// Sorted newest first: 260215 before 260101
	if data.QuickTasks[0].Date != "260215" {
		t.Errorf("expected QuickTasks[0].Date=%q (newest), got %q", "260215", data.QuickTasks[0].Date)
	}
	if data.QuickTasks[0].Status != StatusInProgress {
		t.Errorf("expected QuickTasks[0].Status=%q, got %q", StatusInProgress, data.QuickTasks[0].Status)
	}
	if data.QuickTasks[1].Date != "260101" {
		t.Errorf("expected QuickTasks[1].Date=%q (older), got %q", "260101", data.QuickTasks[1].Date)
	}
	if data.QuickTasks[1].Status != StatusComplete {
		t.Errorf("expected QuickTasks[1].Status=%q, got %q", StatusComplete, data.QuickTasks[1].Status)
	}
}
