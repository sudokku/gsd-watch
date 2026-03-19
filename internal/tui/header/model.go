package header

import "github.com/radu/gsd-watch/internal/parser"

// HeaderModel is a stub — tests will fail (RED phase).
type HeaderModel struct{}

func New(data parser.ProjectData) HeaderModel { return HeaderModel{} }
func (h HeaderModel) SetData(data parser.ProjectData) HeaderModel { return h }
func (h HeaderModel) Height() int { return 0 }
func (h HeaderModel) View(width int) string { return "" }
