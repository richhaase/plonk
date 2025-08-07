# plonk upgrade

Upgrade packages to their latest versions across supported package managers.

## Syntax

```bash
# Upgrade all outdated packages across all managers
plonk upgrade

# Upgrade all packages for a specific package manager
plonk upgrade [manager]:

# Upgrade specific packages
plonk upgrade [manager:]package1 [manager:]package2 ...

# Upgrade with options
plonk upgrade --dry-run
plonk upgrade --format json
```

## Description

The `plonk upgrade` command identifies and upgrades packages that have newer versions available. It works with all supported package managers and updates the plonk.lock file to reflect new versions.

The command operates in three phases:
1. **Discovery**: Identify outdated packages using each package manager's `Outdated()` method
2. **Upgrade**: Execute upgrades using each package manager's `Upgrade()` method
3. **Reconciliation**: Update plonk.lock with new package versions

## Examples

### Upgrade All Packages
```bash
plonk upgrade
```
Checks all installed packages across all package managers and upgrades any that have newer versions available.

### Upgrade Specific Package Manager
```bash
plonk upgrade brew:
plonk upgrade npm:
plonk upgrade pip:
```
Upgrades all packages managed by the specified package manager.

### Upgrade Specific Packages
```bash
plonk upgrade brew:neovim
plonk upgrade npm:typescript brew:ripgrep
plonk upgrade pip:requests
```
Upgrades only the specified packages.

### Dry Run
```bash
plonk upgrade --dry-run
```
Shows what packages would be upgraded without actually performing the upgrades.

## Flags

### Global Flags
- `--dry-run`: Show what would be upgraded without executing changes
- `--format`: Output format (table, json, yaml)
- `--verbose`: Show detailed upgrade information
- `--quiet`: Suppress non-error output

## Behavior

### Package Manager Integration
Each package manager handles upgrades according to its own capabilities:

- **Homebrew**: Runs `brew update` internally if needed, then `brew upgrade`
- **NPM**: Upgrades global packages via `npm update -g`
- **Pip**: Upgrades packages via `pip install --upgrade`
- **Cargo**: Updates via `cargo install` (reinstalls latest version)
- **Go**: Reinstalls packages with `go install package@latest`
- **Gem**: Upgrades via `gem update`
- **UV**: Upgrades tools via `uv tool upgrade`
- **Pixi**: Upgrades global environments via `pixi global upgrade`
- **Composer**: Upgrades global packages via `composer global update`
- **dotnet**: Upgrades tools via `dotnet tool update -g`

### Lock File Updates
After successful upgrades, plonk automatically updates the lock file with:
- New version numbers
- Updated installation timestamps
- Any metadata changes

### Error Handling
- Individual package upgrade failures do not stop the entire operation
- Failed upgrades are reported with specific error messages
- Partial success scenarios are handled gracefully
- Exit codes reflect overall operation success/failure

### Progress Indication
The command provides progress feedback during:
- Discovery phase (checking for outdated packages)
- Upgrade execution (per-package progress)
- Lock file reconciliation

## Output Formats

### Table Format (Default)
```
PACKAGE MANAGER    PACKAGE         FROM      TO        STATUS
brew              ripgrep         14.1.0    14.1.1    upgraded
npm               typescript      5.3.3     5.4.2     upgraded
pip               requests        2.31.0    2.32.0    failed
```

### JSON Format
```json
{
  "upgrades": [
    {
      "manager": "brew",
      "package": "ripgrep",
      "from_version": "14.1.0",
      "to_version": "14.1.1",
      "status": "upgraded"
    }
  ],
  "summary": {
    "total": 3,
    "upgraded": 2,
    "failed": 1
  }
}
```

## Exit Codes

- `0`: All requested upgrades completed successfully
- `1`: Some upgrades failed (partial success)
- `2`: Command failed to execute (no upgrades attempted)

## Integration with Other Commands

### Relationship to plonk status
Use `plonk status --outdated` to preview what packages would be upgraded before running `plonk upgrade`.

### Relationship to plonk doctor
Run `plonk doctor` if upgrade operations are failing to check for package manager configuration issues.

### Relationship to plonk apply
The `plonk apply` command focuses on achieving desired state from lock files, while `plonk upgrade` focuses on updating existing packages to latest versions.

## Performance Considerations

- Checking for outdated packages can be slow for some package managers
- Bulk upgrade operations are used when supported by the package manager
- Operations are cancellable and respect context timeouts
- Network-dependent operations may require retry logic

## Safety Features

- Dry-run capability for preview operations
- Atomic lock file updates (all or nothing)
- Backup creation before destructive operations
- Clear rollback instructions for failed upgrades

## Limitations

- Cannot downgrade packages (use specific install commands)
- Some package managers may require manual intervention for major version changes
- Network connectivity required for most upgrade operations
- Package manager-specific constraints apply (e.g., dependency conflicts)
