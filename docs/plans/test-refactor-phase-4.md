# Testability Refactor: Phase 4

Status: Completed (2025-09-25)

## Goal

Extract pure logic from commands to enable direct unit testing without the CLI harness, while preserving CLI behavior.

## Delivered Changes

- Search command
  - Added `commands.Search(ctx, cfg, packageSpec) (SearchOutput, error)` that performs search and returns typed results.
  - `runSearch` now loads config, calls `Search`, and renders results (no behavior change).

- Info command
  - Added `commands.Info(ctx, packageSpec) (InfoOutput, error)` that implements priority logic and manager-specific lookups.
  - `runInfo` delegates to `Info` and renders results (no behavior change).

## Validation

- Full unit test suite passes locally.
