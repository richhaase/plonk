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

## Improvements

- Make `plonk config edit` work with a complete config file at all times so users can easily see all the options, then on save it would only write the non-default values to plonk.yaml
- Highlight user-defined values in `plonk config show` output to distinguish them from defaults
