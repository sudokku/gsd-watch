---
phase: 09-quick-tasks-tui-section
plan: 01
subsystem: parser
tags: [go, parser, quick-tasks, tdd]

requires:
  - phase: 08-debug-mode
    provides: debugf() function in debug.go used for quick_task_dir events

provides:
  - QuickTask type in types.go with DirName, DisplayName, Date, Status fields
  - QuickTasks []QuickTask field on ProjectData
  - parseQuickTasks function walking .planning/quick/ and detecting status from file presence
  - ParseProject populates ProjectData.QuickTasks

affects:
  - 09-02-quick-tasks-tui-section (TUI plan consuming ProjectData.QuickTasks)

tech-stack:
  added: []
  patterns:
    - Status detection from file presence (SUMMARY.md=complete, PLAN.md-only=in_progress, neither=pending)
    - Newest-first sort by YYMMDD date prefix
    - Display name humanization (strip date-id prefix, replace dashes with spaces)
    - Malformed dir skipping via regex guard on quickTaskDirRe

key-files:
  created:
    - internal/parser/quick.go
    - internal/parser/quick_test.go
    - internal/parser/testdata/quick/260101-ab1-sample-complete-task/260101-ab1-PLAN.md
    - internal/parser/testdata/quick/260101-ab1-sample-complete-task/260101-ab1-SUMMARY.md
    - internal/parser/testdata/quick/260215-cd2-another-in-progress-task/260215-cd2-PLAN.md
    - internal/parser/testdata/project/quick/260101-ab1-sample-complete-task/260101-ab1-PLAN.md
    - internal/parser/testdata/project/quick/260101-ab1-sample-complete-task/260101-ab1-SUMMARY.md
    - internal/parser/testdata/project/quick/260215-cd2-another-in-progress-task/260215-cd2-PLAN.md
  modified:
    - internal/parser/types.go
    - internal/parser/parser.go

key-decisions:
  - "quickTaskDirRe matches ^(\\d{6})-(\\w+)-(.+)$ — requires exactly 6-digit date, alphanumeric ID, and slug; non-matching dirs silently skipped"
  - "parseQuickTasks returns nil (not empty slice) for missing dir or empty dir — consistent with parsePhases behavior"
  - "TestParseQuickTasks_EmptyDir uses t.TempDir() directly (no matching subdirs) — returns nil not empty slice"
  - "Tests placed in package parser (not parser_test) to access unexported parseQuickTasks directly"

patterns-established:
  - "Pattern: os.Stat error nil = file exists (hasSummary, hasPlan detection)"
  - "Pattern: sort.Slice with Date string comparison (lexicographic YYMMDD descending = newest first)"

requirements-completed: [QT-02]

duration: 2min
completed: 2026-03-24
---

# Phase 9 Plan 01: Quick Tasks TUI Section — Parser Summary

**QuickTask parser with status detection from file presence, newest-first sorting, display name humanization, and 11 unit tests covering all paths**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-24T21:11:37Z
- **Completed:** 2026-03-24T21:13:25Z
- **Tasks:** 1 (TDD)
- **Files modified:** 10

## Accomplishments

- QuickTask type added to types.go with DirName, DisplayName, Date, Status fields
- ProjectData.QuickTasks field added — data layer ready for TUI plan 02 consumption
- parseQuickTasks walks .planning/quick/, detects complete/in_progress/pending from SUMMARY.md and PLAN.md presence, humanizes names, sorts newest-first, emits debugf events
- ParseProject wired: data.QuickTasks = parseQuickTasks(filepath.Join(root, "quick"))
- 11 unit tests pass including integration test TestParseProject_QuickTasks
- Full test suite green with no regressions across all packages

## Task Commits

Each task was committed atomically:

1. **Task 1: QuickTask type + parseQuickTasks parser + test fixtures** - `f12ce5a` (feat)

## Files Created/Modified

- `internal/parser/types.go` - Added QuickTask struct and QuickTasks []QuickTask to ProjectData
- `internal/parser/quick.go` - parseQuickTasks implementation with quickTaskDirRe, status detection, debugf, sort
- `internal/parser/parser.go` - Wired parseQuickTasks call before return in ParseProject
- `internal/parser/quick_test.go` - 11 unit tests: Complete, InProgress, Pending, MissingDir, EmptyDir, Sort, DisplayName, SkipsNonDirs, SkipsMalformedDirs, Debug, TestParseProject_QuickTasks
- `internal/parser/testdata/quick/` - Test fixtures for standalone parseQuickTasks tests
- `internal/parser/testdata/project/quick/` - Test fixtures for ParseProject integration test

## Decisions Made

- Tests placed in `package parser` (not `parser_test`) to access unexported `parseQuickTasks` directly — same pattern as phases_test.go
- `parseQuickTasks` returns nil for empty/missing dirs (not empty slice) — consistent with `parsePhases` nil-return-on-error behavior
- `quickTaskDirRe` requires exactly 6-digit date prefix — dirs with 5 or 7 digits are silently skipped per D-01 design

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed os.Stat call syntax**
- **Found during:** Task 1 (GREEN phase compilation)
- **Issue:** Plan pseudocode used `hasSummary := os.Stat(...)` but os.Stat returns two values
- **Fix:** Changed to `_, errSummary := os.Stat(...)` and `_, errPlan := os.Stat(...)` with nil error checks
- **Files modified:** internal/parser/quick.go
- **Verification:** go test passes after fix
- **Committed in:** f12ce5a (part of task commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Minimal — pseudocode in plan was illustrative, fix was trivial. No scope creep.

## Issues Encountered

None beyond the os.Stat syntax fix noted above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- QuickTask type and parseQuickTasks fully implemented and tested
- ProjectData.QuickTasks populated on every ParseProject call
- Ready for Plan 02: TUI tree section rendering ProjectData.QuickTasks as collapsible tree rows

---
*Phase: 09-quick-tasks-tui-section*
*Completed: 2026-03-24*
