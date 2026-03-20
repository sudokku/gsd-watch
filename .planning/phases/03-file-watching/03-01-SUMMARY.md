---
phase: 03-file-watching
plan: 01
subsystem: watcher
tags: [fsnotify, goroutine, debounce, tdd]

# Dependency graph
requires:
  - phase: 01-core-tui-scaffold
    provides: FileChangedMsg type in internal/tui/messages.go
  - phase: 02-live-data-layer
    provides: parser.ProjectData and tui.ParsedMsg established

provides:
  - internal/watcher package with Run(root, events) goroutine
  - Recursive fsnotify monitoring of .planning/ via filepath.WalkDir
  - Dynamic directory addition on fsnotify.Create events
  - 300ms per-path debounce via map[string]*time.Timer
  - Comprehensive tests: TestRunAddsSubdirs, TestDynamicDirAdd, TestDebounce, TestFilterOps

affects: [03-02-parser-cache, 03-03-app-wiring]

# Tech tracking
tech-stack:
  added: [github.com/fsnotify/fsnotify v1.9.0 (direct)]
  patterns:
    - per-path timer map debounce using time.AfterFunc
    - kqueue-safe recursive watching via explicit WalkDir + dynamic add
    - directory Create events filtered from FileChangedMsg output

key-files:
  created:
    - internal/watcher/watcher.go
    - internal/watcher/watcher_test.go
  modified:
    - go.mod (fsnotify promoted to direct dependency)
    - go.sum

key-decisions:
  - "Directory Create events do not produce FileChangedMsg — only file-level Create/Write ops do"
  - "time.AfterFunc with t.Stop() + t.Reset() is the debounce primitive (Go 1.26.1 has clean timer.Reset semantics)"
  - "path := e.Name captured locally before AfterFunc closure to be explicit about variable scope even on Go 1.22+"

patterns-established:
  - "Debounce pattern: map[string]*time.Timer with AfterFunc + Stop + Reset(300ms) per path"
  - "kqueue compatibility: WalkDir on startup + dynamic watcher.Add on directory Create"
  - "TDD: failing tests committed first (RED), then implementation (GREEN)"

requirements-completed: [WATCH-01, WATCH-02, WATCH-03]

# Metrics
duration: 8min
completed: 2026-03-20
---

# Phase 3 Plan 01: File Watcher Package Summary

**fsnotify-based recursive watcher with 300ms per-path debounce — internal/watcher package delivering WATCH-01, WATCH-02, WATCH-03**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-20T03:14:00Z
- **Completed:** 2026-03-20T03:22:00Z
- **Tasks:** 1 (TDD: 2 commits — RED + GREEN)
- **Files modified:** 4

## Accomplishments

- Created `internal/watcher` package with `Run(root string, events chan<- tea.Msg)` goroutine
- WalkDir recursively adds all subdirectories on startup (kqueue/macOS safe) — WATCH-01
- Dynamic directory watching: new directories added on fsnotify.Create events — WATCH-02
- 300ms per-path debounce collapses rapid writes into single FileChangedMsg — WATCH-03
- All 4 tests pass: TestRunAddsSubdirs, TestDynamicDirAdd, TestDebounce, TestFilterOps

## Task Commits

Each TDD phase committed atomically:

1. **RED — Failing tests** - `13e34f7` (test)
2. **GREEN — Implementation** - `f5ea57f` (feat)
3. **Dependency update** - `02a2740` (chore)

**Plan metadata:** (docs commit below)

_Note: TDD task produces separate test and feat commits; implementation fixed one test failure during GREEN phase (directory Create filtering)._

## Files Created/Modified

- `internal/watcher/watcher.go` — Run() goroutine with fsnotify loop, WalkDir, dynamic dir add, 300ms debounce
- `internal/watcher/watcher_test.go` — Integration-style unit tests using real temp dirs and real fsnotify
- `go.mod` — fsnotify v1.9.0 promoted to direct dependency
- `go.sum` — updated checksums

## Decisions Made

- **Directory Creates do not produce FileChangedMsg**: When a fsnotify.Create event fires for a directory, we add it to the watcher (WATCH-02) and `continue` without debouncing — directories are not "file changes" for the TUI update pipeline.
- **time.AfterFunc with explicit Stop**: Timer is created in stopped state (AfterFunc with MaxInt64 then Stop) so the first t.Reset(300ms) starts the clock cleanly. This is the pattern from the official fsnotify dedup example.
- **Local path capture**: `path := e.Name` captured before AfterFunc closure for clarity, even though Go 1.26.1 loop variable semantics don't require it.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Directory Create events produced spurious FileChangedMsg**
- **Found during:** Task 1 GREEN phase (first test run)
- **Issue:** `TestDynamicDirAdd` failed because the directory creation event itself was being debounced and sent as a `FileChangedMsg` with the directory path. The test drained messages after mkdir but the debounced message arrived after the drain window.
- **Fix:** Added `continue` after `w.Add(e.Name)` for directory Create events — directories should not produce FileChangedMsg, only files should.
- **Files modified:** internal/watcher/watcher.go
- **Verification:** TestDynamicDirAdd passes; all 4 tests green
- **Committed in:** f5ea57f (GREEN implementation commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Fix was necessary for correctness — the plan spec says "only Write + Create ops produce messages" implying file-level events. Directory events are infrastructure-only. No scope creep.

## Issues Encountered

None beyond the auto-fixed bug above.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `internal/watcher` package complete and fully tested
- `Run(root, events)` function signature matches the wiring spec from CONTEXT.md
- Ready for Plan 02: parser cache (WATCH-04) and Plan 03: app wiring (WATCH-05)
- Remaining concern: `internal/parser/cache.go` exists as untracked file from prior session — Plan 02 will address this

---
*Phase: 03-file-watching*
*Completed: 2026-03-20*
