# Phase 11: Archive Detection - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-25
**Phase:** 11-archive-detection
**Areas discussed:** Archive discovery, Completion date source, Name field format

---

## Archive Discovery

| Option | Description | Selected |
|--------|-------------|----------|
| vX.Y-phases/ dir only | Only detect milestones that have a phases directory. v1.0 = found, v1.1 = not found. Clean and unambiguous. | ✓ |
| Any vX.Y-ROADMAP.md | Detect any milestone with an archived roadmap, phases dir or not. Phase count 0 when no phases dir. | |
| MILESTONES.md as source of truth | Parse MILESTONES.md for authoritative list, cross-reference with dirs. | |

**User's choice:** vX.Y-phases/ dir only
**Notes:** v1.1 was archived docs-only (no phases dir), which should not show in the TUI. Only fully archived milestones with a phases dir count.

---

## Completion Date Source

| Option | Description | Selected |
|--------|-------------|----------|
| MILESTONES.md | Parse .planning/MILESTONES.md for '## vX.Y ... (Shipped: YYYY-MM-DD)'. Reliable, structured format. | ✓ |
| vX.Y-ROADMAP.md header | Read archived roadmap and extract date — problem: only per-phase dates exist, no milestone-level date. | |
| Leave date always empty | Don't try to read a date — CompletionDate always blank. | |

**User's choice:** MILESTONES.md
**Notes:** Slight conceptual stretch (not literally inside the archive dir) but it's the authoritative record and already exists.

---

## Name Field Format

| Option | Description | Selected |
|--------|-------------|----------|
| Version only — "v1.0" | Derived from dir name. Matches Phase 12 display spec exactly. | ✓ |
| Full name — "v1.0 gsd-watch MVP" | Richer, from MILESTONES.md. Requires extra parsing. | |
| Two fields: Version + DisplayName | More flexible but adds struct complexity. | |

**User's choice:** Version only — "v1.0"
**Notes:** Phase 12 success criteria already shows `▸ v1.0 — 6 phases ✓` — version-only is the right fit.

---

## Claude's Discretion

- Exact regex patterns for version extraction and MILESTONES.md date parsing
- Sort order of returned slice
- Placement of `ArchivedMilestones` field in `ProjectData`

## Deferred Ideas

None.
