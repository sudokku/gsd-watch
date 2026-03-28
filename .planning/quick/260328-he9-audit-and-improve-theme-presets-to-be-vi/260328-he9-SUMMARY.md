---
quick_task: 260328-he9
plan: 02
status: complete
completed_date: "2026-03-28"
duration_minutes: 22
subsystem: tui/themes
tags: [theme, badge, styling, visual-differentiation, structural-chrome, connectors, progress-bar, separator]
dependency_graph:
  requires: []
  provides: [THEME-VISUAL-DIFFERENTIATION, THEME-STRUCTURAL-CHROME]
  affects:
    - internal/tui/styles.go
    - internal/tui/tree/view.go
    - internal/tui/app/model.go
    - internal/tui/header/model.go
    - internal/tui/footer/model.go
tech_stack:
  added: []
  patterns:
    - BadgeStyle map field on Theme (plan 01)
    - per-theme badge palette (plan 01)
    - theme-parameterized BadgeString (plan 01)
    - SeparatorFg/ProgressFilled/ProgressEmpty/ConnectorFg/ExpandIndicatorFg/ArchiveSeparatorFg on Theme (plan 02)
    - InProgressStyle/HeaderNameStyle on Theme (plan 02)
    - SetTheme() on HeaderModel and FooterModel (plan 02)
key_files:
  created: []
  modified:
    - internal/tui/styles.go
    - internal/tui/styles_test.go
    - internal/tui/theme_test.go
    - internal/tui/tree/view.go
    - internal/tui/tree/model_test.go
    - internal/tui/app/model.go
    - internal/tui/header/model.go
    - internal/tui/footer/model.go
decisions:
  - "BadgeStyle is map[string]lipgloss.Style on Theme; keys are badge names matching BadgeString switch cases"
  - "BadgeString gains Theme param; emoji mode is unchanged; noEmoji mode applies theme.BadgeStyle[badge] if present"
  - "Zero Theme{} (nil BadgeStyle map) falls back to plain text — backward safe"
  - "helpView gains Theme param; badge legend in noEmoji mode uses BadgeString so overlay colors match active theme"
  - "high-contrast Highlight uses Reverse(true) — bg/fg reversal is the single biggest visual differentiator"
  - "HeaderModel and FooterModel gain SetTheme(); New() defaults to ThemeDefault() for backward safety"
  - "ExpandIndicatorFg on default theme is lipgloss.NoColor{} — terminal default, zero-diff from before"
  - "RenderArchiveSeparator gains theme param to apply ArchiveSeparatorFg"
metrics:
  duration: 22 min
  completed_date: "2026-03-28"
  tasks_completed: 5
  files_modified: 8
---

# Quick Task 260328-he9: Audit and Improve Theme Presets to Be Visually Distinct — Summary

**One-liner:** Nine rendering sites that bypassed the Theme struct now thread structural chrome colors (separators, progress bar, connectors, expand arrows, archive separator, in-progress icon, header name) through all three theme presets, widening the visual delta between default, minimal, and high-contrast.

## What Was Built

### Plan 01 — Badge styling (prior session)

Three theme presets produce distinct visual output for phase stage badges in noEmoji mode. See plan 01 details in the commits section.

### Plan 02 — Structural chrome fields (this session)

Eight new fields were added to the `Theme` struct and wired into all render sites that previously hardcoded colors:

**ThemeDefault** — preserves existing visual feel, makes it explicit:
- SeparatorFg: ColorGray (═ and ─ lines stay gray)
- ProgressFilled: ColorGreen, ProgressEmpty: ColorGray (unchanged)
- ConnectorFg: ColorGray (├──/└──/│ stay gray)
- ExpandIndicatorFg: lipgloss.NoColor{} (terminal default — no change)
- ArchiveSeparatorFg: ColorGray
- InProgressStyle: plain style (no change)
- HeaderNameStyle: bold (no change)

**ThemeMinimal** — chrome recedes, structure disappears, only content remains:
- SeparatorFg: color 240 — very dark, barely visible separators
- ProgressFilled: color 243, ProgressEmpty: color 238 — nearly flat bar
- ConnectorFg: color 240 — tree lines dim to near-invisible
- ExpandIndicatorFg: color 243 — arrows dimmed
- ArchiveSeparatorFg: color 238 — archive section recedes further
- InProgressStyle: muted foreground (color 245) — not green
- HeaderNameStyle: no bold — reduces visual weight of project name

**ThemeHighContrast** — punchy, max brightness, reverse-video cursor:
- Highlight: Reverse(true) — bg/fg reversal for cursor row (biggest single differentiator)
- SeparatorFg: color 7 (white) — maximum visibility separators
- ProgressFilled: color 2 (green), ProgressEmpty: color 0 (black) — high-contrast bar
- ConnectorFg: color 7 (white) — white connectors pop on dark terminals
- ExpandIndicatorFg: color 3 (yellow) — arrows pop in yellow
- ArchiveSeparatorFg: color 7 (white)
- InProgressStyle: Bold + yellow foreground (color 3)
- HeaderNameStyle: bold (same as default, keeps strong project name)

## Changes Made

### `internal/tui/styles.go`
- Added 8 new fields to `Theme` struct: `SeparatorFg`, `ProgressFilled`, `ProgressEmpty`, `ConnectorFg`, `ExpandIndicatorFg`, `ArchiveSeparatorFg`, `InProgressStyle`, `HeaderNameStyle`
- All three theme constructors (`ThemeDefault`, `ThemeMinimal`, `ThemeHighContrast`) populated with per-theme values
- `ThemeHighContrast`: `Highlight` changed to `lipgloss.NewStyle().Reverse(true)`
- `StatusIcon`: in_progress branches now use `theme.InProgressStyle.Render()` instead of bare string

### `internal/tui/header/model.go`
- Added `theme tui.Theme` field to `HeaderModel`
- Added `SetTheme(th tui.Theme) HeaderModel` method
- `New()` defaults to `ThemeDefault()`
- `View()`: project name uses `theme.HeaderNameStyle`, separator uses `theme.SeparatorFg`
- `progressBar()` gains `th tui.Theme` param; uses `theme.ProgressFilled` and `theme.ProgressEmpty`

### `internal/tui/footer/model.go`
- Added `theme tui.Theme` field to `FooterModel`
- Added `SetTheme(th tui.Theme) FooterModel` method
- `New()` defaults to `ThemeDefault()`
- `View()`: footer ─ separator uses `theme.SeparatorFg`; spinner uses `theme.RefreshFlash`

### `internal/tui/app/model.go`
- `New()`: calls `header.New(...).SetTheme(th)` and `footer.New(...).SetTheme(th)` to propagate resolved theme
- `helpView()`: inner/box styles use `theme.HelpFg` and `theme.HelpBorder` (previously hardcoded `tui.ColorGray`)

### `internal/tui/tree/view.go`
- `RowPhase`: expand indicator (`▶`/`▼`) rendered with `theme.ExpandIndicatorFg`; cursor row uses raw indicator for clean Highlight wrapping
- `RowPlan`: connector (`├──`/`└──`) and continuation (`│`) rendered with `theme.ConnectorFg`; dimmed rows override with Pending color
- `RowQuickSection`: expand indicator rendered with `theme.ExpandIndicatorFg`
- `RowQuickTask`: connector and continuation rendered with `theme.ConnectorFg`
- `RenderArchiveSeparator(width, th)`: gains theme param; applies `theme.ArchiveSeparatorFg`
- `RenderArchiveZone`: passes theme through to `RenderArchiveSeparator`

### Test files updated
- `internal/tui/theme_test.go`: 4 new tests for structural fields
- `internal/tui/header/model_test.go`: 2 new tests for `SetTheme` + all presets
- `internal/tui/footer/model_test.go`: 1 new test for `SetTheme` + all presets
- `internal/tui/tree/model_test.go`: updated `RenderArchiveSeparator` call to new signature

## Commits

| # | Hash | Description |
|---|------|-------------|
| 1 | 692ac33 | feat(quick-260328-he9-01): add BadgeStyle to Theme; update BadgeString signature |
| 2 | 6f9daed | feat(quick-260328-he9-01): wire theme badge styling into help overlay |
| 3 | d7e8903 | docs(quick-260328-he9-01): add SUMMARY.md |
| 4 | e1e63dc | feat(quick-260328-he9-02): add structural chrome fields to Theme struct; update all theme constructors |
| 5 | 6f7d5e4 | feat(quick-260328-he9-02): wire all theme fields into render sites |
| 6 | cbde3fa | test(quick-260328-he9-02): add tests for new Theme structural chrome fields |

## Test Results

```
ok  github.com/radu/gsd-watch/internal/tui          4.116s
ok  github.com/radu/gsd-watch/internal/tui/app      (cached)
ok  github.com/radu/gsd-watch/internal/tui/footer   0.665s
ok  github.com/radu/gsd-watch/internal/tui/header   1.526s
ok  github.com/radu/gsd-watch/internal/tui/tree     (cached)
```

## Deviations from Plan

### Plan 01 (prior session)

**1. [Rule 1 - Bug] Renamed `dim` var in ThemeMinimal to use correct color 248**
- Changed Active from color 245 to 248 per plan spec

**2. [Rule 3 - Blocking] Fixed lipgloss.WithColorProfile API mismatch**
- Used `r.SetColorProfile(termenv.ANSI256)` instead of non-existent `lipgloss.WithColorProfile`

### Plan 02 (this session)

**1. [Rule 1 - Bug] Width calculation for expand indicator after color styling**
- Styled indicator (`lipgloss.NewStyle().Foreground(...).Render("▶ ")`) changes string bytes but not display width
- Split into `rawIndicator` (for width math + cursor Highlight wrapping) and `expandIndicator` (for styled normal rendering) — width calculation stays correct

**2. [Rule 2 - Missing critical] RenderArchiveSeparator call site in tree test updated**
- Signature changed (gains theme param), existing test broke at compile time
- Updated call to `RenderArchiveSeparator(80, tui.ThemeDefault())` — included in Task B commit

## Known Stubs

None — all structural chrome colors are fully wired end-to-end through themes.
