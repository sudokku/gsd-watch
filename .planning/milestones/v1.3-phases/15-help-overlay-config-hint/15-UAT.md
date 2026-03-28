---
status: complete
phase: 15-help-overlay-config-hint
source: [15-01-SUMMARY.md]
started: 2026-03-27T18:10:00Z
updated: 2026-03-27T18:20:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Config path in help overlay
expected: Press `?` to open the help overlay. A "Config" section should appear (after "Phase stages") with a line: `Config:  ~/.config/gsd-watch/config.toml`
result: pass
note: Section header renamed to "Configurations" during UAT to avoid redundancy with "Config:" label

### 2. Theme name in help overlay
expected: With the overlay open (press `?`), a second line in the Config section should read `Theme:   default` (or the active theme name if you've set one in config.toml)
result: pass

### 3. Both lines present with no config file
expected: If `~/.config/gsd-watch/config.toml` does not exist on disk, pressing `?` still shows both the `Config:` path line and the `Theme:   default` line — nothing is missing or blank
result: pass
note: Config dir did not exist on disk — path shown correctly from compile-time constant (intended behavior per DISC-01)

## Summary

total: 3
passed: 3
issues: 0
pending: 0
skipped: 0

## Gaps

[none yet]
