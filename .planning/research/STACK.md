# Stack Research

**Domain:** Go terminal TUI with filesystem watching and Claude Code plugin integration
**Researched:** 2026-03-26
**Confidence:** HIGH (all versions verified against pkg.go.dev and official docs)

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.1 (module minimum 1.22+) | Language runtime | Loop variable fix (each `for` iteration gets its own variable — eliminates a classic goroutine closure bug in TUI event handlers). go.mod currently declares `go 1.26.1`. |
| Bubble Tea | v1.3.10 | TUI event loop, Model/Update/View lifecycle | The standard Go TUI framework. v1.x is stable and widely adopted. v2.x exists at `github.com/charmbracelet/bubbletea/v2` but is a different import path and requires the Cursed Renderer — stick with v1.x per PROJECT.md constraint. |
| Lip Gloss | v1.1.0 | Terminal styling and layout | Pairs directly with Bubble Tea v1.x. Provides `JoinVertical`, `JoinHorizontal`, `Place`, border styles, and styled string composition. v1.x avoids the I/O control changes introduced in v2. |
| Bubbles | v1.0.0 | Reusable TUI components (viewport, key, spinner) | The charmbracelet component library for Bubble Tea. v1.0.0 released 2026-02-09 is the stable release for Bubble Tea v1.x projects. Provides `viewport` (scrollable content), `key` (rebindable keybindings with auto-generated help), and `progress` (animated progress bar). |
| fsnotify | v1.9.0 | Filesystem event watching | The canonical Go filesystem watcher. v1.9.0 is the latest (Apr 2025). Does NOT support recursive watching on any platform including macOS/kqueue — requires explicit per-directory `Add()` calls. |
| gopkg.in/yaml.v3 | v3.0.1 | YAML frontmatter parsing for `*-PLAN.md` files | Standard Go YAML library. Used to decode the `---` frontmatter block in GSD PLAN.md files. Supports struct tags and inline embedding. |
| BurntSushi/toml | v1.6.0 | Config file parsing (`~/.config/gsd-watch/config.toml`) | See "v1.3 additions" section below for rationale. NOT yet in go.mod — must be added via `go get`. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/charmbracelet/bubbles/viewport` | (part of bubbles v1.0.0) | Scrollable content pane for the phase/plan tree | Use for the main tree view — handles scroll position, page-up/down, and renders content larger than terminal height |
| `github.com/charmbracelet/bubbles/key` | (part of bubbles v1.0.0) | Keybinding definitions with help text | Use to define all keyboard bindings (↑↓/jk/←→/hl/q) so they appear consistently in the footer help bar |
| `github.com/charmbracelet/bubbles/progress` | (part of bubbles v1.0.0) | Animated progress bar for the header | Use for the phase completion percentage bar in the header area |
| `github.com/charmbracelet/bubbles/spinner` | (part of bubbles v1.0.0) | Loading indicator during startup file scan | Use only during the initial `.planning/` directory walk — not during normal operation |
| Go stdlib `net` | stdlib | Unix domain socket IPC listener | Use `net.Listen("unix", socketPath)` for the goroutine that accepts `refresh` signals from `gsd-watch-signal.sh`. No third-party library needed. |
| Go stdlib `os/signal` | stdlib | Graceful shutdown on SIGINT/SIGTERM | Wrap program exit to clean up the socket file before terminating |
| Go stdlib `encoding/json` | stdlib | Parsing `config.json` from `.planning/` | GSD config is JSON — no external library required |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `make` | Build, install, plugin-install targets | Makefile with `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64/amd64` for static binary |
| `go build` | Produces static binary | Use `-ldflags="-s -w"` to strip debug info and keep binary under 15MB |
| `goreleaser` (optional) | Cross-compile darwin/arm64 + darwin/amd64 into universal binary | Only needed if distributing; manual `lipo` merge works for personal use |

---

## v1.3 Additions: TOML Config + Theme System

### What's NEW vs existing go.mod

The current `go.mod` (verified 2026-03-26) does NOT include `github.com/BurntSushi/toml`. All other dependencies listed in go.mod are already present. The only new dependency for v1.3 is:

```bash
go get github.com/BurntSushi/toml@v1.6.0
```

Everything else (XDG path, theme system, help overlay path) uses stdlib or the existing lipgloss dependency.

---

### TOML Parsing: BurntSushi/toml v1.6.0

**Recommendation: Add `github.com/BurntSushi/toml@v1.6.0` as a new dependency.**

**Why BurntSushi/toml over pelletier/go-toml v2:**

The config file has exactly 2 keys (`emoji` and `theme`). pelletier/go-toml v2's performance advantage (2–5x faster) is irrelevant when the file is parsed once at startup. BurntSushi/toml is:
- Simpler API — `toml.DecodeFile(path, &cfg)` in one call
- Widely adopted (37,875 importers vs ~1,742 for go-toml v2 at time of research)
- Actively maintained — v1.6.0 released December 18, 2025 (verified on pkg.go.dev)
- Familiar `encoding/json`-compatible struct tags

**Why NOT yaml.v3 for TOML:** yaml.v3 cannot parse TOML — these are different formats. yaml.v3 is only for YAML.

```go
type Config struct {
    Emoji string `toml:"emoji"` // "on" | "off" | "" (empty = defer to flag)
    Theme string `toml:"theme"` // "default" | "minimal" | "high-contrast"
}

func LoadConfig(path string) (Config, error) {
    var cfg Config
    if _, err := toml.DecodeFile(path, &cfg); err != nil {
        return cfg, err
    }
    return cfg, nil
}
```

**Missing file is NOT an error** — `os.IsNotExist(err)` check in the caller; return zero-value Config and nil error. The PROJECT.md requirement is "missing file uses defaults silently."

**Why NOT pelletier/go-toml v2:** Same API complexity for no benefit at this scale. Its v2.3.0 release (March 24, 2026) is very recent — less battle-tested at this patch level. BurntSushi/toml has 10x the adoption.

---

### XDG Config Directory: stdlib only, manual XDG pattern

**Do NOT use `os.UserConfigDir()` for this project.**

`os.UserConfigDir()` on macOS/Darwin returns `$HOME/Library/Application Support` — this is documented behavior, not a bug. The project spec requires `~/.config/gsd-watch/config.toml`. On macOS, `~/.config` follows the XDG Base Directory Specification, which is the convention for CLI/TUI tools (as opposed to GUI apps that use `Library/Preferences`).

**No new dependency needed.** `os.UserHomeDir()` (stdlib) is sufficient:

```go
// configDir returns ~/.config (XDG_CONFIG_HOME if set, else $HOME/.config).
// Intentionally bypasses os.UserConfigDir() which returns
// $HOME/Library/Application Support on macOS — wrong for CLI tools.
func configDir() (string, error) {
    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" && filepath.IsAbs(xdg) {
        return xdg, nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".config"), nil
}

// ConfigFilePath returns ~/.config/gsd-watch/config.toml
func ConfigFilePath() (string, error) {
    base, err := configDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(base, "gsd-watch", "config.toml"), nil
}
```

The `XDG_CONFIG_HOME` check mirrors Go stdlib's own `os.UserConfigDir` source for Linux and respects the spec. `filepath.IsAbs()` prevents accepting relative paths in the env var.

**Why not a third-party XDG library (adrg/xdg, kyoh86/xdg):** The project constraint is "no new external dependencies if possible." The 4-line pattern above handles the only case needed: macOS with optional `XDG_CONFIG_HOME` override.

---

### Theme System: lipgloss.AdaptiveColor, no new dependencies

**No new dependency needed.** The existing `lipgloss.AdaptiveColor` pattern already in `internal/tui/styles.go` is the correct foundation.

**Current state of styles.go (verified 2026-03-26):**

`internal/tui/styles.go` currently defines four package-level `var` color values:
- `ColorGreen`, `ColorAmber`, `ColorRed`, `ColorGray` (all `lipgloss.AdaptiveColor`)

Plus derived styles: `CompleteStyle`, `ActiveStyle`, `PendingStyle`, `FailedStyle`, `NowMarkerStyle`, `RefreshFlashStyle`, `QuitPendingStyle`.

**Current state of tree.Options (verified 2026-03-26):**

`internal/tui/tree/model.go` defines:
```go
type Options struct {
    NoEmoji bool
}
```

`SetOptions(o Options)` is available to apply options to a `TreeModel` copy.

**Design: Palette struct + named presets, injected via Options**

Add a `Palette` type to `internal/tui/` (same package as `styles.go`) and extend `Options` to carry it:

```go
// Palette defines the adaptive color values for a named theme.
type Palette struct {
    Green  lipgloss.AdaptiveColor
    Amber  lipgloss.AdaptiveColor
    Red    lipgloss.AdaptiveColor
    Gray   lipgloss.AdaptiveColor
}

var (
    ThemeDefault = Palette{
        Green: lipgloss.AdaptiveColor{Light: "2", Dark: "2"},
        Amber: lipgloss.AdaptiveColor{Light: "3", Dark: "3"},
        Red:   lipgloss.AdaptiveColor{Light: "1", Dark: "1"},
        Gray:  lipgloss.AdaptiveColor{Light: "8", Dark: "8"},
    }

    ThemeMinimal = Palette{
        Green: lipgloss.AdaptiveColor{Light: "8", Dark: "7"},
        Amber: lipgloss.AdaptiveColor{Light: "8", Dark: "7"},
        Red:   lipgloss.AdaptiveColor{Light: "8", Dark: "7"},
        Gray:  lipgloss.AdaptiveColor{Light: "8", Dark: "240"},
    }

    ThemeHighContrast = Palette{
        Green: lipgloss.AdaptiveColor{Light: "10", Dark: "10"}, // bright green
        Amber: lipgloss.AdaptiveColor{Light: "11", Dark: "11"}, // bright yellow
        Red:   lipgloss.AdaptiveColor{Light: "9", Dark: "9"},   // bright red
        Gray:  lipgloss.AdaptiveColor{Light: "15", Dark: "15"}, // bright white
    }
)

func ThemeByName(name string) Palette {
    switch name {
    case "minimal":
        return ThemeMinimal
    case "high-contrast":
        return ThemeHighContrast
    default:
        return ThemeDefault
    }
}
```

**Integration path — extend existing Options:**

```go
// In tree/model.go:
type Options struct {
    NoEmoji bool
    Theme   tui.Palette  // zero value falls back to ThemeDefault
}
```

Sub-models receive `Palette` through `Options`, matching the existing `NoEmoji` threading pattern. Components call `opts.Theme.Green` etc. instead of the package-level color vars.

**Flag / config precedence (implement in main.go):**
1. `--no-emoji` flag wins over `config.emoji = "off"` (flag is explicit user action)
2. If no `--no-emoji` flag AND `config.emoji = "off"`, activate no-emoji mode
3. `config.theme` sets the palette; no flag override is needed for v1.3 scope
4. Unknown theme names fall back to `default` silently — never error

---

### Help Overlay: no new dependencies

The `?` key help overlay showing the config file path requires:
- `ConfigFilePath()` (stdlib, see above) — call at startup to resolve the path
- Pass the resolved path string into the relevant model via `Options` or a dedicated field
- Render using existing lipgloss styles — no new library needed

---

## Bubble Tea v1 API Reference

This section documents the specific v1.x API patterns critical to this project.

### Model Interface (3 required methods)

```go
type model struct { /* your state */ }

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m model) View() string { return "" }
```

`Init()` returns an initial `tea.Cmd` (or `nil`). `Update()` dispatches on `msg` type with a type switch. `View()` returns a plain string — Bubble Tea handles redrawing.

### Program Setup

```go
p := tea.NewProgram(
    initialModel(),
    tea.WithAltScreen(),       // Full-window mode — recommended for sidebar
)
if _, err := p.Run(); err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}
```

**Key program options for this project:**
- `tea.WithAltScreen()` — takes over the full terminal pane (appropriate for tmux split)
- `tea.WithContext(ctx)` — allows cancellation from a signal handler goroutine
- Do NOT use `tea.WithMouseCellMotion()` — PROJECT.md specifies keyboard only

### External Message Injection (socket refresh pattern)

`p.Send(msg)` is thread-safe and can be called from any goroutine. This is the mechanism for the Unix socket listener to trigger a TUI re-render:

```go
// In your socket listener goroutine:
p.Send(RefreshMsg{})  // non-blocking once program is running

// In your Update():
case RefreshMsg:
    m = reloadFromDisk(m)
    return m, nil
```

### WindowSizeMsg — handle terminal resize

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.viewport.Width = msg.Width
    m.viewport.Height = msg.Height - headerHeight - footerHeight
```

### v1 vs v0 critical differences

| Concern | v0 | v1 |
|---------|----|----|
| Key press type | `tea.KeyMsg` | `tea.KeyMsg` (same in v1; `KeyPressMsg` is v2 only) |
| Program run | `p.Start()` | `p.Run()` — `Start()` deprecated |
| View return | `string` | `string` (still a string in v1; `tea.View` struct is v2 only) |
| Import path | `github.com/charmbracelet/bubbletea` | Same — unchanged in v1 |

**Note:** v2 (import `github.com/charmbracelet/bubbletea/v2`) changes `View()` to return `tea.View` and splits `KeyMsg` into `KeyPressMsg`/`KeyReleaseMsg`. Do not mix v1 and v2 patterns.

---

## Lip Gloss v1 Layout Patterns

### Composing the sidebar layout

```go
// Header | Tree | Footer composition:
view := lipgloss.JoinVertical(
    lipgloss.Left,
    renderHeader(m),     // fixed height, full width
    m.viewport.View(),   // fills remaining space
    renderFooter(m),     // fixed height, full width
)
```

### Styling constants

```go
var (
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("62")).
        Padding(0, 1)

    phaseActiveStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("10"))  // bright green for ▶ active

    dimStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))  // gray for ○ upcoming
)
```

---

## fsnotify v1.9 on macOS: Recursive Watching Pattern

**Critical:** fsnotify does NOT support recursive watching on macOS/kqueue (or any platform). Each directory must be added explicitly.

### Required pattern for `.planning/` watching

```go
// On startup: walk all dirs and add each one
err := filepath.WalkDir(planningDir, func(path string, d fs.DirEntry, err error) error {
    if err != nil { return nil }  // skip unreadable dirs
    if d.IsDir() {
        return watcher.Add(path)
    }
    return nil
})

// On Create event: if a new directory appears, add it too
case event.Has(fsnotify.Create):
    info, err := os.Stat(event.Name)
    if err == nil && info.IsDir() {
        watcher.Add(event.Name)
    }
```

### Debouncing (required — 300ms per PROJECT.md)

```go
var debounce *time.Timer
for event := range watcher.Events {
    if debounce != nil {
        debounce.Stop()
    }
    debounce = time.AfterFunc(300*time.Millisecond, func() {
        p.Send(RefreshMsg{})
    })
}
```

---

## Installation

```bash
# v1.3: Only new dependency — TOML config file support
go get github.com/BurntSushi/toml@v1.6.0

# Already present in go.mod (no action needed):
# github.com/charmbracelet/bubbletea@v1.3.10
# github.com/charmbracelet/lipgloss@v1.1.0
# github.com/charmbracelet/bubbles@v1.0.0
# github.com/fsnotify/fsnotify@v1.9.0
# gopkg.in/yaml.v3@v3.0.1

# Static binary build (in Makefile)
# CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/gsd-watch-arm64 .
# CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/gsd-watch-amd64 .
# lipo -create dist/gsd-watch-arm64 dist/gsd-watch-amd64 -output dist/gsd-watch
```

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| BurntSushi/toml v1.6.0 | pelletier/go-toml v2.3.0 | If parsing very large TOML files at high frequency — pelletier/go-toml v2 is 2–5x faster. For a 2-key config file read once at startup, this difference is immeasurable. |
| BurntSushi/toml v1.6.0 | stdlib only (hand-rolled TOML parser) | Never — TOML has enough edge cases (multiline strings, inline tables, datetime types) that hand-rolling is error-prone. |
| Manual XDG pattern (stdlib only) | adrg/xdg or kyoh86/xdg | If cross-platform XDG support becomes needed (Linux, Windows). Current project is macOS-only — a third-party XDG library is unnecessary surface area. |
| `os.UserHomeDir()` + manual XDG | `os.UserConfigDir()` | Use `os.UserConfigDir()` only if targeting `$HOME/Library/Application Support` (macOS GUI app convention). For a CLI/TUI tool targeting `~/.config`, bypass it entirely. |
| Bubble Tea v1.3.10 | Bubble Tea v2 (`/v2`) | If you need progressive keyboard enhancement (shift+enter, ctrl+m detection) or are building a Wish-based SSH TUI server. Not needed here. |
| Bubble Tea v1.3.10 | tview / tcell | If you need a widget-based immediate-mode UI (like a form builder). Bubble Tea's Elm-architecture is better for reactive read-only displays. |
| Lip Gloss v1.1.0 | termenv directly | Only if you need raw terminal capability detection without layout primitives. Lip Gloss wraps termenv — no reason to use it directly. |
| Bubbles v1.0.0 | Custom viewport | Only if viewport's scroll model doesn't fit (e.g., need horizontal scroll). Bubbles viewport handles all the edge cases around terminal resize. |
| fsnotify v1.9.0 | `inotify` / `kqueue` directly | Never — cross-platform abstraction is valuable even for macOS-only apps. fsnotify's kqueue backend is production-grade. |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `os.UserConfigDir()` for `~/.config` on macOS | Returns `$HOME/Library/Application Support` on Darwin — documented stdlib behavior, not a bug. Wrong path for CLI tools following XDG convention. | `os.UserHomeDir()` + `filepath.Join(home, ".config")` with `XDG_CONFIG_HOME` override |
| `pelletier/go-toml/v2` | Heavier dependency for no benefit at this scale; very recent v2.3.0 release (March 2026) less battle-tested | `github.com/BurntSushi/toml@v1.6.0` |
| `gopkg.in/yaml.v3` for TOML | yaml.v3 cannot parse TOML — these are different formats. yaml.v3 stays for YAML frontmatter only. | `github.com/BurntSushi/toml@v1.6.0` |
| `github.com/charmbracelet/bubbletea/v2` | Different import path, `View()` returns `tea.View` (not string), `KeyMsg` split into `KeyPressMsg`/`KeyReleaseMsg` — incompatible with v1 Bubbles and Lip Gloss v1 | `github.com/charmbracelet/bubbletea@v1.3.10` |
| `github.com/charmbracelet/lipgloss/v2` | Requires explicit `HasDarkBackground()` calls; `Color` returns `color.Color` not `TerminalColor` — needs v2-specific wiring | `github.com/charmbracelet/lipgloss@v1.1.0` |
| `fsnotify.AddWith(..., WithBufferSize(...))` | `WithBufferSize` is Windows-only — no-op on macOS/kqueue | `watcher.Add(path)` (plain Add) |
| CGO or C bindings | Violates static binary requirement; breaks cross-compilation | Pure Go packages only |

---

## Stack Patterns by Variant

**If distributing as standalone (short slash command `/gsd-watch`):**
- Place `gsd-watch.md` in `.claude/commands/gsd-watch.md` directly
- Hooks go in `.claude/settings.local.json` under `"hooks"` key
- This is the approach PROJECT.md implies (manual install, no marketplace)

**If building for macOS universal binary:**
- Compile arm64 and amd64 separately then `lipo -create` to merge
- Both targets work with CGO_ENABLED=0 (pure Go)

**If theme selection via flag is added (`--theme`) in a future phase:**
- Parse `--theme` before config file load
- Flag value wins over config `theme` key
- Unknown theme names fall back to `default` silently — never error

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| bubbletea@v1.3.10 | lipgloss@v1.1.0 | Both use the same string-based View() contract |
| bubbletea@v1.3.10 | bubbles@v1.0.0 | Bubbles v1 targets Bubble Tea v1.x explicitly |
| bubbletea@v1.3.10 | fsnotify@v1.9.0 | No shared interface — used in separate goroutines |
| lipgloss@v1.1.0 | bubbles@v1.0.0 | Bubbles components return strings compatible with Lip Gloss styling |
| BurntSushi/toml@v1.6.0 | Go 1.22+ | Pure Go, no CGO — compatible with static binary build |
| bubbletea@v1.x | bubbletea@v2 | NOT compatible — different import paths, different View() return type |

---

## Sources

- `pkg.go.dev/github.com/BurntSushi/toml?tab=versions` — verified v1.6.0 as latest stable (Dec 18, 2025) — HIGH confidence
- `pkg.go.dev/github.com/pelletier/go-toml/v2?tab=versions` — verified v2.3.0 as latest (Mar 24, 2026); considered and rejected — HIGH confidence
- `pkg.go.dev/github.com/charmbracelet/bubbletea?tab=versions` — verified v1.3.10 as latest v1.x — HIGH confidence
- `pkg.go.dev/github.com/charmbracelet/lipgloss?tab=versions` — verified v1.1.0 as latest v1.x (Mar 2025) — HIGH confidence
- `pkg.go.dev/github.com/charmbracelet/bubbles?tab=versions` — verified v1.0.0 as latest stable (Feb 2026) — HIGH confidence
- `pkg.go.dev/github.com/fsnotify/fsnotify?tab=versions` — verified v1.9.0 (Apr 2025), confirmed no recursive support — HIGH confidence
- `go.mod` (project file, read directly) — verified actual Go module version is 1.26.1; confirmed BurntSushi/toml NOT yet present — HIGH confidence
- `internal/tui/styles.go` (read directly) — confirmed 4 package-level AdaptiveColor vars; theme integration surface identified — HIGH confidence
- `internal/tui/tree/model.go` (read directly) — confirmed `Options{NoEmoji bool}` is the extension point for `Theme Palette` — HIGH confidence
- `go doc os.UserConfigDir` (documented behavior) — confirmed Darwin returns `$HOME/Library/Application Support`, NOT `~/.config` — HIGH confidence
- `golang.org/x/proposal #29960` — os.UserConfigDir addition rationale, confirmed Darwin behavior — HIGH confidence

---

*Stack research for: Go terminal TUI sidebar — v1.3 config file and theme system additions*
*Researched: 2026-03-26*
