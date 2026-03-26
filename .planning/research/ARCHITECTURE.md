# Architecture Research

**Domain:** Config + theme integration into existing Go Bubble Tea TUI (gsd-watch v1.3)
**Researched:** 2026-03-26
**Confidence:** HIGH (direct analysis of actual codebase — no speculation)

---

## Context: What Already Exists (v1.2 baseline)

Source of truth is the live codebase, read 2026-03-26.

```
cmd/gsd-watch/main.go
  flag parse (--no-emoji bool, --debug bool)
  app.New(events chan tea.Msg, noEmoji bool)    ← bare bool
  tea.NewProgram(model, tea.WithAltScreen()).Run()

internal/tui/app/model.go         root Bubble Tea model
  Model.noEmoji bool               sole config field on root model
  tree.SetOptions(Options{NoEmoji: noEmoji})
  helpView(width int, noEmoji bool)    standalone function, no config path

internal/tui/styles.go            package-level vars + two helper functions
  var ColorGreen  = lipgloss.AdaptiveColor{Light: "2", Dark: "2"}
  var ColorAmber  = lipgloss.AdaptiveColor{Light: "3", Dark: "3"}
  var ColorRed    = lipgloss.AdaptiveColor{Light: "1", Dark: "1"}
  var ColorGray   = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}
  var CompleteStyle     = lipgloss.NewStyle().Foreground(ColorGreen)
  var ActiveStyle       = lipgloss.NewStyle().Foreground(ColorGreen)
  var PendingStyle      = lipgloss.NewStyle().Foreground(ColorGray)
  var FailedStyle       = lipgloss.NewStyle().Foreground(ColorRed)
  var NowMarkerStyle    = lipgloss.NewStyle().Foreground(ColorAmber)
  var RefreshFlashStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorGreen)
  var QuitPendingStyle  = lipgloss.NewStyle().Bold(true).Foreground(ColorAmber)
  func StatusIcon(status string, noEmoji bool) string  ← calls package vars directly
  func BadgeString(badge string, noEmoji bool) string  ← no style calls, string only

internal/tui/tree/model.go
  type Options struct { NoEmoji bool }   only field — no Theme
  func (t TreeModel) SetOptions(o Options) TreeModel

internal/tui/tree/view.go
  imports tui; calls tui.PendingStyle.Render() in multiple places:
    - RowPhase: tui.PendingStyle.Render(part/prefixStr/continuation/badgeLine/"(no plans yet)")
    - RowPlan:  tui.PendingStyle.Render(connector/continuation/rawText)
    - RowQuickTask: same pattern as RowPlan
    - RenderArchiveRow: tui.PendingStyle.Render(row)       ← exported; takes noEmoji bool
    - RenderArchiveSeparator: tui.PendingStyle.Render(result)
    - Empty state: lipgloss.NewStyle().Foreground(tui.ColorGray)
    - highlightStyle: package-level var (lipgloss.NewStyle().Bold(true)) — no theme needed
  func tui.StatusIcon(...) called via opts.NoEmoji (already threaded correctly)
  func tui.BadgeString(...) called via opts.NoEmoji (already threaded correctly)

internal/tui/header/model.go
  uses tui.ColorGray (separator), tui.ColorGreen + tui.ColorGray (progress bar) directly

internal/tui/footer/model.go
  uses tui.ColorGray directly (grayStyle local var)
  uses tui.RefreshFlashStyle, tui.QuitPendingStyle directly

go.mod: NO TOML library present (BurntSushi/toml not in dependencies)
```

---

## System Overview: Target State (v1.3)

```
cmd/gsd-watch/main.go
  flag parse (--no-emoji, --debug)
  config.Load()                        NEW — reads ~/.config/gsd-watch/config.toml
    missing file  ->  Config defaults, nil error
    bad TOML      ->  Config defaults, non-nil error (warn + continue)
  if *noEmojiFlag { cfg.NoEmoji = true }   flag always wins
  app.New(events, cfg)                 signature change: bool -> config.Config

internal/config/config.go             NEW PACKAGE (leaf — no tui imports)
  type Config struct {
      NoEmoji bool   `toml:"emoji"`    inverted: emoji=false in file -> NoEmoji=true
      Theme   string `toml:"theme"`    "default" | "minimal" | "high-contrast"
  }
  func Load() (Config, error)
  func ConfigPath() string             returns ~/.config/gsd-watch/config.toml

internal/tui/styles.go                ADDITIVE CHANGE — existing vars and functions untouched
  type Theme struct {
      CompleteStyle  lipgloss.Style
      ActiveStyle    lipgloss.Style
      PendingStyle   lipgloss.Style
      FailedStyle    lipgloss.Style
      NowMarkerStyle lipgloss.Style
      // RefreshFlashStyle, QuitPendingStyle stay as package-level vars (footer/header unchanged)
  }
  func DefaultTheme() Theme
  func MinimalTheme() Theme
  func HighContrastTheme() Theme
  func ThemeFromName(name string) Theme   switch on name, fallback to Default

internal/tui/app/model.go             MODIFIED
  Model.cfg config.Config              replaces noEmoji bool
  New(events, cfg config.Config) calls tui.ThemeFromName(cfg.Theme), passes via SetOptions
  helpView: gains cfgPath display (no signature change needed — call config.ConfigPath() directly)

internal/tui/tree/model.go            MODIFIED
  type Options struct {
      NoEmoji bool
      Theme   tui.Theme                NEW field
  }

internal/tui/tree/view.go             MODIFIED (~15 call sites)
  opts.Theme.PendingStyle.Render(...)  replaces tui.PendingStyle.Render(...)
  RenderArchiveRow: add theme tui.Theme param or pass opts.Theme into call chain
  RenderArchiveSeparator: add theme tui.Theme param
  Empty state: opts.Theme.PendingStyle as local style
  StatusIcon / BadgeString: remain as package-level helpers, accept opts.NoEmoji as before

internal/tui/header/model.go          UNCHANGED in v1.3
internal/tui/footer/model.go          UNCHANGED in v1.3
  continue using package-level tui.ColorXxx and style vars
```

---

## Answering the Four Architecture Questions

### Q1: Where does config loading live?

**Answer: `internal/config/` package, called from `main.go` before `app.New()`.**

The config package is a leaf. It imports only stdlib and the TOML library. It must not import anything from `internal/tui/` or `internal/parser/`.

`app.New()` must not call `config.Load()` internally because:
- `app.New()` returns `Model` (no error return path). Errors during config loading have no clean recovery surface inside a constructor.
- Tests instantiate `app.New()` directly without config files on disk.
- Config loading is a startup side effect, not a model concern.

Call site in `main.go` (after existing flag parsing):
```go
cfg, err := config.Load()
if err != nil {
    fmt.Fprintf(os.Stderr, "gsd-watch: config warning: %v\n", err)
    // continue — err means malformed TOML, cfg already holds defaults
}
// --no-emoji flag always wins over config file value
if *noEmojiFlag {
    cfg.NoEmoji = true
}
p := tea.NewProgram(app.New(events, cfg), tea.WithAltScreen())
```

### Q2: How do named themes work with existing lipgloss.AdaptiveColor usage in styles.go?

**Answer: Add a `Theme` struct to `internal/tui/styles.go` alongside the existing package-level vars. Existing vars are NOT removed.**

The existing `PendingStyle`, `CompleteStyle` etc. vars stay in place for header and footer (unchanged in v1.3). The `Theme` struct holds the same logical styles as fields, letting tree/view.go switch from package-var references to struct-field references.

```go
// internal/tui/styles.go (additions only — existing vars below unchanged)

type Theme struct {
    CompleteStyle  lipgloss.Style
    ActiveStyle    lipgloss.Style
    PendingStyle   lipgloss.Style
    FailedStyle    lipgloss.Style
    NowMarkerStyle lipgloss.Style
}

func DefaultTheme() Theme {
    // mirrors the current package-level vars exactly
    return Theme{
        CompleteStyle:  lipgloss.NewStyle().Foreground(ColorGreen),
        ActiveStyle:    lipgloss.NewStyle().Foreground(ColorGreen),
        PendingStyle:   lipgloss.NewStyle().Foreground(ColorGray),
        FailedStyle:    lipgloss.NewStyle().Foreground(ColorRed),
        NowMarkerStyle: lipgloss.NewStyle().Foreground(ColorAmber),
    }
}

func MinimalTheme() Theme { ... }       // reduced color palette
func HighContrastTheme() Theme { ... }  // bold + high-contrast ANSI values

func ThemeFromName(name string) Theme {
    switch name {
    case "minimal":
        return MinimalTheme()
    case "high-contrast":
        return HighContrastTheme()
    default:
        return DefaultTheme()
    }
}
```

`StatusIcon()` in styles.go already takes `noEmoji bool` and calls the package-level style vars directly. For v1.3, leave `StatusIcon()` unchanged — tree/view.go passes `opts.Theme` for direct style calls, and calls `tui.StatusIcon(status, opts.NoEmoji)` for icon rendering. A full `StatusIcon(status string, noEmoji bool, theme Theme)` signature change is a v1.4 consideration.

### Q3: How does config.emoji interact with the existing Options.NoEmoji pattern?

**Answer: config.emoji is the persistent default; --no-emoji flag is a one-shot override that always wins. Both land on the same `cfg.NoEmoji bool` before any model is constructed.**

Precedence resolved entirely in `main.go`, two lines after `config.Load()`:

```
config.toml emoji = false   ->  cfg.NoEmoji = true   (inverted mapping)
config.toml emoji = true    ->  cfg.NoEmoji = false
config.toml absent          ->  cfg.NoEmoji = false   (default)
--no-emoji flag present     ->  cfg.NoEmoji = true    (overrides config)
both present, flag wins     ->  cfg.NoEmoji = true
```

The inversion (`emoji = false` in TOML -> `NoEmoji = true` in struct) is handled in `config.Load()`:

```go
type rawConfig struct {
    Emoji *bool  `toml:"emoji"`  // pointer to detect presence
    Theme string `toml:"theme"`
}
// After decode:
if raw.Emoji != nil && !*raw.Emoji {
    cfg.NoEmoji = true
}
```

Alternatively, keep the field name matching TOML (`Emoji bool`) and invert at the assignment in `main.go`. Either way, the inversion happens once, not scattered across render paths.

The `Options.NoEmoji` field on `tree.TreeModel` continues to work identically — it receives `cfg.NoEmoji` from `app.New()` via `SetOptions()`, unchanged interface.

### Q4: How does the help overlay get config file path without tight coupling?

**Answer: Call `config.ConfigPath()` directly inside `helpView()`. No tight coupling — `app/model.go` already imports `internal/config` (it receives `config.Config` from `main.go`).**

The concern about tight coupling does not apply here because:
- `internal/config` is a leaf package (no imports from `internal/tui/`)
- `internal/tui/app` already imports `internal/config` to accept the `config.Config` argument in `New()`
- `helpView()` is a private function in `internal/tui/app/model.go` — same file, same package, same import graph node

Adding `config.ConfigPath()` to the help text is two lines:

```go
// In helpView() — after existing phaseStages block:
helpText := `...existing text...

Config
` + config.ConfigPath()
```

No new parameter. No new import. No new dependency edge.

If `helpView` were ever to be extracted to a separate package or made testable without disk I/O, the correct move would be to accept `cfgPath string` as a parameter instead. But for v1.3, direct call is the least-ceremony approach consistent with the codebase style.

---

## Component Inventory: New vs Modified

| Component | Status | Action Required |
|-----------|--------|-----------------|
| `internal/config/config.go` | NEW | Create from scratch |
| `internal/tui/styles.go` | MODIFIED (additive) | Add Theme struct + 4 constructor funcs; existing vars and StatusIcon/BadgeString untouched |
| `internal/tui/tree/model.go` | MODIFIED | Add `Theme tui.Theme` field to Options struct |
| `internal/tui/tree/view.go` | MODIFIED | Replace ~15 `tui.PendingStyle.Render()` calls with `opts.Theme.PendingStyle.Render()`; add theme param to RenderArchiveRow/RenderArchiveSeparator |
| `internal/tui/app/model.go` | MODIFIED | `New()` signature (bool -> config.Config), store cfg, call ThemeFromName, pass via SetOptions, add cfgPath to helpView |
| `cmd/gsd-watch/main.go` | MODIFIED | Add config.Load(), apply flag override, update app.New() call |
| `internal/tui/header/model.go` | UNCHANGED | Continues using tui.ColorXxx package vars |
| `internal/tui/footer/model.go` | UNCHANGED | Continues using tui.ColorXxx package vars + RefreshFlashStyle/QuitPendingStyle |
| `internal/parser/` | UNCHANGED | No config or theme concerns |
| `internal/watcher/` | UNCHANGED | No config or theme concerns |
| `go.mod` | MODIFIED | Add `github.com/BurntSushi/toml` |

---

## Import Graph: Cycle Risk Analysis

The existing rule (PROJECT.md): `tui/*` sub-packages import `tui` for shared types. The new `internal/config` package must be a leaf.

Safe import graph after v1.3:

```
cmd/gsd-watch/main.go
    imports: internal/config          (NEW — leaf, stdlib + toml only)
    imports: internal/tui/app

internal/tui/app
    imports: internal/config          (NEW — for config.Config type + config.ConfigPath())
    imports: internal/tui             (styles, keys, messages)
    imports: internal/tui/tree
    imports: internal/tui/header
    imports: internal/tui/footer
    imports: internal/parser
    imports: internal/watcher

internal/tui/tree
    imports: internal/tui             (for tui.Theme, tui.KeyMap, tui.StatusIcon etc.)
    imports: internal/parser

internal/tui/header, internal/tui/footer
    imports: internal/tui             (unchanged)
    imports: internal/parser          (unchanged)

internal/config
    imports: NOTHING from internal/   (stdlib + BurntSushi/toml only)
```

Cycle check: `internal/config` imports nothing from the project. No cycle possible. The new edge `internal/tui/app -> internal/config` is safe: app is a leaf consumer, config is a leaf provider, and no sub-package of tui imports config.

---

## Data Flow Diagrams

### Startup Config-to-Render Path

```
main() invoked
    |
    +-- flag.Parse()
    |
    +-- config.Load()
    |       reads ~/.config/gsd-watch/config.toml
    |       ENOENT     ->  Config{NoEmoji:false, Theme:"default"}, nil
    |       decode err ->  Config defaults, non-nil error (warn + continue)
    |
    +-- if *noEmojiFlag { cfg.NoEmoji = true }   (flag always wins)
    |
    +-- app.New(events, cfg)
    |       tui.ThemeFromName(cfg.Theme) -> tui.Theme value (resolved once)
    |       tree.SetOptions(Options{NoEmoji: cfg.NoEmoji, Theme: theme})
    |
    +-- tea.NewProgram(model).Run()
    |
    +-- model.Init()
            cache.ParseFull()  ->  ParsedMsg  ->  tree.SetData()
            viewport.SetContent(tree.View(width, vpHeight))
                    opts.Theme.PendingStyle.Render(row)   <- theme applied here
```

### Flag vs Config Precedence

```
config.toml present + emoji = false   ->  cfg.NoEmoji = true
--no-emoji flag present               ->  cfg.NoEmoji = true  (overrides)
config.toml absent                    ->  cfg.NoEmoji = false (default)
neither                               ->  cfg.NoEmoji = false (default)
```

Flag always wins. Two lines in `main.go` after `config.Load()` is all that's needed.

---

## Build Order Across Phases

### Phase 1: Config infrastructure (zero visual change)

1. `go get github.com/BurntSushi/toml` — add TOML dependency to go.mod
2. Create `internal/config/config.go`:
   - `Config` struct with `NoEmoji bool` and `Theme string`
   - `ConfigPath() string` — expands `~/.config/gsd-watch/config.toml`
   - `Load() (Config, error)` — ENOENT returns defaults + nil error; decode error returns defaults + non-nil error
   - Default: `Config{NoEmoji: false, Theme: "default"}`
   - TOML inversion: `emoji = false` in file maps to `NoEmoji = true`
3. Update `internal/tui/tree/model.go`:
   - Add `Theme tui.Theme` to `Options` struct (zero value is fine — DefaultTheme() called in app.New())
4. Update `internal/tui/app/model.go`:
   - Change `New(events chan tea.Msg, noEmoji bool)` -> `New(events chan tea.Msg, cfg config.Config)`
   - Store `cfg config.Config` on `Model` (replaces `noEmoji bool`)
   - Replace all `m.noEmoji` reads with `m.cfg.NoEmoji`
   - `helpView` call: `helpView(m.width, m.cfg.NoEmoji)` (no change yet)
5. Update `cmd/gsd-watch/main.go`:
   - Import `internal/config`
   - Call `config.Load()` before `app.New()`
   - Apply flag override: `if *noEmojiFlag { cfg.NoEmoji = true }`
   - Pass `cfg` to `app.New()`
6. Unit tests for `internal/config/`:
   - Missing file → defaults + nil error
   - Valid TOML → correct struct values
   - Malformed TOML → defaults + non-nil error
   - Unknown keys → ignored (BurntSushi/toml ignores unknown keys by default)
   - `emoji = false` → `NoEmoji = true`; `emoji = true` → `NoEmoji = false`

Verifiable: binary runs identically to v1.2. No visual change.

### Phase 2: Theme system

1. Add `Theme` struct, `DefaultTheme()`, `MinimalTheme()`, `HighContrastTheme()`, `ThemeFromName()` to `internal/tui/styles.go`
2. Update `internal/tui/app/model.go`:
   - In `New()`: call `tui.ThemeFromName(cfg.Theme)`, pass via `SetOptions(Options{NoEmoji: cfg.NoEmoji, Theme: theme})`
3. Update `internal/tui/tree/view.go`:
   - Replace every `tui.PendingStyle` with `opts.Theme.PendingStyle` (approx. 15 sites)
   - Replace `tui.CompleteStyle`, `tui.ActiveStyle`, `tui.FailedStyle`, `tui.NowMarkerStyle` equivalents
   - Update `RenderArchiveRow(am, noEmoji, theme tui.Theme)` signature + body
   - Update `RenderArchiveSeparator(width, theme tui.Theme)` signature + body
   - Update `RenderArchiveZone(archives, width, noEmoji, theme tui.Theme)` signature + body
   - Update callers of those exported functions (ArchiveZone method, test files)
4. Tests:
   - Each named theme produces non-zero styles
   - Snapshot/golden test: DefaultTheme renders identically to pre-v1.3 output (regression guard)
   - `RenderArchiveRow` and `RenderArchiveSeparator` unit tests pass updated signatures

### Phase 3: Help overlay config file path

1. `config.ConfigPath()` already available from Phase 1
2. Update `helpView()` in `app/model.go`:
   - Add `Config` section displaying `config.ConfigPath()`
   - No signature change needed — app already imports internal/config
3. Snapshot test for help overlay text includes the config file path string

---

## Anti-Patterns Specific to This Integration

### Anti-Pattern 1: Mutating package-level style vars at runtime

**What people do:** Write `SetTheme(name string)` that reassigns `tui.PendingStyle = newStyle`.
**Why it's wrong:** Global mutable state breaks parallel tests and contradicts the Elm immutable-update pattern used throughout this codebase.
**Do this instead:** Pass `tui.Theme` value through `tree.Options`. Each render uses its local Options copy.

### Anti-Pattern 2: Loading config inside app.New() or model.Init()

**What people do:** `app.New()` calls `config.Load()` internally so callers don't have to think about it.
**Why it's wrong:** `app.New()` returns `Model` (no error path). Errors surface inside model construction with no clean recovery. Tests must either provide real config files on disk or mock the filesystem.
**Do this instead:** Load in `main.go`, pass the loaded `Config` to `app.New()` as a plain value.

### Anti-Pattern 3: Resolving theme name on every render frame

**What people do:** Call `tui.ThemeFromName(m.cfg.Theme)` inside `View()` or `tree.View()`.
**Why it's wrong:** Allocates a new Theme struct on every render (multiple times per second during scrolling). Theme does not change at runtime.
**Do this instead:** Resolve once in `app.New()`. Store the resolved `Theme` in model state. Pass through `SetOptions()` once.

### Anti-Pattern 4: Migrating header and footer to Theme in v1.3

**What people do:** Migrate all four TUI components simultaneously.
**Why it's wrong:** Header and footer each have 3-4 color references. The blast radius (changed files, test updates, regression risk) is disproportionate to the visual impact. These components have no `Options` struct to extend.
**Do this instead:** Migrate tree only in v1.3 (highest visual weight, already has `Options` pattern). Header and footer remain on package-level vars. Revisit in v1.4 if theme completeness matters.

### Anti-Pattern 5: Storing config file path as a hardcoded string in multiple places

**What people do:** Paste `"~/.config/gsd-watch/config.toml"` in both `config.go` and `app/model.go` (for help overlay).
**Why it's wrong:** Two sources of truth for the same path; diverge on rename.
**Do this instead:** Export `config.ConfigPath() string` from the config package. Both `Load()` and `helpView()` call the same function.

### Anti-Pattern 6: Changing RenderArchiveRow/RenderArchiveSeparator signatures without updating all callers

**What people do:** Update the function signatures in view.go but forget the exported test callers in tree_test package.
**Why it's wrong:** These functions are exported (PROJECT.md key decision: "FormatArchiveDate, RenderArchiveRow, RenderArchiveSeparator, RenderArchiveZone exported" for testability). External test files call them directly. Signature change without updating tests causes compilation failure.
**Do this instead:** Search all call sites of exported archive functions before changing signatures. Update tree_test usages in the same commit as the signature change.

---

## TOML Library Decision

`go.mod` has no TOML library. Two viable options:

| Library | Stars | API | Transitive deps | Recommendation |
|---------|-------|-----|-----------------|----------------|
| `github.com/BurntSushi/toml` v1.x | ~4.5K | Minimal: `toml.DecodeFile()` | Zero | **Use this** |
| `github.com/pelletier/go-toml/v2` | ~1.7K | Larger, richer | A few | Overkill for a 2-key config |

BurntSushi/toml is the canonical Go TOML library. Unknown keys are silently ignored by default — exactly the behavior needed for a config file that may evolve (future keys in newer binaries don't break old binaries).

Install: `go get github.com/BurntSushi/toml@v1`

---

## Style Reference Audit (view.go — all sites to migrate)

Confirmed by reading `internal/tui/tree/view.go` directly:

| Line context | Current call | Migration target |
|---|---|---|
| Empty state paragraph | `lipgloss.NewStyle().Foreground(tui.ColorGray)` | `opts.Theme.PendingStyle` |
| RowPhase: dimmed name part | `tui.PendingStyle.Render(part)` | `opts.Theme.PendingStyle.Render(part)` |
| RowPhase: dimmed prefix | `tui.PendingStyle.Render(prefixStr)` | `opts.Theme.PendingStyle.Render(prefixStr)` |
| RowPhase: dimmed continuation | `tui.PendingStyle.Render(continuation)` | `opts.Theme.PendingStyle.Render(continuation)` |
| RowPhase: dimmed badge line | `tui.PendingStyle.Render(badgeLine)` | `opts.Theme.PendingStyle.Render(badgeLine)` |
| RowPhase: no plans yet | `tui.PendingStyle.Render("(no plans yet)")` | `opts.Theme.PendingStyle.Render("(no plans yet)")` |
| RowPlan: dimmed connector | `tui.PendingStyle.Render(connector)` | `opts.Theme.PendingStyle.Render(connector)` |
| RowPlan: dimmed continuation | `tui.PendingStyle.Render(continuation)` | `opts.Theme.PendingStyle.Render(continuation)` |
| RowPlan: dimmed text | `tui.PendingStyle.Render(rawText)` | `opts.Theme.PendingStyle.Render(rawText)` |
| RowQuickTask: dimmed connector | `tui.PendingStyle.Render(connector)` | `opts.Theme.PendingStyle.Render(connector)` |
| RowQuickTask: dimmed continuation | `tui.PendingStyle.Render(continuation)` | `opts.Theme.PendingStyle.Render(continuation)` |
| RowQuickTask: dimmed text | `tui.PendingStyle.Render(rawText)` | `opts.Theme.PendingStyle.Render(rawText)` |
| RowQuickTask: no quick tasks | `tui.PendingStyle.Render("(no quick tasks)")` | `opts.Theme.PendingStyle.Render("(no quick tasks)")` |
| RenderArchiveRow | `tui.PendingStyle.Render(row)` | add `theme tui.Theme` param |
| RenderArchiveSeparator | `tui.PendingStyle.Render(result)` | add `theme tui.Theme` param |

`NowMarkerStyle` is used once for the `← now` marker — add to Theme struct.
`highlightStyle` (bold only, no color) is a view.go package-level var — leave unchanged, not theme-sensitive.

---

## Sources

- Direct code analysis (read 2026-03-26):
  - `cmd/gsd-watch/main.go`
  - `internal/tui/app/model.go`
  - `internal/tui/styles.go`
  - `internal/tui/tree/model.go`
  - `internal/tui/tree/view.go`
  - `internal/tui/header/model.go`
  - `internal/tui/footer/model.go`
  - `go.mod`
- `.planning/PROJECT.md` — key decisions, v1.3 milestone goals, constraints

---
*Architecture research for: gsd-watch v1.3 config + theme integration*
*Researched: 2026-03-26*
