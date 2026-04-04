---
phase: 19-slash-command-cmux-detection
plan: "01"
subsystem: cli
tags: [slash-command, cmux, tmux, multiplexer, bash]

# Dependency graph
requires:
  - phase: 18-go-binary-multiplexer-detection
    provides: CMUX_WORKSPACE_ID env var convention and OS-aware error message format
provides:
  - Three-branch Step 2 in gsd-watch slash command (cmux, tmux, error paths)
  - cmux instructional stub message (Phase 20 replaces with real spawning)
  - OS-aware not-in-any-multiplexer error naming both tmux and cmux
affects:
  - 20-cmux-pane-spawning

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Slash command multiplexer check: cmux (CMUX_WORKSPACE_ID) before tmux (TMUX) — mirrors Go binary inTmux/inCmux pattern from Phase 18"
    - "OS-aware error via uname -s in Bash — same approach as Makefile uname -m arch detection"

key-files:
  created: []
  modified:
    - commands/gsd-watch.md

key-decisions:
  - "Check CMUX_WORKSPACE_ID before TMUX — mirrors Phase 18 D-01 check order"
  - "cmux branch prints instructional stub and stops — Phase 20 replaces with real socket-based spawning"
  - "Error message names both tmux and cmux with OS-aware install hints via uname -s"
  - "Steps 1, 3, and 4 left verbatim unchanged — tmux regression prevention per SPAWN-02"

patterns-established:
  - "Three-branch multiplexer check pattern: cmux-first, tmux-second, error-third"

requirements-completed: [SPAWN-01, SPAWN-02]

# Metrics
duration: 5min
completed: 2026-04-04
---

# Phase 19 Plan 01: Slash Command cmux Detection Summary

**Three-branch Step 2 in gsd-watch slash command: cmux stub (CMUX_WORKSPACE_ID), tmux proceed, OS-aware error naming both multiplexers**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-04T15:30:00Z
- **Completed:** 2026-04-04T15:35:00Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Replaced single-branch tmux-only Step 2 with three-branch multiplexer check
- cmux path (CMUX_WORKSPACE_ID set): prints instructional stub and stops cleanly before Step 3
- tmux path (TMUX set): proceeds to existing Steps 3 and 4 unchanged
- Neither path: OS-aware error via `uname -s` naming both tmux and cmux with correct install hints

## Task Commits

Each task was committed atomically:

1. **Task 1: Rewrite Step 2 — three-branch multiplexer check** - `8a00a76` (feat)

**Plan metadata:** (docs commit — see below)

## Files Created/Modified
- `commands/gsd-watch.md` - Step 2 rewritten with cmux-first three-branch multiplexer check

## Decisions Made
- Check CMUX_WORKSPACE_ID before TMUX per D-01 in 19-CONTEXT.md (mirrors Phase 18 Go binary pattern)
- cmux branch terminates before Steps 3+4 (tmux-only) — correct isolation
- uname -s OS detection inline in the error branch — same pattern as Makefile arch detection

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 20 (cmux pane spawning) can now replace the stub branch with real `nc -U $CMUX_SOCKET_PATH` JSON-RPC spawning logic
- Steps 3 and 4 (tmux duplicate detection and spawn) are unchanged and ready for any future tmux work

---
*Phase: 19-slash-command-cmux-detection*
*Completed: 2026-04-04*
