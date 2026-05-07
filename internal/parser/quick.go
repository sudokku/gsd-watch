package parser

import (
	"bufio"
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

		if rich := readRichDescription(taskDir, date+"-"+id); rich != "" {
			displayName = rich
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

// readRichDescription tries PLAN.md <objective>, SUMMARY.md **One-liner:**, and
// SUMMARY.md H1 in order. Returns the first non-empty result, or "" when no
// rich source is available.
func readRichDescription(taskDir, base string) string {
	if s := scanForObjective(filepath.Join(taskDir, base+"-PLAN.md")); s != "" {
		return s
	}
	summaryPath := filepath.Join(taskDir, base+"-SUMMARY.md")
	if s := scanForOneLiner(summaryPath); s != "" {
		return s
	}
	if s := scanForSummaryHeading(summaryPath, base); s != "" {
		return s
	}
	return ""
}

// scanForObjective opens path, finds the line equal to "<objective>", then
// returns the sanitized first non-empty content line. Returns "" on any error.
func scanForObjective(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	inObjective := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if !inObjective {
			if trimmed == "<objective>" {
				inObjective = true
			}
			continue
		}
		if trimmed == "</objective>" {
			return ""
		}
		if trimmed == "" {
			continue
		}
		return sanitizeDescription(trimmed)
	}
	return ""
}

// scanForOneLiner opens path and returns the sanitized text following a
// `**One-liner:**` marker on a line. Returns "" on any error or no match.
func scanForOneLiner(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	const marker = "**One-liner:**"
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(trimmed, marker) {
			return sanitizeDescription(strings.TrimPrefix(trimmed, marker))
		}
	}
	return ""
}

// scanForSummaryHeading opens path and returns the title captured from a
// SUMMARY-style H1 such as "# Quick Task <base>: Title — Summary".
// Returns "" on any error or non-match.
func scanForSummaryHeading(path, base string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	re, err := regexp.Compile(`^# Quick (?:Task )?` + regexp.QuoteMeta(base) + `:?\s*(.+?)\s*(?:—\s*)?Summary\s*$`)
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(strings.TrimSpace(line), "# ") {
			continue
		}
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			return ""
		}
		title := strings.TrimSpace(m[1])
		if title == "" {
			return ""
		}
		return sanitizeDescription(title)
	}
	return ""
}

// sanitizeDescription normalizes an extracted description line for display.
func sanitizeDescription(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "</objective>")
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimPrefix(s, "* ")
	// collapse internal whitespace runs
	s = strings.Join(strings.Fields(s), " ")
	// first sentence only when a sentence break exists
	if i := strings.Index(s, ". "); i > 0 {
		s = s[:i+1]
	}
	const maxDisplayLen = 200
	if len([]rune(s)) > maxDisplayLen {
		r := []rune(s)
		s = string(r[:maxDisplayLen]) + "…"
	}
	return s
}
