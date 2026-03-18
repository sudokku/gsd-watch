# Stack Research

**Domain:** Go terminal TUI with filesystem watching and Claude Code plugin integration
**Researched:** 2026-03-18
**Confidence:** HIGH (all versions verified against pkg.go.dev and official docs)

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.22+ | Language runtime | Loop variable fix (each `for` iteration gets its own variable — eliminates a classic goroutine closure bug in TUI event handlers). Required minimum per PROJECT.md. |
| Bubble Tea | v1.3.10 | TUI event loop, Model/Update/View lifecycle | The standard Go TUI framework. v1.x is stable and widely adopted. v2.x exists at `github.com/charmbracelet/bubbletea/v2` but is a different import path and requires the Cursed Renderer — stick with v1.x per PROJECT.md constraint. |
| Lip Gloss | v1.1.0 | Terminal styling and layout | Pairs directly with Bubble Tea v1.x. Provides `JoinVertical`, `JoinHorizontal`, `Place`, border styles, and styled string composition. v1.x avoids the I/O control changes introduced in v2. |
| Bubbles | v1.0.0 | Reusable TUI components (viewport, key, spinner) | The charmbracelet component library for Bubble Tea. v1.0.0 released 2026-02-09 is the stable release for Bubble Tea v1.x projects. Provides `viewport` (scrollable content), `key` (rebindable keybindings with auto-generated help), and `progress` (animated progress bar). |
| fsnotify | v1.9.0 | Filesystem event watching | The canonical Go filesystem watcher. v1.9.0 is the latest (Apr 2025). Does NOT support recursive watching on any platform including macOS/kqueue — requires explicit per-directory `Add()` calls. |
| gopkg.in/yaml.v3 | v3.0.1 | YAML frontmatter parsing for `*-PLAN.md` files | Standard Go YAML library. Used to decode the `---` frontmatter block in GSD PLAN.md files. Supports struct tags and inline embedding. |

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

The `Send()` call blocks until the program starts, then becomes asynchronous. Calling it after program exit is a no-op — safe to call without lifecycle checks.

### WindowSizeMsg — handle terminal resize

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.viewport.Width = msg.Width
    m.viewport.Height = msg.Height - headerHeight - footerHeight
```

Sent automatically on startup and on every terminal resize.

### Cmd patterns for async work

```go
// Wrap blocking work as a Cmd
func loadFilesCmd(path string) tea.Cmd {
    return func() tea.Msg {
        result, err := walkPlanning(path)
        if err != nil { return ErrorMsg{err} }
        return FilesLoadedMsg{result}
    }
}

// Return from Init() or Update()
return m, loadFilesCmd(m.planningDir)
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

### Progress bar (header)

Use `bubbles/progress` for animated transitions. For a static bar, Lip Gloss Width is sufficient:

```go
// Static progress bar via Lip Gloss:
filled := int(float64(m.width-2) * m.progressPct)
bar := lipgloss.NewStyle().
    Foreground(lipgloss.Color("62")).
    Render(strings.Repeat("█", filled)) +
    lipgloss.NewStyle().
    Foreground(lipgloss.Color("240")).
    Render(strings.Repeat("░", m.width-2-filled))
```

---

## fsnotify v1.9 on macOS: Recursive Watching Pattern

**Critical:** fsnotify does NOT support recursive watching on macOS/kqueue (or any platform). Each directory must be added explicitly. This is a known limitation tracked in issue [#18](https://github.com/fsnotify/fsnotify/issues/18).

### Required pattern for `.planning/` watching

```go
// On startup: walk all dirs and add each one
err := filepath.WalkDir(planningDir, func(path string, d fs.DirEntry, err error) error {
    if err != nil { return nil }  // skip unreadable dirs
    if d.IsDir() {
        return watcher.Add(path)  // add every directory individually
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

GSD writes multiple files during a phase execution. Without debouncing, a single `execute-phase` run triggers 10-50 events within milliseconds.

```go
// Canonical debounce pattern with time.AfterFunc:
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

### Events channel and error handling

```go
go func() {
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok { return }
            handleEvent(event)
        case err, ok := <-watcher.Errors:
            if !ok { return }
            // log but don't crash — PROJECT.md: never crash on errors
            log.Printf("watcher error: %v", err)
        }
    }
}()
```

---

## Unix Socket IPC Pattern

### Socket path convention (from PROJECT.md)

```go
// Hash project dir path for unique socket per project
import "crypto/sha256"
import "fmt"

func socketPath(projectDir string) string {
    h := sha256.Sum256([]byte(projectDir))
    return fmt.Sprintf("/tmp/gsd-watch-%x.sock", h[:4])
}
```

### Stale socket handling (required — SIGKILL won't clean up)

```go
func listenOnSocket(path string) (net.Listener, error) {
    // Try to connect; if connection refused, socket is stale
    if conn, err := net.Dial("unix", path); err == nil {
        conn.Close()
        return nil, fmt.Errorf("gsd-watch already running at %s", path)
    }
    // Remove stale socket file
    os.Remove(path)
    return net.Listen("unix", path)
}
```

### Cleanup on exit

```go
// Deferred cleanup — also register os.Signal handler for SIGTERM/SIGINT
defer os.Remove(socketPath)
```

---

## YAML Frontmatter Parsing (PLAN.md files)

GSD `*-PLAN.md` files have YAML frontmatter between `---` delimiters:

```go
type PlanFrontmatter struct {
    Title    string `yaml:"title"`
    Status   string `yaml:"status"`    // "pending" | "in_progress" | "complete"
    Phase    string `yaml:"phase"`
    Priority int    `yaml:"priority"`
}

func parseFrontmatter(content []byte) (PlanFrontmatter, error) {
    var fm PlanFrontmatter
    // Split on "---\n" — first part is empty, second is YAML, third is body
    parts := bytes.SplitN(content, []byte("---\n"), 3)
    if len(parts) < 3 {
        return fm, nil  // no frontmatter — not an error per PROJECT.md
    }
    err := yaml.Unmarshal(parts[1], &fm)
    return fm, err
}
```

**Important:** Treat parse failures as non-fatal — PROJECT.md requires graceful handling of missing/malformed files.

---

## Claude Code Plugin Integration

### Plugin structure

This project ships as a standalone configuration (not a marketplace plugin), installed at project scope:

```
gsd-watch/
├── .claude-plugin/
│   └── plugin.json          # plugin manifest
├── commands/
│   └── gsd-watch.md         # /gsd-watch slash command
├── hooks/
│   └── hooks.json           # Stop and SubagentStop hooks
└── scripts/
    └── gsd-watch-signal.sh  # signal script invoked by hooks
```

### plugin.json manifest

```json
{
  "name": "gsd-watch",
  "version": "1.0.0",
  "description": "GSD project status sidebar for Claude Code",
  "author": {
    "name": "radu"
  }
}
```

Minimal manifest — `name` is the only required field. Skills are invoked as `/gsd-watch:gsd-watch` but PROJECT.md specifies `/gsd-watch` which requires standalone (non-plugin) installation in `.claude/commands/gsd-watch.md`. See "Stack Patterns by Variant" below.

### commands/gsd-watch.md (slash command)

```markdown
---
description: Open GSD watch sidebar in a tmux split pane
---

Check if $ARGUMENTS contains "stop". If so, run: gsd-watch --stop
Otherwise, if not in a tmux session, tell the user to run gsd-watch manually
in a tmux split pane with: gsd-watch $CLAUDE_PROJECT_DIR
If in tmux: run `tmux split-window -h -l 40 gsd-watch "$CLAUDE_PROJECT_DIR"`
```

### hooks/hooks.json — Stop and SubagentStop

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/gsd-watch-signal.sh",
            "async": true
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/gsd-watch-signal.sh",
            "async": true
          }
        ]
      }
    ]
  }
}
```

**Use `"async": true`** — the signal script must NOT block Claude Code from completing. The script writes to the Unix socket and exits immediately.

### scripts/gsd-watch-signal.sh

```bash
#!/bin/bash
# Signals gsd-watch to refresh its view.
# Called by Claude Code Stop and SubagentStop hooks.
# Reads CWD from stdin JSON.
INPUT=$(cat)
CWD=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('cwd',''))" 2>/dev/null || echo "$PWD")

# Derive socket path (must match Go implementation)
HASH=$(echo -n "$CWD" | sha256sum | cut -c1-8)
SOCK="/tmp/gsd-watch-${HASH}.sock"

# Send refresh signal — ignore errors if gsd-watch isn't running
echo "refresh" | nc -U "$SOCK" -w 1 2>/dev/null || true
```

### Hook input payload (Stop event)

```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/Users/radu/Developer/my-project",
  "hook_event_name": "Stop",
  "stop_hook_active": false,
  "last_assistant_message": "..."
}
```

The script receives this on stdin. Only `cwd` is needed to derive the socket path.

---

## Installation

```bash
# go.mod setup
go mod init github.com/radu/gsd-watch

# Core TUI stack
go get github.com/charmbracelet/bubbletea@v1.3.10
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/bubbles@v1.0.0

# File watching
go get github.com/fsnotify/fsnotify@v1.9.0

# YAML frontmatter
go get gopkg.in/yaml.v3@v3.0.1

# Static binary build (in Makefile)
# CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/gsd-watch-arm64 .
# CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/gsd-watch-amd64 .
# lipo -create dist/gsd-watch-arm64 dist/gsd-watch-amd64 -output dist/gsd-watch
```

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Bubble Tea v1.3.10 | Bubble Tea v2 (`/v2`) | If you need progressive keyboard enhancement (shift+enter, ctrl+m detection) or are building a Wish-based SSH TUI server. Not needed here. |
| Bubble Tea v1.3.10 | tview / tcell | If you need a widget-based immediate-mode UI (like a form builder). Bubble Tea's Elm-architecture is better for reactive read-only displays. |
| Lip Gloss v1.1.0 | termenv directly | Only if you need raw terminal capability detection without layout primitives. Lip Gloss wraps termenv — no reason to use it directly. |
| Bubbles v1.0.0 | Custom viewport | Only if viewport's scroll model doesn't fit (e.g., need horizontal scroll). Bubbles viewport handles all the edge cases around terminal resize. |
| fsnotify v1.9.0 | `inotify` / `kqueue` directly | Never — cross-platform abstraction is valuable even for macOS-only apps. fsnotify's kqueue backend is production-grade. |
| gopkg.in/yaml.v3 | `github.com/ghodss/yaml` | Only if consuming JSON-compatible YAML (converts to JSON first). GSD frontmatter is pure YAML — v3 is the correct choice. |
| Unix sockets | Named pipes (FIFO) | If targeting systems without Unix socket support. Named pipes are messier on macOS — Unix sockets are simpler and well-supported. |
| Unix sockets | `net/http` webhook | If you want HTTP semantics or need to call from remote hosts. Overkill for same-machine IPC. |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `github.com/charmbracelet/bubbletea/v2` | Different import path, `View()` returns `tea.View` (not string), `KeyMsg` split into `KeyPressMsg`/`KeyReleaseMsg` — incompatible with v1 Bubbles and Lip Gloss v1 | `github.com/charmbracelet/bubbletea@v1.3.10` |
| `github.com/charmbracelet/bubbles/v2` | Requires Bubble Tea v2, uses functional option constructors and getter/setter methods — API not compatible with v1 program | `github.com/charmbracelet/bubbles@v1.0.0` |
| `github.com/charmbracelet/lipgloss/v2` | Requires explicit `HasDarkBackground()` calls; `Color` returns `color.Color` not `TerminalColor` — needs v2-specific wiring | `github.com/charmbracelet/lipgloss@v1.1.0` |
| `fsnotify.AddWith(..., WithBufferSize(...))` | `WithBufferSize` is Windows-only — no-op on macOS/kqueue | `watcher.Add(path)` (plain Add) |
| CGO or C bindings | Violates static binary requirement; breaks cross-compilation | Pure Go packages only |
| `gopkg.in/yaml.v2` | Older API; v3 adds better error messages, `Node` type for complex parsing, and `inline` struct embedding | `gopkg.in/yaml.v3` |

---

## Stack Patterns by Variant

**If distributing as a plugin (namespaced slash commands):**
- Plugin structure with `.claude-plugin/plugin.json` at root
- Slash command becomes `/gsd-watch:gsd-watch`
- Install via: `claude plugin install --scope project ./gsd-watch-plugin/`

**If distributing as standalone (short slash command `/gsd-watch`):**
- Place `gsd-watch.md` in `.claude/commands/gsd-watch.md` directly
- Hooks go in `.claude/settings.local.json` under `"hooks"` key
- Simpler setup, not shareable as a plugin
- **This is the approach PROJECT.md implies** (manual install, no marketplace)

**If building for macOS universal binary:**
- Compile arm64 and amd64 separately then `lipo -create` to merge
- Both targets work with CGO_ENABLED=0 (pure Go)
- Test on both architectures before distributing

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| bubbletea@v1.3.10 | lipgloss@v1.1.0 | Both use the same string-based View() contract |
| bubbletea@v1.3.10 | bubbles@v1.0.0 | Bubbles v1 targets Bubble Tea v1.x explicitly |
| bubbletea@v1.3.10 | fsnotify@v1.9.0 | No shared interface — used in separate goroutines |
| lipgloss@v1.1.0 | bubbles@v1.0.0 | Bubbles components return strings compatible with Lip Gloss styling |
| Go 1.22+ | all above | Loop variable semantics fix required; all charmbracelet libs support 1.22+ |
| bubbletea@v1.x | bubbletea@v2 | NOT compatible — different import paths, different View() return type |

---

## Sources

- `pkg.go.dev/github.com/charmbracelet/bubbletea?tab=versions` — verified v1.3.10 as latest v1.x (Sep 2025)
- `pkg.go.dev/github.com/charmbracelet/lipgloss?tab=versions` — verified v1.1.0 as latest v1.x (Mar 2025)
- `pkg.go.dev/github.com/charmbracelet/bubbles?tab=versions` — verified v1.0.0 as latest stable (Feb 2026)
- `pkg.go.dev/github.com/fsnotify/fsnotify?tab=versions` — verified v1.9.0 (Apr 2025), confirmed no recursive support
- `pkg.go.dev/gopkg.in/yaml.v3` — verified v3.0.1
- `pkg.go.dev/github.com/charmbracelet/bubbletea` — Program type, Send(), WindowSizeMsg, ProgramOption patterns — HIGH confidence
- `pkg.go.dev/github.com/charmbracelet/lipgloss@v1.1.0` — JoinVertical, JoinHorizontal, Place, borders — HIGH confidence
- `code.claude.com/docs/en/plugins` — plugin.json manifest schema, commands/ vs skills/, hooks.json in plugin — HIGH confidence
- `code.claude.com/docs/en/hooks` — Stop/SubagentStop hook input payload, async field, shell script invocation — HIGH confidence
- `code.claude.com/docs/en/plugins-reference` — complete manifest schema, directory structure, env vars — HIGH confidence
- `go.dev/doc/go1.22` — loop variable semantics, range over integers — HIGH confidence

---

*Stack research for: Go terminal TUI sidebar with fsnotify and Claude Code plugin integration*
*Researched: 2026-03-18*
