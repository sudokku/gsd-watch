---
phase: 02-live-data-layer
plan: 03
subsystem: parser
tags: [go, parser, tui, bubbletea, filesystem]

# Dependency graph
requires:
  - phase: 02-01
    provides: parsePlan, parseConfig, extractFrontmatter functions
  - phase: 02-02
    provides: parseState, parseRoadmap functions
provides:
  - ParseProject(): single public entry point assembling all parsers into ProjectData
  - parsePhases(): filesystem directory walker with badge detection and SUMMARY.md override
  - ProgressPercent field on ProjectData (float64, from STATE.md progress.percent)
  - app.Init() dispatches async ParseProject command via tea.Cmd
  - header uses ProgressPercent (STATE.md milestone progress, not computed from plan counts)
affects:
  - 03-file-watching (watcher triggers re-parse via ParseProject)
  - 04-plugin-delivery (binary runs ParseProject from project root on launch)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - ParseProject is always best-effort — never returns error, always returns valid ProjectData with "unknown" defaults
    - SUMMARY.md presence overrides PLAN.md frontmatter status to "complete" (PARSE-02)
    - Badge detection via fixed suffix list (CONTEXT.md, RESEARCH.md, VERIFICATION.md, UAT.md) mapped to badge constants
    - derivePhaseStatus computes from plan statuses: all complete → complete, any in_progress → in_progress, else pending

key-files:
  created:
    - internal/parser/parser.go — ParseProject top-level assembler, single public entry point
    - internal/parser/phases.go — parsePhases, parsePlansInDir, derivePhaseStatus
    - internal/parser/phases_test.go — 12 TestParsePhases tests (badge detection, SUMMARY override, active plan, etc.)
    - internal/parser/parser_test.go — ParseProject integration tests (full fixture, missing root, empty root)
    - internal/parser/testdata/phases/ — fixture directories for phases tests
    - internal/parser/testdata/project/ — fixture .planning/ structure for integration tests
  modified:
    - internal/parser/types.go — added ProgressPercent float64 field to ProjectData
    - internal/tui/header/model.go — SetData/New use data.ProgressPercent (not CompletionPercent())
    - internal/tui/header/model_test.go — progress bar tests use ProgressPercent directly
    - internal/tui/app/model.go — Init() dispatches async parser.ParseProject command

key-decisions:
  - "parsePhases walks filesystem — phase list is filesystem truth, not config (PARSE-07)"
  - "SUMMARY.md presence overrides plan status to complete regardless of frontmatter (PARSE-02)"
  - "app.Init() returns tea.Cmd closure calling ParseProject from os.Getwd()/.planning"
  - "header ProgressPercent reads STATE.md progress.percent (milestone-level), not computed from plan counts"
  - "ParseProject never returns error — missing/malformed files yield best-effort defaults (PARSE-08)"
  - "testdata/project/ uses copied (not symlinked) fixture files for hermetic tests"

patterns-established:
  - "ParseProject pattern: read config → read state → read roadmap → walk phases — always returns ProjectData"
  - "Phase directory naming: NN-name prefix regex, sorted alphabetically"
  - "planFileRe = NN-NN-PLAN.md pattern distinguishes plans from lifecycle docs in phase dirs"

requirements-completed: [PARSE-02, PARSE-06, PARSE-07, PARSE-08]

# Metrics
duration: 5min
completed: 2026-03-20
---

# Phase 2 Plan 3: ParseProject Assembler + TUI Wiring Summary

**Filesystem-driven phase walker with SUMMARY.md override, badge detection, and async ParseProject wired into app.Init() — TUI now shows live .planning/ data on launch**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-19T23:24:07Z
- **Completed:** 2026-03-20T00:28:00Z
- **Tasks:** 2 completed (Task 3 is checkpoint awaiting visual verification)
- **Files modified:** 10

## Accomplishments
- Implemented `parsePhases()` directory walker with badge detection, SUMMARY.md override, phase status derivation, and active plan marking
- Added `ProgressPercent float64` field to `ProjectData` — sourced from STATE.md progress.percent (0.0-1.0)
- Created `ParseProject()` — the single public assembler that composes config, state, roadmap, and phase parsers
- Wired live data into TUI: `app.Init()` dispatches async `ParseProject` command; header reads `ProgressPercent` from STATE.md

## Task Commits

Each task was committed atomically:

1. **Task 1: Add ProgressPercent field + phases.go directory walker + badge detection + SUMMARY.md override** - `483f974` (feat)
2. **Task 2: ParseProject assembler + TUI wiring (app.Init + header ProgressPercent)** - `1d0b537` (feat)

_Task 3 is a human-verify checkpoint (visual TUI verification) — no commit required._

## Files Created/Modified
- `internal/parser/types.go` — Added ProgressPercent float64 field to ProjectData
- `internal/parser/phases.go` — parsePhases, parsePlansInDir, derivePhaseStatus (new file)
- `internal/parser/phases_test.go` — 12 TestParsePhases tests covering all behaviors (new file)
- `internal/parser/parser.go` — ParseProject top-level assembler (new file)
- `internal/parser/parser_test.go` — Integration tests: full fixture, missing root, empty root (new file)
- `internal/parser/testdata/phases/` — Phase fixture dirs for unit tests (new)
- `internal/parser/testdata/project/` — Full .planning/ fixture for integration tests (new)
- `internal/tui/header/model.go` — SetData/New use data.ProgressPercent instead of CompletionPercent()
- `internal/tui/header/model_test.go` — Updated progress bar tests to use ProgressPercent field directly
- `internal/tui/app/model.go` — Init() dispatches async parser.ParseProject via tea.Cmd

## Decisions Made
- `parsePhases` walks the filesystem as primary source of truth (PARSE-07) — phase list is not config-driven
- SUMMARY.md presence is the sole override for plan status (PARSE-02), checked after frontmatter parse
- `ParseProject` never returns error — callers always get a valid ProjectData (PARSE-08)
- Header uses STATE.md progress.percent as milestone-level progress, not computed from plan counts
- `app.Init()` calls `os.Getwd()` to locate `.planning/` — binary must run from project root
- Updated header tests to use `ProgressPercent` directly since plan status no longer drives bar width

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Updated header_test.go progress bar tests to use ProgressPercent**
- **Found during:** Task 2 (header/model.go modification)
- **Issue:** Existing tests set plan statuses to drive bar fill (50%, 0%, 100%), which relied on CompletionPercent(). After switching to ProgressPercent, those tests would break because ProgressPercent defaults to 0.
- **Fix:** Updated TestHeaderView_ProgressBar50Percent, TestHeaderView_ZeroPercent, TestHeaderView_HundredPercent to set data.ProgressPercent directly
- **Files modified:** internal/tui/header/model_test.go
- **Verification:** go test ./internal/tui/header/ passes
- **Committed in:** 1d0b537 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 2 - correctness)
**Impact on plan:** Test update was required for correctness after header API change. No scope creep.

## Issues Encountered
None — plan executed as specified with one test update required by the ProgressPercent API change.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Live data layer complete: TUI launches and reads real .planning/ files via ParseProject
- Phase 3 (file watching) can attach fsnotify watcher to call ParseProject on file changes
- ParseProject is the stable public API — watcher just needs to call it and send ParsedMsg

## Self-Check: PASSED

- internal/parser/parser.go: FOUND
- internal/parser/phases.go: FOUND
- internal/parser/phases_test.go: FOUND
- internal/parser/parser_test.go: FOUND
- internal/parser/testdata/phases/01-core-tui-scaffold/01-01-SUMMARY.md: FOUND
- internal/parser/testdata/phases/01-core-tui-scaffold/01-CONTEXT.md: FOUND
- .planning/phases/02-live-data-layer/02-03-SUMMARY.md: FOUND
- commit 483f974: FOUND
- commit 1d0b537: FOUND

---
*Phase: 02-live-data-layer*
*Completed: 2026-03-20*
