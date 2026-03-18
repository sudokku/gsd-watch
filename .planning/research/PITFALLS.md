# Pitfalls Research

**Domain:** Go TUI binary — Bubble Tea + fsnotify + Unix socket IPC + YAML frontmatter + Claude Code plugin
**Researched:** 2026-03-18
**Confidence:** HIGH (core Bubble Tea/fsnotify/socket pitfalls from official sources and primary maintainer discussions; MEDIUM for Claude Code hook edge cases from official docs)

---

## Critical Pitfalls

### Pitfall 1: Mutating Model State Outside Update()

**What goes wrong:**
Any goroutine that writes directly to model fields (e.g., `m.content = value` inside a `go func()`) races against `View()`. The result is intermittent rendering of partial or stale state that's non-deterministic and extremely hard to reproduce in tests.

**Why it happens:**
Bubble Tea's event loop serializes all `Update()` calls — nothing else is serialized. Developers familiar with mutex-based Go concurrency assume they can write from any goroutine. They cannot.

**How to avoid:**
All state mutations happen exclusively in `Update()`. Background work (fsnotify event reading, socket accepts, file I/O) returns results as `tea.Msg` values via `tea.Cmd`. Never store a pointer to the model and write through it from a goroutine. Never use `sync.Mutex` or `sync.RWMutex` to "protect" model fields — the correct fix is to use commands.

**Warning signs:**
- Intermittent wrong content in the TUI that fixes itself on next keypress
- Data race reports from `go test -race`
- `View()` crashing on nil pointer dereference even though `Update()` always initializes fields

**Phase to address:** Core TUI scaffold (Phase 1) — establish the command/message pattern before any background I/O is introduced.

---

### Pitfall 2: Calling tea.Program.Send() Before Program.Run()

**What goes wrong:**
The Unix socket goroutine or fsnotify goroutine may fire events before `p.Run()` has started the event loop. `Program.Send()` will block (it uses a buffered channel that only drains once the loop is running), causing the goroutine to hang and the program to deadlock on startup.

**Why it happens:**
It is tempting to start the socket listener and file watcher immediately at the top of `main()` and pass the `*tea.Program` reference into them. If an event arrives in the sub-millisecond window before `p.Run()` begins draining the channel, the send blocks.

**How to avoid:**
Start background goroutines (socket listener, file watcher) only after `p.Run()` has returned its initial render, or — better — start them as `tea.Cmd` returned from `Init()`. Using `Init()` to return a batch of commands that spin up watchers guarantees they start inside the event loop lifecycle.

**Warning signs:**
- Program hangs at startup with no output
- Deadlock detected by Go runtime: `all goroutines are asleep`
- Works when there are no files to watch on startup

**Phase to address:** File watcher integration (Phase 2) and socket IPC (Phase 3) — start both subsystems from `Init()` commands, not from `main()`.

---

### Pitfall 3: fsnotify kqueue File Descriptor Exhaustion on macOS

**What goes wrong:**
kqueue (macOS's watch backend used by fsnotify v1.x) opens one file descriptor per watched file AND directory. A `.planning/` tree with 50+ PLAN.md files plus their parent directories can consume 100–200 fds. Under SIGKILL loops or rapid restarts this approaches macOS's default per-process fd limit (256 in some configurations), causing `too many open files` errors.

**Why it happens:**
fsnotify on macOS uses kqueue, which does not support recursive watching natively. The manual dir-walk approach required for this project (watch each subdirectory explicitly) compounds the fd usage because every directory entry also requires an fd.

**How to avoid:**
- Watch directories only, not individual files. fsnotify fires events for files inside a watched directory — there is no need to add individual file watches.
- On startup, call `watcher.Add(dir)` for each directory found during the recursive walk, but skip adding individual file paths.
- Set `ulimit -n 1024` in the Makefile's test target to surface exhaustion early.
- Track the number of watched directories; log a warning if it exceeds 200.
- Keep watches minimal: watch `.planning/` and its first-level subdirectories; the project structure does not go deeper than phase directories.

**Warning signs:**
- `watcher.Add()` returns an error containing "too many open files"
- Events stop arriving after the watcher has been running for a while
- `lsof -p <pid> | wc -l` grows unboundedly during testing

**Phase to address:** File watcher integration (Phase 2).

---

### Pitfall 4: Missing Recursive Watch for Newly Created Directories

**What goes wrong:**
`fsnotify.Watcher.Add()` is called once at startup for all existing directories. If a new phase directory (e.g., `.planning/phase-3/`) is created after startup, fsnotify never watches it and changes to files inside it are silently ignored.

**Why it happens:**
kqueue and inotify do not propagate watches to newly created subdirectories. The recursive walk at startup only covers what exists at that moment.

**How to avoid:**
In the fsnotify event handler, check for `fsnotify.Create` events where the target is a directory (use `os.Stat()` to confirm `IsDir()`). When a new directory is detected, immediately call `watcher.Add()` on it. This is the standard user-space recursive watcher pattern described in fsnotify's own issue tracker.

**Warning signs:**
- Adding a new phase via GSD does not trigger a TUI refresh
- Only the top-level `.planning/` directory and pre-existing phase dirs see updates

**Phase to address:** File watcher integration (Phase 2).

---

### Pitfall 5: Stale Unix Socket File Prevents Startup After SIGKILL

**What goes wrong:**
If the process is killed with SIGKILL (or crashes), the deferred `os.Remove()` on the socket path never runs. The next invocation of `gsd-watch` calls `net.Listen("unix", sockPath)` and gets `bind: address already in use`, causing an immediate fatal exit or a silent fallback that leaves the IPC mechanism broken.

**Why it happens:**
Unix socket files are filesystem artifacts. `net.UnixListener.Close()` removes the file on graceful shutdown, but SIGKILL bypasses all deferred cleanup. The socket file at `/tmp/gsd-watch-<hash>.sock` persists.

**How to avoid:**
On startup, before calling `net.Listen()`:
1. Attempt to connect to the existing socket path (`net.Dial("unix", sockPath)`).
2. If the connect succeeds, another instance is already running — exit with a clear error message.
3. If the connect fails (connection refused / no such file), call `os.Remove(sockPath)` unconditionally, then `net.Listen()`.

This is the "try-connect, delete-if-dead" pattern already identified in the project decisions. Implement it in startup, not as a defer.

**Warning signs:**
- `gsd-watch` exits immediately with "address already in use" after a crash
- Socket file exists in `/tmp/` but no `gsd-watch` process is running (`lsof /tmp/gsd-watch-*.sock` returns nothing)

**Phase to address:** Unix socket IPC (Phase 3).

---

### Pitfall 6: Unix Socket Goroutine Not Stopped on Program Quit

**What goes wrong:**
The socket accept loop runs in a goroutine started from `Init()`. When the user presses `q`, `tea.Quit` is sent and `p.Run()` returns — but the goroutine is still blocked on `listener.Accept()`. The process stays alive indefinitely (or until the OS reclaims it), and the socket file is never cleaned up.

**Why it happens:**
`p.Run()` returning does not cancel any commands or goroutines started outside the event loop. Background goroutines need explicit cancellation signals.

**How to avoid:**
Use a `context.Context` with `context.WithCancel`. Pass the cancel function into the socket goroutine. When `tea.Quit` is processed (or on `WindowFinalMsg` in Bubble Tea v1), call `cancel()`. In the goroutine, use `listener.SetDeadline()` or select on `ctx.Done()` to exit the accept loop. In the `Init()` command pattern, the context is created in `main()` and passed by closure.

**Warning signs:**
- `gsd-watch` process remains in `ps` after the TUI window closes
- Socket file remains in `/tmp/` after expected exit
- Port reuse test fails: launching a second instance immediately after quitting gets "address already in use"

**Phase to address:** Unix socket IPC (Phase 3).

---

### Pitfall 7: Debounce Timer Race with Go's time.Timer.Reset()

**What goes wrong:**
A naive debounce for fsnotify events looks like: reset a `time.Timer` every time an event arrives; fire the re-parse on expiry. If the timer fires concurrently while the goroutine is inside `timer.Reset()`, the drain-before-reset idiom is required but subtle. Pre-Go 1.23 code that doesn't drain `t.C` before `Reset()` drops or double-fires the debounce tick, causing either a missed refresh or two concurrent re-parses writing to shared state.

**Why it happens:**
`time.Timer.Reset()` is documented to have a race condition in Go versions < 1.23 unless the channel is drained first. The fix landed in Go 1.23 (unbuffered timer channel), but the project targets Go 1.22+, so the pre-1.23 behaviour must be handled.

**How to avoid:**
Use the safe drain pattern:
```go
if !timer.Stop() {
    select {
    case <-timer.C:
    default:
    }
}
timer.Reset(300 * time.Millisecond)
```
Or: target Go 1.23+ minimum and drop the drain entirely (timer channel is unbuffered in 1.23+). Document the minimum Go version in the Makefile.

Alternatively, implement debounce with a separate goroutine and a channel: send events to a channel, the goroutine waits for a quiet period using `time.After` and restarts it on each new event. This pattern sidesteps `Reset()` entirely.

**Warning signs:**
- `go test -race` reports a race on the timer channel
- Occasionally, two re-parse operations run simultaneously; last-write wins clobbers incremental cache

**Phase to address:** File watcher integration (Phase 2).

---

### Pitfall 8: YAML Frontmatter Delimiter Detection Fragility

**What goes wrong:**
`gopkg.in/yaml.v3` does not parse YAML frontmatter out of a mixed markdown+YAML file. You must split the `---` delimiters manually. Common mistakes:
- Splitting on `\n---\n` misses files where the opening `---` is on the very first line with no preceding newline (e.g., a file that starts with `---` at byte 0).
- Splitting on `---` (without newline anchoring) matches `---` inside a YAML value string.
- Files written on Windows may have `\r\n` line endings; `\n---\n` never matches.
- A file with only one `---` (opening delimiter present, closing missing) causes the second split part to contain the entire remaining file including markdown body — `yaml.Unmarshal` on that will either silently ignore fields or produce partial data with no error.

**How to avoid:**
Use `bytes.Index` on `\n---\n` and check for the special case where the file begins with `---\n` (offset 0). For the closing delimiter, search only after the opening. Enforce that both delimiters are found before attempting unmarshal; if either is missing, log a warning and return zero-value struct (never crash). Strip trailing `\r` before delimiter checks. Test with malformed files explicitly.

**Warning signs:**
- Plan files with `status: active` in frontmatter show as "upcoming" in the TUI
- No error logged but no frontmatter fields populated
- Works for all files in one phase dir, fails for files in another

**Phase to address:** File parsing (Phase 2 or wherever PLAN.md parsing is introduced).

---

### Pitfall 9: Rendering Panic When Terminal Width Is Below Minimum

**What goes wrong:**
Lip Gloss `Width()` constraints and padding calculations can produce negative values when the terminal pane is narrower than expected. `lipgloss.NewStyle().Width(-1)` does not error — it silently produces garbage output or panics in some Lip Gloss versions. In a tmux split pane that starts narrow, this crashes the TUI immediately on startup.

**Why it happens:**
Layout arithmetic assumes a minimum width. Developers calculate `paneWidth - padding - borderWidth` without clamping to zero. A tmux pane 20 columns wide with 4 columns of padding and 2-column border leaves 14 columns — but if the user resizes aggressively, the pane can go to 10 or even 5 columns.

**How to avoid:**
Always clamp computed width to a minimum (e.g., `max(computedWidth, 10)`). Establish a `minWidth` constant and render a "too narrow" placeholder message instead of the full tree when `msg.Width < minWidth`. Handle `tea.WindowSizeMsg` at every child model level to propagate current dimensions.

**Warning signs:**
- TUI crashes immediately when tmux split pane is very narrow
- Lip Gloss border characters wrap unexpectedly
- Width truncation breaks border rendering (known Lip Gloss v0.12.0 regression)

**Phase to address:** TUI rendering (Phase 1 or wherever View() is first built).

---

### Pitfall 10: Claude Code Stop Hook Causing Infinite Loop

**What goes wrong:**
The `Stop` hook fires every time Claude finishes a response. If the hook script always exits with a blocking decision, Claude enters forced continuation and the hook fires again — infinitely. The `gsd-watch-signal.sh` script is async and non-blocking (it just signals the socket), so infinite loops are not a risk by default — but any future change that adds a blocking decision must account for this.

**Why it happens:**
Claude Code passes `stop_hook_active: true` in the hook input when Claude is already in a forced-continuation state from a previous Stop block. Hooks that do not check this flag will block again, creating an infinite cycle.

**How to avoid:**
Even though `gsd-watch-signal.sh` is async, build in the guard explicitly:
```bash
STOP_ACTIVE=$(jq -r '.stop_hook_active // false' < /dev/stdin)
[ "$STOP_ACTIVE" = "true" ] && exit 0
```
Mark this comment in the script: "do not remove — prevents infinite Stop loop."

**Warning signs:**
- Claude Code session CPU spikes and never idles
- Hook script invoked thousands of times in a session
- Terminal floods with hook output in verbose mode

**Phase to address:** Claude Code plugin (Phase 4).

---

### Pitfall 11: Async Hook Output and Exit Codes Are Silently Ignored

**What goes wrong:**
Claude Code async hooks (`"async": true`) discard all stdout, stderr, exit codes, and JSON decisions. If `gsd-watch-signal.sh` fails to connect to the socket (because `gsd-watch` is not running), there is no error surface — Claude Code neither logs a warning nor retries. The TUI simply does not refresh.

**Why it happens:**
Async hooks run fire-and-forget. This is by design for non-blocking hook patterns. The documentation states: "response fields like decision and continue have no effect."

**How to avoid:**
This behaviour is acceptable for this project (the TUI refreshing on socket signal is a nice-to-have; missing one refresh is not a bug). However:
- The shell script must handle a missing socket gracefully (check for socket existence before `nc` or `socat`, exit 0 either way).
- Do not rely on async hook success for any correctness guarantee. File watcher is the primary refresh mechanism; socket signal is an accelerator only.

**Warning signs:**
- Assuming the hook worked because Claude Code showed no error (it never shows async errors)
- Debugging by checking hook exit codes (impossible with async hooks)

**Phase to address:** Claude Code plugin (Phase 4).

---

### Pitfall 12: tmux Detection Relies on $TMUX Being Set

**What goes wrong:**
`$TMUX` is set by tmux when a process is started inside a tmux session, but subprocesses launched by Claude Code (which itself runs in a shell) may not inherit `$TMUX` if the shell strips or resets the environment. The slash command logic that checks whether to spawn a tmux pane will misfire: either trying to create a split pane outside tmux (fails silently) or refusing to create one when tmux is actually available.

**Why it happens:**
Environment variable propagation in nested shells is not guaranteed. Claude Code spawns hooks via `sh -c`, which may start a non-interactive shell that does not source profile files. `$TMUX` may be absent even in a tmux-managed session.

**How to avoid:**
Use `$TMUX` as the primary check but fall back to `tmux list-sessions 2>/dev/null` as a secondary check. In the slash command documentation, make it explicit that the user must be in a tmux session and that the command is a no-op otherwise. Do not auto-detect and auto-create tmux sessions — the project already decided against this complexity.

**Warning signs:**
- `/gsd-watch` slash command does nothing when the user is in tmux
- `tmux split-window` returns an error about no current client

**Phase to address:** Claude Code plugin (Phase 4).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Full re-parse all PLAN.md files on every fsnotify event | Simpler code, no cache invalidation logic | Latency spike with 50+ files; potential for re-parse racing with ongoing write | Never — implement event-targeted cache from the start |
| Ignore fsnotify errors channel | Fewer code paths | Silent loss of watch events; watcher silently dies on fd exhaustion | Never — always drain the `watcher.Errors` channel in the event loop |
| Parse STATE.md with regex for all fields | Faster to implement | STATE.md is prose; regex breaks on any LLM rephrasing | Only for "best-effort" fields (current action text); never as source of truth for status |
| Use `os.Exit(1)` on any parse error | Simpler error handling | Crashes TUI on any malformed PLAN.md, destroying user's working session | Never — all file parsing must be fault-tolerant |
| Skip context cancellation for background goroutines | Fewer plumbing lines | Socket and watcher goroutines outlive the TUI; process zombie on quit | Never — context cancellation is mandatory for goroutine cleanup |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| fsnotify + new directories | Add watch only at startup, miss newly created phase dirs | Detect `fsnotify.Create` events on directories; call `watcher.Add()` dynamically |
| fsnotify + debounce | Reset `time.Timer` without draining on Go <1.23 | Use safe stop-drain-reset pattern or `time.After` channel pattern |
| Unix socket + startup | Call `net.Listen()` without removing stale socket file | Try-connect then remove-if-dead before listen |
| Unix socket + shutdown | Rely on `defer os.Remove()` to clean up | Use context cancellation + signal handler; defer is not called on SIGKILL |
| Claude Code hooks + async | Expect async hook to report errors or be retried | Design assuming async hook may silently fail; watcher is primary refresh path |
| Claude Code hooks + Stop | Missing `stop_hook_active` guard | Always check flag; add comment explaining the guard is mandatory |
| Lip Gloss + narrow panes | Compute `width - N` without clamping | Clamp all dimension arithmetic to a minimum; render placeholder when too narrow |
| `gopkg.in/yaml.v3` + frontmatter | Pass full file bytes to `yaml.Unmarshal` | Split on `---` delimiters first, handle missing/malformed delimiters gracefully |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Re-parsing all PLAN.md files on every event | TUI stutters on rapid GSD writes; goroutine backlog | Incremental cache: only re-parse the specific file path reported by fsnotify | With 50+ PLAN.md files during execute-phase (many writes/second) |
| Rendering full tree on every `tea.Msg` | Excessive terminal writes; flickering | Only re-render when model state actually changes; return `nil` cmd when no state change in Update() | During fsnotify event storms (many events within debounce window) |
| Spotlight indexing generating phantom events | Spurious re-parses; false "file changed" indicators | Debounce is sufficient mitigation; optionally filter events from paths containing `.spotlight-V100` | On any macOS system with Spotlight enabled |
| Walking entire `.planning/` on every fsnotify event to rebuild watcher | `O(n)` dir walk per event | Walk only on startup and when a new directory Create event is received | With large project trees |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| World-readable Unix socket at `/tmp/gsd-watch-<hash>.sock` | Any local process can send "refresh" signals | Socket is in `/tmp/` with default 0600 permissions; `net.Listen("unix", path)` uses umask — explicitly `chmod 0600` after creation |
| Parsing YAML frontmatter from arbitrary `.planning/` files without size limits | A maliciously crafted YAML with deeply nested aliases causes CPU/memory exhaustion (known `yaml.v3` vulnerability) | Check file size before parsing (reject files > 1MB); this project is personal-use so risk is low but the check costs nothing |
| Socket path derived from cwd without sanitization | Path traversal in socket name | Use `filepath.Abs(cwd)` then hash with `crc32` or `fnv` — avoid embedding raw path segments in socket name |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No "too narrow" message — TUI just corrupts | User sees garbage characters, assumes TUI is broken | Detect width < minWidth (e.g., 30 cols) and render single-line "pane too narrow" message |
| Panic in tea.Cmd leaves terminal in raw mode | User must manually run `reset` to recover terminal | Recover panics in all background goroutines; always call `p.Quit()` on unrecoverable error, never `os.Exit()` directly inside a goroutine |
| First render before file parse completes shows empty tree | Looks broken; user may assume no project is loaded | Show explicit "loading..." state in initial model; transition to tree once first parse completes |
| Missing/malformed files cause silent empty state | User sees empty tree but no indication of why | Log warnings to a debug log file; surface "N files could not be parsed" in footer |
| No visual feedback when socket signal received | User cannot tell if the "instant refresh" is working | Flash a subtle indicator (e.g., update "last-updated" timestamp) on every signal-triggered refresh |

---

## "Looks Done But Isn't" Checklist

- [ ] **fsnotify watcher:** Verify `watcher.Errors` channel is drained in the event loop — silently closing the watcher on error is invisible
- [ ] **Socket cleanup:** Verify the socket file is removed after `q` press, after `Ctrl+C`, AND after `kill -9` (the last one won't be cleaned; verify behavior is graceful, not a crash on next start)
- [ ] **Graceful file errors:** Verify TUI stays alive and shows partial data when any single PLAN.md is malformed YAML
- [ ] **Terminal reset:** Verify terminal state is restored after `q`, after `Ctrl+C`, and after panic in a background goroutine
- [ ] **Narrow pane:** Verify TUI renders without corruption in a 25-column tmux pane
- [ ] **High-frequency writes:** Verify debounce works correctly during a GSD execute-phase that writes multiple files rapidly
- [ ] **Stop hook guard:** Verify `stop_hook_active` check is present in the hook script before any blocking logic is ever added
- [ ] **New directory:** Verify a new phase directory created after TUI startup is automatically watched

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Stale socket prevents startup | LOW | `rm /tmp/gsd-watch-*.sock` — startup try-connect logic handles this automatically in next version |
| Terminal left in raw mode after panic | LOW | Run `reset` in shell; investigate panic cause with debug log |
| fd exhaustion in watcher | MEDIUM | Restart TUI; reduce watched directories by pruning deeply nested dirs; check `ulimit -n` |
| Goroutine leak (socket or watcher running after quit) | LOW | Kill process with `pkill gsd-watch`; fix: add context cancellation in Phase 3 |
| yaml.v3 alias bomb on crafted file | LOW | Delete or fix malformed file; add file-size guard in parser |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Model mutation outside Update() | Phase 1: Core TUI scaffold | `go test -race` passes; no mutex in model struct |
| Send() before Run() deadlock | Phase 1: Core TUI scaffold | All background goroutines start from `Init()` commands |
| Narrow pane rendering panic | Phase 1: Core TUI scaffold | TUI renders correctly in 25-column pane without panic |
| kqueue fd exhaustion | Phase 2: File watcher | `lsof -p <pid>` fd count stable under 200 after startup |
| Missing watch for new directories | Phase 2: File watcher | Creating a new dir in `.planning/` while TUI runs triggers refresh |
| Timer.Reset() race | Phase 2: File watcher | `go test -race` on debounce logic; or target Go 1.23+ |
| YAML frontmatter delimiter fragility | Phase 2: File watcher / parsing | Malformed PLAN.md test cases; zero crashes on parse errors |
| Stale socket on restart | Phase 3: Unix socket IPC | Kill -9 then re-launch succeeds; no "address already in use" |
| Socket goroutine not stopped | Phase 3: Unix socket IPC | Socket file removed after `q`; `ps` shows no zombie process |
| Stop hook infinite loop | Phase 4: Claude Code plugin | `stop_hook_active` guard present in script; manual test of repeated Stop |
| Async hook silent failure | Phase 4: Claude Code plugin | TUI still refreshes via watcher when hook is disabled; no crash |
| tmux detection failure | Phase 4: Claude Code plugin | `/gsd-watch` slash command works in tmux; graceful message outside tmux |

---

## Sources

- [Bubble Tea concurrency and goroutines (DeepWiki)](https://deepwiki.com/charmbracelet/bubbletea/5.1-concurrency-and-goroutines)
- [Tips for building Bubble Tea programs — leg100.github.io](https://leg100.github.io/en/posts/building-bubbletea-programs/)
- [Injecting messages from outside the program loop — bubbletea issue #25](https://github.com/charmbracelet/bubbletea/issues/25)
- [Race condition on repaint fix — bubbletea PR #330](https://github.com/charmbracelet/bubbletea/pull/330)
- [Teardown-related deadlock and race condition fixes — bubbletea PR #1373](https://github.com/charmbracelet/bubbletea/pull/1373)
- [fsnotify user-space recursive watcher discussion — issue #18](https://github.com/fsnotify/fsnotify/issues/18)
- [kqueue "too many open files" — notify-rs issue #596](https://github.com/notify-rs/notify/issues/596)
- [Spotlight indexing extra events — fsnotify issue #15](https://github.com/fsnotify/fsnotify/issues/15)
- [Go timer reset race condition — blogtitle.github.io](https://blogtitle.github.io/go-advanced-concurrency-patterns-part-2-timers/)
- [Go 1.23 timer reset fix — antonz.org](https://antonz.org/timer-reset/)
- [Unix domain socket file not removed on exit — golang issue #70985](https://github.com/golang/go/issues/70985)
- [net.UnixListener.Close() removes socket file — Golang-nuts discussion](https://groups.google.com/g/Golang-nuts/c/UtBR4IfgaEw)
- [Claude Code hooks reference — official docs](https://code.claude.com/docs/en/hooks)
- [Claude Code async hooks explainer — reading.sh](https://reading.sh/claude-code-async-hooks-what-they-are-and-when-to-use-them-61b21cd71aad)
- [Width truncation broke lipgloss border rendering — charmbracelet/x issue #123](https://github.com/charmbracelet/x/issues/123)
- [gopkg.in/yaml.v3 — official package docs](https://pkg.go.dev/gopkg.in/yaml.v3)

---
*Pitfalls research for: Go Bubble Tea TUI with fsnotify, Unix socket IPC, YAML frontmatter, Claude Code plugin*
*Researched: 2026-03-18*
