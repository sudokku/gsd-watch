---
phase: 12-archive-display
plan: "01"
subsystem: tui/tree
tags: [archive, rendering, tdd, tui, two-pass-view]
one_liner: "Pinned archive zone in TUI tree: FormatArchiveDate/RenderArchiveRow/RenderArchiveSeparator/RenderArchiveZone helpers + two-pass View(width, height) with height-aware scrollable/pinned split"
dependency_graph:
  requires:
    - "internal/parser/types.go (ArchivedMilestone struct from Phase 11)"
    - "internal/tui/styles.go (PendingStyle, ColorGray)"
  provides:
    - "FormatArchiveDate, RenderArchiveRow, RenderArchiveSeparator, RenderArchiveZone (exported, testable)"
    - "View(width, height int) — two-pass render with pinned archive zone"
  affects:
    - "internal/tui/app/model.go — all View calls updated to pass viewport.Height"
tech_stack:
  added:
    - "time (stdlib) — for time.Parse/time.Format in FormatArchiveDate"
    - "fmt (stdlib) — for fmt.Sprintf in RenderArchiveRow"
  patterns:
    - "TDD red/green cycle per plan spec"
    - "Two-pass render: scrollable zone + pinned zone"
    - "Exported helpers for testability from tree_test package"
    - "999 height sentinel for existing tests"
key_files:
  created: []
  modified:
    - "internal/tui/tree/view.go"
    - "internal/tui/tree/model_test.go"
    - "internal/tui/app/model.go"
decisions:
  - "FormatArchiveDate exported (not unexported) — tree_test package needs direct call"
  - "View(width, height) replaces View(width) — height required for pinned zone height math"
  - "999 height sentinel used in existing tests — preserves all pre-existing test behavior"
  - "app/model.go passes m.viewport.Height as height param — already available on Model"
  - "D-10 left-padding applied to archive zone lines separately — mirrors scrollable zone padding"
metrics:
  duration: "3 min"
  completed: "2026-03-26"
  tasks: 2
  files: 3
requirements:
  - ARC-02
---

# Phase 12 Plan 01: Archive Display TDD Summary

## What Was Built

Implemented the archive zone rendering helpers and updated the TUI tree View function to use a two-pass render with a pinned archive zone at the bottom.

### Archive Helpers (Task 1)

Four exported package-level functions in `internal/tui/tree/view.go`:

- `FormatArchiveDate(iso string) string` — converts ISO date to "Jan 2026" format; empty or invalid dates return ""
- `RenderArchiveRow(am ArchivedMilestone, noEmoji bool) string` — renders one archive row in emoji (▸/✓) or noEmoji (>/[done]) mode; styled with PendingStyle
- `RenderArchiveSeparator(width int) string` — builds "- - Archived Milestones - - -..." padded to width-1 (D-10 compensation); styled with PendingStyle
- `RenderArchiveZone(archives []ArchivedMilestone, width int, noEmoji bool) string` — returns "" when empty (D-04); separator + one row per archive joined by newline

### Two-Pass View (Task 2)

View signature changed from `View(width int)` to `View(width, height int)`. When archives exist:
1. Compute `pinnedH = len(ArchivedMilestones) + 1`
2. Compute `scrollH = height - pinnedH`
3. Cap padded scrollable lines to `scrollH`
4. Apply D-10 padding to archive zone lines
5. Join scrollable + archive zones with newline

When archives is empty: return scrollable content as-is (no height capping).

All 16 existing `m.View(80)` calls in model_test.go updated to `m.View(80, 999)`. All `m.tree.View(m.width)` calls in app/model.go updated to `m.tree.View(m.width, m.viewport.Height)`.

## Tests Added

8 archive helper unit tests + 5 two-pass View integration tests = 13 new tests. All 48 tree tests pass.

### New Tests

- TestFormatArchiveDate (table-driven, 5 cases)
- TestRenderArchiveRow_Emoji
- TestRenderArchiveRow_NoEmoji
- TestRenderArchiveRow_NoDate
- TestRenderArchiveSeparator
- TestRenderArchiveZone_Empty
- TestRenderArchiveZone_NonEmpty
- TestArchiveRowsNotInVisibleRows
- TestView_ArchiveZonePinned
- TestView_NoArchiveSectionWhenEmpty
- TestView_ArchiveRowFormat
- TestView_ArchiveRowFormatNoEmoji
- TestView_ScrollZoneHeightReduced

## Deviations from Plan

None — plan executed exactly as written. The app/model.go update (not in the plan's files_modified list) was required by Rule 3 (blocking issue: signature change breaks compilation). App model updated to pass `m.viewport.Height` as the height param.

## Known Stubs

None. All archive data flows from `ProjectData.ArchivedMilestones` (populated by Phase 11 parser work) through `RenderArchiveZone` and into `View` output.

## Commits

- `e87822b` — test(12-01): add failing tests for archive rendering helpers
- `2715fb6` — feat(12-01): implement archive rendering helpers
- `9c34b99` — test(12-01): add failing tests for two-pass View with pinned archive zone
- `6bc023d` — feat(12-01): implement two-pass View(width, height) with pinned archive zone

## Self-Check: PASSED
