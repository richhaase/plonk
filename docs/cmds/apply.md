# Apply Command

Reconciles system state by installing missing packages and deploying missing dotfiles.

## Synopsis

```bash
plonk apply [options]
```

## Description

The apply command reconciles the system state with the desired configuration by installing packages listed in `plonk.lock` and deploying dotfiles from `$PLONK_DIR`. It acts like a sync operation, bringing the local environment in line with the managed configuration.

Apply targets "missing" resources (tracked but not present) and "drifted" dotfiles (modified since deployment), transitioning them to "managed" state. The command uses plonk's internal reconciliation system to identify what needs to be applied.

## Options

- `--dry-run, -n` - Preview changes without applying them
- `--packages` - Apply packages only (mutually exclusive with `--dotfiles`)
- `--dotfiles` - Apply dotfiles only (mutually exclusive with `--packages`)

## Behavior

### Core Operation

Apply performs two main operations in sequence:
1. **Package installation** - Installs all packages listed in `plonk.lock` that are not currently installed
2. **Dotfile deployment** - Deploys all dotfiles from `$PLONK_DIR` to their corresponding locations in `$HOME`

### Resource States

Plonk tracks resources in four states:
- **managed** - Resource is known to plonk and exists in the user's environment
- **missing** - Resource is known to plonk but does NOT exist in the user's environment
- **drifted** - Dotfile is known to plonk and exists but has been modified
- **unmanaged** - Resource exists in the user's environment but is not known to plonk

Apply specifically targets "missing" and "drifted" resources.

### Execution Flow

1. Execute pre-apply hooks (if configured)
2. Read plonk.lock (if exists) to determine packages to install
3. Read $PLONK_DIR contents to determine dotfiles to deploy
4. Apply packages first (if --packages or no flags)
5. Apply dotfiles second (if --dotfiles or no flags)
6. Execute post-apply hooks (if configured)
7. Report results with summary counts

### File Mapping

Dotfiles are deployed with automatic dot-prefix handling:
- `$PLONK_DIR/vimrc` → `$HOME/.vimrc`
- `$PLONK_DIR/config/nvim/init.lua` → `$HOME/.config/nvim/init.lua`

Files matching `ignore_patterns` are excluded from deployment.

### Error Handling

- Errors are reported as they occur but do not stop the apply process
- Failed resources remain in "missing" state for retry on next apply
- Package conflicts (already installed) are considered successful
- Dotfile conflicts result in overwriting the existing file
- Drifted dotfiles are backed up before restoration (with timestamp)

## Examples

```bash
# Apply all changes (packages and dotfiles)
plonk apply

# Preview what would be changed
plonk apply --dry-run

# Apply packages only
plonk apply --packages

# Apply dotfiles only
plonk apply --dotfiles
```

## Integration

- Use `plonk status` to see what resources need to be applied
- Run after `plonk clone` to set up a new system
- Run after adding packages with `plonk install` or dotfiles with `plonk add`
- Pre/post-apply hooks can be configured in `plonk.yaml`

## Notes

- The `--packages` and `--dotfiles` flags cannot be used together
- Hooks are an experimental feature and may change in future versions
