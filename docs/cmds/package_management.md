# Package Management Commands

Commands for managing packages: `install`, `uninstall`, `search`, and `info`.

## Description

The package management commands handle system package operations across multiple package managers. All commands support package manager prefixes (e.g., `brew:htop`) to target specific managers, defaulting to the configured `default_manager` when no prefix is specified. Package state is tracked in plonk.lock, which is updated atomically with each operation.

## Behavior

### Package Manager Prefixes

- `brew:` - Homebrew
- `npm:` - NPM (global packages)
- `cargo:` - Cargo (Rust)
- `pip:` - Pip (Python)
- `gem:` - RubyGems
- `go:` - Go modules

Without prefix, uses `default_manager` from configuration.

### Install Command

- **Purpose**: Install packages and add to plonk management
- **Flags**: `--dry-run`, `--force` (reinstall if already managed)
- **Behavior**:
  - Not installed → installs package, adds to plonk.lock
  - Already installed → adds to plonk.lock (success)
  - Already managed → skips unless --force
  - Updates plonk.lock atomically with each success

### Uninstall Command

- **Purpose**: Remove packages from system and plonk management
- **Flags**: `--dry-run`, `--force` (remove even if not managed)
- **Behavior**:
  - Removes package and plonk.lock entry
  - Dependency handling by package manager
  - Fails if not managed unless --force

### Search Command

- **Purpose**: Find packages across package managers
- **Behavior**:
  - Without prefix: searches all managers in parallel (3-second timeout)
  - With prefix: searches only specified manager
  - Shows package names only
  - Slow managers may not return results due to timeout

### Info Command

- **Purpose**: Show package details and installation status
- **Priority order**:
  1. Managed by plonk
  2. Installed but not managed
  3. Available but not installed
- **Shows**: name, status, manager, description, homepage, install command

### Cross-Command Behaviors

- All commands process multiple packages independently
- Failures don't block other operations
- Summary shows succeeded/skipped/failed counts
- Output formats: table (default), json, yaml
- plonk.lock updated atomically per operation

### State Impact

**Install Command**:
- Modifies: `plonk.lock` (adds package entry)
- System changes: Package installed via manager
- Atomic: Lock file updated only on successful install

**Uninstall Command**:
- Modifies: `plonk.lock` (removes package entry)
- System changes: Package removed via manager
- Atomic: Lock file updated only on successful uninstall

**Search/Info Commands**:
- Read-only operations
- No state modifications
- Query package managers directly

## Implementation Notes

## Improvements

- Consider making search timeout configurable
- Add verbose search mode showing descriptions and versions
- Support version pinning in install command
- Add update command to upgrade managed packages
- Show installation progress for long-running operations
- Add --all flag to uninstall all packages from a manager
- Consider showing dependencies in info output
- Add package count to search results per manager
