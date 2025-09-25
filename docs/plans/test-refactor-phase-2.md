# Testability Refactor: Phase 2

Status: Completed (2025-09-25)

## Goals

- Normalize timeout handling with a central Timeouts struct.
- Make `clone` setup a thin, testable wrapper around helpers.
- Preserve CLI behavior and UX.

## Delivered Changes

1) Centralized Timeouts
- Added `internal/config/timeouts.go` with `Timeouts` struct and `GetTimeouts(cfg)` helper.
- Updated commands to use cfg-derived durations:
  - `internal/commands/install.go` and `uninstall.go` now use `GetTimeouts(cfg).Package`.
  - `internal/commands/search.go` uses `GetTimeouts(cfg).Operation` for overall and per-manager contexts.

2) Doctor uses configurable timeout
- Added `diagnostics.RunHealthChecksWithContext(ctx)`; default `RunHealthChecks()` remains 30s for compatibility.
- `internal/commands/doctor.go` now builds a context from `cfg.OperationTimeout` and calls `RunHealthChecksWithContext`.

3) Clone thin wrapper
- Extracted `SetupFromClonedRepo(ctx, plonkDir, hasConfig, noApply)` from `CloneAndSetup`.
- `CloneAndSetup` now clones and delegates to `SetupFromClonedRepo`; detection and apply remain unchanged.

## Validation

- Full unit test suite passes locally.
- CLI behavior unchanged; timeouts are now centrally derived.

## Notes

- Phase 3 (optional) could introduce more fine-grained FS interfaces where useful, but not necessary now.
