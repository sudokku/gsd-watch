---
name: gsd-watch
description: Open gsd-watch sidebar in a tmux split pane (35% width, right side)
disable-model-invocation: true
allowed-tools: Bash
---

Open a gsd-watch sidebar in a tmux pane. Follow these steps exactly using the Bash tool:

**Step 1 — Check if gsd-watch binary is available:**

Run: `which gsd-watch`

If the command exits with a non-zero exit code (binary not found), print exactly:

`gsd-watch not found. Install it first: clone the gsd-watch repo and run 'make all'.`

Then stop. Do not continue to step 2.

**Step 2 — Check if running inside tmux:**

Run: `echo $TMUX`

If the output is empty (not inside a tmux session), print exactly:

`gsd-watch requires tmux. Start a session first: 'tmux new-session', then run /gsd-watch again.`

Then stop. Do not continue to step 3.

**Step 3 — Check for duplicate instance:**

Run: `tmux list-panes -a -F '#{pane_current_command}'` to list the current command of all panes.

If any line from the output is exactly `gsd-watch`, print exactly:

`gsd-watch is already running in this session. Use Ctrl+C in that pane to stop it first.`

Then stop. Do not continue to step 4.

**Step 4 — Spawn gsd-watch in a new right-side pane:**

Run: `tmux split-window -h -p 35 -d "cd \"$PWD\" && gsd-watch"`

The `-h` flag creates a right-side vertical split. The `-p 35` sets the new pane to 35% of the current pane width. The `-d` flag keeps focus on the original pane. The `cd "$PWD"` ensures the binary runs from the correct project directory.

Then print: `gsd-watch sidebar opened.`
