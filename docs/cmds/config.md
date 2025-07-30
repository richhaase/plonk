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

User-defined values are highlighted with blue "(user-defined)" annotations in table format, making it easy to see which settings you've customized. JSON and YAML output formats remain clean without annotations for scripting compatibility.

### Edit Command Behavior

Works like visudo - opens a temporary file with the full runtime configuration for editing:

- Shows all configuration values (defaults merged with user overrides)
- User-defined values are marked with `# (user-defined)` annotations
- Uses system editor (determined by `$VISUAL`, then `$EDITOR`, then fallback to vim)
- Validates configuration after editing
- If validation fails, offers options to (e)dit again, (r)evert changes, or (q)uit
- On successful save, writes only non-default values to `plonk.yaml`
- Changes take effect immediately (no restart required)

### Configuration Precedence

1. Built-in defaults (hardcoded)
2. User overrides from `plonk.yaml`

Configuration validation occurs during editing. Invalid entries are reported with specific error messages, and users can choose to fix errors or cancel the edit operation.

### Default Configuration Values

The system provides defaults for:
- Package manager settings (`default_manager`, timeout values)
- Directory handling (`expand_directories`)
- File filtering (`ignore_patterns`, `dotfiles.unmanaged_filters`)
- Hook configuration (pre/post apply hooks)

User overrides in `plonk.yaml` only need to specify values that differ from defaults.

### Minimal Configuration Philosophy

Plonk follows a minimal configuration approach. When you run `plonk config edit`, you see all available options with their current values, but only settings that differ from defaults are saved to `plonk.yaml`. This keeps your configuration file clean and makes it clear what you've actually customized.

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

3. **Edit Command (Visudo-style):**
   - Creates config directory if missing
   - Generates temporary file with full runtime configuration
   - Marks user-defined values with `# (user-defined)` annotations
   - Validates edited configuration using `SimpleValidator`
   - Provides edit/revert/quit loop on validation failures
   - Editor selection: Checks `$VISUAL`, then `$EDITOR`, then defaults to vim
   - Saves only non-default values to maintain minimal config files

4. **Validation:**
   - Uses `go-playground/validator` for struct validation
   - Edit command actively validates and reports errors with detailed messages
   - Validation includes field constraints (e.g., timeout ranges)
   - Returns structured `ValidationResult` with errors and warnings

**Configuration Structure:**
- Defaults defined in `defaultConfig` variable
- Extensive default `ignore_patterns` and `unmanaged_filters`
- YAML unmarshaling overlays user config on defaults
- All fields are optional with `omitempty` tags

**Bugs Identified:**
None - all discrepancies have been resolved.
