# Config Command

Manages plonk configuration through subcommands.

## Synopsis

```bash
plonk config <subcommand> [options]
```

## Description

The config command provides access to plonk's configuration system through two subcommands: `show` and `edit`. Plonk uses a zero-configuration approach with sensible defaults, allowing users to override specific values through a YAML configuration file.

Configuration is stored at `$PLONK_DIR/plonk.yaml` (default: `~/.config/plonk/plonk.yaml`). The system merges user overrides with built-in defaults to provide the effective runtime configuration.

## Subcommands

### config show

Displays the effective runtime configuration.

```bash
plonk config show [options]
```

**Options:**
- `-o, --output` - Output format (table/json/yaml)

**Behavior:**
- Shows all configuration values (defaults merged with user overrides)
- Table format displays user-defined values with blue "(user-defined)" annotations
- JSON/YAML formats remain clean for scripting compatibility
- Shows configuration file path at the top of output

### config edit

Opens the configuration file in an editor with visudo-style validation.

```bash
plonk config edit
```

**Behavior:**
- Creates config directory if missing
- Opens temporary file with full runtime configuration
- User-defined values marked with `# (user-defined)` comments
- Validates configuration after editing
- On validation failure, offers options to:
  - (e)dit again to fix errors
  - (r)evert changes and exit
  - (q)uit without saving
- On success, saves only non-default values to `plonk.yaml`
- Changes take effect immediately (no restart required)

**Editor Selection:**
1. `$VISUAL` environment variable
2. `$EDITOR` environment variable
3. Falls back to `vim`

## Configuration Structure

### Available Settings

- **Package Management:**

  - `default_manager` - Default package manager (brew, npm, pnpm, cargo, gem, conda, uv, pipx)
  - `package_timeout` - Timeout for package operations (seconds)
  - `operation_timeout` - Timeout for search operations (seconds)

- **Dotfile Management:**
  - `dotfile_timeout` - Timeout for dotfile operations (seconds)
  - `expand_directories` - Directories to scan for dotfiles
  - `ignore_patterns` - Patterns to exclude from dotfile operations

- **Tools:**
  - `diff_tool` - Tool for showing dotfile differences

### Minimal Configuration Philosophy

When you run `plonk config edit`, you see all available options with their current values, but only settings that differ from defaults are saved to `plonk.yaml`. This keeps your configuration file clean and makes it clear what you've actually customized.

## Examples

```bash
# View current configuration
plonk config show

# View configuration as JSON
plonk config show -o json

# Edit configuration with validation
plonk config edit
```

## Integration

- Configuration affects all plonk commands
- `PLONK_DIR` environment variable can override default location
- Invalid configuration falls back to defaults gracefully
- See [Configuration Guide](../configuration.md) for detailed examples

## Notes

- Configuration precedence: Environment variables > User config > Defaults
- Missing configuration file is valid (uses all defaults)
- Validation includes field constraints (e.g., timeout ranges)
- All configuration fields are optional
