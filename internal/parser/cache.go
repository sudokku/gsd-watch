package parser

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectCache wraps existing parsers with an incremental update mechanism.
// On each file change, only the affected file is re-parsed while the rest of
// the project data is served from the in-memory cache, satisfying WATCH-04.
type ProjectCache struct {
	root   string               // path to .planning/ directory
	data   ProjectData          // cached project data
	mtimes map[string]time.Time // path -> last known mtime
}

// NewCache returns a new ProjectCache rooted at root (the .planning/ directory).
// The cache is empty; call ParseFull() to populate it.
func NewCache(root string) *ProjectCache {
	return &ProjectCache{
		root:   root,
		mtimes: make(map[string]time.Time),
	}
}

// ParseFull performs a full parse of the project tree, populates the mtime index
// for every file under root, and returns the resulting ProjectData.
// Subsequent calls to Update(path) will only re-parse the affected file.
func (c *ProjectCache) ParseFull() ProjectData {
	c.data = ParseProject(c.root)

	// Walk the entire tree to seed the mtime index.
	_ = filepath.WalkDir(c.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !d.IsDir() {
			if info, statErr := d.Info(); statErr == nil {
				c.mtimes[path] = info.ModTime()
			}
		}
		return nil
	})

	return c.data
}

// Update re-parses only the file at path if its mtime has changed since the last
// parse, then returns the updated (or unchanged) ProjectData.
//
// Routing table:
//   - STATE.md        → re-parse state fields + phases (active markers may change)
//   - config.json     → re-parse model/mode only
//   - ROADMAP.md      → full re-parse (phase names affect the whole tree)
//   - NN-NN-PLAN.md   → re-parse plans in that phase directory only
//   - *-SUMMARY.md    → re-parse plans in that phase directory only
//   - badge files     → re-detect badges for that phase
//   - anything else   → full re-parse (safe fallback)
func (c *ProjectCache) Update(path string) ProjectData {
	// Mtime guard: skip re-parse if the file has not changed.
	info, err := os.Stat(path)
	if err == nil {
		if c.mtimes[path].Equal(info.ModTime()) {
			debugf("cache", "HIT %s (mtime unchanged)", filepath.Base(path))
			return c.data
		}
		c.mtimes[path] = info.ModTime()
		debugf("cache", "MISS %s (mtime changed)", filepath.Base(path))
	}
	// If os.Stat failed (file deleted), fall through to full re-parse.

	base := filepath.Base(path)

	switch {
	case base == "STATE.md":
		c.updateFromState(path)

	case base == "config.json":
		c.updateFromConfig(path)

	case base == "ROADMAP.md":
		// Roadmap changes affect phase names globally — full re-parse is simplest and correct.
		c.data = ParseProject(c.root)

	case planFileRe.MatchString(base) || strings.HasSuffix(base, "-SUMMARY.md"):
		c.updatePhasePlans(path)

	case isBadgeFile(base):
		c.updatePhaseBadges(path)

	default:
		// Safe fallback: unknown file type, full re-parse.
		c.data = ParseProject(c.root)
	}

	c.data.LastUpdated = time.Now()
	return c.data
}

// updateFromState re-parses STATE.md and updates CurrentAction, ProgressPercent,
// Name, and Phases (to refresh active plan markers).
func (c *ProjectCache) updateFromState(path string) {
	st, err := parseState(path)
	if err != nil {
		return
	}
	if st.MilestoneName != "" {
		c.data.Name = st.MilestoneName
	}
	c.data.CurrentAction = st.StoppedAt
	if c.data.CurrentAction == "" {
		c.data.CurrentAction = "unknown"
	}

	// Re-parse phases so active plan markers reflect the new STATE.md values.
	phaseNames := parseRoadmap(filepath.Join(c.root, "ROADMAP.md"))
	c.data.Phases = parsePhases(
		filepath.Join(c.root, "phases"),
		phaseNames,
		st.ActivePhase,
		st.ActivePlan,
	)

	// Recompute progress from actual phase completion (mirrors ParseProject logic).
	if len(c.data.Phases) > 0 {
		var done int
		for _, ph := range c.data.Phases {
			if ph.Status == "complete" {
				done++
			}
		}
		c.data.ProgressPercent = float64(done) / float64(len(c.data.Phases))
	}
}

// updateFromConfig re-parses config.json and updates ModelProfile and Mode.
func (c *ProjectCache) updateFromConfig(path string) {
	cfg, err := parseConfig(path)
	if err != nil {
		return
	}
	c.data.ModelProfile = cfg.ModelProfile
	c.data.Mode = cfg.Mode
}

// updatePhasePlans re-parses only the plans in the phase that owns the changed file.
func (c *ProjectCache) updatePhasePlans(changedPath string) {
	phaseDir := filepath.Dir(changedPath)
	phaseDirName := filepath.Base(phaseDir)

	// Extract phase number from directory name (e.g. "01-test" -> 1).
	m := phaseDirRe.FindStringSubmatch(phaseDirName)
	if m == nil {
		// Cannot determine phase number — fall back to full re-parse.
		c.data = ParseProject(c.root)
		return
	}
	phaseNum := mustAtoi(m[1])

	// Get current active phase/plan from STATE.md so IsActive is set correctly.
	activePhase, activePlan := c.activePosition()

	plans := parsePlansInDir(phaseDir, phaseNum, activePhase, activePlan)

	// Update the matching phase in the cached data.
	for i, ph := range c.data.Phases {
		if ph.DirName == phaseDirName {
			c.data.Phases[i].Plans = plans
			c.data.Phases[i].Status = derivePhaseStatus(plans)
			return
		}
	}
	// Phase not found in cache — full re-parse as fallback.
	c.data = ParseProject(c.root)
}

// updatePhaseBadges re-detects badges for the phase that owns the changed file.
func (c *ProjectCache) updatePhaseBadges(changedPath string) {
	phaseDir := filepath.Dir(changedPath)
	phaseDirName := filepath.Base(phaseDir)

	m := phaseDirRe.FindStringSubmatch(phaseDirName)
	if m == nil {
		c.data = ParseProject(c.root)
		return
	}
	prefix := m[1] // e.g. "01"

	// Re-detect all badges for this phase.
	var badges []string
	for _, bf := range badgeFiles {
		badgePath := filepath.Join(phaseDir, prefix+"-"+bf.suffix)
		if _, statErr := os.Stat(badgePath); statErr == nil {
			badges = append(badges, bf.badge)
		}
	}

	for i, ph := range c.data.Phases {
		if ph.DirName == phaseDirName {
			c.data.Phases[i].Badges = badges
			return
		}
	}
	// Phase not in cache — full re-parse.
	c.data = ParseProject(c.root)
}

// activePosition reads STATE.md to get the current active phase and plan numbers.
// Returns (0, 0) on any error.
func (c *ProjectCache) activePosition() (int, int) {
	st, err := parseState(filepath.Join(c.root, "STATE.md"))
	if err != nil {
		return 0, 0
	}
	return st.ActivePhase, st.ActivePlan
}

// isBadgeFile reports whether the base name looks like a phase badge file
// (e.g. "01-CONTEXT.md") but NOT a PLAN.md file.
func isBadgeFile(base string) bool {
	for _, bf := range badgeFiles {
		if strings.HasSuffix(base, "-"+bf.suffix) {
			return true
		}
	}
	return false
}

// mustAtoi converts a string to int; returns 0 on error.
func mustAtoi(s string) int {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
