# Apply Command

The `plonk apply` command installs missing packages and deploys missing dotfiles.

For CLI syntax and flags, see [CLI Reference](../cli.md#plonk-apply).

## Description

The apply command reconciles the system state with the desired configuration by installing packages listed in `plonk.lock` and deploying dotfiles from `$PLONK_DIR`. It acts like a sync operation, bringing the local environment in line with the managed configuration. The command uses plonk's internal reconciliation system to identify missing resources and applies them using the appropriate resource managers.

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

Apply specifically targets "missing" and "drifted" resources and attempts to transition them to "managed" state. Drifted dotfiles are restored from their source in `$PLONK_DIR`.

### Command Options

- `--dry-run, -n` - Preview changes without applying them
- `--packages` - Apply packages only (mutually exclusive with `--dotfiles`)
- `--dotfiles` - Apply dotfiles only (mutually exclusive with `--packages`)

Note: The `--packages` and `--dotfiles` flags are mutually exclusive - you cannot use both together.

### Execution Flow

1. Execute pre-apply hooks (if configured) - **experimental feature**
2. Read plonk.lock (if exists) to determine packages to install
3. Read $PLONK_DIR contents to determine dotfiles to deploy
4. Apply packages first (if --packages or no flags)
   - Updates plonk.lock for each successful package installation
5. Apply dotfiles second (if --dotfiles or no flags)
6. Execute post-apply hooks (if configured) - **experimental feature**
7. Report results with summary counts

### Dry Run Behavior

With `--dry-run`, apply shows what would be changed without making modifications. Output includes a summary of packages and dotfiles that would be affected.

### Output

Apply reports results for both packages and dotfiles, showing successful operations and any failures. Always displays summary counts at the end.

### Error Handling

- Errors are reported as they occur but do not stop the apply process
- Failed resources remain in "missing" state for retry on next apply
- Package conflicts (already installed) are considered successful
- Dotfile conflicts result in overwriting the existing file
- Drifted dotfiles are backed up before restoration (with timestamp)

### Special Cases

- **No plonk.lock**: Valid scenario, no packages will be installed (dotfiles-only mode)
- **Empty $PLONK_DIR**: No dotfiles to deploy
- **Partial apply**: Using `--packages` or `--dotfiles` limits operation and output to specified resources

### File Mapping

Dotfiles are deployed with automatic dot-prefix handling:
- `$PLONK_DIR/vimrc` → `$HOME/.vimrc`
- `$PLONK_DIR/config/nvim/init.lua` → `$HOME/.config/nvim/init.lua`

Files matching `ignore_patterns` are excluded from deployment. For configuration details, see [Configuration Guide](../configuration.md#ignore-patterns).

## Implementation Notes

The apply command orchestrates package installation and dotfile deployment through a layered architecture:

**Command Structure:**
- Entry point: `internal/commands/apply.go`
- Orchestration: `internal/orchestrator/orchestrator.go` and `apply.go`
- Resource management: `internal/resources/packages/` and `internal/resources/dotfiles/`

**Key Implementation Flow:**

1. **Command Processing:**
   - Parses flags: `--dry-run`, `--backup`, `--packages`, `--dotfiles`
   - Flags `--packages` and `--dotfiles` are mutually exclusive via `applyCmd.MarkFlagsMutuallyExclusive()`
   - Creates orchestrator with functional options pattern
   - Calls orchestrator.Apply() and converts result for output

2. **Orchestration Layer:**
   - Runs pre-apply hooks if configured
   - Conditionally applies packages (unless `--dotfiles` only)
   - Conditionally applies dotfiles (unless `--packages` only)
   - Runs post-apply hooks if configured
   - Aggregates results and error handling

3. **Package Apply Flow:**
   - Uses `packages.Reconcile()` to identify missing packages from lock file
   - Groups missing packages by manager type
   - Applies packages through `MultiPackageResource`
   - Lock file updates occur during package installation operations as documented

4. **Dotfile Apply Flow:**
   - Uses `manager.GetConfiguredDotfiles()` to scan `$PLONK_DIR`
   - Creates dotfile resource and sets desired state
   - Uses `resources.ReconcileResource()` to find missing deployments
   - Performs file operations (copy with dot-prefix transformation)

**Backup Implementation:**
- **Removed**: Previously had `--backup` flag for dotfile backups, but this functionality has been removed
- Creates backup files before overwriting existing dotfiles
- Backup naming follows pattern: `{original}.backup.{timestamp}`

**Execution Order:**
- Pre-apply hooks → Packages → Dotfiles → Post-apply hooks
- Hook execution (pre-apply and post-apply) is integrated into the apply flow as an experimental feature

**Error Handling:**
- Partial failure support: continues operation despite individual failures
- Aggregates errors by type (package vs dotfile)
- Post-apply hook failures are non-fatal
- Returns non-zero exit code if any operation failed

**Output Structure:**
- Complex result types with detailed operation status
- Table output shows per-item results with status icons
- Summary section with counts and overall status
- **DISCREPANCY**: More detailed output than documented

**Bugs Identified:**
None - all discrepancies have been resolved.

## Improvements

- Consider adding progress indicators for large apply operations
- Add verbose mode for detailed operation logging
- Add support for selective dotfile deployment based on patterns
