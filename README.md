# gsd-watch

A read-only terminal sidebar that shows your GSD project state live in a tmux split pane. It renders a collapsible phase/plan tree with status icons, phase lifecycle badges, and automatic file-watching updates. Keyboard-navigable. The pane sits alongside Claude Code so you always know where you are in your project without switching context.

## Demo

![gsd-watch demo](docs/demo.gif)

<!-- Placeholder — record with: ttyrec or vhs (https://github.com/charmbracelet/vhs) -->

Sidebar updates live as GSD phases progress.

## Dependencies

tmux is the only dependency:

```bash
brew install tmux
```

macOS only — builds for darwin/arm64 and darwin/amd64.

## Installation

```bash
git clone https://github.com/radu/gsd-watch.git
cd gsd-watch
make all
```

`make all` cross-compiles both architectures and copies the architecture-appropriate binary to `~/.local/bin/gsd-watch`. Make sure `~/.local/bin` is on your `$PATH`.

Then install the Claude Code slash command:

```bash
# Available in all projects (recommended):
make plugin-install-global

# Or available in current project only:
make plugin-install-local
```

`plugin-install-global` copies `commands/gsd-watch.md` to `~/.claude/commands/`. `plugin-install-local` copies it to `.claude/commands/` in your current project.

## Starting the sidebar

Inside a Claude Code session that is running in tmux, run:

```
/gsd-watch
```

What happens:

- **Inside tmux:** opens a 35%-width right-side split pane running `gsd-watch` from your current project directory. Focus stays on the original pane.
- **Not in tmux:** prints a message asking you to start a tmux session first.
- **Already running:** detects the existing pane by its title and tells you gsd-watch is already open — no duplicates.

## Usage

Switch between Claude Code and the gsd-watch pane using standard tmux pane navigation: `Ctrl+b` then an arrow key, or `Ctrl+b o` to cycle panes.

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `l` / `→` | Expand phase |
| `h` / `←` | Collapse phase |
| `e` | Expand all phases |
| `w` | Collapse all phases |
| `?` | Open help overlay |
| `qq` | Quit |
| `Esc Esc` | Quit |
| `Ctrl+C` | Quit immediately |

Press `?` at any time to open the help overlay, which shows keybindings and badge meanings inline.

### Phase lifecycle badges

| Badge | Meaning |
|-------|---------|
| 💬 | Discussed |
| 🔎 | Researched |
| 📋 | Planned |
| ✅ | Verified |
| 🧪 | UAT |

These appear under each phase in the tree when the corresponding lifecycle file (e.g. `01-CONTEXT.md`, `01-RESEARCH.md`) exists in the phase directory.

## Building

```bash
make build    # compiles build/gsd-watch-arm64 and build/gsd-watch-amd64
make install  # copies the arch-appropriate binary to ~/.local/bin/gsd-watch
make clean    # removes the build/ directory
```

Written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea). No CGO dependencies — static binary, zero runtime dependencies except tmux.

## Contributing

This project is actively maintained and very much open to contributions. If you run into a bug, have a feature idea, or want to propose an improvement — open a GitHub issue. I'm genuinely interested in hearing how other GSD users are using the tool and what would make it better for their workflow.

All input is welcome: bug reports, feature proposals, questions, or just a note that you're using it. Looking forward to it.

---

MIT License
