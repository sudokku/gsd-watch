package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// WordWrap splits text into lines of at most maxWidth display cells,
// breaking at word boundaries. Always returns at least one element.
func WordWrap(text string, maxWidth int) []string {
	if maxWidth <= 0 || lipgloss.Width(text) <= maxWidth {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var result []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
		} else if candidate := current + " " + word; lipgloss.Width(candidate) <= maxWidth {
			current = candidate
		} else {
			result = append(result, current)
			current = word
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
