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
