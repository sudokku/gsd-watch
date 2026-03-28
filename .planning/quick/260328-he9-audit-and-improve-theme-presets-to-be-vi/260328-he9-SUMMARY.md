---
quick_task: 260328-he9
plan: 01
status: awaiting-verification
completed_date: "2026-03-28"
duration_minutes: 6
subsystem: tui/themes
tags: [theme, badge, styling, visual-differentiation]
dependency_graph:
  requires: []
  provides: [THEME-VISUAL-DIFFERENTIATION]
  affects: [internal/tui/styles.go, internal/tui/tree/view.go, internal/tui/app/model.go]
tech_stack:
  added: []
  patterns: [BadgeStyle map field on Theme, per-theme badge palette, theme-parameterized BadgeString]
key_files:
  created: []
  modified:
    - internal/tui/styles.go
    - internal/tui/styles_test.go
    - internal/tui/theme_test.go
    - internal/tui/tree/view.go
    - internal/tui/app/model.go
decisions:
  - "BadgeStyle is map[string]lipgloss.Style on Theme; keys are badge names matching BadgeString switch cases"
  - "BadgeString gains Theme param; emoji mode is unchanged (no styling); noEmoji mode applies theme.BadgeStyle[badge] if present"
  - "Zero Theme{} (nil BadgeStyle map) falls back to plain text — backward safe for any callers using zero-value Theme"
  - "helpView gains Theme param; badge legend in noEmoji mode uses BadgeString so overlay colors match active theme"
  - "View() resolves theme via ThemeByName + ApplyColorOverrides with io.Discard before calling helpView"
metrics:
  duration: 6 min
  completed_date: "2026-03-28"
  tasks_completed: 2
  files_modified: 5
---

# Quick Task 260328-he9: Audit and Improve Theme Presets to Be Visually Distinct — Summary

**One-liner:** Per-theme badge color palettes (256-color default, muted minimal, bold+bright high-contrast) via BadgeStyle map on Theme; BadgeString gains Theme param applied at all call sites.

## What Was Built

Three theme presets now produce meaningfully distinct visual output for phase stage badges:

**ThemeDefault** — 256-color ANSI palette with distinct colors per badge category:
- Discussed/Researched: Cyan (36)
- UI Spec: Blue (33)
- Planned: Blue/Indigo (69)
- Executed: Magenta (133)
- Verified: Green (34)
- UAT: Yellow (178)

**ThemeMinimal** — Uniform muted tone (color 243) for all badges; Active style uses color 248 (slightly brighter) so active phases still stand out. Content-first aesthetic preserved.

**ThemeHighContrast** — Bold + bright 16-color ANSI only:
- Discussed/Researched: Bright Cyan (14)
- UI Spec/Planned: Bright Blue (12)
- Executed: Bright Magenta (13)
- Verified: Bright Green (10)
- UAT: Bright Yellow (11)
- Active style gets Bold+Underline; Pending uses Bright White (15)

## Changes Made

### `internal/tui/styles.go`
- Added `BadgeStyle map[string]lipgloss.Style` field to `Theme` struct
- `ThemeDefault()`: populated BadgeStyle with distinct 256-color palette
- `ThemeMinimal()`: populated BadgeStyle with uniform muted tone; Active uses color 248 (was 245)
- `ThemeHighContrast()`: populated BadgeStyle with bold+bright 16-color palette; Active gets Underline; Pending uses Bright White (15)
- `BadgeString()` signature changed from `(badge, noEmoji)` to `(badge, noEmoji, theme)` — applies `theme.BadgeStyle[badge]` in noEmoji mode; emoji mode unchanged

### `internal/tui/tree/view.go`
- Updated both `BadgeString` call sites to pass `th` (already in scope from `themeFor(t.opts)`)

### `internal/tui/app/model.go`
- `helpView()` gains `theme tui.Theme` parameter
- Badge legend in noEmoji mode uses `tui.BadgeString` calls instead of hardcoded strings
- `View()` resolves theme before calling `helpView`; uses `io.Discard` to suppress color-override warnings in render path
- Added `io` import

### `internal/tui/styles_test.go`
- All `BadgeString` calls updated to pass `tui.ThemeDefault()` or `tui.Theme{}`
- `TestBadgeString_NoEmoji` now uses `tui.Theme{}` (nil BadgeStyle) for exact plain-text matching

### `internal/tui/theme_test.go`
- Added `TestBadgeStyle_DefaultDistinct`: verifies disc/exec/vrfy produce different ANSI output in default theme
- Added `TestBadgeStyle_HighContrastBold`: verifies all HC badge styles produce non-plain output (bold+color)
- Added `TestBadgeStyle_ThemesDiffer`: verifies same badge produces different output across all 3 themes
- Added `TestBadgeString_EmojiNoThemeChange`: verifies emoji mode is theme-invariant

## Test Results

```
ok  github.com/radu/gsd-watch/internal/tui          (all badge/theme tests pass)
ok  github.com/radu/gsd-watch/internal/tui/app      (no regressions)
ok  github.com/radu/gsd-watch/internal/tui/tree     (no regressions)
go vet ./...                                         (clean)
```

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| 1 (TDD) | 692ac33 | feat: add BadgeStyle to Theme; update BadgeString signature |
| 2 | 6f9daed | feat: wire theme badge styling into help overlay |

## Deviations from Plan

### Auto-applied

**1. [Rule 1 - Bug] Renamed `dim` var in ThemeMinimal to use correct color 248**
- **Found during:** Task 1 implementation
- **Issue:** Original `dim` was color 245 for Active. Plan spec said Active should use 248 for slightly-brighter differentiation from Complete (243).
- **Fix:** Changed `dim` to lipgloss.Color("248") directly; renamed variable to avoid confusion with `muted` (243)
- **Files modified:** internal/tui/styles.go

**2. [Rule 3 - Blocking] Fixed lipgloss.WithColorProfile API mismatch**
- **Found during:** Task 1 TDD (RED phase compile errors)
- **Issue:** lipgloss v1.1.0 does not export `WithColorProfile` or color profile constants directly — the plan's test snippet used a newer API. Correct API is `r.SetColorProfile(termenv.ANSI256)`.
- **Fix:** Updated test helper to use `lipgloss.NewRenderer(nil)` + `r.SetColorProfile(termenv.ANSI256)` from `github.com/muesli/termenv`
- **Files modified:** internal/tui/theme_test.go

## Known Stubs

None — all badge styles are fully wired and applied at render time.

## Status

Awaiting human verification (checkpoint:human-verify). Build instructions:

```bash
cd /Users/radu/Developer/gsd-watch && go build -o gsd-watch ./cmd/gsd-watch
./gsd-watch                      # default theme — distinct colored badges
./gsd-watch --theme minimal      # muted/gray badges
./gsd-watch --theme high-contrast # bold+bright badges
```
