package parser

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// planFrontmatter holds only the fields we need from PLAN.md YAML.
// yaml.v3 silently ignores unknown fields by default — do NOT use KnownFields(true).
type planFrontmatter struct {
	Status string `yaml:"status"`
	Wave   int    `yaml:"wave"`
}

var objectiveRe = regexp.MustCompile(`(?s)<objective>\s*\n(.+?)[\n]`)

// parsePlan reads a PLAN.md file and extracts frontmatter fields + title from objective block.
// Returns a partial Plan and error. Caller decides fallback behavior.
func parsePlan(path string) (Plan, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Plan{}, err
	}
	fm, _ := extractFrontmatter(string(content))
	var pf planFrontmatter
	if fm != "" {
		if err := yaml.Unmarshal([]byte(fm), &pf); err != nil {
			return Plan{}, err
		}
	}
	title := ""
	if m := objectiveRe.FindStringSubmatch(string(content)); len(m) > 1 {
		title = strings.TrimSpace(m[1])
	}
	return Plan{
		Status: pf.Status,
		Wave:   pf.Wave,
		Title:  title,
	}, nil
}

// extractFrontmatter splits a file starting with "---\n...\n---" into YAML block and prose body.
func extractFrontmatter(content string) (string, string) {
	if !strings.HasPrefix(content, "---") {
		return "", content
	}
	rest := content[3:]
	// Find the closing "---" — must be on its own line
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return "", content
	}
	return strings.TrimSpace(rest[:idx]), rest[idx+4:]
}
