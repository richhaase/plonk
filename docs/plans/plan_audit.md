# Plan Audit – Package Manager Refactor

**Date**: 2025-11-15
**Branch**: `cleanup/config-driven-package-managers`
**Scope**: Work tracked in `docs/plans/*` vs. current code

## Summary of Completed Work

- **Core Package Code (4/4 completed)**
  - npm/pnpm list handling migrated to JSON/`json-map` and driven by `ManagerConfig.List`:
    - npm uses `list -g --depth=0 --json` + `Parse: "json-map"` with `JSONField: "dependencies"`.
    - pnpm uses `list -g --depth=0 --json` + `Parse: "json"` with `JSONField: "name"`.
    - All npm/pnpm `node_modules` path parsing and Homebrew alias expansion were removed from `GenericManager`.
  - Go manager special-casing removed:
    - All `if manager == "go"` branches are gone from `internal/resources/packages/operations.go` and `internal/commands/upgrade.go`.
    - The `ExtractBinaryNameFromPath` helper and Go-specific tests were deleted or neutralized.
  - npm scoped package metadata is produced via config-driven metadata extractors:
    - `ManagerConfig.MetadataExtractors` for npm define `scope` and `full_name` extraction.
    - `applyMetadataExtractors` in `operations.go` populates these fields generically.

- **CLI / Orchestration (2/3 completed)**
  - CLI help examples are partly dynamic and config-driven:
    - `install`, `uninstall`, and `upgrade` commands no longer embed hard-coded manager examples in `Long`.
    - Their `Example` fields are computed at runtime by helpers in `internal/commands/helpers.go`:
      - Helpers read the active config via `config.LoadWithDefaults` and build examples from the configured manager names (plus a few generic examples).
  - Clone/setup and search messaging use configuration:
    - `internal/clone/setup.go`’s `getManagerDescription` and `getManualInstallInstructions` now derive text from `GetDefaultManagers()` (via `ManagerConfig.Description`/`InstallHint`), with generic fallbacks; no inline manager-specific strings remain.
    - `internal/output/search_formatter.go`’s “no-managers” hint uses `cfg.Managers` from config to mention example managers instead of hard-coded “Homebrew or NPM”.
  - Upgrade targeting is per-manager and config-driven:
    - `ManagerConfig.UpgradeTarget` added; npm uses `full_name_preferred` by default.
    - `determineUpgradeTarget` in `upgrade.go` selects targets based on this field, eliminating the previous `info.Manager == "npm"` special case.

- **Shared Types / Config (5/6 completed)**
  - Config schema extended and wired:
    - `ManagerConfig` now includes `Description`, `InstallHint`, `HelpURL`, `UpgradeTarget`, `NameTransform`, and `MetadataExtractors`.
    - Defaults for built-in managers (brew, npm, pnpm, cargo, pipx, conda, gem, uv) are defined in `GetDefaultManagers` and used consistently across clone, diagnostics, operations, and help.
  - Validation is registry/config-driven:
    - `internal/config/validators.go` no longer uses a hard-coded `knownManagers` list; it relies on `SetValidManagers` and, if unset, the keys from `GetDefaultManagers()`.

## Deviations / Additions vs. Plan

- **Upgrade FullName Tracking – implemented via `UpgradeTarget`**
  - The plan suggested a “unified package identifier system” and name-normalization pipeline; instead, we implemented a minimal, targeted mechanism:
    - `ManagerConfig.UpgradeTarget` controls how each manager chooses the upgrade target (`name` vs. `full_name_preferred`).
    - npm uses `full_name_preferred` so scoped packages use `full_name` when present; other managers default to `name`.
  - This is slightly simpler than the original proposal but satisfies the architectural goal (no hard-coded npm logic in `upgrade`).

- **Homebrew → brew normalization**
  - The original references documented a `homebrew -> brew` normalization as a low-priority compat shim.
  - We ultimately **removed** this normalization from:
    - `internal/resources/packages/reconcile.go` (no manager renaming during reconcile).
    - `internal/clone/setup.go` (no aliasing `homebrew` to `brew` when looking up descriptions/hints).
  - This goes slightly beyond the plan by eliminating even this legacy manager alias in code; any aliasing now must come from configuration or lock data, not hard-coded logic.

- **CLI help generation**
  - The plan called for dynamic help generation from the registry; we implemented:
    - Dynamic examples for `install`, `uninstall`, and `upgrade` using configuration (`ManagerConfig`) instead of the registry APIs directly.
  - Other commands still use static examples, but:
    - Their examples are either generic or refer only to built-in managers for which we ship config (“batteries included”), which is acceptable under the reframed goal.

## Remaining Follow-ups

1. **CLI Help Examples (Other Commands)**
   - Optionally extend the dynamic-example helpers to other commands (e.g., `status`, `apply`, `doctor`, `config`) if they start to accumulate manager-specific examples.
   - For now, their examples are either generic or limited to doc-level mentions, so this is nice-to-have rather than required.

2. **Comments / Documentation References**
   - A few comments and docs still mention specific managers (e.g., examples like `brew:wget` or `npm:@types/node` in comments and docs). These are acceptable as examples but could be updated over time to match the latest defaults.

3. **Metadata & Validation Enhancements**
   - Add validation for `MetadataExtractors` (regex compilation, group bounds) and future `NameTransform` usage.
   - Consider whether any future lock-file shape changes (beyond additive metadata) will require a formal migration path; for now, behavior changes have been additive and backward-compatible.

Overall, the codebase now matches the core intent of the plans: all manager-specific behavior flows through configuration (`ManagerConfig` and defaults), “go” is no longer treated as a built-in manager, and any remaining references to specific managers in CLI output are either config-driven or exist only in examples for the default, shipped managers.***
