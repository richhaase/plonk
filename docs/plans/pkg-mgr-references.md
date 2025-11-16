# Manager-Specific References in Code (To Remove or Encapsulate)

Goal: The core should be manager-agnostic. This document lists all places in code that currently embed package-manager specifics and should be removed, normalized behind config, or encapsulated in dedicated adapters.

## Core Package Code
- ~~`internal/resources/packages/generic.go:182`~~ ✅ **RESOLVED**
  - Previously special-cased npm/pnpm parseable paths (`node_modules`) when parsing `list` output.
  - Now removed; npm/pnpm use JSON-based list output configured via `ManagerConfig` and parsed generically (`json`/`json-map` strategies).
- ~~`internal/resources/packages/generic.go:215`~~ ✅ **RESOLVED**
  - Homebrew alias expansion during list parsing has been removed in favor of config-driven behavior.
- ~~`internal/resources/packages/generic.go:223`~~ ✅ **RESOLVED**
  - `expandBrewAliases` helper and brew-specific alias logic have been removed from core.

- ~~`internal/resources/packages/operations.go:124,171,188,223`~~ ✅ **RESOLVED**
  - Previously contained multiple branches for `manager == "go"` (dead code – Go is not a built-in manager).
  - These branches have been removed; Go can now be supported purely as a config-defined manager, without special handling in core logic.
- ~~`internal/resources/packages/operations.go:193`~~ ✅ **RESOLVED**
  - Previously had an inline special case for npm scoped packages (`scope`/`full_name`).
  - This is now handled via `MetadataExtractors` in `ManagerConfig` and a generic `applyMetadataExtractors` helper; core code no longer hard-codes npm semantics.
- ~~`internal/resources/packages/operations.go:305-313`~~ ✅ **RESOLVED**
  - Previously used a hard-coded suggestions map for brew, npm, cargo, gem, uv, and go.
  - `getManagerInstallSuggestion` now prefers config-driven `InstallHint` values (from `ManagerConfig`) and falls back to a generic message for unknown managers.

- `internal/resources/packages/reconcile.go:34-37`
  - Normalizes `homebrew -> brew` during package reconcile.

## CLI / Orchestration Surfaces
- `internal/clone/setup.go:179-210` ✅ **RESOLVED**
  - Previously hard-coded manager descriptions and install instructions (Homebrew, npm, cargo, uv, gem, go).
  - Now derives descriptions/install hints from `ManagerConfig` defaults via `GetDefaultManagers()`; core logic no longer embeds manager-specific strings.

- `internal/output/search_formatter.go:76` ✅ **RESOLVED**
  - Previously emitted “Homebrew or NPM” explicitly.
  - Now builds the hint using the configured manager names (via `config.LoadWithDefaults` and `cfg.Managers`), showing examples like `brew or npm` only when those managers are configured.

- `internal/commands/upgrade.go:54,97,178,194-195` ✅ **RESOLVED**
  - Previously tracked npm scoped `FullName` with an explicit `info.Manager == "npm"` check.
  - Now uses a per-manager `UpgradeTarget` setting in `ManagerConfig` (e.g., `full_name_preferred` for npm) so upgrade targeting is configuration-driven rather than manager-hardcoded.

## Shared Types / Config
- `internal/resources/types.go:48`
  - Comment references "homebrew", "npm".

- `internal/config/validators.go:21` ✅ **RESOLVED**
  - Previously used a `knownManagers` slice with embedded names (`apt, brew, npm, uv, gem, go, cargo, test-unavailable`) as a fallback.
  - Now derives allowed managers from the dynamically registered list (via `ManagerRegistry`) or from `GetDefaultManagers()`; no hand-maintained manager-name slice remains.

- `internal/config/config.go:34`
  - Default `DefaultManager: "brew"`.
- `internal/config/config.go:62-64`
  - Default ignore patterns include `.npm`, `.gem`, `.cargo`.

- `internal/resources/packages/spec.go:20`
  - Comment examples include `brew:wget`, `npm:@types/node`.

## Notes
- Tests and help text (usage examples) naturally reference manager names; they are acceptable in test fixtures and CLI help, but functional/logic code should remain manager-agnostic.
- The most critical offenders are in GenericManager (parse special-casing, alias expansion) and operations (npm/go branches, install suggestions). These should be addressed by:
  - Config-driven parse strategies (json_keys/regex) and normalized defaults (npm/pnpm JSON list).
  - **Removing dead code**: Delete all `if manager == "go"` blocks (Go is not a built-in manager).
  - Moving npm scoped package handling to metadata framework.
  - A small, isolated name normalizer per manager (e.g., brew canonicalization/aliases) invoked outside GenericManager, or lock-time canonicalization.
