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

- `--managed` - Show only managed dotfiles (tracked and deployed)
- `--missing` - Show only missing dotfiles (tracked but not deployed)
- `--unmanaged` - Show only unmanaged dotfiles (exist but not tracked)
- `--untracked` - Alias for `--unmanaged`
- `-v, --verbose` - Show verbose output (includes untracked items)
- `-o, --output` - Output format (table/json/yaml)

## Behavior

### Resource States

Dotfiles can exist in four states:
- **managed/deployed** - Dotfile is tracked by plonk and deployed to $HOME
- **missing** - Dotfile is tracked by plonk but not deployed
- **drifted** - Dotfile is tracked and deployed but content has been modified
- **unmanaged** - Dotfile exists in $HOME but isn't tracked by plonk

### Default Display

Without flags, `dotfiles` shows all managed and missing dotfiles:
- **$HOME**: Full path in home directory (deployed location)
- **$PLONK_DIR**: Path relative to plonk configuration directory
- **STATUS**: Current state (deployed/missing/drifted)

### Filter Behavior

- `--managed` - Shows only dotfiles that are tracked and deployed
- `--missing` - Shows only dotfiles that need to be deployed
- `--unmanaged` or `--untracked` - Shows only dotfiles that exist but aren't tracked
- No flags - Shows managed and missing dotfiles (default)

### Table Format Display

**Dotfiles Table** (managed/missing):
- $HOME: Full path in home directory (deployed location)
- $PLONK_DIR: Path relative to plonk configuration directory
- STATUS: Current state with icon

**Dotfiles List** (unmanaged):
- Simple list of unmanaged dotfile paths (no source/target/status columns)

Summary counts are displayed at the end, except when using `--unmanaged` or `--missing` flags.

### Structured Output

JSON and YAML output formats include:
- Summary with counts (managed, missing, untracked)
- Items array with detailed dotfile information including source and target paths

When filtering with `--missing`, `--managed`, or `--unmanaged`, the summary reflects only the filtered state.

## Examples

```bash
# Show all managed dotfiles
plonk dotfiles

# Show only missing dotfiles
plonk dotfiles --missing

# Show only managed dotfiles
plonk dotfiles --managed

# Show only unmanaged dotfiles
plonk dotfiles --unmanaged

# Show untracked dotfiles (alias)
plonk dotfiles --untracked

# Output as JSON
plonk dotfiles -o json

# Output as YAML
plonk dotfiles -o yaml

# Short alias
plonk d --missing
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
- Summary is hidden when using `--unmanaged` or `--missing` flags in table format
- Drifted files are shown with "drifted" status in the managed list
