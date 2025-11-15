# Manager-Specific References in Code (To Remove or Encapsulate)

Goal: The core should be manager-agnostic. This document lists all places in code that currently embed package-manager specifics and should be removed, normalized behind config, or encapsulated in dedicated adapters.

## Core Package Code
- `internal/resources/packages/generic.go:182`
  - Special-cases npm/pnpm parseable paths (node_modules) when parsing `list` output.
- `internal/resources/packages/generic.go:215`
  - Applies Homebrew alias expansion during list parsing.
- `internal/resources/packages/generic.go:223`
  - `expandBrewAliases` helper implements brew-specific alias logic.

- `internal/resources/packages/operations.go:124,171,188,223`
  - Multiple branches for `manager == "go"` (dead code - Go is NOT a built-in manager).
- `internal/resources/packages/operations.go:193`
  - npm scoped package handling (adds `full_name`/`scope`).
- `internal/resources/packages/operations.go:305-313`
  - Hard-coded install suggestions for brew, npm, cargo, gem, uv (note: includes 'go' which is not a built-in manager).

- `internal/resources/packages/reconcile.go:34-37`
  - Normalizes `homebrew -> brew` during package reconcile.

## CLI / Orchestration Surfaces
- `internal/clone/setup.go:179-210`
  - Hard-coded manager descriptions and install instructions (Homebrew, npm, cargo, uv, gem, go).

- `internal/output/search_formatter.go:76`
  - Message mentions “Homebrew or NPM”.

- `internal/commands/upgrade.go:54,97,178,194-195`
  - Tracks npm scoped `FullName`; explicit checks for `info.Manager == "npm"`.

## Shared Types / Config
- `internal/resources/types.go:48`
  - Comment references "homebrew", "npm".

- `internal/config/validators.go:21`
  - `knownManagers` slice embeds manager names: `apt, brew, npm, uv, gem, go, cargo, test-unavailable`.

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
