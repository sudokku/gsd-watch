---
phase: 15-help-overlay-config-hint
plan: 01
subsystem: ui
tags: [tui, help-overlay, config, theme, lipgloss]

# Dependency graph
requires:
  - phase: 13-config-infrastructure
    provides: config.ConfigPath constant, config.Config.Theme field, config.Defaults()
  - phase: 14-theme-system
    provides: Theme sentinel — empty string means "default"
provides:
  - helpView renders Config section showing tilde-abbreviated config path and active theme name
affects: [future-settings-panel]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - helpView extended with string params (configPath, themeName) following same pattern as noEmoji bool in Phase 10
    - Caller resolves display values inline in View() before passing to pure render function

key-files:
  created: []
  modified:
    - internal/tui/app/model.go
    - internal/tui/model_test.go

key-decisions:
  - "helpView(width, noEmoji, configPath, themeName) — two string params added; caller resolves before calling, keeps function pure"
  - "Config path tilde-abbreviated inline in View() via filepath.Join(home, config.ConfigPath) then strings.Replace"
  - "Empty theme sentinel normalized to 'default' string in View() before passing to helpView"
  - "fmt.Sprintf used for configSection interpolation inside helpView"

patterns-established:
  - "Extend helpView with additional display params computed at call site — avoids storing render-only state in struct"

requirements-completed: ["DISC-01", "DISC-02"]

# Metrics
duration: 8min
completed: 2026-03-27
---

# Phase 15 Plan 01: Add Config Path and Theme Name to Help Overlay Summary

**helpView extended with Config section showing `~/.config/gsd-watch/config.toml` and active theme name; two acceptance tests cover DISC-01 and DISC-02**

## Performance

- **Duration:** 8 min
- **Started:** 2026-03-27T00:00:00Z
- **Completed:** 2026-03-27T00:08:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Extended `helpView` signature with `configPath, themeName string` params following the Phase 10 `noEmoji bool` extension pattern
- Added a `Config` section to the help overlay displaying the tilde-abbreviated config file path and resolved theme name
- Updated `View()` to compute both values inline and pass them to `helpView`
- Added `TestHelpOverlay_ContainsConfigPath` (DISC-01) and `TestHelpOverlay_ContainsThemeName` (DISC-02) tests

## Task Commits

Each task was committed atomically:

1. **Task 1 + Task 2: Extend helpView and add tests** - `94e079e` (feat)

## Files Created/Modified
- `internal/tui/app/model.go` — extended `helpView` signature, added Config section, updated `View()` call site; added `fmt` and `strings` imports
- `internal/tui/model_test.go` — added two new tests for DISC-01 and DISC-02

## Decisions Made
- `helpView` receives pre-computed `configPath` and `themeName` strings — caller owns resolution, function stays pure and testable
- Tilde abbreviation computed via `strings.Replace(filepath.Join(home, config.ConfigPath), home, "~", 1)` — one call site, no helper needed
- Empty `cfg.Theme` sentinel normalized to `"default"` string in `View()`, not inside `helpView` — consistent with Phase 14 D-06 pattern

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- DISC-01 and DISC-02 requirements fulfilled; v1.3 Settings milestone is complete
- Help overlay now surfaces both config file path and active theme name for user discoverability
- No blockers for next phase

## Self-Check: PASSED

- FOUND: `internal/tui/app/model.go`
- FOUND: `internal/tui/model_test.go`
- FOUND: `.planning/phases/15-help-overlay-config-hint/15-01-SUMMARY.md`
- FOUND: commit `94e079e`

---
*Phase: 15-help-overlay-config-hint*
*Completed: 2026-03-27*
