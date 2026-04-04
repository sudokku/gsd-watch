---
name: gsd-watch
description: Open gsd-watch sidebar in a tmux split pane (35% width, right side)
disable-model-invocation: true
allowed-tools: Bash
---

Open a gsd-watch sidebar in a tmux pane. Follow these steps exactly using the Bash tool:

**Step 1 — Check if gsd-watch binary is available:**

Run: `GSD_BIN=$(which gsd-watch)`

If the command exits with a non-zero exit code (binary not found), print exactly:

`gsd-watch not found. Install it first: clone the gsd-watch repo and run 'make all'.`

Then stop. Do not continue to step 2.

**Step 2 — Check for a supported multiplexer:**

Run: `echo $CMUX_WORKSPACE_ID` to check for cmux. Then run: `echo $TMUX` to check for tmux.

If `$CMUX_WORKSPACE_ID` is non-empty (inside cmux), print exactly:

`cmux detected. Run \`gsd-watch\` in a cmux pane manually — automatic pane spawning is not yet supported.`

Then stop. Do not continue to step 3.

If `$TMUX` is non-empty (inside tmux), proceed to step 3. No output.

If both `$CMUX_WORKSPACE_ID` and `$TMUX` are empty (not inside any multiplexer), run `uname -s` to detect OS, then print exactly:

On macOS (uname output is `Darwin`):
```
gsd-watch requires tmux or cmux.
tmux:  brew install tmux
       then: tmux new-session
cmux:  open cmux — gsd-watch will work inside it automatically
```

On Linux (uname output is `Linux`):
```
gsd-watch requires tmux or cmux.
tmux:  sudo apt install tmux
       then: tmux new-session
cmux:  open cmux — gsd-watch will work inside it automatically
```

Then stop. Do not continue to step 3.

**Step 3 — Check for duplicate instance:**

Run: `tmux list-panes -s -F '#{pane_title}'` to list the title of all panes in the current session.

If any line from the output starts with `gsd-watch:`, print exactly:

`gsd-watch is already running in this session. Use Ctrl+C in that pane to stop it first.`

Then stop. Do not continue to step 4.

**Step 4 — Spawn gsd-watch in a new right-side pane:**

Run: `tmux split-window -h -p 35 -d "cd \"$PWD\" && $GSD_BIN"`

The `-h` flag creates a right-side vertical split. The `-p 35` sets the new pane to 35% of the current pane width. The `-d` flag keeps focus on the original pane. The `cd "$PWD"` ensures the binary runs from the correct project directory.

Then print: `gsd-watch sidebar opened.`
