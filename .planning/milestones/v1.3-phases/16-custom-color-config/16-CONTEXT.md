# Phase 16: Custom Color Config - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Users can set individual hex color overrides in config.toml under `[theme]`, with the active `preset` as the base. Unoverridden fields keep the preset value. Only the 5 visible status-tree colors are exposed. The app always starts — invalid values emit a warning and fall back to the preset color.

This phase also renames the config key `theme` → `preset` (clean break, no migration shim).

</domain>

<decisions>
## Implementation Decisions

### TOML Schema Change
- **D-01:** Rename config key `theme` → `preset`. Existing configs with `theme = "..."` get an unknown-key warning (CFG-03 path) — no migration alias. This frees `[theme]` as a section name.
- **D-02:** Color overrides live in `[theme]` table in config.toml. `Config` struct gains a nested `ThemeColors` struct with TOML tag `theme`, holding 5 optional string fields.

### Exposed Fields (5 status-tree colors)
- **D-03:** The following fields are user-overrideable via `[theme]` keys. TOML key names use snake_case matching the struct field names:
  - `complete` — complete status color
  - `active` — in-progress/active status color
  - `pending` — pending/default status color
  - `failed` — failed status color
  - `now_marker` — current-phase arrow color (NowMarker)

  Transient UI fields (`RefreshFlash`, `QuitPending`, `Highlight`, `EmptyFg`, `HelpBorder`, `HelpFg`) are NOT exposed — user does not need to tweak them.

### Color Value Format
- **D-04:** Only `#RRGGBB` hex strings are accepted (e.g., `"#00cc00"`). ANSI index strings are not supported. Validation: 7-character string starting with `#`.

### Invalid Color Handling
- **D-05:** Invalid hex strings emit a stderr warning naming the field and the bad value, then fall back to the preset's color for that field. App starts normally. Consistent with CFG-03 unknown-key behavior — never fatal for a color override error.

### Apply Logic
- **D-06:** `ThemeByName(preset)` resolves the base `Theme`. Then each non-empty `ThemeColors` field is validated and, if valid, overrides the corresponding `Theme` field style with `lipgloss.NewStyle().Foreground(lipgloss.Color(hexValue))`. Applied in `app.New()` after `ThemeByName` returns (or in a new `ApplyColorOverrides(theme, overrides)` helper in `styles.go`).

### Claude's Discretion
- Whether `ThemeColors` is a named struct or a `map[string]string`
- Whether the apply logic lives as a method on `ThemeColors`, a free function in `styles.go`, or inline in `app.New()`
- Exact stderr warning message format (must name the field and bad value)
- Whether short hex (`#RGB`) is silently rejected or expanded to `#RRGGBB`

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — §Future Requirements (Config Extensions) for `[colors]` table intent; §Out of Scope for header/footer full theme coverage deferral
- `.planning/ROADMAP.md` §Phase 16 — scope boundary

### Prior Phase Context
- `.planning/phases/13-config-infrastructure/13-CONTEXT.md` — D-04: BurntSushi/toml; D-05: flag.Visit override pattern; D-06: config file path
- `.planning/phases/14-theme-system/14-CONTEXT.md` — D-04: Theme struct shape; D-06: ThemeByName() function signature
- `.planning/phases/15-help-overlay-config-hint/15-CONTEXT.md` — D-03: helpView signature (uses cfg.Theme — will need updating to cfg.Preset after rename)

### Out-of-Scope Constraints
- `.planning/REQUIREMENTS.md` §Out of Scope — header/footer full theme coverage deferred to v1.4+
- Multiple named custom profiles / `[profiles.X]` switching — deferred idea (see below)

</canonical_refs>

<code_context>
## Existing Code Insights

### Key Files
- `internal/config/load.go` — `Config` struct with `Emoji bool`, `Theme string` (rename to `Preset`); `Defaults()`, `Load()`, `UnknownKeysError`; `[theme]` table needs a new nested struct field
- `internal/tui/styles.go` — `Theme` struct, `ThemeByName()`, three preset constructors; new `ApplyColorOverrides()` helper lands here
- `cmd/gsd-watch/main.go` — calls `ThemeByName(cfg.Theme)` (update to `cfg.Preset`); handles unknown-theme warning
- `internal/tui/app/model.go` — `app.New(events, cfg)` where theme is resolved; color overrides applied here or via helper
- `internal/tui/app/model.go:267` — `helpView` uses `cfg.Theme` for display (update to `cfg.Preset`)

### Downstream Impact of Rename
- Every reference to `cfg.Theme` across the codebase needs updating to `cfg.Preset`
- `Defaults()` returns `Theme: ""` → becomes `Preset: ""`
- Phase 15 CONTEXT D-02: "display `default` when cfg.Theme is `""`" → same logic, just field rename

</code_context>

<specifics>
## Specific Ideas

- Example valid config after Phase 16:
  ```toml
  preset = "minimal"
  emoji  = true

  [theme]
  complete   = "#00cc00"
  failed     = "#cc0000"
  now_marker = "#ffaa00"
  ```
- `complete` and `active` are independent fields even though both default to green in most presets — user can set them differently.
- The `[theme]` section with no keys is valid TOML and should produce no warnings.
- Short hex (`#RGB`) behavior is Claude's discretion — safest to reject with a warning rather than expand silently.

</specifics>

<deferred>
## Deferred Ideas

- **Multiple named custom profiles** — User asked about switching between different named color sets (e.g., `[profiles.dark]` and `[profiles.light]`). Great poweruser feature; scope for a future phase (v1.4+).
- **`theme = "custom"` preset** — REQUIREMENTS.md future note mentions a preset that fully defers to the `[colors]` table. Phase 16 covers partial overrides; `custom` (full override) is a natural follow-on if there's demand.

</deferred>

---

*Phase: 16-custom-color-config*
*Context gathered: 2026-03-27*
