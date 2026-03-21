---
phase: 06-onboarding-docs-ux
plan: 01
subsystem: tui, binary
tags: [ux, discoverability, onboarding, footer, help-overlay, word-wrap, cli-flags]
dependency_graph:
  requires: []
  provides: [footer-help-hint, help-overlay-phase-stages, phase-name-wrapping, binary-help-flag, tmux-detection]
  affects: [internal/tui/footer/model.go, internal/tui/app/model.go, internal/tui/tree/view.go, cmd/gsd-watch/main.go]
tech_stack:
  added: []
  patterns: [word-wrap-phase-names, flag-stdlib, tmux-env-check]
key_files:
  created: []
  modified:
    - internal/tui/footer/model.go
    - internal/tui/footer/model_test.go
    - internal/tui/app/model.go
    - internal/tui/model_test.go
    - internal/tui/tree/view.go
    - internal/tui/tree/model_test.go
    - cmd/gsd-watch/main.go
decisions:
  - "[06-01] Footer hint uses static string '? help' appended to existing collapse/expand hints with same separator"
  - "[06-01] Help overlay Phase stages uses actual BadgeString() emojis: 💬 🔎 📋 ✅ 🧪"
  - "[06-01] Phase name wrapping applies highlight/dim per-line independently to avoid ANSI reset kill"
  - "[06-01] renderedRowLines RowPhase uses fixed expandIndicator width (2 cells) for prefix calculation"
  - "[06-01] --help uses flag stdlib, no third-party CLI library per project constraints"
  - "[06-01] TMUX check uses os.Getenv('TMUX') — empty string means not in tmux"
metrics:
  duration: 2 min
  completed_date: "2026-03-21"
  tasks_completed: 2
  files_modified: 7
---

# Phase 6 Plan 1: Footer hint, help overlay phase stages, phase name wrapping, --help flag, tmux detection Summary

**One-liner:** Footer shows "? help" hint, help overlay has Phase stages badge legend, phase names word-wrap, binary self-documents with --help and blocks outside tmux.

## Tasks Completed

| # | Task | Commit | Key Files |
|---|------|--------|-----------|
| 1 | Footer hint, help overlay phase stages, phase name wrapping | 29ed38c | footer/model.go, app/model.go, tree/view.go + tests |
| 2 | Binary --help flag and outside-tmux detection | b999b3a | cmd/gsd-watch/main.go |

## What Was Built

### Task 1: TUI Discoverability Improvements

**Footer hint update:** Added `· ? help` to the footer's action hints line. The string `w collapse · e expand · ? help` now appears as the second hint line.

**Help overlay Phase stages section:** Added a "Phase stages" block after the Quit section listing each badge emoji with its label: 💬 discussed, 🔎 researched, 📋 planned, ✅ verified, 🧪 UAT. These match the actual `BadgeString()` values rendered in the tree.

**Phase name word-wrapping:** Replaced single-line phase header rendering with multi-line wrapped rendering. The implementation:
1. Computes `prefixStr = expandIndicator + icon + " "` and its display width
2. Calculates `wrapWidth = width - 1 - prefixWidth` (matching the D-10 left-padding offset)
3. Calls `tui.WordWrap(row.Phase.Name, wrapWidth)` to split the name
4. First line: `prefixStr + nameParts[0]`; continuation lines: `strings.Repeat(" ", prefixWidth) + nameParts[j]`
5. Applies highlight (cursor) and dimming (completed) per-line to avoid ANSI reset kill
6. Updated `renderedRowLines` RowPhase case to call `tui.WordWrap` instead of hardcoded `n := 1`

### Task 2: Binary Self-Documentation

**--help flag:** Uses `flag.Bool("help", ...)` from stdlib. Prints a one-liner description, slash command reference, keybindings table, and GitHub URL then exits 0.

**Outside-tmux detection:** Checks `os.Getenv("TMUX") == ""` after the help check. Prints install instructions to stderr and exits 1 if not in a tmux session.

**Execution order in main():**
1. flag.Parse + --help check
2. TMUX env var check
3. OSC 2 pane title (existing)
4. tea.NewProgram (existing)

## Verification Results

All tests pass:
- `go test ./internal/tui/... -count=1` — all packages OK
- `/tmp/gsd-watch-test --help` — prints usage, exits 0
- `TMUX="" /tmp/gsd-watch-test` — prints tmux requirement message, exits 1

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None — all implemented functionality is wired to real data sources.

## Self-Check: PASSED

Files verified to exist:
- internal/tui/footer/model.go — contains "w collapse · e expand · ? help"
- internal/tui/app/model.go — contains "Phase stages"
- internal/tui/tree/view.go — calls tui.WordWrap for phase names
- cmd/gsd-watch/main.go — contains flag.Bool("help") and os.Getenv("TMUX")

Commits verified:
- 29ed38c — feat(06-01): footer help hint, help overlay phase stages, phase name wrapping
- b999b3a — feat(06-01): --help flag and outside-tmux detection in main
