# Roadmap: gsd-watch

## Milestones

- ✅ **v1.0 gsd-watch MVP** — Phases 1-6 (shipped 2026-03-23)
- 🚧 **v1.1 Parser Reliability + Observability + Quick Tasks** — Phases 7-10 (active)
- 📋 **v1.2 Archived Milestone Visibility** — Phases 11-12 (planned)
- 📋 **v1.3 Settings** — Phases 13-15 (planned)

## Phases

<details>
<summary>✅ v1.0 gsd-watch MVP (Phases 1-6) — SHIPPED 2026-03-23</summary>

- [x] Phase 1: Core TUI Scaffold (4/4 plans) — completed 2026-03-19
- [x] Phase 2: Live Data Layer (3/3 plans) — completed 2026-03-20
- [x] Phase 3: File Watching (3/3 plans) — completed 2026-03-20
- [x] Phase 4: Plugin & Delivery (2/2 plans) — completed 2026-03-21
- [x] Phase 5: TUI Polish (3/3 plans) — completed 2026-03-21
- [x] Phase 6: Onboarding, Documentation & UX (2/2 plans) — completed 2026-03-21

Full phase details: `.planning/milestones/v1.0-ROADMAP.md`

</details>

---

## Milestone v1.1: Parser Reliability + Observability + Quick Tasks

**Goal:** Make the parser verifiably correct against real-world naming variations, add observability for silent failures, and surface quick tasks in the TUI.

**Phases:** 7, 8, 9, 10 (continuing from v1.0 Phase 6)

### v1.1 Phases

- [ ] **Phase 7: Parser Reliability + Fixture Corpus** - Fix phase sorting, BOM frontmatter, heading regex, PROJECT.md name fallback; add test fixture corpus
- [ ] **Phase 8: Debug Mode** - `--debug` flag that prints parser decisions to stderr
- [ ] **Phase 9: Quick Tasks TUI Section** - Collapsible quick tasks tree section reading `.planning/quick/`
- [ ] **Phase 10: Emoji/Text Toggle** - `--no-emoji` flag for ASCII fallback in SSH and minimal terminals

### Phase 7: Parser Reliability + Fixture Corpus
**Goal**: The parser correctly handles every known real-world edge case — out-of-ROADMAP phase dirs, BOM-prefixed frontmatter, non-H3 phase headings, missing STATE.md milestone name — and a test fixture corpus confirms all cases pass
**Depends on**: Phase 6
**Requirements**: PARSE-09, PARSE-10, PARSE-11, PARSE-12, TEST-01
**Success Criteria** (what must be TRUE):
  1. A phase directory like `07-foo` that has no entry in ROADMAP.md sorts correctly in the tree at position 7, not position 0 or at the end
  2. A PLAN.md whose YAML frontmatter begins with a UTF-8 BOM or leading whitespace is parsed correctly — title, status, and all fields are extracted, not treated as prose
  3. ROADMAP.md phase headings formatted as `## Phase N`, `### Phase N`, or `#### Phase N` are all detected and matched to their phase dirs
  4. When STATE.md has no `milestone_name` field (or the field is blank), the header shows the project name from PROJECT.md's H1 title instead of "unknown"
  5. Running `go test ./...` passes with fixture corpus covering BOM frontmatter, H2/H4 headings, and phases absent from ROADMAP.md; all pre-existing fixtures continue to pass
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### Phase 8: Debug Mode
**Goal**: A developer running `gsd-watch --debug` can see every parser decision printed to stderr, making silent parse failures diagnosable without adding print statements to the source
**Depends on**: Phase 7
**Requirements**: OBS-01
**Success Criteria** (what must be TRUE):
  1. Running `gsd-watch --debug` prints parser events to stderr: phase dir detection, PLAN.md frontmatter parse results (field values or error), badge file detection, and cache hit/miss events
  2. Running `gsd-watch` without `--debug` produces no debug output — stderr is clean
  3. Debug output includes enough context to identify which file triggered each event (path is always present in the log line)
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### Phase 9: Quick Tasks TUI Section
**Goal**: Users see a collapsible "Quick tasks" section at the bottom of the TUI tree that shows tasks from `.planning/quick/`, with correct status determined by file naming convention
**Depends on**: Phase 8
**Requirements**: QT-01, QT-02
**Success Criteria** (what must be TRUE):
  1. A "Quick tasks" section appears in the TUI tree below the phase list; it can be expanded and collapsed with the same h/l or arrow keys used for phases
  2. Files matching `YYMMDD-ID-PLAN.md` in `.planning/quick/` appear as task rows; a task with a corresponding `YYMMDD-ID-SUMMARY.md` shows as complete, one without shows as in-progress or pending
  3. If `.planning/quick/` does not exist or is empty, the "Quick tasks" section shows an empty state placeholder rather than crashing or hiding entirely
  4. New files dropped into `.planning/quick/` appear in the TUI within 300ms via the existing fsnotify watcher (no new watcher config required)
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### Phase 10: Emoji/Text Toggle
**Goal**: Users who run gsd-watch in SSH sessions or terminals without emoji support can switch all status icons and lifecycle badges to ASCII equivalents with a single flag
**Depends on**: Phase 9
**Requirements**: A11Y-01
**Success Criteria** (what must be TRUE):
  1. Running `gsd-watch --no-emoji` replaces all emoji status icons (✓, ▶, ○, ✗) with ASCII equivalents (e.g. `[x]`, `[>]`, `[ ]`, `[!]`)
  2. Running `gsd-watch --no-emoji` replaces all lifecycle badge emoji (📝, 🔬, 📋, 🧪) with ASCII text equivalents (e.g. `[discussed]`, `[researched]`, `[verified]`, `[uat]`)
  3. Running `gsd-watch` without `--no-emoji` renders emoji unchanged — the flag is strictly opt-in
  4. The `--help` output lists `--no-emoji` with a description mentioning SSH and minimal terminal use
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### v1.1 Progress

**Execution Order:**
Phases execute in numeric order: 7 → 8 → 9 → 10

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 7. Parser Reliability + Fixture Corpus | 0/2 | Not started | - |
| 8. Debug Mode | 0/? | Not started | - |
| 9. Quick Tasks TUI Section | 0/? | Not started | - |
| 10. Emoji/Text Toggle | 0/? | Not started | - |

---

## Milestone v1.2: Archived Milestone Visibility

**Goal:** Completed milestones are acknowledged in the TUI without cluttering current work.

**Phases:** 11, 12 (continuing from v1.1 Phase 10)

### Phase 11: Archive Detection
**Goal**: The parser detects archived milestone directories and returns structured metadata (name, phase count, completion date) for each
**Depends on**: Phase 10
**Requirements**: ARC-01
**Success Criteria** (what must be TRUE):
  1. Parser returns one `ArchivedMilestone` struct per archived dir with name, phase count, and completion date populated
  2. Phase count derived from subdirectory count inside the archived dir
  3. Completion date read from metadata inside the archive; left empty (not crash) if absent
  4. Malformed or missing archive dir is skipped with optional `--debug` log, no crash
  5. No archive directory present returns empty list with no error
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### Phase 12: Archive Display
**Goal**: Users see a collapsed, non-interactive row per completed milestone below the active section in the TUI tree
**Depends on**: Phase 11
**Requirements**: ARC-02
**Success Criteria** (what must be TRUE):
  1. Each archived milestone renders as `▸ v1.0 — 6 phases ✓` below the active section and Quick Tasks section
  2. j/k and arrow keys skip archive rows entirely; h/l produces no expand/collapse action on them
  3. No archive section or placeholder appears when there are no archived milestones
  4. Archive rows render in a visually distinct dimmed/muted style
  5. New archive dir triggers a TUI update within 300ms via the existing fsnotify watcher
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### v1.2 Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 11. Archive Detection | 0/? | Not started | - |
| 12. Archive Display | 0/? | Not started | - |

---

## Milestone v1.3: Settings

**Goal:** Users can configure appearance once via a config file, and discover it through the help overlay.

**Phases:** 13, 14, 15 (continuing from v1.2 Phase 12)

### Phase 13: Config File Infrastructure
**Goal**: The app reads `~/.config/gsd-watch/config.toml` on startup and applies `emoji` and `theme` settings; a missing or malformed file silently uses defaults
**Depends on**: Phase 12
**Requirements**: CFG-01, CFG-02, CFG-03
**Success Criteria** (what must be TRUE):
  1. When `~/.config/gsd-watch/config.toml` does not exist, gsd-watch starts normally with defaults — no error, no crash
  2. When config contains `emoji = false`, emoji is suppressed exactly as if `--no-emoji` was passed
  3. When both `--no-emoji` flag and `emoji = true` in config are present, emoji is disabled (flag takes precedence)
  4. When config contains `theme = "default"` or the key is absent, the TUI renders with the existing default color scheme unchanged
  5. A config file with invalid TOML or unrecognised keys is skipped silently — gsd-watch starts with defaults
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### Phase 14: Themes
**Goal**: Users can select a named color theme via config; all three presets render coherently throughout the TUI using Lip Gloss
**Depends on**: Phase 13
**Requirements**: THEME-01, THEME-02, THEME-03
**Success Criteria** (what must be TRUE):
  1. `theme = "default"` (or omitted) produces no visual regression from pre-v1.3
  2. `theme = "minimal"` renders text-only status icons and muted/dimmed colors throughout
  3. `theme = "high-contrast"` renders bold, high-contrast foreground colors throughout
  4. All three themes use `lipgloss.AdaptiveColor` — no new external dependencies
  5. Switching themes by editing config and restarting applies the new theme fully
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### Phase 15: Help Overlay Config Hint
**Goal**: The `?` help overlay shows the config file path so users can discover and edit their settings
**Depends on**: Phase 14
**Requirements**: HELP-01
**Success Criteria** (what must be TRUE):
  1. The `?` overlay displays a line showing `Config: ~/.config/gsd-watch/config.toml`
  2. The path is shown regardless of whether the file currently exists on disk
  3. The config path line is visually distinct from keybinding rows (muted style or separator)
**Plans**: 2 plans
Plans:
- [ ] 07-01-PLAN.md — Fix BOM/whitespace, heading regex, phase sorting + unit tests
- [ ] 07-02-PLAN.md — PROJECT.md name fallback + integration regression

### v1.3 Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 13. Config File Infrastructure | 0/? | Not started | - |
| 14. Themes | 0/? | Not started | - |
| 15. Help Overlay Config Hint | 0/? | Not started | - |
