---
phase: 17-linux-build-targets
plan: 01
subsystem: infra
tags: [makefile, cross-compilation, linux, go-build]

# Dependency graph
requires:
  - phase: 04-plugin-delivery
    provides: "Makefile with darwin build targets and codesign pattern — extended here with Linux variants"
provides:
  - "build-linux target producing static linux/arm64 and linux/amd64 binaries"
  - "build-all meta-target for all four platform binaries in one invocation"
  - "OS-agnostic install target with uname-based 4-branch platform detection"
  - "Missing-binary guard with human-readable error in install"
affects: [release, distribution, linux-users]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "CGO_ENABLED=0 GOOS=linux for pure-Go static Linux cross-compilation from macOS"
    - "4-branch uname OS+arch shell matrix with aarch64/x86_64 normalization before dispatch"
    - "No-build-dependency install: users explicitly call build target before install"

key-files:
  created: []
  modified:
    - "Makefile"

key-decisions:
  - "build target renamed to build-darwin (D-01); old build: removed cleanly — forces callers to be explicit about platform"
  - "build-linux produces linux/arm64 and linux/amd64 via GOOS=linux CGO_ENABLED=0; no codesign step (D-02, D-08)"
  - "build-all: build-darwin build-linux — prerequisite chain, not inline shell, enables make -j parallelism (D-03)"
  - "all: target removed entirely — was darwin-centric alias for install; removing avoids platform assumption (D-04)"
  - "install has no build prerequisite (D-05) — users run build target explicitly first"
  - "uname -s detects OS (Darwin vs Linux); uname -m detects arch; aarch64->arm64 and x86_64->amd64 normalized BEFORE 4-way dispatch (D-06)"
  - "Missing-binary guard prints error naming the missing binary and suggesting correct build target (D-07)"
  - "All shell logic in a single @ block with backslash continuations — avoids Make's per-line shell spawning pitfall"

patterns-established:
  - "Linux file targets: mkdir -p build/ first, then CGO_ENABLED=0 GOOS=linux GOARCH=<arch> go build"
  - "Multi-line shell in Make: single @ prefix, all lines joined with backslash continuations"

requirements-completed: [BUILD-01, BUILD-02, BUILD-03, BUILD-04]

# Metrics
duration: 2min
completed: 2026-04-04
---

# Phase 17: Linux Build Targets Summary

**Makefile rewritten with build-darwin/build-linux/build-all split, static Linux binaries via GOOS=linux CGO_ENABLED=0, and OS-agnostic 4-branch install with aarch64/x86_64 normalization and missing-binary guard**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-04-04T15:11:29Z
- **Completed:** 2026-04-04T15:12:51Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- `build-linux` cross-compiles `build/gsd-watch-linux-arm64` and `build/gsd-watch-linux-amd64` from macOS using pure-Go GOOS=linux — no external toolchain required
- `build-all` builds all four platform binaries in one invocation as a prerequisite chain (enables `make -j` parallelism)
- `install` detects host OS and arch via `uname -s`/`uname -m`, normalizes `aarch64`→`arm64` and `x86_64`→`amd64` before dispatch, prints actionable error if binary missing
- Old `build:` and `all:` targets removed — callers must now be explicit about platform

## Task Commits

Each task was committed atomically:

1. **Task 1: Rewrite Makefile with Linux targets and OS-agnostic install** - `db41ef0` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `Makefile` - Added BINARY_LINUX_ARM64/AMD64 vars, build-darwin/build-linux/build-all targets, Linux file targets without codesign, rewrote install to 4-branch OS+arch matrix

## Decisions Made
None beyond what was locked in CONTEXT.md (D-01 through D-10). Plan executed exactly as specified in the research and context documents.

## Deviations from Plan

None - plan executed exactly as written.

Note: `make build-all` triggers a pre-existing codesign ambiguity error on darwin targets (two matching "Apple Development" certificates in keychain) — this is unrelated to this plan's changes and was present before. The Linux-only path `make build-linux` works cleanly.

## Issues Encountered

Pre-existing codesign ambiguity on the darwin build target prevented verifying `make build-all` end-to-end, but:
- `make build-linux` succeeds and produces both Linux binaries (BUILD-01, BUILD-02)
- Makefile structure for `build-all` is correct (BUILD-03 — verified by code review)
- `install` shell logic reviewed against all four platform branches (BUILD-04)
- `go test ./...` passes with no regressions

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Linux users can now `git clone && make build-linux && make install` from source
- Developer can cross-compile distribution binaries for all four platforms from macOS via `make build-linux`
- `make build-all` will work once the codesign ambiguity is resolved (separate pre-existing issue)
- No blockers for next phase

---
*Phase: 17-linux-build-targets*
*Completed: 2026-04-04*
