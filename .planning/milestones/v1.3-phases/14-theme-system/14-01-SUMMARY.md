---
phase: 14-theme-system
plan: 01
subsystem: tui/styles, tui/tree, tui/app, cmd/gsd-watch
tags: [theme, lipgloss, styles, tree, validation]

requires:
  - "13-02: cfg config.Config in app.New(); cfg.Theme field set from config/CLI"

provides:
  - "THEME-01: ThemeDefault() identical to pre-14 globals — no visual regression"
  - "THEME-02: ThemeMinimal() muted 256-color preset wired into tree via Options.Theme"
  - "THEME-03: ThemeHighContrast() 16-color ANSI bold preset wired into tree via Options.Theme"
  - "THEME-04: unknown theme name prints stderr warning and falls back to default"
  - "tui.Theme struct with Complete/Active/Pending/Failed/NowMarker/RefreshFlash/QuitPending/Highlight/EmptyFg/HelpBorder/HelpFg"
  - "tui.ThemeByName() lookup: known names return (theme, true); unknown returns (default, false)"
  - "tree.Options.Theme field: single authority for color selection in view.go"

affects:
  - "15-help-overlay: cfg.Theme ready; ThemeByName(cfg.Theme) available for display"

tech-stack:
  added: []
  patterns:
    - "Theme struct bundles all lipgloss styles — single allocation, passed via tree.Options"
    - "themeFor(opts) helper: returns opts.Theme if set, else ThemeDefault() — safe zero-value handling"
    - "lipgloss.TerminalColor interface for color fields — accepts both AdaptiveColor and Color"
    - "ThemeByName(name) (Theme, bool) — ok=false signals unknown name at call site for warning"

key-files:
  created:
    - "internal/tui/theme_test.go"
    - ".planning/phases/14-theme-system/14-01-PLAN.md"
  modified:
    - "internal/tui/styles.go"
    - "internal/tui/tree/model.go"
    - "internal/tui/tree/view.go"
    - "internal/tui/tree/model_test.go"
    - "internal/tui/app/model.go"
    - "cmd/gsd-watch/main.go"

key-decisions:
  - "Theme struct uses lipgloss.TerminalColor (interface) for color fields — lipgloss.AdaptiveColor satisfies it, enabling adaptive dark/light colors for default/minimal themes"
  - "themeFor(opts) zero-check uses opts.Theme.Pending.GetForeground() != nil — Pending is always set in every ThemeX() constructor"
  - "tree.Options.Theme zero value resolves to ThemeDefault() at render time — existing tests need no Options.Theme setup"
  - "THEME-04 validation in main.go placed after flag.Visit block — ensures CLI overrides are applied before validation"
  - "cfg.Theme reset to empty string on unknown name — ThemeByName('') returns default, consistent with omitted config key"

requirements-completed: [THEME-01, THEME-02, THEME-03, THEME-04]

duration: 6min
completed: 2026-03-27
---

# Phase 14 Plan 01: Theme System — Theme Struct, Presets, Call-site Migration, Validation Summary

**Theme struct with three named presets (default/minimal/high-contrast) in styles.go; tree/view.go migrated to use Theme via Options; unknown theme name warns on stderr and falls back to default**

## Performance

- **Duration:** ~6 min
- **Completed:** 2026-03-27
- **Tasks:** 3
- **Files modified:** 6 (internal/tui/styles.go, tree/model.go, tree/view.go, tree/model_test.go, app/model.go, cmd/gsd-watch/main.go)
- **Files created:** 2 (internal/tui/theme_test.go, .planning/phases/14-theme-system/14-01-PLAN.md)

## Accomplishments

- `tui.Theme` struct with 11 fields covering all color needs in tree rendering
- `ThemeDefault()`, `ThemeMinimal()`, `ThemeHighContrast()` constructors covering all three THEME requirements
- `ThemeByName(name string) (Theme, bool)` lookup for name dispatch and unknown detection
- `tree.Options.Theme tui.Theme` field; `themeFor(opts)` helper handles zero-value safely
- All `tui.PendingStyle`, `tui.NowMarkerStyle`, `tui.ColorGray`, `highlightStyle` references in `view.go` replaced with theme fields
- `RenderArchiveRow`, `RenderArchiveZone` updated to accept `tui.Theme` parameter
- `app/model.go` wires `tui.ThemeByName(cfg.Theme)` into `tree.SetOptions()`
- `main.go` validates theme name; unknown names print `gsd-watch: unknown theme "X", using default` to stderr
- 5 new theme tests + all existing tests passing (no regressions)

## Task Commits

1. **Task 1: Add Theme struct and three presets** - `6991d98` (feat)
2. **Task 2: Migrate tree/view.go to use Theme via Options** - `1fd2eda` (feat)
3. **Task 3: Validate theme name at startup** - `efcb114` (feat)

## Files Created/Modified

- `internal/tui/styles.go` — Theme struct, ThemeDefault/Minimal/HighContrast(), ThemeByName(); existing globals preserved
- `internal/tui/theme_test.go` — TestThemeByName_Known, TestThemeByName_Unknown, TestThemeDefault_NotNil, TestThemeMinimal_NotNil, TestThemeHighContrast_NotNil
- `internal/tui/tree/model.go` — Options.Theme field; themeFor() helper
- `internal/tui/tree/view.go` — All style references use th.Pending/th.Highlight/th.NowMarker/th.EmptyFg; archive functions accept tui.Theme
- `internal/tui/tree/model_test.go` — Updated RenderArchiveRow/RenderArchiveZone call sites; added tui import
- `internal/tui/app/model.go` — ThemeByName(cfg.Theme) wired into tree.SetOptions()
- `cmd/gsd-watch/main.go` — tui import; theme validation block after flag.Visit

## Decisions Made

- `lipgloss.TerminalColor` interface for `EmptyFg`/`HelpBorder`/`HelpFg` fields — `lipgloss.AdaptiveColor` (used by default/minimal themes) and `lipgloss.Color` (used by high-contrast) both satisfy the interface
- `themeFor(opts)` zero-check: `opts.Theme.Pending.GetForeground() != nil` — Pending is always set in ThemeX() constructors; zero lipgloss.Style has nil foreground
- `cfg.Theme` reset to `""` on unknown name (not `"default"`) — `ThemeByName("")` returns default theme, consistent with omitted config key behavior
- Package-level style vars (`PendingStyle`, etc.) kept intact — `footer/model.go`, `header/model.go`, and existing tests reference them directly; only `tree/view.go` migrated

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed model_test.go call sites after RenderArchiveRow/RenderArchiveZone signature change**
- **Found during:** Task 2 verification (`go test ./...`)
- **Issue:** `internal/tui/tree/model_test.go` called `tree.RenderArchiveRow(am, false)` and `tree.RenderArchiveZone(archives, 80, false)` — no longer valid after theme parameter was added
- **Fix:** Added `tui.ThemeDefault()` as the third argument to each call; added `tui` import
- **Files modified:** `internal/tui/tree/model_test.go`
- **Commit:** `1fd2eda`

## Known Stubs

None — all three themes are fully wired. Phase 15 will display the active theme name in the help overlay.

## Self-Check: PASSED

- `internal/tui/styles.go`: FOUND
- `internal/tui/theme_test.go`: FOUND
- `internal/tui/tree/model.go`: FOUND
- `internal/tui/tree/view.go`: FOUND
- `internal/tui/app/model.go`: FOUND
- `cmd/gsd-watch/main.go`: FOUND
- Commit `6991d98` (feat(14-01): add Theme struct...): FOUND
- Commit `1fd2eda` (feat(14-01): migrate tree/view.go...): FOUND
- Commit `efcb114` (feat(14-01): validate theme name...): FOUND

---
*Phase: 14-theme-system*
*Completed: 2026-03-27*
