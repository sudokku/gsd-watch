---
name: gsd-watch
description: Open gsd-watch sidebar in a tmux split pane (35% width, right side)
disable-model-invocation: true
allowed-tools: Bash
---

Run the following bash script exactly as a single Bash tool call:

```bash
GSD_BIN=$(which gsd-watch 2>/dev/null)
if [ -z "$GSD_BIN" ]; then
  echo "gsd-watch not found. Install it first: clone the gsd-watch repo and run 'make all'."
  exit 0
fi

if [ -n "$TMUX" ]; then
  if tmux list-panes -s -F '#{pane_title}' | grep -q '^gsd-watch:'; then
    echo "gsd-watch is already running in this session. Use Ctrl+C in that pane to stop it first."
    exit 0
  fi
  PANE_ID=$(tmux split-window -h -p 35 -d -P -F '#{pane_id}')
  tmux send-keys -t "$PANE_ID" "cd \"$PWD\" && $GSD_BIN $ARGUMENTS" Enter
  echo "gsd-watch sidebar opened."
  exit 0
fi

if [ -n "$CMUX_WORKSPACE_ID" ]; then
  NEW_SURFACE=$(cmux new-split right | cut -d' ' -f2)
  cmux send --surface "$NEW_SURFACE" "cd \"$PWD\" && $GSD_BIN $ARGUMENTS\n"
  echo "gsd-watch sidebar opened."
  exit 0
fi

OS=$(uname -s)
if [ "$OS" = "Darwin" ]; then
  echo "gsd-watch requires tmux or cmux."
  echo "tmux:  brew install tmux"
  echo "       then: tmux new-session"
  echo "cmux:  open cmux — gsd-watch will work inside it automatically"
else
  echo "gsd-watch requires tmux or cmux."
  echo "tmux:  sudo apt install tmux"
  echo "       then: tmux new-session"
  echo "cmux:  open cmux — gsd-watch will work inside it automatically"
fi
```
