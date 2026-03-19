package footer

import (
	"time"

	"github.com/radu/gsd-watch/internal/parser"
	tui "github.com/radu/gsd-watch/internal/tui"
)

// FooterModel is a stub — tests will fail (RED phase).
type FooterModel struct {
	currentAction string
	lastUpdated   time.Time
	keys          tui.KeyMap
}

func New(data parser.ProjectData, keys tui.KeyMap) FooterModel { return FooterModel{} }
func (f FooterModel) SetData(data parser.ProjectData) FooterModel { return f }
func (f FooterModel) Height() int { return 0 }
func (f FooterModel) View(width int) string { return "" }
