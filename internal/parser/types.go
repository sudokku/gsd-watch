package parser

import "time"

// Status constants for phases and plans.
const (
	StatusComplete   = "complete"
	StatusInProgress = "in_progress"
	StatusPending    = "pending"
	StatusFailed     = "failed"
)

// Badge constants for phase lifecycle badges (in lifecycle order).
const (
	BadgeDiscussed  = "discussed"
	BadgeResearched = "researched"
	BadgeUISpec     = "ui_spec"
	BadgePlanned    = "planned"
	BadgeExecuted   = "executed"
	BadgeVerified   = "verified"
	BadgeUAT        = "uat"
)

// Plan represents a single plan within a phase.
type Plan struct {
	Filename string
	Title    string
	Status   string
	IsActive bool
	Wave     int
}

// Phase represents a project phase containing multiple plans.
type Phase struct {
	DirName string
	Name    string
	Status  string
	Badges  []string
	Plans   []Plan
}

// QuickTask represents a quick task from .planning/quick/.
type QuickTask struct {
	DirName     string // e.g. "260323-re2-fix-gsd-watch-sidebar-closing-immediatel"
	DisplayName string // humanized: "fix gsd watch sidebar closing immediatel"
	Date        string // "260323" — YYMMDD for sort
	Status      string // StatusComplete / StatusInProgress / StatusPending
}

// ProjectData is the root data model for the entire project view.
type ProjectData struct {
	Name            string
	ModelProfile    string
	Mode            string
	Phases          []Phase
	QuickTasks      []QuickTask
	CurrentAction   string
	LastUpdated     time.Time
	ProgressPercent float64 // 0.0 to 1.0, from STATE.md progress.percent
}

// CompletionPercent returns the fraction of plans with status "complete"
// divided by total plans across all phases. Returns 0.0 if no plans exist.
func (p ProjectData) CompletionPercent() float64 {
	total := 0
	complete := 0
	for _, phase := range p.Phases {
		for _, plan := range phase.Plans {
			total++
			if plan.Status == StatusComplete {
				complete++
			}
		}
	}
	if total == 0 {
		return 0.0
	}
	return float64(complete) / float64(total)
}
