package tui

import "github.com/radu/gsd-watch/internal/parser"

// ParsedMsg is sent when the parser successfully parses the project data.
// Phase 1 uses mock data loaded synchronously; Phase 2 will dispatch this
// message from the parser goroutine.
type ParsedMsg struct {
	Project parser.ProjectData
}

// ParseErrorMsg is sent when the parser encounters an error.
// Phase 2 will use this for error handling when parsing .planning/ files.
type ParseErrorMsg struct {
	Err error
}

// FileChangedMsg is sent by the fsnotify file watcher when a file changes.
// Phase 3 will inject this into the event loop via p.Send().
type FileChangedMsg struct {
	Path string
}

// RefreshMsg is sent by the Unix socket listener to trigger a full re-parse.
// Phase 3 will use this for the Stop hook signal from Claude Code.
type RefreshMsg struct{}

// RefreshFlashMsg is sent by a tea.Tick to clear the activity indicator.
// Gen is matched against the model's flashGen so stale ticks from earlier
// file-change bursts are ignored when rapid changes overlap.
type RefreshFlashMsg struct{ Gen int }

// ClockTickMsg is sent every second to force the footer timestamp to re-render,
// keeping the "Ns ago" counter visually up-to-date without requiring a file change.
type ClockTickMsg struct{}

// SpinTickMsg is sent at ~80ms intervals to advance the braille spinner frame.
// The spin loop self-terminates when activeChanges is false.
type SpinTickMsg struct{}

// QuitTimeoutMsg is sent by a tea.Tick after the quit-confirm window expires.
// If quitPending is still true when this arrives, it is cleared (user did not
// confirm in time and the pending state is reset).
type QuitTimeoutMsg struct{}
