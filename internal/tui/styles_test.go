package tui_test

import (
	"strings"
	"testing"

	"github.com/radu/gsd-watch/internal/tui"
)

// TestStatusIcon_Emoji: all 4 statuses with noEmoji=false produce non-empty strings.
func TestStatusIcon_Emoji(t *testing.T) {
	statuses := []string{"complete", "in_progress", "pending", "failed"}
	for _, status := range statuses {
		got := tui.StatusIcon(status, false)
		if got == "" {
			t.Errorf("StatusIcon(%q, false) returned empty string", status)
		}
	}
}

// TestStatusIcon_NoEmoji: all 4 statuses with noEmoji=true contain the right ASCII brackets.
func TestStatusIcon_NoEmoji(t *testing.T) {
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
		got := tui.StatusIcon(tt.status, true)
		if !strings.Contains(got, tt.want) {
			t.Errorf("StatusIcon(%q, true) = %q; want it to contain %q", tt.status, got, tt.want)
		}
	}
}

// TestStatusIcon_NoEmoji_Default: unknown status falls back to pending "[ ]".
func TestStatusIcon_NoEmoji_Default(t *testing.T) {
	got := tui.StatusIcon("unknown_status", true)
	if !strings.Contains(got, "[ ]") {
		t.Errorf("StatusIcon(unknown, true) = %q; want it to contain [ ]", got)
	}
}

// TestBadgeString_Emoji: all 7 badges with noEmoji=false produce non-empty strings.
func TestBadgeString_Emoji(t *testing.T) {
	badges := []string{"discussed", "researched", "ui_spec", "planned", "executed", "verified", "uat"}
	for _, badge := range badges {
		got := tui.BadgeString(badge, false)
		if got == "" {
			t.Errorf("BadgeString(%q, false) returned empty string", badge)
		}
	}
}

// TestBadgeString_NoEmoji: all 7 badges with noEmoji=true return exact bracketed short codes.
func TestBadgeString_NoEmoji(t *testing.T) {
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
		got := tui.BadgeString(tt.badge, true)
		if got != tt.want {
			t.Errorf("BadgeString(%q, true) = %q; want %q", tt.badge, got, tt.want)
		}
	}
}

// TestBadgeString_Unknown: unknown badge returns "" for both modes.
func TestBadgeString_Unknown(t *testing.T) {
	if got := tui.BadgeString("unknown_badge", false); got != "" {
		t.Errorf("BadgeString(unknown, false) = %q; want empty string", got)
	}
	if got := tui.BadgeString("unknown_badge", true); got != "" {
		t.Errorf("BadgeString(unknown, true) = %q; want empty string", got)
	}
}
