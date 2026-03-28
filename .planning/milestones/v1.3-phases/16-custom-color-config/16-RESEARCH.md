# Phase 16: Custom Color Config - Research

**Researched:** 2026-03-27
**Domain:** Go TOML config schema extension, lipgloss hex color application
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**TOML Schema Change**
- D-01: Rename config key `theme` -> `preset`. Existing configs with `theme = "..."` get an unknown-key warning (CFG-03 path) — no migration alias. This frees `[theme]` as a section name.
- D-02: Color overrides live in `[theme]` table in config.toml. `Config` struct gains a nested `ThemeColors` struct with TOML tag `theme`, holding 5 optional string fields.

**Exposed Fields (5 status-tree colors)**
- D-03: The following fields are user-overrideable via `[theme]` keys (TOML key names use snake_case matching struct field names):
  - `complete` — complete status color
  - `active` — in-progress/active status color
  - `pending` — pending/default status color
  - `failed` — failed status color
  - `now_marker` — current-phase arrow color (NowMarker)

  Transient UI fields (RefreshFlash, QuitPending, Highlight, EmptyFg, HelpBorder, HelpFg) are NOT exposed.

**Color Value Format**
- D-04: Only `#RRGGBB` hex strings accepted. Validation: 7-character string starting with `#`. ANSI index strings not supported.

**Invalid Color Handling**
- D-05: Invalid hex strings emit a stderr warning naming the field and the bad value, then fall back to the preset's color. App starts normally. Consistent with CFG-03 behavior — never fatal.

**Apply Logic**
- D-06: `ThemeByName(preset)` resolves the base `Theme`. Then each non-empty `ThemeColors` field is validated and, if valid, overrides the corresponding `Theme` field style with `lipgloss.NewStyle().Foreground(lipgloss.Color(hexValue))`. Applied in `app.New()` after `ThemeByName` returns (or in a new `ApplyColorOverrides(theme, overrides)` helper in `styles.go`).

### Claude's Discretion

- Whether `ThemeColors` is a named struct or a `map[string]string`
- Whether the apply logic lives as a method on `ThemeColors`, a free function in `styles.go`, or inline in `app.New()`
- Exact stderr warning message format (must name the field and bad value)
- Whether short hex (`#RGB`) is silently rejected or expanded to `#RRGGBB`

### Deferred Ideas (OUT OF SCOPE)

- **Multiple named custom profiles** — `[profiles.dark]` / `[profiles.light]` switching — v1.4+
- **`theme = "custom"` preset** — full override preset that defers all colors to `[colors]` table — future phase
</user_constraints>

---

## Summary

Phase 16 is a self-contained config schema extension and color application pass. Two distinct work streams exist: (1) a rename of the top-level config key `theme` -> `preset` affecting `Config` struct, its callers in main.go and app/model.go, and the existing test suite; (2) adding a `[theme]` TOML table with a `ThemeColors` nested struct, hex validation, and an apply step that patches the `Theme` struct returned by `ThemeByName`.

Both streams touch the same small set of files and have no external dependencies beyond libraries already in go.mod (BurntSushi/toml v1.6.0, lipgloss v1.1.0). The existing CFG-03 unknown-key warning path will naturally surface old `theme = "..."` configs after the rename — no special migration code is needed.

The apply logic is straightforward: `lipgloss.Color(hexString)` accepts `#RRGGBB` directly; wrap in `lipgloss.NewStyle().Foreground(...)` to produce a replacement `lipgloss.Style`. The only real judgment call is struct-vs-map for `ThemeColors` — a named struct is the right choice (compile-time exhaustiveness, TOML field tags, testability).

**Primary recommendation:** Implement `ThemeColors` as a named struct with pointer-typed optional string fields (`*string`), place `ApplyColorOverrides(theme tui.Theme, overrides config.ThemeColors) tui.Theme` as a package-level function in `styles.go`, and call it from `app.New()` immediately after `ThemeByName`.

---

## Project Constraints (from CLAUDE.md)

No CLAUDE.md exists in this project. No project-specific overrides apply.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/BurntSushi/toml | v1.6.0 | TOML decode + undecoded-key detection | Already in use; `toml.DecodeFile` + `md.Undecoded()` pattern established in Phase 13 |
| github.com/charmbracelet/lipgloss | v1.1.0 | Style construction from hex color | Already in use; `lipgloss.Color("#RRGGBB")` accepted directly |

Both are already present in go.mod. No new dependencies required.

### How BurntSushi/toml handles nested struct fields

When a TOML file contains a `[theme]` table, BurntSushi/toml maps it to any Go struct field tagged `toml:"theme"`. Pointer vs value struct:

- **Value struct (`ThemeColors`):** TOML omits keys for unset fields. Optional string fields inside should be `*string` so the zero value (nil) is distinguishable from `""` (empty string explicitly set).
- **Alternative `map[string]string`:** Simpler iteration but loses compile-time field enumeration and makes tests more verbose.

The `md.Undecoded()` call in `Load()` reports any key inside `[theme]` that does not match a field of `ThemeColors` — this is the existing CFG-03 mechanism and requires no changes, it fires automatically for free.

**CRITICAL insight on `md.Undecoded()` and the `[theme]` table:** After adding `ThemeColors` as a field, the 5 declared keys (`complete`, `active`, `pending`, `failed`, `now_marker`) will be "decoded" and not appear in `md.Undecoded()`. Any other key the user puts in `[theme]` (e.g., `[theme] highlight = "..."`) will correctly surface as an unknown key via the existing CFG-03 path. No changes needed to the `Load()` undecoded logic.

**CRITICAL insight on the `theme` -> `preset` rename and `md.Undecoded()`:** After renaming `Config.Theme` to `Config.Preset` (with `toml:"preset"`), a config file still containing `theme = "minimal"` (old key) will be reported by `md.Undecoded()` as key `"theme"`. The existing warning loop in main.go prints it. No migration shim needed — behavior falls naturally out of the existing infrastructure.

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Standard `strings` | stdlib | Hex validation (len check, prefix check) | In the validation function — no regex needed |
| Standard `fmt` | stdlib | Warning message formatting | Stderr warning in main.go |

**Installation:** No new packages. All dependencies already present.

---

## Architecture Patterns

### Recommended Project Structure (changes only)

```
internal/config/
├── load.go              — Config struct: Theme -> Preset rename, add ThemeColors nested struct
└── load_test.go         — Update existing tests: Theme -> Preset; add ThemeColors decode tests

internal/tui/
├── styles.go            — Add ApplyColorOverrides(theme Theme, overrides config.ThemeColors) Theme
└── theme_test.go        — Add tests for ApplyColorOverrides

internal/tui/app/
└── model.go             — Update cfg.Theme -> cfg.Preset; call ApplyColorOverrides after ThemeByName

cmd/gsd-watch/
└── main.go              — Update cfg.Theme -> cfg.Preset (2 references); update flag.Visit case
```

### Pattern 1: ThemeColors as Named Struct with Optional Pointer Fields

**What:** A flat struct with `*string` fields, each tagged for TOML. Nil means "not set by user"; non-nil means "user provided a value (valid or invalid)."

**When to use:** When you need compile-time exhaustiveness (can't miss a field), TOML tag control, and easy nil-check in apply logic.

```go
// internal/config/load.go

// ThemeColors holds optional hex color overrides for the 5 user-facing status colors.
// Each field is a pointer so nil means "not overridden" and "" is never ambiguous.
type ThemeColors struct {
    Complete  *string `toml:"complete"`
    Active    *string `toml:"active"`
    Pending   *string `toml:"pending"`
    Failed    *string `toml:"failed"`
    NowMarker *string `toml:"now_marker"`
}

type Config struct {
    Emoji  bool        `toml:"emoji"`
    Preset string      `toml:"preset"`
    Colors ThemeColors `toml:"theme"`
}

func Defaults() Config {
    return Config{Emoji: true, Preset: ""}
}
```

**TOML tag on `Colors` field is `toml:"theme"`** — this maps the `[theme]` TOML table to the `ThemeColors` struct. The field name in Go can be `Colors` (or `Theme`, `ThemeColors`, etc.) — only the TOML tag matters for parsing. Using `Colors` avoids naming confusion with the `tui.Theme` type.

### Pattern 2: ApplyColorOverrides as Free Function in styles.go

**What:** A pure function that takes a resolved `Theme` and a `ThemeColors` override set, validates each non-nil field, applies valid ones, emits warnings for invalid ones, and returns the modified `Theme`.

**When to use:** Keeps `app.New()` readable. Keeps the validation+application logic in the `tui` package co-located with `Theme` and `ThemeByName`. Avoids import cycles (config -> tui would be a cycle; tui -> config is the wrong direction too — see below).

**Import cycle warning:** `styles.go` is in package `tui`. `config.ThemeColors` is in package `config`. If `ApplyColorOverrides` lives in `tui` and accepts `config.ThemeColors`, that creates an import `tui -> config`. Check existing imports: `app/model.go` already imports both `tui` and `config`. But `styles.go` (package `tui`) does NOT currently import `config`. Adding `tui -> config` would create a cycle if `config` imports `tui`.

**Checking actual import graph:** `internal/config/load.go` imports only `errors`, `io/fs`, and `github.com/BurntSushi/toml`. It does NOT import `tui`. So `tui -> config` is safe.

**Alternative (avoids the import):** Accept raw `map[string]*string` or a simple struct defined in `tui` itself. But the cleanest solution is to either:
- Accept `config.ThemeColors` in `tui` (requires `tui` to import `config` — safe per above)
- Define the function in `app/model.go` or `main.go` (keeps it local to the call site)

**Recommendation:** Place it in `styles.go` as `ApplyColorOverrides(theme Theme, overrides config.ThemeColors) Theme` with the warning writer passed as an `io.Writer` parameter (consistent with Phase 8 `DebugOut` pattern — but for this phase, warning output can be handled by the caller in main.go instead). Given the warning must name the field and bad value, the simplest approach is:

```go
// internal/tui/styles.go
// (imports "github.com/radu/gsd-watch/internal/config" added)

// ApplyColorOverrides returns a copy of theme with each non-nil ThemeColors field applied.
// Invalid hex values are appended to the warnings slice so the caller can emit them.
func ApplyColorOverrides(theme Theme, overrides config.ThemeColors, warnings *[]string) Theme {
    apply := func(style *lipgloss.Style, field string, val *string) {
        if val == nil {
            return
        }
        if isValidHex(*val) {
            *style = lipgloss.NewStyle().Foreground(lipgloss.Color(*val))
        } else {
            *warnings = append(*warnings, fmt.Sprintf("gsd-watch: invalid color %q for [theme].%s (ignored)", *val, field))
        }
    }
    apply(&theme.Complete,  "complete",   overrides.Complete)
    apply(&theme.Active,    "active",     overrides.Active)
    apply(&theme.Pending,   "pending",    overrides.Pending)
    apply(&theme.Failed,    "failed",     overrides.Failed)
    apply(&theme.NowMarker, "now_marker", overrides.NowMarker)
    return theme
}

// isValidHex returns true for exactly 7-character strings starting with '#'.
func isValidHex(s string) bool {
    return len(s) == 7 && s[0] == '#'
}
```

Caller in `main.go` collects the warnings slice and prints to stderr. This keeps `ApplyColorOverrides` pure (no direct stderr writes), consistent with the project's testing pattern (no I/O side effects in library functions).

**Alternative warning delivery:** Pass `io.Writer` directly and write inside (Phase 8 `DebugOut` style). Either works; the warnings-slice approach makes unit testing easier (no mock writer needed).

### Pattern 3: Calling ApplyColorOverrides in app.New()

```go
// internal/tui/app/model.go — inside New()

th, _ := tui.ThemeByName(cfg.Preset)  // renamed from cfg.Theme
// Color overrides applied after preset resolution.
// Warnings collected and printed by caller (main.go) before calling app.New(),
// OR collected here if app.New() receives an io.Writer.
// Simplest: apply here, ignore warnings (main.go already validated preset).
// Better: apply in main.go before calling app.New(), passing the final Theme.
```

**Recommendation for apply site:** Apply in `main.go` after theme validation, before calling `app.New()`. This keeps the warning loop in one place (main.go already has the warning pattern from CFG-03 and THEME-04). Pass the final resolved+overridden `tui.Theme` into `app.New()` as an additional parameter, or fold it into the `Config` struct.

The cleanest option given current code: apply in main.go, store the resolved theme in `app.New()` directly. This avoids threading `config.ThemeColors` into the `tui` package at all.

**Two viable call-site designs:**

| Option | Apply Site | Pro | Con |
|--------|------------|-----|-----|
| A | `main.go`, pass `tui.Theme` to `app.New()` | All warnings in one place; `app.New()` signature already has `cfg` | Requires changing `app.New()` signature |
| B | `app.New()` with warnings returned or via writer | `app.New()` self-contained | Warnings surface after program starts, not before |

Option A is cleaner for this codebase style. But changing `app.New()` signature is a larger ripple. **Option B is simpler:** call `ApplyColorOverrides` inside `app.New()` with a pre-collected warnings slice, then have `New()` return warnings alongside the model (breaking current signature) — or just print warnings inside `ApplyColorOverrides` directly to `os.Stderr`.

**Final recommendation:** Have `ApplyColorOverrides` accept an `io.Writer` for warnings (pass `os.Stderr` from main.go or `app.New()`), call it inside `app.New()` after `ThemeByName`. This requires no signature change to `app.New()` and no new return values.

### Anti-Patterns to Avoid

- **Using `string` instead of `*string` for ThemeColors fields:** If `Complete string` is used, you cannot distinguish "user set `complete = ""`" from "user did not set `complete`". Empty string and nil are different. Use `*string`.
- **Applying overrides before `ThemeByName`:** The override must layer on top of the resolved preset, not modify a zero-value Theme. Always call `ThemeByName` first.
- **Calling `lipgloss.Color(val)` without validation:** lipgloss accepts any string as a color and will not panic on invalid hex — but it produces undefined rendering output. Always validate before applying.
- **Adding `theme` to flag.Visit switch:** The CLI `--theme` flag sets `cfg.Preset` (renamed). The flag itself remains `--theme` for user-facing backward compatibility (or it gets renamed too — see Open Questions). Do not confuse the TOML key with the flag name.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Hex color application | Custom ANSI escape builder | `lipgloss.NewStyle().Foreground(lipgloss.Color("#RRGGBB"))` | lipgloss handles terminal profile detection, 256/true-color negotiation, light/dark adaptation |
| TOML nested struct decode | Manual key parsing | `toml.DecodeFile` into nested struct | BurntSushi/toml handles nested tables natively via struct tags |
| Unknown-key detection for `[theme]` sub-keys | Custom key list tracking | `md.Undecoded()` | Already implemented; fires automatically for any key not matched by `ThemeColors` fields |

---

## Runtime State Inventory

Not applicable. This is a greenfield feature addition (new struct fields, new function). No rename/refactor of stored data is involved.

The `theme` -> `preset` config key rename is a TOML key rename only. There are no databases, registries, or stored state containing the string "theme" as a config key. The rename only affects: Go struct field tag, `Defaults()` return value, field references in `.go` source files. All are code edits.

---

## Common Pitfalls

### Pitfall 1: BurntSushi/toml and `*string` pointer fields

**What goes wrong:** BurntSushi/toml does not automatically set pointer fields to non-nil when a TOML key is absent. A `*string` field stays nil if the key is omitted. This is the correct behavior — but only if `cfg := Defaults()` is called before `toml.DecodeFile`, which the existing `Load()` already does (Phase 13 decision). If `Defaults()` is not called first, the zero value of `ThemeColors` is a struct with all nil pointers, which is correct for "no overrides."

**Why it happens:** Go zero value for pointer is nil. TOML decode sets fields only when the key is present.

**How to avoid:** Keep the existing pattern: `cfg := Defaults()` then `toml.DecodeFile(path, &cfg)`. ThemeColors zero value (all nil pointers) is the correct "no overrides" state.

**Warning signs:** If you see `ThemeColors` fields as `string` instead of `*string` and tests fail to detect "not set" vs "set to empty string."

### Pitfall 2: The `[theme]` section appearing in md.Undecoded() before struct is added

**What goes wrong:** If the developer adds `[theme]` to a test config file but forgets to add the `ThemeColors` field to the `Config` struct, the entire `[theme]` table key shows up in `md.Undecoded()` as a single unknown key. The CFG-03 warning fires, but it reports the table-level key not the individual sub-keys.

**Why it happens:** BurntSushi/toml treats an unmatched table as a single undecoded key.

**How to avoid:** Add the `Colors ThemeColors \`toml:"theme"\`` field to `Config` before running any config tests with `[theme]`.

### Pitfall 3: `--theme` flag name vs `preset` field name divergence

**What goes wrong:** After renaming `Config.Theme` to `Config.Preset`, main.go has `flag.Visit` cases for `"theme"` (the flag name) and sets `cfg.Theme` (now `cfg.Preset`). Forgetting to update `cfg.Theme` -> `cfg.Preset` inside the `flag.Visit` switch will cause the flag override to silently do nothing.

**Why it happens:** The flag name `"theme"` and the field name `Theme`/`Preset` are two different things. The flag name can remain `"theme"` (user-facing), but the field access must be `cfg.Preset`.

**How to avoid:** Search for all occurrences of `cfg.Theme` across the codebase and update every one. The grep result confirms 3 files contain it: `main.go`, `model_test.go` (via `config.Defaults()`), `app/model.go`.

**Exact locations to update:**
- `main.go:84`: `cfg.Theme = *themeFlag` -> `cfg.Preset = *themeFlag`
- `main.go:89-93`: `cfg.Theme` references in THEME-04 warning block
- `app/model.go:53`: `tui.ThemeByName(cfg.Theme)` -> `tui.ThemeByName(cfg.Preset)`
- `app/model.go:351`: `themeName := m.cfg.Theme` -> `themeName := m.cfg.Preset`
- `load_test.go:28,43,107`: test assertions on `Config.Theme` field -> `Config.Preset`
- `load_test.go:27`: `wantConfig: Config{Emoji: false, Theme: "minimal"}` -> `Config{Emoji: false, Preset: "minimal"}`

### Pitfall 4: `--theme` flag help text in main.go and --help output

**What goes wrong:** The `--help` block in main.go documents `--theme` as "Color theme name (overrides config file)". After the TOML key rename, the config file key is now `preset`, not `theme`. The help text might become confusing if it says "overrides config file" but the config key is now `preset`.

**Why it happens:** CLI flags and TOML keys are decoupled; documentation must track both.

**How to avoid:** Decide whether `--theme` flag also gets renamed to `--preset`. The CONTEXT.md does not address this. The `--theme` flag name is CLI-facing and can remain `--theme` for backward compatibility — it's orthogonal to the TOML key rename. No action needed unless the user decides otherwise.

### Pitfall 5: lipgloss style modification semantics

**What goes wrong:** `lipgloss.Style` is a value type. Calling `.Foreground()` on an existing style returns a new style — it does not mutate in place. If you do `theme.Complete.Foreground(...)` without assigning the result back, the override is silently dropped.

**Why it happens:** lipgloss builder pattern: each method returns a new `lipgloss.Style` value.

**How to avoid:** Always assign: `theme.Complete = lipgloss.NewStyle().Foreground(lipgloss.Color(val))`. Note this creates a fresh style (losing Bold, etc. from the preset). For the 5 status-tree colors this is fine — none have bold set (only `high-contrast` uses `.Bold(true)` on status styles). If the user overrides a color on `high-contrast`, the bold is lost. This is acceptable per D-06 which says `lipgloss.NewStyle().Foreground(...)` explicitly.

**Warning signs:** Override appears to succeed at compile time but has no visible effect at runtime.

---

## Code Examples

### ThemeColors struct and Config update

```go
// internal/config/load.go

// ThemeColors holds optional hex color overrides for the 5 status-tree colors.
// Pointer fields: nil = not set by user; non-nil = user provided a value.
type ThemeColors struct {
    Complete  *string `toml:"complete"`
    Active    *string `toml:"active"`
    Pending   *string `toml:"pending"`
    Failed    *string `toml:"failed"`
    NowMarker *string `toml:"now_marker"`
}

type Config struct {
    Emoji  bool        `toml:"emoji"`
    Preset string      `toml:"preset"`    // renamed from Theme
    Colors ThemeColors `toml:"theme"`     // [theme] table in TOML
}

func Defaults() Config {
    return Config{Emoji: true, Preset: ""}
}
```

### Hex validation function

```go
// internal/tui/styles.go (or internal/config/load.go — either works)

// isValidHex returns true if s is a valid #RRGGBB hex string.
// Short form #RGB is rejected per D-04.
func isValidHex(s string) bool {
    return len(s) == 7 && s[0] == '#'
}
```

### ApplyColorOverrides function

```go
// internal/tui/styles.go

import (
    "fmt"
    "io"
    "github.com/radu/gsd-watch/internal/config"
)

// ApplyColorOverrides returns a copy of theme with each non-nil ThemeColors field
// applied as a hex foreground color. Invalid hex values emit a warning to w.
// Pass os.Stderr for production; pass a bytes.Buffer for tests.
func ApplyColorOverrides(theme Theme, overrides config.ThemeColors, w io.Writer) Theme {
    applyField := func(style *lipgloss.Style, field string, val *string) {
        if val == nil {
            return
        }
        if isValidHex(*val) {
            *style = lipgloss.NewStyle().Foreground(lipgloss.Color(*val))
        } else {
            fmt.Fprintf(w, "gsd-watch: invalid color %q for [theme].%s (ignored)\n", *val, field)
        }
    }
    applyField(&theme.Complete,  "complete",   overrides.Complete)
    applyField(&theme.Active,    "active",     overrides.Active)
    applyField(&theme.Pending,   "pending",    overrides.Pending)
    applyField(&theme.Failed,    "failed",     overrides.Failed)
    applyField(&theme.NowMarker, "now_marker", overrides.NowMarker)
    return theme
}
```

### app.New() update (minimal diff)

```go
// internal/tui/app/model.go — inside New()

th, _ := tui.ThemeByName(cfg.Preset)                           // was cfg.Theme
th = tui.ApplyColorOverrides(th, cfg.Colors, os.Stderr)        // new line
t = t.SetOptions(tree.Options{NoEmoji: !cfg.Emoji, Theme: th}) // unchanged
```

### Valid TOML after Phase 16

```toml
preset = "minimal"
emoji  = true

[theme]
complete   = "#00cc00"
failed     = "#cc0000"
now_marker = "#ffaa00"
```

### Test fixture for ThemeColors decode

```toml
# internal/config/testdata/theme-colors.toml
preset = "default"
emoji  = true

[theme]
complete = "#00ff00"
failed   = "#ff0000"
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| TOML `theme` key (string) | TOML `preset` key (string) + `[theme]` table | Phase 16 | Old `theme = "..."` configs warn via CFG-03; `[theme]` is now reserved for color overrides |

**Deprecated after Phase 16:**
- `Config.Theme string \`toml:"theme"\`` field — replaced by `Config.Preset string \`toml:"preset"\`` and `Config.Colors ThemeColors \`toml:"theme"\``

---

## Open Questions

1. **Should `--theme` CLI flag be renamed to `--preset`?**
   - What we know: The CONTEXT.md renames the TOML key only. The flag name is separate.
   - What's unclear: Whether keeping `--theme` flag while the TOML key is `preset` is confusing.
   - Recommendation: Keep `--theme` flag name for backward compatibility. The flag and the TOML key do not need to match. Document in `--help` output: "Color theme preset name (overrides `preset` in config)".

2. **Should `--preset` flag be added as an alias?**
   - What we know: Not mentioned in CONTEXT.md.
   - What's unclear: User expectation after reading config docs that say `preset`.
   - Recommendation: Out of scope for Phase 16. Not mentioned in CONTEXT.md.

3. **Should Phase 16 update the `--help` text in main.go to reflect `preset` key?**
   - What we know: The `--help` output currently says `--theme  Color theme name (overrides config file)`.
   - What's unclear: Whether "config file" reference is sufficient or should mention `preset =`.
   - Recommendation: Update `--help` text inline while touching main.go. Low risk, improves accuracy.

---

## Environment Availability

Step 2.6: SKIPPED (no external dependencies — all required libraries already in go.mod; no CLI tools, databases, or services needed beyond the Go toolchain already confirmed working).

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go standard `testing` package |
| Config file | none (go test ./... convention) |
| Quick run command | `go test ./internal/config/ ./internal/tui/` |
| Full suite command | `go test ./...` |

Current baseline: all 10 packages pass (`go test ./...` output confirmed above).

### Phase Requirements -> Test Map

Phase 16 has no assigned requirement IDs (TBD per phase description). The behaviors to test are derived from the locked decisions:

| Behavior ID | Behavior | Test Type | Automated Command | File Exists? |
|-------------|----------|-----------|-------------------|-------------|
| P16-CFG-A | `Config.Preset` field decoded from TOML `preset` key | unit | `go test ./internal/config/ -run TestLoad` | Existing — needs update |
| P16-CFG-B | `Config.Colors.Complete` etc. decoded from `[theme]` table | unit | `go test ./internal/config/ -run TestLoad_ThemeColors` | Wave 0 gap |
| P16-CFG-C | `[theme]` section with no keys produces no warnings | unit | `go test ./internal/config/ -run TestLoad_EmptyTheme` | Wave 0 gap |
| P16-CFG-D | Old `theme = "..."` TOML key reports unknown-key warning | unit | `go test ./internal/config/ -run TestLoad_OldThemeKey` | Wave 0 gap |
| P16-APPLY-A | Valid hex override replaces preset color in Theme | unit | `go test ./internal/tui/ -run TestApplyColorOverrides_Valid` | Wave 0 gap |
| P16-APPLY-B | Invalid hex string emits warning, preset color preserved | unit | `go test ./internal/tui/ -run TestApplyColorOverrides_Invalid` | Wave 0 gap |
| P16-APPLY-C | nil field (not set) leaves preset color unchanged | unit | `go test ./internal/tui/ -run TestApplyColorOverrides_NilUnchanged` | Wave 0 gap |
| P16-APPLY-D | Short hex `#RGB` rejected (7-char rule) | unit | `go test ./internal/tui/ -run TestIsValidHex` | Wave 0 gap |
| P16-RENAME-A | All `cfg.Theme` references compile-clean after rename | compile | `go build ./...` | n/a |
| P16-INT-A | app.New() with color overrides does not panic | unit | `go test ./internal/tui/ -run TestNew_WithColorOverrides` | Wave 0 gap |

### Sampling Rate

- **Per task commit:** `go test ./internal/config/ ./internal/tui/`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `go test ./...` green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/config/testdata/theme-colors.toml` — fixture with `[theme]` overrides
- [ ] `internal/config/testdata/theme-colors-invalid.toml` — fixture with invalid hex values
- [ ] `internal/config/testdata/old-theme-key.toml` — fixture with legacy `theme = "minimal"` key (triggers CFG-03)
- [ ] Test cases in `internal/config/load_test.go` — `TestLoad_ThemeColors`, `TestLoad_EmptyTheme`, `TestLoad_OldThemeKey`
- [ ] `TestApplyColorOverrides_Valid`, `TestApplyColorOverrides_Invalid`, `TestApplyColorOverrides_NilUnchanged`, `TestIsValidHex` in `internal/tui/theme_test.go` (or new `color_overrides_test.go`)
- [ ] `TestNew_WithColorOverrides` in `internal/tui/model_test.go`

Existing `TestLoad` cases will need field name updates (`Theme` -> `Preset`) but the test logic is unchanged.

---

## Sources

### Primary (HIGH confidence)

- Direct code inspection: `/Users/radu/Developer/gsd-watch/internal/config/load.go` — Config struct, Load(), Defaults(), UnknownKeysError pattern
- Direct code inspection: `/Users/radu/Developer/gsd-watch/internal/tui/styles.go` — Theme struct, ThemeByName(), all 5 overrideable field names confirmed
- Direct code inspection: `/Users/radu/Developer/gsd-watch/cmd/gsd-watch/main.go` — cfg.Theme usage locations, flag.Visit pattern, warning stderr pattern
- Direct code inspection: `/Users/radu/Developer/gsd-watch/internal/tui/app/model.go` — app.New() signature, ThemeByName call site, helpView cfg.Theme usage
- Direct code inspection: `/Users/radu/Developer/gsd-watch/go.mod` — confirmed BurntSushi/toml v1.6.0, lipgloss v1.1.0, Go 1.26.1
- BurntSushi/toml documentation: nested struct decode via struct tags, `md.Undecoded()` behavior — HIGH confidence from Phase 13 prior use in this project
- lipgloss v1.x: `lipgloss.Color()` accepts `#RRGGBB` hex strings directly — HIGH confidence from existing codebase usage

### Secondary (MEDIUM confidence)

- Phase 13 CONTEXT.md D-04 (BurntSushi/toml selection), D-05 (flag.Visit pattern), prior art in project
- Phase 14 CONTEXT.md D-04 (Theme struct shape), D-06 (ThemeByName signature)
- Phase 15 CONTEXT.md D-03 (helpView signature with cfg.Theme)

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries already in use in this project; no new dependencies
- Architecture: HIGH — direct code inspection of all affected files; patterns established in prior phases
- Pitfalls: HIGH — derived from direct reading of actual code paths that will be touched

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable dependencies; lipgloss/toml APIs are stable)
