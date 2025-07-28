# Apply Command

The `plonk apply` command installs missing packages and deploys missing dotfiles.

## Description

The apply command reconciles the system state with the desired configuration by installing packages listed in `plonk.lock` and deploying dotfiles from `$PLONK_DIR`. It acts like a sync operation, bringing the local environment in line with the managed configuration. The command uses plonk's internal reconciliation system to identify missing resources and applies them using the appropriate resource managers.

## Behavior

### Core Operation

Apply performs two main operations in sequence:
1. **Package installation** - Installs all packages listed in `plonk.lock` that are not currently installed
2. **Dotfile deployment** - Deploys all dotfiles from `$PLONK_DIR` to their corresponding locations in `$HOME`

### Resource States

Plonk tracks resources in three states:
- **managed** - Resource is known to plonk and exists in the user's environment
- **missing** - Resource is known to plonk but does NOT exist in the user's environment
- **unmanaged** - Resource exists in the user's environment but is not known to plonk

Apply specifically targets "missing" resources and attempts to transition them to "managed" state.

### Command Options

- `--dry-run, -n` - Preview changes without applying them
- `--packages` - Apply packages only
- `--dotfiles` - Apply dotfiles only
- `--backup` - Create backups before overwriting existing dotfiles (feature under review)

Note: Using `--packages` and `--dotfiles` together is redundant and equivalent to running with no flags.

### Execution Flow

```
1. Read plonk.lock (if exists) → determine packages to install
2. Read $PLONK_DIR contents → determine dotfiles to deploy
3. If --packages or no flags:
   - Query package managers for installed packages
   - Install missing packages via appropriate managers
4. If --dotfiles or no flags:
   - Check existing dotfiles in $HOME
   - Deploy missing/changed dotfiles from $PLONK_DIR
5. Report results
```

### Dry Run Behavior

With `--dry-run`, apply shows what would be changed without making modifications:

```
Plonk Apply (Dry Run)
=====================

Dotfiles:
  → .config/nvim/lua/plugins/disabled.lua (would deploy)

Summary:
--------
📦 Packages: 0 would be installed
📄 Dotfiles: 1 would be deployed
```

### Success Output

A successful apply operation reports results for both packages and dotfiles:

```
Plonk Apply
===========

Dotfiles:
  ✓ .config/nvim/lua/plugins/disabled.lua

Summary:
--------
📦 Packages: All up to date
📄 Dotfiles: 1 deployed, 0 failed

Total: 1 succeeded, 0 failed
```

### Error Handling

- Errors are reported as they occur but do not stop the apply process
- Failed resources remain in "missing" state for retry on next apply
- Package conflicts (already installed) are considered successful
- Dotfile conflicts result in overwriting the existing file

### Special Cases

- **No plonk.lock**: Valid scenario, no packages will be installed (dotfiles-only mode)
- **Empty $PLONK_DIR**: No dotfiles to deploy
- **Partial apply**: Using `--packages` or `--dotfiles` limits operation and output to specified resources

### File Mapping

Dotfiles are deployed with automatic dot-prefix handling:
- `$PLONK_DIR/vimrc` → `$HOME/.vimrc`
- `$PLONK_DIR/config/nvim/init.lua` → `$HOME/.config/nvim/init.lua`

Files matching `ignore_patterns` in configuration are excluded from deployment.

## Implementation Notes

## Improvements

- Review and properly document the `--backup` flag behavior
- Consider adding progress indicators for large apply operations
- Add verbose mode for detailed operation logging
