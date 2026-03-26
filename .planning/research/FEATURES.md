# Feature Research

**Domain:** Config file + theme system for an existing Go TUI binary (gsd-watch v1.3 Settings milestone)
**Researched:** 2026-03-26
**Confidence:** HIGH — patterns are well-established across lazygit, k9s, gitui, and gh CLI; verified against BurntSushi/toml docs and lipgloss v1 source

---

> **Scope note:** This file replaces the original v1.0 FEATURES.md and focuses exclusively on the v1.3 Settings milestone additions. Existing features (--no-emoji, AdaptiveColor, keyboard nav, file watching) are already shipped and are referenced only where they create dependencies.

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users expect from any Go CLI/TUI tool that introduces a config file. Missing these makes the tool feel amateur or hostile.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Missing config file = silent defaults | Every serious Go tool (gh, k9s, lazygit) boots normally when config is absent. Users should never see an error on first launch before they've created any config | LOW | `os.IsNotExist` check; return defaults struct, no log output. This is the single most important behavior to get right. |
| Malformed TOML = fatal error with clear message | When the file exists but is invalid TOML, the user made a mistake that needs fixing. Silent ignore is worse than failing (wrong behavior with no feedback). k9s and gh both exit with parse error + file path | LOW | `toml.Decode` returns error; print `config error: ~/.config/gsd-watch/config.toml: <parse detail>`; `os.Exit(1)`. Never silently ignore a corrupted config. |
| Unknown keys = warn, do not fail | Users copy-paste config snippets, rename keys when upgrading, or make typos. Crashing on unknown keys destroys the upgrade path. BurntSushi/toml silently ignores unknown keys by default — this is the correct behavior for user config files | LOW | Use `toml.Decode` default (lenient). Optionally log unknown keys at `--debug` level only. Do NOT use `Undecoded()` to error. |
| Config path documented in --help | Users ask "where does the config live?" before reading any README. k9s shows config paths in `k9s info`. The `?` help overlay is the natural home for this in gsd-watch | LOW | `--help` output and `?` overlay both print the resolved config path. Use `os.UserConfigDir()` (stdlib, returns `~/Library/Application Support` on macOS) or hardcode `~/.config/gsd-watch/config.toml` per PROJECT.md spec. |
| Flag overrides config | The standard precedence for Go CLI tools is: flag > config > default. When a user passes `--no-emoji`, it must win over `emoji = true` in config. Users expect explicit flags to be authoritative | LOW | In main.go: if `--no-emoji` was explicitly set (check `flag.Lookup("no-emoji").Value.String() != default`), use flag value; otherwise use config value. |
| Config changes take effect on next launch | Hot-reload of config during a running session is a significant complexity jump (fsnotify on the config file, propagate to all submodels). No TUI in this class does it. Users accept that config changes apply after restart | LOW (no-op) | Do NOT implement config file watching. Document "restart to apply" in the `?` overlay. |
| No config directory creation on startup | Tools that silently `mkdir ~/.config/gsd-watch/` during first launch are considered presumptuous. Create the directory only when the user explicitly invokes a "create config" action (out of v1.3 scope) | LOW | Only read; never write. If the directory doesn't exist, treat it the same as a missing file. |

### Differentiators (Competitive Advantage)

Features that give gsd-watch a better config + theme experience than the average small TUI tool.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Named theme presets (not raw colors) | Raw color customization (lazygit, k9s) requires the user to know ANSI color names or hex values. Named presets (`default`, `minimal`, `high-contrast`) work immediately with zero color knowledge | LOW | A Go `Theme` struct with lipgloss.AdaptiveColor fields; a `ThemeByName(string) Theme` lookup function. Three presets to implement. |
| `high-contrast` theme usable via SSH + low-color terminals | Developers SSHing into remote machines often hit 8-color or 16-color terminals where the default theme looks washed out. A high-contrast preset with bold + 16-color ANSI palette covers this case without needing `--no-emoji` | MEDIUM | Use lipgloss color index values 1-15 (standard 16-color ANSI); avoid 256-color or truecolor. Combine with `--no-emoji` for full SSH compatibility. |
| `minimal` theme: content only, no decorative color | Users who prefer monochromatic terminals or who find colors distracting want the tree to be visually quiet. Minimal = no color differentiation on status, just structural indentation | LOW | Most styles render with `lipgloss.NoColor` / empty foreground. Status icons and badges still render but without color. |
| Config path shown in `?` help overlay | Closes the "where do I put it?" question without the user having to find docs. k9s has `k9s info` for this; gsd-watch surfaces it directly in the already-open help overlay | LOW | One line added to the help overlay render function. Requires the help overlay to be built (v1.3 scope). |
| Documented TOML example in --help output | Lowers barrier to first config. Print a commented example config block in `--help` output so the user can copy-paste and modify | LOW | Static string appended to --help. No file I/O. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Per-color overrides in config (`[colors] complete = "#00ff00"`) | Power users want exact brand colors | Requires color validation, 256-color vs truecolor detection, documentation for each field, migration when fields are renamed. k9s has this and its skin files are complex YAML with 30+ fields. Overkill for a personal tool | Named presets cover 95% of needs. Source code is available for the 5% case. |
| Config file auto-creation with defaults | "Teach me the options by showing me a populated file" | Writes to the user's filesystem without consent; presupposes `~/.config/gsd-watch/` exists; creates surprise files in `ls ~/.config/`. gh CLI does NOT do this. | `--help` prints an example config. The `?` overlay shows the config path. User creates the file manually when they want it. |
| In-TUI settings panel | Visual discovery of config options | Requires a full modal UI, cursor management in the settings panel, applying changes live vs on-save, and writing the config file back to disk. This is 3-5x the work of everything else in v1.3 combined | Explicitly deferred to v1.4+ in PROJECT.md. Document as future work in `?` overlay. |
| Config hot-reload (live apply without restart) | "I want to tweak theme without restarting" | Requires fsnotify on the config file path, a settings-changed tea.Msg, propagating Theme to every submodel, and handling the race between a file write and the next render. No small TUI tool does this | Restart is 100ms. "Restart to apply" is a fine UX for a settings-heavy change. |
| Environment variable overrides (`GSD_WATCH_THEME=minimal`) | Scripting / CI / per-project overrides | Adds a third resolution layer (flag > env > config > default) that most users never need. Env vars are invisible and hard to debug. The target users (the author + small friend group) will not use this | Config file covers per-machine preferences. Flags cover per-invocation overrides. |
| `--theme` flag as alternative to config | Per-invocation theme switching | Adds redundancy with the config `theme` key without adding value. The slash command invocation makes flags awkward to pass. One config file is cleaner | Config file is the right place for stable preferences. |
| Validating that a theme name is correct at parse time vs at render time | "Better error messages" | Both approaches are valid; parse-time validation is slightly better UX. However, it requires a static list of valid names that must be kept in sync. For three presets, the risk is low either way | Validate at parse time with a clear error: `unknown theme "dracula"; valid themes: default, minimal, high-contrast`. |

## Feature Dependencies

```
[--no-emoji flag (existing)]
    └──defers-to──> [config emoji key (NEW)]
                        └──requires──> [Config loader (NEW)]
                                           └──reads──> [~/.config/gsd-watch/config.toml]

[Theme preset (NEW)]
    └──requires──> [Config loader (NEW)]
    └──requires──> [Theme struct with lipgloss.AdaptiveColor fields (NEW)]
                       └──replaces──> [package-level vars in tui/styles.go (existing)]
                       └──threaded-through──> [app.New() constructor (existing)]
                                                  └──propagates-to──> [header, tree, footer submodels]

[? help overlay (NEW)]
    └──displays──> [config file path]
    └──displays──> [current theme name]
    └──requires──> [Config struct accessible in app model]

[high-contrast theme (NEW)]
    └──pairs-well-with──> [--no-emoji (existing)]
    └──independent-of──> [--debug (existing)]
```

### Dependency Notes

- **Config loader is the single new foundation:** Everything else in v1.3 (theme, emoji config, help overlay path display) flows from reading and validating the config file once at startup.
- **Theme struct replaces package-level style vars:** Currently `tui/styles.go` has package-level `var ColorGreen = lipgloss.AdaptiveColor{...}`. The Theme struct approach moves these into a passed value, enabling different presets. This is a refactor of existing code, not net-new code.
- **Flag precedence requires knowing if a flag was explicitly set:** Go's `flag` package doesn't distinguish "user passed --no-emoji" from "no-emoji defaulted to false." A boolean flag default of `false` is indistinguishable from "not passed." Solution: use a pointer flag (`flag.Bool`) where `nil` means "not provided" and `false` means "explicitly passed as false" — or accept that `--no-emoji` is a one-way override (false by default, can only be set to true from flag, config can set it true without flag).
- **Theme propagation to submodels:** Currently `noEmoji bool` is passed into `app.New()`. The Theme struct follows the same pattern — one constructor argument, stored on the app model, passed to submodel render functions. This is already the right architecture.

## MVP Definition

### Launch With (v1.3)

Minimum set to deliver "configure appearance once via a file."

- [ ] Config loader: reads `~/.config/gsd-watch/config.toml`; missing = silent defaults; malformed = fatal error with file path; unknown keys = silently ignored
- [ ] Config struct with two keys: `emoji` (bool, default true) and `theme` (string, default "default")
- [ ] Flag-over-config precedence: `--no-emoji` flag wins over `emoji = false` in config
- [ ] Theme struct: lipgloss.AdaptiveColor fields replacing package-level vars in styles.go
- [ ] Three theme presets: `default` (current colors), `minimal` (no status color), `high-contrast` (bold + 16-color ANSI)
- [ ] Theme validation at config load time: unknown theme name = fatal error with valid options listed
- [ ] `?` help overlay showing config file path and current theme name
- [ ] `--help` output showing TOML example

### Add After Validation (v1.x)

- [ ] Fourth theme preset (e.g., `monokai` or `solarized`) — only if users request it; three presets cover the functional space
- [ ] Config path respects `XDG_CONFIG_HOME` env var — only if a user is on a non-standard XDG setup (unlikely for this audience)

### Future Consideration (v2+)

- [ ] In-TUI settings panel — explicitly deferred per PROJECT.md; add to v1.4+ if there's demand
- [ ] Per-color overrides in config — only if named presets prove insufficient
- [ ] Config hot-reload — only if restart friction becomes a real complaint

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Config loader (missing/malformed/unknown handling) | HIGH | LOW | P1 |
| `emoji` config key + flag precedence | HIGH | LOW | P1 |
| `theme` config key + validation | HIGH | LOW | P1 |
| Theme struct refactor (styles.go) | HIGH | MEDIUM | P1 — required to make themes work |
| `default` theme preset (current colors) | HIGH | LOW | P1 |
| `minimal` theme preset | MEDIUM | LOW | P1 |
| `high-contrast` theme preset | MEDIUM | LOW | P1 |
| `?` help overlay with config path | MEDIUM | LOW | P1 |
| `--help` example TOML block | LOW | LOW | P2 |
| XDG_CONFIG_HOME support | LOW | LOW | P3 |
| In-TUI settings panel | MEDIUM | HIGH | P3 — v1.4+ |

**Priority key:**
- P1: Must have for v1.3 launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Behavior | lazygit | k9s | gitui | gsd-watch (v1.3 plan) |
|----------|---------|-----|-------|----------------------|
| Config format | YAML | YAML | RON | TOML (already chosen per PROJECT.md) |
| Config location | `~/Library/Application Support/lazygit/` (macOS) | `$XDG_CONFIG_HOME/k9s/` | `$XDG_CONFIG_HOME/gitui/` | `~/.config/gsd-watch/config.toml` |
| Missing config | Silent defaults | Silent defaults | Silent defaults | Silent defaults |
| Malformed config | Fatal error + path | Likely silent revert | Fatal error | Fatal error + path |
| Unknown keys | Ignored | Ignored | Ignored | Ignored (BurntSushi/toml default) |
| Theme system | Raw color fields in config | Separate skin YAML files | Separate RON file + `-t` flag | Named presets in config (simpler, appropriate for scope) |
| Number of themes | Community-sourced (many) | ~20 built-in skins | ~5 built-in | 3 presets |
| Flag vs config | Flags win | Env var wins over config | `-t` flag selects theme file | Flag wins |
| Config path in UI | Not shown in TUI | `k9s info` command | Not shown | `?` overlay |

## Theme Preset Design Notes

These notes inform implementation decisions without prescribing exact color values.

### `default`
The current color palette (ColorGreen=2, ColorAmber=3, ColorRed=1, ColorGray=8) using `lipgloss.AdaptiveColor`. This is already the behavior — the Theme struct for `default` is a direct extraction of existing styles.go vars. Zero visual change for existing users.

### `minimal`
Goal: content-first, no status coloring. Structural hierarchy visible via indentation only. Good for users who find colors distracting or are in contexts where color means something (e.g., presenting to others).
- Status icons: rendered without color (no Foreground style applied)
- Progress bar: single color (gray or terminal default)
- Phase badges: rendered without color
- Borders / separators: terminal default or bold only

### `high-contrast`
Goal: maximal legibility in degraded environments (SSH, 8-color terminals, high ambient light, accessibility needs).
- Uses only 16-color ANSI palette indices (1–15), never 256-color or truecolor
- Complete status: Bold + color index 2 (green)
- Active/in-progress: Bold + color index 3 (yellow)
- Failed: Bold + color index 1 (red) with possible reverse video
- Pending: color index 7 (white) — not 8 (dark gray), which disappears on dark SSH backgrounds
- Progress bar: Bold, no gradient
- Combines well with `--no-emoji` for full SSH compatibility (document this in --help)

## Sources

- [lazygit Config.md](https://github.com/jesseduffield/lazygit/blob/master/docs/Config.md) — YAML config structure, theme color properties, missing/unknown key handling (HIGH confidence, official)
- [k9s skins documentation](https://k9scli.io/topics/skins/) — skin file format, location, per-context overrides, fallback to stock (HIGH confidence, official)
- [gitui THEMES.md](https://github.com/gitui-org/gitui/blob/master/THEMES.md) — RON theme file, `-t` flag override, XDG location (MEDIUM confidence, official)
- [BurntSushi/toml](https://pkg.go.dev/github.com/BurntSushi/toml) — default lenient behavior (unknown keys ignored), `Undecoded()` for strict validation, `MetaData` (HIGH confidence, official Go package docs)
- [lipgloss v1](https://pkg.go.dev/github.com/charmbracelet/lipgloss) — AdaptiveColor, style composition, color profile (HIGH confidence, official)
- [Go flag package behavior](https://pkg.go.dev/flag) — no built-in "was this flag explicitly set?" mechanism; pointer-flag pattern as workaround (HIGH confidence, stdlib docs)
- Standard flag > config > default precedence — documented in Viper, gh CLI, and multiple Go configuration guides; considered idiomatic (MEDIUM confidence, multiple consistent sources)
- PROJECT.md (gsd-watch) — config path, key names, theme names, scope constraints (authoritative)

---
*Feature research for: gsd-watch v1.3 Settings milestone — config file + theme presets*
*Researched: 2026-03-26*
