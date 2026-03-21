---
phase: 05-tui-polish
plan: 03
subsystem: ui
tags: [bubbletea, lipgloss, tui, keyboard, overlay, state-machine]

# Dependency graph
requires:
  - phase: 05-01
    provides: tree ExpandAll/CollapseAll methods, keys.go KeyMap with ExpandAll/CollapseAll/Help bindings
  - phase: 05-02
    provides: footer SetRefreshFlash(), Height()=3, two-line hint layout

provides:
  - Double-quit state machine (qq / EscEsc) with Ctrl+C immediate exit
  - Help overlay (?) with rounded-border lipgloss box, dismissable with q or Esc
  - Expand-all (e) and collapse-all (w) delegated to tree.ExpandAll()/CollapseAll()
  - Refresh flash lifecycle: FileChangedMsg -> SetRefreshFlash(true) -> tea.Tick(1s) -> RefreshFlashMsg -> SetRefreshFlash(false)
  - Correct viewport height math: 18 for 80x24 terminal (header 3 + footer 3 = 6)

affects: [future TUI features that add key bindings or overlay layers]

# Tech tracking
tech-stack:
  added: [time (stdlib — tea.Tick usage)]
  patterns:
    - Double-quit state machine via quitPending bool field on Model
    - Help overlay via helpVisible bool; View() returns helpView(width) when true
    - Ctrl+C checked before all other key routing as unconditional escape hatch
    - Refresh flash lifecycle wired through tea.Tick (not a timer reset pattern)

key-files:
  created: []
  modified:
    - internal/tui/app/model.go
    - internal/tui/model_test.go

key-decisions:
  - "[05-03] helpView() is a package-level function (not a method) taking width — keeps View() readable and avoids accessing model state in render path"
  - "[05-03] quitPending reset happens on every non-quit key — simpler than a timeout, matches expected UX for CLI tools"
  - "[05-03] Help overlay captures all keys except Ctrl+C — consistent with overlay conventions; q single-press closes overlay without entering double-quit flow"

patterns-established:
  - "Overlay pattern: helpVisible bool on Model; View() short-circuits to helpView(width) before normal layout"
  - "Key routing order: Ctrl+C first, overlay capture second, double-quit third, feature keys fourth, navigation last"
  - "Refresh flash: trigger SetRefreshFlash(true) on event, schedule tea.Tick, handle *Msg to call SetRefreshFlash(false)"

requirements-completed: [D-05, D-06, D-08]

# Metrics
duration: ~15min
completed: 2026-03-21
---

# Phase 5 Plan 3: App Model Wiring Summary

**Double-quit state machine, help overlay, expand/collapse-all, and refresh flash lifecycle wired into root app model with full TDD coverage**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-03-21T03:30:00Z
- **Completed:** 2026-03-21T03:45:00Z
- **Tasks:** 2 (1 auto + 1 human-verify)
- **Files modified:** 2

## Accomplishments

- Implemented double-quit (qq / EscEsc) and single-key Ctrl+C via quitPending state machine — single q/Esc no longer accidentally quits
- Wired help overlay: ? opens full-pane lipgloss rounded-border box, q or Esc dismisses without quitting, Ctrl+C exits through overlay
- Delegated e/w keys to tree.ExpandAll() / tree.CollapseAll() with viewport content and offset refresh
- Wired refresh flash lifecycle: FileChangedMsg sets footer flash on, tea.Tick(1s) fires RefreshFlashMsg, which sets flash off
- Fixed viewport height math to account for footer Height()=3 (previously 2), giving viewport 18 rows on 80x24
- Added 8 new tests covering all new behaviors; updated 2 existing tests for changed height and double-quit contract

## Task Commits

1. **Task 1: App model — help overlay, double-quit, expand/collapse-all, refresh flash** — `ea1adb3` (feat)
2. **Task 2: Visual verification** — approved by user (checkpoint, no separate commit)

## Files Created/Modified

- `internal/tui/app/model.go` — Added helpVisible/quitPending fields, rewrote KeyMsg handler, added helpView() function, updated View() overlay short-circuit, added RefreshFlashMsg handler, updated FileChangedMsg handler with tea.Tick
- `internal/tui/model_test.go` — Renamed TestQuitQ to TestQuit_DoubleQ, updated TestWindowSizeNormal height expectation to 18, added TestQuit_DoubleEsc, TestQuit_QResetByOtherKey, TestQuit_CtrlCAlwaysQuits, TestHelpOverlay_OpenClose, TestHelpOverlay_CtrlCQuits, TestHelpOverlay_EscCloses, TestExpandAllKey, TestCollapseAllKey

## Decisions Made

- helpView() is a package-level function (not a method) taking width — keeps View() readable and avoids accessing model state in the render path
- quitPending reset on every non-quit key — simpler than a timeout, matches expected UX for CLI tools
- Help overlay captures all keys except Ctrl+C — q single-press closes overlay without entering the double-quit flow

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- All Phase 05 TUI polish decisions (D-01 through D-10) are implemented and visually verified
- Phase 05 is complete — the TUI is production-ready with polish, keyboard ergonomics, and live refresh indication
- No blockers for future enhancement phases

---
*Phase: 05-tui-polish*
*Completed: 2026-03-21*
