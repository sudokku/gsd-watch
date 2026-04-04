---
plan: 20-01
phase: 20-slash-command-cmux-pane-spawning
status: complete
completed: 2026-04-04
requirements:
  - SPAWN-03
  - SPAWN-04
  - SPAWN-05
---

## What Was Built

Replaced the cmux instructional stub in `commands/gsd-watch.md` Step 2 with real pane spawning using the cmux CLI. Running `/gsd-watch` inside a cmux workspace now automatically creates a right-side split pane with gsd-watch running in the correct project directory.

## Key Changes

- `commands/gsd-watch.md` — cmux branch in Step 2 replaced: stub text removed, real `cmux new-split right` + `cmux send` commands added
- Follow-up: all steps consolidated into a single bash script (one Bash tool call instead of 3) for faster execution
- `~/.claude/commands/gsd-watch.md` — global slash command updated via `make plugin-install-global`

## Self-Check: PASSED

- `grep "cmux new-split right" commands/gsd-watch.md` → match found
- `grep "cmux send --surface" commands/gsd-watch.md` → match found
- `grep "automatic pane spawning is not yet supported" commands/gsd-watch.md` → no match
- `grep "tmux split-window" commands/gsd-watch.md` → match found (Step 4 untouched)
- `go test ./...` → all 8 packages pass
- Human verified: cmux pane spawning works end-to-end (surface:5, surface:6 confirmed)

## Deviations

- Slash command consolidated into single bash script post-checkpoint (user request, improves latency)
- `make plugin-install-global` added to gsd-watch-builder agent and QC checklist (gap discovered during execution)

## key-files

created:
  - .planning/phases/20-slash-command-cmux-pane-spawning/20-01-SUMMARY.md
modified:
  - commands/gsd-watch.md
