# Requirements: gsd-watch

**Defined:** 2026-03-18
**Core Value:** A developer running GSD can always see exactly where they are in their project — without context-switching out of Claude Code — and the view updates automatically within one second of any GSD action completing.

## v1 Requirements

### TUI Core

- [x] **TUI-01**: User sees a collapsible tree of phases and plans with status icons (✓ complete, ▶ in-progress, ○ pending, ✗ failed)
- [x] **TUI-02**: User can expand/collapse phases with `←`/`→` or `h`/`l` keys; expanded state survives re-parses (path-keyed)
- [x] **TUI-03**: User can scroll the tree with `↑`/`↓` or `j`/`k` keys via a scrollable viewport
- [x] **TUI-04**: User can quit the TUI with `q` or `Ctrl+C`
- [x] **TUI-05**: Header bar shows project name, GSD model profile, and mode (from config.json)
- [x] **TUI-06**: Header bar shows an overall progress bar (calculated from plan completion ratios)
- [x] **TUI-07**: Footer bar shows current GSD action (from STATE.md), time since last state change, and keybinding hints
- [x] **TUI-08**: TUI reflows gracefully on terminal resize without crashing (Lip Gloss width clamped to minimum)
- [x] **TUI-09**: Currently active plan is marked with a `← now` indicator in the tree
- [x] **TUI-10**: Phase lifecycle badges displayed under each phase (📝 discussed, 🔬 researched, 📋 verified, 🧪 UAT)

### Data Parsing

- [x] **PARSE-01**: Parser reads PLAN.md YAML frontmatter (phase, plan, title, status, wave, depends_on); missing fields use sensible defaults
- [x] **PARSE-02**: Parser treats SUMMARY.md presence as the definitive status for a plan (overrides PLAN.md frontmatter status)
- [x] **PARSE-03**: Parser reads ROADMAP.md to extract phase names and checked/unchecked success criteria
- [x] **PARSE-04**: Parser extracts STATE.md fields via regex (Milestone, Phase, Status, Plan, Progress, Stopped at) as best-effort; gracefully falls back to "unknown" if field absent
- [x] **PARSE-05**: Parser reads config.json for model profile and mode
- [x] **PARSE-06**: Parser infers phase lifecycle badge state from file presence (CONTEXT.md, RESEARCH.md, VERIFICATION.md, UAT.md)
- [x] **PARSE-07**: Filesystem directory structure is primary source of truth for phase list; STATE.md is supplemental only
- [x] **PARSE-08**: Any parser failure (missing file, malformed YAML, invalid JSON) is handled gracefully — TUI never crashes, shows "unknown" or skips the item

### File Watching

- [x] **WATCH-01**: fsnotify watcher monitors `.planning/` recursively — all subdirectories added explicitly on startup via filepath.WalkDir
- [x] **WATCH-02**: Newly created directories (e.g. new phase dir) are added to the watcher dynamically on fsnotify.Create events
- [x] **WATCH-03**: File change events are debounced at 300ms — rapid writes during execute-phase produce a single re-parse, not a storm
- [x] **WATCH-04**: On fsnotify event, only the changed file is re-parsed (incremental cache keyed by path + mtime); full re-parse only on startup
- [x] **WATCH-05**: TUI displays updated state within 300ms of any `.planning/` file change

### Plugin & Delivery

- [ ] **PLUGIN-01**: `/gsd-watch` slash command spawns a tmux split pane (35% width, right side) running the gsd-watch binary
- [ ] **PLUGIN-02**: Slash command detects if running inside tmux (`$TMUX`); if not, prints instructions to start tmux manually rather than attempting to wrap the session
- [ ] **PLUGIN-03**: Slash command detects if gsd-watch is already running (socket or pane title check) and avoids spawning a duplicate
- [x] **PLUGIN-04**: Go binary compiles with `CGO_ENABLED=0` to a static binary under 15MB with no runtime dependencies except tmux
- [x] **PLUGIN-05**: Makefile provides `build` (darwin/arm64), `install` (→ ~/.local/bin/), `plugin-install` (→ Claude Code commands dir), `all`, and `clean` targets
- [x] **PLUGIN-06**: Makefile also builds `darwin/amd64` target for Intel Mac friends

## v2 Requirements

### Unix Socket IPC

- **IPC-01**: Unix socket listener receives "refresh\n" signal for sub-100ms TUI update
- **IPC-02**: Stop and SubagentStop hooks signal TUI via gsd-watch-signal.sh on Claude Code response completion
- **IPC-03**: Stale socket file cleanup on startup (try-connect, delete if dead) for clean restarts after SIGKILL
- **IPC-04**: stop_hook_active guard in signal script to prevent infinite hook loops
- **IPC-05**: Socket path derived from project directory hash (consistent between Go binary and shell script)

### Enhancements

- **ENH-01**: Expand-all / collapse-all keybinding (e.g. `e` / `E`)
- **ENH-02**: Visual flash or timestamp update when a refresh event is received
- **ENH-03**: GSD v2 support (`.gsd/` directory structure)
- **ENH-04**: Linux support (inotify-based fsnotify)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Mouse interaction | tmux captures mouse events before TUI process; incompatible with target workflow |
| Triggering GSD commands from sidebar | Read-only by design — sidebar is an observer, not an actor |
| Editing project state from sidebar | Read-only by design |
| VS Code extension / web dashboard | Contradicts terminal-only, no-web-server constraint |
| Multi-project support | One project per invocation is correct for v1 |
| Windows / Linux support | macOS only for now; fsnotify behavior differs |
| Zellij / WezTerm / iTerm2 splits | tmux only |
| Cost tracking / token usage | No signal for this in .planning/ files |
| Plugin marketplace publishing | Manual install only |
| Color theme configuration | Hardcode sensible defaults using lipgloss.AdaptiveColor |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| TUI-01 | Phase 1 | Complete |
| TUI-02 | Phase 1 | Complete |
| TUI-03 | Phase 1 | Complete |
| TUI-04 | Phase 1 | Complete |
| TUI-05 | Phase 1 | Complete (01-01) |
| TUI-06 | Phase 1 | Complete (01-01) |
| TUI-07 | Phase 1 | Complete (01-01) |
| TUI-08 | Phase 1 | Complete |
| TUI-09 | Phase 1 | Complete (01-01) |
| TUI-10 | Phase 1 | Complete (01-01) |
| PARSE-01 | Phase 2 | Complete |
| PARSE-02 | Phase 2 | Complete |
| PARSE-03 | Phase 2 | Complete |
| PARSE-04 | Phase 2 | Complete |
| PARSE-05 | Phase 2 | Complete |
| PARSE-06 | Phase 2 | Complete |
| PARSE-07 | Phase 2 | Complete |
| PARSE-08 | Phase 2 | Complete |
| WATCH-01 | Phase 3 | Complete |
| WATCH-02 | Phase 3 | Complete |
| WATCH-03 | Phase 3 | Complete |
| WATCH-04 | Phase 3 | Complete |
| WATCH-05 | Phase 3 | Complete |
| PLUGIN-01 | Phase 4 | Pending |
| PLUGIN-02 | Phase 4 | Pending |
| PLUGIN-03 | Phase 4 | Pending |
| PLUGIN-04 | Phase 4 | Complete |
| PLUGIN-05 | Phase 4 | Complete |
| PLUGIN-06 | Phase 4 | Complete |

**Coverage:**
- v1 requirements: 29 total
- Mapped to phases: 29
- Unmapped: 0

---
*Requirements defined: 2026-03-18*
*Last updated: 2026-03-19 after 01-01 completion*
