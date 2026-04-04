# gsd-watch

## What This Is

A terminal-based companion sidebar for Claude Code that displays a live, real-time view of the current GSD v1 project state. It runs as a compiled Go TUI binary in a tmux split pane, reading `.planning/` on disk and rendering a collapsible tree of milestones, phases, plans, and task statuses. Targeted at macOS users of Claude Code + GSD v1 who want a persistent status panel while working.

## Core Value

A developer running GSD can always see exactly where they are in their project — without context-switching out of Claude Code — and the view updates automatically within one second of any GSD action completing.

## Requirements

### Validated (v1.0)

- ✓ TUI binary renders collapsible phase/plan tree from `.planning/` directory — v1.0
- ✓ File watcher (fsnotify, recursive via manual dir-walk) triggers re-render on changes — v1.0
- ✓ Header shows project name, model profile, mode, and progress bar — v1.0
- ✓ Tree shows phase status icons and plan status icons — v1.0
- ✓ Phase lifecycle badges shown (discussed, researched, verified, UAT) — v1.0
- ✓ Footer shows current GSD action, time-since-last-update, and keybindings — v1.0
- ✓ Keyboard navigation: ↑↓/jk scroll, ←→/hl collapse/expand, q/Ctrl+C quit — v1.0
- ✓ Claude Code plugin with `/gsd-watch` slash command — v1.0
- ✓ Makefile with `build`, `install`, `plugin-install`, `all`, `clean` targets — v1.0
- ✓ Static Go binary under 15MB, zero runtime dependencies except tmux — v1.0
- ✓ Graceful handling of missing/malformed files — never crash — v1.0
- ✓ `--help` flag, outside-tmux detection with clear error message — v1.0
- ✓ Project README with installation, usage, and contributing docs — v1.0

### Validated (v1.1)

- ✓ PARSE-09: Parser sorts phases absent from ROADMAP.md by extracting number from dir name — v1.1
- ✓ PARSE-10: Parser handles BOM and leading whitespace in PLAN.md frontmatter — v1.1
- ✓ PARSE-11: ROADMAP.md phase heading detection works for H2, H3, and H4 formats — v1.1
- ✓ PARSE-12: App shows project name from PROJECT.md when STATE.md milestone_name is missing — v1.1
- ✓ OBS-01: `--debug` flag prints parser decisions to stderr — v1.1
- ✓ TEST-01: Test fixture corpus covers BOM, alternate headings, and missing-from-ROADMAP phases — v1.1
- ✓ QT-01: Collapsible "Quick tasks" section in TUI tree reading `.planning/quick/` — v1.1
- ✓ QT-02: Quick task parser detects `YYMMDD-ID-PLAN.md` / `YYMMDD-ID-SUMMARY.md` convention — v1.1
- ✓ A11Y-01: `--no-emoji` flag switches all TUI status icons and badges to ASCII equivalents — v1.1

### Validated (v1.2)

- ✓ ARC-01: Parser detects archived milestone directories and extracts completion metadata — v1.2 (Phase 11)
- ✓ ARC-02: User sees collapsed, non-interactive row per completed archived milestone in TUI tree — v1.2 (Phase 12)

### Out of Scope

- VS Code extension / sidebar — web/GUI UI is out of scope entirely
- GSD v2 support (`.gsd/` directory) — v1 only
- Windows support — macOS and Linux supported; Windows is out of scope
- Mouse interaction — keyboard navigation only
- Triggering or editing GSD commands from the sidebar — read-only
- Cost tracking / token usage display — not enough signal in `.planning/`
- Multi-project support — one project at a time
- Plugin marketplace publishing — manual install only
- Zellij / WezTerm / iTerm2 / other multiplexers — tmux and cmux only

### Validated (v1.3)

- ✓ CFG-01: Missing config → silent defaults — v1.3 (Phase 13)
- ✓ CFG-02: Malformed TOML → fatal error with config path — v1.3 (Phase 13)
- ✓ CFG-03: Unknown config keys → stderr warning, still starts — v1.3 (Phase 13)
- ✓ CFG-04: `--no-emoji` flag overrides config emoji key — v1.3 (Phase 13)
- ✓ CFG-05: `--theme` flag overrides config theme key — v1.3 (Phase 13)
- ✓ THEME-01: `default` preset — zero visual regression from v1.2 — v1.3 (Phase 14)
- ✓ THEME-02: `minimal` preset — muted 256-color, content-first appearance — v1.3 (Phase 14)
- ✓ THEME-03: `high-contrast` preset — bold 16-color ANSI, SSH/degraded-terminal safe — v1.3 (Phase 14)
- ✓ THEME-04: Unknown theme name → stderr warning + fallback to default — v1.3 (Phase 14)
- ✓ DISC-01: Help overlay (`?`) shows config file path — v1.3 (Phase 15)
- ✓ DISC-02: Help overlay shows currently active theme name — v1.3 (Phase 15)
- ✓ COLOR-01: `[theme]` TOML section → `ThemeColors` struct with 5 `*string` override fields — v1.3 (Phase 16)
- ✓ COLOR-02: `ApplyColorOverrides()` wired into `app.New()` after preset resolution — v1.3 (Phase 16)

### Validated (v1.4)

- ✓ BUILD-01: `make build-linux` cross-compiles static arm64 + amd64 ELF binaries — v1.4 (Phase 17)
- ✓ BUILD-02: `build-all` builds all 4 platform binaries as a prerequisite chain — v1.4 (Phase 17)
- ✓ BUILD-03: `install` detects host OS/arch at runtime and dispatches to correct binary — v1.4 (Phase 17)
- ✓ BUILD-04: Missing-binary guard prints actionable error and exits 2 — v1.4 (Phase 17)
- ✓ SPAWN-01: `/gsd-watch` passes multiplexer guard when `$CMUX_WORKSPACE_ID` is set — v1.4 (Phase 19)
- ✓ SPAWN-02: `/gsd-watch` outside both tmux and cmux shows error naming both multiplexers — v1.4 (Phase 19)
- ✓ SPAWN-03: `/gsd-watch` inside cmux creates a right-side split pane — v1.4 (Phase 20)
- ✓ SPAWN-04: New cmux pane automatically runs gsd-watch in the correct project directory — v1.4 (Phase 20)
- ✓ SPAWN-05: tmux code path (Steps 3-4) byte-identical to v1.3 — v1.4 (Phase 20)

## Current State

**In progress:** v1.4 cmux + Linux — Phases 17-20 complete (2026-04-04)
- Go binary detects cmux (`$CMUX_WORKSPACE_ID`) alongside tmux (`$TMUX`)
- OS-aware error message with brew/apt install hints when neither multiplexer present
- Pane title switched from OSC 2 → OSC 0 for cross-multiplexer compatibility
- `/gsd-watch` slash command: cmux → real pane spawning via `cmux new-split right` + `cmux send`, tmux → existing flow, neither → OS-aware error
- Slash command consolidated into single bash script (one tool call, faster execution)

**Shipped:** v1.3 Settings (2026-03-27) — 4 phases, 7 plans, 7,082 Go LOC
- `~/.config/gsd-watch/config.toml` configures emoji, preset theme, and per-color hex overrides
- 3 named themes: `default`, `minimal`, `high-contrast`; unknown theme warns + falls back
- `?` help overlay reveals config path and active theme name
- `[theme]` TOML section applies hex overrides on top of chosen preset

## Current Milestone: v1.4 cmux + Linux

**Goal:** Expand gsd-watch to support cmux as a second terminal multiplexer (macOS) and ship Linux binaries (arm64 + amd64).

**Target features:**
- Linux builds (arm64 + amd64, CGO_ENABLED=0, no codesign)
- Multiplexer detection in Go binary: accept `$TMUX` OR `$CMUX_WORKSPACE_ID`
- Slash command cmux pane spawning via Unix socket API (`nc -U`)
- Documentation: platform matrix, Linux install, cmux usage

**Key constraints:**
- cmux is macOS-only; Linux users always get tmux
- Socket API preferred over cmux CLI for pane spawning (more stable long-term)
- Lockfile-based duplicate detection for cmux deferred to v1.4.1 (best-effort for now)
- Binary release (v1.4.0 tag) at end of milestone

## Future Milestone: v1.5+ (TBD)

**Candidate features from v1.3/v1.4 deferred list:**
- `--config <path>` flag for alternate config file
- `theme = "custom"` preset deferring all colors to `[theme]` table
- Header and footer full theme coverage
- `[behavior].refresh_debounce_ms` tuning
- `[behavior].show_completed_phases` toggle
- Lockfile-based cmux duplicate detection (deferred from v1.4)

## Context

- GSD v1 uses `.planning/` with `STATE.md` (free-form prose, regex-parsed), `ROADMAP.md` (markdown phases), `REQUIREMENTS.md`, `config.json`, and phase directories containing `*-PLAN.md` (YAML frontmatter), `*-SUMMARY.md`, `*-VERIFICATION.md`, `*-UAT.md`, `*-CONTEXT.md`, `*-RESEARCH.md`
- STATE.md is intentionally unstructured (written for LLM consumption). Filesystem structure is the primary source of truth; STATE.md supplements with current action text and milestone name
- The existing `gsd-statusline.js` uses regex matching on STATE.md — same approach, known fragile; mitigate by treating STATE.md as best-effort only
- Target users: builder and small friend group, all macOS, all tmux users, all Claude Code terminal users, all familiar with GSD v1
- Unix socket path convention: `/tmp/gsd-watch-<project-hash>.sock` where hash is derived from the project directory path

## Constraints

- **Tech Stack**: Go 1.22+, Bubble Tea v1.x, Bubbles, Lip Gloss v1.x, fsnotify v1.x, gopkg.in/yaml.v3, net (stdlib) — no substitutions
- **Build**: CGO_ENABLED=0 static binary, darwin/arm64 + darwin/amd64
- **Runtime**: tmux is the only external dependency
- **No services**: No web server, WebSocket, browser, React, or Node.js runtime

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Filesystem is primary source of truth, STATE.md is supplemental | STATE.md is free-form prose; parsing it for structure is fragile. Phase dirs and PLAN.md frontmatter are authoritative | ✓ Good |
| Debounce fsnotify events at 300ms | Rapid file writes during execute-phase would cause render storms without debouncing | ✓ Good — no render storms observed |
| Unix socket IPC deferred to v2; fsnotify alone is the refresh path | fsnotify is sufficient for 3-10 min GSD phase cadence; socket adds complexity | ✓ Good — v1.0 shipped without it |
| On startup, walk .planning/ recursively to add all dirs to fsnotify | macOS kqueue doesn't support recursive watching natively; must add each dir explicitly | ✓ Good |
| Slash command tells user to start tmux manually if not in tmux | Automating tmux-wrapping of a live Claude Code session is too complex and risky | ✓ Good |
| Incremental cache: only re-parse changed file on fsnotify event | 50+ PLAN.md files parsed on every event would cause latency; full re-parse only on startup | ✓ Good — sub-100ms re-parse latency |
| Root model in `internal/tui/app` sub-package | Avoids import cycle: tui/* sub-packages import tui for shared types | ✓ Good |
| All tea.Msg types in single messages.go | Establishes message contract up front for all phases | ✓ Good |
| lipgloss.AdaptiveColor for all colors | Dark/light terminal support without detection logic | ✓ Good |
| Pane title duplicate detection via OSC 2 | Prevents duplicate gsd-watch panes; works across PATH and binary location differences | ✓ Good — required quick fix (260323-re2) to scope to current session only |
| Archive zone pinned outside viewport via ArchiveZoneHeight() + external section append | Avoids storing height in TreeModel; app/model.go reduces vpHeight by archiveH and appends zone below viewport render | ✓ Good — clean separation, View(width,height) height param ended up unused |
| FormatArchiveDate, RenderArchiveRow, RenderArchiveSeparator, RenderArchiveZone exported | tree_test package (external) needs direct access for unit tests | ✓ Good — testability pattern established |
| ArchivedMilestones parsed from vX.Y-phases/ dirs in .planning/milestones/; MILESTONES.md for completion date | Mirrors actual on-disk structure; fallback to empty string if MILESTONES.md absent | ✓ Good — 12 tests covering all edge cases |
| `BurntSushi/toml` for config parsing; `md.Undecoded()` for unknown-key detection | Avoids exposing toml.Key in public API; `Undecoded()` gives []toml.Key cleanly | ✓ Good — UnknownKeysError type works well |
| `flag.Visit()` for CLI override of config-loaded values | Only fires for explicitly-set flags — avoids bool/zero-value ambiguity when flag is absent | ✓ Good — clean override semantics |
| `ThemeColors` uses `*string` pointer fields (nil = not set) | Enables zero-value detection without sentinel strings for optional per-field color overrides | ✓ Good — nil check is clear and idiomatic |
| `Config.Theme` renamed to `Config.Preset`; `ThemeColors` uses `toml:"theme"` tag | TOML field name collision: both `preset` and `[theme]` section needed distinct struct fields | ✓ Good — naming now matches user-facing TOML keys |
| `ApplyColorOverrides` takes `io.Writer` for warnings | `io.Writer` injection enables `bytes.Buffer` in tests without real stderr; same pattern as `DebugOut` from v1.1 | ✓ Good — consistent with established pattern |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-04 — Phase 20 complete: cmux pane spawning live, slash command consolidated to single bash script*
