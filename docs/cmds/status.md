# Status Command

Shows managed packages and dotfiles with their current state.

## Synopsis

```bash
plonk status [options]
```

## Description

The status command provides a comprehensive view of all plonk-managed resources, including packages and dotfiles. It displays the current state of each resource (managed, missing, drifted, or unmanaged), helping users understand what's tracked, what needs attention, and what exists outside of plonk's management.

The command supports filtering by resource type and state, with multiple output formats for different use cases. It's the primary tool for understanding your current plonk configuration state.

## Options

- `--packages` - Show only package status
- `--dotfiles` - Show only dotfile status
- `--unmanaged` - Show only unmanaged items
- `--missing` - Show only missing resources (mutually exclusive with --unmanaged)
- `--outdated` - Show packages with available updates (implies --packages)
- `-o, --output` - Output format (table/json/yaml)

## Behavior

### Resource States

Status displays resources in four possible states:
- **managed/deployed** - Resource is tracked by plonk and exists in the environment
- **missing** - Resource is tracked by plonk but doesn't exist in the environment
- **drifted** - Dotfile is tracked and exists but content has been modified
- **unmanaged** - Resource exists in the environment but isn't tracked by plonk

### Package Update States (--outdated flag)

When using the `--outdated` flag, packages also show update information:
- **current** - Package is at the latest available version
- **outdated** - Package has a newer version available for upgrade
- **unknown** - Version information could not be determined

### Default Display

Without flags, status shows all managed and missing resources in two sections:
- **PACKAGES**: Shows name, manager, and status
- **DOTFILES**: Shows source (in $PLONK_DIR), target (in $HOME), and status

Always displays summary counts at the end.

### Filter Combinations

Filters can be combined:
- `--packages --unmanaged` - Show only unmanaged packages
- `--dotfiles --unmanaged` - Show only unmanaged dotfiles
- `--packages --missing` - Show only missing packages
- `--packages --outdated` - Show only packages with available updates

Using both `--packages` and `--dotfiles` together has the same effect as using neither.

**Note**: The `--outdated` flag automatically implies `--packages` and is mutually exclusive with `--dotfiles`, `--unmanaged`, and `--missing`.

### Table Format Display

**Packages Table**:
- NAME: Package name
- MANAGER: Package manager (brew, npm, cargo, pipx, gem, go, uv, pixi, composer, dotnet)
- STATUS: Current state with icon

**Packages Table with --outdated**:
- NAME: Package name
- MANAGER: Package manager
- CURRENT: Currently installed version
- LATEST: Latest available version
- UPDATE: Update status (current/outdated/unknown)

**Dotfiles Table** (managed/missing):
- SOURCE: Path relative to `$PLONK_DIR`
- TARGET: Full path in `$HOME`
- STATUS: Current state with icon

**Dotfiles List** (unmanaged):
- Simple list of unmanaged dotfile paths (no source/target/status columns)

### Structured Output

JSON and YAML output formats include:
- Configuration file paths and validity
- Detailed summary with counts by domain
- Managed items in flat array with domain field
- When using `--outdated`, includes version and update information

Note: Structured output formats (JSON/YAML) do not support the `--unmanaged` flag and will always show managed items only. Use table format to view unmanaged items.

### Error Handling

- Continues operation even with invalid configuration
- Missing lock file treated as informational, not error
- Reconciliation errors are reported but don't prevent partial results

## Examples

```bash
# Show all managed resources
plonk status

# Show only packages
plonk status --packages

# Show only dotfiles
plonk status --dotfiles

# Show unmanaged items
plonk status --unmanaged

# Show missing resources
plonk status --missing

# Show missing packages only
plonk status --packages --missing

# Show packages with available updates
plonk status --outdated

# Show outdated packages as JSON
plonk status --outdated -o json

# Output as JSON
plonk status -o json
```

## Integration

- Use before `plonk apply` to see what will be changed
- Missing items can be resolved with `plonk apply`
- Drifted dotfiles can be restored with `plonk apply` or updated with `plonk add`
- Unmanaged items can be added with `plonk install` or `plonk add`
- Use `plonk status --outdated` before `plonk upgrade` to preview available updates
- Outdated packages can be upgraded with `plonk upgrade`

## Notes

- Alias: `st` (short form)
- Colors are applied to status words only, not full lines
- Respects NO_COLOR environment variable for accessibility
- Summary is hidden when using `--unmanaged` flag in table format
