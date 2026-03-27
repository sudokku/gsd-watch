// Package app provides the root Bubble Tea model for gsd-watch, composing
// the tree, header, footer, and viewport sub-models into a working TUI.
package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/config"
	"github.com/radu/gsd-watch/internal/parser"
	"github.com/radu/gsd-watch/internal/tui"
	"github.com/radu/gsd-watch/internal/tui/footer"
	"github.com/radu/gsd-watch/internal/tui/header"
	"github.com/radu/gsd-watch/internal/tui/tree"
	"github.com/radu/gsd-watch/internal/watcher"
)

// Model is the root Bubble Tea model that composes tree, header, footer, and viewport.
// All sub-models are stored as value types following the Elm pattern.
type Model struct {
	tree         tree.TreeModel
	header       header.HeaderModel
	footer       footer.FooterModel
	viewport     viewport.Model
	keys         tui.KeyMap
	width        int
	height       int
	ready        bool                 // set to true after first WindowSizeMsg
	helpVisible  bool                 // true when the help overlay is shown
	quitPending  bool                 // true after first q/Esc press (double-quit state machine)
	flashGen     int                  // incremented on each FileChangedMsg; used to discard stale RefreshFlashMsgs
	cache        *parser.ProjectCache // incremental cache backed by .planning/
	events       chan tea.Msg          // watcher event channel
	planningRoot string               // path to .planning/ dir
	cfg          config.Config        // user configuration (emoji, theme, etc.)
}

// New returns a Model initialized with empty project data. Data arrives via
// ParsedMsg dispatched from Init(). The events channel is created in main()
// and passed here so the watcher goroutine and Bubble Tea runtime share it.
// cfg holds user configuration (emoji, theme, etc.) loaded from the config file.
func New(events chan tea.Msg, cfg config.Config) Model {
	root, _ := os.Getwd()
	planningRoot := filepath.Join(root, ".planning")
	keys := tui.DefaultKeyMap()
	t := tree.New()
	t = t.SetOptions(tree.Options{NoEmoji: !cfg.Emoji})
	h := header.New(parser.ProjectData{})
	f := footer.New(parser.ProjectData{}, keys)
	vp := viewport.New(0, 0)
	return Model{
		tree:         t,
		header:       h,
		footer:       f,
		viewport:     vp,
		keys:         keys,
		ready:        false,
		events:       events,
		planningRoot: planningRoot,
		cache:        parser.NewCache(planningRoot),
		cfg:          cfg,
	}
}

// waitForEvent returns a tea.Cmd that blocks until the next message arrives on
// ch. It is returned from Init() and re-armed after every FileChangedMsg so
// the event loop perpetuates indefinitely.
func waitForEvent(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

// clockTickCmd returns a Cmd that fires ClockTickMsg at the next 1-second wall-clock
// boundary, keeping the footer timestamp visually up-to-date every second.
func clockTickCmd() tea.Cmd {
	return tea.Every(time.Second, func(time.Time) tea.Msg {
		return tui.ClockTickMsg{}
	})
}

// spinTickCmd returns a Cmd that fires SpinTickMsg after 80ms, advancing the
// braille spinner one frame. The spin loop self-terminates when activeChanges is false.
func spinTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
		return tui.SpinTickMsg{}
	})
}

// Init implements tea.Model. Starts the watcher goroutine and dispatches both
// an async full-parse cmd and a waitForEvent cmd so the loop is live from
// the first frame.
func (m Model) Init() tea.Cmd {
	// Start watcher goroutine (PROJECT.md decision: from Init(), not New()).
	go watcher.Run(m.planningRoot, m.events)

	cache := m.cache
	return tea.Batch(
		func() tea.Msg {
			project := cache.ParseFull()
			return tui.ParsedMsg{Project: project}
		},
		waitForEvent(m.events),
		clockTickCmd(),
	)
}

// Update implements tea.Model. Handles resize, quit, key delegation, ParsedMsg,
// FileChangedMsg, and RefreshFlashMsg.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ctrl+C always quits immediately, even during overlay (Pitfall 3).
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// Help overlay captures all keys except Ctrl+C (D-08).
		if m.helpVisible {
			if msg.String() == "q" || msg.Type == tea.KeyEscape {
				m.helpVisible = false
			}
			// All other keys ignored while overlay is open.
			return m, nil
		}

		// Double-quit with 1.5s confirm window (D-06).
		// First press: show confirmation prompt and start timeout.
		// Second press within window: quit. Timeout or any other key: reset.
		if msg.String() == "q" || msg.Type == tea.KeyEscape {
			if m.quitPending {
				return m, tea.Quit
			}
			m.quitPending = true
			m.footer = m.footer.SetQuitPending(true)
			return m, tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg {
				return tui.QuitTimeoutMsg{}
			})
		}
		// Any non-quit key resets quitPending immediately.
		if m.quitPending {
			m.quitPending = false
			m.footer = m.footer.SetQuitPending(false)
		}

		// Help key opens overlay (D-08).
		if key.Matches(msg, m.keys.Help) {
			m.helpVisible = true
			return m, nil
		}

		// Expand-all / collapse-all delegation (D-07).
		if key.Matches(msg, m.keys.ExpandAll) {
			m.tree = m.tree.ExpandAll()
			m.viewport.SetContent(m.tree.View(m.width, m.viewport.Height))
			return m, nil
		}
		if key.Matches(msg, m.keys.CollapseAll) {
			m.tree = m.tree.CollapseAll()
			m.viewport.SetContent(m.tree.View(m.width, m.viewport.Height))
			m.viewport.SetYOffset(0)
			return m, nil
		}

		// Delegate navigation keys to tree (existing behavior).
		var cmd tea.Cmd
		m.tree, cmd = m.tree.Update(msg)
		// Sync viewport content and scroll to keep cursor visible.
		m.viewport.SetContent(m.tree.View(m.width, m.viewport.Height))
		cursorLine := m.tree.RenderedCursorLine(m.width)
		if cursorLine < m.viewport.YOffset {
			m.viewport.SetYOffset(cursorLine)
		} else if cursorLine >= m.viewport.YOffset+m.viewport.Height {
			m.viewport.SetYOffset(cursorLine - m.viewport.Height + 1)
		}
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.footer = m.footer.SetWidth(msg.Width)
		headerH := m.header.Height()
		footerH := m.footer.Height()
		archiveH := m.tree.ArchiveZoneHeight()
		vpHeight := max(msg.Height-headerH-footerH-archiveH, 0)
		m.viewport.Width = msg.Width
		m.viewport.Height = vpHeight
		m.viewport.SetContent(m.tree.View(m.width, m.viewport.Height))
		m.ready = true
		return m, func() tea.Msg { return tea.ClearScreen() }

	case tui.ParsedMsg:
		// Live data flows here from both ParseFull (startup) and cache.Update (incremental).
		m.tree = m.tree.SetData(msg.Project)
		m.header = m.header.SetData(msg.Project)
		m.footer = m.footer.SetData(msg.Project)
		// Recalculate viewport height: footer height may change if the left label wraps,
		// and archive zone height may change if archived milestones were added/removed.
		if m.ready {
			archiveH := m.tree.ArchiveZoneHeight()
			vpHeight := max(m.height-m.header.Height()-m.footer.Height()-archiveH, 0)
			m.viewport.Height = vpHeight
		}
		m.viewport.SetContent(m.tree.View(m.width, m.viewport.Height))
		return m, nil

	case tui.FileChangedMsg:
		// Increment generation so any in-flight RefreshFlashMsg from a prior change is ignored.
		m.flashGen++
		gen := m.flashGen
		m.footer = m.footer.SetLastFile(filepath.Base(msg.Path))
		m.footer = m.footer.SetActiveChanges(true)
		path := msg.Path
		cache := m.cache
		return m, tea.Batch(
			func() tea.Msg {
				project := cache.Update(path)
				return tui.ParsedMsg{Project: project}
			},
			waitForEvent(m.events),
			spinTickCmd(),
			tea.Tick(3*time.Second, func(time.Time) tea.Msg {
				return tui.RefreshFlashMsg{Gen: gen}
			}),
		)

	case tui.RefreshFlashMsg:
		// Only clear active state if this tick belongs to the latest file-change burst.
		if msg.Gen == m.flashGen {
			m.footer = m.footer.SetActiveChanges(false)
		}
		return m, nil

	case tui.ClockTickMsg:
		// Re-arm the 1-second clock so the footer timestamp increments every second.
		return m, clockTickCmd()

	case tui.SpinTickMsg:
		// Advance spinner frame and re-arm only while changes are still active.
		if m.footer.ActiveChanges() {
			m.footer = m.footer.AdvanceSpinFrame()
			return m, spinTickCmd()
		}
		return m, nil

	case tui.QuitTimeoutMsg:
		// Confirm window expired — reset pending state if user didn't press again.
		if m.quitPending {
			m.quitPending = false
			m.footer = m.footer.SetQuitPending(false)
		}
		return m, nil
	}

	// Let viewport handle its own messages (scroll, mouse, etc.).
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// helpView renders the full-pane help overlay.
// When noEmoji is true, ASCII bracket codes replace emoji in the Phase stages section.
func helpView(width int, noEmoji bool) string {
	if width < tui.MinWidth {
		return "\u25c0 too narrow"
	}

	var phaseStages string
	if noEmoji {
		phaseStages = `Phase stages
[disc]   discussed
[rsrch]  researched
[ui]     ui spec
[plan]   planned
[exec]   executed
[vrfy]   verified
[uat]    UAT`
	} else {
		phaseStages = `Phase stages
💬  discussed
🔎  researched
🎨  ui spec
📋  planned
🚀  executed
✅  verified
🧪  UAT`
	}

	helpText := `gsd-watch help

Navigation
←/h  move left / collapse
↓/j  move down
↑/k  move up
→/l  move right / expand

Tree
e    expand all
w    collapse all
?    show this help

Quit
qq   quit gsd-watch
esc  quit gsd-watch

` + phaseStages + `

press q or esc to close`

	inner := lipgloss.NewStyle().
		Padding(1, 2).
		Foreground(tui.ColorGray)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorGray)

	content := box.Render(inner.Render(helpText))
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(content)
}

// View implements tea.Model. Renders header, viewport (tree), and footer.
// When helpVisible is true, renders the full-pane help overlay instead.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.width < tui.MinWidth {
		return "\u25c0 pane too narrow"
	}
	if m.helpVisible {
		return helpView(m.width, !m.cfg.Emoji)
	}
	// Sync viewport content with current tree state.
	m.viewport.SetContent(m.tree.View(m.width, m.viewport.Height))
	sections := []string{
		m.header.View(m.width),
		m.viewport.View(),
	}
	if az := m.tree.ArchiveZone(m.width); az != "" {
		sections = append(sections, az)
	}
	sections = append(sections, m.footer.View(m.width))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// ViewportHeight returns the current viewport height. Used in tests.
func (m Model) ViewportHeight() int {
	return m.viewport.Height
}

// Width returns the stored terminal width. Used in tests.
func (m Model) Width() int {
	return m.width
}

// TreeCursor returns the tree's current cursor position. Used in tests.
func (m Model) TreeCursor() int {
	return m.tree.Cursor()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
