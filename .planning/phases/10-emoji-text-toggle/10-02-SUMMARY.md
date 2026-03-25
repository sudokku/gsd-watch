---
phase: 10-emoji-text-toggle
plan: "02"
subsystem: tui
tags: [a11y, flag, no-emoji, ascii, integration-test]
dependency_graph:
  requires: ["10-01"]
  provides: ["A11Y-01"]
  affects: ["cmd/gsd-watch/main.go", "internal/tui/app/model.go", "internal/tui/model_test.go"]
tech_stack:
  added: []
  patterns: ["flag.Bool for CLI flag", "TDD integration tests", "noEmoji propagation via function param"]
key_files:
  created: []
  modified:
    - cmd/gsd-watch/main.go
    - internal/tui/app/model.go
    - internal/tui/model_test.go
decisions:
  - "[10-02] helpView accepts noEmoji bool param — keeps View() clean and avoids storing render-only state in struct (struct already has noEmoji field but helpView takes it explicitly for purity)"
  - "[10-02] newTestModel() passes false; newTestModelNoEmoji() is a separate helper — avoids changing existing test signatures and makes intent explicit"
metrics:
  duration: "10 min"
  completed: "2026-03-25"
  tasks_completed: 2
  files_modified: 3
---

# Phase 10 Plan 02: Wire --no-emoji CLI Flag End-to-End Summary

**One-liner:** --no-emoji CLI flag parsed in main.go, propagated through app.New(noEmoji bool) to tree.SetOptions, help overlay shows ASCII bracket codes when active.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Wire --no-emoji flag from main.go through app to tree | afbaeea | cmd/gsd-watch/main.go, internal/tui/app/model.go, internal/tui/model_test.go |
| 2 | Integration tests for noEmoji rendering | d242d53 | internal/tui/model_test.go |

## What Was Built

- `--no-emoji` flag added to `cmd/gsd-watch/main.go` with description "Use ASCII status icons and badges (for SSH and minimal terminals)"
- `--help` output updated to list `--no-emoji` in the Flags section
- `app.New(events chan tea.Msg, noEmoji bool)` — signature updated to accept noEmoji
- `app.Model.noEmoji bool` field added; set from New() param
- `t.SetOptions(tree.Options{NoEmoji: noEmoji})` called in New() to wire flag to tree
- `helpView(width int, noEmoji bool)` — Phase stages section conditionally renders ASCII bracket codes (`[disc]`, `[rsrch]`, `[ui]`, `[plan]`, `[exec]`, `[vrfy]`, `[uat]`) vs emoji
- `newTestModel()` updated to `app.New(make(chan tea.Msg, 10), false)`
- `newTestModelNoEmoji()` added: `app.New(make(chan tea.Msg, 10), true)`

## Integration Tests Added

| Test | Assertion |
|------|-----------|
| TestNoEmoji_TreeRenders_ASCIIIcons | View contains "[x]", does NOT contain "✓" |
| TestNoEmoji_TreeRenders_ASCIIBadges | View contains "[disc]"/"[plan]"/etc., does NOT contain "📋" |
| TestNoEmoji_HelpOverlay_ASCIIBadges | Help overlay contains "[disc]" and "[uat]", does NOT contain "💬" |
| TestNoEmoji_False_RendersEmoji | Default model does NOT contain "[x]" |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed model_test.go compilation failure**
- **Found during:** Task 1 verification
- **Issue:** Updating `app.New` signature broke `model_test.go` which still called `app.New(make(chan tea.Msg, 10))` with the old single-arg signature. Build failed.
- **Fix:** Updated `newTestModel()` and added `newTestModelNoEmoji()` as part of Task 1 commit (these changes are also the core requirement of Task 2, so they were included early to unblock verification).
- **Files modified:** internal/tui/model_test.go
- **Commit:** afbaeea

**Note on TDD:** The integration tests (Task 2) passed immediately on the first run (GREEN from the start) because the implementation was already complete from Task 1. The RED phase was skipped since the `TestNoEmoji_*` test bodies depend on behavior already implemented.

## Verification Results

- `go build ./...` exits 0
- `go test ./... -count=1` — all packages pass
- `go run ./cmd/gsd-watch --help` lists `--no-emoji  Use ASCII status icons and badges (for SSH and minimal terminals)`
- `go test ./internal/tui/... -run "TestNoEmoji" -v -count=1` — all 4 tests PASS

## Known Stubs

None.

## Self-Check: PASSED

- [x] cmd/gsd-watch/main.go contains `flag.Bool("no-emoji"`
- [x] cmd/gsd-watch/main.go contains `--no-emoji  Use ASCII status icons and badges`
- [x] cmd/gsd-watch/main.go contains `app.New(events, *noEmoji)`
- [x] internal/tui/app/model.go contains `noEmoji bool` field
- [x] internal/tui/app/model.go contains `func New(events chan tea.Msg, noEmoji bool) Model`
- [x] internal/tui/app/model.go contains `t.SetOptions(tree.Options{NoEmoji: noEmoji})`
- [x] internal/tui/app/model.go contains `helpView(m.width, m.noEmoji)`
- [x] internal/tui/app/model.go contains `[disc]` (ASCII badge in help overlay)
- [x] internal/tui/model_test.go contains `func newTestModelNoEmoji()`
- [x] internal/tui/model_test.go contains `TestNoEmoji_TreeRenders_ASCIIIcons`
- [x] internal/tui/model_test.go contains `TestNoEmoji_TreeRenders_ASCIIBadges`
- [x] internal/tui/model_test.go contains `TestNoEmoji_HelpOverlay_ASCIIBadges`
- [x] internal/tui/model_test.go contains `TestNoEmoji_False_RendersEmoji`
- [x] Commits afbaeea and d242d53 exist
