package tui

import "github.com/charmbracelet/lipgloss"

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
var (
	CompleteStyle     = lipgloss.NewStyle().Foreground(ColorGreen)
	ActiveStyle       = lipgloss.NewStyle().Foreground(ColorGreen)
	PendingStyle      = lipgloss.NewStyle().Foreground(ColorGray)
	FailedStyle       = lipgloss.NewStyle().Foreground(ColorRed)
	NowMarkerStyle    = lipgloss.NewStyle().Foreground(ColorAmber)
	RefreshFlashStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen)
	QuitPendingStyle  = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber)
)

// StatusIcon returns a styled status icon string for the given status value.
// When noEmoji is true, ASCII bracket equivalents are returned instead of emoji.
func StatusIcon(status string, noEmoji bool) string {
	if noEmoji {
		switch status {
		case "complete":
			return CompleteStyle.Render("[x]")
		case "in_progress":
			return "[>]"
		case "failed":
			return FailedStyle.Render("[!]")
		default:
			return PendingStyle.Render("[ ]")
		}
	}
	switch status {
	case "complete":
		return CompleteStyle.Render("✓")
	case "in_progress":
		return "▶"
	case "failed":
		return FailedStyle.Render("✗")
	default:
		return PendingStyle.Render("○")
	}
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
