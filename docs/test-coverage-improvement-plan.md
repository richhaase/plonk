# Plonk Test Coverage Improvement Plan

This document outlines a detailed, incremental plan to substantially improve automated test coverage and test quality for Plonk, while ensuring tests never modify a developer’s real system.

## Objectives

- Raise Go unit/integration test coverage from ~50% to 80%+ over multiple phases
- Reduce regression risk with stable, hermetic tests that do not mutate the host system
- Establish uniform behavioral contracts across all package managers and commands
- Keep a small number of end-to-end BATS tests for smoke/UX while shifting behavior coverage to Go tests

## Constraints and Safety

- Do not invoke real package managers in tests (brew, npm, etc.)
- Do not write to the real home directory or global config
- Tests must be hermetic, using temporary HOME and PLONK_DIR
- All external process execution must be mocked in unit/integration tests

## Testing Strategy Overview

1. Environment isolation for all tests
   - Set `PLONK_DIR` to a temp directory
   - Set `HOME` to a temp directory for dotfile operations
   - Force non-terminal writer by default (`output.SetWriter(...)`) to stabilize spinner output
   - Replace default command executor with a mock (`packages.SetDefaultExecutor(...)`)

2. Shared Manager Contract Tests (Compliance Suite)
   - One reusable suite executed for every package manager to enforce common interface behavior
   - Validate: `IsAvailable`, `ListInstalled`, `Install`, `Uninstall`, `IsInstalled`, `Info`, `Search`, `Upgrade`, `Dependencies`
   - Explicitly assert behavior for unsupported features (e.g., return empty results for search when unsupported)

3. Orchestrator Apply Tests
   - Package-only, dotfiles-only, and combined paths
   - Dry-run vs. real (mocked) apply, success and failure aggregation, `ApplyResult.Success` logic

4. Go-based CLI Integration Tests
   - Programmatically invoke Cobra commands with isolated env + mock executor
   - Validate table/JSON/YAML outputs (unmarshal for JSON/YAML, golden files for table)
   - Commands: `install`, `uninstall`, `apply`, `status`, `diff`, `search`, `info`, `upgrade`, `config`, `dotfiles`

5. Property-Based Testing for Reconciliation
   - Fuzz/property tests for `ReconcileItems`/`ReconcileItemsWithKey`
   - Invariants: no duplicates, correct partitioning, merge semantics maintained

6. Timeout and Cancellation Tests
   - Simulate long-running/blocked manager operations that respect context deadline
   - Verify that command-level timeouts drive cancellation and return predictable errors

7. Lock Service Scaling + Edge Cases
   - Large lock files (e.g., 1k resources) for marshal/unmarshal + performance sanity
   - Version mismatch behavior already covered; keep and extend if migration is introduced

8. Output Formatting Golden Tests
   - Stable golden files for representative outputs of `apply`, `status`, `search`, `info`, `upgrade`
   - Normalize nondeterministic bits (timestamps) and sort lists for stability

9. Dotfiles Apply Scenarios
   - Add/update/unchanged/failed
   - Ignore patterns and expand_directories interactions
   - If backups are created on overwrite, verify naming and presence

## Enabling Infrastructure and Helpers

- Registry isolation (tests-only helper):
  - `WithTemporaryRegistry(t, register func(*packages.ManagerRegistry))` swaps out the global registry with a fresh instance for test and restores on cleanup

- CLI integration harness:
  - `RunCobra(t, args []string) (stdout string, err error)` sets temp `HOME`, temp `PLONK_DIR`, `NO_COLOR=1`, sets writer and executor mocks, then executes Cobra

- Output writer control:
  - Default tests run with a non-terminal writer to suppress spinner animation
  - Spinner-specific tests can opt-in to terminal mode via `testutil.NewBufferWriter(true)`

- Deterministic ordering:
  - Where result ordering is undefined, sort by `(Manager, Name)` prior to assertions

## Concrete Fixes and Cleanups

- internal/commands/apply.go: update the comment "Run apply with hooks..." to remove hook mention (documentation drift)
- internal/resources/packages/interfaces.go: remove unused install metadata types/interfaces (`InstallationInfo`, `InstallMethod`, `SecurityLevel`, `SelfInstaller.GetInstallationInfo`) unless they are to be used imminently (doctor/clone display)
- internal/config/constants.go: remove unused default timeout constants or re-plumb `defaultConfig` to reference them to avoid dual sources of truth

## Milestones and Coverage Targets

- Phase 1 (Foundations) — target +10–15% coverage
  - Add registry test helper, CLI harness, orchestrator apply tests (happy-path + errors)
  - Add initial CLI integration tests: install/uninstall/status/apply (mocked)
  - Fix comment drift and remove unused constants/API (if agreed)

- Phase 2 (Breadth) — target 75–80% total coverage
  - Manager compliance suite: apply to brew, npm, pipx, pnpm first, then remaining managers
  - Add timeout/cancellation tests; add golden tests for apply/status/search/info/upgrade
  - Property-based reconciliation tests

- Phase 3 (Depth + Stability)
  - Large lock file scenarios, dotfiles backup/overwrite tests, more CLI coverage (diff/config/dotfiles)
  - Introduce coverage gating in CI (start at 70%; raise to 80%)

## CI and Tooling

- Add `just test-coverage` (or equivalent):
  - `go test ./... -race -coverprofile=coverage.out`
  - `go tool cover -func=coverage.out`
- Enforce minimum coverage in CI with a threshold gate (incrementally increased)
- Keep BATS tests as smoke/system checks; main behavior validation lives in Go tests
- Run golangci-lint with `deadcode`, `unused`, `gosimple`, `staticcheck`, `stylecheck`

## Risks and Mitigation

- Risk: Over-mocking hides integration issues → keep a slim BATS/system test suite
- Risk: Brittle outputs → use JSON/YAML unmarshal and deterministic sorting
- Risk: Test flakiness due to timeouts/spinners → disable spinner animations for non-terminal writers; use short, controlled timeouts

---

## Tracking Grid

Use this grid to track the work, results, and learnings.

| ID | Workstream | Task | Owner | Status | Target Coverage Δ | PR/Issue | Notes/Learnings |
|----|------------|------|-------|--------|-------------------|----------|-----------------|
| T1 | Infra | Add `WithTemporaryRegistry` helper (tests-only) | | Planned | +2% | | Enables hermetic manager tests |
| T2 | Infra | Add `RunCobra` CLI test harness | | Planned | +3% | | Captures CLI output, sets env + mocks |
| T3 | Orchestrator | Apply tests (packages-only, dotfiles-only, combined; dry-run/real; errors) | | Planned | +4% | | Assert `ApplyResult.Success` and error aggregation |
| T4 | CLI | Integration tests: install/uninstall/status/apply (mock exec) | | Planned | +6% | | Validate JSON/YAML via unmarshal; golden for table |
| T5 | Managers | Compliance suite scaffold + run for brew/npm/pipx/pnpm | | Planned | +6% | | Enforces uniform contract behavior |
| T6 | Timeouts | Timeout/cancellation unit tests for install/uninstall/search | | Planned | +3% | | Use blocking mocks + deadlines |
| T7 | Reconcile | Property/fuzz tests for `ReconcileItems` (+WithKey) | | Planned | +2% | | Assert invariants and merges |
| T8 | Lock | Large lock file round-trip + atomic write sanity | | Planned | +2% | | Performance and stability check |
| T9 | Output | Golden tests for apply/status/search/info/upgrade | | Planned | +4% | | Sorted, timestamp-normalized outputs |
| T10 | Dotfiles | Apply scenarios: add/update/unchanged/failed; ignore/expand | | Planned | +3% | | Optional: backup/overwrite checks |
| F1 | Cleanup | Fix apply comment (remove hooks mention) | | Merged | 0% | ae68375 | Doc-only change in code |
| F2 | Cleanup | Remove unused install metadata API (or wire into doctor/clone) | | Merged | +1% | 9c6b400 | Shrinks surface, clarifies intent |
| F3 | Cleanup | Unify/remove duplicate default constants | | Merged | 0% | 390db0b | Single source of truth for defaults |

Legend for Status: Planned, In Progress, Merged, Deferred, Blocked.

---

## Acceptance Criteria

- Tests never call real package managers or touch the real home directory
- Coverage increases to ≥75% by end of Phase 2, ≥80% by end of Phase 3
- Manager compliance suite covers all supported managers (with explicit opt-outs where unsupported)
- CLI integration tests validate command behavior across table/JSON/YAML outputs
- CI enforces a minimum coverage threshold and runs linters

## Next Steps

1. Implement registry helper and CLI harness (T1, T2)
2. Add orchestrator apply tests and initial CLI tests (T3, T4)
3. Stand up manager compliance suite and port first four managers (T5)
4. Add timeout tests and golden outputs (T6, T9)
5. Expand to reconciliation fuzzing, lock scaling, dotfiles scenarios (T7, T8, T10)
