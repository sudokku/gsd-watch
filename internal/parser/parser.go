package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ParseProject reads .planning/ directory and returns a fully populated ProjectData.
// This is the single public entry point for all parsing.
// NEVER returns an error — always returns best-effort ProjectData with "unknown" defaults.
// root parameter is the path to the .planning/ directory (e.g. "/path/to/project/.planning").
func ParseProject(root string) ProjectData {
	// Project name: PROJECT.md H1 at the project root (one level above .planning/),
	// falling back to the project directory name if PROJECT.md is absent or has no H1.
	projectRoot := filepath.Dir(root)
	name := filepath.Base(projectRoot)
	if projectBytes, err := os.ReadFile(filepath.Join(projectRoot, "PROJECT.md")); err == nil {
		h1Re := regexp.MustCompile(`(?m)^# (.+)`)
		if m := h1Re.FindSubmatch(projectBytes); len(m) > 1 {
			name = strings.TrimSpace(string(m[1]))
		}
	}

	data := ProjectData{
		Name:          name,
		ModelProfile:  "unknown",
		Mode:          "unknown",
		CurrentAction: "unknown",
		LastUpdated:   time.Now(),
	}

	// Parse config.json.
	if cfg, err := parseConfig(filepath.Join(root, "config.json")); err == nil {
		data.ModelProfile = cfg.ModelProfile
		data.Mode = cfg.Mode
	}

	// Parse STATE.md.
	if st, err := parseState(filepath.Join(root, "STATE.md")); err == nil {
		if st.StoppedAt != "" {
			data.CurrentAction = st.StoppedAt
		}
		// Parse ROADMAP.md for phase names.
		phaseNames := parseRoadmap(filepath.Join(root, "ROADMAP.md"))

		// Parse phases directory.
		data.Phases = parsePhases(
			filepath.Join(root, "phases"),
			phaseNames,
			st.ActivePhase,
			st.ActivePlan,
		)
	} else {
		// STATE.md missing — still try roadmap + phases without active plan.
		phaseNames := parseRoadmap(filepath.Join(root, "ROADMAP.md"))
		data.Phases = parsePhases(filepath.Join(root, "phases"), phaseNames, 0, 0)
	}

	// Compute progress from actual phase completion on disk (not STATE.md percent,
	// which is updated infrequently by gsd-tools and reflects milestone-level accounting).
	if len(data.Phases) > 0 {
		var done int
		for _, ph := range data.Phases {
			if ph.Status == "complete" {
				done++
			}
		}
		data.ProgressPercent = float64(done) / float64(len(data.Phases))
	}

	return data
}
