package parser

import (
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

// stateData holds extracted STATE.md values.
type stateData struct {
	MilestoneName   string
	StoppedAt       string
	ProgressPercent int
	ActivePhase     int // parsed from prose "Phase: N of M"
	ActivePlan      int // parsed from prose "Plan: N of M"
}

// stateFrontmatter matches STATE.md YAML frontmatter structure.
type stateFrontmatter struct {
	MilestoneName string `yaml:"milestone_name"`
	StoppedAt     string `yaml:"stopped_at"`
	Progress      struct {
		Percent int `yaml:"percent"`
	} `yaml:"progress"`
}

var (
	phaseLineRe = regexp.MustCompile(`(?m)^Phase:\s+(\d+)`)
	planLineRe  = regexp.MustCompile(`(?m)^Plan:\s+(\d+)`)
)

// parseState reads STATE.md and extracts frontmatter fields + active plan from prose.
// Returns non-nil error only on file read failure. YAML errors are silently ignored
// (zero-value fields). ActivePhase and ActivePlan are only parsed when frontmatter
// is present (prose after frontmatter delimiter).
func parseState(path string) (stateData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return stateData{}, err
	}

	var result stateData

	// Split into frontmatter + prose using extractFrontmatter from plan.go.
	fm, prose := extractFrontmatter(string(content))

	// Parse YAML frontmatter fields.
	if fm != "" {
		var sf stateFrontmatter
		if yamlErr := yaml.Unmarshal([]byte(fm), &sf); yamlErr == nil {
			result.MilestoneName = sf.MilestoneName
			result.StoppedAt = sf.StoppedAt
			result.ProgressPercent = sf.Progress.Percent
		}
		// On YAML error: leave defaults, don't propagate.

		// Only parse active phase/plan from prose when frontmatter was present.
		if m := phaseLineRe.FindStringSubmatch(prose); len(m) > 1 {
			result.ActivePhase, _ = strconv.Atoi(m[1])
		}
		if m := planLineRe.FindStringSubmatch(prose); len(m) > 1 {
			result.ActivePlan, _ = strconv.Atoi(m[1])
		}
	}

	return result, nil
}
