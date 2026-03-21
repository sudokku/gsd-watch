# Roadmap: gsd-watch

## Overview

Build a read-only terminal TUI sidebar that shows live GSD project state in a tmux split pane. Four phases follow the strict dependency graph: a static TUI scaffold establishes the Bubble Tea message contract first, then the data parsers and file watcher wire in live data, then the Claude Code plugin and delivery layer make it installable and usable end-to-end.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Core TUI Scaffold** - Bubble Tea root model, collapsible tree, header/footer, keyboard nav — static mock data (completed 2026-03-19)
- [x] **Phase 2: Live Data Layer** - All parsers (PLAN.md, ROADMAP.md, STATE.md, config.json), phase lifecycle badges, wired to TUI (completed 2026-03-19)
- [x] **Phase 3: File Watching** - fsnotify recursive watcher, debounce, incremental cache, sub-300ms live updates (completed 2026-03-20)
- [x] **Phase 4: Plugin & Delivery** - Slash command, Stop hooks, Makefile, static binary, installable (completed 2026-03-21)
- [x] **Phase 5: TUI Polish** - Visual hierarchy improvements, empty-state handling, live refresh indicator, discoverability (completed 2026-03-21)

## Phase Details

### Phase 1: Core TUI Scaffold
**Goal**: Users can launch a working TUI that renders a collapsible phase/plan tree, navigate it with keyboard shortcuts, and see a header and footer — all against static mock data
**Depends on**: Nothing (first phase)
**Requirements**: TUI-01, TUI-02, TUI-03, TUI-04, TUI-05, TUI-06, TUI-07, TUI-08, TUI-09, TUI-10
**Success Criteria** (what must be TRUE):
  1. User can expand and collapse phases with h/l or arrow keys; expand state persists across re-renders
  2. User can scroll a tree taller than the terminal with j/k or arrow keys without the cursor leaving the visible viewport
  3. User can quit the TUI with q or Ctrl+C and return to the shell cleanly
  4. Header shows project name, model profile, mode, and a progress bar; footer shows current action, time since last update, and keybinding hints
  5. Terminal resize reflows the layout without any crash or rendering artifact; narrow panes are clamped to minimum width
**Plans:** 4/4 plans complete

Plans:
- [x] 01-01-PLAN.md — Foundation: types, messages, mock data, keys, styles
- [ ] 01-02-PLAN.md — Collapsible tree component with status icons and badges
- [ ] 01-03-PLAN.md — Header and footer components
- [ ] 01-04-PLAN.md — Root model integration + main.go + visual verification

### Phase 2: Live Data Layer
**Goal**: Users see real project state parsed from `.planning/` — phases, plans with status icons, lifecycle badges, progress bar, and current action — using live files on disk
**Depends on**: Phase 1
**Requirements**: PARSE-01, PARSE-02, PARSE-03, PARSE-04, PARSE-05, PARSE-06, PARSE-07, PARSE-08
**Success Criteria** (what must be TRUE):
  1. Tree reflects actual phase/plan structure from `.planning/` filesystem; PLAN.md frontmatter drives status icons
  2. A plan with a SUMMARY.md shows as complete regardless of its PLAN.md frontmatter status field
  3. Phase lifecycle badges (discussed, researched, verified, UAT) appear under each phase when the corresponding file exists in the phase directory
  4. Header progress bar and footer current-action field reflect values from config.json and STATE.md respectively
  5. Missing, empty, or malformed files (bad YAML, invalid JSON, absent fields) never crash the TUI — affected items show "unknown" or are skipped silently
**Plans:** 3/3 plans complete

Plans:
- [ ] 02-01-PLAN.md — PLAN.md frontmatter parser + config.json parser
- [ ] 02-02-PLAN.md — ROADMAP.md parser + STATE.md parser
- [ ] 02-03-PLAN.md — Phase assembler, badge detection, ParseProject + TUI wiring

### Phase 3: File Watching
**Goal**: Users see the TUI update automatically within 300ms of any `.planning/` file change — including new phase directories and new plan files — without restarting the process
**Depends on**: Phase 2
**Requirements**: WATCH-01, WATCH-02, WATCH-03, WATCH-04, WATCH-05
**Success Criteria** (what must be TRUE):
  1. Saving any `.planning/` file causes the TUI to re-render updated state within 300ms
  2. Creating a new phase directory causes it to appear in the tree on the next file change — the new directory is watched automatically
  3. Rapid consecutive saves (e.g. during execute-phase) produce exactly one re-parse, not a render storm
  4. Only the changed file is re-parsed on a watch event; the rest of the tree is served from cache, keeping re-render latency imperceptible even with 50+ PLAN.md files
**Plans:** 3/3 plans complete

Plans:
- [ ] 03-01-PLAN.md — Watcher package: fsnotify, WalkDir, debounce, dynamic dir add
- [ ] 03-02-PLAN.md — ProjectCache: incremental re-parse keyed by path + mtime
- [ ] 03-03-PLAN.md — App wiring: channel, waitForEvent, FileChangedMsg handler + visual verification

### Phase 4: Plugin & Delivery
**Goal**: Users can install gsd-watch with two make commands and have the `/gsd-watch` slash command spawn a live sidebar in a tmux split pane automatically on both arm64 and amd64 Macs
**Depends on**: Phase 3
**Requirements**: PLUGIN-01, PLUGIN-02, PLUGIN-03, PLUGIN-04, PLUGIN-05, PLUGIN-06
**Success Criteria** (what must be TRUE):
  1. Running `/gsd-watch` inside a Claude Code session in tmux opens a 35%-width right-side split pane running gsd-watch
  2. Running `/gsd-watch` when not in tmux prints clear manual instructions rather than erroring or hanging
  3. Running `/gsd-watch` a second time does not spawn a duplicate pane
  4. `make install && make plugin-install` completes without error and the binary is available at `~/.local/bin/gsd-watch`
  5. The installed binary is under 15MB, passes `file` as a static Mach-O, and runs on both darwin/arm64 and darwin/amd64
**Plans:** 2/2 plans complete

Plans:
- [x] 04-01-PLAN.md — Makefile cross-compilation + main.go pane title
- [x] 04-02-PLAN.md — Slash command + end-to-end verification

### Phase 5: TUI Polish
**Goal**: Users see a polished TUI with clear visual hierarchy, graceful empty states, and enough discoverability that a new user understands the tool within 30 seconds
**Depends on**: Phase 4
**Requirements**: D-01, D-02, D-03, D-04, D-05, D-06, D-07, D-08, D-09, D-10
**Success Criteria** (what must be TRUE):
  1. When no `.planning/` directory exists, tree shows a centered "No GSD project found" message in gray
  2. Phases with no plans show "(no plans yet)" placeholder when expanded
  3. Completed phases render in dimmed gray for both phase and plan rows
  4. Footer shows a refresh icon that briefly flashes green on file change events
  5. Single q/Esc does not quit; double-q or double-Esc quits; Ctrl+C always quits immediately
  6. "?" opens a full-pane help overlay; single q/Esc dismisses it without quitting
  7. "e" expands all phases; "w" collapses all phases
  8. Footer displays two-line keybinding hints (navigation + actions)
  9. All TUI content has 1-character left/right padding
**Plans:** 3/3 plans complete

Plans:
- [x] 05-01-PLAN.md — Foundation + tree polish (keys, messages, styles, empty states, dimming, padding)
- [x] 05-02-PLAN.md — Footer redesign (refresh indicator, two-line hints)
- [x] 05-03-PLAN.md — App model wiring (double-quit, help overlay, expand/collapse-all, refresh flash routing)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Core TUI Scaffold | 4/4 | Complete   | 2026-03-19 |
| 2. Live Data Layer | 3/3 | Complete   | 2026-03-20 |
| 3. File Watching | 3/3 | Complete   | 2026-03-20 |
| 4. Plugin & Delivery | 2/2 | Complete   | 2026-03-21 |
| 5. TUI Polish | 3/3 | Complete   | 2026-03-21 |

### Phase 6: Onboarding, documentation, and UX improvements

**Goal:** [To be planned]
**Requirements**: TBD
**Depends on:** Phase 5
**Plans:** 0 plans

Plans:
- [ ] TBD (run /gsd:plan-phase 6 to break down)
