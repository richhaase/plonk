# Testability Refactor: Phase 1

Status: Completed (2025-09-25)

This plan delivers the lowest‑risk, highest‑impact refactors to improve testability and determinism without changing CLI behavior or UX. It focuses on widening key seams, removing hidden dependencies, and stabilizing ordering so unit tests can be written hermetically and predictably.

## Goals

- Inject dependencies (lock, registry, config) to enable fast, hermetic unit tests.
- Remove hidden, hardcoded timeouts in inner layers; let callers control timeouts.
- Make upgrade execution order deterministic to eliminate flaky tests.
- Preserve all user‑visible behavior and CLI output.

## Scope (Phase 1)

1) Widen hot‑path helpers to accept interfaces, not concretes.
2) Add explicit‑dependency entrypoints for install/uninstall flows.
3) Eliminate hardcoded 5‑minute package operation timeouts in inner layer.
4) Make upgrade ordering deterministic and inject registry dependency.
5) Add focused unit tests to lock in behavior and ordering.

Non‑goals (Phase 1):
- No new features, flags, or CLI output changes.
- No cross‑cutting orchestrator redesign.
- No changes to search behavior (already timeout‑configurable).

## Detailed Changes

### A. Use interfaces in package operations (install/uninstall)

- Change signatures to accept `lock.LockService` instead of `*lock.YAMLLockService`:
  - `installSinglePackage(ctx context.Context, configDir string, lockSvc lock.LockService, packageName, manager string, dryRun bool)`
  - `uninstallSinglePackage(ctx context.Context, configDir string, lockSvc lock.LockService, packageName, manager string, dryRun bool)`

Rationale:
- Tests can pass a trivial in‑memory/mock lock implementation to exercise branches (exists/not exists, write error, etc.) without touching disk.

Compatibility:
- All call sites continue to work; the concrete `YAMLLockService` satisfies the interface.

### B. Add explicit‑dependency entrypoints and delegate

- New functions in `internal/resources/packages/operations.go`:
  - `InstallPackagesWith(ctx, cfg *config.Config, lockSvc lock.LockService, registry *packages.ManagerRegistry, pkgs []string, opts InstallOptions) ([]resources.OperationResult, error)`
  - `UninstallPackagesWith(ctx, cfg *config.Config, lockSvc lock.LockService, registry *packages.ManagerRegistry, pkgs []string, opts UninstallOptions) ([]resources.OperationResult, error)`

- Update existing exported functions to be thin wrappers:
  - `InstallPackages(ctx, configDir string, pkgs []string, opts InstallOptions)` becomes: construct `cfg := config.LoadWithDefaults(configDir)`, `lockSvc := lock.NewYAMLLockService(configDir)`, `registry := packages.NewManagerRegistry()`, then delegate to `InstallPackagesWith`.
  - `UninstallPackages(...)` mirrors the same pattern.

Rationale:
- Library code becomes testable without relying on filesystem or global registries; commands keep using the simple API.

### C. Remove hardcoded inner timeouts (package operations)

- In `installSinglePackage` and `uninstallSinglePackage`, remove the creation of 5‑minute child contexts. Use the passed `ctx` directly for `IsAvailable`, `Install`, `Uninstall`, and `InstalledVersion` calls.

Rationale:
- Commands already pass contexts derived from configured timeouts (e.g., `cfg.PackageTimeout` in `install.go`). Let the caller own cancellation to enable short, deterministic unit tests.

Compatibility:
- No visible change. Parent command contexts already constrain operation time.

### D. Deterministic ordering in upgrade + injected registry

- Sort manager keys before iterating `spec.ManagerTargets`.
- Sort each manager’s package list before processing.
- Change `executeUpgrade` to accept a `*packages.ManagerRegistry` argument (or introduce `executeUpgradeWithRegistry` and delegate from `executeUpgrade`). Default command path still constructs the registry and passes it in.

Rationale:
- Stabilizes JSON/table output ordering for tests and logs. Registry injection unlocks pure/unit test coverage of upgrade logic without global state.

Compatibility:
- No behavioral change; only iteration order becomes deterministic.

## File‑Level Work Items

- `internal/resources/packages/operations.go`
  - Change helper signatures to use `lock.LockService`.
  - Add `InstallPackagesWith` / `UninstallPackagesWith` and delegate from existing functions.
  - Remove 5‑minute `context.WithTimeout` calls inside install/uninstall helpers; rely on caller `ctx`.

- `internal/commands/upgrade.go`
  - Sort keys of `spec.ManagerTargets` and sort each manager’s package slice before processing.
  - Inject registry into upgrade execution (adjust function signature or add a delegating variant).
  - Keep CLI output and summary identical.

## Tests to Add/Adjust

- Package operations
  - New unit tests that pass a fake `LockService` to `installSinglePackage`/`uninstallSinglePackage` to cover:
    - already‑managed, add/remove success, lock write failure, pass‑through uninstall.
  - Tests for honoring caller context cancellation (short timeouts) without inner 5‑minute overrides.

- Upgrade ordering
  - Test that results are processed in deterministic order when multiple managers and packages are present.
  - Test that injected registry is used by the upgraded execution path (e.g., via `WithTemporaryRegistry`).

No changes expected to existing CLI goldens except stable ordering (which should only make them less flaky).

## Acceptance Criteria

- All unit and integration tests pass locally.
- No CLI behavior or UX changes (flags, text, exit codes) except deterministic ordering where it was previously map‑dependent.
- `installSinglePackage` / `uninstallSinglePackage` accept `lock.LockService` and tests successfully inject a fake lock.
- `InstallPackagesWith` / `UninstallPackagesWith` exist; legacy functions delegate and are covered by tests.
- No hardcoded inner 5‑minute timeouts remain in package install/uninstall helpers.
- Upgrade iterates managers and packages in sorted order and accepts an injected registry (directly or via a delegating variant).

## Risks and Mitigations

- Risk: Signature drift in internal helpers.
  - Mitigation: Keep existing exported APIs and add `With(...)` variants; only unexported helper params change to interfaces.

- Risk: Timeout behavior change.
  - Mitigation: Commands already set explicit timeouts using config; add tests to confirm behavior matches expectations.

- Risk: Output order changes affect brittle tests.
  - Mitigation: Sorting provides stable order; update or relax any order‑dependent tests as needed.

## Rollout Plan

1) Land helper signature changes and `With(...)` variants with unit tests.
2) Remove inner hardcoded timeouts; add cancellation/timeout tests.
3) Add deterministic ordering to upgrade and registry injection; add ordering tests.
4) Run full test suite and adjust any brittle order‑dependent assertions.

## Notes

- Phase 2 can introduce a small `Timeouts` struct to thread config defaults more explicitly, but Phase 1 relies on caller‑owned contexts to keep changes minimal.
- Search path already uses configurable timeouts; no changes required there in Phase 1.

---

Completion Summary (2025-09-25)

- Implemented interface-first helpers and With(...) entrypoints:
  - Updated `installSinglePackage` and `uninstallSinglePackage` to accept `lock.LockService` and registry param in `internal/resources/packages/operations.go`.
  - Added `InstallPackagesWith` / `UninstallPackagesWith` and delegated from existing functions.
- Removed hardcoded inner timeouts:
  - Package operations now honor caller context; commands already pass `cfg.PackageTimeout` timeouts.
- Deterministic ordering + registry injection for upgrade:
  - `executeUpgrade` now sorts managers and per-manager packages and accepts an injected registry in `internal/commands/upgrade.go`.
- Added focused unit tests:
  - `internal/resources/packages/operations_injection_test.go` for lock injection, already-managed, and lock write failure.
  - `internal/commands/upgrade_order_test.go` for deterministic ordering with injected registry.
- All tests pass locally.
