---
phase: 16
slug: custom-color-config
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-27
---

# Phase 16 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing go test infrastructure |
| **Quick run command** | `go test ./internal/config/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/config/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 3 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 16-01-01 | 01 | 0 | color-parse | unit | `go test ./internal/config/... -run TestHexColor -count=1` | ❌ W0 | ⬜ pending |
| 16-01-02 | 01 | 1 | ThemeColors struct | unit | `go test ./internal/config/... -run TestThemeColors -count=1` | ✅ | ⬜ pending |
| 16-01-03 | 01 | 1 | preset rename | unit | `go test ./internal/config/... -count=1` | ✅ | ⬜ pending |
| 16-01-04 | 01 | 2 | ApplyOverrides | unit | `go test ./internal/tui/... -run TestApplyColorOverrides -count=1` | ❌ W0 | ⬜ pending |
| 16-01-05 | 01 | 3 | integration | integration | `go test ./... -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/config/config_test.go` — add `TestHexColorValidation` stub (hex parse: valid, invalid, empty, 3-char short)
- [ ] `internal/tui/styles_test.go` — add `TestApplyColorOverrides` stub (nil fields pass through, non-nil fields override)

*Existing `go test` infrastructure covers all other requirements.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual color rendering | hex override display | lipgloss color rendering requires terminal | Run `./gsd-watch` with `[theme.colors] dir_name = "#ff0000"` — tree dir names should render red |
| Unknown-key warning for old `theme =` | CFG-03 compat | stderr output requires runtime | Add `theme = "minimal"` to config.toml, run tool, confirm "unknown key" warning on stderr |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 3s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
