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

### Phase 15: Help Overlay Config Hint
**Goal**: The `?` help overlay shows the config file path so users can discover and edit their settings
**Depends on**: Phase 14
**Requirements**: HELP-01
**Success Criteria** (what must be TRUE):
  1. The `?` overlay displays a line showing `Config: ~/.config/gsd-watch/config.toml`
  2. The path is shown regardless of whether the file currently exists on disk
  3. The config path line is visually distinct from keybinding rows (muted style or separator)
**Plans**: 2 plans

### v1.3 Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 13. Config File Infrastructure | 0/? | Not started | - |
| 14. Themes | 0/? | Not started | - |
| 15. Help Overlay Config Hint | 0/? | Not started | - |
