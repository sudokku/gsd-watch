---
phase: 03-file-watching
plan: 03
subsystem: ui
tags: [bubbletea, fsnotify, go, incremental-cache, file-watching]

# Dependency graph
requires:
  - phase: 03-01
    provides: watcher.Run() goroutine with fsnotify + debounce
  - phase: 03-02
    provides: parser.ProjectCache with ParseFull() and Update() methods
  - phase: 01-04
    provides: app.Model root TUI model structure
provides:
  - Live-updating TUI app model wired to watcher goroutine and ProjectCache
  - FileChangedMsg event loop via waitForEvent re-arm pattern
  - events channel created in main() and passed through to app.New(events)
affects: [04-plugin-delivery]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "waitForEvent pattern: tea.Cmd blocking on channel, re-armed after every FileChangedMsg for perpetual event loop"
    - "events channel created in main(), stored in Model, passed to watcher goroutine — single ownership point"
    - "cache and planningRoot initialised in New() (pointer receiver), watcher goroutine started in Init() (value receiver)"

key-files:
  created: []
  modified:
    - cmd/gsd-watch/main.go
    - internal/tui/app/model.go
    - internal/tui/model_test.go

key-decisions:
  - "Watcher goroutine started from Init() not New() — Init() runs after Bubble Tea runtime is ready"
  - "events channel buffer size 10 per CONTEXT.md decision — prevents goroutine blocking during re-parse"
  - "FileChangedMsg handler returns tea.Batch(asyncParseCmd, waitForEvent) — perpetuates loop and parses concurrently"
  - "Model starts with empty ProjectData; data arrives via ParsedMsg from ParseFull() in Init() cmd"

patterns-established:
  - "waitForEvent(ch chan tea.Msg) tea.Cmd: idiomatic Bubble Tea channel bridge for external goroutine messages"
  - "tea.Batch in FileChangedMsg: dispatch async work AND re-arm channel listener simultaneously"

requirements-completed: [WATCH-05]

# Metrics
duration: 25min
completed: 2026-03-20
---

# Phase 3 Plan 03: Wire Watcher + Cache into TUI Summary

**fsnotify watcher and ProjectCache wired into Bubble Tea app model via events channel and waitForEvent perpetual loop, completing the filesystem-to-screen live-update pipeline**

## Performance

- **Duration:** ~25 min (including human-verify checkpoint)
- **Started:** 2026-03-20T01:30:10Z
- **Completed:** 2026-03-20T01:56:25Z
- **Tasks:** 2 of 2 (Task 1: implementation; Task 2: human-verify approved)
- **Files modified:** 3

## Accomplishments
- `main.go` creates `events := make(chan tea.Msg, 10)` and passes to `app.New(events)`
- `Model` struct gains `cache *parser.ProjectCache`, `events chan tea.Msg`, `planningRoot string` fields
- `New(events)` initialises cache via `parser.NewCache(planningRoot)`, stores channel — no more mock data
- `Init()` starts `go watcher.Run(m.planningRoot, m.events)` goroutine and returns `tea.Batch(parseFull, waitForEvent)`
- `Update()` handles `FileChangedMsg` — dispatches async `cache.Update(path)` and re-arms `waitForEvent`
- Full build and all tests pass — complete live-update pipeline: filesystem event → watcher → channel → waitForEvent → Update() → cache.Update() → ParsedMsg → SetData()

## Task Commits

Each task was committed atomically:

1. **Task 1: Wire watcher and cache into app model** - `f1fcdf2` (feat)
2. **Task 2: Verify live file watching works end-to-end** - checkpoint:human-verify (approved — no code commit)

**Plan metadata:** (docs commit — this summary)

## Files Created/Modified
- `cmd/gsd-watch/main.go` - Creates events channel (buf 10), passes to app.New(events)
- `internal/tui/app/model.go` - Adds cache/events/planningRoot fields, updates New/Init/Update, removes mock import
- `internal/tui/model_test.go` - Updated to pass events channel to app.New(); data-dependent tests inject via ParsedMsg

## Decisions Made
- Watcher goroutine started from `Init()` not `New()` — aligns with PROJECT.md decision; `Init()` runs after Bubble Tea runtime is ready to receive messages
- Channel buffer size 10 per CONTEXT.md — sufficient to absorb burst writes without blocking the watcher goroutine
- `cache` and `planningRoot` initialised in `New()` (value receiver, safe to mutate returned value) — not in `Init()` which uses value receiver and would lose mutations
- `waitForEvent` re-armed in `tea.Batch` alongside async parse cmd — both run concurrently, loop perpetuates

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated model_test.go to match new app.New(events) signature**
- **Found during:** Task 1 (build verification)
- **Issue:** Existing test file called `app.New()` with no arguments; new signature requires `chan tea.Msg`
- **Fix:** Added `newTestModel()` helper creating `app.New(make(chan tea.Msg, 10))`; updated 6 test calls; fixed 2 data-dependent tests to inject `tui.ParsedMsg{Project: mock.MockProject()}` since empty state no longer has project data
- **Files modified:** `internal/tui/model_test.go`
- **Verification:** `go test ./... -timeout 60s` all pass
- **Committed in:** f1fcdf2 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary fix — test file was written for the old no-args signature. No scope creep.

## Issues Encountered
None beyond the test signature mismatch (documented as deviation above).

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Complete live-update pipeline is implemented, tests pass, and human verification approved
- User confirmed: TUI updates within 300ms, new directories detected without crash, rapid writes debounced to single update, clean quit works
- Phase 3 (File Watching) is fully complete — Phase 4 (Plugin & Delivery) is unblocked

---
*Phase: 03-file-watching*
*Completed: 2026-03-20*
