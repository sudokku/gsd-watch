# Pitfalls Research — v1.3 Config + Theme

**Domain:** Adding TOML config file + 3-preset theme system to existing Go Bubble Tea TUI
**Researched:** 2026-03-26
**Confidence:** HIGH for TOML/flag precedence (stdlib + BurntSushi docs); HIGH for XDG gotcha (verified against open Go issue #76320); HIGH for lipgloss package-var mutation (verified against lipgloss source); HIGH for AdaptiveColor loss in theme presets (verified against lipgloss AdaptiveColor semantics); MEDIUM for Bubble Tea theme threading (pattern from maintainer discussions, no canonical guide exists)

> This file is the v1.3-specific supplement to the broader ecosystem pitfalls previously documented.
> The prior v1.0 concerns (Bubble Tea concurrency, fsnotify, socket IPC, YAML frontmatter, tmux detection)
> are already shipped and are not duplicated here.

---

## Critical Pitfalls

### Pitfall 1: Zero Value Cannot Be Distinguished From "Not Set" When Merging Flag + Config

**What goes wrong:**
`--no-emoji` is a `bool` flag with a zero value of `false`. After `flag.Parse()`, both "user didn't pass the flag" and "user explicitly passed `--no-emoji=false`" produce `*noEmoji == false`. When the config file sets `emoji = false`, the merge logic cannot tell whether the command-line value should win — both are the same zero value. The result: config file overrides a user-supplied `--no-emoji=false`, or the flag always wins and config is ignored.

**Why it happens:**
Go's `flag` package does not expose whether a flag was set; `flag.Lookup("no-emoji").Value.String()` returns `"false"` regardless of whether the flag appeared on the command line. There is no built-in `flag.IsSet()`. (Go issue #21226, open since 2017, marked "not planned" for the standard library.) Developers who check `if *noEmoji { ... }` silently implement "flag wins always" with no config merge.

**How to avoid:**
Use `flag.Visit` to build a set of flags that were explicitly provided, before falling back to config:
```go
setByUser := map[string]bool{}
flag.Visit(func(f *flag.Flag) { setByUser[f.Name] = true })

cfg := loadConfig()            // load TOML, defaults applied
noEmoji := cfg.NoEmoji         // config value is the base
if setByUser["no-emoji"] {
    noEmoji = *noEmojiFlag     // explicit CLI flag wins
}
```
This is the only stdlib-compatible pattern. Precedence order: explicit CLI flag > config file > compiled default.

**Warning signs:**
- `--no-emoji` on the command line has no effect when config file has `emoji = true`
- Config `emoji = false` is ignored when `--no-emoji` is not present on CLI
- Tests pass individually but fail when run with different flag combinations

**Phase to address:** Config loading phase — implement `flag.Visit` merge at startup before constructing `app.New()`.

---

### Pitfall 2: BurntSushi/toml Silently Ignores Unknown Keys by Default

**What goes wrong:**
A user adds a typo (`theem = "dark"`) or a future key (`accent_color = "#ff0000"`) to `~/.config/gsd-watch/config.toml`. BurntSushi/toml decodes it successfully, returns no error, and silently drops the unknown field. The user has no way to know their config key was ignored. If this is `emoji = true` misspelled as `emojis = true`, the feature appears broken.

**Why it happens:**
BurntSushi/toml's default mode is "loose": TOML values that have no corresponding struct field are silently dropped. There is no `DisallowUnknownFields` option (unlike `pelletier/go-toml/v2`). The `MetaData.Undecoded()` method exists but must be called explicitly — it is not automatic.

**How to avoid:**
After every `toml.Decode` call, check `md.Undecoded()` and log a warning (not an error — never crash on config) to stderr:
```go
md, err := toml.Decode(string(data), &cfg)
if err != nil {
    // parse error — use defaults, warn
}
if keys := md.Undecoded(); len(keys) > 0 {
    fmt.Fprintf(os.Stderr, "gsd-watch: unknown config keys: %v\n", keys)
}
```
This surfaces typos without breaking the user's session. Do NOT return an error — the app must start with defaults regardless of config problems.

**Warning signs:**
- User reports setting `theme = "minimal"` has no effect (maybe they typed `themes = "minimal"`)
- No error message shown despite config key mismatch
- Adding a new config key in a later version silently works on old config files with old keys

**Phase to address:** Config loading phase — add `md.Undecoded()` check immediately after decode.

---

### Pitfall 3: Package-Level `var` Styles Cannot Be Swapped at Runtime Without Global Mutation

**What goes wrong:**
`internal/tui/styles.go` currently declares `ColorGreen`, `ColorAmber`, `CompleteStyle`, `PendingStyle`, etc. as `var` at package scope. A naive theme implementation reassigns these package-level vars after config loads:
```go
// WRONG — global mutation, not concurrency-safe, affects all goroutines
tui.ColorGreen = lipgloss.AdaptiveColor{Light: "10", Dark: "10"}
```
This has two failure modes:
1. **Test pollution:** Any test that runs after a theme-mutating test sees the mutated globals. Tests become order-dependent and flaky.
2. **Concurrency unsafety:** `View()` reads package-level vars at render time; mutating them from outside the Bubble Tea event loop (e.g., at startup before `p.Run()`) is technically a data race under `-race` even if it happens before the first render.

**Why it happens:**
Lipgloss `Style` is a value type (safe to copy), but the *vars that hold the color constants* are global pointers to named values. Reassigning the var itself races with any reader. Developers coming from CSS or React theming assume "just swap the global" is safe.

**How to avoid:**
Do NOT mutate package-level style vars. Instead, define a `Theme` struct that holds all color/style values, construct it once at startup from the config, and pass it into `app.New()` alongside `noEmoji`. Sub-models receive the theme through their constructor or via an `Options` struct (matching the existing `tree.Options{NoEmoji: noEmoji}` pattern already in the codebase):
```go
type Theme struct {
    ColorGreen lipgloss.AdaptiveColor
    ColorAmber lipgloss.AdaptiveColor
    // ...
    CompleteStyle lipgloss.Style
    PendingStyle  lipgloss.Style
}

func DefaultTheme() Theme { ... }
func MinimalTheme() Theme { ... }
func HighContrastTheme() Theme { ... }
```
Keep the existing package-level vars as the `DefaultTheme()` values — they are still valid for tests that don't configure a theme. This is a zero-breaking-change approach.

**Warning signs:**
- `go test -race ./...` reports a race on `tui.ColorGreen` or any style var
- Changing theme in one test causes another test to see the wrong colors
- `styles.go` gains an `init()` function that reads config — a strong anti-pattern

**Phase to address:** Theme definition phase — define `Theme` struct before wiring any color changes; never mutate package vars.

---

### Pitfall 4: Theme Presets Using Hardcoded Colors Break Dark/Light Terminal Adaptivity

**What goes wrong:**
The existing codebase uses `lipgloss.AdaptiveColor{Light: "2", Dark: "2"}` throughout `styles.go`. This gives lipgloss two color values to choose from based on the detected terminal background (dark vs light). When writing new theme presets, developers switch to `lipgloss.Color("10")` (a non-adaptive single value) for brevity. Now `minimal` and `high-contrast` themes render with hardcoded colors that look wrong on one of the two terminal modes — for example, a gray meant for dark mode is invisible on a light terminal.

**Why it happens:**
`lipgloss.Color("10")` looks nearly identical to `lipgloss.AdaptiveColor{...}` in usage at a call site. The distinction is invisible until tested on a terminal with the opposite background. The 3-preset theme is implemented and tested on the developer's dark terminal, then ships broken for light-background users.

**How to avoid:**
All color values in all Theme preset constructors must use `lipgloss.AdaptiveColor` with explicit `Light` and `Dark` fields, never `lipgloss.Color`:
```go
// CORRECT — both modes handled
Green: lipgloss.AdaptiveColor{Light: "2", Dark: "10"}

// WRONG — breaks one terminal mode
Green: lipgloss.Color("10")
```
For the `default` theme, copy the exact values from the existing package-level vars. For `minimal` and `high-contrast`, choose `Light` and `Dark` values independently — do not assume the same ANSI index works on both backgrounds. Add a CI or manual verification step: run `gsd-watch --theme=minimal` on both a dark and light terminal before merging.

**Warning signs:**
- Theme preset constructors call `lipgloss.Color(...)` instead of `lipgloss.AdaptiveColor{...}`
- Theme only tested on one terminal background (dark or light)
- `minimal` theme text disappears or becomes unreadable on light-mode terminals
- `high-contrast` theme loses contrast on one of the two modes

**Phase to address:** Theme definition phase — enforce `AdaptiveColor` in all preset constructors; code review should reject any `lipgloss.Color(...)` usage in theme presets.

---

### Pitfall 5: Storing Theme in the Model Causes Re-render Drift When Theme Changes

**What goes wrong:**
If the `Theme` struct is stored as a field on `app.Model`, it becomes model state. The Bubble Tea architecture requires model state to only change inside `Update()`. But theme is loaded once at startup and never changes at runtime (there is no in-TUI settings panel in v1.3). Storing it in the model bloats the model snapshot and confuses the mental model of what "state" means.

A subtler failure: if theme is stored in model AND the tree/header/footer sub-models also each cache a copy, a theme-change message (even a future one) would require updating 4+ places in the model tree. If any copy is missed, the TUI renders with mixed themes.

**Why it happens:**
Developers default to "everything the model needs goes in the model." Render-only configuration (theme, noEmoji) feels like it should live there. This works but creates unnecessary coupling.

**How to avoid:**
Pass theme through sub-model constructors (same as `noEmoji` already does) and store it in the `Options` struct of each sub-model. For v1.3 (theme is immutable after startup), the sub-model constructors receive the theme once at `New()` time. This is consistent with how `noEmoji` is already threaded through `tree.Options`. No `ThemeChangedMsg` is needed for v1.3 scope.

If a future milestone adds live theme switching, introduce `ThemeChangedMsg` at that point — do not pre-engineer it now.

**Warning signs:**
- `app.Model` struct gains a `Theme` field next to `tree`, `header`, `footer` (it should live in Options, not top-level model state)
- Sub-models have their own `theme` field with no corresponding update path
- `helpView()` in `app/model.go` uses hardcoded `tui.ColorGray` instead of the theme's gray

**Phase to address:** Theme wiring phase — pass `Theme` through `Options` structs, not as model state.

---

### Pitfall 6: `os.UserConfigDir()` Returns `~/Library/Application Support` on macOS, Not `~/.config`

**What goes wrong:**
`os.UserConfigDir()` on macOS returns `/Users/<name>/Library/Application Support`. The target config path is `~/.config/gsd-watch/config.toml`. If the code uses `os.UserConfigDir()` naively to construct the config path, it reads from `~/Library/Application Support/gsd-watch/config.toml` instead — a path users will never find or create.

More specifically: Go deliberately chose the macOS convention over XDG on Darwin, and this was closed as "not planned" in Go issue #76320 (November 2025). Even if a user sets `XDG_CONFIG_HOME=/Users/name/.config`, `os.UserConfigDir()` ignores it on macOS.

**Why it happens:**
Developers assume `os.UserConfigDir()` is the portable "correct" way to find the config directory on all platforms. It is correct for the OS convention — just not the convention this project targets. The project targets CLI power users who expect `~/.config/` (the XDG convention used by most developer tools on macOS too).

**How to avoid:**
Hardcode the XDG path construction manually, with `XDG_CONFIG_HOME` fallback:
```go
func configDir() (string, error) {
    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
        return filepath.Join(xdg, "gsd-watch"), nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".config", "gsd-watch"), nil
}
```
Do NOT use `os.UserConfigDir()`. Document this explicitly in the config loading code.

**Warning signs:**
- Config file exists at `~/.config/gsd-watch/config.toml` but is never loaded
- No error is logged (the file at `~/Library/Application Support/gsd-watch/config.toml` simply doesn't exist, so missing-file path is taken silently)
- Works on Linux CI but not on developer's Mac

**Phase to address:** Config loading phase — the very first thing written in config path resolution.

---

### Pitfall 7: Missing Config File Must Silently Use Defaults (Never Error)

**What goes wrong:**
The loader calls `os.ReadFile(configPath)` and returns the error if the file doesn't exist. `main()` treats any config error as fatal and exits. A fresh install with no `~/.config/gsd-watch/config.toml` fails to start.

**Why it happens:**
Go's idiomatic error handling propagates errors up. Developers write `if err != nil { return err }` habitually. A missing optional config file is not an error.

**How to avoid:**
```go
data, err := os.ReadFile(configPath)
if errors.Is(err, os.ErrNotExist) {
    return DefaultConfig(), nil  // missing = use defaults, not an error
}
if err != nil {
    fmt.Fprintf(os.Stderr, "gsd-watch: cannot read config: %v\n", err)
    return DefaultConfig(), nil  // unreadable = use defaults, warn only
}
```
The function signature must be `func LoadConfig() Config` (not `(Config, error)`) to make it impossible for callers to treat config errors as fatal.

**Warning signs:**
- `gsd-watch` exits with "no such file or directory" on a fresh install
- Config loading returns `(Config, error)` — any error-returning function is a footgun
- Integration test creates a temp dir without a config file and panics

**Phase to address:** Config loading phase — define function signature before writing the body.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Mutate package-level style vars for themes | Fewer lines to thread theme through constructors | Flaky tests (order-dependent), data race under `-race`, impossible to unit test | Never |
| Use `os.UserConfigDir()` on macOS | One stdlib call instead of custom path logic | Reads wrong directory silently; user config never loaded | Never — it reads `~/Library/Application Support`, not `~/.config` |
| `return (Config, error)` from config loader | Idiomatic Go error signature | Callers can treat missing config as fatal; app fails to start on fresh install | Never for optional config |
| Store full `Theme` struct in `app.Model` state | Simple to access everywhere | Couples render-only config to model snapshot; complicates future live-theme switching | Acceptable if only one place stores it and sub-models receive values not the whole struct |
| Check `if *noEmojiFlag` without `flag.Visit` | One-line merge | Config emoji setting permanently ignored when `--no-emoji=false` | Never — the silent override will confuse users |
| Use `lipgloss.Color(...)` in theme presets | Fewer characters per color definition | Breaks dark/light adaptivity; `minimal`/`high-contrast` presets look wrong on one terminal mode | Never — always use `lipgloss.AdaptiveColor{Light: ..., Dark: ...}` |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| BurntSushi/toml + unknown keys | Ignore `md.Undecoded()` result | Always call `md.Undecoded()` and log warnings; never crash |
| BurntSushi/toml + missing file | Propagate `os.ErrNotExist` as error | Detect with `errors.Is(err, os.ErrNotExist)`; return defaults silently |
| `flag` + TOML config | Check `*flag == zero value` to detect "not set" | Use `flag.Visit` to build explicit-set map; CLI flag only wins when in that map |
| lipgloss styles + theme | Reassign package-level `var` at runtime | Construct `Theme` struct at startup; pass through `Options`; never mutate globals |
| lipgloss AdaptiveColor + theme presets | Use `lipgloss.Color(...)` for brevity | Use `lipgloss.AdaptiveColor{Light: ..., Dark: ...}` for every color in every preset |
| `os.UserConfigDir` + macOS | Assume it returns `~/.config` | Hardcode `$XDG_CONFIG_HOME` → `~/.config` fallback; do not call `os.UserConfigDir` |
| config path in tests | Use real `~/.config` in test | Use `t.TempDir()` + inject config path via function parameter or env var override |
| TOML theme value | Accept any string for `theme` field | Validate against known preset names after decode; warn and fall back to default |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Re-reading config file on every fsnotify event | Unnecessary disk I/O; config could change mid-session unexpectedly | Read config once at startup only; v1.3 has no live-reload | N/A for v1.3 scope |
| Constructing `lipgloss.NewStyle()` inside `View()` | Style objects allocated per render frame; GC pressure | Construct all styles once in `Theme` struct at startup; reuse in `View()` | High-frequency renders (e.g., rapid key presses) |
| Calling `ThemeFromName()` inside `View()` or `Update()` | Allocates new Theme struct on every render/update cycle | Resolve theme name once in `app.New()`; store resolved `Theme` in `Options` | Visible render latency during scrolling with large trees |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No warning for unknown config keys | User typos `theem = "dark"`, theme silently stays default, user assumes bug | Log `gsd-watch: unknown config key "theem"` to stderr on startup |
| No warning for invalid theme name | User sets `theme = "dracula"`, silently falls back to default | Log `gsd-watch: unknown theme "dracula", using "default"` to stderr |
| Help overlay doesn't mention config file path | User doesn't know config exists | Help overlay (`?`) must show the exact config file path (PROJECT.md v1.3 goal) |
| Config file created automatically on first run | User surprised by new file in `~/.config/` | Never auto-create config; only read if it exists |
| Theme looks right on dark terminal, broken on light | Light-mode users see invisible or low-contrast text | Test each preset on both dark and light terminal backgrounds before shipping |

---

## "Looks Done But Isn't" Checklist

- [ ] **Config missing file:** Verify `gsd-watch` starts correctly with no `~/.config/gsd-watch/config.toml` present — use a clean `t.TempDir()` in tests
- [ ] **Config unknown keys:** Verify a config file with `theem = "dark"` logs a warning but does not crash or fail silently
- [ ] **Flag precedence:** Verify `--no-emoji` on CLI overrides `emoji = true` in config; and that absence of `--no-emoji` allows config to set emoji mode
- [ ] **Config path on macOS:** Verify config is read from `~/.config/gsd-watch/config.toml`, NOT `~/Library/Application Support/gsd-watch/config.toml`
- [ ] **Theme mutation:** Verify `go test -race ./...` passes after theme wiring; no mutations to package-level style vars
- [ ] **Theme fallback:** Verify invalid `theme = "unknown"` falls back to `default` with a warning, not a crash
- [ ] **Help overlay config path:** Verify `?` overlay shows the config file path (exact path, not just the directory)
- [ ] **Theme sub-model reach:** Verify header, footer, tree, and help overlay all respect the active theme (no hardcoded color references remaining in their render paths)
- [ ] **AdaptiveColor in presets:** Verify every color in `MinimalTheme()` and `HighContrastTheme()` uses `lipgloss.AdaptiveColor{Light: ..., Dark: ...}`, not `lipgloss.Color(...)`
- [ ] **Light terminal test:** Verify all three theme presets are readable on a light-background terminal (iTerm2 or Terminal.app with light profile)

---

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| `os.UserConfigDir()` used instead of XDG path | LOW | Change the one call site in `config.go`; re-run tests. No data migration needed — user config simply was never loaded from the wrong path |
| Package-level vars mutated for theme | MEDIUM | Extract `Theme` struct (see Pitfall 3 pattern); mechanically replace ~10-15 call sites in `view.go`; fix any failing race tests. All changes are in `styles.go` and `tree/view.go` |
| `(Config, error)` signature already in callers | LOW | Change signature to `Config` only; move error handling inside the function; update the 1-2 call sites in `main.go` |
| flag/config merge broken — `flag.Visit` missing | LOW | Add `flag.Visit` block to `main.go` after `flag.Parse()`; no other files change |
| Hardcoded `lipgloss.Color(...)` in theme presets | LOW | Replace each instance with `lipgloss.AdaptiveColor{Light: ..., Dark: ...}`; confined to theme preset constructors in `styles.go` |
| Config tests touching real `~/.config/` | LOW | Add `path string` parameter to `LoadConfig()`; update tests to use `t.TempDir()`; takes under an hour |
| Theme stored in `app.Model` state | MEDIUM | Move `Theme` out of model fields into `Options` struct; thread through `SetOptions()` calls; risk is limited to `app/model.go` and `tree/model.go` |

---

## Testing Pitfalls

### Pitfall: Config Tests That Touch Real `~/.config/`

**What goes wrong:**
Config loading tests call `LoadConfig()` which reads from the real `~/.config/gsd-watch/config.toml` on the developer's machine (or CI). Tests pass locally because the developer has a config file, fail on CI where the path doesn't exist, or — worse — tests mutate the developer's real config.

**How to avoid:**
Extract the config path as a parameter. The loader must accept a path, not compute it internally:
```go
func LoadConfigFrom(path string) Config    // used by tests
func LoadConfig() Config {                 // used by main — resolves path then calls LoadConfigFrom
    return LoadConfigFrom(configFilePath())
}
```
In tests:
```go
func TestLoadConfig_Theme(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "config.toml")
    os.WriteFile(path, []byte(`theme = "minimal"`), 0600)
    cfg := LoadConfigFrom(path)
    // assert
}
```
`t.TempDir()` is automatically cleaned up after the test. Never use `os.TempDir()` directly (no automatic cleanup). Never hardcode paths inside the loader.

**Warning signs:**
- Config tests pass on developer machine, fail in CI (CI has no config file)
- Test output changes depending on developer's personal `~/.config/gsd-watch/config.toml`
- `LoadConfig()` has no injectable path parameter

**Phase to address:** Config loading phase — define the function signature as `LoadConfigFrom(path string)` from the start.

---

### Pitfall: Theme Tests That Depend on Package-Level Style Var State

**What goes wrong:**
A test for the `minimal` theme calls a function that mutates `tui.PendingStyle`. A later test for the `default` theme expects the original `tui.PendingStyle` value. The tests fail in any order and are very hard to debug because the failure is in the assertion, not the mutation.

**How to avoid:**
Theme tests must operate on `Theme` struct instances, not package-level vars:
```go
func TestMinimalTheme_PendingColor(t *testing.T) {
    theme := MinimalTheme()
    // assert on theme.PendingStyle, not tui.PendingStyle
}
```
Package-level vars in `styles.go` should be treated as read-only constants for tests. If a test needs to verify that `View()` uses the right theme colors, pass the theme through `Options` and assert on the rendered output.

**Phase to address:** Theme definition phase — enforce the `Theme` struct approach before any test is written.

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Zero value vs not-set in flag/config merge | Config loading phase | `flag.Visit` pattern present in code; test with explicit `--no-emoji=false` vs absent |
| Unknown TOML keys silently dropped | Config loading phase | `md.Undecoded()` check in loader; test with unknown key logs warning |
| Package-level style var mutation | Theme definition phase | `go test -race` green; no `tui.ColorX = ...` assignments anywhere |
| Hardcoded `lipgloss.Color` in theme presets breaks adaptivity | Theme definition phase | Code review: no `lipgloss.Color(...)` in preset constructors; manual light-terminal test |
| Theme stored as model state (wrong layer) | Theme wiring phase | `app.Model` struct has no `Theme` field; theme only in `Options` |
| `os.UserConfigDir()` on macOS wrong dir | Config loading phase | Unit test with `XDG_CONFIG_HOME` set and unset; verify `~/Library/...` never used |
| Missing config file treated as error | Config loading phase | Test with nonexistent path returns defaults, no error |
| Config tests touching real `~/.config/` | Config loading phase | All config tests use `t.TempDir()` + `LoadConfigFrom(path)` |
| Theme tests depending on global var state | Theme definition phase | Tests operate on `Theme` struct instances; no package-var assignment in tests |
| `ThemeFromName()` called per render frame | Theme wiring phase | `ThemeFromName()` call only in `app.New()`; `View()` and `Update()` never call it |

---

## Sources

- [BurntSushi/toml pkg.go.dev — Decode, MetaData.Undecoded()](https://pkg.go.dev/github.com/BurntSushi/toml)
- [Go issue #21226 — flag.IsSet proposal (not planned)](https://github.com/golang/go/issues/21226)
- [Go issue #76320 — os.UserConfigDir should respect XDG_CONFIG_HOME on Darwin (closed: not planned, Nov 2025)](https://github.com/golang/go/issues/76320)
- [flag.Visit pattern for detecting explicitly set flags — golang/go discussion](https://github.com/golang/go/issues/21226)
- [lipgloss AdaptiveColor — charmbracelet/lipgloss v1.1.0](https://pkg.go.dev/github.com/charmbracelet/lipgloss@v1.1.0)
- [lipgloss Style is a value type (safe copy semantics) — charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- [lipgloss compat package global-var impurity acknowledgment — lipgloss v2 docs](https://pkg.go.dev/github.com/charmbracelet/lipgloss/v2/compat)
- [t.TempDir() automatic cleanup — Go testing package docs](https://pkg.go.dev/testing#T.TempDir)
- [Bubble Tea model context and render-only state — bubbletea issue #1010](https://github.com/charmbracelet/bubbletea/issues/1010)
- [Rob Pike on flag default-vs-set detection — golang/go issue #21226 comment](https://github.com/golang/go/issues/21226)

---
*Pitfalls research for: v1.3 TOML config + theme system added to existing Go Bubble Tea TUI*
*Researched: 2026-03-26*
