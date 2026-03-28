package tui_test

import (
	"strings"
	"testing"

	"github.com/radu/gsd-watch/internal/tui"
)

// TestStatusIcon_Emoji: all 4 statuses with noEmoji=false produce non-empty strings.
func TestStatusIcon_Emoji(t *testing.T) {
	theme := tui.ThemeDefault()
	statuses := []string{"complete", "in_progress", "pending", "failed"}
	for _, status := range statuses {
		got := tui.StatusIcon(status, false, theme)
		if got == "" {
			t.Errorf("StatusIcon(%q, false) returned empty string", status)
		}
	}
}

// TestStatusIcon_NoEmoji: all 4 statuses with noEmoji=true contain the right ASCII brackets.
func TestStatusIcon_NoEmoji(t *testing.T) {
	theme := tui.ThemeDefault()
	tests := []struct {
		status string
		want   string
	}{
		{"complete", "[x]"},
		{"in_progress", "[>]"},
		{"pending", "[ ]"},
		{"failed", "[!]"},
	}
	for _, tt := range tests {
		got := tui.StatusIcon(tt.status, true, theme)
		if !strings.Contains(got, tt.want) {
			t.Errorf("StatusIcon(%q, true) = %q; want it to contain %q", tt.status, got, tt.want)
		}
	}
}

// TestStatusIcon_NoEmoji_Default: unknown status falls back to pending "[ ]".
func TestStatusIcon_NoEmoji_Default(t *testing.T) {
	got := tui.StatusIcon("unknown_status", true, tui.ThemeDefault())
	if !strings.Contains(got, "[ ]") {
		t.Errorf("StatusIcon(unknown, true) = %q; want it to contain [ ]", got)
	}
}

// TestBadgeString_Emoji: all 7 badges with noEmoji=false produce non-empty strings.
func TestBadgeString_Emoji(t *testing.T) {
	theme := tui.ThemeDefault()
	badges := []string{"discussed", "researched", "ui_spec", "planned", "executed", "verified", "uat"}
	for _, badge := range badges {
		got := tui.BadgeString(badge, false, theme)
		if got == "" {
			t.Errorf("BadgeString(%q, false, theme) returned empty string", badge)
		}
	}
}

// TestBadgeString_NoEmoji: all 7 badges with noEmoji=true return bracketed short codes
// (may be styled, so we check for Contains rather than exact equality).
func TestBadgeString_NoEmoji(t *testing.T) {
	// Use a zero Theme (no BadgeStyle) to get plain text output for exact matching.
	emptyTheme := tui.Theme{}
	tests := []struct {
		badge string
		want  string
	}{
		{"discussed", "[disc]"},
		{"researched", "[rsrch]"},
		{"ui_spec", "[ui]"},
		{"planned", "[plan]"},
		{"executed", "[exec]"},
		{"verified", "[vrfy]"},
		{"uat", "[uat]"},
	}
	for _, tt := range tests {
		got := tui.BadgeString(tt.badge, true, emptyTheme)
		if got != tt.want {
			t.Errorf("BadgeString(%q, true, emptyTheme) = %q; want %q", tt.badge, got, tt.want)
		}
	}
}

// TestBadgeString_Unknown: unknown badge returns "" for both modes.
func TestBadgeString_Unknown(t *testing.T) {
	theme := tui.ThemeDefault()
	if got := tui.BadgeString("unknown_badge", false, theme); got != "" {
		t.Errorf("BadgeString(unknown, false, theme) = %q; want empty string", got)
	}
	if got := tui.BadgeString("unknown_badge", true, theme); got != "" {
		t.Errorf("BadgeString(unknown, true, theme) = %q; want empty string", got)
	}
}
