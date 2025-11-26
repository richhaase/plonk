# Packages Command

Display the status of all plonk-managed packages.

## Synopsis

```bash
plonk packages [options]
plonk p [options]  # Short alias
```

## Description

The `packages` command provides a focused view of package status, independent of dotfiles. It shows which packages are managed, missing, or unmanaged, helping you understand the package state of your system.

This command is useful when you want to focus specifically on packages without the additional context of dotfiles. For a combined view of packages and dotfiles, use `plonk status`.

## Options

- `--missing` - Show only missing packages (packages tracked but not installed)
- `--unmanaged` - Show only unmanaged packages (packages installed but not tracked)
- `-o, --output` - Output format (table/json/yaml)

## Behavior

### Resource States

Packages can exist in three states:
- **managed** - Package is tracked by plonk and installed on the system
- **missing** - Package is tracked by plonk but not installed
- **unmanaged** - Package is installed but not tracked by plonk

### Default Display

Without flags, `packages` shows all managed and missing packages grouped by manager:
- **NAME**: Package name
- **MANAGER**: Package manager (brew, npm, pnpm, cargo, gem, conda, uv, pipx)
- **STATUS**: Current state with icon

### Filter Behavior

- `--missing` - Shows only packages that need to be installed
- `--unmanaged` - Shows only packages that exist but aren't tracked
- No flags - Shows managed and missing packages (default)

### Table Format Display

**Packages Table**:
- NAME: Package name
- MANAGER: Package manager
- STATUS: Current state (managed/missing/untracked)

Summary counts are displayed at the end, except when using `--unmanaged` or `--missing` flags.

### Structured Output

JSON and YAML output formats include:
- Summary with counts (managed, missing, untracked)
- Items array with detailed package information

When filtering with `--missing` or `--unmanaged`, the summary reflects only the filtered state.

## Examples

```bash
# Show all managed packages
plonk packages

# Show only missing packages
plonk packages --missing

# Show only unmanaged packages
plonk packages --unmanaged

# Output as JSON
plonk packages -o json

# Output as YAML
plonk packages -o yaml

# Short alias
plonk p --missing
```

## Integration

- Use `plonk status` for combined package and dotfile status
- Missing packages can be installed with `plonk apply`
- Unmanaged packages can be added with `plonk install`
- Use `plonk upgrade` to upgrade managed packages

## Notes

- Alias: `p` (short form)
- Colors are applied to status words only
- Respects NO_COLOR environment variable for accessibility
- Summary is hidden when using `--unmanaged` or `--missing` flags in table format
