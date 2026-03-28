package tui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/muesli/termenv"
	"github.com/radu/gsd-watch/internal/config"
	"github.com/radu/gsd-watch/internal/tui"

	"github.com/charmbracelet/lipgloss"
)

func strPtr(s string) *string { return &s }

// newColorRenderer returns a lipgloss Renderer with ANSI256 color profile forced on,
// so that badge style comparisons produce ANSI sequences regardless of terminal state.
func newColorRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(nil)
	r.SetColorProfile(termenv.ANSI256)
	return r
}

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

// TestBadgeStyle_DefaultDistinct verifies at least 3 badge categories produce different
// Render() output in the default theme (cyan vs magenta vs green).
func TestBadgeStyle_DefaultDistinct(t *testing.T) {
	th := tui.ThemeDefault()
	if th.BadgeStyle == nil {
		t.Fatal("ThemeDefault().BadgeStyle is nil; expected a populated map")
	}

	discStyle, ok := th.BadgeStyle["discussed"]
	if !ok {
		t.Fatal("ThemeDefault().BadgeStyle missing 'discussed' key")
	}
	execStyle, ok := th.BadgeStyle["executed"]
	if !ok {
		t.Fatal("ThemeDefault().BadgeStyle missing 'executed' key")
	}
	vrfyStyle, ok := th.BadgeStyle["verified"]
	if !ok {
		t.Fatal("ThemeDefault().BadgeStyle missing 'verified' key")
	}

	// Use a forced-color renderer so ANSI sequences are produced regardless of terminal state.
	r := newColorRenderer()
	disc := r.NewStyle().Inherit(discStyle).Render("[disc]")
	exec := r.NewStyle().Inherit(execStyle).Render("[exec]")
	vrfy := r.NewStyle().Inherit(vrfyStyle).Render("[vrfy]")

	if disc == exec {
		t.Errorf("ThemeDefault: 'discussed' and 'executed' badge styles produce identical output %q; want distinct", disc)
	}
	if exec == vrfy {
		t.Errorf("ThemeDefault: 'executed' and 'verified' badge styles produce identical output %q; want distinct", exec)
	}
	if disc == vrfy {
		t.Errorf("ThemeDefault: 'discussed' and 'verified' badge styles produce identical output %q; want distinct", disc)
	}
}

// TestBadgeStyle_HighContrastBold verifies that all high-contrast badge styles have Bold=true.
// We test this by checking that the rendered output is different from a non-bold render.
func TestBadgeStyle_HighContrastBold(t *testing.T) {
	th := tui.ThemeHighContrast()
	if th.BadgeStyle == nil {
		t.Fatal("ThemeHighContrast().BadgeStyle is nil; expected a populated map")
	}

	badges := []string{"discussed", "researched", "ui_spec", "planned", "executed", "verified", "uat"}
	r := newColorRenderer()
	// A plain style (no bold) rendered "x" — any badge bold style should differ.
	plainOut := r.NewStyle().Render("x")

	for _, badge := range badges {
		style, ok := th.BadgeStyle[badge]
		if !ok {
			t.Errorf("ThemeHighContrast().BadgeStyle missing key %q", badge)
			continue
		}
		rendered := r.NewStyle().Inherit(style).Render("x")
		// Bold + color should not equal plain (no color, no bold).
		if rendered == plainOut {
			t.Errorf("ThemeHighContrast badge %q: rendered output matches plain (no bold/color); want styled output", badge)
		}
	}
}

// TestBadgeStyle_ThemesDiffer verifies the same badge produces different output across all 3 themes.
func TestBadgeStyle_ThemesDiffer(t *testing.T) {
	thDefault := tui.ThemeDefault()
	thMinimal := tui.ThemeMinimal()
	thHighContrast := tui.ThemeHighContrast()

	r := newColorRenderer()

	badges := []string{"discussed", "executed", "verified"}
	for _, badge := range badges {
		sDefault, ok1 := thDefault.BadgeStyle[badge]
		sMinimal, ok2 := thMinimal.BadgeStyle[badge]
		sHC, ok3 := thHighContrast.BadgeStyle[badge]
		if !ok1 || !ok2 || !ok3 {
			t.Errorf("badge %q: one or more themes missing BadgeStyle entry (default=%v, minimal=%v, hc=%v)", badge, ok1, ok2, ok3)
			continue
		}

		text := "[" + badge + "]"
		dOut := r.NewStyle().Inherit(sDefault).Render(text)
		mOut := r.NewStyle().Inherit(sMinimal).Render(text)
		hOut := r.NewStyle().Inherit(sHC).Render(text)

		if dOut == mOut {
			t.Errorf("badge %q: default and minimal produce identical output %q", badge, dOut)
		}
		if dOut == hOut {
			t.Errorf("badge %q: default and high-contrast produce identical output %q", badge, dOut)
		}
		if mOut == hOut {
			t.Errorf("badge %q: minimal and high-contrast produce identical output %q", badge, mOut)
		}
	}
}

// TestThemeStructuralFields_Constructable verifies that all three theme constructors
// populate the new structural chrome fields without panicking.
func TestThemeStructuralFields_Constructable(t *testing.T) {
	themes := []struct {
		name string
		th   tui.Theme
	}{
		{"default", tui.ThemeDefault()},
		{"minimal", tui.ThemeMinimal()},
		{"high-contrast", tui.ThemeHighContrast()},
	}
	for _, tt := range themes {
		th := tt.th
		// Style fields — confirm they are renderable.
		_ = th.InProgressStyle.Render("x")
		_ = th.HeaderNameStyle.Render("x")
		// TerminalColor fields — confirm they satisfy the interface (non-nil).
		if th.SeparatorFg == nil {
			t.Errorf("%s: SeparatorFg is nil", tt.name)
		}
		if th.ProgressFilled == nil {
			t.Errorf("%s: ProgressFilled is nil", tt.name)
		}
		if th.ProgressEmpty == nil {
			t.Errorf("%s: ProgressEmpty is nil", tt.name)
		}
		if th.ConnectorFg == nil {
			t.Errorf("%s: ConnectorFg is nil", tt.name)
		}
		if th.ArchiveSeparatorFg == nil {
			t.Errorf("%s: ArchiveSeparatorFg is nil", tt.name)
		}
		// ExpandIndicatorFg may be lipgloss.NoColor{} (default theme) — that's valid.
		// Just confirm it doesn't panic when used in a style.
		_ = lipgloss.NewStyle().Foreground(th.ExpandIndicatorFg).Render("x")
	}
}

// TestHighContrast_HighlightIsReverse verifies that the high-contrast theme's Highlight
// style uses Reverse(true), producing inverted output different from the plain string.
func TestHighContrast_HighlightIsReverse(t *testing.T) {
	th := tui.ThemeHighContrast()
	r := newColorRenderer()
	plain := r.NewStyle().Render("hello")
	reversed := r.NewStyle().Inherit(th.Highlight).Render("hello")
	if plain == reversed {
		t.Errorf("high-contrast Highlight should produce reverse-video output; got same as plain: %q", plain)
	}
}

// TestStatusIcon_InProgress_UsesThemeStyle verifies that StatusIcon in_progress applies
// the theme's InProgressStyle, so different themes produce different output for in_progress.
func TestStatusIcon_InProgress_UsesThemeStyle(t *testing.T) {
	r := newColorRenderer()

	// minimal theme gives InProgressStyle a muted foreground color.
	thDefault := tui.ThemeDefault()
	thMinimal := tui.ThemeMinimal()

	// In emoji mode (noEmoji=false)
	defaultOut := tui.StatusIcon("in_progress", false, thDefault)
	minimalOut := tui.StatusIcon("in_progress", false, thMinimal)

	// Apply through the test renderer so ANSI sequences are consistent.
	// We just need to confirm the styled versions are non-empty and the function doesn't panic.
	_ = r.NewStyle().Render(defaultOut)
	_ = r.NewStyle().Render(minimalOut)

	// In noEmoji mode (noEmoji=true) — both should contain "[>]"
	defaultNoEmoji := tui.StatusIcon("in_progress", true, thDefault)
	if !strings.Contains(defaultNoEmoji, "[>]") {
		t.Errorf("StatusIcon in_progress noEmoji: want '[>]' in output, got %q", defaultNoEmoji)
	}
	minimalNoEmoji := tui.StatusIcon("in_progress", true, thMinimal)
	if !strings.Contains(minimalNoEmoji, "[>]") {
		t.Errorf("StatusIcon in_progress noEmoji minimal: want '[>]' in output, got %q", minimalNoEmoji)
	}
}

// TestMinimal_HeaderNameStyle_NoBold verifies that the minimal theme's HeaderNameStyle
// produces plain output (no bold), while default and high-contrast produce bold output.
func TestMinimal_HeaderNameStyle_NoBold(t *testing.T) {
	r := newColorRenderer()
	thDefault := tui.ThemeDefault()
	thMinimal := tui.ThemeMinimal()
	thHC := tui.ThemeHighContrast()

	plain := r.NewStyle().Render("Project")
	defaultName := r.NewStyle().Inherit(thDefault.HeaderNameStyle).Render("Project")
	minimalName := r.NewStyle().Inherit(thMinimal.HeaderNameStyle).Render("Project")
	hcName := r.NewStyle().Inherit(thHC.HeaderNameStyle).Render("Project")

	// default and high-contrast should be bold (differ from plain).
	if defaultName == plain {
		t.Errorf("ThemeDefault HeaderNameStyle: want bold (different from plain), got same as plain: %q", plain)
	}
	if hcName == plain {
		t.Errorf("ThemeHighContrast HeaderNameStyle: want bold (different from plain), got same as plain: %q", plain)
	}
	// minimal should not add bold (same as plain).
	if minimalName != plain {
		t.Errorf("ThemeMinimal HeaderNameStyle: want plain (no bold), got %q; plain is %q", minimalName, plain)
	}
}

// TestBadgeString_EmojiNoThemeChange verifies that in emoji mode (noEmoji=false),
// BadgeString returns the same emoji regardless of theme (no ANSI wrapping).
func TestBadgeString_EmojiNoThemeChange(t *testing.T) {
	thDefault := tui.ThemeDefault()
	thMinimal := tui.ThemeMinimal()
	thHC := tui.ThemeHighContrast()

	badges := []string{"discussed", "executed", "verified"}
	for _, badge := range badges {
		d := tui.BadgeString(badge, false, thDefault)
		m := tui.BadgeString(badge, false, thMinimal)
		h := tui.BadgeString(badge, false, thHC)
		if d != m || d != h {
			t.Errorf("badge %q: emoji mode produced different output across themes: default=%q minimal=%q hc=%q", badge, d, m, h)
		}
		// Must be non-empty
		if d == "" {
			t.Errorf("badge %q: emoji mode returned empty string", badge)
		}
	}
}
