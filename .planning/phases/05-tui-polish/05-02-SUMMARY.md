---
phase: 05-tui-polish
plan: "02"
subsystem: ui
tags: [lipgloss, bubbletea, footer, tui, refresh-indicator]

# Dependency graph
requires:
  - phase: 05-01
    provides: RefreshFlashStyle in styles.go, updated KeyMap with ExpandAll/CollapseAll

provides:
  - FooterModel.SetRefreshFlash() method for refresh flash animation
  - Idle icon (↺ gray) and flash icon (⟳ bold green) in footer timestamp line
  - Two-line keybinding hints layout (nav line + actions/quit line)
  - Footer Height() returns 3 (was 2)

affects:
  - 05-03 (viewport math update for footer height=3)
  - internal/tui/app (viewport height recalculated dynamically from footer.Height())

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Value-receiver pattern: SetRefreshFlash(bool) returns copy of FooterModel"
    - "TDD: RED commit (failing tests) -> GREEN commit (implementation)"

key-files:
  created: []
  modified:
    - internal/tui/footer/model.go
    - internal/tui/footer/model_test.go
    - internal/tui/styles.go
    - internal/tui/model_test.go

key-decisions:
  - "Footer Height() default (width==0) changed from 2 to 3 to match new two-hint-line layout"
  - "Two-line hints are static strings, not derived from KeyMap.ShortHelp() — decoupled from key binding struct for layout control"
  - "TestWindowSizeNormal updated: viewport height now 18 (24-3header-3footer) not 19"

patterns-established:
  - "SetRefreshFlash follows existing value-receiver immutable update pattern (like SetWidth, SetData)"

requirements-completed:
  - D-05
  - D-09

# Metrics
duration: 3min
completed: 2026-03-21
---

# Phase 5 Plan 02: Footer Redesign Summary

**Footer refresh indicator (idle ↺/flash ⟳) with two-line keybinding hints layout replacing single ShortHelp() line**

## Performance

- **Duration:** ~3 min
- **Started:** 2026-03-21T03:20:00Z
- **Completed:** 2026-03-21T03:23:00Z
- **Tasks:** 1 (TDD)
- **Files modified:** 4

## Accomplishments
- Added `refreshFlash bool` field and `SetRefreshFlash(bool)` method to FooterModel
- Idle state shows gray "↺ Xs ago"; flash state shows bold-green "⟳ Xs ago"
- Added `RefreshFlashStyle` (bold + green) to shared styles.go
- Replaced single ShortHelp()-based hints with two static lines: nav (←h · ↓j · ↑k · →l) + actions/quit
- Updated Height() to return len(actionLines())+2 (was +1), default 3 (was 2)
- 10 footer tests all pass; integration test updated for new viewport height

## Task Commits

Each task was committed atomically:

1. **RED: Failing tests** - `96ba5ff` (test)
2. **GREEN: Implementation** - `412afed` (feat)

## Files Created/Modified
- `internal/tui/footer/model.go` - Added refreshFlash field, SetRefreshFlash(), two-line hints in View(), updated Height()
- `internal/tui/footer/model_test.go` - Added 4 new tests (RefreshIdle, RefreshFlash, HeightThreeLines, SetRefreshFlash); updated ContainsKeyHints and Height expectations
- `internal/tui/styles.go` - Added RefreshFlashStyle (bold + green)
- `internal/tui/model_test.go` - Updated TestWindowSizeNormal to expect viewport height 18 (footer now 3 lines)

## Decisions Made
- Two-line hints use static strings rather than KeyMap.ShortHelp() — gives full layout control and avoids the ShortHelp() format not matching the desired compact display
- Footer Height() default (no width set) changed from 2 to 3 to match the actual rendered line count
- TestWindowSizeNormal updated inline: 24-3(header)-3(footer)=18 viewport height

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed TestWindowSizeNormal test stale expectation**
- **Found during:** Task 1 (implementation — GREEN phase)
- **Issue:** test expected viewport height 19 (footer=2) but footer is now 3 lines, making correct answer 18
- **Fix:** Updated test comment and assertion from 19 to 18
- **Files modified:** internal/tui/model_test.go
- **Verification:** Test passes with correct assertion
- **Committed in:** 412afed (feat commit)

---

**Total deviations:** 1 auto-fixed (1 bug/stale test)
**Impact on plan:** Required for test suite to pass; no scope creep.

## Issues Encountered
- `internal/tui/tree` tests fail due to sibling parallel agent (05-01) adding Phase 5 to MockProject — 2 extra phases cause row count mismatches. These are out of scope for this plan (not caused by my changes). Documented in deferred-items.

## Next Phase Readiness
- Footer Height()=3 is ready; Plan 03 must update viewport math if needed
- SetRefreshFlash() API is ready for app model wiring in Plan 03
- RefreshFlashStyle available in styles.go for all tui packages

## Self-Check: PASSED
- FOUND: internal/tui/footer/model.go
- FOUND: internal/tui/styles.go
- FOUND: .planning/phases/05-tui-polish/05-02-SUMMARY.md
- FOUND: commit 96ba5ff (test)
- FOUND: commit 412afed (feat)

---
*Phase: 05-tui-polish*
*Completed: 2026-03-21*
