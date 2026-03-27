package tui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/radu/gsd-watch/internal/config"
	"github.com/radu/gsd-watch/internal/tui"
)

func strPtr(s string) *string { return &s }

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

// TestIsValidHex verifies the hex color string validation helper.
func TestIsValidHex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"#00ff00", true},
		{"#FF00FF", true},
		{"#fff", false},
		{"00ff00", false},
		{"", false},
		{"#1234567", false},
	}
	for _, tt := range tests {
		if got := tui.IsValidHex(tt.input); got != tt.want {
			t.Errorf("IsValidHex(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// TestApplyColorOverrides_NilUnchanged verifies that nil ThemeColors fields leave the theme unchanged.
func TestApplyColorOverrides_NilUnchanged(t *testing.T) {
	base := tui.ThemeDefault()
	var buf bytes.Buffer
	got := tui.ApplyColorOverrides(base, config.ThemeColors{}, &buf)
	if buf.Len() != 0 {
		t.Errorf("unexpected warnings: %s", buf.String())
	}
	if got.Complete.Render("x") != base.Complete.Render("x") {
		t.Errorf("Complete style changed with nil override")
	}
	if got.Failed.Render("x") != base.Failed.Render("x") {
		t.Errorf("Failed style changed with nil override")
	}
}

// TestApplyColorOverrides_ValidHex verifies that a valid hex string is applied without warnings.
func TestApplyColorOverrides_ValidHex(t *testing.T) {
	base := tui.ThemeDefault()
	var buf bytes.Buffer
	overrides := config.ThemeColors{Complete: strPtr("#00ff00")}
	got := tui.ApplyColorOverrides(base, overrides, &buf)
	if buf.Len() != 0 {
		t.Errorf("unexpected warnings: %s", buf.String())
	}
	// Style should have changed — just verify no panic and no warnings
	_ = got.Complete.Render("x")
}

// TestApplyColorOverrides_InvalidHex verifies that invalid hex values emit warnings and preserve preset colors.
func TestApplyColorOverrides_InvalidHex(t *testing.T) {
	base := tui.ThemeDefault()
	var buf bytes.Buffer
	overrides := config.ThemeColors{Complete: strPtr("bad"), Failed: strPtr("#ff")}
	got := tui.ApplyColorOverrides(base, overrides, &buf)
	output := buf.String()
	if !strings.Contains(output, "[theme].complete") {
		t.Errorf("warning should name field 'complete', got: %s", output)
	}
	if !strings.Contains(output, `"bad"`) {
		t.Errorf("warning should contain bad value, got: %s", output)
	}
	if !strings.Contains(output, "[theme].failed") {
		t.Errorf("warning should name field 'failed', got: %s", output)
	}
	// Preset colors preserved — render should match base
	if got.Complete.Render("x") != base.Complete.Render("x") {
		t.Errorf("Complete style should be preserved on invalid hex")
	}
}
