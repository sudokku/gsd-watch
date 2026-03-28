# Phase 14: Theme System - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-27
**Phase:** 14-theme-system
**Areas discussed:** Palette design, Unknown-theme validation, Theme struct shape, Archive function signatures

---

## Palette design

### Minimal appearance

| Option | Description | Selected |
|--------|-------------|----------|
| Gray-only | Everything in gray/dim — no green/red/amber. Content stays white. | ✓ |
| Muted tints | Keep green/red/amber but desaturated. Colors still signal meaning. | |
| Monochrome with bold active | All gray except cursor row (bold white). No color at all. | |

**User's choice:** Gray-only
**Notes:** Archive zone uses same gray (already PendingStyle in default — no special-casing needed).

---

### High-contrast palette

| Option | Description | Selected |
|--------|-------------|----------|
| Standard 8 (1–8) | ANSI 1–8: green=2, red=1, yellow=3, gray=7. Safe everywhere. | ✓ |
| Bright 8 (9–15) | Bright variants: bright-green=10, bright-red=9. More vivid. | |
| Both + bold | Standard 1–8, all bold. Bold compensates for dim rendering. | |

**User's choice:** Standard 8 (1–8)
**Notes:** All status colors get Bold(true) applied. No AdaptiveColor — fixed ANSI indices.

---

### Archive zone in minimal

| Option | Description | Selected |
|--------|-------------|----------|
| Same gray as everything | Archive already uses PendingStyle (gray) — consistent with gray-only intent. | ✓ |
| Slightly dimmer | Dimmer gray (dim=true) to de-emphasize history even more. | |

**User's choice:** Same gray as everything

---

## Unknown-theme validation

| Option | Description | Selected |
|--------|-------------|----------|
| theme.Resolve() in styles.go | `ResolveTheme(name string) (Theme, bool)` in styles.go. app.New() calls it, prints warning. Phase 13 untouched. | ✓ |
| Directly in app.New() | Inline switch/case in app.New(). No new exported function. | |
| In config.Load() | Extend Phase 13 config package. Cross-package dependency concern. | |

**User's choice:** `theme.Resolve()` in styles.go
**Notes:** `ResolveTheme("")` returns `(DefaultTheme(), true)` — empty string is the default sentinel.

---

## Theme struct shape

### Struct contents

| Option | Description | Selected |
|--------|-------------|----------|
| Pre-built lipgloss.Style fields | `Complete, Active, Pending, Failed, NowMarker, Highlight lipgloss.Style`. Styles built once. | ✓ |
| Raw color values | Theme holds colors; styles built at call sites. More flexible but more changes. | |

**User's choice:** Pre-built lipgloss.Style fields

---

### Propagation path

| Option | Description | Selected |
|--------|-------------|----------|
| Add Theme to tree.Options | `tree.Options{NoEmoji bool, Theme tui.Theme}` — natural extension of Phase 10 pattern. | ✓ |
| Pass Theme as param to View() | `tree.View(width, height, theme)` — all call sites need updating. | |
| Store in app model, pass to helpers | app/model.go stores Theme, passes via setter. | |

**User's choice:** Add Theme to tree.Options

---

## Archive function signatures

| Option | Description | Selected |
|--------|-------------|----------|
| Add Theme param | `RenderArchiveRow(am, noEmoji bool, theme tui.Theme)` — consistent with existing pattern. | ✓ |
| Bundle into RenderOpts struct | `RenderArchiveRow(am, opts tui.RenderOpts)` — new exported type, cleaner for future. | |

**User's choice:** Add Theme param
**Notes:** Test callers in `tree_test` package pass `tui.DefaultTheme()`.

---

## Claude's Discretion

- Name of default-theme constructor (`DefaultTheme()` vs alternatives)
- Whether old package-level style vars stay as aliases or are removed
- `StatusIcon()` and `BadgeString()` theme param decision
- Exact stderr warning message format

## Deferred Ideas

None — discussion stayed within phase scope.
