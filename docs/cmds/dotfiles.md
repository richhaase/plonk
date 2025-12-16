# Dotfiles Command

Display the status of all plonk-managed dotfiles.

## Synopsis

```bash
plonk dotfiles [options]
plonk d [options]  # Short alias
```

## Description

The `dotfiles` command provides a focused view of dotfile status, independent of packages. It shows which dotfiles are managed, missing, drifted, or unmanaged, helping you understand the dotfile state of your system.

This command is useful when you want to focus specifically on dotfiles without the additional context of packages. For a combined view of packages and dotfiles, use `plonk status`.

## Options

None. The dotfiles command displays table output only.

## Behavior

### Resource States

Dotfiles can exist in four states:
- **managed/deployed** - Dotfile is tracked by plonk and deployed to $HOME
- **missing** - Dotfile is tracked by plonk but not deployed
- **drifted** - Dotfile is tracked and deployed but content has been modified
- **unmanaged** - Dotfile exists in $HOME but isn't tracked by plonk

### Default Display

The `dotfiles` command shows all managed and missing dotfiles:
- **$HOME**: Full path in home directory (deployed location)
- **$PLONK_DIR**: Path relative to plonk configuration directory
- **STATUS**: Current state (deployed/missing/drifted)

### Table Format Display

**Dotfiles Table**:
- $HOME: Full path in home directory (deployed location)
- $PLONK_DIR: Path relative to plonk configuration directory
- STATUS: Current state with icon

Summary counts are displayed at the end.

## Examples

```bash
# Show all managed dotfiles
plonk dotfiles
plonk d                     # Short alias
```

## Integration

- Use `plonk status` for combined package and dotfile status
- Missing dotfiles can be deployed with `plonk apply`
- Unmanaged dotfiles can be added with `plonk add`
- Drifted dotfiles can be restored with `plonk apply` or synced back with `plonk add -y`
- Use `plonk diff` to see differences for drifted dotfiles

## Notes

- Alias: `d` (short form)
- Colors are applied to status words only
- Respects NO_COLOR environment variable for accessibility
- Drifted files are shown with "drifted" status in the managed list
