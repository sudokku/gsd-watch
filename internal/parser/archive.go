package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

// archiveDirRe matches vX.Y-phases directory names, capturing the version string.
var archiveDirRe = regexp.MustCompile(`^(v\d+\.\d+)-phases$`)

// parseArchivedMilestones scans milestonesDir for vX.Y-phases/ directories and returns
// a slice of ArchivedMilestone sorted newest-first by version string.
// Never returns an error — missing/malformed dirs are skipped with optional debugf.
// Returns empty (not nil) slice when no archives found (per D-08).
func parseArchivedMilestones(milestonesDir, milestonesFile string) []ArchivedMilestone {
	result := []ArchivedMilestone{} // empty, not nil (D-08)

	entries, err := os.ReadDir(milestonesDir)
	if err != nil {
		return result
	}

	// Pre-read MILESTONES.md once for all version lookups
	milestonesData, _ := os.ReadFile(milestonesFile) // nil on error — lookups return ""

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		m := archiveDirRe.FindStringSubmatch(entry.Name())
		if m == nil {
			debugf("archive_dir", "skipping non-matching dir %s", entry.Name())
			continue
		}
		version := m[1] // e.g. "v1.0"

		// Count subdirectories inside vX.Y-phases/ for PhaseCount (D-05)
		phasesPath := filepath.Join(milestonesDir, entry.Name())
		phaseEntries, err := os.ReadDir(phasesPath)
		if err != nil {
			debugf("archive_dir", "cannot read phases dir %s: skipped", entry.Name())
			continue
		}
		phaseCount := 0
		for _, pe := range phaseEntries {
			if pe.IsDir() {
				phaseCount++
			}
		}

		// Lookup completion date from MILESTONES.md (D-03)
		completionDate := lookupCompletionDate(milestonesData, version)

		debugf("archive_dir", "detected %s name=%q phases=%d date=%q", entry.Name(), version, phaseCount, completionDate)

		result = append(result, ArchivedMilestone{
			Name:           version,
			PhaseCount:     phaseCount,
			CompletionDate: completionDate,
		})
	}

	// Sort newest-first by version string (lexicographic descending — works for vX.Y single-digit)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name > result[j].Name
	})

	return result
}

// lookupCompletionDate searches milestonesData for a heading matching the given version
// and returns the Shipped date, or empty string if not found.
func lookupCompletionDate(milestonesData []byte, version string) string {
	if milestonesData == nil {
		return ""
	}
	// Build regex: ^## v1.0 ... (Shipped: YYYY-MM-DD)
	// Use QuoteMeta to escape the dot in version (e.g. v1.0 → v1\.0)
	re := regexp.MustCompile(`(?m)^## ` + regexp.QuoteMeta(version) + `[^\n]*\(Shipped:\s*(\d{4}-\d{2}-\d{2})\)`)
	m := re.FindSubmatch(milestonesData)
	if m == nil {
		return ""
	}
	return string(m[1])
}
