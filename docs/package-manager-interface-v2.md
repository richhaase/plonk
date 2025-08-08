# Package Manager Interface v2

## Overview

This document outlines the plan to extend the PackageManager interface with new methods to improve health checking, self-installation, and package upgrade capabilities.

## New Interface Methods

### CheckHealth() HealthCheck ✅ IMPLEMENTED
**Purpose**: Each package manager validates its own configuration and reports health status.

**Returns**: HealthCheck struct with status, issues, and suggestions specific to that package manager.

**Replaced**: Hardcoded PATH checking in `health.go:checkPathConfiguration()` - now uses dynamic discovery

**Implementation Details** (All Complete):
- ✅ Check if package manager binary is available via `IsAvailable()`
- ✅ Dynamically discover bin directories using package manager-specific commands:
  - Homebrew: `brew --prefix` + `/bin`
  - NPM: `npm config get prefix` + `/bin`
  - Pip: `python3 -m site --user-base` + `/bin`
  - Cargo: `CARGO_HOME` detection + `~/.cargo/bin`
  - Go: `go env GOBIN`/`go env GOPATH` + `/bin`
  - Gem: `gem environment` parsing
  - UV: `uv tool dir` discovery
  - Pixi: `pixi global bin` discovery
  - Composer: `composer global config bin-dir --absolute`
  - .NET: `~/.dotnet/tools` standard path
  - Pipx: `pipx environment` parsing for PIPX_BIN_DIR
- ✅ Validate PATH configuration for discovered directories
- ✅ Generate shell-specific export commands for PATH fixes
- ✅ Professional status reporting (pass/warn/fail) with actionable feedback

### SelfInstall() error ✅ IMPLEMENTED
**Purpose**: Package manager installs itself when needed during environment setup.

**Usage**: Called by `plonk clone` only for package managers that have packages in the cloned plonk.lock file.

**Implementation Details** (All Complete):
- ✅ Each package manager knows how to install itself using official installation methods
- ✅ Three-tier classification system:
  - Tier 1 (Independent): Homebrew, Cargo, Go, UV, Pixi - direct installation via curl | sh
  - Tier 2 (Runtime-dependent): Composer, Pip, Pipx - secure installation via runtime commands
  - Tier 3 (Package manager dependent): NPM, Gem, .NET - delegation to other package managers
- ✅ Idempotent operations - safe to call multiple times if already installed
- ✅ HTTPS verification and official source validation for security
- ✅ Context cancellation support with proper error handling
- ✅ No interactive prompting - fully automated installation process

### Upgrade(ctx context.Context, packages []string) error ✅ IMPLEMENTED
**Purpose**: Upgrade one or more packages to their latest versions.

**Parameters**:
- `packages`: List of package names to upgrade. Empty slice means upgrade all installed packages.

**Implementation Details** (All Complete):
- ✅ Handle repository updates internally if needed (e.g., `brew update` before `brew upgrade`)
- ✅ Support both single package and bulk upgrade operations
- ✅ Return meaningful errors for failed upgrades with proper error categorization
- ✅ Update package versions in plonk.lock via existing lock file mechanisms
- ✅ Comprehensive error handling for "not installed", permission, and network failures
- ✅ Individual package upgrade support with batch processing for efficiency

### ~~Outdated(ctx context.Context) ([]PackageUpdate, error)~~ ❌ REMOVED
**Decision**: This interface method has been removed from the scope as it provides limited value relative to implementation complexity.

**Reasons for removal**:
- Network-dependent operations would slow down `plonk status` significantly
- Information becomes stale quickly as packages are published frequently
- Not immediately actionable - users still need separate upgrade workflow
- Complex maintenance burden across 10 different package managers
- Users can run native package manager commands (`brew outdated`, `npm outdated`, etc.) when needed

**Alternative approach**: Focus implementation effort on robust `Upgrade()` functionality that provides direct user value.

## Command Integration

### plonk doctor ✅ IMPLEMENTED
- ✅ Iterates through all 10 available package managers
- ✅ Calls `CheckHealth()` on each manager via `checkPackageManagerHealth()`
- ✅ Aggregates results into comprehensive health report
- ✅ Removed hardcoded PATH checking logic from `health.go`
- ✅ Added overall ecosystem health assessment with `calculateOverallPackageManagerHealth()`

### plonk clone ✅ IMPLEMENTED
- ✅ Parse cloned plonk.lock file via `DetectRequiredManagers()`
- ✅ Identify which package managers are needed (have managed packages)
- ✅ Call `SelfInstall()` only on required package managers via `installDetectedManagers()`
- ✅ Proceed with existing package installation logic
- ✅ Removed all interactive prompting functionality (deleted `prompts.go`)
- ✅ Fully automated setup process integrated with package manager registry

### plonk upgrade (NEW COMMAND) ✅ IMPLEMENTED
**Syntax**:
- `plonk upgrade` - Upgrade all installed packages across all managers
- `plonk upgrade [manager:]package` - Upgrade specific package(s)
- `plonk upgrade [manager]:` - Upgrade all packages for specific manager

**Behavior** (All Complete):
- ✅ Call `Upgrade()` on relevant package managers for specified packages
- ✅ Update plonk.lock with new versions using lockfile integration
- ✅ Provide progress feedback and error handling with colored output
- ✅ For bulk upgrades, delegate to native package manager upgrade commands
- ✅ Support for Go package matching by both binary name and source path
- ✅ Comprehensive error reporting and graceful failure handling

### ~~plonk status --outdated~~ ❌ REMOVED
**Decision**: The `--outdated` flag for status command has been removed as it would depend on the removed `Outdated()` interface method.

**Alternative**: Users can check for outdated packages using native package manager commands when needed.

## Migration Plan

### Phase 1: Interface Extension ✅ COMPLETED
- ✅ Added `CheckHealth(ctx context.Context) (*HealthCheck, error)` to PackageManager interface
- ✅ Added HealthCheck struct with comprehensive status reporting fields
- ✅ Created helper functions in `health_helpers.go` and `path_helpers.go`
- ✅ Updated interface documentation

### Phase 2: Health Check Migration ✅ COMPLETED
- ✅ Implemented `CheckHealth()` for all 10 package managers
- ✅ Updated `health.go` to use new interface method via `checkPackageManagerHealth()`
- ✅ Removed hardcoded PATH checking logic (300+ lines of obsolete code)
- ✅ Tested health checks for all supported package managers via `plonk doctor`
- ✅ Added dynamic PATH discovery using package manager-specific commands
- ✅ Implemented shell-specific configuration suggestions (zsh, bash, fish)

### Phase 3: Self-Installation ✅ COMPLETED
- ✅ Implemented `SelfInstall()` for all 10 package managers using official installation methods
- ✅ Updated `clone` command to use new interface method via `installDetectedManagers()`
- ✅ Created helper functions in `install_helpers.go` for secure installation
- ✅ Removed obsolete installation functions from `tools.go` (replaced with SelfInstall interface)
- ✅ Deleted interactive prompting system (`prompts.go`) per project requirements
- ✅ Tested environment setup scenarios with comprehensive unit and integration tests

### Phase 4: Upgrade Functionality ✅ COMPLETED
- ✅ Implemented `Upgrade()` for all 10 package managers with error handling
- ✅ Created new `upgrade` command with comprehensive argument parsing
- ✅ Added extensive testing including unit tests and BATS integration tests
- ✅ Focused on robust upgrade workflows with lockfile integration
- ✅ Added upgrade output formatting with progress indicators and colored feedback
- ✅ Removed pip package manager entirely (deprecated Python 2 tool)

## Implementation Considerations

### Error Handling
- Consistent error types across all package managers
- Graceful degradation when package managers are unavailable
- Clear error messages for user-facing operations

### Performance
- `Upgrade()` operations should be efficient and cancellable
- Bulk operations should be optimized where possible
- Delegate to native package manager bulk upgrade commands when available

### Testing
- Unit tests for all new interface methods
- Integration tests for command behavior
- BATS tests for end-to-end scenarios

### Backward Compatibility
- Existing commands should continue to work during migration
- Gradual rollout of new functionality
- Clear migration path for existing installations

## Success Criteria

1. ✅ **CheckHealth Implementation**: All 10 package managers implement CheckHealth() method
2. ✅ **Dynamic Health Checks**: `plonk doctor` provides comprehensive health checks without hardcoded logic
3. ✅ **Self-Installation**: `plonk clone` automatically installs required package managers
4. ⏳ **Upgrade Command**: `plonk upgrade` command works reliably across all supported package managers
5. ✅ **No Regression**: No regression in existing functionality - all existing commands work
6. ✅ **Test Coverage**: Comprehensive test coverage maintained for CheckHealth and SelfInstall functionality

**Phase 3 Complete**: SelfInstall system fully implemented and tested. Ready for Phase 4 (Upgrade Functionality).

**Scope Refinement**: Removed `Outdated()` interface method and `--outdated` flag from scope to focus implementation effort on more valuable upgrade functionality that provides direct user benefit.

**Recent Addition**: Added pipx package manager support as an alternative to pip for Python application management. pipx provides isolated environments for Python CLI applications, making it safer than pip for installing global tools.
