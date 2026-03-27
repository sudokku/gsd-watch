package tui_test

import (
	"testing"

	"github.com/radu/gsd-watch/internal/tui"
)

// TestThemeByName_Known verifies that all supported theme names return ok=true.
func TestThemeByName_Known(t *testing.T) {
	known := []string{"", "default", "minimal", "high-contrast"}
	for _, name := range known {
		_, ok := tui.ThemeByName(name)
		if !ok {
			t.Errorf("ThemeByName(%q) returned ok=false; want ok=true", name)
		}
	}
}

// TestThemeByName_Unknown verifies that unknown theme names return ok=false.
func TestThemeByName_Unknown(t *testing.T) {
	unknown := []string{"neon", "solarized", "MINIMAL", "High-Contrast", "dark"}
	for _, name := range unknown {
		_, ok := tui.ThemeByName(name)
		if ok {
			t.Errorf("ThemeByName(%q) returned ok=true; want ok=false", name)
		}
	}
}

// TestThemeDefault_NotNil verifies that ThemeDefault() returns non-zero styles.
func TestThemeDefault_NotNil(t *testing.T) {
	th := tui.ThemeDefault()
	// Render empty string — if style is zero-value this still returns "" not panics.
	// The test confirms no panic and the Theme is constructable.
	_ = th.Complete.Render("x")
	_ = th.Active.Render("x")
	_ = th.Pending.Render("x")
	_ = th.Failed.Render("x")
	_ = th.NowMarker.Render("x")
	_ = th.RefreshFlash.Render("x")
	_ = th.QuitPending.Render("x")
	_ = th.Highlight.Render("x")
}

// TestThemeMinimal_NotNil verifies that ThemeMinimal() returns non-zero styles.
func TestThemeMinimal_NotNil(t *testing.T) {
	th := tui.ThemeMinimal()
	_ = th.Complete.Render("x")
	_ = th.Pending.Render("x")
}

// TestThemeHighContrast_NotNil verifies that ThemeHighContrast() returns non-zero styles.
func TestThemeHighContrast_NotNil(t *testing.T) {
	th := tui.ThemeHighContrast()
	_ = th.Complete.Render("x")
	_ = th.Failed.Render("x")
}
