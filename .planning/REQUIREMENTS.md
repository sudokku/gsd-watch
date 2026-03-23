# Requirements: gsd-watch

**Defined:** 2026-03-18
**Core Value:** A developer running GSD can always see exactly where they are in their project — without context-switching out of Claude Code — and the view updates automatically within one second of any GSD action completing.

> **v1.0 requirements** archived to `.planning/milestones/v1.0-REQUIREMENTS.md` (all 29 complete)

## v1.1 Requirements

### Parser Reliability

- [x] **PARSE-09**: Parser correctly sorts phases with no ROADMAP.md entry by extracting the phase number from the directory name (e.g. `07-foo` → phase 7), not returning 0
- [x] **PARSE-10**: Parser handles BOM (`\xEF\xBB\xBF`) and leading whitespace in PLAN.md frontmatter without treating the file as all-prose
- [x] **PARSE-11**: ROADMAP.md phase heading detection works for H2 (`##`), H3 (`###`), and H4 (`####`) heading formats — not only H3
- [ ] **PARSE-12**: App displays project name from PROJECT.md `# Title` when STATE.md `milestone_name` field is missing or empty

### Observability

- [ ] **OBS-01**: `--debug` flag prints parser decisions to stderr: phase dir detection, plan file matching, frontmatter parse results, badge detection, and cache hit/miss events

### Test Coverage

- [x] **TEST-01**: Test fixture corpus covers BOM-prefixed frontmatter, alternate heading formats (H2/H4), and phases missing from ROADMAP.md — all parsed correctly, with existing fixtures still passing

### Quick Tasks

- [ ] **QT-01**: User sees a collapsible "Quick tasks" section in the TUI tree showing tasks parsed from `.planning/quick/`
- [ ] **QT-02**: Quick task parser detects `YYMMDD-ID-PLAN.md` / `YYMMDD-ID-SUMMARY.md` naming convention; status determined by SUMMARY.md presence (complete) or absence (in-progress/pending)

### Accessibility

- [ ] **A11Y-01**: `--no-emoji` CLI flag switches all TUI status icons and badges to ASCII text equivalents (for SSH and minimal terminal environments)

## v1.2 Requirements

### Archived Milestone Visibility

- [ ] **ARC-01**: Parser detects archived milestone directories and extracts completion metadata (milestone name, phase count, completion date)
- [ ] **ARC-02**: User sees a collapsed, non-interactive row in the TUI tree for each completed archived milestone (e.g. "▸ v1.0 — 6 phases ✓"), displayed below the active milestone section

## v1.3 Requirements

### Config File

- [ ] **CFG-01**: App reads `~/.config/gsd-watch/config.toml` on startup; a missing or malformed file silently uses defaults and never errors
- [ ] **CFG-02**: Config `emoji = false` disables emoji in the TUI, superseding the `--no-emoji` flag if both are set
- [ ] **CFG-03**: Config `theme = "default"` key selects the active color theme

### Themes

- [ ] **THEME-01**: App ships named color presets: `default`, `minimal`, `high-contrast`
- [ ] **THEME-02**: `minimal` theme renders text-only status icons and muted/dimmed colors throughout
- [ ] **THEME-03**: `high-contrast` theme renders bold, high-contrast colors for maximum legibility

### Help Discoverability

- [ ] **HELP-01**: The `?` help overlay displays the config file path (e.g. `Config: ~/.config/gsd-watch/config.toml`) so users can discover and edit it

## v2 Requirements

### Unix Socket IPC

- **IPC-01**: Unix socket listener receives "refresh\n" signal for sub-100ms TUI update
- **IPC-02**: Stop and SubagentStop hooks signal TUI via gsd-watch-signal.sh on Claude Code response completion
- **IPC-03**: Stale socket file cleanup on startup (try-connect, delete if dead) for clean restarts after SIGKILL
- **IPC-04**: stop_hook_active guard in signal script to prevent infinite hook loops
- **IPC-05**: Socket path derived from project directory hash (consistent between Go binary and shell script)

### Enhancements

- **ENH-01**: GSD v2 support (`.gsd/` directory structure)
- **ENH-02**: Linux support (inotify-based fsnotify)

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
| In-TUI settings panel | Config file preferred for v1.3; panel deferred to future milestone |

## Traceability

**v1.0 Traceability:** archived to `.planning/milestones/v1.0-REQUIREMENTS.md` (29 requirements, all Complete)

**v1.1 Traceability:**

| Requirement | Phase | Status |
|-------------|-------|--------|
| PARSE-09 | Phase 7 | Complete |
| PARSE-10 | Phase 7 | Complete |
| PARSE-11 | Phase 7 | Complete |
| PARSE-12 | Phase 7 | Pending |
| TEST-01 | Phase 7 | Complete |
| OBS-01 | Phase 8 | Pending |
| QT-01 | Phase 9 | Pending |
| QT-02 | Phase 9 | Pending |
| A11Y-01 | Phase 10 | Pending |

**v1.2 Traceability:**

| Requirement | Phase | Status |
|-------------|-------|--------|
| ARC-01 | Phase 11 | Pending |
| ARC-02 | Phase 12 | Pending |

**v1.3 Traceability:**

| Requirement | Phase | Status |
|-------------|-------|--------|
| CFG-01 | Phase 13 | Pending |
| CFG-02 | Phase 13 | Pending |
| CFG-03 | Phase 13 | Pending |
| THEME-01 | Phase 14 | Pending |
| THEME-02 | Phase 14 | Pending |
| THEME-03 | Phase 14 | Pending |
| HELP-01 | Phase 15 | Pending |

**Coverage:**
- v1.0 requirements: 29 total, all complete (archived)
- v1.1 requirements: 9 total, 0 complete
- v1.2 requirements: 2 total, 0 complete
- v1.3 requirements: 7 total, 0 complete
- Unmapped: 0

---
*Requirements defined: 2026-03-18*
*Last updated: 2026-03-23 — v1.0 archived, v1.1/v1.2/v1.3 active*
