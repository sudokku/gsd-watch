---
phase: 08-debug-mode
plan: "01"
subsystem: parser
tags: [debug, observability, parser, OBS-01]
dependency_graph:
  requires: []
  provides: [parser.DebugOut, parser.debugf]
  affects: [internal/parser/phases.go, internal/parser/cache.go]
tech_stack:
  added: []
  patterns: [package-level io.Writer toggle, best-effort debug logging]
key_files:
  created:
    - internal/parser/debug.go
  modified:
    - internal/parser/phases.go
    - internal/parser/cache.go
decisions:
  - "DebugOut is io.Writer not bool — enables bytes.Buffer injection in tests without real stderr"
  - "debugf is unexported — internal helper; DebugOut is exported — main.go sets it"
  - "Format [debug HH:MM:SS] event: details per D-02; Go time format 15:04:05"
  - "D-04 scope boundary: no debug calls in updateFromState, updateFromConfig (STATE.md/config.json/ROADMAP.md/PROJECT.md paths)"
  - "plan_error emits before continue; plan emits after title fallback but before SUMMARY.md override"
metrics:
  duration_minutes: 2
  completed_date: "2026-03-24"
  tasks_completed: 2
  files_changed: 3
---

# Phase 08 Plan 01: Debug Infrastructure Summary

**One-liner:** Parser observability via package-level `io.Writer` with `debugf()` emitting five event types (phase_dir, plan, plan_error, badge, cache HIT/MISS).

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Create debug.go with DebugOut var and debugf helper | 6f6e279 | internal/parser/debug.go |
| 2 | Add debug emit calls to phases.go and cache.go | 14746be | internal/parser/phases.go, internal/parser/cache.go |

## What Was Built

Created the debug infrastructure for the parser package:

- `internal/parser/debug.go`: exported `DebugOut io.Writer` (nil = silent default) and unexported `debugf(event, format string, args ...any)` that writes `[debug HH:MM:SS] event: details\n` when DebugOut is non-nil.
- `internal/parser/phases.go`: 4 debugf calls added — phase_dir (after name resolution), badge (inside badge detection loop), plan (after title fallback), plan_error (on parsePlan error branch).
- `internal/parser/cache.go`: 2 debugf calls added — cache HIT (mtime-equal early return), cache MISS (after mtime update).

D-04 scope boundary respected: `updateFromState`, `updateFromConfig`, and full re-parse paths for ROADMAP.md have no debugf calls.

## Verification

- `go build ./internal/parser/` exits 0
- `go test ./internal/parser/ -count=1` passes (0 regressions)
- `grep -c 'debugf(' internal/parser/phases.go` returns 4
- `grep -c 'debugf(' internal/parser/cache.go` returns 2

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None. DebugOut is wired but not yet set from main.go (that is the --debug flag work, OBS-01, which is a separate plan).

## Self-Check: PASSED

- internal/parser/debug.go: FOUND
- internal/parser/phases.go: FOUND (modified)
- internal/parser/cache.go: FOUND (modified)
- Commit 6f6e279: FOUND
- Commit 14746be: FOUND
