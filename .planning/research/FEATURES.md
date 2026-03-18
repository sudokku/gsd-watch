# Feature Research

**Domain:** Terminal TUI sidebar companion tool (Go, developer tooling, read-only project state display)
**Researched:** 2026-03-18
**Confidence:** HIGH — established TUI ecosystem with well-documented patterns; specific feature decisions grounded in PROJECT.md constraints

## Feature Landscape

### Table Stakes (Users Expect These)

Features any developer TUI sidebar must have. Missing these makes the tool feel unfinished or untrustworthy.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Collapsible tree with expand/collapse | Any hierarchical status view collapses nodes — k9s, lazygit, and file explorers all do this. Users arrive with muscle memory for `h/l` or arrow keys | MEDIUM | Bubble Tea has no tree component; must be hand-rolled using viewport + list model with indent-level tracking |
| Status icons per item (complete / active / upcoming) | Developers expect visual state differentiation at a glance — colored dots, checkmarks, arrows are the norm in k9s, git TUIs, and CI dashboards | LOW | Unicode glyphs (✓ ▶ ○) work in all modern terminals; lipgloss colors the glyphs |
| Keyboard navigation: j/k or ↑/↓ scroll, q/Ctrl+C quit | Vim-style bindings are the default expectation for any Go/terminal developer tool. Absence of j/k is jarring | LOW | Bubble Tea key handling is straightforward; bubbles/key package handles binding definitions |
| Header: project name + progress summary | k9s and lazygit both lead with a contextual header. Without one, the user has no orientation | LOW | Static render each frame; progress bar requires completed/total counts parsed from filesystem |
| Footer: current action + keybindings hint | Lazygit, k9s, bottom — every mature TUI shows keybindings in the footer to reduce cognitive load | LOW | Bubbles `help` component provides a standard keybinding footer view |
| Progress bar for milestone completion | Developer status tools are expected to show proportion complete, not just lists | LOW | Bubbles `progress` component with gradient fill; recomputed on re-render |
| Real-time / near-real-time updates | The core value proposition. A static snapshot is just a file viewer. Without auto-refresh the tool is useless as a companion | MEDIUM | fsnotify + debounce (300ms) + socket IPC; Bubble Tea `tea.Cmd` delivers async events to the update loop |
| Terminal resize handling | Every TUI must respond to `tea.WindowSizeMsg`. A broken layout on resize feels broken overall | LOW | Bubble Tea emits `WindowSizeMsg` automatically; re-render respects new width/height |
| Graceful error states (missing/malformed files) | Filesystem is messy during active GSD runs. Crash-on-bad-input is unacceptable in a persistent companion | LOW | Return zero-value structs on parse error; display placeholder text, never panic |
| Clean exit + terminal state restore | If the raw terminal state is not restored on exit, the user's shell becomes unusable | LOW | Bubble Tea `tea.Quit` command handles this automatically via `tea/v2` lifecycle |
| Time-since-last-update in footer | Immediately answers "is this stale?" — same pattern as tmux status bar, htop, k9s cluster age column | LOW | Store last-refreshed timestamp; compute elapsed on each tick; `tea.Tick` every second |

### Differentiators (Competitive Advantage)

Features that distinguish this tool from a generic file viewer or a hand-rolled tmux status line.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Unix socket IPC for instant refresh on Stop hook | fsnotify has ~300ms debounce lag. The Stop hook signal delivers an immediate, authoritative refresh the moment Claude Code finishes an action — not eventually | MEDIUM | Net listener in a goroutine; `tea.Cmd` bridge from goroutine to Bubble Tea loop via channel; stale socket detection on startup |
| Phase lifecycle badges (discussed / researched / verified / UAT) | Surfaces GSD-specific workflow state that no generic file viewer can infer. Makes the sidebar meaningful for GSD users specifically | MEDIUM | Detect presence of `*-SUMMARY.md`, `*-VERIFICATION.md`, `*-UAT.md`, `*-RESEARCH.md` in phase dirs; badge render as colored tags via lipgloss |
| Incremental cache: re-parse only changed file | With 50+ PLAN.md files, full re-parse on every fsnotify event causes visible latency. Per-file cache makes updates imperceptible | MEDIUM | Map of `filepath -> (mtime, parsedStruct)`; on event, stat the file, compare mtime, re-parse only if changed |
| Adaptive color via lipgloss `AdaptiveColor` | Respects the user's terminal theme (dark or light). A tool that looks broken in a light terminal will be discarded immediately | LOW | `lipgloss.AdaptiveColor{Light: "...", Dark: "..."}` on all color definitions; no manual theme switching needed |
| Sub-1-second update latency end-to-end | The stated core value is "within one second." Most file-watcher TUIs settle for 1-2s. This requires debounce tuned to 300ms, not the default 500ms-1s | LOW | Configuration decision; already committed in PROJECT.md |
| Zero runtime dependencies (static binary) | Install is `cp gsd-watch ~/bin/`. No Node.js, no Python, no shared libs. This is the correct tradeoff for a tiny personal tool | LOW | `CGO_ENABLED=0`; already in constraints; no additional work needed |
| Current GSD action text in footer | STATE.md's "current action" field surfaces exactly what Claude Code is doing right now — the companion tells you what's happening without switching windows | MEDIUM | Regex parse STATE.md for current action field; treat as best-effort (empty string if unreadable) |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem natural to add but should be explicitly rejected.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Mouse click to expand/collapse | Modern TUIs support mouse; seems polished | tmux captures mouse events and routes them to tmux itself, not the pane's process. Mouse interaction breaks the tmux workflow entirely for the target user group. PROJECT.md explicitly lists this as out-of-scope | Keyboard-only: `h/l` or `←/→` to collapse/expand; fast and consistent |
| Edit / trigger GSD actions from sidebar | "While I can see state, I should be able to act" — natural CRUD instinct | This is a read-only companion by design. Adding write paths requires error handling, confirmation dialogs, and undo — complexity that dwarfs the tool's value | Users run GSD commands in the Claude Code pane; sidebar is purely observational |
| Multi-project switching | Power users manage multiple projects | Requires a project picker, persistent state, socket management per project. This is a v2 feature at minimum. The target user has one active project at a time | One project per invocation; restart binary in different directory for different project |
| Mouse scroll / pointer | Seems ergonomic | Same tmux mouse capture problem as click-to-expand. Additionally, Bubble Tea's mouse support requires `tea.EnableMouseAllMotion` which interferes with terminal text selection | `j/k` scroll is faster for keyboard-centric developers anyway |
| Configuration file (colors, keybindings) | Power-user customization | Adds a config schema, migration path, docs, and error states for every option. The target audience is a tiny group with shared preferences | Hardcode sensible defaults; change via source and rebuild |
| Fuzzy search / filter within tree | File explorers have fuzzy find | Adds a text input, filtering state machine, and match highlighting. The tree is small (tens of items, not thousands) — search is unnecessary | Full tree visible with collapse/expand; j/k navigation is sufficient |
| Log / history panel | "Show me what changed" | STATE.md does not contain history. Implementing history requires diff tracking across refresh cycles — a significant state management problem with no data source | Footer shows time-since-last-update; that's the extent of history needed |
| Web/HTTP server or REST API | Remote monitoring | Contradicts the "No web server" constraint directly. Adds a large attack surface and dependency for a personal local tool | Unix socket IPC is sufficient for the one consumer (Stop hook) |
| Automatic tmux pane creation | Convenience on start | Automating tmux-wrapping of a live session is risky (PROJECT.md notes this). Slash command already handles it with a manual instruction path | `/gsd-watch` slash command tells user to run the command in a new pane; user does it once |

## Feature Dependencies

```
[File Parser: PLAN.md / ROADMAP.md / STATE.md]
    └──requires──> [Filesystem Walker + fsnotify Watcher]
                       └──enables──> [Incremental Cache]
                                         └──enables──> [Tree Model (phases/plans/tasks)]
                                                            └──enables──> [Collapsible Tree View]
                                                                               └──enables──> [Status Icons]
                                                                               └──enables──> [Phase Lifecycle Badges]

[Unix Socket Listener]
    └──triggers──> [Force Refresh via tea.Cmd]
                       └──requires──> [Incremental Cache (to apply efficiently)]

[Progress Bar]
    └──requires──> [Tree Model (completed/total counts)]

[Footer: time-since-last-update]
    └──requires──> [tea.Tick (1s interval)]
    └──requires──> [last-refreshed timestamp set on every refresh cycle]

[Header: project name + model profile]
    └──requires──> [config.json parser]

[Footer: current GSD action]
    └──requires──> [STATE.md regex parser]
    └──is independent of──> [Tree Model]

[Adaptive Color]
    └──enhances──> [All rendered components]
    └──has no hard dependency — can be added at any point]

[Keyboard Navigation (j/k, h/l, q)]
    └──requires──> [Bubble Tea key handler]
    └──requires──> [Tree Model (cursor position, collapse state)]
```

### Dependency Notes

- **Collapsible Tree View requires Tree Model:** The rendering logic (indent, glyph, color per node) depends on having a stable data structure that tracks node depth, expand/collapse state, and cursor position.
- **Incremental Cache requires Filesystem Walker:** You need the initial full parse before you can track per-file mtimes for incremental updates.
- **Unix Socket Listener enhances Incremental Cache:** The socket trigger bypasses the fsnotify debounce delay — but it still routes through the same cache invalidation path to avoid duplicating parse logic.
- **Progress Bar requires Tree Model:** Completed/total counts come from the same phase/plan data that drives the tree.
- **Phase Lifecycle Badges require Filesystem Walker:** Badge state is derived from presence/absence of specific files in phase directories, not from PLAN.md frontmatter.

## MVP Definition

### Launch With (v1)

Minimum set that delivers the core value: "always know where you are in your project within one second."

- [ ] Filesystem walker + fsnotify watcher with 300ms debounce — without this, nothing updates
- [ ] PLAN.md YAML frontmatter parser + ROADMAP.md parser + STATE.md regex parser — the data layer
- [ ] Incremental file cache (mtime-keyed) — required for sub-second latency at scale
- [ ] Tree model: phases and plans with status (complete/active/upcoming), depth-aware
- [ ] Collapsible tree view: `h/l` expand/collapse, `j/k` scroll, cursor tracking
- [ ] Status icons: ✓ ▶ ○ colored via lipgloss
- [ ] Phase lifecycle badges: detected from file presence
- [ ] Header: project name + progress bar (from config.json + tree model counts)
- [ ] Footer: current GSD action (STATE.md best-effort) + time-since-last-update + keybindings
- [ ] Unix socket listener: receives "refresh" from Stop hook, triggers immediate re-parse
- [ ] `gsd-watch-signal.sh` + Claude Code Stop/SubagentStop hooks
- [ ] `/gsd-watch` slash command with tmux pane instructions
- [ ] Terminal resize handling: re-render on `WindowSizeMsg`
- [ ] Graceful error states: never crash on malformed or missing files
- [ ] Static binary: `CGO_ENABLED=0`, darwin/arm64 + darwin/amd64
- [ ] Makefile: `build`, `install`, `plugin-install`, `all`, `clean`

### Add After Validation (v1.x)

Features to add if v1 friction points are confirmed.

- [ ] `expand-all` / `collapse-all` keybinding — add if users report frustration navigating deep trees
- [ ] Stale-socket detection improvement — add if users report startup failures after abnormal exits
- [ ] Configurable debounce interval — add if 300ms proves wrong for slow disks or large project trees

### Future Consideration (v2+)

Defer until there is evidence of need beyond the original author.

- [ ] Multi-project support — requires significant state management; defer until multiple users request it
- [ ] GSD v2 (`.gsd/` directory) support — separate product version; not until v2 ships
- [ ] Linux support — different fsnotify behavior, different socket paths; defer until non-macOS users exist
- [ ] Color theme configuration — defer until someone finds the default palette unworkable

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Filesystem watcher + debounce | HIGH | MEDIUM | P1 |
| File parsers (PLAN.md, ROADMAP.md, STATE.md) | HIGH | MEDIUM | P1 |
| Incremental cache | HIGH | MEDIUM | P1 |
| Collapsible tree view | HIGH | MEDIUM | P1 |
| Status icons (✓ ▶ ○) | HIGH | LOW | P1 |
| Header (project name + progress bar) | HIGH | LOW | P1 |
| Footer (action + time-since + keys) | HIGH | LOW | P1 |
| Unix socket IPC + Stop hook integration | HIGH | MEDIUM | P1 |
| Phase lifecycle badges | MEDIUM | MEDIUM | P1 |
| Terminal resize handling | HIGH | LOW | P1 |
| Graceful error states | HIGH | LOW | P1 |
| Adaptive color (lipgloss AdaptiveColor) | MEDIUM | LOW | P2 |
| Expand-all / collapse-all keybinding | LOW | LOW | P2 |
| Stale socket detection | MEDIUM | LOW | P1 — safety, not UX |
| Static binary + Makefile | HIGH | LOW | P1 — delivery mechanism |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | k9s (Kubernetes TUI) | lazygit (Git TUI) | gsd-watch (our approach) |
|---------|----------------------|-------------------|--------------------------|
| Collapsible tree | Resource group tree, expand/collapse | File staging panels, not a tree | Phase/plan tree, h/l keys |
| Real-time refresh | Continuous Kubernetes API polling | Manual refresh on action | fsnotify + socket IPC |
| Status icons | Color-coded resource states | Staged/unstaged symbols | ✓ ▶ ○ with lifecycle badges |
| Header | Cluster + namespace breadcrumb | Branch + repo name | Project + model profile |
| Footer | Keybindings hint | Keybindings hint | Action + time-since + keys |
| Progress indicator | Pod ready ratio in columns | None | Progress bar in header |
| Read-only mode | No (full CRUD) | No (full CRUD) | Yes — by design |
| Mouse support | Yes | Yes | No — tmux incompatibility |
| Theme support | Skin files (YAML) | Hardcoded | AdaptiveColor auto-detect |
| Resize handling | Yes | Yes | Yes (WindowSizeMsg) |
| Graceful errors | Yes | Yes | Yes — never crash |

## Sources

- [Bubble Tea framework](https://github.com/charmbracelet/bubbletea) — architecture, key handling, WindowSizeMsg, tea.Cmd patterns (HIGH confidence, official)
- [Bubbles component library](https://github.com/charmbracelet/bubbles) — viewport, list, progress bar, spinner, help components (HIGH confidence, official)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — AdaptiveColor, style composition (HIGH confidence, official)
- [k9s](https://k9scli.io/) — header/footer patterns, real-time monitoring, icon toggles (HIGH confidence, official)
- [Lazygit](https://github.com/jesseduffield/lazygit) — panel UX, keyboard-first navigation, status indicators (HIGH confidence, official)
- [Tree View UX patterns — Interaction Design for Trees](https://medium.com/@hagan.rivers/interaction-design-for-trees-5e915b408ed2) — expand-all/collapse-all, expand-on-active-child (MEDIUM confidence, design article)
- [Primer Tree View component docs](https://primer.style/components/tree-view/) — parent expansion on active child, UX heuristics (MEDIUM confidence, official GitHub design system)
- [TUI resize and graceful degradation — notcurses discussion](https://github.com/dankamongmen/notcurses/discussions/2160) — resize event handling, minimum-size constraints (MEDIUM confidence)
- PROJECT.md constraints (explicit scope, out-of-scope, and key decisions already made)

---
*Feature research for: gsd-watch — terminal TUI sidebar companion for GSD v1*
*Researched: 2026-03-18*
