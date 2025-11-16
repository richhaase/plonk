# Field-Wise Config Merge & Highlighting Plan

**Status**: Draft – not started
**Scope**: Introduce field-wise merge semantics for configuration (especially `managers:`) and enhance `plonk config show` to clearly highlight customized fields.

---

## Problem Statement

Plonk’s configuration system is largely config-driven, but there are two related UX/behavior gaps:

1. **Manager overrides are all-or-nothing**
   - The load path (`LoadFromPath`) currently:
     - Starts from `defaultConfig`.
     - Merges YAML into `cfg`.
     - Replaces `cfg.Managers` with defaults from `GetDefaultManagers()`.
     - Overwrites those defaults with any user-defined `cfg.Managers` from YAML.
   - For a given manager, the user-provided `ManagerConfig` replaces the entire default `ManagerConfig`.
   - This means:
     - To change just `list.command`, a user effectively has to re-specify all other fields (install, uninstall, idempotent errors, etc.) or risk zeroing them.
     - The documentation’s desired semantics (“only specify fields you want to change; others fall back to defaults”) are not true in code.

2. **`plonk config show` does not visually distinguish custom vs default fields**
   - `config show` currently:
     - Marshals the effective `config.Config` to YAML and prints it.
   - There is no visual indication which fields were customized vs inherited from defaults.
   - `plonk config edit` plus `UserDefinedChecker` and `saveNonDefaultValues` keep `plonk.yaml` minimal, but it’s still difficult for a user to see at a glance:
     - “What did I actually change?”
     - “Which parts of the manager definitions are mine vs shipped defaults?”

These gaps undermine one of Plonk’s core goals: a config-driven system that is both safe to extend and easy to understand.

---

## Goals

1. **Field-wise merge for managers**
   - When a user overrides a built-in manager under `managers:`, only the provided fields should change; any omitted fields should fall back to `GetDefaultManagers()`.
   - This should apply recursively to nested structs such as:
     - `ListConfig` (command, parse, json_field).
     - `CommandConfig` (command, idempotent_errors).
     - Optional fields (`Description`, `InstallHint`, `HelpURL`, `UpgradeTarget`, `NameTransform`, `MetadataExtractors`).

2. **Consistent semantics across config layers**
   - `LoadFromPath`, `LoadWithDefaults`, `SimpleValidator.ValidateConfigFromYAML`, and `plonk config edit` must all see the same effective field-wise merge behavior.
   - Field-wise merge behavior must be compatible with:
     - Zero-config defaults.
     - Minimal `plonk.yaml` (only diffs).
     - The manager diffing logic (`GetNonDefaultManagers`).

3. **Visual highlighting of custom fields in `config show`**
   - `plonk config show` should clearly indicate fields that differ from defaults, ideally via color (e.g., blue) without changing the underlying YAML semantics.
   - At a glance, a user should be able to see:
     - Which top-level fields they customized.
     - Which manager definitions and subfields are custom.

4. **First-class managers**
   - Field-wise merge must be compatible with the “all configured managers are first-class” goal:
     - Any manager present in `cfg.Managers` is usable as `default_manager` and should not be treated as “second-class” due to merge behavior.

5. **Backward compatibility**
   - Existing configs (that fully override manager fields) should remain valid and retain their behavior.
   - The lock file and runtime behavior for package operations must not break.

---

## Non-Goals

- Do **not** change the lock file format or semantics.
- Do **not** introduce per-field precedence rules beyond “user overrides default”; no multi-source layering beyond default + user YAML.
- Do **not** attempt to colorize JSON/YAML output formats; highlighting is table-mode only.
- Do **not** change how `applyDefaults` handles basic scalar fields (timeouts, etc.) apart from documenting behavior where needed.

---

## Current Behavior (Summary)

### Load & validation

- `LoadFromPath`:
  - Copies `defaultConfig` into `cfg`.
  - Unmarshals YAML over `cfg` (scalar fields overlay; lists/maps override).
  - Saves `cfg.Managers` into `userManagers`.
  - Replaces `cfg.Managers` with `GetDefaultManagers()`.
  - Overwrites defaults with `userManagers` entries (full replacement per manager name).
  - Calls `updateValidManagersFromConfig(&cfg)` to feed `validmanager`.
  - Validates via `validator.Struct(&cfg)` with `validmanager`.

- `SimpleValidator.ValidateConfigFromYAML`:
  - Unmarshals YAML into `cfg`.
  - Calls `applyDefaults(&cfg)` for scalars & common fields.
  - Calls `updateValidManagersFromConfig(&cfg)`.
  - Validates via `validator.Struct(&cfg)`.

### Manager diffing

- `UserDefinedChecker.GetNonDefaultManagers(cfg *Config)`:
  - Compares `cfg.Managers` against `GetDefaultManagers()`.
  - Returns:
    - All managers not present in `GetDefaultManagers()` (purely user-defined).
    - Built-ins where `reflect.DeepEqual` differs between effective and default `ManagerConfig`.
- `saveNonDefaultValues`:
  - Writes `GetNonDefaultFields` and, if non-empty, `managers: GetNonDefaultManagers(cfg)` to `plonk.yaml`.

### config show / config edit

- `config show`:
  - Uses `config.LoadWithDefaults`.
  - `ConfigShowFormatter.TableOutput` just `yaml.Marshal`s the full `config.Config` to YAML.

- `config edit`:
  - Uses `config.LoadWithDefaults`.
  - `createTempConfigFile` writes a header plus `writeFullConfig` (full YAML).
  - `parseAndValidateConfig` strips header comments and feeds the YAML through `SimpleValidator`.
  - `saveNonDefaultValues` persists only diffs (top-level + manager diffs).

---

## High-Level Design Options

### Manager merge strategy

1. **Field-wise merge helper (recommended)**
   - Introduce a helper:
     ```go
     func MergeManagerConfig(base, override ManagerConfig) ManagerConfig
     ```
   - Behavior:
     - Start from `base` (default).
     - For each field in `override`, if non-zero/non-empty, copy it over.
     - For nested structs:
       - `ListConfig`: merge `Command`, `Parse`, `ParseStrategy`, `JSONField`.
       - `CommandConfig`: merge `Command`, `IdempotentErrors`.
       - `NameTransformConfig`: if override non-nil, use it fully (no further merge for now).
       - `MetadataExtractors`: treat override map as full replacement (per-key) for now; deeper merge is optional future work.
   - Pros:
     - Maintains default behavior while allowing partial overrides.
     - Isolates complexity in one helper function.
   - Cons:
     - Requires careful design for “zero/empty” vs “intentional clear”.

2. **Explicit “full override” mode**
   - Keep current behavior (full replacement of `ManagerConfig`), but:
     - Document it explicitly.
     - Add validation/warnings if required fields are missing for overridden managers.
   - Pros:
     - Minimal code change.
   - Cons:
     - Does not satisfy the desired UX; users must copy a lot of boilerplate.

**We choose Option 1** with a conservative “zero means use base” rule and no explicit “clear” semantics for now.

### Highlighting strategy in config show

1. **Colorize full YAML lines (recommended)**
   - Reuse `UserDefinedChecker` and `config.GetDefaults` to determine which fields differ from defaults.
   - Parse the YAML into an AST (`yaml.Node`), traverse it with context, and:
     - For each scalar node corresponding to a field that differs from defaults, wrap the scalar text in a color function (e.g., `output.ColorInfo`).
   - Pros:
     - Keeps YAML structure untouched; only adds ANSI color codes in table mode.
     - Works for both top-level fields and nested manager fields.
   - Cons:
     - Requires a small amount of AST mapping (field path → isCustom).

2. **Add inline comments or markers**
   - Reintroduce `(user-defined)` comments next to custom fields.
   - Pros:
     - No color required, easier to parse visually.
   - Cons:
     - Pollutes YAML for users who may copy/paste it elsewhere.

**We choose Option 1**, using color (blue) for custom fields in table output only; JSON/YAML formats remain unmodified.

---

## Proposed Implementation

### Phase 0 – Baseline & Tests (1–2h)

1. Document current manager behavior:
   - Add tests in `internal/config/config_test.go` that capture the current “full replacement” behavior for manager overrides.
   - Ensure these tests are clearly named as “pre-merge behavior” for easy update.

2. Document current config show behavior:
   - Confirm via a golden test or assertion that `ConfigShowFormatter.TableOutput` is currently plain `yaml.Marshal`.

> This phase mainly documents the starting point so behavior changes are deliberate and test-driven.

### Phase 1 – ManagerConfig field-wise merge helper (3–4h) – ✅ Completed 2025-11-16

1. Add merge helpers in `internal/config/managers.go` (✅ implemented):
   ```go
   func MergeManagerConfig(base, override ManagerConfig) ManagerConfig
   func mergeListConfig(base, override ListConfig) ListConfig
   func mergeCommandConfig(base, override CommandConfig) CommandConfig
   ```
   - Rules:
     - For `[]string` fields (`Command`, `IdempotentErrors`):
       - Non-empty override slice replaces base; empty slice means “inherit base”.
     - For scalar/string fields (`Parse`, `ParseStrategy`, `JSONField`, `Description`, `InstallHint`, `HelpURL`, `UpgradeTarget`):
       - Non-empty override replaces base; empty string means “inherit base”.
     - For maps (`MetadataExtractors`):
       - If override map is non-nil:
         - Start from a copy of the base map.
         - For each key in override, set/replace the value in the result.
         - Keys not mentioned in override remain as in base.
       - Nil override map means “inherit base” (return a copy of base or nil when base is nil).
     - For pointer fields (`NameTransform`):
       - If override is nil, keep base.
       - If non-nil, use override as-is.

2. Tests (✅ implemented in `internal/config/managers_test.go`):
   - Cases for each field type (scalars, slices, maps, pointer) verifying:
     - Empty override fields do not clobber base.
     - Non-empty override fields override base.
     - `MetadataExtractors` merge per-key (existing keys overridden, new keys added, unchanged keys preserved).

### Phase 2 – Integrate field-wise merge into Load (3–4h) – ✅ Completed 2025-11-16

1. Update `LoadFromPath` manager merging (✅ implemented in `internal/config/config.go`):
   - Replaced the direct overwrite logic:
     ```go
     for name, userMgr := range userManagers {
         cfg.Managers[name] = userMgr
     }
     ```
   - With field-wise merge:
     ```go
     for name, userMgr := range userManagers {
         base := cfg.Managers[name] // may be zero if not in defaults
         cfg.Managers[name] = MergeManagerConfig(base, userMgr)
     }
     ```
   - Behavior:
     - For managers with shipped defaults: effective config is default merged with user overrides.
     - For user-only managers (no default): effective config is `MergeManagerConfig(ManagerConfig{}, userMgr)`, i.e., essentially the user config.

2. Adjusted tests in `internal/config/config_test.go` (✅ implemented):
   - Added `TestLoad_ManagerFieldWiseMerge`:
     - Defines a config that overrides only `managers.npm.install.command`.
     - Verifies that:
       - `cfg.Managers["npm"].Install.Command` matches the override.
       - `Install.IdempotentErrors` remains as in the default npm manager.
       - Other sub-configs such as `List` and `Uninstall` still match defaults.

3. `GetNonDefaultManagers` remains valid:
   - Its behavior (treat managers not in defaults as custom, and built-ins as non-default when the merged config differs from defaults) is still correct under field-wise merge and did not require code changes, but the new tests indirectly validate compatibility by exercising merged configs through `Load`.

### Phase 3 – Expand field-wise merge to other config areas (optional but recommended, 3–5h) – ✅ Completed 2025-11-16

1. Evaluate additional merge candidates (✅ completed by analysis, no code changes):
   - `Dotfiles.UnmanagedFilters`:
     - Current behavior: user-provided `dotfiles.unmanaged_filters` replaces the default list.
     - Changing this to additive semantics would make it harder for users to *remove* default filters.
   - `IgnorePatterns` / `ExpandDirectories`:
     - Current behavior: user-provided lists fully replace defaults.
     - Making these additive by default could unexpectedly hide or include files compared to existing setups.

2. Decision (✅ documented behavior, no behavioral change):
   - To avoid surprising users and silently changing how dotfiles and ignore patterns behave, we keep the current “user list replaces default list” semantics for:
     - `dotfiles.unmanaged_filters`
     - `ignore_patterns`
     - `expand_directories`
   - Field-wise merge is therefore scoped to:
     - Manager definitions (`ManagerConfig` and nested types).
     - Scalar config fields already handled by `applyDefaults`.
   - This decision is documented in configuration docs so users understand that:
     - Manager configs can be partially overridden.
     - List fields (ignore patterns, expand directories, unmanaged filters) remain full overrides for now.

> If a future need arises for additive list semantics, we can introduce explicit mechanisms (e.g., “append” sections or flags) without changing the current default behavior.

### Phase 4 – config show highlighting (4–6h) – ✅ Completed 2025-11-16

1. Highlight custom fields in `ConfigShowFormatter` (✅ implemented in `internal/output/config_formatter.go`):
   - `ConfigShowFormatter.TableOutput` now:
     - Detects when `Config` is a `*config.Config` and `Checker` is a `*config.UserDefinedChecker`.
     - Uses `checker.GetNonDefaultFields(cfg)` to identify top-level fields that differ from defaults.
     - Uses `checker.GetNonDefaultManagers(cfg)` to identify managers that differ from defaults.
     - Marshals the effective config to YAML and post-processes the text line-by-line:
       - Top-level keys in the non-default set are colorized with `ColorInfo(...)` (entire line).
       - Within the `managers:` block, manager name lines (e.g., `  npm:`) are colorized when the manager name is in the non-default manager set.
     - Falls back to plain `yaml.Marshal` output when type assertions or highlighting fail.
   - The YAML structure and keys remain unchanged; only color codes are added in table output.

2. Tests (✅ implemented in `internal/output/config_formatter_test.go`):
   - `TestConfigShowFormatter_TableAndStructured`:
     - Confirms the basic table output and `StructuredData` behavior still work.
   - `TestConfigShowFormatter_HighlightsCustomFields`:
     - Creates a config with a non-default `DefaultManager` (`npm`) and defaults for other fields.
     - Uses a `UserDefinedChecker` backed by a temp config directory.
     - Asserts that:
       - The output contains the `default_manager:` key.
       - The value `npm` appears in the output (indicating the custom value is present; exact escape sequences are not asserted).

3. NO_COLOR & terminals:
   - Highlighting uses `ColorInfo`, which is built on the existing color subsystem:
     - Honors `NO_COLOR` and terminal detection via `output.InitColors`.
     - When colors are disabled, all lines fall back to plain text with no escape sequences.

---

## Risks & Edge Cases

- **Zero vs “clear” semantics**
  - Field-wise merge currently treats empty/zero override fields as “inherit default”.
  - There is no way to intentionally clear a default value (e.g., remove an `idempotent_errors` entry) via YAML alone.
  - This is acceptable for now but should be documented.

- **Map merge subtleties**
  - `MetadataExtractors` and similar maps might require more nuanced merge behavior in the future (e.g., removing default extractors). For now, we treat override entries as “set/replace” and leave unspecified keys intact.

- **Config show path mapping**
  - The path mapping between AST nodes and config fields must be carefully tested to avoid miscoloring unrelated fields.

- **Behavioral changes for existing configs**
  - Field-wise merge can change behavior for users who currently rely on full replacement of `ManagerConfig`.
  - Mitigation:
    - Keep tests for old behavior and update them deliberately.
    - Document the change in the changelog and migration notes.

---

## Summary

- This plan moves Plonk’s configuration system to a true field-wise merge model for manager definitions, so users can safely override only what they care about.
- It also upgrades `plonk config show` to visually highlight custom fields using color, making it much easier to audit what’s been customized.
- The design maintains:
  - The “all configured managers are first-class” principle.
  - Minimal on-disk `plonk.yaml` via diff-based saving.
  - Backward compatibility via careful test-driven changes and documentation.
