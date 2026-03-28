---
phase: 13-config-infrastructure
plan: 01
subsystem: config
tags: [toml, BurntSushi/toml, config, settings]

requires: []
provides:
  - "internal/config package with Load(), Defaults(), Config struct, UnknownKeysError, ConfigPath"
  - "CFG-01: silent defaults when config file is missing"
  - "CFG-02: error return on malformed TOML"
  - "CFG-03: UnknownKeysError return on unknown keys"
affects:
  - "13-02-wire-config: wires config package into main.go and app.New()"
  - "14-themes: consumes config.Config.Theme"

tech-stack:
  added: ["github.com/BurntSushi/toml v1.6.0"]
  patterns:
    - "Defaults() pre-initializes struct before DecodeFile to avoid Go zero-value pitfall for bool fields"
    - "errors.Is(err, fs.ErrNotExist) for missing file detection — toml.DecodeFile wraps OS error"
    - "md.Undecoded() for unknown key detection without exposing toml.Key in public API"
    - "UnknownKeysError.Keys as []string via k.String() conversion"

key-files:
  created:
    - "internal/config/load.go"
    - "internal/config/load_test.go"
    - "internal/config/testdata/valid.toml"
    - "internal/config/testdata/malformed.toml"
    - "internal/config/testdata/unknown-keys.toml"
    - "internal/config/testdata/empty.toml"
  modified:
    - "go.mod"
    - "go.sum"

key-decisions:
  - "Store unknown keys as []string (via k.String()) to avoid exposing toml.Key type in public API"
  - "Initialize cfg := Defaults() before DecodeFile so Emoji defaults to true (Go zero-value for bool is false)"
  - "Use errors.Is(err, fs.ErrNotExist) not == — toml.DecodeFile wraps the OS error"
  - "Single-file package (load.go) — no sub-splitting needed for 3 types + 2 functions"

patterns-established:
  - "TDD: test file + testdata committed first (RED), then implementation (GREEN), then full suite verification"
  - "Config package is isolated: internal/config/ has no dependency on any other internal package"

requirements-completed: [CFG-01, CFG-02, CFG-03]

duration: 2min
completed: 2026-03-26
---

# Phase 13 Plan 01: Config Infrastructure Summary

**TOML config package with Load()/Defaults()/UnknownKeysError using BurntSushi/toml, covering missing-file defaults (CFG-01), malformed-TOML errors (CFG-02), and unknown-key warnings (CFG-03)**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-03-26T11:40:44Z
- **Completed:** 2026-03-26T11:41:54Z
- **Tasks:** 1 (TDD — 2 commits: RED + GREEN)
- **Files modified:** 8

## Accomplishments

- `internal/config/` package created with full TDD coverage
- `Load()` handles all three config-file states (missing, malformed, unknown keys)
- `Defaults()` returns `Config{Emoji: true, Theme: ""}` as documented
- `ConfigPath` constant exports the XDG path for use in main.go and help overlay
- `github.com/BurntSushi/toml v1.6.0` added; full test suite remains green

## Task Commits

1. **Task 1 (RED): failing tests for config package** - `4b45d41` (test)
2. **Task 1 (GREEN): implement config package** - `4258443` (feat)

## Files Created/Modified

- `internal/config/load.go` - Config struct, Load(), Defaults(), UnknownKeysError, ConfigPath
- `internal/config/load_test.go` - TestLoad (5 subtests) + TestDefaults
- `internal/config/testdata/valid.toml` - emoji=false, theme="minimal"
- `internal/config/testdata/malformed.toml` - invalid TOML syntax
- `internal/config/testdata/unknown-keys.toml` - known keys + color="blue" unknown key
- `internal/config/testdata/empty.toml` - empty file (valid TOML, all defaults)
- `go.mod` - added github.com/BurntSushi/toml v1.6.0
- `go.sum` - updated

## Decisions Made

- `cfg := Defaults()` before `toml.DecodeFile` — ensures `Emoji` starts as `true` (not Go zero `false`)
- `errors.Is(err, fs.ErrNotExist)` — toml.DecodeFile wraps the OS error; direct `==` comparison would miss it
- `[]string` for `UnknownKeysError.Keys` — avoids leaking `toml.Key` type out of the config package boundary
- Single-file package — no need to split 3 types + 2 functions across files

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Config package ready for Plan 02 wiring into `main.go` and `app.New()`
- `Load()` returns typed errors; main.go can use type assertion for the warning path (CFG-03)
- `ConfigPath` constant ready for help overlay display in Phase 14
- No blockers

---
*Phase: 13-config-infrastructure*
*Completed: 2026-03-26*
