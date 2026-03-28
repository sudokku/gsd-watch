---
phase: 13-config-infrastructure
plan: 02
subsystem: tui/app, cmd/gsd-watch
tags: [config, toml, main, app-model, flags]

requires:
  - "13-01: internal/config package with Load(), Defaults(), Config struct, UnknownKeysError"
provides:
  - "CFG-01: silent defaults when config file is missing — wired in main.go"
  - "CFG-02: fatal error with path on malformed TOML — wired in main.go"
  - "CFG-03: stderr warning per unknown key, continues — wired in main.go"
  - "CFG-04: --no-emoji flag overrides config Emoji field via flag.Visit"
  - "CFG-05: --theme flag overrides config Theme field via flag.Visit"
  - "app.New() accepts config.Config instead of noEmoji bool"
affects:
  - "14-themes: app.Model.cfg.Theme ready for theme dispatch"

tech-stack:
  added: []
  patterns:
    - "flag.Visit() for CLI override of config-loaded values — only fires for explicitly-set flags"
    - "!cfg.Emoji inversion at call sites: Emoji=true means show emoji, NoEmoji=true means suppress"
    - "errors.As(err, &ukErr) for typed UnknownKeysError dispatch"
    - "_ = flag.Bool(...) to register flag for flag.Visit without needing the pointer value"

key-files:
  created: []
  modified:
    - "cmd/gsd-watch/main.go"
    - "internal/tui/app/model.go"
    - "internal/tui/model_test.go"

key-decisions:
  - "Use _ = flag.Bool('no-emoji', ...) — pointer not needed since flag.Visit checks by name; avoids 'declared and not used' compiler error"
  - "!cfg.Emoji inversion at call sites only — helpView keeps noEmoji bool param; inversion happens in View() and New()"
  - "flag.Visit override block placed after config.Load() — ensures config values are set before flags can override them"
  - "Test helpers updated to use config.Defaults() and cfg.Emoji=false — matches new API, test intent preserved"

requirements-completed: [CFG-01, CFG-02, CFG-03, CFG-04, CFG-05]

duration: 5min
completed: 2026-03-26
---

# Phase 13 Plan 02: Wire Config into main.go and app.New() Summary

**Config loading, three-case dispatch (silent/fatal/warn), flag.Visit overrides for --no-emoji and --theme, and app.Model migration from noEmoji bool to cfg config.Config**

## Performance

- **Duration:** ~5 min
- **Completed:** 2026-03-26
- **Tasks:** 2
- **Files modified:** 3 (cmd/gsd-watch/main.go, internal/tui/app/model.go, internal/tui/model_test.go)

## Accomplishments

- `main.go` loads config via `config.Load()` with three-case dispatch (CFG-01/02/03)
- `main.go` adds `--theme` flag alongside `--no-emoji` (CFG-05)
- `flag.Visit` override block applies CLI flags on top of config values (CFG-04, CFG-05)
- `app.New()` signature changed from `noEmoji bool` to `cfg config.Config`
- `Model.noEmoji bool` field removed, replaced by `cfg config.Config`
- `helpView` call site inverts: `!m.cfg.Emoji` (function signature kept stable)
- All tests updated and passing (22 tests across all packages)

## Task Commits

1. **Task 1: Wire config loading and --theme flag in main.go** - `30b7b64` (feat)
2. **Task 2: Migrate app.Model from noEmoji bool to cfg config.Config** - `15259d0` (feat)

## Files Created/Modified

- `cmd/gsd-watch/main.go` — config.Load() call, three-case dispatch, --theme flag, flag.Visit overrides, app.New(events, cfg)
- `internal/tui/app/model.go` — cfg config.Config field, New() signature update, !cfg.Emoji inversions
- `internal/tui/model_test.go` — newTestModel/newTestModelNoEmoji updated to use config.Config

## Decisions Made

- `_ = flag.Bool("no-emoji", ...)` — flag registration without capturing pointer; Go compiler requires all declared variables to be used; the flag still registers with the flag package for `flag.Visit` detection
- `!cfg.Emoji` inversion at call sites — `Config.Emoji=true` means "show emoji", but `tree.Options.NoEmoji=true` means "suppress emoji"; inversion done at the boundary
- `helpView(width, noEmoji bool)` signature kept stable — Phase 15 may adjust help overlay; keeping signature avoids churn

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed model_test.go call sites after app.New() signature change**
- **Found during:** Task 2 verification (`go test ./...`)
- **Issue:** `internal/tui/model_test.go` called `app.New(ch, false)` and `app.New(ch, true)` — no longer valid after signature change to `config.Config`
- **Fix:** Updated `newTestModel()` to use `config.Defaults()` and `newTestModelNoEmoji()` to use `cfg.Emoji = false`; added `config` import
- **Files modified:** `internal/tui/model_test.go`
- **Commit:** `15259d0`

## Issues Encountered

None beyond the expected test helper update.

## User Setup Required

None.

## Next Phase Readiness

- All CFG requirements (CFG-01 through CFG-05) fulfilled
- `app.Model.cfg.Theme` ready for Phase 14 theme dispatch
- `config.ConfigPath` constant available for help overlay path display
- No blockers

## Self-Check: PASSED

- `cmd/gsd-watch/main.go`: FOUND
- `internal/tui/app/model.go`: FOUND
- `13-02-SUMMARY.md`: FOUND
- Commit `30b7b64` (feat(13-02): wire config loading...): FOUND
- Commit `15259d0` (feat(13-02): migrate app.Model...): FOUND

---
*Phase: 13-config-infrastructure*
*Completed: 2026-03-26*
