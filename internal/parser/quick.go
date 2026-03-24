package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// quickTaskDirRe matches YYMMDD-ID-description subdirectory names.
var quickTaskDirRe = regexp.MustCompile(`^(\d{6})-(\w+)-(.+)$`)

// parseQuickTasks walks quickDir and returns a slice of QuickTask, sorted newest-first.
// Returns nil if the directory is absent or empty (no matching entries).
func parseQuickTasks(quickDir string) []QuickTask {
	entries, err := os.ReadDir(quickDir)
	if err != nil {
		return nil
	}

	var tasks []QuickTask
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		m := quickTaskDirRe.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}
		date := m[1]
		id := m[2]
		rest := m[3]
		displayName := strings.ReplaceAll(rest, "-", " ")

		taskDir := filepath.Join(quickDir, entry.Name())
		planFile := date + "-" + id + "-PLAN.md"
		summaryFile := date + "-" + id + "-SUMMARY.md"

		_, errSummary := os.Stat(filepath.Join(taskDir, summaryFile))
		_, errPlan := os.Stat(filepath.Join(taskDir, planFile))

		var status string
		switch {
		case errSummary == nil:
			status = StatusComplete
		case errPlan == nil:
			status = StatusInProgress
		default:
			status = StatusPending
		}

		debugf("quick_task_dir", "%s status=%q display=%q", entry.Name(), status, displayName)

		tasks = append(tasks, QuickTask{
			DirName:     entry.Name(),
			DisplayName: displayName,
			Date:        date,
			Status:      status,
		})
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Date > tasks[j].Date
	})

	return tasks
}
