// Package app provides the root Bubble Tea model for gsd-watch, composing
// the tree, header, footer, and viewport sub-models into a working TUI.
package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/radu/gsd-watch/internal/tui"
	"github.com/radu/gsd-watch/internal/tui/footer"
	"github.com/radu/gsd-watch/internal/tui/header"
	"github.com/radu/gsd-watch/internal/tui/mock"
	"github.com/radu/gsd-watch/internal/tui/tree"
)

// Model is the root Bubble Tea model that composes tree, header, footer, and viewport.
// All sub-models are stored as value types following the Elm pattern.
type Model struct {
	tree     tree.TreeModel
	header   header.HeaderModel
	footer   footer.FooterModel
	viewport viewport.Model
	keys     tui.KeyMap
	width    int
	height   int
	ready    bool // set to true after first WindowSizeMsg
}

// New returns a Model initialized with mock project data.
func New() Model {
	data := mock.MockProject()
	keys := tui.DefaultKeyMap()
	t := tree.New().SetData(data)
	h := header.New(data)
	f := footer.New(data, keys)
	vp := viewport.New(0, 0)
	return Model{
		tree:     t,
		header:   h,
		footer:   f,
		viewport: vp,
		keys:     keys,
		ready:    false,
	}
}

// Init implements tea.Model. Returns nil — no async work in Phase 1.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. Handles resize, quit, key delegation, and ParsedMsg.
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
		// Phase 2: live data comes through here.
		m.tree = m.tree.SetData(msg.Project)
		m.header = m.header.SetData(msg.Project)
		m.footer = m.footer.SetData(msg.Project)
		m.viewport.SetContent(m.tree.View(m.width))
		return m, nil
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
