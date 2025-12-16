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

None. The packages command displays table output only.

## Behavior

### Resource States

Packages can exist in three states:
- **managed** - Package is tracked by plonk and installed on the system
- **missing** - Package is tracked by plonk but not installed
- **unmanaged** - Package is installed but not tracked by plonk

### Default Display

The `packages` command shows all managed and missing packages grouped by manager:
- **NAME**: Package name
- **MANAGER**: Package manager (brew, npm, pnpm, cargo, gem, conda, uv, pipx)
- **STATUS**: Current state with icon

### Table Format Display

**Packages Table**:
- NAME: Package name
- MANAGER: Package manager
- STATUS: Current state (managed/missing)

Summary counts are displayed at the end.

## Examples

```bash
# Show all managed packages
plonk packages
plonk p                     # Short alias
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
