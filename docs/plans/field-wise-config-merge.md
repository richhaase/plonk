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

### Phase 1 – ManagerConfig field-wise merge helper (3–4h)

1. Add merge helpers in `internal/config/managers.go`:
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
         - For each key in override:
           - If value is non-zero `MetadataExtractorConfig`, replace base’s value for that key.
         - Keys not mentioned in override remain as in base.
       - Nil override map means “inherit base” (no change).
     - For pointer fields (`NameTransform`):
       - If override is nil, keep base.
       - If non-nil, use override as-is.

2. Tests:
   - Add `internal/config/managers_test.go`:
     - Cases for each field type (scalars, slices, maps, pointer) verifying:
       - Empty override fields do not clobber base.
       - Non-empty override fields override base.
       - MetadataExtractors merge per-key.

### Phase 2 – Integrate field-wise merge into Load (3–4h)

1. Update `LoadFromPath` manager merging:
   - Instead of:
     ```go
     for name, userMgr := range userManagers {
         cfg.Managers[name] = userMgr
     }
     ```
   - Use:
     ```go
     for name, userMgr := range userManagers {
         base := cfg.Managers[name] // may be zero if not in defaults
         cfg.Managers[name] = MergeManagerConfig(base, userMgr)
     }
     ```
   - Behavior:
     - For managers with shipped defaults: effective config is default merged with user overrides.
     - For user-only managers (no default): effective config is `MergeManagerConfig(ManagerConfig{}, userMgr)`, i.e., essentially the user config.

2. Adjust `GetNonDefaultManagers` expectations:
   - Reconfirm that `GetNonDefaultManagers` still:
     - Treats managers not in defaults as custom.
     - Treats built-ins as non-default if the merged `ManagerConfig` differs from the default one.
   - Add/update tests in `internal/config/user_defined_test.go` to confirm behavior with partially overridden managers (only a subset of fields changed).

3. Tests:
   - Update config loading tests to:
     - Verify that overriding just one field (e.g., `managers.npm.list.command`) leaves other default fields intact.

### Phase 3 – Expand field-wise merge to other config areas (optional but recommended, 3–5h)

1. Evaluate additional merge candidates:
   - `Dotfiles.UnmanagedFilters`:
     - Possibly treat user list as additive to defaults rather than full replacement (or leave as-is if that’s undesirable).
   - `IgnorePatterns` / `ExpandDirectories`:
     - Decide whether to keep current “user list replaces default list” semantics or support additive patterns (e.g., `append` vs `replace`).

2. Plan minimal changes:
   - To avoid surprises, we may:
     - Start by merging only managers (Phase 2).
     - Document current behavior for lists (full replacement) and add explicit “additive” options later if needed.

> For now, consider this phase optional and driven by concrete user need.

### Phase 4 – config show highlighting (4–6h)

1. Introduce field-diffing helper:
   - Add a helper in `internal/config` or `internal/output`:
     ```go
     type FieldDiffChecker interface {
         IsFieldUserDefined(fieldPath []string, value interface{}) bool
     }
     ```
   - Reuse or extend `UserDefinedChecker` to support nested paths, especially under `managers`:
     - Example paths:
       - `["default_manager"]`
       - `["managers", "npm", "install", "command"]`

2. YAML AST traversal in `ConfigShowFormatter`:
   - Modify `ConfigShowFormatter.TableOutput` to:
     - Marshal `c.Config` to a `yaml.Node` instead of raw bytes.
     - Walk the AST, keeping track of the current path (e.g., using parent mapping).
     - For each scalar node:
       - If `IsFieldUserDefined(path, value)` is true, wrap `node.Value` with `output.ColorInfo`.
     - Render the modified AST back to YAML text.

3. Tests:
   - Add tests in `internal/output/config_formatter_test.go`:
     - With a config that differs from defaults in a few known spots (e.g., `default_manager`, `package_timeout`, `managers.npm.install.command`).
     - Assert that:
       - The YAML structure is unchanged (same keys/shape).
       - The custom values are wrapped in the color escape sequences produced by `ColorInfo`.
     - Ensure that JSON/YAML structured outputs remain uncolored (no escape sequences).

4. NO_COLOR & terminals:
   - Ensure colorization respects `NO_COLOR` and terminal detection via `output.InitColors`.
   - If colors are disabled, the highlighting should fall back to plain text without loss of information.

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
