---
phase: 13
slug: config-infrastructure
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-26
---

# Phase 13 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` |
| **Config file** | none — go test is convention-based |
| **Quick run command** | `go test ./internal/config/... -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/config/... -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 13-01-01 | 01 | 0 | CFG-01 | unit | `go test ./internal/config/... -run TestLoad/missing_file` | ❌ W0 | ⬜ pending |
| 13-01-02 | 01 | 0 | CFG-02 | unit | `go test ./internal/config/... -run TestLoad/malformed_toml` | ❌ W0 | ⬜ pending |
| 13-01-03 | 01 | 0 | CFG-03 | unit | `go test ./internal/config/... -run TestLoad/unknown_keys` | ❌ W0 | ⬜ pending |
| 13-01-04 | 01 | 1 | CFG-01,CFG-02,CFG-03 | unit | `go test ./internal/config/... -v` | ❌ W0 | ⬜ pending |
| 13-02-01 | 02 | 1 | CFG-04,CFG-05 | manual | run binary with `--no-emoji` + config `emoji = true`; run with `--theme minimal` + config `theme = "default"` | N/A | ⬜ pending |
| 13-02-02 | 02 | 1 | CFG-01..05 | integration | `go test ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/config/load.go` — Config struct, Load(), Defaults() (package doesn't exist yet)
- [ ] `internal/config/load_test.go` — table-driven tests for CFG-01/02/03
- [ ] `internal/config/testdata/valid.toml` — two-key TOML fixture (`emoji = true`, `theme = "default"`)
- [ ] `internal/config/testdata/malformed.toml` — invalid TOML syntax fixture
- [ ] `internal/config/testdata/unknown-keys.toml` — valid keys plus one unknown key (e.g. `color = "blue"`)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `--no-emoji` overrides `emoji = true` in config | CFG-04 | `flag.Visit` wiring in `main.go` — no `main_test.go` exists | Build binary; create config with `emoji = true`; run `./gsd-watch --no-emoji`; confirm ASCII icons |
| `--theme minimal` overrides `theme = "default"` in config | CFG-05 | same — flag wiring in main.go is boundary logic | Build binary; create config with `theme = "default"`; run `./gsd-watch --theme minimal`; confirm theme stored in cfg |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
