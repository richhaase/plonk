# Release Notes – Config-Driven Managers & Config UX

**Scope**: Summary of behavior changes and new UX for manager configuration and `plonk config show/edit` in the `cleanup/config-driven-package-managers` branch.

## Manager Configuration

- Manager definitions are now fully config-driven via `ManagerConfig`:
  - Defaults are provided by `GetDefaultManagers` (brew, npm, pnpm, cargo, pipx, conda, gem, uv).
  - Users can override built-ins or add new managers under `managers:` in `plonk.yaml`.
- Manager overrides are merged **field-wise**:
  - For each manager, the effective config is `defaults + overrides`:
    - Non-empty override fields (`binary`, `list`, `install`, `uninstall`, etc.) replace defaults.
    - Omitted fields inherit their default values.
  - This allows partial overrides (e.g., only changing `list.command` or `install.command`) without copying the entire default definition.
- Limitations:
  - Field-wise merge treats empty/zero override fields as “inherit defaults”.
  - There is currently no way to explicitly clear default values from YAML (e.g., to remove a default `idempotent_errors` entry or metadata extractor).

## default_manager Behavior

- `default_manager` now supports **any manager present in the effective config**, including user-defined managers:
  - Validation of `default_manager` is driven by the union of:
    - Shipped defaults from `GetDefaultManagers`.
    - Managers defined under `managers:` in `plonk.yaml`.
- Examples:
  - Valid:
    ```yaml
    default_manager: custom-manager
    managers:
      custom-manager:
        binary: "custom"
    ```
  - Invalid:
    ```yaml
    default_manager: unknown-manager  # not present in defaults or managers:
    ```

## plonk config edit

- `plonk config edit` now:
  - Loads the full effective config via `config.LoadWithDefaults`.
  - Writes a temp file with header comments plus the full config (including the merged `managers:` map).
  - Opens the temp file in `$VISUAL`/`$EDITOR`/`vim`.
  - Validates the edited YAML with `SimpleValidator` (same rules as normal load).
  - On success, saves **only non-default values** back to `plonk.yaml`:
    - Top-level diffs via `GetNonDefaultFields`.
    - Manager diffs via `GetNonDefaultManagers` (user-defined managers + overridden built-ins).
- Result:
  - `plonk.yaml` stays minimal.
  - Users can see and edit the full, effective manager configuration (including shipped defaults).

## plonk config show

- Table output (`plonk config show` with default `--output table`) now highlights customization:
  - Top-level fields that differ from defaults (e.g., `default_manager`, `operation_timeout`) are highlighted in blue.
  - Manager entries under `managers:` are highlighted in blue when the manager is user-defined or differs from its default config.
  - List fields:
    - `expand_directories` / `ignore_patterns`:
      - Added items (present in current config but not defaults) are highlighted in green.
      - Removed default items are shown as red comments:
        ```yaml
        # removed: - .config
        ```
- JSON/YAML outputs (`--output json` / `--output yaml`) remain uncolored and structurally unchanged.
- Color handling:
  - Honors `NO_COLOR` and terminal detection via `output.InitColors`.
  - When colors are disabled or output is piped, highlighting falls back to plain text.

## Manager-Agnostic Core

- All manager-specific logic has been moved into configuration and metadata pipelines:
  - npm/pnpm list handling is JSON-based and driven by `ManagerConfig.List`.
  - npm scoped metadata (`scope`, `full_name`) is produced via `MetadataExtractors` and applied generically.
  - Upgrade targeting for scoped packages is driven by `ManagerConfig.UpgradeTarget` (e.g., `full_name_preferred` for npm) instead of hard-coded `manager == "npm"` checks.
- Go is no longer treated as a built-in manager:
  - Any Go-related behavior must come from user-defined manager entries in `plonk.yaml`.

## Notes for Reviewers

- Global validation state:
  - The set of valid managers used by the `validmanager` validator is now updated from the effective config via `updateValidManagersFromConfig`.
  - This is appropriate for a CLI that loads a single config, but remains a mutable global; if Plonk ever becomes a long-lived multi-config process, we may prefer instance-based validators.
- List semantics:
  - `ignore_patterns`, `expand_directories`, and `dotfiles.unmanaged_filters` remain **full overrides** when specified in `plonk.yaml`.
  - We intentionally did not introduce additive semantics to avoid surprising changes in which files are ignored or managed.
