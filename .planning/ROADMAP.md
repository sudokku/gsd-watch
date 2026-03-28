# Roadmap: gsd-watch

## Milestones

- ✅ **v1.0 gsd-watch MVP** — Phases 1-6 (shipped 2026-03-23)
- ✅ **v1.1 Parser Reliability + Observability + Quick Tasks** — Phases 7-10 (shipped 2026-03-25)
- ✅ **v1.2 Archived Milestone Visibility** — Phases 11-12 (shipped 2026-03-26)
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

<details>
<summary>✅ v1.1 Parser Reliability + Observability + Quick Tasks (Phases 7-10) — SHIPPED 2026-03-25</summary>

- [x] Phase 7: Parser Reliability + Fixture Corpus (2/2 plans) — completed 2026-03-23
- [x] Phase 8: Debug Mode (2/2 plans) — completed 2026-03-24
- [x] Phase 9: Quick Tasks TUI Section (2/2 plans) — completed 2026-03-24
- [x] Phase 10: Emoji/Text Toggle (2/2 plans) — completed 2026-03-25

Full phase details: `.planning/milestones/v1.1-ROADMAP.md`

</details>

<details>
<summary>✅ v1.2 Archived Milestone Visibility (Phases 11-12) — SHIPPED 2026-03-26</summary>

- [x] Phase 11: Archive Detection (2/2 plans) — completed 2026-03-25
- [x] Phase 12: Archive Display (2/2 plans) — completed 2026-03-26

Full phase details: `.planning/milestones/v1.2-ROADMAP.md`

</details>

### Phase 16: Custom Color Config

**Goal:** Users can override individual theme colors in config.toml under [theme], with the chosen preset as the base and per-field hex overrides applied on top
**Requirements**: TBD
**Depends on:** Phase 15
**Plans:** 2 plans

Plans:
- [x] 16-01-PLAN.md — Config schema: rename Theme to Preset, add ThemeColors struct with 5 *string fields, test fixtures and tests
- [x] 16-02-PLAN.md — ApplyColorOverrides + IsValidHex in styles.go, wire into app.New/main.go, rename all cfg.Theme to cfg.Preset

---

## Milestone v1.3: Settings

**Goal:** Users can configure appearance once via a config file, and discover it through the help overlay.

**Phases:** 13, 14, 15, 16 (continuing from v1.2 Phase 12)

- [x] **Phase 13: Config Infrastructure** — New `internal/config/` package; TOML loading with silent-defaults, fatal-error, and unknown-key-warning behaviors; `--no-emoji` and `--theme` flags override config via `flag.Visit` (completed 2026-03-26)
- [x] **Phase 14: Theme System** — `Theme` struct + three named presets in `styles.go`; call-site migration in `tree/view.go`; exported archive function signatures updated; theme name validated at startup (completed 2026-03-27)
- [x] **Phase 15: Help Overlay Config Hint** — `?` overlay shows config file path and active theme name (completed 2026-03-27)
- [x] **Phase 16: Custom Color Config** — Per-field hex color overrides in config.toml under [theme.colors], applied on top of the active preset (completed 2026-03-27)

## Phase Details

### Phase 13: Config Infrastructure
**Goal**: Users can start gsd-watch with or without a config file; the app reads `~/.config/gsd-watch/config.toml`, applies `emoji` and `theme` settings, and CLI flags always override config values
**Depends on**: Phase 12
**Requirements**: CFG-01, CFG-02, CFG-03, CFG-04, CFG-05
**Success Criteria** (what must be TRUE):
  1. When `~/.config/gsd-watch/config.toml` does not exist, gsd-watch starts normally with defaults — no error, no log noise, no crash
  2. When config.toml exists but is invalid TOML, gsd-watch exits with a fatal error message that includes the file path
  3. When config.toml contains unrecognised keys, gsd-watch prints a warning to stderr and starts normally with defaults for those keys
  4. When `--no-emoji` is passed on the command line, emoji is suppressed regardless of the `emoji` key value in config
  5. When `--theme <name>` is passed on the command line, that theme is used regardless of the `theme` key value in config
**Plans:** 2/2 plans complete
Plans:
- [x] 13-01-PLAN.md — Config package: Load(), Defaults(), UnknownKeysError, tests, testdata fixtures
- [x] 13-02-PLAN.md — Wire config into main.go (three-case dispatch, flag.Visit, --theme) and migrate app.New() signature

### Phase 14: Theme System
**Goal**: Users can select a named color theme (`default`, `minimal`, `high-contrast`) via config; all three presets render coherently in the tree view; an unknown theme name warns and falls back to default
**Depends on**: Phase 13
**Requirements**: THEME-01, THEME-02, THEME-03, THEME-04
**Success Criteria** (what must be TRUE):
  1. `theme = "default"` (or key omitted) produces no visual regression from gsd-watch v1.2
  2. `theme = "minimal"` renders muted status colors and a content-first appearance throughout the tree
  3. `theme = "high-contrast"` renders bold foreground colors using only 16-color ANSI palette indices — visible over SSH and in degraded terminals
  4. An unknown theme name (in config or via `--theme`) prints a stderr warning and falls back to `default` — gsd-watch does not crash
  5. Switching themes by editing config and restarting applies the new theme fully with no leftover colors from the previous theme
**Plans:** 2/2 plans complete
Plans:
- [x] 14-01-PLAN.md — Theme struct, three preset constructors, ResolveTheme, update tree.Options
- [x] 14-02-PLAN.md — Migrate view.go call sites, archive function signatures, main.go wiring + unknown-theme warning
**UI hint**: yes

### Phase 15: Help Overlay Config Hint
**Goal**: The `?` help overlay exposes the config file path and active theme name so users can discover and edit their settings without consulting docs
**Depends on**: Phase 14
**Requirements**: DISC-01, DISC-02
**Success Criteria** (what must be TRUE):
  1. Pressing `?` shows a line with the full config file path (e.g. `Config: ~/.config/gsd-watch/config.toml`) regardless of whether that file exists on disk
  2. Pressing `?` shows a line with the currently active theme name (e.g. `Theme: default`)
  3. Both lines are present whether the config file is absent, present-with-defaults, or present-with-explicit-values
**Plans:** 1/1 plans complete
Plans:
- [x] 15-01-PLAN.md — Extend helpView signature; render Config section with path + theme name; add DISC-01/DISC-02 tests
**UI hint**: yes

### v1.3 Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 13. Config Infrastructure | 2/2 | Complete    | 2026-03-26 |
| 14. Theme System | 2/2 | Complete    | 2026-03-27 |
| 15. Help Overlay Config Hint | 1/1 | Complete   | 2026-03-27 |
