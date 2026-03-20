---
phase: 03-file-watching
plan: 02
subsystem: parser
tags: [go, cache, incremental-parsing, mtime, fsnotify]

# Dependency graph
requires:
  - phase: 02-live-data-layer
    provides: ParseProject, parsePlan, parseState, parseConfig, parsePhases, parsePlansInDir, derivePhaseStatus
provides:
  - ProjectCache struct with root/data/mtimes fields
  - NewCache(root) constructor
  - ParseFull() method wrapping ParseProject
  - Update(path) incremental re-parse routing on filepath.Base
affects:
  - 03-file-watching (watcher integrates ProjectCache for WATCH-04)
  - future phases that consume live ProjectData

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Mtime guard: compare os.Stat(path).ModTime() against cached mtime to skip redundant re-parses
    - Path routing on filepath.Base(path) for deterministic file-type dispatch without regex on full path
    - Mtime index seeded by filepath.WalkDir in ParseFull() for O(1) guard lookups

key-files:
  created:
    - internal/parser/cache.go
    - internal/parser/cache_test.go
  modified: []

key-decisions:
  - "ProjectCache.Update routes on filepath.Base(path) not full path — avoids fragile full-path regex and stays correct regardless of .planning/ location"
  - "STATE.md update triggers parsePhases re-call to refresh IsActive markers — active plan display must reflect new STATE.md active phase/plan values"
  - "ROADMAP.md changes trigger full ParseProject re-parse — phase names affect entire tree, targeted update not worth complexity"
  - "isBadgeFile() checks suffix (ends-with) not exact match — badge filenames include phase prefix (e.g. 01-CONTEXT.md) not just CONTEXT.md"
  - "mustAtoi() helper avoids strconv import in cache.go by inline digit parsing — keeps imports minimal"

patterns-established:
  - "Mtime guard pattern: Stat file, compare Equal(), update map, then route to partial parse"
  - "Fallback pattern: unrecognized file or failed phase lookup falls back to full ParseProject() rather than returning stale data"

requirements-completed: [WATCH-04]

# Metrics
duration: 10min
completed: 2026-03-20
---

# Phase 3 Plan 02: Parser Cache Summary

**ProjectCache struct wrapping ParseProject() with mtime-guarded incremental Update() routing — STATE.md, config.json, PLAN.md, badge files each trigger only the minimal re-parse needed**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-20T01:22:36Z
- **Completed:** 2026-03-20T01:32:00Z
- **Tasks:** 1 (TDD: test + feat commits)
- **Files modified:** 2

## Accomplishments
- `ProjectCache` struct with `root`, `data`, `mtimes` fields targeting WATCH-04 incremental update requirement
- `ParseFull()` wraps existing `ParseProject()` and seeds mtime index via `filepath.WalkDir`
- `Update(path)` routes on `filepath.Base()` to STATE.md, config.json, ROADMAP.md, PLAN.md, badge file, or default full re-parse
- Mtime guard prevents redundant re-parses when file unchanged since last index
- 8 new tests covering all routing paths plus mtime-skip guard; zero regressions in existing 27 parser tests

## Task Commits

Each task was committed atomically (TDD: test then implementation):

1. **Task 1 RED: Failing cache tests** - `ff3104b` (test)
2. **Task 1 GREEN: ProjectCache implementation** - `2962519` (feat)

## Files Created/Modified
- `internal/parser/cache.go` - ProjectCache struct, NewCache, ParseFull, Update with path routing
- `internal/parser/cache_test.go` - 8 tests covering all Update routing paths and mtime guard

## Decisions Made
- `Update()` routes on `filepath.Base(path)` rather than full-path regex — simpler, more robust, correct regardless of .planning/ root location.
- STATE.md updates trigger `parsePhases()` re-call because IsActive markers depend on ActivePhase/ActivePlan from STATE.md; skipping this would show stale active plan highlighting.
- ROADMAP.md triggers full `ParseProject()` — phase name changes are global, targeted update would need to reconstruct the entire phases slice anyway.
- `isBadgeFile()` uses `HasSuffix(base, "-"+bf.suffix)` because badge files are prefixed with the phase number (e.g. `01-CONTEXT.md`), not bare `CONTEXT.md`.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- `ProjectCache` is ready to be wired into the watcher package from plan 03-01 — the watcher calls `cache.Update(changedPath)` on each debounced fsnotify event and sends the result via `p.Send()`.
- No blockers.

---
*Phase: 03-file-watching*
*Completed: 2026-03-20*
