package mock

import (
	"time"

	"github.com/radu/gsd-watch/internal/parser"
)

// MockProject returns a static ProjectData representing gsd-watch's own roadmap.
// This mock exercises all visual states: complete, in_progress, pending, failed,
// phase badges, the IsActive marker, and enough plans to require scrolling.
func MockProject() parser.ProjectData {
	return parser.ProjectData{
		Name:          "gsd-watch",
		ModelProfile:  "balanced",
		Mode:          "yolo",
		CurrentAction: "Phase 1 — building TUI scaffold",
		LastUpdated:   time.Now(),
		Phases: []parser.Phase{
			{
				DirName: "01-core-tui-scaffold",
				Name:    "Phase 1: Core TUI Scaffold",
				Status:  "in_progress",
				Badges:  []string{"discussed", "researched"},
				Plans: []parser.Plan{
					{Filename: "01-01-PLAN.md", Title: "Foundation: types, messages, mock data", Status: "complete"},
					{Filename: "01-02-PLAN.md", Title: "Tree model + viewport", Status: "in_progress", IsActive: true},
					{Filename: "01-03-PLAN.md", Title: "Header + footer components", Status: "pending"},
					{Filename: "01-04-PLAN.md", Title: "Root model + integration", Status: "pending"},
				},
			},
			{
				DirName: "02-live-data-layer",
				Name:    "Phase 2: Live Data Layer",
				Status:  "pending",
				Badges:  nil,
				Plans: []parser.Plan{
					{Filename: "02-01-PLAN.md", Title: "PLAN.md + ROADMAP.md parsers", Status: "pending"},
					{Filename: "02-02-PLAN.md", Title: "STATE.md + config.json parsers", Status: "pending"},
					{Filename: "02-03-PLAN.md", Title: "Wire parsers to TUI (blocked)", Status: "failed"},
				},
			},
			{
				DirName: "03-file-watching",
				Name:    "Phase 3: File Watching",
				Status:  "pending",
				Badges:  nil,
				Plans: []parser.Plan{
					{Filename: "03-01-PLAN.md", Title: "fsnotify watcher + debounce", Status: "pending"},
					{Filename: "03-02-PLAN.md", Title: "Incremental cache", Status: "pending"},
				},
			},
			{
				DirName: "04-plugin-delivery",
				Name:    "Phase 4: Plugin & Delivery",
				Status:  "pending",
				Badges:  nil,
				Plans: []parser.Plan{
					{Filename: "04-01-PLAN.md", Title: "Slash command + Makefile", Status: "pending"},
					{Filename: "04-02-PLAN.md", Title: "Static binary + install", Status: "pending"},
				},
			},
			{
				DirName: "05-tui-polish",
				Name:    "Phase 5: TUI Polish",
				Status:  "complete",
				Badges:  []string{"discussed", "researched", "planned", "executed", "verified", "uat"},
				Plans: []parser.Plan{
					{Filename: "05-01-PLAN.md", Title: "Tree + footer polish", Status: "complete"},
					{Filename: "05-02-PLAN.md", Title: "App model wiring", Status: "complete"},
				},
			},
			{
				DirName: "06-future",
				Name:    "Phase 6: Future Work",
				Status:  "pending",
				Badges:  nil,
				Plans:   nil,
			},
		},
	}
}
