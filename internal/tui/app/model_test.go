package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/radu/gsd-watch/internal/config"
)

func strPtr(s string) *string { return &s }

// TestNew_WithColorOverrides verifies that New() applies color overrides from
// Config.Colors without panicking. This is the integration test for the
// ApplyColorOverrides wiring added in model.go.
func TestNew_WithColorOverrides(t *testing.T) {
	events := make(chan tea.Msg, 1)
	cfg := config.Config{
		Emoji:  true,
		Preset: "",
		Colors: config.ThemeColors{
			Complete: strPtr("#ff0000"),
		},
	}
	// Must not panic — exercises ThemeByName + ApplyColorOverrides path
	_ = New(events, cfg)
}
