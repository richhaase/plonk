# Config Command

The `plonk config` command manages plonk configuration.

## Description

The config command provides access to plonk's configuration system through two subcommands: `show` and `edit`. Plonk uses a zero-configuration approach with sensible defaults, allowing users to override specific values through a YAML configuration file stored at `$PLONK_DIR/plonk.yaml` (default: `~/.config/plonk/plonk.yaml`).

## Behavior

### Subcommands

- **`plonk config show`** - Display the effective runtime configuration
- **`plonk config edit`** - Open the configuration file in an editor

### Configuration Storage

- Configuration file location: `$PLONK_DIR/plonk.yaml`
- Default `$PLONK_DIR`: `~/.config/plonk`
- Environment variable override: `PLONK_DIR` can be set to use alternate location

### Show Command Behavior

Displays all configuration values (both defaults and user overrides):

- Shows configuration file path at the top
- Maintains nested structure for complex values
- Supports multiple output formats via `-o/--output` flag:
  - `table` (default): Human-readable YAML-like format
  - `json`: Structured JSON with PascalCase field names
  - `yaml`: Standard YAML format with snake_case field names

Output structure differences:
- Table format: Simple, indented display
- JSON format: Wrapped in object with `config_path` and `config` fields
- YAML format: Same structure as JSON but in YAML syntax

### Edit Command Behavior

Opens the configuration file for editing:

- Uses system editor (determined by `$EDITOR` environment variable)
- If `plonk.yaml` doesn't exist, creates it with default values
- No validation on save - invalid YAML entries are silently ignored
- Changes take effect immediately (no restart required)

### Configuration Precedence

1. Built-in defaults (hardcoded)
2. User overrides from `plonk.yaml`

Invalid configuration entries in `plonk.yaml` are ignored and defaults are used instead. No warnings or error messages are displayed for invalid entries.

### Default Configuration Values

The system provides defaults for:
- Package manager settings (`default_manager`, timeout values)
- Directory handling (`expand_directories`)
- File filtering (`ignore_patterns`, `dotfiles.unmanaged_filters`)
- Hook configuration (pre/post apply hooks)

User overrides in `plonk.yaml` only need to specify values that differ from defaults.

## Implementation Notes

The config command is implemented as a parent command with two subcommands:

**Command Structure:**
- Parent command: `internal/commands/config.go` - Simple command group
- Show subcommand: `internal/commands/config_show.go`
- Edit subcommand: `internal/commands/config_edit.go`

**Configuration Management:**
- Core configuration: `internal/config/config.go`
- Compatibility layer: `internal/config/compat.go`
- Environment variable support: `PLONK_DIR` overrides default location

**Key Implementation Details:**

1. **Config Loading Flow:**
   - `GetDefaultConfigDirectory()` checks `PLONK_DIR` env var first
   - Falls back to `~/.config/plonk` if not set
   - `LoadWithDefaults()` provides zero-config behavior
   - Missing config files return defaults silently

2. **Show Command:**
   - Loads merged configuration (defaults + user overrides)
   - Supports three output formats via `RenderOutput()`
   - Table format displays YAML representation
   - JSON/YAML formats wrap config in structured output

3. **Edit Command:**
   - Creates config directory if missing
   - **DISCREPANCY**: Creates a template config file with example content, not actual defaults
   - Validates configuration after editing using `SimpleValidator`
   - **DISCREPANCY**: Shows validation errors and allows re-editing (documented as "silently ignored")
   - Editor selection: `$EDITOR` → `$VISUAL` → `nano` → `vi`

4. **Validation:**
   - Uses `go-playground/validator` for struct validation
   - **DISCREPANCY**: Edit command actively validates and reports errors
   - Validation includes field constraints (e.g., timeout ranges)
   - Returns structured `ValidationResult` with errors and warnings

**Configuration Structure:**
- Defaults defined in `defaultConfig` variable
- Extensive default `ignore_patterns` and `unmanaged_filters`
- YAML unmarshaling overlays user config on defaults
- All fields are optional with `omitempty` tags

**Bugs Identified:**
1. Documentation states invalid entries are "silently ignored" but edit command shows validation errors
2. Edit command creates template config, not file with actual default values
3. Documentation doesn't mention `$VISUAL` environment variable check

## Improvements

- Make `plonk config edit` work with a complete config file at all times so users can easily see all the options, then on save it would only write the non-default values to plonk.yaml
- Highlight user-defined values in `plonk config show` output to distinguish them from defaults
