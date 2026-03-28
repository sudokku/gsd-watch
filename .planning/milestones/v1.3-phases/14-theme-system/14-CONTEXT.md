# Phase 14: Theme System - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

New `Theme` struct in `internal/tui/styles.go` with three named presets (`default`, `minimal`, `high-contrast`). A `ResolveTheme(name string) (Theme, bool)` helper validates theme names and returns the resolved struct. `tree.Options` gains a `Theme tui.Theme` field. All render call sites in `tree/view.go` migrate from package-level style vars (`tui.PendingStyle` etc.) to `opts.Theme.*` fields. Exported archive functions (`RenderArchiveRow`, `RenderArchiveZone`) gain a `theme tui.Theme` param. Unknown theme name prints a stderr warning in `app.New()` and falls back to default — no crash.

Header and footer theming is explicitly **NOT in scope** — deferred to v1.4+ per REQUIREMENTS.md.

</domain>

<decisions>
## Implementation Decisions

### Palette: `minimal` preset
- **D-01:** `minimal` is gray-only — all status colors (complete, active, failed, now-marker) render in the same gray as `PendingStyle`. No green, red, or amber. Content (phase/plan names) stays white (uncolored). The archive zone uses the same gray — no special-casing needed since it already uses PendingStyle in default.

### Palette: `high-contrast` preset
- **D-02:** `high-contrast` uses standard 8-color ANSI palette (indices 1–8): green=2, red=1, yellow/amber=3, white/pending=7. All status colors get `Bold(true)` applied. No `lipgloss.AdaptiveColor` — fixed foreground indices only. This ensures compatibility with SSH and degraded terminals.

### Palette: `default` preset
- **D-03:** `default` reproduces the current v1.2 appearance exactly — `lipgloss.AdaptiveColor` for all colors, same values as existing `ColorGreen/Red/Amber/Gray` vars. Zero visual regression (THEME-01).

### Theme Struct Shape
- **D-04:** `Theme` struct in `internal/tui/styles.go` holds pre-built `lipgloss.Style` fields:
  ```go
  type Theme struct {
      Complete  lipgloss.Style
      Active    lipgloss.Style
      Pending   lipgloss.Style
      Failed    lipgloss.Style
      NowMarker lipgloss.Style
      Highlight lipgloss.Style
  }
  ```
  Styles are constructed once when the theme is resolved. Call sites do `theme.Pending.Render(x)` — direct drop-in for `tui.PendingStyle.Render(x)`.

### Theme Propagation
- **D-05:** `tree.Options` gains a `Theme tui.Theme` field alongside existing `NoEmoji bool`. `app.New()` calls `tui.ResolveTheme(cfg.Theme)`, then passes the resolved theme via `t.SetOptions(tree.Options{NoEmoji: !cfg.Emoji, Theme: resolvedTheme})`. Consistent with the Phase 10 options pattern.

### Unknown-Theme Validation (THEME-04)
- **D-06:** `tui.ResolveTheme(name string) (Theme, bool)` lives in `styles.go`. Returns `(DefaultTheme(), false)` for any name not in the known set. `app.New()` calls it, and when `ok == false`, prints a stderr warning (e.g. `gsd-watch: unknown theme %q, falling back to default`) and proceeds. Phase 13 `config.Load()` code is untouched.

### Archive Function Signatures
- **D-07:** `RenderArchiveRow(am parser.ArchivedMilestone, noEmoji bool, theme tui.Theme)` — adds `theme` as a third param, consistent with the existing `noEmoji bool` param pattern. Same for `RenderArchiveZone(archives []parser.ArchivedMilestone, width int, noEmoji bool, theme tui.Theme)`. Test callers in `tree_test` package pass `tui.DefaultTheme()`.

### Claude's Discretion
- Name of the default-theme constructor (`DefaultTheme()` vs `NewDefaultTheme()` vs `Themes["default"]`)
- Whether the old package-level vars (`CompleteStyle`, `PendingStyle`, etc.) stay as aliases pointing at `DefaultTheme()` fields, or are removed
- `StatusIcon()` and `BadgeString()` in `styles.go` — whether they gain a `theme Theme` param or remain pure string functions (badge strings have no color, status icons delegate to styles)
- Exact stderr warning message format (must include the bad theme name per REQUIREMENTS.md implied behavior)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — THEME-01 through THEME-04 define all acceptance criteria for this phase
- `.planning/ROADMAP.md` §Phase 14 — success criteria and scope boundary

### Prior Phase Context
- `.planning/phases/13-config-infrastructure/13-CONTEXT.md` — D-01 through D-06: config package shape, `cfg.Theme = ""` means "default", Phase 14 owns validation

### Out-of-Scope Constraints
- `.planning/REQUIREMENTS.md` §Future Requirements — header/footer full theme coverage deferred to v1.4+
- `.planning/REQUIREMENTS.md` §Out of Scope — per-color `[colors]` TOML table deferred

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/tui/styles.go` — current color vars (`ColorGreen`, `ColorAmber`, `ColorRed`, `ColorGray`) and 7 named styles (`CompleteStyle`, `ActiveStyle`, `PendingStyle`, `FailedStyle`, `NowMarkerStyle`, `RefreshFlashStyle`, `QuitPendingStyle`); Theme struct and ResolveTheme() land here
- `internal/tui/tree/model.go` — `tree.Options{NoEmoji bool}` and `SetOptions()` pattern from Phase 10; `Theme tui.Theme` is a natural extension
- `internal/tui/tree/view.go` — all call sites to migrate: `tui.PendingStyle`, `tui.NowMarkerStyle`, `tui.StatusIcon()`, archive functions; exported functions need `theme tui.Theme` param
- `internal/config/load.go` — `Config.Theme string`; `""` is the default (Phase 13 decision D-03)
- `cmd/gsd-watch/main.go` — `app.New(events, cfg)` call site; Phase 14 adds `ResolveTheme` call + warning before `app.New()`
- `internal/tui/app/model.go` — `app.New(events chan tea.Msg, cfg config.Config)` signature (Phase 13); sets `t.SetOptions()` to propagate options

### Established Patterns
- `lipgloss.AdaptiveColor` for default theme (Phase 1 principle — keep for `default` preset only)
- `tree.Options` struct + `SetOptions()` method for propagating render options from app → tree (Phase 10)
- `noEmoji bool` param pattern on exported render functions (Phase 10)
- Best-effort defaults on any error — consistent throughout the codebase
- Package-level functions exported for testability in external `_test` packages (Phase 12)

### Integration Points
- `internal/tui/styles.go`: add `Theme` struct, `ResolveTheme()`, `DefaultTheme()`, three preset constructors
- `internal/tui/tree/model.go`: add `Theme tui.Theme` to `Options` struct
- `internal/tui/tree/view.go`: migrate all `tui.*Style` references to `t.opts.Theme.*`; update `RenderArchiveRow/Zone` signatures
- `cmd/gsd-watch/main.go`: call `tui.ResolveTheme(cfg.Theme)`, print warning if unknown, pass resolved theme via `app.New()` or via `SetOptions()` after init
- Test files in `tree_test` package: update `RenderArchiveRow/Zone` call sites to pass `tui.DefaultTheme()`

</code_context>

<specifics>
## Specific Ideas

- `default` preset must produce byte-for-byte identical output to v1.2 for the same input (THEME-01 zero regression requirement). Easiest to achieve by constructing it from the same `lipgloss.AdaptiveColor` values currently in `styles.go`.
- `minimal` gray: use the same `ColorGray = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}` already defined — Complete/Active/Failed/NowMarker all render with `PendingStyle` equivalent.
- `high-contrast` bold: all styles use `lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))` (no AdaptiveColor) for the green role, etc.
- `ResolveTheme("")` should return `(DefaultTheme(), true)` — empty string is the "default" sentinel from Phase 13 config defaults.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 14-theme-system*
*Context gathered: 2026-03-27*
