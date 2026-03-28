# Phase 13: Config Infrastructure - Context

**Gathered:** 2026-03-26
**Status:** Ready for planning

<domain>
## Phase Boundary

New `internal/config/` package that loads `~/.config/gsd-watch/config.toml` using BurntSushi/toml. Implements three behaviors: silent defaults (missing file), fatal error with path (malformed TOML), and stderr warning (unknown keys). Adds `--theme <name>` flag alongside existing `--no-emoji`. CLI flags always override config via `flag.Visit`. Resolved config propagates to `app.New()` as a `config.Config` struct.

Theme validation is NOT in scope — Phase 13 stores and passes theme strings through. Phase 14 owns validation.

</domain>

<decisions>
## Implementation Decisions

### Config Propagation
- **D-01:** `app.New()` signature changes from `app.New(events, noEmoji bool)` to `app.New(events chan tea.Msg, cfg config.Config)` — Phase 14 adds fields to `config.Config` without touching the signature again
- **D-02:** `main.go` resolves config, applies flag overrides via `flag.Visit`, then passes the final `config.Config` to `app.New()` — config package stays out of the app package

### `--theme` Flag in Phase 13
- **D-03:** Phase 13 adds `--theme <name>` to `main.go` and stores whatever string is provided in `cfg.Theme` — no validation, no warning for unknown names. Phase 14 owns validation and the unknown-theme warning (CFG-05 is about override precedence only)

### TOML Library
- **D-04:** Use `github.com/BurntSushi/toml` — `toml.MetaData.Undecoded()` provides the unknown key list for CFG-03 stderr warning with zero extra logic

### Flag Override Mechanism
- **D-05:** `flag.Visit` iterates over explicitly-set flags after `flag.Parse()`. If `--no-emoji` was set, override `cfg.Emoji = false`. If `--theme` was set, override `cfg.Theme = <value>`. Config file values apply only when the corresponding flag was NOT explicitly passed.

### Config File Path
- **D-06:** `os.UserHomeDir()` + manual XDG join (`~/.config/gsd-watch/config.toml`) — `os.UserConfigDir()` is excluded per REQUIREMENTS.md (returns wrong path on macOS)

### Claude's Discretion
- Internal structure of `internal/config/` package (single file vs multiple)
- `config.Config` field names and TOML tag names (recommend: `Emoji bool \`toml:"emoji"\``, `Theme string \`toml:"theme"\``)
- Fatal error message format for malformed TOML (must include file path per CFG-02)
- Unknown key warning format for stderr (must name the key per CFG-03)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — CFG-01 through CFG-05 define all acceptance criteria for this phase
- `.planning/ROADMAP.md` §Phase 13 — success criteria and scope boundary

### Out-of-Scope Constraints
- `.planning/REQUIREMENTS.md` §Out of Scope — `os.UserConfigDir()` excluded; config auto-creation excluded; Windows/Linux paths excluded

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/gsd-watch/main.go` — existing flag setup with stdlib `flag`; `flag.Parse()` already called; `--no-emoji` already defined — Phase 13 adds `--theme` alongside it and introduces `flag.Visit` block after parse
- `internal/tui/app/model.go` — `app.New(events, noEmoji bool)` is the call site to migrate to `app.New(events, cfg config.Config)`
- `internal/tui/tree/model.go` — `tree.Options{NoEmoji: bool}` pattern from Phase 10 shows how config propagates downward from app

### Established Patterns
- Best-effort parsing: missing/malformed inputs return zero-value structs (Phase 02-01 pattern) — config loading follows the same shape but with explicit error returns at the boundary (main.go handles fatal vs warning)
- `internal/` sub-package per concern: `internal/parser/`, `internal/watcher/`, `internal/tui/` — new `internal/config/` fits this pattern

### Integration Points
- `cmd/gsd-watch/main.go`: flag definition block (add `--theme`), post-parse block (add `flag.Visit` overrides), `app.New()` call site (change signature)
- `internal/tui/app/model.go`: `New()` signature and `noEmoji bool` field → becomes `cfg config.Config`
- `go.mod` / `go.sum`: add `github.com/BurntSushi/toml`

</code_context>

<specifics>
## Specific Ideas

- The `config.Config` struct should use TOML struct tags matching the roadmap-specified key names: `emoji` and `theme`
- Default for `Emoji` field should be `true` (emoji on by default); `--no-emoji` flag sets `cfg.Emoji = false` via `flag.Visit`
- `Theme` default is empty string `""` (meaning "default" — Phase 14 interprets this)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 13-config-infrastructure*
*Context gathered: 2026-03-26*
