# Package Manager Interface v2

## Overview

This document outlines the plan to extend the PackageManager interface with new methods to improve health checking, self-installation, and package upgrade capabilities.

## New Interface Methods

### CheckHealth() HealthCheck
**Purpose**: Each package manager validates its own configuration and reports health status.

**Returns**: HealthCheck struct with status, issues, and suggestions specific to that package manager.

**Replaces**: Current hardcoded PATH checking in `health.go:checkPathConfiguration()`

**Implementation Details**:
- Check if package manager binary is available
- Validate PATH configuration for package manager's bin directory
- Check permissions for package installation/removal
- Verify package manager-specific configuration files
- Test basic functionality (e.g., list command)

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

### plonk doctor
- Iterate through all available package managers
- Call `CheckHealth()` on each manager
- Aggregate results into comprehensive health report
- Remove hardcoded PATH checking logic from `health.go`

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

### Phase 1: Interface Extension
- Add new methods to PackageManager interface
- Add default implementations that return "not implemented" errors
- Update interface documentation

### Phase 2: Health Check Migration
- Implement `CheckHealth()` for all existing package managers
- Update `health.go` to use new interface method
- Remove hardcoded PATH checking logic
- Test health checks for all supported package managers

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

1. All package managers implement the new interface methods
2. `plonk doctor` provides comprehensive health checks without hardcoded logic
3. `plonk clone` automatically installs required package managers
4. `plonk upgrade` command works reliably across all supported package managers
5. `plonk status --outdated` provides useful update information
6. No regression in existing functionality
7. Comprehensive test coverage for all new features
