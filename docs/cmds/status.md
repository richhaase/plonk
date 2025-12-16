# Status Command

Shows managed packages and dotfiles with their current state.

## Synopsis

```bash
plonk status [options]
plonk st [options]  # Short alias
```

## Description

The `status` command provides a comprehensive view of all plonk-managed resources, including packages and dotfiles. It displays the current state of each resource (managed, missing, drifted, or unmanaged), helping users understand what's tracked, what needs attention, and what exists outside of plonk's management.

The command supports filtering by resource type and state, with multiple output formats for different use cases. It's the primary tool for understanding your current plonk configuration state.

**Note**: For focused views of specific resource types, use the dedicated commands:
- `plonk packages` (or `plonk p`) - Show only package status
- `plonk dotfiles` (or `plonk d`) - Show only dotfile status

## Options

None. The status command displays table output only.

## Behavior

### Resource States

Status displays resources in four possible states:
- **managed/deployed** - Resource is tracked by plonk and exists in the environment
- **missing** - Resource is tracked by plonk but doesn't exist in the environment
- **drifted** - Dotfile is tracked and exists but content has been modified
- **unmanaged** - Resource exists in the environment but isn't tracked by plonk

### Default Display

Without flags, status shows all managed and missing resources in two sections:
- **PACKAGES**: Shows name, manager, and status
- **DOTFILES**: Shows source (in $PLONK_DIR), target (in $HOME), and status


### Table Format Display

**Packages Table**:
- NAME: Package name
- MANAGER: Package manager (brew, npm, pnpm, cargo, gem, conda, uv, pipx)
- STATUS: Current state with icon

**Dotfiles Table** (managed/missing):
- $HOME: Full path in home directory (deployed location)
- $PLONK_DIR: Path relative to plonk configuration directory
- STATUS: Current state with icon

Summary counts are displayed at the end.

### Error Handling

- Continues operation even with invalid configuration
- Missing lock file treated as informational, not error
- Reconciliation errors are reported but don't prevent partial results

## Examples

```bash
# Show all managed resources
plonk status
plonk st                    # Short alias

# Use focused commands for specific resource types
plonk packages              # Show only package status (alias: p)
plonk dotfiles              # Show only dotfile status (alias: d)
```

## Integration

- Use before `plonk apply` to see what will be changed
- Missing items can be resolved with `plonk apply`
- Drifted dotfiles can be restored with `plonk apply` or synced back with `plonk add -y`
- Unmanaged items can be added with `plonk install` or `plonk add`
- For focused views, use `plonk packages` or `plonk dotfiles` instead

## Notes

- Alias: `st` (short form)
- Colors are applied to status words only, not full lines
- Respects NO_COLOR environment variable for accessibility
