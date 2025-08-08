# plonk upgrade

Upgrade packages to their latest versions across supported package managers.

## Syntax

```bash
# Upgrade all packages across all managers
plonk upgrade

# Upgrade all packages for a specific package manager
plonk upgrade [manager]

# Upgrade specific packages
plonk upgrade [manager:]package1 [manager:]package2 ...
plonk upgrade package1 package2 ...
```

## Description

The `plonk upgrade` command upgrades packages across supported package managers and updates the plonk.lock file to reflect new versions.

The command operates in two phases:
1. **Upgrade**: Execute upgrades using each package manager's `Upgrade()` method
2. **Reconciliation**: Update plonk.lock with new package versions

## Examples

### Upgrade All Packages
```bash
plonk upgrade
```
Upgrades all packages managed by plonk across all package managers.

### Upgrade Specific Package Manager
```bash
plonk upgrade brew
plonk upgrade npm
plonk upgrade uv
```
Upgrades all packages managed by the specified package manager.

### Upgrade Specific Packages
```bash

# Upgrade specific packages from specific managers
plonk upgrade brew:neovim
plonk upgrade npm:typescript brew:ripgrep
plonk upgrade uv:httpx

# Upgrade packages by name across all managers that have them
plonk upgrade ripgrep neovim
plonk upgrade typescript requests
```
When a package manager is specified (e.g., `brew:neovim`), only that specific package is upgraded.
When no manager is specified (e.g., `ripgrep`), all packages with that name across all managers are upgraded.

## Flags

### Global Flags
- `--format`: Output format (table, json, yaml)
- `--verbose`: Show detailed upgrade information
- `--quiet`: Suppress non-error output

## Behavior

### Package Manager Integration
Each package manager handles upgrades according to its own capabilities:

- **Homebrew**: Runs `brew update` internally if needed, then `brew upgrade`
- **NPM**: Upgrades global packages via `npm update -g`
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
uv                httpx           0.26.0    0.27.0    failed
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

### Relationship to plonk doctor
Run `plonk doctor` if upgrade operations are failing to check for package manager configuration issues.

### Relationship to plonk apply
The `plonk apply` command focuses on achieving desired state from lock files, while `plonk upgrade` focuses on updating existing packages to latest versions.

## Performance Considerations

- Bulk upgrade operations are used when supported by the package manager
- Operations are cancellable and respect context timeouts
- Network-dependent operations may require retry logic

## Safety Features

- Atomic lock file updates (all or nothing)
- Safety is delegated to the underlying package managers

## Limitations

- Cannot downgrade packages (use specific install commands)
- Some package managers may require manual intervention for major version changes
- Network connectivity required for most upgrade operations
- Package manager-specific constraints apply (e.g., dependency conflicts)
