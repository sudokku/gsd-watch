package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/config"
)

// MinWidth is the minimum terminal width before showing a "too narrow" placeholder.
const MinWidth = 30

// Adaptive color palette — works on both dark and light terminals.
var (
	ColorGreen = lipgloss.AdaptiveColor{Light: "2", Dark: "2"}
	ColorAmber = lipgloss.AdaptiveColor{Light: "3", Dark: "3"}
	ColorRed   = lipgloss.AdaptiveColor{Light: "1", Dark: "1"}
	ColorGray  = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}
)

// Shared styles for status rendering.
// These package-level vars are used by footer/model.go, header/model.go, and tests.
// tree/view.go uses Theme fields instead (set via tree.Options.Theme).
var (
	CompleteStyle     = lipgloss.NewStyle().Foreground(ColorGreen)
	ActiveStyle       = lipgloss.NewStyle().Foreground(ColorGreen)
	PendingStyle      = lipgloss.NewStyle().Foreground(ColorGray)
	FailedStyle       = lipgloss.NewStyle().Foreground(ColorRed)
	NowMarkerStyle    = lipgloss.NewStyle().Foreground(ColorAmber)
	RefreshFlashStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen)
	QuitPendingStyle  = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber)
)

// Theme bundles all lipgloss styles needed for tree and archive rendering.
// Each Theme field replaces a direct reference to the package-level style vars in view.go.
type Theme struct {
	Complete     lipgloss.Style
	Active       lipgloss.Style
	Pending      lipgloss.Style
	Failed       lipgloss.Style
	NowMarker    lipgloss.Style
	RefreshFlash lipgloss.Style
	QuitPending  lipgloss.Style
	Highlight    lipgloss.Style
	EmptyFg      lipgloss.TerminalColor
	HelpBorder   lipgloss.TerminalColor
	HelpFg       lipgloss.TerminalColor
}

// ThemeDefault returns the default theme — identical to pre-Phase-14 global style vars.
// THEME-01: no visual regression from gsd-watch v1.2.
func ThemeDefault() Theme {
	return Theme{
		Complete:     lipgloss.NewStyle().Foreground(ColorGreen),
		Active:       lipgloss.NewStyle().Foreground(ColorGreen),
		Pending:      lipgloss.NewStyle().Foreground(ColorGray),
		Failed:       lipgloss.NewStyle().Foreground(ColorRed),
		NowMarker:    lipgloss.NewStyle().Foreground(ColorAmber),
		RefreshFlash: lipgloss.NewStyle().Bold(true).Foreground(ColorGreen),
		QuitPending:  lipgloss.NewStyle().Bold(true).Foreground(ColorAmber),
		Highlight:    lipgloss.NewStyle().Bold(true),
		EmptyFg:      ColorGray,
		HelpBorder:   ColorGray,
		HelpFg:       ColorGray,
	}
}

// ThemeMinimal returns a muted, content-first theme with subdued status colors.
// THEME-02: muted status colors and content-first appearance throughout the tree.
func ThemeMinimal() Theme {
	muted := lipgloss.AdaptiveColor{Light: "243", Dark: "243"}
	dim := lipgloss.AdaptiveColor{Light: "245", Dark: "245"}
	return Theme{
		Complete:     lipgloss.NewStyle().Foreground(muted),
		Active:       lipgloss.NewStyle().Foreground(dim),
		Pending:      lipgloss.NewStyle().Foreground(muted),
		Failed:       lipgloss.NewStyle().Foreground(muted),
		NowMarker:    lipgloss.NewStyle().Foreground(dim),
		RefreshFlash: lipgloss.NewStyle().Foreground(dim),
		QuitPending:  lipgloss.NewStyle().Foreground(muted),
		Highlight:    lipgloss.NewStyle().Bold(true),
		EmptyFg:      muted,
		HelpBorder:   muted,
		HelpFg:       muted,
	}
}

// ThemeHighContrast returns a theme using only 16-color ANSI palette indices.
// THEME-03: bold foreground colors visible over SSH and in degraded terminals.
func ThemeHighContrast() Theme {
	green := lipgloss.Color("2")
	yellow := lipgloss.Color("3")
	red := lipgloss.Color("1")
	white := lipgloss.Color("7")
	return Theme{
		Complete:     lipgloss.NewStyle().Bold(true).Foreground(green),
		Active:       lipgloss.NewStyle().Bold(true).Foreground(green),
		Pending:      lipgloss.NewStyle().Foreground(white),
		Failed:       lipgloss.NewStyle().Bold(true).Foreground(red),
		NowMarker:    lipgloss.NewStyle().Bold(true).Foreground(yellow),
		RefreshFlash: lipgloss.NewStyle().Bold(true).Foreground(green),
		QuitPending:  lipgloss.NewStyle().Bold(true).Foreground(yellow),
		Highlight:    lipgloss.NewStyle().Bold(true),
		EmptyFg:      white,
		HelpBorder:   white,
		HelpFg:       white,
	}
}

// ThemeByName returns the Theme for the given name and true, or ThemeDefault() and false
// for an unrecognised name. Empty string resolves to the default theme with ok=true.
// THEME-04: unknown theme name falls back to default without crash.
func ThemeByName(name string) (Theme, bool) {
	switch name {
	case "", "default":
		return ThemeDefault(), true
	case "minimal":
		return ThemeMinimal(), true
	case "high-contrast":
		return ThemeHighContrast(), true
	default:
		return ThemeDefault(), false
	}
}

// StatusIcon returns a styled status icon string for the given status value.
// When noEmoji is true, ASCII bracket equivalents are returned instead of emoji.
// The theme parameter controls the colors applied to the icon.
func StatusIcon(status string, noEmoji bool, theme Theme) string {
	if noEmoji {
		switch status {
		case "complete":
			return theme.Complete.Render("[x]")
		case "in_progress":
			return "[>]"
		case "failed":
			return theme.Failed.Render("[!]")
		default:
			return theme.Pending.Render("[ ]")
		}
	}
	switch status {
	case "complete":
		return theme.Complete.Render("✓")
	case "in_progress":
		return "▶"
	case "failed":
		return theme.Failed.Render("✗")
	default:
		return theme.Pending.Render("○")
	}
}

// IsValidHex returns true if s is a valid #RRGGBB hex color string.
// Only checks length (7) and # prefix per D-04. Does not validate hex digits.
func IsValidHex(s string) bool {
	return len(s) == 7 && s[0] == '#'
}

// ApplyColorOverrides returns a copy of theme with each non-nil ThemeColors
// field applied as a hex foreground color. Invalid hex values emit a warning
// to w and preserve the preset color. Per D-05: never fatal for color errors.
func ApplyColorOverrides(theme Theme, overrides config.ThemeColors, w io.Writer) Theme {
	apply := func(style *lipgloss.Style, field string, val *string) {
		if val == nil {
			return
		}
		if IsValidHex(*val) {
			*style = lipgloss.NewStyle().Foreground(lipgloss.Color(*val))
		} else {
			fmt.Fprintf(w, "gsd-watch: invalid color %q for [theme].%s (ignored)\n", *val, field)
		}
	}
	apply(&theme.Complete, "complete", overrides.Complete)
	apply(&theme.Active, "active", overrides.Active)
	apply(&theme.Pending, "pending", overrides.Pending)
	apply(&theme.Failed, "failed", overrides.Failed)
	apply(&theme.NowMarker, "now_marker", overrides.NowMarker)
	return theme
}

// BadgeString returns the emoji (or ASCII short code) for a given phase lifecycle badge.
// When noEmoji is true, bracketed short codes are returned instead of emoji.
func BadgeString(badge string, noEmoji bool) string {
	if noEmoji {
		switch badge {
		case "discussed":
			return "[disc]"
		case "researched":
			return "[rsrch]"
		case "ui_spec":
			return "[ui]"
		case "planned":
			return "[plan]"
		case "executed":
			return "[exec]"
		case "verified":
			return "[vrfy]"
		case "uat":
			return "[uat]"
		default:
			return ""
		}
	}
	switch badge {
	case "discussed":
		return "💬"
	case "researched":
		return "🔎"
	case "ui_spec":
		return "🎨"
	case "planned":
		return "📋"
	case "executed":
		return "🚀"
	case "verified":
		return "✅"
	case "uat":
		return "🧪"
	default:
		return ""
	}
}
