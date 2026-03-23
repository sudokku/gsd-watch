package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// testdataPhasesDir is the root of the phases fixture directory.
const testdataPhasesDir = "testdata/phases"

func TestParsePhases_ValidStructure(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	if len(phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(phases))
	}
	if phases[0].DirName != "01-core-tui-scaffold" {
		t.Errorf("expected first phase DirName=01-core-tui-scaffold, got %q", phases[0].DirName)
	}
	if phases[1].DirName != "02-live-data-layer" {
		t.Errorf("expected second phase DirName=02-live-data-layer, got %q", phases[1].DirName)
	}
	if phases[2].DirName == "" || phases[2].DirName != "09-roadmap-absent" {
		t.Errorf("expected third phase DirName=09-roadmap-absent, got %q", phases[2].DirName)
	}
	if phases[2].Name != "Phase 9: roadmap absent" {
		t.Errorf("expected third phase Name=%q, got %q", "Phase 9: roadmap absent", phases[2].Name)
	}
}

func TestParsePhases_PlanFromFrontmatter(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	// 01-02-PLAN.md has status: in_progress, no SUMMARY.md
	phase1 := phases[0]
	var plan02 *Plan
	for i := range phase1.Plans {
		if phase1.Plans[i].Filename == "01-02-PLAN.md" {
			plan02 = &phase1.Plans[i]
			break
		}
	}
	if plan02 == nil {
		t.Fatal("01-02-PLAN.md not found in phase 1 plans")
	}
	if plan02.Status != StatusInProgress {
		t.Errorf("expected 01-02-PLAN.md Status=in_progress, got %q", plan02.Status)
	}
	if plan02.Wave != 2 {
		t.Errorf("expected 01-02-PLAN.md Wave=2, got %d", plan02.Wave)
	}
}

func TestParsePhases_SummaryOverride(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	// 01-01-PLAN.md has status: pending BUT 01-01-SUMMARY.md exists -> complete
	phase1 := phases[0]
	var plan01 *Plan
	for i := range phase1.Plans {
		if phase1.Plans[i].Filename == "01-01-PLAN.md" {
			plan01 = &phase1.Plans[i]
			break
		}
	}
	if plan01 == nil {
		t.Fatal("01-01-PLAN.md not found in phase 1 plans")
	}
	if plan01.Status != StatusComplete {
		t.Errorf("expected 01-01-PLAN.md Status=complete (SUMMARY.md override), got %q", plan01.Status)
	}
}

func TestParsePhases_BadgeDetection(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	phase1 := phases[0]

	hasDiscussed := false
	hasResearched := false
	hasVerified := false
	for _, b := range phase1.Badges {
		switch b {
		case BadgeDiscussed:
			hasDiscussed = true
		case BadgeResearched:
			hasResearched = true
		case BadgeVerified:
			hasVerified = true
		}
	}
	if !hasDiscussed {
		t.Errorf("expected badge %q in phase 1 (01-CONTEXT.md exists), badges=%v", BadgeDiscussed, phase1.Badges)
	}
	if !hasResearched {
		t.Errorf("expected badge %q in phase 1 (01-RESEARCH.md exists), badges=%v", BadgeResearched, phase1.Badges)
	}
	if hasVerified {
		t.Errorf("did not expect badge %q in phase 1 (no 01-VERIFICATION.md), badges=%v", BadgeVerified, phase1.Badges)
	}
}

func TestParsePhases_PhaseStatusComplete(t *testing.T) {
	// phase 1: plan01 is complete (summary override), plan02 is in_progress -> in_progress
	// We need a phase where ALL plans are complete. Create a temp dir.
	dir := t.TempDir()
	phaseDir := filepath.Join(dir, "01-all-complete")
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create plan with pending status + summary override
	writePlanFile(t, filepath.Join(phaseDir, "01-01-PLAN.md"), "pending", 1)
	writeEmptyFile(t, filepath.Join(phaseDir, "01-01-SUMMARY.md"))

	phases := parsePhases(dir, map[int]string{}, 0, 0)
	if len(phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(phases))
	}
	if phases[0].Status != StatusComplete {
		t.Errorf("expected phase status=complete (all plans complete), got %q", phases[0].Status)
	}
}

func TestParsePhases_PhaseStatusInProgress(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	// phase1: plan01=complete(override), plan02=in_progress -> in_progress
	if phases[0].Status != StatusInProgress {
		t.Errorf("expected phase 1 status=in_progress, got %q", phases[0].Status)
	}
}

func TestParsePhases_PhaseStatusPending(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	// phase2: 02-01-PLAN.md has status=pending, no SUMMARY.md -> pending
	if phases[1].Status != StatusPending {
		t.Errorf("expected phase 2 status=pending, got %q", phases[1].Status)
	}
}

func TestParsePhases_SkipNonPlanFiles(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	phase1 := phases[0]
	for _, p := range phase1.Plans {
		if p.Filename == "01-CONTEXT.md" || p.Filename == "01-RESEARCH.md" || p.Filename == "01-01-SUMMARY.md" {
			t.Errorf("non-plan file %q was incorrectly parsed as a plan", p.Filename)
		}
	}
}

func TestParsePhases_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	phaseDir := filepath.Join(dir, "01-empty")
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		t.Fatal(err)
	}
	phases := parsePhases(dir, map[int]string{}, 0, 0)
	if len(phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(phases))
	}
	if len(phases[0].Plans) != 0 {
		t.Errorf("expected empty Plans slice for empty dir, got %d plans", len(phases[0].Plans))
	}
	if phases[0].Status != StatusPending {
		t.Errorf("expected Status=pending for empty phase, got %q", phases[0].Status)
	}
}

func TestParsePhases_MissingPhasesDir(t *testing.T) {
	phases := parsePhases("/nonexistent/path/to/phases", map[int]string{}, 0, 0)
	if phases != nil {
		t.Errorf("expected nil slice for nonexistent path, got %v", phases)
	}
}

func TestParsePhases_ActivePlan(t *testing.T) {
	// activePhase=1, activePlan=1 -> 01-01 plan IsActive=true
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 1, 1)
	phase1 := phases[0]
	var plan01 *Plan
	for i := range phase1.Plans {
		if phase1.Plans[i].Filename == "01-01-PLAN.md" {
			plan01 = &phase1.Plans[i]
			break
		}
	}
	if plan01 == nil {
		t.Fatal("01-01-PLAN.md not found")
	}
	if !plan01.IsActive {
		t.Errorf("expected 01-01-PLAN.md IsActive=true with activePhase=1, activePlan=1")
	}
	// 01-02-PLAN.md should NOT be active
	for _, p := range phase1.Plans {
		if p.Filename == "01-02-PLAN.md" && p.IsActive {
			t.Errorf("01-02-PLAN.md should not be active when activePlan=1")
		}
	}
}

func TestParsePhases_PlanTitle(t *testing.T) {
	phases := parsePhases(testdataPhasesDir, map[int]string{}, 0, 0)
	phase1 := phases[0]

	// 01-02-PLAN.md has <objective>\nTree model + viewport\n</objective>
	var plan02 *Plan
	for i := range phase1.Plans {
		if phase1.Plans[i].Filename == "01-02-PLAN.md" {
			plan02 = &phase1.Plans[i]
			break
		}
	}
	if plan02 == nil {
		t.Fatal("01-02-PLAN.md not found")
	}
	if plan02.Title != "Tree model + viewport" {
		t.Errorf("expected Title=%q from <objective>, got %q", "Tree model + viewport", plan02.Title)
	}

	// 01-01-PLAN.md has <objective>\nFoundation: types, messages, mock data\n</objective>
	var plan01 *Plan
	for i := range phase1.Plans {
		if phase1.Plans[i].Filename == "01-01-PLAN.md" {
			plan01 = &phase1.Plans[i]
			break
		}
	}
	if plan01 == nil {
		t.Fatal("01-01-PLAN.md not found")
	}
	// If title extracted, use it; otherwise filename stem
	if plan01.Title == "" {
		t.Errorf("expected non-empty Title for 01-01-PLAN.md")
	}
}

func TestParsePhases_RoadmapAbsentSorting(t *testing.T) {
	// Provide phaseNames for 1 and 2 but NOT 9 — phase 9 must still sort correctly
	phaseNames := map[int]string{1: "Core TUI Scaffold", 2: "Live Data Layer"}
	phases := parsePhases(testdataPhasesDir, phaseNames, 0, 0)

	if len(phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(phases))
	}

	// Phase 9 should be at index 2 (sorted after 1 and 2)
	if phases[2].DirName != "09-roadmap-absent" {
		t.Errorf("expected phases[2].DirName=09-roadmap-absent, got %q", phases[2].DirName)
	}
	if phases[2].Name != "Phase 9: roadmap absent" {
		t.Errorf("expected phases[2].Name=%q, got %q", "Phase 9: roadmap absent", phases[2].Name)
	}

	// Verify strictly ascending sort order
	n0 := extractPhaseNum(phases[0].Name)
	n1 := extractPhaseNum(phases[1].Name)
	n2 := extractPhaseNum(phases[2].Name)
	if !(n0 < n1 && n1 < n2) {
		t.Errorf("phases not in ascending order: %d, %d, %d", n0, n1, n2)
	}
}

// helpers

func writePlanFile(t *testing.T, path, status string, wave int) {
	t.Helper()
	content := "---\nstatus: " + status + "\nwave: " + string(rune('0'+wave)) + "\n---\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeEmptyFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
}
