# Phase 15: Help Overlay Config Hint - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-27
**Phase:** 15-help-overlay-config-hint
**Areas discussed:** Section placement, Theme name display, helpView signature, Config path abbreviation

---

## Section Placement

| Option | Description | Selected |
|--------|-------------|----------|
| New section at bottom | Append Config section after Phase stages, before close hint | ✓ |
| Top of overlay | After "gsd-watch help" title, before Navigation | |
| After Quit section | Between Quit and Phase stages | |

**User's choice:** New section at bottom
**Notes:** Consistent with existing section pattern (Navigation / Tree / Quit / Phase stages / Config)

---

## Theme Name Display

| Option | Description | Selected |
|--------|-------------|----------|
| "default" (canonical) | Show resolved canonical name; "" sentinel → "default" | ✓ |
| Raw cfg.Theme value | Show empty string or blank if key absent | |

**User's choice:** Show "default" for empty string
**Notes:** Matches success criteria example `Theme: default`; caller normalises before passing to helpView

---

## helpView Signature

| Option | Description | Selected |
|--------|-------------|----------|
| Two string params | `helpView(width, noEmoji, configPath, themeName string)` — pure, testable | ✓ |
| cfg config.Config param | Pass full config struct; function resolves path/name internally | |
| Method on Model | `(m Model) helpView()` — accesses m.cfg directly | |

**User's choice:** Two string params
**Notes:** Consistent with Phase 10 noEmoji bool extension pattern; keeps helpView pure and testable

---

## Config Path Abbreviation

| Option | Description | Selected |
|--------|-------------|----------|
| Inline in View() | `strings.Replace(path, home, "~", 1)` at call site — no new helper | ✓ |
| tui.ConfigPath() helper | New exported function for tilde-abbreviated path | |

**User's choice:** Inline in View() call site
**Notes:** Single call site, no reuse needed

---

## Claude's Discretion

- Whether `config.DefaultPath()` already exists or needs to be added
- Column alignment of Config/Theme label-value pairs
- Test coverage approach for updated helpView signature

## Deferred Ideas

None
