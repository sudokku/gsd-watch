---
phase: 08-debug-mode
plan: "02"
subsystem: parser, cli
tags: [debug, observability, parser, OBS-01, testing]
dependency_graph:
  requires: [parser.DebugOut, parser.debugf]
  provides: [--debug CLI flag, debug_test.go coverage]
  affects: [cmd/gsd-watch/main.go, internal/parser/debug_test.go]
tech_stack:
  added: []
  patterns: [flag.Bool CLI flag, bytes.Buffer test injection, regex output assertion]
key_files:
  created:
    - internal/parser/debug_test.go
  modified:
    - cmd/gsd-watch/main.go
decisions:
  - "--debug flag set before TMUX check and tea.NewProgram — matches D-01 wiring requirement"
  - "Tests go directly GREEN because infrastructure was fully in place from plan 01"
  - "TestDebugCacheHIT calls Update() immediately after ParseFull() without sleep — mtime recorded by ParseFull guarantees HIT"
  - "TestDebugSilentByDefault asserts DebugOut==nil directly (package zero value test)"
metrics:
  duration_minutes: 2
  completed_date: "2026-03-24"
  tasks_completed: 2
  files_changed: 2
---

# Phase 08 Plan 02: Debug Flag Wiring and Tests Summary

**One-liner:** `--debug` CLI flag wires `parser.DebugOut = os.Stderr` in main.go; 8 tests cover all five debug event types, silence-by-default, and log format, completing OBS-01.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Wire --debug flag in main.go and update help text | 2dafb05 | cmd/gsd-watch/main.go |
| 2 | Create debug_test.go with tests for all event types | 6d75ac0 | internal/parser/debug_test.go |

## What Was Built

**Task 1 — main.go `--debug` flag:**
- Added `debugMode := flag.Bool("debug", false, "Print parser decisions to stderr")` alongside existing `showHelp`
- Set `parser.DebugOut = os.Stderr` when `*debugMode` is true, after `flag.Parse()` and before the TMUX check
- Added `"github.com/radu/gsd-watch/internal/parser"` import
- Updated `--help` output to include a `Flags:` section listing both `--help` and `--debug`

**Task 2 — debug_test.go (8 tests):**
- `TestDebugSilentByDefault`: Confirms `DebugOut` is nil at package init; `debugf` with nil DebugOut does not panic
- `TestDebugPhaseDir`: `parsePhases` emits `phase_dir:` with `num=N` and `name="Phase N: ..."`
- `TestDebugPlan`: `parsePhases` emits `plan:` with `status=`, `title=`, `wave=`
- `TestDebugPlanError`: `parsePhases` emits `plan_error:` with `err=` for invalid YAML frontmatter (`[invalid yaml`)
- `TestDebugBadge`: `parsePhases` emits `badge:` containing `discussed` for `01-CONTEXT.md`
- `TestDebugCacheHIT`: `cache.Update(path)` with unchanged mtime emits `cache:` and `HIT`
- `TestDebugCacheMISS`: `cache.Update(path)` after file write emits `cache:` and `MISS`
- `TestDebugFormat`: `debugf` output matches `\[debug \d{2}:\d{2}:\d{2}\] event: details\n`

All tests use `DebugOut = &buf` injection + `defer func() { DebugOut = nil }()` cleanup pattern to isolate state between tests.

## Verification

- `go build ./cmd/gsd-watch/` exits 0
- `./gsd-watch --help` lists `--debug    Print parser decisions to stderr`
- `go test ./internal/parser/ -run TestDebug -v -count=1` — all 8 tests PASS
- `go test ./... -count=1` — full suite PASS, no regressions

## Deviations from Plan

**TDD note:** Tests were written first (RED step) but went GREEN immediately — the debug infrastructure (DebugOut, debugf, emit calls in phases.go and cache.go) was fully in place from plan 01. No implementation code was needed in this plan; plan 01 pre-built everything this plan tests.

No correctness or behavior deviations.

## Known Stubs

None. OBS-01 is fully satisfied: `--debug` flag wires `parser.DebugOut = os.Stderr`; all event types have test coverage.

## Self-Check: PASSED

- cmd/gsd-watch/main.go: FOUND (modified)
- internal/parser/debug_test.go: FOUND (created)
- Commit 2dafb05: FOUND
- Commit 6d75ac0: FOUND
