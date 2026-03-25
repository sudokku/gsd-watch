# Phase 11: Archive Detection - Context

**Gathered:** 2026-03-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Add `parseArchivedMilestones()` to the parser package — scan `.planning/milestones/` for archived milestone directories and return a slice of `ArchivedMilestone` structs. Wire the result into `ProjectData`. No TUI changes in this phase.

</domain>

<decisions>
## Implementation Decisions

### Archive Discovery
- **D-01:** An "archived milestone" is identified by the presence of a `vX.Y-phases/` directory inside `.planning/milestones/`. No phases dir = not an archived milestone. Partial archives (ROADMAP.md only, no phases dir) are silently ignored.
- **D-02:** Version is extracted from the dir name pattern `vX.Y-phases` → `"v1.0"`. Non-matching directory names are skipped.

### Completion Date
- **D-03:** Read completion date from `.planning/MILESTONES.md` by matching `## vX.Y ... (Shipped: YYYY-MM-DD)` for the corresponding version. If MILESTONES.md is absent, unreadable, or has no matching entry for a given version, leave `CompletionDate` empty — no crash.

### Struct Fields
- **D-04:** `ArchivedMilestone.Name` = version string only (e.g. `"v1.0"`), derived from the dir name. No separate display name or full milestone title.
- **D-05:** `PhaseCount` = count of subdirectories inside `vX.Y-phases/`. Non-directory entries are ignored.
- **D-06:** `CompletionDate` = string (ISO date from MILESTONES.md) or empty string. Not `time.Time` — avoids parse failures and keeps the struct simple.

### Error Handling
- **D-07:** Follows established best-effort pattern: `parseArchivedMilestones()` never returns an error. A malformed/missing archive dir is skipped with an optional `debugf()` call.
- **D-08:** No archived milestones returns an empty (not nil) slice to keep callers simple.

### Claude's Discretion
- Exact regex for version extraction from dir name (`vX.Y-phases`)
- Exact regex for parsing `(Shipped: YYYY-MM-DD)` from MILESTONES.md
- Sort order of returned `ArchivedMilestone` slice (newest-first by version seems natural)
- Where `ArchivedMilestones []ArchivedMilestone` is added to `ProjectData` struct

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Existing Parser Patterns
- `internal/parser/quick.go` — `parseQuickTasks()` is the direct structural analogue: ReadDir → regex match → build struct → sort. Follow the same pattern.
- `internal/parser/types.go` — All types live here; add `ArchivedMilestone` struct here.
- `internal/parser/parser.go` — `ParseProject()` is the wiring point; call `parseArchivedMilestones()` and assign to `ProjectData.ArchivedMilestones`.
- `internal/parser/debug.go` — `debugf()` for skip/detection events.

### Archive Format (live examples)
- `.planning/milestones/` — the directory to scan
- `.planning/milestones/v1.0-phases/` — example archived phases dir (6 subdirs)
- `.planning/MILESTONES.md` — contains `(Shipped: YYYY-MM-DD)` for each version

### Requirements
- `.planning/ROADMAP.md` Phase 11 success criteria — canonical acceptance test list

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `parseQuickTasks(quickDir string) []QuickTask` — identical structure: ReadDir loop, regex match on entry names, status/metadata extraction, sort. Copy the pattern directly.
- `debugf(event, format, args...)` — already used in quick.go for skip events; use same pattern here.
- `quickTaskDirRe` regex var — same declaration pattern for `archiveDirRe`.

### Established Patterns
- Best-effort: every parse function returns zero-value / empty slice on error, never propagates errors.
- `os.ReadDir()` over `filepath.Walk()` for single-level directory scanning (quick.go uses this).
- Types in `types.go`, parsing functions in dedicated files (e.g. `quick.go` → `archive.go`).
- Tests in `*_test.go` with testdata fixtures (`internal/parser/testdata/`).

### Integration Points
- `ProjectData` struct in `types.go` — add `ArchivedMilestones []ArchivedMilestone` field.
- `ParseProject()` in `parser.go` — call `parseArchivedMilestones(filepath.Join(root, "milestones"), filepath.Join(root, "MILESTONES.md"))` and assign to `data.ArchivedMilestones`.
- Cache layer (`cache.go`) — may need updating if `ArchivedMilestone` changes affect cache invalidation logic.

</code_context>

<specifics>
## Specific Ideas

- The `vX.Y-phases` dir name pattern is the single detection signal — if that dir exists, it's an archive. No secondary confirmation needed.
- MILESTONES.md lookup: match version string extracted from dir name against the heading, then capture the date in parens.
- Phase 12 will consume `ProjectData.ArchivedMilestones` and render `▸ v1.0 — 6 phases ✓`. The struct fields `Name`, `PhaseCount`, and `CompletionDate` are designed exactly for that display.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 11-archive-detection*
*Context gathered: 2026-03-25*
