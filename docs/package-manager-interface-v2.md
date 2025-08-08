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
- ✅ Validate PATH configuration for discovered directories
- ✅ Generate shell-specific export commands for PATH fixes
- ✅ Professional status reporting (pass/warn/fail) with actionable feedback

### SelfInstall() error
**Purpose**: Package manager installs itself when needed during environment setup.

**Usage**: Called by `plonk clone` only for package managers that have packages in the cloned plonk.lock file.

**Implementation Details**:
- Each package manager knows how to install itself
- May use system package managers (e.g., brew installs npm via `brew install node`)
- Should be idempotent - safe to call if already installed
- Should validate successful installation

### Upgrade(ctx context.Context, packages []string) error
**Purpose**: Upgrade one or more packages to their latest versions.

**Parameters**:
- `packages`: List of package names to upgrade. Empty slice means upgrade all installed packages.

**Implementation Details**:
- Handle repository updates internally if needed (e.g., `brew update` before `brew upgrade`)
- Support both single package and bulk upgrade operations
- Return meaningful errors for failed upgrades
- Update package versions in plonk.lock via existing mechanisms

### Outdated(ctx context.Context) ([]PackageUpdate, error)
**Purpose**: List packages that have newer versions available.

**Returns**: Slice of PackageUpdate structs containing current and available versions.

**Usage**: Called by `plonk status --outdated` to show update information.

**PackageUpdate struct**:
```go
type PackageUpdate struct {
    Name           string
    CurrentVersion string
    LatestVersion  string
    Manager        string
}
```

## Command Integration

### plonk doctor ✅ IMPLEMENTED
- ✅ Iterates through all 10 available package managers
- ✅ Calls `CheckHealth()` on each manager via `checkPackageManagerHealth()`
- ✅ Aggregates results into comprehensive health report
- ✅ Removed hardcoded PATH checking logic from `health.go`
- ✅ Added overall ecosystem health assessment with `calculateOverallPackageManagerHealth()`

### plonk clone
- Parse cloned plonk.lock file
- Identify which package managers are needed (have managed packages)
- Call `SelfInstall()` only on required package managers
- Proceed with existing package installation logic

### plonk upgrade (NEW COMMAND)
**Syntax**:
- `plonk upgrade` - Upgrade all outdated packages across all managers
- `plonk upgrade [manager:]package` - Upgrade specific package(s)
- `plonk upgrade [manager]:` - Upgrade all packages for specific manager

**Behavior**:
- Check for outdated packages using `Outdated()`
- Call `Upgrade()` on relevant package managers
- Update plonk.lock with new versions
- Provide progress feedback and error handling

### plonk status --outdated
**Behavior**:
- Existing status output plus outdated package information
- Call `Outdated()` on all managers with installed packages
- Display packages with available updates
- Performance consideration: Only call when flag is explicitly used

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

### Phase 3: Self-Installation
- Implement `SelfInstall()` for all package managers
- Update `clone` command to use new interface method
- Test environment setup scenarios

### Phase 4: Upgrade Functionality
- Implement `Outdated()` and `Upgrade()` for all package managers
- Create new `upgrade` command
- Update `status` command with `--outdated` flag
- Add comprehensive testing for upgrade scenarios

## Implementation Considerations

### Error Handling
- Consistent error types across all package managers
- Graceful degradation when package managers are unavailable
- Clear error messages for user-facing operations

### Performance
- `Outdated()` calls should be efficient and cancellable
- Bulk operations should be optimized where possible
- Consider caching for expensive operations

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
3. ⏳ **Self-Installation**: `plonk clone` automatically installs required package managers
4. ⏳ **Upgrade Command**: `plonk upgrade` command works reliably across all supported package managers
5. ⏳ **Outdated Status**: `plonk status --outdated` provides useful update information
6. ✅ **No Regression**: No regression in existing functionality - all existing commands work
7. ✅ **Test Coverage**: Comprehensive test coverage maintained for CheckHealth functionality

**Phase 2 Complete**: CheckHealth system fully implemented and tested. Ready for Phase 3 (Self-Installation).
