# Project Research Summary

**Project:** gsd-watch v1.3 — Settings Milestone (Config File + Theme Presets)
**Domain:** TOML config + named theme presets added to an existing Go Bubble Tea TUI
**Researched:** 2026-03-26
**Confidence:** HIGH

## Executive Summary

gsd-watch v1.3 is a focused settings milestone on an already-shipping Go TUI binary. The work is not greenfield — it extends an existing Bubble Tea v1.x application with two new capabilities: a TOML config file that persists appearance preferences (`~/.config/gsd-watch/config.toml`), and a three-preset theme system (`default`, `minimal`, `high-contrast`) that replaces hardcoded color constants in the tree view. All architecture research was conducted against the real codebase (2026-03-26), so there is no speculation about the current state. The one new external dependency is `github.com/BurntSushi/toml@v1.6.0` — everything else uses stdlib or existing charmbracelet packages already in `go.mod`.

The recommended build order is three phases ordered strictly by hard dependencies: config infrastructure first (new `internal/config/` leaf package, zero visual change to confirm correctness), theme wiring second (the main refactor touching `styles.go` and `tree/view.go`), and help overlay addition third (pure string addition with no structural risk). Total change surface is bounded: one new package, five modified files, and approximately 10–15 call-site replacements in `tree/view.go`.

The primary risk area is flag-vs-config merge for the existing `--no-emoji` flag. Go's `flag` package cannot natively detect whether a boolean flag was explicitly set by the user, so a naive merge silently ignores either the flag or the config. The fix is a `flag.Visit` pattern documented with exact code in the pitfalls research. The second high-probability silent failure is `os.UserConfigDir()` on macOS — it returns `~/Library/Application Support` rather than `~/.config`, and Go issue #76320 was closed "not planned" in November 2025. Both risks have exact, tested mitigations.

## Key Findings

### Recommended Stack

The v1.3 stack adds exactly one new dependency to a stable Bubble Tea v1.3.10 / Lip Gloss v1.1.0 / Bubbles v1.0.0 / fsnotify v1.9.0 core. BurntSushi/toml@v1.6.0 was selected over pelletier/go-toml v2 because the config file has exactly two keys read once at startup — parse performance is irrelevant, BurntSushi has 10x the adoption (37K+ importers vs ~1.7K), a simpler single-call API (`toml.DecodeFile`), zero transitive dependencies, and the correct default behavior for a config file that will evolve (unknown keys silently ignored, `md.Undecoded()` available for explicit warnings).

**Core technologies:**
- `github.com/BurntSushi/toml@v1.6.0`: TOML config parsing — only new dependency; minimal API, zero transitive deps, unknown-keys-ignored by default
- `lipgloss.AdaptiveColor` (existing, `lipgloss@v1.1.0`): theme color representation — `Theme` struct uses the same type as current `styles.go` vars; no new API surface needed
- `os.UserHomeDir()` + manual XDG path join (stdlib only): config path resolution — bypasses `os.UserConfigDir()` which returns the wrong macOS directory
- `flag.Visit` (stdlib): explicit-flag detection — the only stdlib-compatible way to distinguish "user passed this flag" from "flag defaulted to false"

**What not to add:**
- `os.UserConfigDir()` — returns `~/Library/Application Support` on macOS; Darwin-specific behavior confirmed against Go stdlib source and issue #76320 (closed Nov 2025)
- Third-party XDG libraries (adrg/xdg, kyoh86/xdg) — 4 lines of stdlib handle the only case needed for this macOS-only tool
- Theme hot-reload, in-TUI settings panel, per-color overrides — explicitly deferred to v1.4+ per PROJECT.md

### Expected Features

All features below are scoped to v1.3. Competitor analysis of lazygit, k9s, and gitui confirms the table-stakes behaviors — these are the industry-standard contracts users expect from any Go CLI/TUI tool that introduces a config file.

**Must have (table stakes):**
- Missing config file = silent defaults — every serious Go CLI tool (gh, k9s, lazygit) boots normally when config is absent; failing on a missing optional config is user-hostile
- Malformed TOML = fatal error with file path and parse detail — when the file exists but is invalid, the user made a mistake that needs actionable feedback
- Unknown config keys = warning to stderr, never a crash — the upgrade path depends on this; old binaries must not fail on new config keys they don't recognize
- Flag overrides config (`--no-emoji` wins over `emoji = false` in config) — the universal CLI precedence contract: flag > config > default
- Config path shown in `?` help overlay — closes the "where does the config live?" question without requiring docs
- Three theme presets: `default` (current colors, zero visual change for existing users), `minimal` (no status color, content-first), `high-contrast` (bold + 16-color ANSI, SSH-compatible)
- Theme name validated at config load time — unknown name warns to stderr and falls back to `default`; never crash

**Should have (differentiators):**
- `high-contrast` theme using only ANSI palette indices 1–15 — most tools only test truecolor; SSH and 8-color terminal users benefit from explicit ANSI choices
- `--help` output includes a copy-pasteable TOML example block — lowers barrier to first config without auto-creating files
- `config.ConfigPath()` exported from config package — single source of truth for both the loader and the help overlay display

**Defer to v1.4+:**
- In-TUI settings panel — 3–5x the work of the entire v1.3 milestone; explicitly deferred in PROJECT.md
- Per-color overrides in config — named presets cover 95% of needs; source is available for the other 5%
- Config hot-reload — restart is 100ms; complexity cost is not justified for this audience
- `--theme` CLI flag as alternative to config — redundant with the config `theme` key; flags are awkward to pass via slash command invocation
- Environment variable overrides (`GSD_WATCH_THEME=...`) — adds a third resolution layer most users will never use

### Architecture Approach

The architecture is additive and non-breaking. A new leaf package `internal/config/` handles all config file concerns (path resolution, TOML decode, defaults). It imports nothing from `internal/tui/`, preventing import cycles. Config is loaded in `main.go` before `app.New()` — keeping model constructors error-free and tests independent of real disk paths. Theme resolution happens once in `app.New()` by calling `tui.ThemeFromName(cfg.Theme)`, and the resulting `tui.Theme` value flows through `tree.Options` into render functions, matching the existing `noEmoji bool` threading pattern exactly. Package-level style vars in `styles.go` are preserved untouched — header and footer components are intentionally not migrated in v1.3 to limit blast radius.

**Major components:**
1. `internal/config/config.go` (NEW) — `Config` struct, `LoadConfigFrom(path string) Config`, `LoadConfig() Config`, `ConfigPath() string`. Leaf package; stdlib + BurntSushi/toml only. No error return — missing/invalid config always produces defaults plus a warning.
2. `internal/tui/styles.go` (ADDITIVE) — `Theme` struct with `lipgloss.Style` fields; `DefaultTheme()`, `MinimalTheme()`, `HighContrastTheme()`, `ThemeFromName(name string) Theme` constructors. Existing package-level vars untouched.
3. `internal/tui/tree/model.go` (MODIFIED) — `Options.Theme tui.Theme` added alongside existing `NoEmoji bool`.
4. `internal/tui/tree/view.go` (MODIFIED) — approximately 10–15 `tui.XxxStyle.Render()` package-var references replaced with `opts.Theme.XxxStyle.Render()`. Exported archive function signatures (`RenderArchiveRow`, `RenderArchiveSeparator`, `RenderArchiveZone`) gain a `theme tui.Theme` parameter.
5. `internal/tui/app/model.go` (MODIFIED) — `New()` signature changes from `(events, noEmoji bool)` to `(events, cfg config.Config)`; resolves theme once via `ThemeFromName`; passes to tree via `SetOptions`.
6. `cmd/gsd-watch/main.go` (MODIFIED) — calls `config.Load()` after `flag.Parse()`; applies `flag.Visit`-based flag-over-config merge; passes `cfg` to `app.New()`.

**Unchanged:** `internal/tui/header/model.go`, `internal/tui/footer/model.go`, `internal/parser/`, `internal/watcher/`.

### Critical Pitfalls

1. **`flag.Visit` is required for boolean flag/config merge** — `*noEmojiFlag == false` is indistinguishable from "flag not set" using Go's `flag` package (issue #21226, closed "not planned"). Use `flag.Visit` to build a `setByUser` map; CLI flag only wins when it appears in that map. Without this, either the flag silently ignores config or config silently ignores the flag.

2. **`os.UserConfigDir()` returns the wrong path on macOS** — It returns `~/Library/Application Support`, not `~/.config`. Go issue #76320 closed "not planned" in November 2025. Always use `os.UserHomeDir()` + `filepath.Join(home, ".config", "gsd-watch")` with `XDG_CONFIG_HOME` env var fallback. Never call `os.UserConfigDir()`.

3. **Package-level style var mutation is never correct for theming** — Reassigning `tui.ColorGreen = ...` at runtime creates a data race under `go test -race` and makes tests order-dependent. Prevention: never mutate package vars. Define a `Theme` struct, construct it once at startup from config, pass through `Options`. Existing package-level vars stay as-is for components not yet migrated.

4. **`lipgloss.Color()` in theme presets breaks dark/light terminal adaptivity** — `lipgloss.Color("10")` is non-adaptive and will look wrong on one terminal background mode. Every color in every preset constructor must use `lipgloss.AdaptiveColor{Light: "...", Dark: "..."}`. This is invisible until tested on the opposite background and has no compile-time signal.

5. **Missing config file must not be an error** — Function signature must be `func LoadConfig() Config` with no error return. `os.ErrNotExist` is handled inside the loader and produces silent defaults. If the signature is `(Config, error)`, callers may accidentally treat the missing-file case as fatal, breaking fresh installs.

6. **Config tests must not touch the real `~/.config/`** — Use `t.TempDir()` + `LoadConfigFrom(path string)` in all tests. A loader without an injectable path parameter will fail on CI (no config file) or silently depend on the developer's personal config, making tests non-deterministic.

## Implications for Roadmap

Based on the dependency graph confirmed by direct codebase analysis, the v1.3 milestone maps cleanly to three phases. Each phase is an independent safe stopping point verifiable before the next begins.

### Phase 1: Config Infrastructure

**Rationale:** `internal/config/` is the dependency for Phase 2 (theme wiring requires `cfg.Theme string`) and Phase 3 (help overlay requires `config.ConfigPath()`). This phase produces zero visual change — identical binary behavior to v1.2 is the acceptance criterion. Any config correctness bugs discovered here are cheaper to fix than after theme wiring is layered on top.

**Delivers:** New `internal/config/` package with `Config` struct, `LoadConfigFrom`, `LoadConfig`, `ConfigPath`; updated `app.New()` signature accepting `config.Config`; `flag.Visit`-based flag/config merge in `main.go`; unit tests for all edge cases (missing file, malformed TOML, unknown keys, emoji flag override, TOML `emoji = false` to `NoEmoji = true` inversion).

**Addresses (from FEATURES.md):** Missing file = silent defaults; malformed TOML = fatal error with path; unknown keys = stderr warning; flag overrides config.

**Avoids (from PITFALLS.md):** Flag zero-value confusion (flag.Visit pattern); `os.UserConfigDir()` wrong macOS directory (manual XDG path); missing config treated as error (no-error function signature); config tests touching real `~/.config/` (injectable path via `LoadConfigFrom`).

**Research flag:** No deeper research needed. All patterns are fully specified in STACK.md and PITFALLS.md with exact code.

### Phase 2: Theme System

**Rationale:** Depends on Phase 1 (`config.Config.Theme` must exist). Highest impact and highest change count — touches four files. The refactor is mechanical (replace package-var references with struct-field references at ~15 sites) but requires careful handling of exported archive function signatures (`RenderArchiveRow`, `RenderArchiveSeparator`, `RenderArchiveZone` are tested from external test packages and callers must be updated atomically with the signature change). Header and footer stay on package-level vars to limit scope.

**Delivers:** `Theme` struct and three named presets in `styles.go`; `tree.Options.Theme` field; call-site replacements in `tree/view.go`; updated exported archive function signatures; `app.New()` resolves theme once via `ThemeFromName`; golden test confirming `DefaultTheme` renders identically to pre-v1.3 output.

**Addresses (from FEATURES.md):** All three theme presets; theme name validation (unknown name warns and falls back); `high-contrast` with 16-color ANSI palette for SSH/degraded terminal compatibility.

**Avoids (from PITFALLS.md):** Package-level var mutation (`Theme` struct, never mutate globals); `lipgloss.Color()` in presets (all colors use `AdaptiveColor`); theme stored as model state (theme goes in `Options`, resolved once in `app.New()`); `ThemeFromName()` called per render frame (called once at construction time only).

**Research flag:** No additional research needed. ARCHITECTURE.md enumerates all modified files, the exact integration path, and the ~10–15 call-site count in `view.go`.

### Phase 3: Help Overlay Config Hint

**Rationale:** Depends on Phase 1 (`config.ConfigPath()` must exist). Pure addition to `helpView()` in `app/model.go` — no structural risk. Phase 1 already adds `internal/config` to `app/model.go`'s import graph, so no new imports are needed.

**Delivers:** `?` overlay shows the exact config file path and current theme name; `--help` output includes a copy-pasteable TOML example block; "restart to apply config changes" note in the overlay.

**Addresses (from FEATURES.md):** Config path in help overlay (P1, PROJECT.md requirement); `--help` example TOML block (P2).

**Avoids (from PITFALLS.md):** Config path hardcoded in multiple files — `config.ConfigPath()` is the single source of truth; no string literal duplication.

**Research flag:** No research needed. This is a string addition to an existing pure function.

### Phase Ordering Rationale

- Phase 1 must be first: `Config` struct is the shared type dependency for Phase 2 (theme name input to `ThemeFromName`) and Phase 3 (config path output to help overlay).
- Phase 2 before Phase 3: theme wiring has the highest blast radius. Stabilizing it before Phase 3 means the help overlay snapshot test can assert on the final rendered output, not an intermediate state.
- Header and footer intentionally deferred: they each have 3–4 color references and no existing `Options` struct. The change-count vs. visual benefit ratio is unfavorable for v1.3. Address in v1.4.
- Exported archive function signatures (`RenderArchiveRow`, `RenderArchiveSeparator`, `RenderArchiveZone`) must be updated atomically with their external test file callers in Phase 2 — grep for these names in `*_test.go` before changing signatures.

### Research Flags

Phases with standard patterns — no `research-phase` invocation needed:

- **Phase 1 (Config):** All edge cases fully documented with exact code patterns. BurntSushi/toml API verified. XDG path behavior verified against Go stdlib source and issue #76320. `flag.Visit` pattern verified against issue #21226 with exact implementation.
- **Phase 2 (Theme):** Exact `styles.go` additions specified with struct definitions and constructor signatures. Call-site count confirmed by direct `view.go` analysis in ARCHITECTURE.md. `AdaptiveColor` and `Style` value-type semantics verified against lipgloss v1.1.0 source.
- **Phase 3 (Help Overlay):** Pure string addition to an existing pure function. No API or behavior research needed.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified against pkg.go.dev. BurntSushi/toml v1.6.0 confirmed Dec 2025. No conflicts with existing charmbracelet packages. go.mod read directly — only missing dependency is BurntSushi/toml. |
| Features | HIGH | Competitor analysis (lazygit, k9s, gitui) validates expected behaviors. All P1 features are low-complexity. Feature boundaries from PROJECT.md are authoritative. Anti-feature rationale grounded in specific cost/complexity evidence. |
| Architecture | HIGH | Based on direct codebase analysis — no speculation. Exact files, function signatures, and call-site counts provided. Import graph verified cycle-free. Anti-patterns grounded in codebase-specific observations. |
| Pitfalls | HIGH | Three of six critical pitfalls reference specific, verified Go issues (#76320, #21226). TOML `Undecoded()` behavior verified against official BurntSushi/toml docs. Package-var mutation race verified against lipgloss source. `AdaptiveColor` adaptivity loss is observable behavior, not inference. |

**Overall confidence:** HIGH

### Gaps to Address

- **`RenderArchiveRow` / `RenderArchiveSeparator` / `RenderArchiveZone` external test callers:** These functions are exported and called from external test packages. The exact test file paths were not enumerated during research. Before changing their signatures in Phase 2, grep for these names in all `*_test.go` files and update callers atomically.
- **Header and footer theme coverage:** Both components stay on package-level vars in v1.3. `minimal` and `high-contrast` themes will affect the tree but not the header progress bar or footer key hints. This is an accepted, documented partial-theme state — address in v1.4.
- **Light-terminal verification:** All three theme presets must be manually verified on a light-background macOS terminal (iTerm2 or Terminal.app with a light profile) before Phase 2 closes. This cannot be automated in unit tests. The `AdaptiveColor` requirement mitigates the risk, but manual verification is the only way to confirm light-mode rendering.
- **`helpView` signature decision:** Currently `helpView(width int, noEmoji bool)`. Adding the config path can be done by either (a) calling `config.ConfigPath()` directly inside the function (no signature change, no new import after Phase 1), or (b) adding a `cfgPath string` parameter. Both are clean; decide at implementation time.

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/BurntSushi/toml` — DecodeFile, MetaData.Undecoded(), unknown-key default behavior
- `pkg.go.dev/github.com/charmbracelet/lipgloss@v1.1.0` — AdaptiveColor, Style value type, style composition
- `pkg.go.dev/github.com/charmbracelet/bubbletea@v1.3.10` — Program, Send(), WindowSizeMsg, Init/Update/View contract, v1 vs v2 API differences
- `pkg.go.dev/github.com/charmbracelet/bubbles@v1.0.0` — viewport, key, progress, spinner components
- `pkg.go.dev/github.com/fsnotify/fsnotify@v1.9.0` — macOS kqueue backend, no recursive watching confirmed
- `go doc os.UserConfigDir` — Darwin returns `$HOME/Library/Application Support`, confirmed against stdlib source
- Go issue #76320 — `os.UserConfigDir` XDG_CONFIG_HOME on Darwin: closed "not planned," November 2025
- Go issue #21226 — `flag.IsSet` proposal: closed "not planned"; `flag.Visit` pattern documented in thread
- Direct codebase analysis (2026-03-26): `internal/tui/styles.go`, `app/model.go`, `tree/model.go`, `tree/view.go`, `header/model.go`, `footer/model.go`, `cmd/gsd-watch/main.go`, `go.mod`
- `.planning/PROJECT.md` — config path, key names, theme names, scope constraints (authoritative source of requirements)

### Secondary (MEDIUM confidence)
- lazygit Config.md — YAML config structure, missing/unknown key handling, theme color property patterns
- k9s skins documentation — skin file format, `k9s info` config path display, fallback-to-stock behavior
- gitui THEMES.md — RON theme file format, XDG location, `-t` flag override pattern
- Go `flag` package behavior — pointer-flag pattern as workaround for "was this flag set?"; consistent across Viper, gh CLI, and multiple Go configuration guides
- lipgloss v2 compat package acknowledgment of global-var impurity — confirms package-var mutation as an anti-pattern even in lipgloss's own ecosystem

---
*Research completed: 2026-03-26*
*Ready for roadmap: yes*
