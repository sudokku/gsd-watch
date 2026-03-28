# Phase 16: Custom Color Config - Discussion Log

**Date:** 2026-03-27
**Workflow:** discuss-phase

---

## Area: TOML Section Name

**Q: What should the TOML section for color overrides be called?**

Options presented:
- `[colors]` top-level — matches REQUIREMENTS.md future note, no conflict
- `[theme_colors]` — underscore-style
- Rename `theme` key → `preset`, use `[theme]` as section (breaking change)

**Selected:** Rename `theme` → `preset`, use `[theme]`

**User notes:** Asked about multiple custom profiles (e.g., `[profiles.dark]` and `[profiles.light]`) for poweruser switching. This was noted as a deferred idea — Phase 16 covers a single `[theme]` override block.

---

**Q: The rename breaks existing `theme = "..."` configs. How to handle it?**

Options:
- Clean break — just rename (unknown-key warning fires on old `theme` key)
- Accept both, warn on old (migration alias)

**Selected:** Clean break — just rename

---

## Area: Exposed Field Names

**Q: Which Theme fields should be overrideable in [theme]?**

Options:
- All status fields: complete, active, pending, failed, now_marker (5 fields)
- All 11 Theme struct fields

**Selected:** All status fields (5)

Fields: `complete`, `active`, `pending`, `failed`, `now_marker`

Transient/structural fields excluded: RefreshFlash, QuitPending, Highlight, EmptyFg, HelpBorder, HelpFg

---

## Area: Color Value Format

**Q: What color formats should be accepted?**

Options:
- Hex only — `#RRGGBB`
- Hex + ANSI indices (decimal string like `"2"`)

**Selected:** Hex only — `#RRGGBB`

---

## Area: Invalid Color Handling

**Q: What should happen when a color value is invalid?**

Options:
- Warning + ignore (use preset value, app starts)
- Fatal error

**Selected:** Warning + ignore — consistent with CFG-03 unknown-key behavior

---

*Discussion complete. CONTEXT.md written.*
