package parser

import (
	"os"
	"regexp"
	"strconv"
)

var phaseHeadingRe = regexp.MustCompile(`(?m)#{2,4} Phase (\d+): (.+)`)

// parseRoadmap reads ROADMAP.md and extracts phase number → name mapping.
// Returns empty map on any error (best-effort).
func parseRoadmap(path string) map[int]string {
	content, err := os.ReadFile(path)
	if err != nil {
		return map[int]string{}
	}
	result := map[int]string{}
	matches := phaseHeadingRe.FindAllStringSubmatch(string(content), -1)
	for _, m := range matches {
		if num, err := strconv.Atoi(m[1]); err == nil {
			result[num] = m[2]
		}
	}
	return result
}
