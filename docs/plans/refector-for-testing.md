# Refactor For Testing: Raising Coverage Without Breaking Behavior

This proposal outlines concrete, low-risk refactors to improve test coverage and long‑term testability without changing Plonk’s existing behavior or CLI surface. The plan focuses on introducing seams (dependency injection), isolating pure logic from I/O, and removing hidden runtime dependencies that currently make some code paths hard to exercise in hermetic unit tests.

The theme: do more with the infrastructure we already have (e.g., MockCommandExecutor, ManagerRegistry swaps, BufferWriter), and widen a few narrow interfaces so we can test code without wiring the world each time.

---

## Current Strengths To Preserve

- Mockable process execution via `packages.CommandExecutor` with `SetDefaultExecutor`.
- Switchable package manager registry (`WithTemporaryRegistry`).
- CLI harness (`RunCLI`) that isolates HOME/PLONK_DIR and captures output.
- Dotfiles domain already abstracts filesystem operations (scanner, expander, fileops) and backs atomic writes.
- Output layer supports JSON/YAML for stable, assertion-friendly tests.

These are excellent “footholds” for testability. The refactors below leverage and extend them.

---

## Status Update (2025-09-25)

- Phase 1 completed. See details in `docs/plans/test-refactor-phase-1.md`.
- Implemented:
  - Interface‑first helpers for install/uninstall (`lock.LockService`).
  - `InstallPackagesWith` / `UninstallPackagesWith` with explicit cfg/lock/registry injection.
  - Removed hardcoded inner 5‑minute timeouts; callers own timeouts.
  - Deterministic upgrade ordering and injected registry for upgrade path.
  - Focused unit tests for injection and ordering.

Next: Phase 2 (optional) to normalize timeout threading with a `Timeouts` struct.

## High-Impact, Low-Risk Refactors

### 1) Accept Interfaces, Not Concretes, In Hot Paths

Problem:
- `installSinglePackage`/`uninstallSinglePackage` accept `*lock.YAMLLockService` (concrete) which forces real disk writes in most unit tests or workarounds like permission tricks.

Refactor:
- Change signatures to accept `lock.LockService` (interface) instead of `*YAMLLockService`.
  - `func installSinglePackage(ctx context.Context, configDir string, lockSvc lock.LockService, ...)`
  - `func uninstallSinglePackage(ctx context.Context, configDir string, lockSvc lock.LockService, ...)`

Benefits:
- Tests can pass a trivial in‑memory/mock lock implementation to hit all branches (write failures, partial success, etc.) deterministically and quickly.

Compatibility Plan:
- Keep call sites the same; they already construct `lock.NewYAMLLockService(configDir)` which satisfies `LockService`.
- This is purely a type‐widening change (no behavior change), safe to ship.

### 2) Remove Hidden Runtime Dependencies From Library Functions

Problem:
- `InstallPackages`/`UninstallPackages` call `config.LoadWithDefaults(configDir)` internally to find defaults (e.g., manager). This makes unit tests depend on actual config file system state and buries dependencies.

Refactor:
- Add variant with explicit dependencies:
  - `InstallPackagesWith(cfg *config.Config, lockSvc lock.LockService, registry *packages.ManagerRegistry, ...)` and use the existing `InstallPackages` as a thin wrapper that resolves defaults and delegates.
- Same for `UninstallPackages`.

Benefits:
- Unit tests can call the `With(...)` variants with in‑memory config and fake registry, avoiding file system and global state entirely.
- Existing callers (commands) keep using the simple API; behavior remains unchanged.

### 3) Extract Pure Logic From Cobra Commands

Problem:
- Command `RunE` functions mix parsing, config/lock I/O, orchestration, and formatting. Unit tests often need the full CLI harness to hit small slices of logic.

Refactor:
- For each long command, extract a pure function that accepts inputs and returns typed output (no printing):
  - `func Upgrade(ctx context.Context, spec upgradeSpec, cfg *config.Config, lockSvc lock.LockService, registry *packages.ManagerRegistry) (upgradeResults, error)` and make `runUpgrade` a thin CLI wrapper.
  - Repeat where applicable (e.g., search path already does this to some extent).

Benefits:
- Unit tests can call the pure function and assert on the return struct (faster, hermetic), while CLI tests continue to use `RunCLI`.
- Improves test coverage in commands without brittle table output assertions.

Compatibility Plan:
- Keep the public CLI behavior unchanged. Only code organization changes (introducing a helper function, no signature change for `RunE`).

### 4) Normalize Timeouts and Context Ownership

Problem:
- Some timeouts are hardcoded (e.g., 5 minutes in `operations.go`), while commands pass their own timeouts based on config. Mixed patterns complicate cancellation tests.

Refactor:
- Thread a `Timeouts` struct (sourced from cfg) through orchestration and operations instead of hardcoding.
- Provide helpers: `WithPackageTimeout(ctx, cfg)` to create scoped contexts.

Benefits:
- Deterministic testing of timeouts/cancellations (simulated long‑running operations) without waiting real minutes.
- Clearer ownership of context lifecycles.

### 5) Stabilize Result Ordering For Deterministic Tests

Problem:
- Some outputs aggregate maps (e.g., by manager) which are then iterated without sorting.

Refactor:
- Sort keys (managers, package names) before building tables. Sorting already exists in places (status formatter); extend this to other aggregations in commands where order is non‑semantic but improves determinism.

Benefits:
- Deterministic output for both table and JSON (
  - JSON usually stable; table sometimes wobbles if order isn’t enforced).
- Less brittle “golden snippet” checks when desired.

### 6) Broaden Existing Seams Where We Already Have Abstractions

What to reinforce:
- `packages.NewManagerRegistry()` currently returns a global default. Introduce an injectable registry path for library calls (e.g., pass a registry into `executeUpgrade/InstallPackagesWith`). Tests can then supply a minimal registry with only managers they care about.
- `output.Writer` is already injectable. Keep the policy that unit tests default to non‑terminal writers.

Benefits:
- Fewer global singletons in unit tests; more isolated and faster tests.

### 7) Git/Clone: Separate Parsing/Detection From Effects

Problem:
- Clone performs: parse URL → filesystem layout decisions → lock inspection → tool installation (via SelfInstall). Tests sometimes need to stub several layers.

Refactor:
- Extract pure helpers:
  - `ParseGitURL(input) (url string, err error)` (already present),
  - `DetectRequiredManagersFromLock(lock *lock.Lock) []string` — pure function.
  - `CreateDefaultConfigContents(defaults *config.Config) []byte` — pure.
- Ensure `CloneAndSetup` is a thin orchestration wrapper around these helpers.

Benefits:
- Unit tests can cover URL formats, detection, and config generation without touching the network or disk. The current tests already cover much of this; making them first‑class pure functions raises coverage further with minimal glue.

---

## Medium-Risk / High-Payoff (Optional) Refactors

These are safe but slightly more invasive. Consider them as a Phase 2.

### A) File System Abstraction Where Needed

If we want to raise coverage for file‑heavy code (dotfiles, lock) further without touching the real FS, introduce a minimal FS interface (or adopt an established one like afero) and inject it into hotspots:
- `AtomicFileWriter` → `Writer` interface with a prod impl and a test impl that writes to an in‑memory map.
- Dotfile scanning (DirectoryScanner) already has abstractions; ensure all direct `os.*` calls are behind those interfaces.

Benefit: broaden hermetic, high‑speed unit testing scope. Risk: churn with limited payoff if current coverage/robustness is adequate.

### B) Replace Global Defaults With Option Structs

Introduce function options for long functions that currently pull defaults implicitly:
- `InstallPackages(ctx, configDir, pkgs, InstallOptions{Manager:..., DryRun:..., Registry:..., LockService:..., Timeouts:...})`

Benefit: fewer hidden dependencies; tests can fully control execution. Risk: signature drift; mitigate via thin overloads/wrappers.

---

## Test Strategy After Refactors

### Unit Tests (Pure / Small)
- Prefer testing pure helpers (arg parsing, detection, mapping) and logic functions extracted from commands.
- Use interface injections (LockService, ManagerRegistry, CommandExecutor) to exercise failure/success branches without FS/network.
- Keep JSON assertions for CLI coverage; avoid brittle table “goldens”.

### Property/Fuzz Tests
- Expand to critical mappers: upgrade spec parsing, reconcile with key, and simple file path resolvers.
- Focus on invariants (e.g., membership, partitioning, idempotence) rather than specific outputs.

### Integration (Hermetic CLI) Tests
- Use `RunCLI` with JSON output where possible; table “snippet” checks only for coverage pokes where JSON isn’t available.
- Continue to avoid real package managers; rely on registry/exec mocks.

### Out of Scope
- BATS tests remain smoke/system only; no expansion.

---

## Phase Plan (PR‑sized Slices)

Phase 1 (Low risk, high value)
1. Change `installSinglePackage`/`uninstallSinglePackage` to accept `lock.LockService`.
2. Add `InstallPackagesWith` / `UninstallPackagesWith` to accept cfg/lock/registry explicitly; make existing functions delegate.
3. Extract `Upgrade(...)` pure function from `runUpgrade` and update tests to use it.
4. Sort result aggregations where nondeterministic order is observed (commands formatters/aggregators).

Phase 2 (Optional refinements)
5. Normalize timeout handling via a `Timeouts` struct passed down from cfg.
6. Make `CloneAndSetup` a thin wrapper around explicit helpers (detection + content builder).

Phase 3 (Only if needed)
7. Introduce a minimal FS interface for targeted areas (atomic write, select scan paths) to enable more hermetic tests.

---

## Risks and Mitigations

- Signature churn: Keep thin wrappers that preserve existing call sites. Use type‑widening (interface in place of concrete) wherever possible.
- Behavior drift: Add regression tests around command outputs (JSON), especially for upgrade and operations where we refactor. Use PRs in small steps, landing tests first where feasible.
- Performance: Sorting for determinism is negligible for current data sizes. Extracting pure functions reduces work per test.

---

## Expected Outcomes

- Easier, faster, and more deterministic unit tests, particularly around:
  - upgrade/operations lock mutation branches,
  - manager availability/error branches,
  - clone detection and config generation.
- Coverage improves meaningfully with less ceremony; unit tests stop depending on read/write to real FS for most cases.
- Reduced reliance on integration harness for basic logic, lowering test flakiness.

---

## Acceptance Criteria

- No user-visible CLI behavior changes; commands still print and format exactly the same.
- Existing tests continue to pass; new tests can inject interfaces and hit branches previously hard to reach.
- `installSinglePackage`/`uninstallSinglePackage` accept `lock.LockService` without downstream breakage.
- `InstallPackagesWith`/`UninstallPackagesWith` available and used by new unit tests; `InstallPackages`/`UninstallPackages` delegate appropriately.
- Upgrade command has a pure logic function; command wrapper delegates.
- Deterministic ordering asserted in affected output aggregations.

---

## Appendix: Concrete Examples

1) Widening types:
```go
// before
func installSinglePackage(ctx context.Context, configDir string, lockSvc *lock.YAMLLockService, name, manager string, dryRun bool) resources.OperationResult

// after (no behavior change)
func installSinglePackage(ctx context.Context, configDir string, lockSvc lock.LockService, name, manager string, dryRun bool) resources.OperationResult
```

2) Wrapper for explicit dependency injection:
```go
// New testable entrypoint
func InstallPackagesWith(ctx context.Context, cfg *config.Config, lockSvc lock.LockService, registry *packages.ManagerRegistry, pkgs []string, opts InstallOptions) ([]resources.OperationResult, error) {
    // identical logic, no implicit config/registry
}

// Old API becomes a thin wrapper
func InstallPackages(ctx context.Context, configDir string, pkgs []string, opts InstallOptions) ([]resources.OperationResult, error) {
    cfg := config.LoadWithDefaults(configDir)
    lockSvc := lock.NewYAMLLockService(configDir)
    registry := packages.NewManagerRegistry()
    return InstallPackagesWith(ctx, cfg, lockSvc, registry, pkgs, opts)
}
```

3) Extract pure logic from upgrade:
```go
// New pure function, returns structured results
func Upgrade(ctx context.Context, spec upgradeSpec, cfg *config.Config, lockSvc lock.LockService, registry *packages.ManagerRegistry) (upgradeResults, error)

// Cobra wrapper (unchanged behavior)
func runUpgrade(cmd *cobra.Command, args []string) error {
    // parse, load cfg/lock, build spec
    res, err := Upgrade(cmd.Context(), spec, cfg, lockSvc, packages.NewManagerRegistry())
    // format + print + exit semantics
}
```

These examples are illustrative; exact naming can follow existing package conventions.
