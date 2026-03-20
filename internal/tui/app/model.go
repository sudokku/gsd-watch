// Package app provides the root Bubble Tea model for gsd-watch, composing
// the tree, header, footer, and viewport sub-models into a working TUI.
package app

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ready        bool // set to true after first WindowSizeMsg
	cache        *parser.ProjectCache // incremental cache backed by .planning/
	events       chan tea.Msg          // watcher event channel
	planningRoot string               // path to .planning/ dir
}

// New returns a Model initialized with empty project data. Data arrives via
// ParsedMsg dispatched from Init(). The events channel is created in main()
// and passed here so the watcher goroutine and Bubble Tea runtime share it.
func New(events chan tea.Msg) Model {
	root, _ := os.Getwd()
	planningRoot := filepath.Join(root, ".planning")
	keys := tui.DefaultKeyMap()
	t := tree.New()
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
	}
}

// waitForEvent returns a tea.Cmd that blocks until the next message arrives on
// ch. It is returned from Init() and re-armed after every FileChangedMsg so
// the event loop perpetuates indefinitely.
func waitForEvent(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
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
	)
}

// Update implements tea.Model. Handles resize, quit, key delegation, ParsedMsg,
// and FileChangedMsg.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Quit handling takes priority.
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		// Delegate navigation keys to tree.
		var cmd tea.Cmd
		m.tree, cmd = m.tree.Update(msg)
		// Sync viewport content after tree state change.
		m.viewport.SetContent(m.tree.View(m.width))
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerH := m.header.Height()
		footerH := m.footer.Height()
		vpHeight := max(msg.Height-headerH-footerH, 0)
		m.viewport.Width = msg.Width
		m.viewport.Height = vpHeight
		m.viewport.SetContent(m.tree.View(m.width))
		m.ready = true
		return m, nil

	case tui.ParsedMsg:
		// Live data flows here from both ParseFull (startup) and cache.Update (incremental).
		m.tree = m.tree.SetData(msg.Project)
		m.header = m.header.SetData(msg.Project)
		m.footer = m.footer.SetData(msg.Project)
		m.viewport.SetContent(m.tree.View(m.width))
		return m, nil

	case tui.FileChangedMsg:
		path := msg.Path
		cache := m.cache
		return m, tea.Batch(
			func() tea.Msg {
				project := cache.Update(path)
				return tui.ParsedMsg{Project: project}
			},
			waitForEvent(m.events),
		)
	}

	// Let viewport handle its own messages (scroll, mouse, etc.).
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements tea.Model. Renders header, viewport (tree), and footer.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.width < tui.MinWidth {
		return "\u25c0 pane too narrow"
	}
	// Sync viewport content with current tree state.
	m.viewport.SetContent(m.tree.View(m.width))
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(m.width),
		m.viewport.View(),
		m.footer.View(m.width),
	)
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
