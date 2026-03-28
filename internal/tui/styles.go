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

	// Structural chrome colors — separator lines, progress bar, tree connectors.
	SeparatorFg       lipgloss.TerminalColor // header ═ and footer ─ separators
	ProgressFilled    lipgloss.TerminalColor // progress bar filled blocks ▓
	ProgressEmpty     lipgloss.TerminalColor // progress bar empty blocks ░
	ConnectorFg       lipgloss.TerminalColor // tree ├──, └──, │ connectors
	ExpandIndicatorFg lipgloss.TerminalColor // ▶ / ▼ expand arrows
	ArchiveSeparatorFg lipgloss.TerminalColor // "- - Archived Milestones - -" line

	// Structural styles — applied to composite rendered elements.
	InProgressStyle lipgloss.Style // ▶ / [>] in-progress icon
	HeaderNameStyle lipgloss.Style // project name in header

	// BadgeStyle maps badge name (e.g. "discussed", "executed") to the lipgloss style
	// used when rendering bracketed badge codes in noEmoji mode. Each theme preset
	// defines its own palette to make themes visually distinct.
	BadgeStyle map[string]lipgloss.Style
}

// ThemeDefault returns the default theme — identical to pre-Phase-14 global style vars.
// THEME-01: no visual regression from gsd-watch v1.2.
// Badge palette uses 256-color ANSI codes for distinct category coloring:
//   - discussed/researched: Cyan (36) — discovery/information gathering
//   - ui_spec: Blue (33) — design
//   - planned: Blue (69) — planning
//   - executed: Magenta (133) — action/shipping
//   - verified: Green (34) — confirmation
//   - uat: Yellow (178) — testing/caution
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

		SeparatorFg:        ColorGray,
		ProgressFilled:     ColorGreen,
		ProgressEmpty:      ColorGray,
		ConnectorFg:        ColorGray,
		ExpandIndicatorFg:  lipgloss.NoColor{},
		ArchiveSeparatorFg: ColorGray,
		InProgressStyle:    lipgloss.NewStyle(),
		HeaderNameStyle:    lipgloss.NewStyle().Bold(true),

		BadgeStyle: map[string]lipgloss.Style{
			"discussed":  lipgloss.NewStyle().Foreground(lipgloss.Color("36")),  // Cyan
			"researched": lipgloss.NewStyle().Foreground(lipgloss.Color("36")),  // Cyan
			"ui_spec":    lipgloss.NewStyle().Foreground(lipgloss.Color("33")),  // Blue
			"planned":    lipgloss.NewStyle().Foreground(lipgloss.Color("69")),  // Blue (lighter)
			"executed":   lipgloss.NewStyle().Foreground(lipgloss.Color("133")), // Magenta
			"verified":   lipgloss.NewStyle().Foreground(lipgloss.Color("34")),  // Green
			"uat":        lipgloss.NewStyle().Foreground(lipgloss.Color("178")), // Yellow
		},
	}
}

// ThemeMinimal returns a muted, content-first theme with subdued status colors.
// THEME-02: muted status colors and content-first appearance throughout the tree.
// Badge palette: all badges use a single muted tone (243), no bold — consistent with content-first aesthetic.
// Active uses color 248 (slightly brighter than 243 Complete) so active phases stand out slightly.
func ThemeMinimal() Theme {
	muted := lipgloss.AdaptiveColor{Light: "243", Dark: "243"}
	dim := lipgloss.AdaptiveColor{Light: "248", Dark: "248"}
	mutedBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
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

		SeparatorFg:        lipgloss.Color("240"),
		ProgressFilled:     lipgloss.Color("243"),
		ProgressEmpty:      lipgloss.Color("238"),
		ConnectorFg:        lipgloss.Color("240"),
		ExpandIndicatorFg:  lipgloss.Color("243"),
		ArchiveSeparatorFg: lipgloss.Color("238"),
		InProgressStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		HeaderNameStyle:    lipgloss.NewStyle(),

		BadgeStyle: map[string]lipgloss.Style{
			"discussed":  mutedBadge,
			"researched": mutedBadge,
			"ui_spec":    mutedBadge,
			"planned":    mutedBadge,
			"executed":   mutedBadge,
			"verified":   mutedBadge,
			"uat":        mutedBadge,
		},
	}
}

// ThemeHighContrast returns a theme using only 16-color ANSI palette indices.
// THEME-03: bold foreground colors visible over SSH and in degraded terminals.
// Status: Active gets Bold+Underline for extra visibility; Pending uses Bright White (15).
// Badge palette: Bold + bright 16-color ANSI only for maximum SSH/degraded terminal visibility:
//   - discussed/researched: Bright Cyan (14)
//   - ui_spec/planned: Bright Blue (12)
//   - executed: Bright Magenta (13)
//   - verified: Bright Green (10)
//   - uat: Bright Yellow (11)
func ThemeHighContrast() Theme {
	green := lipgloss.Color("2")
	yellow := lipgloss.Color("3")
	red := lipgloss.Color("1")
	white := lipgloss.Color("7")
	brightWhite := lipgloss.Color("15")
	return Theme{
		Complete:     lipgloss.NewStyle().Bold(true).Foreground(green),
		Active:       lipgloss.NewStyle().Bold(true).Underline(true).Foreground(green),
		Pending:      lipgloss.NewStyle().Foreground(brightWhite),
		Failed:       lipgloss.NewStyle().Bold(true).Foreground(red),
		NowMarker:    lipgloss.NewStyle().Bold(true).Foreground(yellow),
		RefreshFlash: lipgloss.NewStyle().Bold(true).Foreground(green),
		QuitPending:  lipgloss.NewStyle().Bold(true).Foreground(yellow),
		Highlight:    lipgloss.NewStyle().Reverse(true),
		EmptyFg:      brightWhite,
		HelpBorder:   brightWhite,
		HelpFg:       brightWhite,

		SeparatorFg:        white,
		ProgressFilled:     green,
		ProgressEmpty:      lipgloss.Color("0"),
		ConnectorFg:        white,
		ExpandIndicatorFg:  yellow,
		ArchiveSeparatorFg: white,
		InProgressStyle:    lipgloss.NewStyle().Bold(true).Foreground(yellow),
		HeaderNameStyle:    lipgloss.NewStyle().Bold(true),

		BadgeStyle: map[string]lipgloss.Style{
			"discussed":  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")), // Bright Cyan
			"researched": lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")), // Bright Cyan
			"ui_spec":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")), // Bright Blue
			"planned":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")), // Bright Blue
			"executed":   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13")), // Bright Magenta
			"verified":   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")), // Bright Green
			"uat":        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")), // Bright Yellow
		},
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
			return theme.InProgressStyle.Render("[>]")
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
		return theme.InProgressStyle.Render("▶")
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

// BadgeString returns the emoji (or styled ASCII short code) for a given phase lifecycle badge.
// When noEmoji is false, emoji characters are returned unstyled (theme is ignored).
// When noEmoji is true, bracketed short codes are returned styled with theme.BadgeStyle[badge]
// if the badge key exists in the map; otherwise plain text is returned.
func BadgeString(badge string, noEmoji bool, theme Theme) string {
	if noEmoji {
		var plain string
		switch badge {
		case "discussed":
			plain = "[disc]"
		case "researched":
			plain = "[rsrch]"
		case "ui_spec":
			plain = "[ui]"
		case "planned":
			plain = "[plan]"
		case "executed":
			plain = "[exec]"
		case "verified":
			plain = "[vrfy]"
		case "uat":
			plain = "[uat]"
		default:
			return ""
		}
		if style, ok := theme.BadgeStyle[badge]; ok {
			return style.Render(plain)
		}
		return plain
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
