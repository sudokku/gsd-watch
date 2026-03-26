---
phase: quick
plan: 260326-qai
subsystem: tui/tree
tags: [padding, visual-fix, wrap-width]
tech-stack:
  added: []
  patterns: [symmetric-padding-via-wrapwidth]
key-files:
  modified:
    - internal/tui/tree/view.go
decisions:
  - Right padding implemented implicitly by reducing wrapWidth by 2 instead of 1 — no actual trailing spaces added to lines
metrics:
  duration: 5 min
  completed: "2026-03-26"
  tasks: 1
  files: 1
---

# Quick Task 260326-qai: Fix TUI Trailing Column Visual Debris Summary

Tree View() now uses `width - 2` wrap calculations throughout, giving symmetric 1-char left and right padding so content never reaches the terminal's rightmost column on resize.

## What Was Done

Changed all `wrapWidth := width - 1 - ...` expressions to `wrapWidth := width - 2 - ...` in `internal/tui/tree/view.go`, covering:

- **View() RowPhase** name wrap (~line 73)
- **View() RowPlan** title wrap (~line 159)
- **View() RowQuickTask** name wrap (~line 244)
- **renderedRowLines() RowPhase** (~line 325)
- **renderedRowLines() RowPlan** (~line 353)
- **renderedRowLines() RowQuickTask** (~line 371)

Also updated the D-10 comment at the bottom of View() to explain that right padding is achieved via wrapWidth reduction (not trailing spaces).

## Task Results

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add right padding to tree View() wrap calculations | 3d00bb7 | internal/tui/tree/view.go |

## Verification

- `go build ./...` — clean
- `go test ./internal/tui/tree/... -count=1` — 19/19 PASS
- `go test ./... -count=1` — all packages PASS

## Deviations from Plan

**1. [Rule 1 - Skip] RenderArchiveSeparator not present in this worktree**
- **Found during:** Task 1
- **Issue:** Plan item 7 referenced `RenderArchiveSeparator` at ~line 48, but this function does not exist in the worktree's `internal/tui/tree/view.go`. The worktree is on a different branch (pre-Phase-12) than the main repo where that function was introduced.
- **Fix:** Skipped — the 6 applicable `width - 1` → `width - 2` changes were applied. The separator fix is a no-op here as the function doesn't exist.
- **Impact:** None — the worktree's tree/view.go has no archive separator rendering.

## Self-Check: PASSED

- `internal/tui/tree/view.go` — modified and committed at 3d00bb7
- All `width - 1` expressions in View() and renderedRowLines() changed to `width - 2`
- No trailing `width - 1` patterns remain in wrapWidth calculations
