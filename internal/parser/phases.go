package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	phaseDirRe = regexp.MustCompile(`^(\d{2})-`)
	planFileRe = regexp.MustCompile(`^\d{2}-\d{2}-PLAN\.md$`)
)

// badgeFiles maps filename suffix to badge constant.
var badgeFiles = []struct {
	suffix string
	badge  string
}{
	{"CONTEXT.md", BadgeDiscussed},
	{"RESEARCH.md", BadgeResearched},
	{"VERIFICATION.md", BadgeVerified},
	{"UAT.md", BadgeUAT},
}

// parsePhases walks the phases directory, discovers phase dirs, parses plans,
// detects badges, and applies SUMMARY.md override. phaseNames provides
// human-readable names from ROADMAP.md. activePhase/activePlan identify the
// currently active plan (0 means none).
func parsePhases(phasesDir string, phaseNames map[int]string, activePhase, activePlan int) []Phase {
	entries, err := os.ReadDir(phasesDir)
	if err != nil {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	seenPhaseNums := map[int]bool{}
	var phases []Phase
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		m := phaseDirRe.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}
		phaseNum, _ := strconv.Atoi(m[1])
		seenPhaseNums[phaseNum] = true
		phaseDir := filepath.Join(phasesDir, entry.Name())

		// Parse plans in this phase directory.
		plans := parsePlansInDir(phaseDir, phaseNum, activePhase, activePlan)

		// Detect badges: look for {NN}-{TYPE}.md files.
		prefix := m[1] // e.g. "01"
		var badges []string
		for _, bf := range badgeFiles {
			badgePath := filepath.Join(phaseDir, prefix+"-"+bf.suffix)
			if _, statErr := os.Stat(badgePath); statErr == nil {
				badges = append(badges, bf.badge)
			}
		}

		// Phase name from ROADMAP.md, fallback to directory name.
		name := entry.Name()
		if rname, ok := phaseNames[phaseNum]; ok {
			name = fmt.Sprintf("Phase %d: %s", phaseNum, rname)
		}

		// Derive phase status from plan statuses.
		status := derivePhaseStatus(plans)

		phases = append(phases, Phase{
			DirName: entry.Name(),
			Name:    name,
			Status:  status,
			Badges:  badges,
			Plans:   plans,
		})
	}

	// Add stub entries for roadmap phases that have no directory yet.
	for phaseNum, phaseName := range phaseNames {
		if seenPhaseNums[phaseNum] {
			continue
		}
		phases = append(phases, Phase{
			Name:   fmt.Sprintf("Phase %d: %s", phaseNum, phaseName),
			Status: StatusPending,
		})
	}

	// Sort all phases by phase number.
	sort.Slice(phases, func(i, j int) bool {
		return extractPhaseNum(phases[i].Name) < extractPhaseNum(phases[j].Name)
	})

	return phases
}

// extractPhaseNum pulls the leading integer from a phase name like "Phase 3: File Watching".
// Returns 0 if no number is found.
var phaseNumRe = regexp.MustCompile(`Phase (\d+)`)

func extractPhaseNum(name string) int {
	if m := phaseNumRe.FindStringSubmatch(name); len(m) > 1 {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

// parsePlansInDir reads all NN-NN-PLAN.md files in a phase directory.
func parsePlansInDir(phaseDir string, phaseNum, activePhase, activePlan int) []Plan {
	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var plans []Plan
	for _, entry := range entries {
		if entry.IsDir() || !planFileRe.MatchString(entry.Name()) {
			continue
		}
		planPath := filepath.Join(phaseDir, entry.Name())
		plan, err := parsePlan(planPath)
		if err != nil {
			continue // skip malformed plans silently (PARSE-08)
		}
		plan.Filename = entry.Name()

		// Title fallback: if parsePlan didn't extract a title from <objective>,
		// use filename stem (e.g. "01-02-PLAN" from "01-02-PLAN.md").
		if plan.Title == "" {
			plan.Title = strings.TrimSuffix(entry.Name(), ".md")
		}

		// SUMMARY.md override (PARSE-02): if NN-NN-SUMMARY.md exists, status = complete.
		summaryName := strings.Replace(entry.Name(), "-PLAN.md", "-SUMMARY.md", 1)
		summaryPath := filepath.Join(phaseDir, summaryName)
		if _, statErr := os.Stat(summaryPath); statErr == nil {
			plan.Status = StatusComplete
		}

		// Active plan detection.
		// Extract plan number from filename: "01-02-PLAN.md" -> plan number 2.
		parts := strings.SplitN(entry.Name(), "-", 3) // ["01", "02", "PLAN.md"]
		if len(parts) >= 2 {
			planNum, _ := strconv.Atoi(parts[1])
			if phaseNum == activePhase && planNum == activePlan {
				plan.IsActive = true
			}
		}

		plans = append(plans, plan)
	}
	return plans
}

// derivePhaseStatus computes phase status from constituent plan statuses.
// All complete -> complete; any in_progress -> in_progress; else pending.
func derivePhaseStatus(plans []Plan) string {
	if len(plans) == 0 {
		return StatusPending
	}
	allComplete := true
	for _, p := range plans {
		if p.Status == StatusInProgress {
			return StatusInProgress
		}
		if p.Status != StatusComplete {
			allComplete = false
		}
	}
	if allComplete {
		return StatusComplete
	}
	return StatusPending
}
