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
func StatusIcon(status string) string {
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

// BadgeString returns the emoji for a given phase lifecycle badge.
func BadgeString(badge string) string {
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
