# Project Research Summary

**Project:** gsd-watch
**Domain:** Go terminal TUI sidebar — filesystem watcher + Claude Code plugin integration
**Researched:** 2026-03-18
**Confidence:** HIGH

## Executive Summary

gsd-watch is a read-only terminal TUI companion that displays GSD project state in real time inside a tmux split pane. The expert approach for this domain is well-established: Bubble Tea v1 (Elm architecture) for the event loop, fsnotify v1.9 for filesystem watching, and Unix domain socket IPC for instant refresh signals from Claude Code's Stop hooks. The key architectural insight across all four research areas is that Bubble Tea's single-threaded Update() function is both the strength and the main constraint — all background goroutines (file watcher, socket listener) must communicate exclusively via `p.Send(msg)` and `tea.Cmd`, never by writing to model state directly.

The recommended implementation path is to build in four dependency-ordered phases: core TUI scaffold first (parser types and static rendering), then the live data layer (fsnotify watcher with debounce and incremental parse cache), then the socket IPC (Claude Code Stop hook integration), and finally the plugin delivery layer (slash command, hooks, Makefile, static binary). This ordering is driven by the dependency graph: every other component depends on the parser types and the Bubble Tea message contract being stable before they are wired in.

The primary risks are all concurrency and platform-specific: model mutation from outside Update() causes data races, fsnotify on macOS requires user-space recursive watching and the timer.Reset() race must be handled for Go 1.22, and stale Unix socket files after SIGKILL prevent clean restarts. All of these have known, documented mitigations that must be built in from the start — they are not fixable by refactoring after the fact.

## Key Findings

### Recommended Stack

The charmbracelet stack (Bubble Tea v1.3.10 + Lip Gloss v1.1.0 + Bubbles v1.0.0) is the correct and only realistic choice for this domain. All three packages are at stable v1.x releases and are explicitly compatible with each other. The critical version constraint is to stay on v1.x for all three — the v2 packages use different import paths, a different `View()` return type, and split `KeyMsg` types, making them incompatible with the v1 ecosystem.

**Core technologies:**
- Go 1.22+: language runtime — loop variable fix eliminates a classic goroutine closure bug in TUI event handlers
- Bubble Tea v1.3.10: TUI event loop and Model/Update/View lifecycle — the standard Go TUI framework, stable v1.x
- Lip Gloss v1.1.0: terminal styling and layout — pairs directly with Bubble Tea v1.x; `JoinVertical`, borders, adaptive color
- Bubbles v1.0.0: reusable components (viewport, key, progress, spinner) — stable release for Bubble Tea v1.x projects
- fsnotify v1.9.0: filesystem event watching — canonical Go watcher; does NOT support recursive watching on macOS/kqueue
- gopkg.in/yaml.v3: YAML frontmatter parsing for `*-PLAN.md` files — standard library, handles struct tags
- Go stdlib `net`: Unix domain socket IPC — no third-party library needed for the socket listener

### Expected Features

The MVP feature set is fully defined and validated against comparable TUI tools (k9s, lazygit). Every item in the v1 list is P1 — the core value proposition (sub-second project state visibility) requires the full set to work. No individual feature can be safely deferred without degrading the tool's primary purpose.

**Must have (table stakes):**
- Filesystem walker + fsnotify watcher with 300ms debounce — nothing works without this
- PLAN.md YAML frontmatter parser + ROADMAP.md parser + STATE.md regex parser — the data layer
- Incremental file cache (mtime-keyed) — required for sub-second latency at scale (50+ files)
- Collapsible tree view with `h/l` expand/collapse, `j/k` scroll, cursor tracking
- Status icons: ✓ ▶ ○ colored via lipgloss, phase lifecycle badges from file presence
- Header: project name + progress bar; footer: current action + time-since-last-update + keybindings
- Unix socket listener with Stop/SubagentStop hook integration for instant refresh
- Terminal resize handling; graceful error states (never crash on malformed or missing files)
- Static binary (CGO_ENABLED=0), darwin/arm64 + darwin/amd64, Makefile with install targets

**Should have (competitive differentiators):**
- Adaptive color via `lipgloss.AdaptiveColor` — respects user terminal theme, zero cost
- Expand-all / collapse-all keybinding — add if tree navigation proves tedious
- Visual feedback (timestamp update) when socket signal is received

**Defer (v2+):**
- Multi-project support — significant state management; one project per invocation is correct for v1
- Linux support — different fsnotify behavior and socket paths; defer until non-macOS users exist
- Color theme configuration — hardcode sensible defaults; rebuild from source for any changes
- GSD v2 (`.gsd/` directory) support — separate product version, not until GSD v2 ships

**Anti-features (explicitly rejected):**
- Mouse click / scroll — tmux captures mouse events; incompatible with target workflow
- Edit / trigger GSD actions from sidebar — read-only by design
- Web/HTTP server — contradicts "no web server" constraint; Unix socket is sufficient

### Architecture Approach

The architecture is a Bubble Tea Elm-style root model that owns four sub-models (tree, header, footer, viewport) and two external goroutines (watcher, socket) that communicate exclusively via `p.Send(msg)`. The parser package is pure functions with no Bubble Tea dependency — safe to call from any tea.Cmd. The watcher and socket packages hold only a `*tea.Program` reference; they never touch model state. This design guarantees that all state mutations flow through `Update()`, making the program correct-by-construction for concurrent access.

**Major components:**
1. `internal/parser/` — pure functions: parse ProjectData from `.planning/` filesystem; no goroutines, no state
2. `internal/watcher/` — fsnotify wrapper with 300ms debounce; calls `p.Send(FileChangedMsg{})` only
3. `internal/socket/` — Unix socket listener; calls `p.Send(RefreshMsg{})` on signal; handles stale socket cleanup
4. `internal/tui/tree/` — collapsible tree model and renderer; path-keyed expanded map survives re-parses
5. `internal/tui/model.go` — root model: routes messages, dispatches tea.Cmd for async parse, composes views
6. `cmd/gsd-watch/main.go` — wires program, starts goroutines, calls `p.Run()`
7. Plugin layer — `hooks/hooks.json`, `commands/gsd-watch.md`, `scripts/gsd-watch-signal.sh`

**Build order mandated by dependencies:** parser types → parser → messages → watcher/socket → tree → header/footer → root model → main → plugin files → Makefile.

### Critical Pitfalls

1. **Model mutation outside Update()** — any goroutine writing to model fields directly creates a data race that Go's race detector catches but is intermittent at runtime. Prevention: goroutines only call `p.Send()`; all model writes happen in `Update()`. Build this pattern in Phase 1 before any background I/O is introduced.

2. **p.Send() called before p.Run() deadlocks** — starting background goroutines in `main()` before `p.Run()` risks the send channel blocking on startup. Prevention: start watcher and socket goroutines from `Init()` commands inside the event loop lifecycle, not from `main()`.

3. **fsnotify kqueue does not watch recursively on macOS** — `watcher.Add(".planning/")` only watches the root directory. All subdirectories must be added explicitly at startup via `filepath.WalkDir`. New directories created at runtime must be detected via `fsnotify.Create` events and added dynamically. This is a design-time requirement, not an optimization.

4. **Stale Unix socket file after SIGKILL prevents restart** — `defer os.Remove()` is not called on SIGKILL. Prevention: on startup, attempt `net.Dial()` to the socket; if connection is refused, call `os.Remove()` unconditionally before `net.Listen()`. The "try-connect, delete-if-dead" pattern must be in Phase 3 from the start.

5. **time.Timer.Reset() race on Go 1.22** — the timer channel drain idiom is required for Go < 1.23. Prevention: use the safe stop-drain-reset pattern, or implement debounce via a `time.After` channel that sidesteps `Reset()` entirely. Document minimum Go version in the Makefile.

## Implications for Roadmap

Based on the dependency graph in ARCHITECTURE.md and the pitfall-to-phase mapping in PITFALLS.md, the research points to a clear 4-phase structure.

### Phase 1: Core TUI Scaffold

**Rationale:** Parser types and the root model's Update/View contract are the foundation everything else depends on. Establishing Bubble Tea's message-passing pattern correctly in Phase 1 prevents the data race pitfalls from ever being introduced. This phase has no external I/O dependencies — it can be built and tested without fsnotify or socket concerns.

**Delivers:** A working static TUI that renders mock project data, handles keyboard navigation (j/k, h/l, q), responds to terminal resize, handles narrow panes without crashing, and has a complete header/footer layout. The collapsible tree model with path-keyed expanded state must be finalized here.

**Addresses:** Table stakes features — collapsible tree, status icons, header, footer, keyboard navigation, terminal resize, graceful error display, loading state.

**Avoids:** Pitfall 1 (model mutation race), Pitfall 9 (narrow pane rendering panic). Both must be addressed at scaffold stage.

**Research flag:** Standard patterns — skip research-phase. Bubble Tea v1 patterns are well-documented and verified.

### Phase 2: Live Data Layer

**Rationale:** With the TUI scaffold stable, wire in the real data sources. The filesystem watcher, YAML parser, incremental cache, and ROADMAP.md/STATE.md parsers belong together because they form a single data pipeline. The parser feeds the tree model; the watcher triggers the parser. The incremental cache is not optional — it must be built alongside the watcher, not added later.

**Delivers:** A fully live TUI that reflects real `.planning/` filesystem state within 300ms of any file change. Incremental cache keeps re-parse latency imperceptible even with 50+ PLAN.md files. Phase lifecycle badges (file presence detection) are implemented here.

**Uses:** fsnotify v1.9.0, gopkg.in/yaml.v3, Go stdlib `filepath.WalkDir`.

**Implements:** `internal/watcher/`, `internal/parser/` (all sub-parsers), incremental cache.

**Avoids:** Pitfall 2 (Send before Run — goroutines start from Init()), Pitfall 3 (kqueue fd exhaustion — watch dirs only), Pitfall 4 (new directory watching — dynamic Add on Create), Pitfall 7 (timer.Reset race — safe drain pattern), Pitfall 8 (YAML frontmatter delimiter fragility — explicit split with graceful fallback).

**Research flag:** Standard patterns — skip research-phase. All fsnotify behaviors on macOS are documented and verified.

### Phase 3: Unix Socket IPC

**Rationale:** The socket listener is an independent subsystem that enhances refresh latency from 300ms to near-instant. It depends on Phase 2's full data pipeline being correct (the RefreshMsg triggers the same parser path as FileChangedMsg). Socket lifecycle — stale socket detection, context cancellation, goroutine cleanup on quit — must all be correct at build time.

**Delivers:** Instant TUI refresh when Claude Code finishes any response (Stop/SubagentStop hook fires, shell script writes to socket, socket goroutine calls p.Send(RefreshMsg{})).

**Uses:** Go stdlib `net`, `context`, `crypto/sha256`.

**Implements:** `internal/socket/`, socket path derivation, stale socket cleanup on startup.

**Avoids:** Pitfall 5 (stale socket on restart — try-connect-then-remove pattern), Pitfall 6 (socket goroutine not stopped on quit — context cancellation with SetDeadline).

**Research flag:** Standard patterns — skip research-phase. Unix socket IPC in Go is thoroughly documented.

### Phase 4: Claude Code Plugin + Delivery

**Rationale:** Plugin files, hooks, slash command, and Makefile are the delivery mechanism. They depend on the binary being correct and the socket path logic being finalized. Claude Code hook integration has two non-obvious edge cases (stop_hook_active guard, async hook silent failure) that must be implemented even though gsd-watch-signal.sh is non-blocking today — the guard is forward-safety for any future additions.

**Delivers:** Installable tool — static binary (darwin arm64 + amd64), `make install`, `/gsd-watch` slash command, Stop/SubagentStop hooks. User installs with `make install && make plugin-install` and the sidebar works immediately.

**Implements:** `hooks/hooks.json`, `commands/gsd-watch.md`, `scripts/gsd-watch-signal.sh`, `Makefile`, `plugin.json`.

**Avoids:** Pitfall 10 (Stop hook infinite loop — stop_hook_active guard in script), Pitfall 11 (async hook silent failure — watcher is primary path; script handles missing socket gracefully), Pitfall 12 (tmux detection — $TMUX check with tmux list-sessions fallback).

**Research flag:** Verify Claude Code hook behavior manually — async hook semantics are well-documented but the `stop_hook_active` guard requires a real Claude Code session to test end-to-end. Plan a manual verification step in this phase.

### Phase Ordering Rationale

- Phases 1-2-3-4 follow the strict dependency graph from ARCHITECTURE.md's "Suggested Build Order" section.
- The parser package has no dependencies and all other packages need its types — it goes first (Phase 1/2 boundary).
- Background goroutines (watcher, socket) depend on the TUI message contract being stable — they come after the scaffold (Phases 2 and 3).
- Plugin delivery comes last because it depends on the binary being correct and the socket path being finalized (Phase 4).
- This ordering also clusters pitfall mitigations by phase: race conditions in Phase 1, filesystem edge cases in Phase 2, socket lifecycle in Phase 3, hook edge cases in Phase 4.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 4 (Claude Code plugin):** The async hook `stop_hook_active` behavior and tmux $TMUX environment propagation require manual testing in a real Claude Code session. Documented patterns are clear, but edge cases are environment-dependent.

Phases with standard patterns (skip research-phase):
- **Phase 1 (Core TUI Scaffold):** Bubble Tea v1 API is thoroughly documented with API references and examples verified at HIGH confidence.
- **Phase 2 (Live Data Layer):** fsnotify macOS behavior, YAML v3, and incremental cache patterns are all documented with specific code patterns in STACK.md and PITFALLS.md.
- **Phase 3 (Unix Socket IPC):** Go stdlib `net` Unix socket patterns are standard and fully covered in documentation.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified against pkg.go.dev; Bubble Tea v1/v2 split confirmed; compatibility matrix explicit |
| Features | HIGH | Feature set grounded in PROJECT.md constraints + comparative analysis against k9s/lazygit; MVP scope is clear |
| Architecture | HIGH | Elm architecture patterns for Bubble Tea are well-documented from official sources and community posts; project structure matches standard Go layout |
| Pitfalls | HIGH (core) / MEDIUM (Claude Code hooks) | Bubble Tea, fsnotify, socket, timer pitfalls from official maintainer discussions; Claude Code async hook edge cases from official docs only — no community validation available |

**Overall confidence:** HIGH

### Gaps to Address

- **Claude Code async hook testing:** The `stop_hook_active` guard and tmux environment propagation can only be validated in a live Claude Code session. Flag this for manual verification during Phase 4 execution.
- **Go version target:** Research targets Go 1.22+ as minimum. Going to 1.23+ would eliminate the timer.Reset() drain pattern entirely. Decide at roadmap time whether to raise the minimum — the Makefile should enforce whatever is chosen.
- **State.md parsing:** STATE.md is parsed with regex as a "best-effort" field. The exact regex for the current-action field is not specified in FEATURES.md — this will need to be derived from the actual STATE.md format during Phase 2.
- **Socket path hash algorithm:** STACK.md uses SHA256 (Go), but ARCHITECTURE.md's signal script uses `shasum -a 256` (shell), and PITFALLS.md mentions CRC32/FNV as alternatives. Standardize on one algorithm and ensure Go binary and shell script produce identical paths — validate this in Phase 3.

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/github.com/charmbracelet/bubbletea` — Program type, Send(), WindowSizeMsg, Init/Update/View contract, v1 vs v2 differences
- `pkg.go.dev/github.com/charmbracelet/lipgloss@v1.1.0` — JoinVertical, JoinHorizontal, Place, borders, AdaptiveColor
- `pkg.go.dev/github.com/charmbracelet/bubbles@v1.0.0` — viewport, key, progress, spinner components
- `pkg.go.dev/github.com/fsnotify/fsnotify@v1.9.0` — macOS kqueue behavior, no recursive watching, event types
- `pkg.go.dev/gopkg.in/yaml.v3` — Unmarshal, struct tags, v3.0.1
- `code.claude.com/docs/en/hooks` — Stop/SubagentStop hook payload, async field, stop_hook_active, stdin format
- `code.claude.com/docs/en/plugins-reference` — plugin.json schema, commands/, hooks/, env vars
- `go.dev/doc/go1.22` — loop variable semantics
- `deepwiki.com/charmbracelet/bubbletea/5.1-concurrency-and-goroutines` — p.Send() thread safety, goroutine patterns
- `github.com/charmbracelet/bubbletea/issues/25` — injecting messages from outside the program loop
- `github.com/fsnotify/fsnotify/issues/18` — recursive watch limitation, user-space workaround pattern

### Secondary (MEDIUM confidence)
- `leg100.github.io/en/posts/building-bubbletea-programs/` — real-world Bubble Tea architecture patterns
- `donderom.com/posts/managing-nested-models-with-bubble-tea/` — nested sub-model composition
- `github.com/charmbracelet/bubbletea/discussions/601` — debounce patterns
- `blogtitle.github.io` — Go timer reset race condition analysis
- `antonz.org/timer-reset/` — Go 1.23 timer reset fix
- `github.com/notify-rs/notify/issues/596` — kqueue "too many open files" analysis
- k9s, lazygit — comparative TUI feature baseline
- Primer Tree View component docs — tree UX heuristics

### Tertiary (LOW confidence)
- `reading.sh` — Claude Code async hooks explainer (community article, not official)
- TUI resize and graceful degradation — notcurses discussion (different TUI framework, pattern transfer)

---
*Research completed: 2026-03-18*
*Ready for roadmap: yes*
