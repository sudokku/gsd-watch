# Phase 15: Help Overlay Config Hint - Context

**Gathered:** 2026-03-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Extend the existing `helpView` function in `internal/tui/app/model.go` to render a new "Config" section at the bottom of the `?` overlay. The section shows two lines: the tilde-abbreviated config file path and the active theme name. No new packages, no new files — purely additive changes to `helpView` and its call site in `View()`.

</domain>

<decisions>
## Implementation Decisions

### Overlay Section Placement
- **D-01:** New "Config" section appended after "Phase stages" and before "press q or esc to close". Consistent with the existing section pattern (Navigation / Tree / Quit / Phase stages / Config). Both lines (`Config:` and `Theme:`) live inside this section.

### Theme Name Display
- **D-02:** When `cfg.Theme` is `""` (the default sentinel from Phase 13), display `"default"` — the resolved canonical name. The caller normalises: `themeName := m.cfg.Theme; if themeName == "" { themeName = "default" }`. Never display a blank theme name.

### `helpView` Signature Extension
- **D-03:** Add two string params: `helpView(width int, noEmoji bool, configPath, themeName string)`. Keeps the function pure and testable — caller resolves both values before calling. Consistent with the `noEmoji bool` param extension pattern from Phase 10.

### Config Path Tilde Abbreviation
- **D-04:** Computed inline in `View()` before the `helpView` call:
  ```go
  home, _ := os.UserHomeDir()
  cfgPath := strings.Replace(config.DefaultPath(), home, "~", 1)
  ```
  No new helper function — one call site, no reuse needed elsewhere.

### Claude's Discretion
- Whether `config.DefaultPath()` already exists or needs to be added to `internal/config/` (or inline the path string directly using the same `os.UserHomeDir()` + join logic from Phase 13 D-06)
- Exact column alignment of `Config:` and `Theme:` label/value pairs (aligned or flush-left)
- Whether existing `helpView` tests need updating or new test cases are added

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — DISC-01, DISC-02 define acceptance criteria for this phase
- `.planning/ROADMAP.md` §Phase 15 — success criteria and scope boundary

### Prior Phase Context
- `.planning/phases/13-config-infrastructure/13-CONTEXT.md` — D-06: config path uses `os.UserHomeDir()` + manual join; `cfg.Theme = ""` means "default" sentinel
- `.planning/phases/14-theme-system/14-CONTEXT.md` — D-06: `ResolveTheme("")` returns `(DefaultTheme(), true)`; `""` resolves to "default" canonical name

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/tui/app/model.go:267` — `helpView(width int, noEmoji bool) string` — package-level function to extend; call site at line 339: `return helpView(m.width, !m.cfg.Emoji)`
- `internal/tui/model_test.go` — existing help overlay tests (check for coverage gaps after signature change)
- `internal/config/` — `Config.Theme string` field; `Defaults()` returns `Theme: ""`; config path logic lives here (or can be extracted)

### Established Patterns
- `noEmoji bool` param added to `helpView` in Phase 10 — same extension pattern for `configPath, themeName string`
- `m.cfg` is a `config.Config` on the Model — accessible in `View()` where `helpView` is called
- `os.UserHomeDir()` + manual join used in Phase 13 `config.Load()` for path resolution

### Integration Points
- `internal/tui/app/model.go`: `helpView` signature (add two string params), `View()` call site (compute `cfgPath` + `themeName` inline, pass to `helpView`), `helpText` string (add Config section)
- `internal/tui/model_test.go`: update `helpView` call sites to pass new params; add assertions for Config section content

</code_context>

<specifics>
## Specific Ideas

- Layout of the Config section in the overlay:
  ```
  Config
  Config:  ~/.config/gsd-watch/config.toml
  Theme:   default
  ```
- The tilde path (`~/...`) is always shown regardless of whether the file exists on disk (DISC-01 requirement)
- `themeName` is the canonical name — `"default"` not `""` — even when the user has never set a theme

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 15-help-overlay-config-hint*
*Context gathered: 2026-03-27*
