# Phase 13: Config Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-26
**Phase:** 13-config-infrastructure
**Areas discussed:** Config propagation, --theme behavior in Phase 13, TOML library

---

## Config propagation to `app.New()`

| Option | Description | Selected |
|--------|-------------|----------|
| Pass config.Config struct directly | `app.New(events, cfg config.Config)` — Phase 14 adds fields to config.Config, no signature change needed | ✓ |
| Introduce app.Options struct | `app.New(events, app.Options{...})` — decouples app from config package, extra translation layer in main.go | |
| Expand individual params | `app.New(events, noEmoji bool, theme string)` — simplest now, Phase 14/15 may need another signature change | |

**User's choice:** Pass config.Config struct directly
**Notes:** Clean forward-compatibility — Phase 14 just adds fields to the config struct.

---

## `--theme` behavior in Phase 13

| Option | Description | Selected |
|--------|-------------|----------|
| Store and pass through silently | Phase 13 stores whatever string --theme provides; Phase 14 owns validation | ✓ |
| Warn but continue | Phase 13 emits stderr warning for unknown theme names | |
| Only add --theme in Phase 14 | Skip --theme in Phase 13 entirely | |

**User's choice:** Store and pass through silently
**Notes:** Avoids duplicating validation logic that Phase 14 will own. CFG-05 is only about override precedence.

---

## TOML library

| Option | Description | Selected |
|--------|-------------|----------|
| BurntSushi/toml | `meta.Undecoded()` makes CFG-03 unknown-key warning trivial | ✓ |
| pelletier/go-toml v2 | Modern rewrite, strict-mode via DisallowUnknownFields | |

**User's choice:** BurntSushi/toml
**Notes:** `toml.MetaData.Undecoded()` provides unknown key list with zero extra logic.

---

## Claude's Discretion

- Internal structure of `internal/config/` (single file vs multiple)
- `config.Config` field names and TOML tag names
- Fatal error and unknown key warning message formats

## Deferred Ideas

None.
