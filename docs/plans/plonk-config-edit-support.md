# plonk config edit – Manager Configuration Support Plan

**Status**: Phase 0 completed – ready for Phase 1
**Scope**: Make `plonk config edit` a first-class way to view and edit package manager configuration in `plonk.yaml`, while ensuring `config edit` and `config show` present the same effective configuration data.

---

## Problem Statement

Plonk’s package manager system is now fully config‑driven (`ManagerConfig`, `ManagerRegistry`, `GenericManager`), but:

- `plonk config edit` does **not** surface the `managers:` section at all.
- The visudo‑style editor only shows a handful of top‑level fields (`default_manager`, timeouts, `expand_directories`, `ignore_patterns`, `dotfiles`).
- Only non‑default values are saved back to `plonk.yaml`, and the current “non‑default” detection logic ignores `Managers`.
- As a result, users cannot:
  - Discover that manager config is even configurable from `plonk config edit`.
  - View/modify overrides for built‑in managers.
  - Add/edit custom managers in a guided way.

Today, the only way to change manager configuration is to manually edit `plonk.yaml` with another editor and know the schema from docs. This is inconsistent with the intended UX:

- `plonk config show` and `plonk config edit` should show **identical configuration data** (same fields, same values); `edit` simply offers a safe way to change those values and then persists only the minimal diff back to disk.

---

## Goals

1. **Expose manager configuration via `plonk config edit`**
   - Make the existence and contents of `managers:` visible in the temporary config file.
   - Allow users to add/edit/remove manager definitions within the visudo workflow.

2. **Present the same effective config in `show` and `edit`**
   - The YAML data structure presented by `config edit` (ignoring comments/annotations) must match what `config show` prints for the active configuration.
   - This includes the merged `managers:` map (built‑ins + overrides).

3. **Round‑trip only non‑default overrides to `plonk.yaml`**
   - Keep `plonk.yaml` minimal: store only differences from built‑in manager defaults plus new custom managers, along with non‑default top‑level fields.
   - Do **not** emit full built‑in manager definitions into `plonk.yaml` unless the user explicitly overrides them.

4. **Preserve zero‑config behavior**
   - No `plonk.yaml` → still works. Running `config edit` should not force a large default `managers:` block into a new file.
   - Users who never touch managers should continue to see a simple configuration file.

5. **Avoid regressions for existing configs**
   - Existing `managers:` blocks in `plonk.yaml` must be preserved and round‑trip correctly through `config edit`.
   - Existing commands (`install`, `uninstall`, `upgrade`, `clone`, `doctor`, etc.) must continue to see the same effective runtime config.

6. **Keep responsibilities clear**
   - Both `config show` and `config edit` operate on the **same effective configuration struct**.
   - `config show` renders it to stdout; `config edit` writes it to a temp file for editing and, on save, persists only non‑default differences back to `plonk.yaml`.

7. **Treat all configured managers as first‑class**
   - Any manager present in the effective `cfg.Managers` (whether from shipped defaults or user config) should be usable everywhere a manager name is expected (registry, `default_manager`, validation, help/examples) without behavioral distinction.

---

## Non‑Goals

- Do **not** redesign the overall config schema or `ManagerConfig`.
- Do **not** change lock file format or semantics.
- Do **not** turn `config edit` into an interactive wizard; it remains an editor‑based workflow on YAML.

---

## Current Behavior (Summary)

### Config model

- Runtime config: `internal/config.Config` (`internal/config/config.go`)
  - `Managers map[string]ManagerConfig` holds all known managers (built‑ins from `GetDefaultManagers()` + user overrides from `plonk.yaml`).
  - `LoadFromPath`:
    - Starts from `defaultConfig`.
    - Unmarshals YAML on top.
    - Merges default managers with user managers:
      - `cfg.Managers = GetDefaultManagers()` → then override/add from user file.
    - Validates via `validator` + `RegisterValidators`.
- Default managers: `GetDefaultManagers()` (`internal/config/managers_defaults.go`).
- Manager registry: `ManagerRegistry` + `LoadV2Configs` (`internal/resources/packages/registry.go` and `init.go`).

### Config show

- `plonk config show`:
  - Loads runtime config via `config.LoadWithDefaults(configDir)`.
  - Prints it via `yaml.Marshal(cfg)` in `internal/output/config_formatter.go`.
  - This **does** include `managers:` with the merged default+user definitions.

### Config edit

- `plonk config edit` (`internal/commands/config_edit.go`):
  - Creates a temp file with header comments and then calls `writeFullConfig`.
  - `writeFullConfig`:
    - Serializes the entire effective `config.Config` (from `config.LoadWithDefaults`) to YAML, including the merged `managers:` map.
    - Ensures that, aside from the header comments, the YAML body in the temp file matches what `config show` would print for the same config.
  - `parseAndValidateConfig`:
    - Strips header comments (and any legacy `(user-defined)` annotations) then uses `config.NewSimpleValidator().ValidateConfigFromYAML`.
    - On success, unmarshals into `config.Config`, applies simple defaults to top‑level fields, and returns it.
  - `saveNonDefaultValues`:
    - Uses `config.NewUserDefinedChecker(configDir).GetNonDefaultFields(cfg)` to compute a **map of non‑default top‑level fields**.
    - (After Phase 3) will also use `GetNonDefaultManagers` to add a `managers` entry containing only non‑default/custom managers.
    - Serializes that map via `yaml.Marshal` and writes it to `plonk.yaml`.

### UserDefinedChecker

- Defined in `internal/config/user_defined.go`.
- Holds:
  - `defaults`: copy of `defaultConfig` (top‑level config only, no managers).
  - `userConfig`: result of `Load(configDir)` (already merged defaults + user config).
- `IsFieldUserDefined(fieldName, value)`:
  - Uses `userConfig == nil` as a guard for “no file yet”.
  - Compares `value` vs `defaults` using `reflect.DeepEqual`.
- `GetNonDefaultFields(cfg)`:
  - Returns a map with overrides for:
    - `default_manager`, `operation_timeout`, `package_timeout`, `dotfile_timeout`,
      `expand_directories`, `ignore_patterns`, `dotfiles`.
  - **Ignores `Managers` entirely.**

---

## High‑Level Design Options

### Option A – Dump full runtime config into plonk.yaml

- Change `config edit` to:
  - Write header comments.
  - Dump `yaml.Marshal(cfg)` to the temp file.
  - On save, write that full YAML directly back to `plonk.yaml`.
- Pros:
  - `config show` and `config edit` views would match.
  - Very simple implementation.
- Cons:
  - Bloats `plonk.yaml` with all default values and built‑in managers.
  - Violates the existing “minimal config” philosophy (only diffs in the file).
  - Harder for users to see what they actually changed.

**Verdict**: Reject. Too noisy; we want minimal on‑disk config.

### Option B – Show only user‑overrides and custom managers

- Extend `UserDefinedChecker` & `writeAnnotatedConfig` to:
  - Compute which managers differ from built‑in defaults.
  - Serialize only those overrides + any purely custom managers.
  - Continue to emit only non‑default top‑level fields.
- `config show` prints the full effective config; `config edit` only shows what differs from defaults.
- Pros:
  - Keeps `plonk.yaml` minimal.
  - Mirrors current behavior for top‑level fields.
- Cons:
  - **Conflicts with the requirement** that `config show` and `config edit` present identical configuration data; edit would be a filtered view.
  - Users cannot see default manager definitions while editing unless they cross‑reference `config show` or docs.

**Verdict**: Reject for now. Does not satisfy the “identical data” requirement.

### Option C – Show full effective config in edit, save only diffs

- Make `config edit` show the **full effective config**, identical to `config show`:
  - Use the same `config.LoadWithDefaults` result.
  - Serialize the whole `config.Config` (including `managers`) in the temp file.
- When saving:
  - Parse the edited YAML into a `config.Config`.
  - Compute diffs vs defaults for both top‑level fields and managers.
  - Persist **only** the non‑default differences to `plonk.yaml`.
- Pros:
  - `config show` and `config edit` have the same data (ignoring comments).
  - Users can see and edit full manager definitions in one place.
  - On‑disk config remains minimal.
- Cons:
  - More complex diffing logic for managers (and potential for subtle bugs).
  - We may need to drop or re‑implement per‑field `(user-defined)` annotations.

**Verdict**: **Preferred option.** Matches the intended UX (identical show/edit data) while keeping the current minimal‑config behavior on disk.

---

## Proposed Implementation (Option C – full config view, diffed save)

### Phase 0 – Baseline & Tests (1–2h) – ✅ Completed 2025-11-16

1. Regression tests documenting current behavior before changes:
   - `internal/commands/config_edit_test.go`:
     - Ensures existing top‑level fields are preserved across the current `createTempConfigFile → parseAndValidateConfig → saveNonDefaultValues` pipeline.
   - `internal/config/user_defined_test.go`:
     - Documents that `GetNonDefaultFields` tracks only top‑level fields and explicitly **does not** include `managers` yet.
2. Docs updated to call out the current limitation:
   - Short note in `docs/configuration.md` under “Manager Configuration” stating that `config edit` does not yet expose managers and pointing to this plan file.

> This phase now locks in the current behavior with tests and documentation; subsequent phases can safely change behavior with clear before/after expectations.

### Phase 1 – Manager diffing support in UserDefinedChecker (3–4h) – ✅ Completed 2025-11-16

**Goal**: Teach `UserDefinedChecker` how to compute “non‑default manager configs” so we can save only manager diffs even when the edit view shows full config. **Completed** via implementation of `GetNonDefaultManagers` and tests.

1. Extend `UserDefinedChecker` with manager‑aware helpers:
   - Add:
     ```go
     func (c *UserDefinedChecker) GetNonDefaultManagers(cfg *Config) map[string]ManagerConfig
     ```
   - Implemented behavior:
     - `defaults := GetDefaultManagers()`.
     - For each `name, mgrCfg := range cfg.Managers`:
       - If `name` is not in `defaults`, treat as **custom** → include in result.
       - Else if `!reflect.DeepEqual(mgrCfg, defaultMgr)`, treat as an **override** → include in result.
       - Otherwise, skip (pure default; no need to persist).
     - If no managers differ from defaults, returns an empty map.

2. Tests for manager diffing in `internal/config/user_defined_test.go`:
   - No managers → `GetNonDefaultManagers` returns empty.
   - Custom manager (`custom-manager`) → returned with its binary.
   - Overridden built‑in manager (`npm` with modified `Binary`) → returned as non‑default.
   - Default built‑in manager (`brew` equal to default) → not returned.

3. `HasNonDefaultManagers` remains optional and is not yet implemented; we can add it later if we want to surface this in `config show`.

### Phase 2 – Full‑config view in config edit (4–6h) – ✅ Completed 2025-11-16

**Goal**: Make `config edit` show the same effective config as `config show`, including `managers:`, while keeping the visudo workflow. **Completed** by switching to a full-config writer and adding a matching test.

1. Replaced `writeAnnotatedConfig` with a full‑config writer:
   - Added `writeFullConfig(w *os.File, cfg *config.Config) error` in `internal/commands/config_edit.go`:
     - Uses `yaml.Marshal(cfg)` and writes the result to the temp file.
   - Updated `createTempConfigFile` to:
     - Keep the existing header comments.
     - Call `writeFullConfig(tempFile, cfg)` after the header instead of `writeAnnotatedConfig`.
   - This ensures:
     - The YAML body written by `config edit` is structurally identical to what `config show` prints for the same `cfg`.

2. `(user-defined)` annotations:
   - Inline `(user-defined)` comments are no longer emitted by `config edit`.
   - `parseAndValidateConfig` still strips legacy `(user-defined)` markers for backward compatibility.

3. Tests:
   - Added `TestCreateTempConfigFileWritesFullConfig` in `internal/commands/config_edit_test.go`:
     - Builds a config (including a `managers` block) in a temp directory.
     - Loads `cfg := config.LoadWithDefaults(configDir)` and marshals it to YAML.
     - Calls `createTempConfigFile(configDir)` and reads the temp file.
     - Strips header comment lines and asserts that the remaining YAML matches `yaml.Marshal(cfg)` exactly.
     - Verifies that `managers:` flows through into the edit view.

### Phase 3 – Save diffs (including managers) back to plonk.yaml (3–5h) – ✅ Completed 2025-11-16

**Goal**: Persist only non‑default differences (top‑level and managers) even though the edit view shows full config. **Completed** by extending `saveNonDefaultValues` and adding manager-aware round‑trip tests.

1. Updated `saveNonDefaultValues` in `internal/commands/config_edit.go`:
   - After computing:
     ```go
     nonDefaults := checker.GetNonDefaultFields(cfg)
     ```
   - Added:
     ```go
     nonDefaultManagers := checker.GetNonDefaultManagers(cfg)
     if len(nonDefaultManagers) > 0 {
         if nonDefaults == nil {
             nonDefaults = make(map[string]interface{})
         }
         nonDefaults["managers"] = nonDefaultManagers
     }
     ```
   - The resulting `nonDefaults` map (top‑level fields + manager diffs) is what gets marshaled to `plonk.yaml`.

2. Round‑trip tests:
   - Implemented `TestSaveNonDefaultValuesIncludesManagerDiffs` in `internal/commands/config_edit_test.go`:
     - Uses a temp config dir seeded with an empty `plonk.yaml` (so defaults are merged).
     - Loads `cfg := config.LoadWithDefaults(configDir)`.
     - Adds a custom manager (`custom-manager`) and overrides the shipped `npm` config (e.g., changing `Binary` to `npm-custom`).
     - Calls `saveNonDefaultValues(configDir, cfg)`.
     - Asserts that `plonk.yaml` contains a `managers:` section with entries for `custom-manager` and `npm`, but not a full dump of all defaults.
     - Reloads via `LoadWithDefaults` and asserts:
       - The custom manager is present with the expected binary.
       - The overridden `npm` manager reflects the custom binary.
       - An untouched default manager (e.g., `brew`) remains present and unchanged.

3. Backwards compatibility:
   - Existing tests (e.g., `TestConfigEditRoundTripPreservesTopLevelFields`) still pass:
     - When no managers are set or overridden, `saveNonDefaultValues` continues to omit `managers` from `plonk.yaml`, keeping configs minimal.

### Phase 4 – UX & Documentation (2–3h) – ✅ Completed 2025-11-16

**Goal**: Make manager editing discoverable and clarify the relationship between `config show` and `config edit`. **Completed** by refreshing the Manager Configuration docs and ensuring the `config edit` help text remains accurate for the full-config view.

1. Updated `docs/configuration.md`:
   - Refreshed the “Manager Configuration” section to:
     - Explain that Plonk ships a set of default manager definitions via `GetDefaultManagers` and that both `plonk config show` and `plonk config edit` display the **effective** configuration built from those defaults plus any user overrides.
     - Describe the v2 manager schema (matching `ManagerConfig`): `binary`, `list.command` with `parse`/`json_field`, nested `install` / `upgrade` / `upgrade_all` / `uninstall` blocks with `idempotent_errors`, and optional descriptive fields (`description`, `install_hint`, `help_url`).
     - Provide an updated custom manager example (using `pixi`) in the v2 schema.
     - Provide an updated override example for `npm` that matches the current defaults (JSON-based list, templated commands, idempotent errors).
   - Clarified interaction:
     - `config show` → full effective config (defaults + overrides).
     - `config edit` → same effective config, but edits are validated and only non‑default values (including manager diffs) are written back to `plonk.yaml`.

2. Help text:
   - The `config edit` long description already accurately describes the visudo-style workflow and the “full runtime configuration (defaults + your overrides)” behavior.
   - With Phase 2/3 implemented (full-config view and diffed save, including managers), this help text now implicitly applies to manager configuration as well, and no additional wording changes are required.

### Phase 5 – Dynamic manager validation for default_manager (3–5h) – ✅ Completed 2025-11-16

**Goal**: Ensure all configured managers are truly first‑class by allowing `default_manager` to reference any manager present in `cfg.Managers` (shipped or user‑defined), while keeping validation accurate. **Completed** by driving the `validmanager` set from the effective configuration in both Load and SimpleValidator paths.

Implementation summary:

1. Config-driven valid manager set:
   - Added `updateValidManagersFromConfig(cfg *Config)` in `internal/config/config.go`:
     - Builds a set of manager names from:
       - Keys of `GetDefaultManagers()`.
       - Keys of `cfg.Managers` (which already includes defaults + user managers in the Load path).
     - Calls `SetValidManagers` with this combined list so `validmanager` sees both shipped and user-defined managers as valid.
   - `LoadFromPath` now calls `updateValidManagersFromConfig(&cfg)` after merging managers but before validation:
     - This ensures `default_manager` can be any manager defined in the effective `cfg.Managers`, not just shipped defaults.

2. SimpleValidator integration:
   - Updated `SimpleValidator.ValidateConfigFromYAML` in `internal/config/compat.go`:
     - After `applyDefaults(&cfg)`, it now calls `updateValidManagersFromConfig(&cfg)` before running validation.
     - In this path, `cfg.Managers` contains user-defined managers only, so `updateValidManagersFromConfig` combines `GetDefaultManagers` and `cfg.Managers` to form the allowed set.

3. Tests:
   - `internal/config/helpers_test.go`:
     - Added a subtest to `TestValidateConfigFromYAML` verifying that:
       - A YAML snippet with `default_manager: custom-manager` and a matching `managers.custom-manager` block passes validation.
   - `internal/config/config_test.go`:
     - Added `TestLoad_CustomManagerCanBeDefault`:
       - Writes a `plonk.yaml` with `default_manager: custom-manager` and a `managers.custom-manager` definition.
       - Confirms `Load` succeeds, `cfg.DefaultManager` is `custom-manager`, and the custom manager’s binary is preserved.

With these changes, any manager present in the effective `cfg.Managers` is treated as a valid `default_manager` value, fully aligning validation behavior with the “all configured managers are first‑class” goal. The existing `validManagers` override mechanism (via `SetValidManagers`) remains available for tests and other specialized contexts.

---

## Risks & Edge Cases

- **Diffing logic mistakes**:
  - Incorrect `reflect.DeepEqual` usage could cause default managers to be written into `plonk.yaml`, bloating configs.
  - Mitigation: comprehensive tests for `GetNonDefaultManagers` and full edit/save round‑trips.

- **YAML shape changes**:
  - If we change the structure of `ManagerConfig` or `Config` fields in the future, `GetNonDefaultManagers` / `GetNonDefaultFields` may need updates to avoid spurious diffs.

- **User editing full managers block**:
  - Users might paste a full `managers:` block from `config show` into `config edit`. The diffing logic must handle this gracefully, persisting only the parts that differ from defaults (and treat the rest as defaults).

- **Annotations vs. data equality**:
  - We only guarantee that the **data** shown by `config show` and `config edit` is identical. Comments/annotations may differ between the two, and that’s acceptable as long as the YAML structures match.

---

## Summary

- The core package manager architecture is already fully config‑driven, but `plonk config edit` currently hides `managers:` and shows only a partial view of the config.
- The chosen plan (Option C) will:
  - Make `config edit` display the same effective configuration as `config show`, including `managers:`.
  - Extend `UserDefinedChecker` to understand manager diffs.
  - Teach `config edit` to save only non‑default top‑level fields and manager overrides/custom managers back to `plonk.yaml`.
- This approach satisfies the requirement that `config show` and `config edit` present identical configuration data, while preserving Plonk’s minimal on‑disk config philosophy and leaving room for future enhancements (e.g., dynamic `default_manager` validation).
