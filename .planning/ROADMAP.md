# Roadmap: gsd-watch

## Milestones

- ✅ **v1.0 gsd-watch MVP** — Phases 1-6 (shipped 2026-03-23)
- ✅ **v1.1 Parser Reliability + Observability + Quick Tasks** — Phases 7-10 (shipped 2026-03-25)
- ✅ **v1.2 Archived Milestone Visibility** — Phases 11-12 (shipped 2026-03-26)
- ✅ **v1.3 Settings** — Phases 13-16 (shipped 2026-03-27)
- 📋 **v1.4 cmux + Linux** — Phases 17-21 (planning)

## Phases

<details>
<summary>✅ v1.0 gsd-watch MVP (Phases 1-6) — SHIPPED 2026-03-23</summary>

- [x] Phase 1: Core TUI Scaffold (4/4 plans) — completed 2026-03-19
- [x] Phase 2: Live Data Layer (3/3 plans) — completed 2026-03-20
- [x] Phase 3: File Watching (3/3 plans) — completed 2026-03-20
- [x] Phase 4: Plugin & Delivery (2/2 plans) — completed 2026-03-21
- [x] Phase 5: TUI Polish (3/3 plans) — completed 2026-03-21
- [x] Phase 6: Onboarding, Documentation & UX (2/2 plans) — completed 2026-03-21

Full phase details: `.planning/milestones/v1.0-ROADMAP.md`

</details>

<details>
<summary>✅ v1.1 Parser Reliability + Observability + Quick Tasks (Phases 7-10) — SHIPPED 2026-03-25</summary>

- [x] Phase 7: Parser Reliability + Fixture Corpus (2/2 plans) — completed 2026-03-23
- [x] Phase 8: Debug Mode (2/2 plans) — completed 2026-03-24
- [x] Phase 9: Quick Tasks TUI Section (2/2 plans) — completed 2026-03-24
- [x] Phase 10: Emoji/Text Toggle (2/2 plans) — completed 2026-03-25

Full phase details: `.planning/milestones/v1.1-ROADMAP.md`

</details>

<details>
<summary>✅ v1.2 Archived Milestone Visibility (Phases 11-12) — SHIPPED 2026-03-26</summary>

- [x] Phase 11: Archive Detection (2/2 plans) — completed 2026-03-25
- [x] Phase 12: Archive Display (2/2 plans) — completed 2026-03-26

Full phase details: `.planning/milestones/v1.2-ROADMAP.md`

</details>

<details>
<summary>✅ v1.3 Settings (Phases 13-16) — SHIPPED 2026-03-27</summary>

- [x] Phase 13: Config Infrastructure (2/2 plans) — completed 2026-03-26
- [x] Phase 14: Theme System (2/2 plans) — completed 2026-03-27
- [x] Phase 15: Help Overlay Config Hint (1/1 plans) — completed 2026-03-27
- [x] Phase 16: Custom Color Config (2/2 plans) — completed 2026-03-27

Full phase details: `.planning/milestones/v1.3-ROADMAP.md`

</details>

### v1.4 cmux + Linux (Phases 17-21)

- [ ] **Phase 17: Linux Build Targets** — Extend Makefile with linux-arm64, linux-amd64, build-linux, build-all, and platform-aware install
- [ ] **Phase 18: Go Binary Multiplexer Detection** — Update main.go to accept $CMUX_WORKSPACE_ID, update error message, switch OSC 2 to OSC 0
- [ ] **Phase 19: Slash Command cmux Detection** — Update slash command to detect $CMUX_WORKSPACE_ID vs $TMUX and branch multiplexer check
- [ ] **Phase 20: Slash Command cmux Pane Spawning** — Add cmux pane creation via Unix socket API (nc -U, JSON-RPC)
- [ ] **Phase 21: Documentation** — Update README platform matrix, Linux install, make targets, and PROJECT.md Out of Scope

## Phase Details

### Phase 17: Linux Build Targets
**Goal**: Developers can cross-compile static Linux binaries for arm64 and amd64 from their Mac without Go source changes
**Depends on**: Nothing (parallelizable with Phase 18)
**Requirements**: BUILD-01, BUILD-02, BUILD-03, BUILD-04
**Success Criteria** (what must be TRUE):
  1. `make build-linux` produces two binaries: `build/gsd-watch-linux-arm64` and `build/gsd-watch-linux-amd64`
  2. `make build-all` produces all four binaries (darwin-arm64, darwin-amd64, linux-arm64, linux-amd64) in a single invocation
  3. Running `make install` on a Linux arm64 machine copies the arm64 binary; on a Linux amd64 machine copies the amd64 binary
  4. Linux binaries are static (CGO_ENABLED=0) and carry no codesign step
**Plans**: TBD

### Phase 18: Go Binary Multiplexer Detection
**Goal**: The compiled binary starts normally inside cmux and shows a helpful error outside any supported multiplexer
**Depends on**: Nothing (parallelizable with Phase 17)
**Requirements**: MUXER-01, MUXER-02, MUXER-03
**Success Criteria** (what must be TRUE):
  1. Running the binary inside a cmux workspace (CMUX_WORKSPACE_ID set) shows the TUI — not an error
  2. Running the binary outside both tmux and cmux shows an error message that names both tmux and cmux
  3. Pane title is set via OSC 0 (`\033]0;title\007`), confirmed to work in both tmux and cmux terminals
**Plans**: TBD

### Phase 19: Slash Command cmux Detection
**Goal**: The `/gsd-watch` slash command passes the multiplexer guard inside cmux and surfaces a clear error outside any multiplexer
**Depends on**: Phase 18
**Requirements**: SPAWN-01, SPAWN-02
**Success Criteria** (what must be TRUE):
  1. Running `/gsd-watch` inside cmux (CMUX_WORKSPACE_ID set) proceeds past the multiplexer check with no error
  2. Running `/gsd-watch` inside tmux proceeds past the multiplexer check identically to v1.3 (no regression)
  3. Running `/gsd-watch` outside both tmux and cmux shows an error message that names both tmux and cmux
**Plans**: TBD

### Phase 20: Slash Command cmux Pane Spawning
**Goal**: The `/gsd-watch` slash command opens a right-side pane in cmux and starts gsd-watch automatically, with no change to the tmux path
**Depends on**: Phase 19
**Requirements**: SPAWN-03, SPAWN-04, SPAWN-05
**Success Criteria** (what must be TRUE):
  1. Running `/gsd-watch` inside cmux creates a right-side split pane via `nc -U $CMUX_SOCKET_PATH` with the `surface.split` JSON-RPC call
  2. The new cmux pane automatically runs `gsd-watch` in the correct project directory via a `surface.send_text` JSON-RPC call
  3. Running `/gsd-watch` inside tmux produces the same split pane behavior as v1.3 — tmux code path is untouched
**Plans**: TBD

### Phase 21: Documentation
**Goal**: README and PROJECT.md accurately describe the expanded platform support so Linux users and cmux users can self-serve
**Depends on**: Phases 17, 18, 19, 20
**Requirements**: DOCS-01, DOCS-02, DOCS-03
**Success Criteria** (what must be TRUE):
  1. README contains a platform/multiplexer support matrix showing macOS+Linux and tmux+cmux combinations
  2. README Installation section includes Linux binary download and install instructions
  3. README Building section documents `make build-linux` and `make build-all` targets with descriptions
**Plans**: TBD

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Core TUI Scaffold | v1.0 | 4/4 | Complete | 2026-03-19 |
| 2. Live Data Layer | v1.0 | 3/3 | Complete | 2026-03-20 |
| 3. File Watching | v1.0 | 3/3 | Complete | 2026-03-20 |
| 4. Plugin & Delivery | v1.0 | 2/2 | Complete | 2026-03-21 |
| 5. TUI Polish | v1.0 | 3/3 | Complete | 2026-03-21 |
| 6. Onboarding & Docs | v1.0 | 2/2 | Complete | 2026-03-21 |
| 7. Parser Reliability | v1.1 | 2/2 | Complete | 2026-03-23 |
| 8. Debug Mode | v1.1 | 2/2 | Complete | 2026-03-24 |
| 9. Quick Tasks | v1.1 | 2/2 | Complete | 2026-03-24 |
| 10. Emoji/Text Toggle | v1.1 | 2/2 | Complete | 2026-03-25 |
| 11. Archive Detection | v1.2 | 2/2 | Complete | 2026-03-25 |
| 12. Archive Display | v1.2 | 2/2 | Complete | 2026-03-26 |
| 13. Config Infrastructure | v1.3 | 2/2 | Complete | 2026-03-26 |
| 14. Theme System | v1.3 | 2/2 | Complete | 2026-03-27 |
| 15. Help Overlay Config Hint | v1.3 | 1/1 | Complete | 2026-03-27 |
| 16. Custom Color Config | v1.3 | 2/2 | Complete | 2026-03-27 |
| 17. Linux Build Targets | v1.4 | 0/? | Not started | - |
| 18. Go Binary Multiplexer Detection | v1.4 | 0/? | Not started | - |
| 19. Slash Command cmux Detection | v1.4 | 0/? | Not started | - |
| 20. Slash Command cmux Pane Spawning | v1.4 | 0/? | Not started | - |
| 21. Documentation | v1.4 | 0/? | Not started | - |
