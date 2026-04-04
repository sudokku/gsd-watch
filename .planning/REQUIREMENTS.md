# Requirements: gsd-watch v1.4 cmux + Linux

**Defined:** 2026-04-04
**Core Value:** A developer running GSD can always see exactly where they are in their project — without context-switching out of Claude Code — and the view updates automatically within one second of any GSD action completing.

## v1.4 Requirements

### Build

- [ ] **BUILD-01**: User can cross-compile a Linux arm64 static binary via `make build-linux`
- [ ] **BUILD-02**: User can cross-compile a Linux amd64 static binary via `make build-linux`
- [ ] **BUILD-03**: User can build all four binaries (darwin-arm64, darwin-amd64, linux-arm64, linux-amd64) via `make build-all`
- [ ] **BUILD-04**: User can run `make install` on Linux and have the correct arch binary copied to the install path

### Multiplexer Detection

- [x] **MUXER-01**: User running gsd-watch inside cmux sees the TUI start normally (binary accepts `$CMUX_WORKSPACE_ID` as valid multiplexer)
- [x] **MUXER-02**: User running gsd-watch outside any multiplexer sees an error message that mentions both tmux and cmux
- [x] **MUXER-03**: Binary sets pane title via OSC 0 (works in both tmux and cmux) instead of OSC 2 (tmux-only)

### Slash Command: cmux Spawning

- [ ] **SPAWN-01**: User running `/gsd-watch` inside cmux proceeds past the multiplexer check without error
- [ ] **SPAWN-02**: User running `/gsd-watch` outside any multiplexer sees a clear error mentioning both tmux and cmux
- [ ] **SPAWN-03**: User running `/gsd-watch` inside cmux gets a right-side split pane created via the cmux socket API
- [ ] **SPAWN-04**: New cmux pane automatically starts gsd-watch in the correct project directory via socket send_text
- [ ] **SPAWN-05**: User running `/gsd-watch` inside tmux sees identical behavior to v1.3 (no regression)

### Documentation

- [ ] **DOCS-01**: README Requirements section shows a platform/multiplexer support matrix (macOS+Linux / tmux+cmux)
- [ ] **DOCS-02**: README Installation section includes Linux binary download instructions
- [ ] **DOCS-03**: README Building section documents `build-linux` and `build-all` make targets

## v2 Requirements

### cmux Duplicate Detection

- **DEDUP-01**: User cannot accidentally spawn two gsd-watch instances in the same cmux workspace (lockfile-based detection)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Windows support | macOS and Linux only; Windows has no tmux/cmux |
| Zellij / WezTerm / iTerm2 support | Only tmux and cmux supported in v1.x |
| cmux on Linux | cmux is macOS-only; Linux users always use tmux |
| Binary GitHub release publishing | Deferred to end of milestone (v1.4.0 tag) |
| Lockfile-based cmux duplicate detection | Deferred to v1.4.1 quick task |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BUILD-01 | Phase 17 | Pending |
| BUILD-02 | Phase 17 | Pending |
| BUILD-03 | Phase 17 | Pending |
| BUILD-04 | Phase 17 | Pending |
| MUXER-01 | Phase 18 | Complete |
| MUXER-02 | Phase 18 | Complete |
| MUXER-03 | Phase 18 | Complete |
| SPAWN-01 | Phase 19 | Pending |
| SPAWN-02 | Phase 19 | Pending |
| SPAWN-03 | Phase 20 | Pending |
| SPAWN-04 | Phase 20 | Pending |
| SPAWN-05 | Phase 20 | Pending |
| DOCS-01 | Phase 21 | Pending |
| DOCS-02 | Phase 21 | Pending |
| DOCS-03 | Phase 21 | Pending |

**Coverage:**
- v1.4 requirements: 15 total
- Mapped to phases: 15
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-04*
*Last updated: 2026-04-04 — roadmap finalized, all requirements assigned to phases 17-21*
