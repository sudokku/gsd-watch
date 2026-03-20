package parser

import (
	"path/filepath"
	"time"
)

// ParseProject reads .planning/ directory and returns a fully populated ProjectData.
// This is the single public entry point for all parsing.
// NEVER returns an error — always returns best-effort ProjectData with "unknown" defaults.
// root parameter is the path to the .planning/ directory (e.g. "/path/to/project/.planning").
func ParseProject(root string) ProjectData {
	data := ProjectData{
		Name:          "unknown",
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
		if st.MilestoneName != "" {
			data.Name = st.MilestoneName
		}
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
