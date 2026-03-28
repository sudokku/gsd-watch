---
phase: quick-260328-on8
plan: 01
subsystem: tui/tree
tags: [empty-state, quick-tasks, archived-milestones, rendering]
dependency_graph:
  requires: []
  provides: [empty-state-quick-tasks-rendering]
  affects: [internal/tui/tree/view.go]
tech_stack:
  added: []
  patterns: [conditional-empty-state-rendering]
key_files:
  created: []
  modified:
    - internal/tui/tree/view.go
    - internal/tui/tree/model_test.go
decisions:
  - Empty state archived+quick branch builds quick tasks inline in padded slice — no new helper function, keeps View() self-contained
  - Use StatusPending (not StatusTodo) — correct constant for tasks not yet started
metrics:
  duration: 6 min
  completed: 2026-03-28
  tasks_completed: 1
  files_modified: 2
---

# Quick 260328-on8: Fix Quick Tasks Visibility When No Active Milestone

**One-liner:** Empty-state archived-milestone branch now renders live quick tasks tree (header + rows with connectors) instead of static "/gsd:quick" hint when quick tasks exist; adds 2-line top spacing to archived message.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Update empty-state rendering for archived-milestone case | e34e99c | internal/tui/tree/view.go, internal/tui/tree/model_test.go |

## Changes Made

### internal/tui/tree/view.go

- **Change 1 — vertical spacing:** Archived-milestone msg now prefixed with `"\n\n"` so placeholder renders 2 lines below pane top.
- **Change 2 — conditional quick tasks rendering:** When `len(t.data.QuickTasks) > 0`, the `msg` drops the static hint suffix and the code appends a live quick tasks section (header + rows with `├──`/`└──` connectors, icons, word-wrapped names, Pending dim for completed tasks) to the `padded` slice. The `"\n\n"` prefix applies in both the quick-tasks and no-quick-tasks branches.
- **No structural change** when `QuickTasks` is empty — only the `"\n\n"` prefix is added to the existing hint.
- **"No GSD project found" branch** is unchanged.

### internal/tui/tree/model_test.go

- Added `TestView_ArchivedOnlyWithQuickTasks`: constructs `ProjectData` with one archived milestone and one pending quick task ("Buy milk"), asserts "Quick tasks" header is present, at least one connector is present, and static "/gsd:quick" hint is absent.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Used correct parser status constant**
- **Found during:** Task 1 build
- **Issue:** Plan said `parser.StatusTodo` which doesn't exist — the correct constant is `parser.StatusPending`
- **Fix:** Changed test to use `parser.StatusPending`
- **Files modified:** internal/tui/tree/model_test.go
- **Commit:** e34e99c (same commit)

## Verification

- `go build ./internal/tui/tree/...` — clean
- `go test ./internal/tui/tree/...` — 50 tests pass, including new `TestView_ArchivedOnlyWithQuickTasks`
- `TestView_ArchivedOnly` still passes, confirming the no-quick-tasks branch retains `/gsd:quick` hint

## Known Stubs

None.

## Self-Check: PASSED

- `internal/tui/tree/view.go` — FOUND
- `internal/tui/tree/model_test.go` — FOUND
- Commit e34e99c — FOUND
