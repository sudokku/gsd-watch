# Architecture Research

**Domain:** Go Terminal TUI with Bubble Tea Elm Architecture, fsnotify file watcher, Unix socket IPC
**Researched:** 2026-03-18
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        External Event Sources                        │
│                                                                      │
│  ┌───────────────────┐        ┌──────────────────────────────────┐  │
│  │  fsnotify.Watcher  │        │    Unix Socket Listener          │  │
│  │  (goroutine loop)  │        │    /tmp/gsd-watch-<hash>.sock   │  │
│  └────────┬──────────┘        └──────────────┬───────────────────┘  │
│           │ debounced 300ms                   │ on connect           │
│           │ FileChangedMsg                    │ RefreshMsg           │
└───────────┼───────────────────────────────────┼─────────────────────┘
            │ p.Send(msg)                        │ p.Send(msg)
            ▼                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     tea.Program Event Loop                           │
│                                                                      │
│  msgs channel (unbuffered) ← keyboard input, window resize,         │
│                               FileChangedMsg, RefreshMsg             │
│                                                                      │
│  Update(msg tea.Msg) → (Model, tea.Cmd)   [serial, never blocked]   │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
            ┌──────────────────┼──────────────────┐
            ▼                  ▼                   ▼
┌───────────────────┐  ┌──────────────┐  ┌─────────────────────────┐
│   parser package  │  │  tree model  │  │   header / footer        │
│                   │  │  (collapsible│  │   sub-models (lipgloss)  │
│  parse.Project()  │  │   state)     │  │                          │
│  parse.State()    │  │              │  │  ProjectName, Progress   │
│  parse.Phase()    │  │  cursor int  │  │  LastUpdated, Action     │
│  parse.Plan()     │  │  expanded    │  │  KeyBindings             │
└────────┬──────────┘  │  map[int]bool│  └─────────────────────────┘
         │             └──────────────┘
         │ ProjectData struct
         ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Root Model                                    │
│                                                                      │
│  project   ProjectData      ← parsed .planning/ snapshot             │
│  tree      TreeModel        ← collapsible tree state                 │
│  header    HeaderModel      ← top bar render                         │
│  footer    FooterModel      ← bottom bar render                      │
│  viewport  bubbles.Viewport ← scrollable middle region               │
│  loading   bool             ← initial parse in progress              │
│  err       error            ← graceful error display                 │
└──────────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     View() → string                                  │
│  lipgloss.JoinVertical(                                              │
│    header.View(),                                                    │
│    tree.View(),    ← rendered with status icons, badges, indents     │
│    footer.View(),                                                    │
│  )                                                                   │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Package / Location |
|-----------|---------------|-------------------|
| `watcher` | Wrap fsnotify, walk `.planning/` dirs on start, debounce events via timer, call `p.Send(FileChangedMsg{path})` | `internal/watcher/` |
| `socket` | Listen on Unix socket, accept connections, call `p.Send(RefreshMsg{})` on signal, handle stale socket cleanup on startup | `internal/socket/` |
| `parser` | Parse `STATE.md`, `ROADMAP.md`, `config.json`, `*-PLAN.md` YAML frontmatter into typed structs; treat STATE.md as best-effort | `internal/parser/` |
| `tree` | Hold collapsible state, cursor position, render tree rows with icons and badges | `internal/tui/tree/` |
| `header` | Render project name, model profile, mode, progress bar | `internal/tui/header/` |
| `footer` | Render current GSD action, time-since-last-update, keybindings | `internal/tui/footer/` |
| Root `Model` | Own all sub-models, route messages, orchestrate re-parse on events | `internal/tui/model.go` |
| `main` | Wire up `tea.Program`, start watcher and socket goroutines, pass `*tea.Program` reference to both | `cmd/gsd-watch/main.go` |
| Plugin | Claude Code plugin: slash command, Stop/SubagentStop hooks, signal script | `.claude-plugin/` + `hooks/` + `commands/` |

## Recommended Project Structure

```
gsd-watch/
├── cmd/
│   └── gsd-watch/
│       └── main.go              # Program entry point: wire, start, run
├── internal/
│   ├── watcher/
│   │   └── watcher.go           # fsnotify wrapper with debounce and dir-walk
│   ├── socket/
│   │   └── socket.go            # Unix socket listener and stale socket cleanup
│   ├── parser/
│   │   ├── parser.go            # Orchestrates full and incremental parse
│   │   ├── roadmap.go           # Parse ROADMAP.md → []Phase
│   │   ├── plan.go              # Parse *-PLAN.md YAML frontmatter → Plan
│   │   ├── state.go             # Parse STATE.md → StateInfo (best-effort regex)
│   │   ├── config.go            # Parse config.json → Config
│   │   └── types.go             # ProjectData, Phase, Plan, StateInfo structs
│   └── tui/
│       ├── model.go             # Root Model: Init, Update, View
│       ├── messages.go          # All custom tea.Msg types
│       ├── keys.go              # Global key bindings
│       ├── tree/
│       │   ├── model.go         # TreeModel: collapsible state, cursor
│       │   ├── view.go          # Tree rendering: icons, badges, indents
│       │   └── keys.go          # Tree-specific key bindings
│       ├── header/
│       │   └── model.go         # HeaderModel: project name, progress bar
│       └── footer/
│           └── model.go         # FooterModel: action, last-updated, keys
├── .claude-plugin/
│   └── plugin.json              # Plugin manifest (name, version, hooks path)
├── commands/
│   └── gsd-watch.md             # /gsd-watch slash command definition
├── hooks/
│   └── hooks.json               # Stop and SubagentStop hook config
├── scripts/
│   └── gsd-watch-signal.sh      # Shell script: send refresh via socat/nc
├── Makefile                     # build, install, plugin-install, all, clean
└── go.mod
```

### Structure Rationale

- **`internal/`:** Prevents accidental import as a library; all non-main packages here.
- **`internal/watcher/` and `internal/socket/`:** Isolated packages with no knowledge of Bubble Tea model — they hold only a `*tea.Program` reference and call `p.Send()`. This keeps the event source decoupled from the model.
- **`internal/parser/`:** Pure functions — no goroutines, no state. Input is the `.planning/` path, output is `ProjectData`. Safe to call from any goroutine or `tea.Cmd`.
- **`internal/tui/`:** All Bubble Tea concerns. Sub-packages (`tree/`, `header/`, `footer/`) each implement the `Model/Update/View` triad and are composed by the root model.
- **`internal/tui/messages.go`:** Single file listing every custom `tea.Msg` type. Makes message flow auditable at a glance.
- **`scripts/`:** Lives outside the plugin to be available in `$PATH` after `make install`. Referenced by hooks via `${CLAUDE_PLUGIN_ROOT}/../scripts/`.

## Architectural Patterns

### Pattern 1: External Goroutine → tea.Program.Send()

**What:** External goroutines (watcher, socket) hold a `*tea.Program` reference obtained before `p.Run()` is called. They call `p.Send(msg)` to inject typed messages into the event loop. The event loop processes them serially in `Update()`.

**When to use:** Any persistent background goroutine that produces events asynchronously — file watchers, socket listeners, timers not expressible as `tea.Cmd`.

**Trade-offs:** Simple and safe. `p.Send()` is non-blocking on shutdown (uses `select` with context), so goroutines do not need to check if the program is still running. The one risk is that `p.Send()` blocks while the event loop is busy — this is intentional backpressure and acceptable for low-frequency events like file changes.

**Example:**
```go
// internal/watcher/watcher.go
type Watcher struct {
    program *tea.Program
    fw      *fsnotify.Watcher
    timer   *time.Timer
}

func (w *Watcher) loop() {
    for {
        select {
        case event, ok := <-w.fw.Events:
            if !ok { return }
            // Reset debounce timer on every event
            if w.timer != nil {
                w.timer.Stop()
            }
            path := event.Name
            w.timer = time.AfterFunc(300*time.Millisecond, func() {
                w.program.Send(msgs.FileChangedMsg{Path: path})
            })
        case err, ok := <-w.fw.Errors:
            if !ok { return }
            w.program.Send(msgs.WatcherErrorMsg{Err: err})
        }
    }
}
```

### Pattern 2: tea.Cmd for One-Shot Async Work (Parse)

**What:** When `Update()` receives a `FileChangedMsg` or `RefreshMsg`, it returns a `tea.Cmd` that runs the parser in a goroutine, then returns a `ParsedMsg` with the result. The model applies `ParsedMsg` to update its `project` field.

**When to use:** Work that needs to run once in response to a message — I/O-bound operations like file parsing that should not block the event loop.

**Trade-offs:** Clean separation — the model never calls I/O directly. Multiple in-flight parse commands can exist if events arrive quickly; the incremental cache in the parser mitigates redundant work. Accept last-wins semantics: whichever `ParsedMsg` arrives last wins.

**Example:**
```go
// internal/tui/model.go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case msgs.FileChangedMsg:
        return m, parseCmd(m.planningDir, msg.Path) // tea.Cmd
    case msgs.ParsedMsg:
        m.project = msg.Project
        m.tree = m.tree.SetData(msg.Project)
        return m, nil
    }
    // ...
}

func parseCmd(dir, changedPath string) tea.Cmd {
    return func() tea.Msg {
        project, err := parser.ParseIncremental(dir, changedPath)
        if err != nil {
            return msgs.ParseErrorMsg{Err: err}
        }
        return msgs.ParsedMsg{Project: project}
    }
}
```

### Pattern 3: Collapsible Tree State in Model

**What:** The `TreeModel` holds an `expanded map[int]bool` keyed by a stable node index (computed from phase/plan path, not position, to survive re-parses). Keyboard messages (`KeyLeft`/`h`, `KeyRight`/`l`) toggle expansion. The `cursor` is a flat index over currently-visible rows.

**When to use:** Any hierarchical data structure with stable node identities that needs to survive data refreshes without collapsing everything.

**Trade-offs:** Using a path-based key (e.g., `"milestone-1/phase-2"`) for `expanded` means re-parses do not reset user's expand/collapse state. Using a flat cursor over visible rows simplifies scroll math but requires recomputing visible rows on every data update.

**Example:**
```go
// internal/tui/tree/model.go
type TreeModel struct {
    data     parser.ProjectData
    expanded map[string]bool  // key: "phase/<id>" or "milestone/<id>"
    cursor   int              // index into visibleRows()
}

func (t TreeModel) visibleRows() []Row {
    // Flatten tree: include children only if parent is expanded
}

func (t TreeModel) Update(msg tea.Msg) (TreeModel, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "left", "h":
            row := t.visibleRows()[t.cursor]
            t.expanded[row.Key] = false
        case "right", "l":
            row := t.visibleRows()[t.cursor]
            t.expanded[row.Key] = true
        case "up", "k":
            if t.cursor > 0 { t.cursor-- }
        case "down", "j":
            rows := t.visibleRows()
            if t.cursor < len(rows)-1 { t.cursor++ }
        }
    }
    return t, nil
}
```

### Pattern 4: Debounce via time.AfterFunc (Outside Tea)

**What:** Debouncing is implemented in the watcher goroutine using `time.AfterFunc`, not inside Bubble Tea's event loop. Each fsnotify event resets the timer. Only after 300ms of inactivity does `p.Send(FileChangedMsg{})` fire.

**When to use:** High-frequency external events that must be coalesced before entering the event loop. Doing this outside the event loop prevents the model's `Update` from being called dozens of times per second during a multi-file write.

**Trade-offs:** `time.AfterFunc` runs its callback in a new goroutine, so the `p.Send()` call is already goroutine-safe. The downside is only the last-changed path is passed; if multiple files changed during the 300ms window, the incremental parser receives only one path and must detect other changes itself — or use a full re-parse on socket-triggered refresh.

## Data Flow

### File System → Rendered View

```
.planning/ directory on disk
    │
    ▼ (startup: full parse)
parser.Parse(dir) → ProjectData
    │
    ▼
Root Model receives ParsedMsg → m.project = ProjectData
    │
    ▼
TreeModel.SetData(project) → recomputes visibleRows(), preserves expanded state
    │
    ▼
View() → header.View() + tree.View() + footer.View() → rendered string
    │
    ▼ (lipgloss styles applied inline)
Terminal output (alt screen)
```

### fsnotify Event → Re-render

```
File write on disk
    │
    ▼ (fsnotify.Events channel)
watcher goroutine: debounce 300ms via time.AfterFunc
    │
    ▼ p.Send(FileChangedMsg{path})
tea.Program event loop: msgs channel
    │
    ▼ Update() receives FileChangedMsg
tea.Cmd: parser.ParseIncremental(dir, path) in goroutine
    │
    ▼ returns ParsedMsg
Update() receives ParsedMsg → m.project updated
    │
    ▼ View() called automatically by Bubble Tea
Terminal re-renders
```

### Unix Socket Signal → Re-render

```
Claude Code Stop hook runs gsd-watch-signal.sh
    │
    ▼ (socat/nc writes "refresh" to /tmp/gsd-watch-<hash>.sock)
socket goroutine: conn.Accept()
    │
    ▼ p.Send(RefreshMsg{})
tea.Program event loop: msgs channel
    │
    ▼ Update() receives RefreshMsg
tea.Cmd: parser.Parse(dir) full re-parse in goroutine
    │  (full re-parse on socket signal, not incremental)
    ▼ returns ParsedMsg
Same path as above → View() re-renders
```

### State Management

```
Root Model (immutable value, replaced on each Update)
    ├── project ProjectData     ← replaced atomically on ParsedMsg
    ├── tree    TreeModel       ← updated on key events and ParsedMsg
    ├── header  HeaderModel     ← derived from project on every View()
    └── footer  FooterModel     ← updated on tick (time-since-last-update)

No shared mutable state. No mutexes. All writes go through Update().
```

### Key Data Flows

1. **Startup parse:** `main.go` launches initial parse as the first `tea.Cmd` from `Init()`, so the TUI displays immediately (with a spinner or "loading" state) and updates when parsing completes.
2. **Incremental cache:** `parser` package tracks last-parsed mtime per file path; `ParseIncremental(dir, changedPath)` only re-parses the changed file and returns a fully merged `ProjectData`. Full re-parse only on startup and socket signal.
3. **Window resize:** Bubble Tea automatically delivers `tea.WindowSizeMsg`; root model propagates width/height to `tree`, `header`, `footer` so they can re-flow their layout.

## Scaling Considerations

This is a single-user local TUI binary. Scaling dimensions are irrelevant. The practical performance constraints are:

| Concern | Constraint | Mitigation |
|---------|-----------|-----------|
| Parse latency | 50+ PLAN.md files | Incremental cache; only re-parse changed file on fsnotify |
| Render frequency | fsnotify burst during execute-phase | 300ms debounce in watcher goroutine |
| Memory | Static binary under 15MB | No dynamic plugins, no Node.js runtime |
| Startup time | First render should be near-instant | Show loading state immediately; parse async via Init() Cmd |

## Anti-Patterns

### Anti-Pattern 1: Calling Parser Inside Update()

**What people do:** Call `parser.Parse()` synchronously inside the `Update()` function when a `FileChangedMsg` arrives.

**Why it's wrong:** `Update()` must return fast — it blocks the entire event loop, making the TUI unresponsive during parse. With 50+ files this will cause visible lag.

**Do this instead:** Return a `tea.Cmd` from `Update()` that runs the parser in a goroutine and sends back a `ParsedMsg`.

### Anti-Pattern 2: Accessing Model State from Watcher/Socket Goroutines

**What people do:** Pass a pointer to the model into the watcher goroutine to read/write fields directly.

**Why it's wrong:** `Update()` is the sole writer of model state and runs serially. Any concurrent access creates a data race. Go's race detector will catch it.

**Do this instead:** Goroutines hold only a `*tea.Program` reference and communicate exclusively via `p.Send(msg)`.

### Anti-Pattern 3: Using a Single Expanded-by-Position Key

**What people do:** Key the `expanded` map by cursor position (`expanded[0]`, `expanded[1]`, ...) rather than by node identity.

**Why it's wrong:** After a re-parse, the tree may have different nodes at the same positions (e.g., a phase was completed and a new one appeared). The user's expand/collapse state gets applied to the wrong nodes.

**Do this instead:** Key by a stable string derived from the node's identity (e.g., phase directory name or `milestone-<n>/phase-<n>`).

### Anti-Pattern 4: Watching Only the Root .planning/ Dir

**What people do:** Pass only `.planning/` to `fsnotify.Watcher.Add()` and assume it watches recursively.

**Why it's wrong:** macOS kqueue (which fsnotify uses on darwin) does not support recursive watching. Changes in `.planning/milestone-1/phase-2/` will not fire events.

**Do this instead:** On startup, walk `.planning/` recursively and call `watcher.Add()` for every subdirectory found. When a new directory is created (rare), add it dynamically.

### Anti-Pattern 5: Leaving a Stale Unix Socket File

**What people do:** Create `/tmp/gsd-watch-<hash>.sock` on startup without checking if it already exists.

**Why it's wrong:** If the process was killed (SIGKILL), the socket file remains. `net.Listen("unix", path)` will return `bind: address already in use`.

**Do this instead:** On startup, attempt `net.Dial("unix", path)`. If it succeeds, another instance is running — exit or warn. If it fails, `os.Remove(path)` the stale file, then `net.Listen`.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|--------------------|----|
| fsnotify v1.x | Watcher goroutine in `internal/watcher/`; `p.Send()` bridge | Must manually walk and add all subdirs on darwin |
| Unix socket (`net` stdlib) | Listener goroutine in `internal/socket/`; `p.Send()` bridge | Socket path: `/tmp/gsd-watch-<sha256-of-abs-path[:8]>.sock` |
| Claude Code plugin | hooks in `hooks/hooks.json`; signal script in `scripts/` | Stop and SubagentStop events call `gsd-watch-signal.sh` |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|--------------|-------|
| `watcher` → Root Model | `p.Send(msgs.FileChangedMsg{Path})` | Debounce handled in watcher; model receives coalesced events |
| `socket` → Root Model | `p.Send(msgs.RefreshMsg{})` | Triggers full re-parse, not incremental |
| Root Model → `parser` | `tea.Cmd` (goroutine, returns `msgs.ParsedMsg`) | Pure function call; parser has no knowledge of tea |
| Root Model → `tree` | Direct method call in `Update()`: `m.tree = m.tree.Update(msg)` | Tree is a value type; returned as updated copy |
| Root Model → `header`/`footer` | Direct method call in `Update()` | Same pattern as tree |
| `parser` → `internal/tui` | `parser.ProjectData` struct | Defined in `internal/parser/types.go`; tui imports parser, not vice versa |

## Plugin Architecture

### Claude Code Plugin Structure

```
gsd-watch/                     ← plugin root (same as repo root)
├── .claude-plugin/
│   └── plugin.json            ← manifest (name, version, hooks reference)
├── commands/
│   └── gsd-watch.md           ← /gsd-watch slash command
├── hooks/
│   └── hooks.json             ← Stop and SubagentStop hook definitions
└── scripts/
    └── gsd-watch-signal.sh    ← sends "refresh" to the socket
```

### plugin.json

```json
{
  "name": "gsd-watch",
  "version": "1.0.0",
  "description": "Live GSD project status sidebar for Claude Code",
  "hooks": "./hooks/hooks.json"
}
```

### hooks/hooks.json

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/gsd-watch-signal.sh"
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/gsd-watch-signal.sh"
          }
        ]
      }
    ]
  }
}
```

### scripts/gsd-watch-signal.sh

The signal script must locate the correct socket path (derived from the project directory `$PWD`), check the socket exists, and send a refresh signal:

```bash
#!/usr/bin/env bash
# Compute same hash as the Go binary uses
HASH=$(echo -n "$PWD" | shasum -a 256 | cut -c1-8)
SOCK="/tmp/gsd-watch-${HASH}.sock"
[ -S "$SOCK" ] && echo -n "refresh" | nc -U "$SOCK" -q 1 2>/dev/null || true
```

The script must exit 0 regardless of whether the TUI is running — hooks must not block Claude Code.

### commands/gsd-watch.md

```markdown
---
name: gsd-watch
description: Start gsd-watch status sidebar in a tmux split pane
---

Start the gsd-watch sidebar by opening a new tmux pane and launching the binary:

tmux split-window -h -l 40 "gsd-watch"

If you are not in a tmux session, instruct the user to run `gsd-watch` manually
in a tmux split pane. Do not attempt to automate tmux wrapping of the current session.
```

## Suggested Build Order

Build in dependency order — packages with no dependencies first:

1. **`internal/parser/types.go`** — Define `ProjectData`, `Phase`, `Plan`, `StateInfo` structs. Everything imports this.
2. **`internal/parser/`** — Implement full parse. No Bubble Tea dependency. Fully testable in isolation.
3. **`internal/tui/messages.go`** — Define all custom `tea.Msg` types. Required by both watcher/socket and the TUI.
4. **`internal/watcher/`** — Implement fsnotify wrapper. Depends only on `messages.go` and `*tea.Program`.
5. **`internal/socket/`** — Implement Unix socket listener. Depends only on `messages.go` and `*tea.Program`.
6. **`internal/tui/tree/`** — Implement tree model and renderer. Depends on `parser/types.go`. Core visual component.
7. **`internal/tui/header/` and `internal/tui/footer/`** — Implement header and footer models.
8. **`internal/tui/model.go`** — Assemble root model, wire all sub-models, implement `Init/Update/View`.
9. **`cmd/gsd-watch/main.go`** — Wire everything: create program, start watcher and socket goroutines, call `p.Run()`.
10. **Plugin files** — `plugin.json`, `hooks.json`, `gsd-watch.md`, `gsd-watch-signal.sh`. No Go compilation required.
11. **`Makefile`** — `build`, `install`, `plugin-install`, `all`, `clean` targets.

**Rationale:** Parser first because it has no dependencies and all other packages need its types. Watcher and socket second because they are simple goroutine wrappers. Tree component before root model because the root model composes it. Main last because it wires everything — it's the thinnest layer.

## Sources

- [Bubble Tea pkg.go.dev — v1.3.10 API reference](https://pkg.go.dev/github.com/charmbracelet/bubbletea)
- [charmbracelet/bubbletea GitHub](https://github.com/charmbracelet/bubbletea)
- [charmbracelet/bubbles GitHub](https://github.com/charmbracelet/bubbles)
- [Concurrency and Goroutines — DeepWiki Bubble Tea](https://deepwiki.com/charmbracelet/bubbletea/5.1-concurrency-and-goroutines)
- [Tips for building Bubble Tea programs — leg100.github.io](https://leg100.github.io/en/posts/building-bubbletea-programs/)
- [Managing nested models with Bubble Tea — donderom.com](https://donderom.com/posts/managing-nested-models-with-bubble-tea/)
- [Debounce discussion — charmbracelet/bubbletea #601](https://github.com/charmbracelet/bubbletea/discussions/601)
- [Injecting messages from outside the program loop — Issue #25](https://github.com/charmbracelet/bubbletea/issues/25)
- [tree-bubble package](https://pkg.go.dev/github.com/savannahostrowski/tree-bubble)
- [Claude Code Plugins Reference](https://code.claude.com/docs/en/plugins-reference)

---
*Architecture research for: Go Bubble Tea TUI with fsnotify + Unix socket IPC*
*Researched: 2026-03-18*
