# Phase 13: Config Infrastructure - Research

**Researched:** 2026-03-26
**Domain:** Go config file loading (TOML), flag override wiring, internal package design
**Confidence:** HIGH

## Summary

Phase 13 is a well-scoped Go package addition. The implementation surface is narrow: one new `internal/config/` package, one dependency (`github.com/BurntSushi/toml`), a signature change to `app.New()`, and a `flag.Visit` block in `main.go`. All decisions are locked in CONTEXT.md; there is no ambiguity in approach.

BurntSushi/toml v1.6.0 (released Dec 2025) is the current stable release. The `toml.DecodeFile` function returns `(MetaData, error)` — the same call handles all three error cases: missing file (wrapped `fs.ErrNotExist`), malformed TOML (`toml.ParseError`), and unknown keys (non-empty `md.Undecoded()`). No alternative code paths are needed.

The `app.New()` signature migration from `(events, noEmoji bool)` to `(events chan tea.Msg, cfg config.Config)` is the only callsite change outside the new package. The project's existing test pattern (table-driven tests with `testdata/` fixtures) applies directly to the config loader.

**Primary recommendation:** Create `internal/config/load.go` as a single file exporting `Load(path string) (Config, error)` and a `Defaults()` function. Wire the call in `main.go` using the three-case dispatch (missing → warn, malformed → fatal, unknown keys → stderr). Use `flag.Visit` after `flag.Parse()` to apply overrides.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** `app.New()` signature changes from `app.New(events, noEmoji bool)` to `app.New(events chan tea.Msg, cfg config.Config)` — Phase 14 adds fields to `config.Config` without touching the signature again
- **D-02:** `main.go` resolves config, applies flag overrides via `flag.Visit`, then passes the final `config.Config` to `app.New()` — config package stays out of the app package
- **D-03:** Phase 13 adds `--theme <name>` to `main.go` and stores whatever string is provided in `cfg.Theme` — no validation, no warning for unknown names. Phase 14 owns validation and the unknown-theme warning (CFG-05 is about override precedence only)
- **D-04:** Use `github.com/BurntSushi/toml` — `toml.MetaData.Undecoded()` provides the unknown key list for CFG-03 stderr warning with zero extra logic
- **D-05:** `flag.Visit` iterates over explicitly-set flags after `flag.Parse()`. If `--no-emoji` was set, override `cfg.Emoji = false`. If `--theme` was set, override `cfg.Theme = <value>`. Config file values apply only when the corresponding flag was NOT explicitly passed.
- **D-06:** `os.UserHomeDir()` + manual XDG join (`~/.config/gsd-watch/config.toml`) — `os.UserConfigDir()` is excluded per REQUIREMENTS.md (returns wrong path on macOS)

### Claude's Discretion

- Internal structure of `internal/config/` package (single file vs multiple)
- `config.Config` field names and TOML tag names (recommend: `Emoji bool \`toml:"emoji"\``, `Theme string \`toml:"theme"\``)
- Fatal error message format for malformed TOML (must include file path per CFG-02)
- Unknown key warning format for stderr (must name the key per CFG-03)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CFG-01 | Missing config → silent defaults, no error/crash/log noise | `errors.Is(err, fs.ErrNotExist)` check; return zero-value `Config{Emoji: true}` |
| CFG-02 | Malformed TOML → fatal error with file path | `toml.ParseError` type assertion; `log.Fatalf` with path + error message |
| CFG-03 | Unknown config keys → stderr warning, still starts | `md.Undecoded()` returns `[]toml.Key`; loop and `fmt.Fprintf(os.Stderr, ...)` per key |
| CFG-04 | `--no-emoji` flag overrides config emoji key | `flag.Visit` after `flag.Parse()`; `if name == "no-emoji" { cfg.Emoji = false }` |
| CFG-05 | `--theme` flag overrides config theme key | `flag.Visit` after `flag.Parse()`; `if name == "theme" { cfg.Theme = *themeFlag }` |
</phase_requirements>

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/BurntSushi/toml | v1.6.0 | TOML file decode + unknown key detection | Locked in D-04; `Undecoded()` eliminates custom key-walk logic; most downloaded Go TOML library |
| stdlib `flag` | Go 1.26.1 | CLI flag parsing + `flag.Visit` for override detection | Already in use; `flag.Visit` is the idiomatic way to distinguish "explicitly set" from "default" |
| stdlib `os` | Go 1.26.1 | `os.UserHomeDir()`, file existence check | Locked in D-06 over `os.UserConfigDir()` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib `errors` | Go 1.26.1 | `errors.Is(err, fs.ErrNotExist)` | Distinguish missing-file from parse errors |
| stdlib `fmt` | Go 1.26.1 | `fmt.Fprintf(os.Stderr, ...)` for unknown-key warnings | stderr output without `log` package overhead |
| stdlib `path/filepath` | Go 1.26.1 | Path construction only if needed | Already imported in main.go |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| BurntSushi/toml | pelletier/go-toml | go-toml v2 has a `Strict()` mode but requires different API; BurntSushi locked by D-04 |
| `flag.Visit` | custom boolean tracking | `flag.Visit` is standard library; no extra variables needed |
| `os.UserHomeDir()` | `os.UserConfigDir()` | `UserConfigDir()` returns `~/Library/Application Support` on macOS — excluded by REQUIREMENTS.md |

**Installation:**
```bash
go get github.com/BurntSushi/toml@v1.6.0
```

**Version verification:**
```
github.com/BurntSushi/toml v1.6.0 — published 2025-12-18 (verified via Go module proxy)
```

---

## Architecture Patterns

### Recommended Project Structure
```
internal/
└── config/
    └── load.go         # Config struct, Load(), Defaults(), config path constant
    └── load_test.go    # table-driven tests with testdata/
    └── testdata/
        ├── valid.toml
        ├── malformed.toml
        ├── unknown-keys.toml
        └── empty.toml
```

Single-file package is appropriate: only 2 exported types, 2 functions, ~60 lines of logic. Splitting would add navigation overhead with no benefit.

### Pattern 1: Three-Case Config Loader

**What:** `Load()` returns `(Config, error)` distinguishing three error classes at the callsite boundary.
**When to use:** All config load operations in `main.go`

```go
// Source: BurntSushi/toml v1.6.0 pkg.go.dev docs
package config

import (
    "errors"
    "io/fs"
    "github.com/BurntSushi/toml"
)

type Config struct {
    Emoji bool   `toml:"emoji"`
    Theme string `toml:"theme"`
}

// ConfigPath is the canonical XDG path on macOS.
const ConfigPath = ".config/gsd-watch/config.toml"

type UnknownKeysError struct {
    Keys []toml.Key
}

func (e *UnknownKeysError) Error() string { return "unknown config keys" }

// Load reads the config file at path and returns the decoded Config.
// Three distinct outcomes:
//   - file missing: returns (Defaults(), nil)   — CFG-01
//   - malformed TOML: returns (Defaults(), err) — CFG-02, caller must fatal
//   - unknown keys: returns (cfg, &UnknownKeysError{...}) — CFG-03, caller warns then continues
func Load(path string) (Config, error) {
    cfg := Defaults()
    md, err := toml.DecodeFile(path, &cfg)
    if err != nil {
        if errors.Is(err, fs.ErrNotExist) {
            return Defaults(), nil  // CFG-01: missing is OK
        }
        return Defaults(), err      // CFG-02: parse error, caller fatals
    }
    if undecoded := md.Undecoded(); len(undecoded) > 0 {
        return cfg, &UnknownKeysError{Keys: undecoded} // CFG-03
    }
    return cfg, nil
}

// Defaults returns a Config with all fields at their documented defaults.
func Defaults() Config {
    return Config{Emoji: true, Theme: ""}
}
```

### Pattern 2: flag.Visit Override Block in main.go

**What:** Walk only explicitly-set flags after `flag.Parse()` to apply CLI overrides.
**When to use:** After `Load()` succeeds (or returns defaults for missing/unknown-key cases)

```go
// Source: stdlib flag documentation
flag.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "no-emoji":
        cfg.Emoji = false
    case "theme":
        cfg.Theme = *themeFlag
    }
})
```

`flag.Visit` only visits flags that were explicitly set on the command line (not flags left at their default values). This is the idiomatic Go solution for "flag beats config file."

### Pattern 3: main.go Three-Case Dispatch

**What:** Dispatch on the three return cases from `Load()` before calling `app.New()`
**When to use:** Only in `main.go` — config package returns errors, boundary handles behaviour

```go
homeDir, _ := os.UserHomeDir()
cfgPath := filepath.Join(homeDir, config.ConfigPath)

cfg, err := config.Load(cfgPath)
if err != nil {
    var ukErr *config.UnknownKeysError
    if errors.As(err, &ukErr) {
        // CFG-03: warn on stderr, continue with partial config
        for _, k := range ukErr.Keys {
            fmt.Fprintf(os.Stderr, "gsd-watch: unknown config key %q (ignored)\n", k)
        }
    } else {
        // CFG-02: malformed TOML — fatal with path
        fmt.Fprintf(os.Stderr, "gsd-watch: error reading config %s: %v\n", cfgPath, err)
        os.Exit(1)
    }
}

// Apply CLI flag overrides (D-05)
flag.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "no-emoji":
        cfg.Emoji = false
    case "theme":
        cfg.Theme = *themeFlag
    }
})

// app.New() receives the resolved config (D-01)
p := tea.NewProgram(app.New(events, cfg), tea.WithAltScreen())
```

### Anti-Patterns to Avoid

- **Calling `os.UserConfigDir()`:** Returns `~/Library/Application Support` on macOS, not `~/.config`. Excluded by REQUIREMENTS.md and D-06.
- **Doing config loading in `app.New()` or `app.Init()`:** Config belongs at the `main.go` boundary (D-02). The app package must not import `internal/config`.
- **Using `log.Fatal` for unknown keys (CFG-03):** Unknown keys are a warning, not a fatal error. Mixing log.Fatal with fmt.Fprintf(os.Stderr) creates inconsistent behaviour; use `os.Exit(1)` only for CFG-02.
- **Passing `noEmoji bool` through the call chain after Phase 13:** The `noEmoji` field in `app.Model` should be replaced by reading from `cfg.Emoji` — don't maintain both.
- **Forgetting to update `helpView` call signature:** `helpView(width, noEmoji)` currently passes `m.noEmoji`; after the migration, this becomes `!m.cfg.Emoji` (note the inversion — `Emoji=true` means show emoji, `NoEmoji=true` means suppress).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Unknown key detection | Walk struct fields with reflection | `md.Undecoded()` | BurntSushi tracks decoded keys at parse time; reflection approach misses nested tables and aliased fields |
| "Was this flag explicitly set?" | Boolean sentinel variables per flag | `flag.Visit` | Stdlib-provided; handles all flag types uniformly; zero extra state |
| TOML parse errors | Custom string parsing of TOML errors | `toml.ParseError` type assertion | Library provides structured error with position info |

**Key insight:** The three hard problems in this phase (unknown keys, flag-beats-config, error classification) each have exactly one correct solution in the standard/library layer. Custom solutions would be inferior and harder to test.

---

## Common Pitfalls

### Pitfall 1: `Emoji` field default is `true`, not `false`

**What goes wrong:** Zero-value `bool` in Go is `false`. If `Config{}.Emoji` is `false` and no config file is present, emoji is suppressed by default — the opposite of the intended behaviour.
**Why it happens:** Go zero values don't match the semantic default for this field.
**How to avoid:** Always initialize via `Defaults()` not `Config{}`. The `Load()` function must start with `cfg := Defaults()` before calling `toml.DecodeFile`.
**Warning signs:** Test for `CFG-01` (missing file) unexpectedly shows ASCII icons.

### Pitfall 2: `flag.Visit` must run after `flag.Parse()`

**What goes wrong:** Calling `flag.Visit` before `flag.Parse()` visits nothing (no flags have been set yet).
**Why it happens:** `flag.Visit` reflects parsed state, not declared state.
**How to avoid:** The `flag.Visit` block must appear after `flag.Parse()` in `main()`. Order: declare flags → `flag.Parse()` → `Load()` → `flag.Visit` → `app.New()`.
**Warning signs:** CLI flags appear to have no effect when config file is present.

### Pitfall 3: `--no-emoji` is a bool flag, not a string flag

**What goes wrong:** `flag.Bool("no-emoji", false, ...)` stores a `*bool`. Trying to read its string value in `flag.Visit` via `f.Value.String()` returns `"false"` even when set, not the flag name.
**Why it happens:** `flag.Visit` passes `*flag.Flag`; the name is in `f.Name`, not the value.
**How to avoid:** In the `flag.Visit` callback, switch on `f.Name` (the flag name), not `f.Value`. The `--no-emoji` flag's presence in the visited set is the signal; no value inspection needed.

### Pitfall 4: `noEmoji` field in `app.Model` must be removed, not left alongside `cfg`

**What goes wrong:** After the signature change, `app.Model` still has a `noEmoji bool` field, but `New()` now receives `cfg config.Config`. If both fields exist, they can diverge — the old field is never updated when config changes.
**Why it happens:** Incremental refactors leave dead fields.
**How to avoid:** Delete `noEmoji bool` from `app.Model` and replace all `m.noEmoji` reads with `!m.cfg.Emoji`. Plan should include this cleanup explicitly.

### Pitfall 5: `toml.DecodeFile` wraps the OS error

**What goes wrong:** `err == fs.ErrNotExist` (equality check) returns false even for missing files.
**Why it happens:** `toml.DecodeFile` calls `os.Open` internally; the resulting `*os.PathError` wraps `fs.ErrNotExist` rather than being equal to it.
**How to avoid:** Use `errors.Is(err, fs.ErrNotExist)` (unwrapping check), not `==`.
**Warning signs:** Missing config file causes a fatal error instead of silent defaults.

---

## Code Examples

Verified patterns from official sources:

### DecodeFile with Undecoded() — CFG-01, CFG-02, CFG-03
```go
// Source: https://pkg.go.dev/github.com/BurntSushi/toml#MetaData.Undecoded
cfg := Defaults()
md, err := toml.DecodeFile(path, &cfg)
if err != nil {
    if errors.Is(err, fs.ErrNotExist) {
        return Defaults(), nil      // CFG-01: silent missing
    }
    return Defaults(), err          // CFG-02: parse failure, caller fatals
}
if keys := md.Undecoded(); len(keys) > 0 {
    return cfg, &UnknownKeysError{Keys: keys}  // CFG-03: unknown keys warning
}
return cfg, nil
```

### flag.Visit for override detection — CFG-04, CFG-05
```go
// Source: https://pkg.go.dev/flag#Visit
// flag.Visit calls fn for each flag that has been set (not just declared).
themeFlag := flag.String("theme", "", "Theme name")
noEmoji   := flag.Bool("no-emoji", false, "Use ASCII icons")
flag.Parse()

cfg, _ := config.Load(cfgPath) // simplified

flag.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "no-emoji":
        cfg.Emoji = false
    case "theme":
        cfg.Theme = *themeFlag
    }
})
```

### TOML config file format (what users will write)
```toml
# ~/.config/gsd-watch/config.toml
emoji = false
theme = "default"
```

### app.New() new signature
```go
// internal/tui/app/model.go
func New(events chan tea.Msg, cfg config.Config) Model {
    t := tree.New()
    t = t.SetOptions(tree.Options{NoEmoji: !cfg.Emoji})
    // ...
    return Model{
        // ...
        cfg: cfg,
    }
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `app.New(events, noEmoji bool)` | `app.New(events chan tea.Msg, cfg config.Config)` | Phase 13 | One callsite in main.go; `app.Model.noEmoji` field replaced by `cfg.Emoji` |
| No config file support | `~/.config/gsd-watch/config.toml` via BurntSushi/toml | Phase 13 | New `internal/config/` package |

**Deprecated/outdated after Phase 13:**
- `app.Model.noEmoji bool` field — replaced by `app.Model.cfg config.Config`
- `app.New(events, noEmoji bool)` signature — becomes `app.New(events, cfg)`
- Direct `!*noEmoji` reference in `main.go` → replaced by `flag.Visit` + `config.Load`

---

## Open Questions

1. **Should `UnknownKeysError` be a sentinel or inspected?**
   - What we know: The caller (main.go) only needs to iterate the keys for the stderr message
   - What's unclear: Whether future callers need programmatic access to unknown key names
   - Recommendation: Export `UnknownKeysError.Keys []toml.Key` for testability; the planner can decide to use a simpler `[]string` if the `toml.Key` type import feels heavyweight in tests

2. **`toml.Key` type in tests**
   - What we know: `md.Undecoded()` returns `[]toml.Key`, not `[]string`
   - What's unclear: Whether test assertions comparing unknown keys need to import `toml` package or if string conversion is sufficient
   - Recommendation: Store as `[]string` in `UnknownKeysError` by converting `k.String()` at collection time — avoids exposing the toml package type in the config package's public API

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All compilation | Yes | 1.26.1 (from go.mod) | — |
| github.com/BurntSushi/toml | CFG-01/02/03 | Not yet in go.mod | v1.6.0 (available) | — |
| `go test ./...` | Test suite | Yes | passes (all green) | — |

**Missing dependencies with no fallback:**
- None that would block execution. `go get github.com/BurntSushi/toml@v1.6.0` is a Wave 0 task.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` |
| Config file | none (no pytest.ini / jest equivalent — go test is convention-based) |
| Quick run command | `go test ./internal/config/... -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CFG-01 | Missing config → silent defaults, no error | unit | `go test ./internal/config/... -run TestLoad/missing_file` | No — Wave 0 |
| CFG-02 | Malformed TOML → returns error with path info | unit | `go test ./internal/config/... -run TestLoad/malformed_toml` | No — Wave 0 |
| CFG-03 | Unknown keys → `UnknownKeysError` returned, Config still populated | unit | `go test ./internal/config/... -run TestLoad/unknown_keys` | No — Wave 0 |
| CFG-04 | `--no-emoji` CLI flag overrides `cfg.Emoji` | integration (main.go wiring) | `go test ./internal/config/... -run TestDefaults`; flag-visit logic manual-only (no main_test.go) | No — Wave 0 |
| CFG-05 | `--theme` CLI flag overrides `cfg.Theme` | integration (main.go wiring) | same as CFG-04 | No — Wave 0 |

Note: CFG-04 and CFG-05 are flag-wiring logic in `main.go`. The `flag.Visit` block is simple enough to verify by inspection; unit coverage for `Load()` covers the config side. A manual smoke test (run binary with `--no-emoji` and a config with `emoji = true`) is the verification path.

### Sampling Rate
- **Per task commit:** `go test ./internal/config/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/config/load.go` — package does not yet exist
- [ ] `internal/config/load_test.go` — table-driven tests for CFG-01/02/03
- [ ] `internal/config/testdata/valid.toml` — two-key TOML fixture
- [ ] `internal/config/testdata/malformed.toml` — invalid TOML syntax fixture
- [ ] `internal/config/testdata/unknown-keys.toml` — known keys plus one unknown key

---

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/BurntSushi/toml` — DecodeFile, MetaData.Undecoded(), ParseError, error wrapping (verified via WebFetch 2026-03-26)
- Go module proxy `proxy.golang.org` — v1.6.0 published 2025-12-18 (verified via API)
- `pkg.go.dev/flag` — flag.Visit behaviour (stdlib, always current)
- `pkg.go.dev/os` — os.UserHomeDir() vs os.UserConfigDir() behaviour

### Secondary (MEDIUM confidence)
- REQUIREMENTS.md §Out of Scope — `os.UserConfigDir()` exclusion with Go issue reference (#76320) documented by project author

### Tertiary (LOW confidence)
- None

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — versions verified via Go module proxy; library API verified via pkg.go.dev
- Architecture: HIGH — all decisions locked in CONTEXT.md; patterns are direct applications of standard Go idioms
- Pitfalls: HIGH — `errors.Is` wrapping and bool zero-value pitfalls are Go fundamentals; `flag.Visit` ordering is documented in stdlib

**Research date:** 2026-03-26
**Valid until:** 2026-09-26 (BurntSushi/toml stable API; stdlib patterns don't change)
