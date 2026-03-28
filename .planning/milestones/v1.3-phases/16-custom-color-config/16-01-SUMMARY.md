---
phase: 16-custom-color-config
plan: "01"
subsystem: config
tags: [config, schema, toml, theme-colors]
dependency_graph:
  requires: [13-config-infrastructure, 14-theme-system]
  provides: [ThemeColors struct, Config.Preset field, [theme] TOML decode]
  affects: [internal/config/load.go, cmd/gsd-watch/main.go, internal/tui/app/model.go]
tech_stack:
  added: []
  patterns: [nested TOML struct with pointer fields, strPtr test helper]
key_files:
  created:
    - internal/config/testdata/theme-colors.toml
    - internal/config/testdata/theme-colors-invalid.toml
    - internal/config/testdata/old-theme-key.toml
    - internal/config/testdata/empty-theme-section.toml
  modified:
    - internal/config/load.go
    - internal/config/load_test.go
    - internal/config/testdata/valid.toml
    - internal/config/testdata/unknown-keys.toml
    - cmd/gsd-watch/main.go
    - internal/tui/app/model.go
decisions:
  - "[16-01] Config.Theme renamed to Config.Preset (toml:\"preset\"); Colors ThemeColors added with toml:\"theme\" tag for [theme] section decode"
  - "[16-01] ThemeColors uses *string pointer fields — nil means not set by user, enabling zero-value detection without sentinel strings"
  - "[16-01] old theme = 'string' produces TOML type mismatch error (not UnknownKeysError) because 'theme' maps to ThemeColors table field; test updated to reflect actual decoder behavior"
metrics:
  duration: 8
  completed_date: "2026-03-27"
  tasks_completed: 2
  files_modified: 8
---

# Phase 16 Plan 01: Config Schema Update Summary

Config.Theme renamed to Config.Preset and ThemeColors nested struct added with 5 *string fields decoded from [theme] TOML table via BurntSushi/toml.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Config schema update — rename Theme to Preset, add ThemeColors struct | d267484 | load.go, valid.toml, main.go, app/model.go |
| 2 | Update tests and add new test cases + fixtures | a6db611 | load_test.go, 4 new fixtures, unknown-keys.toml |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated cfg.Theme call sites in main.go and app/model.go**
- **Found during:** Task 1
- **Issue:** Renaming Config.Theme to Config.Preset would break compilation in cmd/gsd-watch/main.go (lines 84, 89-92) and internal/tui/app/model.go (lines 53, 351) which reference cfg.Theme directly.
- **Fix:** Updated all four call sites to cfg.Preset; build passes.
- **Files modified:** cmd/gsd-watch/main.go, internal/tui/app/model.go
- **Commit:** d267484

**2. [Rule 1 - Bug] old_theme_key test adjusted for actual TOML decoder behavior**
- **Found during:** Task 2
- **Issue:** Plan spec stated old `theme = "..."` would produce UnknownKeysError with key "theme". However, `theme` now maps to Config.Colors (ThemeColors struct), so TOML sees a string-vs-table type mismatch — a regular error, not UnknownKeysError.
- **Fix:** Updated test wantUnknown:false and wantConfig to Defaults() (decoder returns default on error). Updated unknown-keys.toml to use `preset = "default"` so the "color" key remains the only unknown key.
- **Files modified:** internal/config/load_test.go, internal/config/testdata/unknown-keys.toml
- **Commit:** a6db611

## Known Stubs

None — this plan modifies the config schema only; no UI or rendering changes.

## Self-Check: PASSED
