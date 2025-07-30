# Package Management Commands

Commands for managing packages: `install`, `uninstall`, `search`, and `info`.

For CLI syntax and flags, see [CLI Reference](../cli.md#package-manager-prefixes).

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

**Purpose**: Install packages and add to plonk management

**Behavior**:
  - Not installed → installs package, adds to plonk.lock
  - Already installed → adds to plonk.lock (success)
  - Already managed → skips (no reinstall)
  - Updates plonk.lock atomically with each success

### Uninstall Command

**Purpose**: Remove packages from system and plonk management

**Behavior**:
  - Removes package and plonk.lock entry
  - Dependency handling by package manager
  - Only removes packages that are currently managed by plonk

### Search Command

**Purpose**: Find packages across package managers

**Behavior**:
  - Without prefix: searches all managers in parallel (configurable timeout)
  - With prefix: searches only specified manager
  - Shows package names only
  - Slow managers may not return results due to timeout
  - Timeout controlled by `operation_timeout` configuration (default: 5 minutes)

### Info Command

**Purpose**: Show package details and installation status

**Priority order**:
  1. Managed by plonk
  2. Installed but not managed
  3. Available but not installed

**Information displayed**: name, status, manager, description, homepage, install command

### Timeout Configuration

Package management operations have configurable timeouts. For complete timeout configuration examples and details, see [Configuration Guide](../CONFIGURATION.md#timeout-configuration).

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

The package management commands provide unified package operations across multiple package managers through a registry-based system:

**Command Structure:**
- Entry points: `internal/commands/install.go`, `internal/commands/uninstall.go`, `internal/commands/search.go`, `internal/commands/info.go`
- Core operations: `internal/resources/packages/operations.go`
- Manager registry: `internal/resources/packages/registry.go`
- Lock file management: `internal/lock/yaml_lock.go`

**Key Implementation Flow:**

1. **Install Command Processing:**
   - Entry point parses package specifications with `ParsePackageSpec()`
   - Uses configurable timeout per package installation (from `package_timeout` config)
   - Validates manager and checks availability before installation
   - Updates `plonk.lock` atomically after successful installation

2. **Uninstall Command Processing:**
   - Uses configurable timeout per package removal (from `package_timeout` config)
   - Automatically determines manager from lock file if no prefix specified
   - Only removes from lock file if package was actually managed

3. **Search Command Implementation:**
   - Uses configurable timeout for parallel manager search (from `operation_timeout` config)
   - Includes all available package managers in search operations
   - Parallel goroutines with per-manager configurable timeouts

4. **Info Command Priority Logic:**
   - Implements exact priority: managed → installed → available
   - Uses `lock.FindPackage()` to check managed status first
   - Falls back to checking installation across all available managers
   - Returns first available package if not installed anywhere

5. **Package Manager Registry:**
   - Central registry pattern for all 6 supported managers
   - `IsAvailable()` checks performed before all operations
   - Manager instances created per-operation (not cached)
   - Supports dynamic manager discovery and validation

6. **Lock File State Management:**
   - Uses `YAMLLockService` for atomic updates
   - **Known Limitation**: Go packages store binary name only, not full module path
   - Implementation: `ExtractBinaryNameFromPath()` converts `golang.org/x/tools/cmd/gopls` → `gopls`
   - This design choice prioritizes binary management but loses reinstallation information
   - Updates happen immediately after successful package operations
   - **Future Enhancement**: v2 lock format could store both binary name and source path in metadata

7. **Error Handling and Timeouts:**
   - Install operations: Configurable timeout per package (default: 3 minutes)
   - Uninstall operations: Configurable timeout per package (default: 3 minutes)
   - Search operations: Configurable timeout total and per-manager (default: 5 minutes)
   - Info operations: No timeout (blocking)
   - Failed operations don't prevent other packages from processing

8. **Validation and Fallbacks:**
   - Empty package names rejected with validation errors
   - Invalid managers rejected with helpful error messages
   - Default manager fallback: configuration → `DefaultManager` constant
   - Manager availability checked before each operation

**Architecture Patterns:**
- Registry pattern for package manager abstraction
- Individual package processing with independent error handling
- Atomic lock file updates with rollback on failure
- Context-based timeout management with cancellation
- Standardized operation result format across all commands

**Error Conditions:**
- Manager unavailability results in operation failure
- Package not found errors are command-specific
- Lock file corruption prevents operations
- Network timeouts are handled gracefully in search

**Integration Points:**
- Uses `config.LoadWithDefaults()` for consistent configuration loading
- Shares `resources.OperationResult` format with dotfile commands
- Lock file service used by status command for reconciliation
- Manager registry used by doctor command for availability checks

**Bugs Identified:**
None - all discrepancies have been resolved.

## Improvements

- **Enhance lock file format**: Store both binary name and full source path for Go packages and npm scoped packages
- Add verbose search mode showing descriptions and versions
- Support version pinning in install command
- Add update command to upgrade managed packages
- Show installation progress for long-running operations
- Add --all flag to uninstall all packages from a manager
- Consider showing dependencies in info output
