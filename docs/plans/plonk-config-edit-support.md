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

---

## Non‑Goals

- Do **not** redesign the overall config schema or `ManagerConfig`.
- Do **not** change lock file format or semantics.
- Do **not** turn `config edit` into an interactive wizard; it remains an editor‑based workflow on YAML.
- Optional / future: allowing `default_manager` to point at custom managers (requires improving the `validmanager` validator around dynamic manager lists).

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
  - Creates a temp file with header comments and then calls `writeAnnotatedConfig`.
  - `writeAnnotatedConfig`:
    - Writes only these fields: `default_manager`, `operation_timeout`, `package_timeout`, `dotfile_timeout`, `expand_directories`, `ignore_patterns`, `dotfiles`.
    - Never writes `managers:`.
  - `parseAndValidateConfig`:
    - Strips header comments (and `(user-defined)` annotations) then uses `config.NewSimpleValidator().ValidateConfigFromYAML`.
    - On success, unmarshals into `config.Config`, applies simple defaults to top‑level fields, and returns it.
  - `saveNonDefaultValues`:
    - Uses `config.NewUserDefinedChecker(configDir).GetNonDefaultFields(cfg)` to compute a **map of non‑default top‑level fields**.
    - Serializes that map via `yaml.Marshal` and writes it to `plonk.yaml`.
    - `GetNonDefaultFields` currently tracks only scalar/lists/dotfiles – **no `managers` support**.

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

### Phase 1 – Manager diffing support in UserDefinedChecker (3–4h)

**Goal**: Teach `UserDefinedChecker` how to compute “non‑default manager configs” so we can save only manager diffs even when the edit view shows full config.

1. Extend `UserDefinedChecker` with manager‑aware helpers:
   - Add:
     ```go
     func (c *UserDefinedChecker) GetNonDefaultManagers(cfg *Config) map[string]ManagerConfig
     ```
   - Behavior:
     - Let `defaults := GetDefaultManagers()`.
     - For each `name, mgrCfg := range cfg.Managers`:
       - If `defaultMgr, ok := defaults[name]; !ok`:
         - Treat `mgrCfg` as **custom** → include in result.
       - Else if `!reflect.DeepEqual(mgrCfg, defaultMgr)`:
         - Treat `mgrCfg` as an **override** → include in result.
       - Else:
         - Skip (pure default; no need to persist).
     - If no managers differ from defaults, return an empty map.
   - `userConfig` is not required for this logic; it is purely defaults vs current runtime config.

2. Add tests for manager diffing in `internal/config/user_defined_test.go`:
   - Case: no user config → `GetNonDefaultManagers` returns empty.
   - Case: user overrides a field for a built‑in manager (e.g., `npm.install.command`) → only `npm` appears in result with updated config.
   - Case: user adds a new custom manager (`my-custom`) → map contains only `my-custom`.

3. (Optional) Add:
   ```go
   func (c *UserDefinedChecker) HasNonDefaultManagers(cfg *Config) bool
   ```
   if we later want to highlight that in `config show`.

### Phase 2 – Full‑config view in config edit (4–6h)

**Goal**: Make `config edit` show the same effective config as `config show`, including `managers:`, while keeping the visudo workflow.

1. Replace `writeAnnotatedConfig` with a full‑config writer:
   - Introduce a new helper in `internal/commands/config_edit.go`, e.g.:
     ```go
     func writeFullConfig(w *os.File, cfg *config.Config) error {
         data, err := yaml.Marshal(cfg)
         if err != nil {
             return err
         }
         _, err = w.Write(data)
         return err
     }
     ```
   - In `createTempConfigFile`:
     - Keep the existing header comments.
     - After the header, call `writeFullConfig(tempFile, cfg)` instead of `writeAnnotatedConfig`.
   - This ensures:
     - The YAML data written by `config edit` is identical (structurally) to what `config show` prints for the same `cfg`.

2. Decide what to do with `(user-defined)` annotations:
   - Simplest path:
     - Drop inline `(user-defined)` comments for now in `config edit`.
     - Rely on `config show` + `UserDefinedChecker` in the future if we want a separate “diff view”.
   - Alternative (more work, optional later):
     - Build a small comment injector that uses `UserDefinedChecker` and a YAML AST to tag user‑defined fields inside the full config. This is not required to satisfy the primary goal.

3. Tests:
   - Add tests in `internal/commands/config_edit_test.go` to assert that:
     - The YAML body (ignoring header comments) written by `config edit` matches the output of `config show` for the same `cfg`.
     - `managers:` is present in the temp file when `cfg.Managers` is non‑empty.

### Phase 3 – Save diffs (including managers) back to plonk.yaml (3–5h)

**Goal**: Persist only non‑default differences (top‑level and managers) even though the edit view shows full config.

1. Update `saveNonDefaultValues` in `internal/commands/config_edit.go`:
   - After computing:
     ```go
     nonDefaults := checker.GetNonDefaultFields(cfg)
     ```
   - Add:
     ```go
     nonDefaultManagers := checker.GetNonDefaultManagers(cfg)
     if len(nonDefaultManagers) > 0 {
         nonDefaults["managers"] = nonDefaultManagers
     }
     ```
   - The resulting `nonDefaults` map is what gets marshaled to `plonk.yaml`.

2. Round‑trip tests:
   - Use a temp config dir and:
     - Seed a `plonk.yaml` with:
       - A custom manager (`my-custom`).
       - An overridden built‑in manager (e.g., `npm` with a modified `install` command).
     - Simulate an edit by:
       - Loading config via `LoadWithDefaults`.
       - Modifying `cfg.Managers` in memory.
       - Passing it through `saveNonDefaultValues`.
     - Reload via `LoadWithDefaults` and assert:
       - Custom manager is present and matches edits.
       - Built‑in manager override is applied on top of defaults.
       - No unexpected managers are written.

3. Backwards compatibility:
   - Test that configs without `managers:` are unchanged:
     - Start with a `plonk.yaml` containing only top‑level fields.
     - Run through the new `parseAndValidateConfig` + `saveNonDefaultValues` path without touching managers.
     - Verify `plonk.yaml` is unchanged (modulo benign formatting/ordering).

### Phase 4 – UX & Documentation (2–3h)

**Goal**: Make manager editing discoverable and clarify the relationship between `config show` and `config edit`.

1. Update `docs/configuration.md`:
   - Add a “Manager configuration” section that:
     - Explains that built‑in manager defaults are defined in code and that both `plonk config show` and `plonk config edit` display the **effective** manager configuration.
     - Shows a minimal example of overriding a built‑in manager:
       ```yaml
       managers:
         npm:
           install:
             command: ["npm", "install", "-g", "{{.Package}}", "--legacy-peer-deps"]
       ```
     - Shows an example of adding a custom manager.
   - Clarify interaction:
     - `config show` → full effective config (merged defaults + overrides).
     - `config edit` → same effective config, but changes are validated and only non‑default values are written back to `plonk.yaml`.

2. Update help text:
   - Extend the `config edit` long description to mention:
     - Manager overrides are supported via the `managers:` section.
     - For the full schema and examples, see `docs/configuration.md`.

3. (Optional) Add a small note to the `config show` header:
   - E.g.: “Note: Not all fields shown here may be present in `plonk.yaml`; defaults are applied at runtime. Use `plonk config edit` to change them.”

### Phase 5 – Optional: Improve validmanager for custom defaults (future)

**Problem**: The `validmanager` validator for `DefaultManager` currently relies on a global `validManagers` slice, populated once in `internal/resources/packages/init.go` from built‑ins. That means:

- Users can define custom managers in `managers:`, but cannot reliably set them as `default_manager` without making `validmanager` aware of those new names.

**Possible approach** (not required for initial feature):

1. Allow `validmanager` to consider runtime config managers:
   - Modify `validatePackageManager` to:
     - Prefer `validManagers` when explicitly set (test harness).
     - Otherwise, derive allowed names from:
       - Keys of `GetDefaultManagers()`.
       - Plus keys from `cfg.Managers` when validating a config instance (this may require threading a `Config` instance into validation or separating validation modes).

2. Alternatively, update `SetValidManagers` at startup based on merged config:
   - After loading config in the CLI entrypoint, call `SetValidManagers` with the manager names from the active `ManagerRegistry`.
   - This needs careful design to avoid init‑time cycles (`config` ↔ `packages`) and to keep tests predictable.

For now, this can be documented as a limitation: users can override built‑ins and add custom managers, but `default_manager` should remain one of the built‑in names until validation is made dynamic.

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
