# Plonk Lock File Implementation

## Overview

This document tracks the implementation of separating state tracking from configuration in Plonk. The goal is to move all package management state from `plonk.yaml` (user configuration) to `plonk.lock` (managed state).

## Design Principles

1. **Clean Separation**: `plonk.yaml` contains only settings, `plonk.lock` contains all package state
2. **No Backwards Compatibility**: Clean break from current implementation
3. **Implicit Management**: Presence in lock file means managed, absence means untracked
4. **Minimal Lock File**: Only essential information stored

## File Structure

### plonk.yaml (User Configuration)
```yaml
settings:
  default_manager: homebrew
  expand_directories:
    - .config
    - .ssh
    - .aws
  
ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"
```

### plonk.lock (Managed State)
```yaml
version: 1
packages:
  homebrew:
    - name: git
      installed_at: "2024-01-15T10:30:00Z"
      version: "2.43.0"
    - name: neovim
      installed_at: "2024-01-15T10:31:00Z"
      version: "0.9.5"
  npm:
    - name: typescript
      installed_at: "2024-01-15T10:32:00Z"
      version: "5.3.3"
  cargo:
    - name: ripgrep
      installed_at: "2024-01-15T10:35:00Z"
      version: "14.1.0"
```

## State Logic

- **Managed**: Package exists in lock file AND is installed on system
- **Missing**: Package exists in lock file BUT not installed on system
- **Untracked**: Package is installed on system BUT not in lock file

## Implementation Plan

### Phase 1: Core Lock File Infrastructure
- [x] Create feature branch `feat/lock-file`
- [x] Design lock file data structures
- [x] Create lock file interfaces (`LockReader`, `LockWriter`)
- [x] Implement YAML-based lock file service
- [x] Add atomic file operations for lock file
- [x] Create lock file adapter for PackageConfigLoader
- [x] Add comprehensive tests for lock service

### Phase 2: Config Refactoring (COMPLETED ✅)
- [x] Remove package fields from `Config` struct (homebrew, npm, cargo)
- [x] Update `Settings` struct to only contain configuration settings
- [x] Remove package-related methods from `ConfigService` and `ConfigAdapter`
- [x] Update config validation to exclude package validation
- [x] Update config tests to reflect package-free structure
- [x] Ensure `plonk config show` only displays settings (not packages)
- [x] Temporarily disable pkg add/remove commands with clear error messages

### Phase 3: Command Updates (COMPLETED ✅)
- [x] Update `pkg list` to read from lock file (use existing adapter)
- [x] Re-enable `pkg add` command with lock file integration
- [x] Re-enable `pkg remove` command with lock file integration
- [x] Update `status` command to read from lock file
- [x] Add lock file path to `doctor` command checks
- [x] Ensure proper error handling for lock file operations
- [x] Add cargo package manager support across all commands
- [ ] Update command tests to use lock file (deferred to Phase 5)
- [ ] Update `apply` command to use lock file for reconciliation (moved to Phase 4)

### Phase 4: Apply Command and Reconciliation Updates
- [ ] Update `apply` command to use lock file for package reconciliation
- [ ] Update reconciler initialization in apply command to use `LockFileAdapter`
- [ ] Test full apply workflow with lock file-based package management
- [ ] Verify package reconciliation flow (install missing, remove untracked)
- [ ] Update apply command tests to use lock file
- [ ] Add error handling for lock file operations during apply
- [ ] Test apply command with mixed package states (managed/missing/untracked)

### Phase 5: Testing and Documentation
- [ ] Update all package-related tests
- [ ] Add lock file service tests
- [ ] Update integration tests
- [ ] Update CLI documentation
- [ ] Update configuration documentation
- [ ] Update architecture documentation
- [ ] Update examples in README

## Technical Details

### Lock File Service Interface
```go
type LockFile struct {
    Version  int                          `yaml:"version"`
    Packages map[string][]PackageEntry   `yaml:"packages"`
}

type PackageEntry struct {
    Name        string    `yaml:"name"`
    InstalledAt time.Time `yaml:"installed_at"`
    Version     string    `yaml:"version"`
}

type LockService interface {
    Load() (*LockFile, error)
    Save(lock *LockFile) error
    AddPackage(manager, name, version string) error
    RemovePackage(manager, name string) error
    GetPackages(manager string) ([]PackageEntry, error)
    HasPackage(manager, name string) bool
}
```

### File Locations
- Lock file: `~/.config/plonk/plonk.lock` (or `$PLONK_DIR/plonk.lock`)
- Config file: `~/.config/plonk/plonk.yaml` (unchanged location)

## Testing Strategy

1. Unit tests for lock file service
2. Integration tests for command updates
3. End-to-end tests for complete workflows
4. Migration testing (even though no backwards compatibility)

## Rollout Plan

Since we're not maintaining backwards compatibility:
1. Single release with complete implementation
2. Clear release notes about breaking changes
3. Update all documentation before release

## Phase 2 Detailed Plan

### Step 1: Remove Package Fields from Config Struct
- Update `internal/config/yaml_config.go` Config struct
- Remove `Homebrew`, `NPM`, `Cargo` fields
- Update YAML tags and struct comments

### Step 2: Update Config Service Methods
- Remove `GetPackagesForManager()` from `ConfigAdapter`
- Remove package-related validation from `SimpleValidator`
- Update `LoadConfig()` to not process package sections
- Update `SaveConfig()` to not save package sections

### Step 3: Update Config Commands
- Modify `config show` to only display settings and ignore patterns
- Update `config validate` to exclude package validation
- Ensure `config edit` still works with package-free config

### Step 4: Update Config Tests
- Remove package-related test cases from config tests
- Update test fixtures to use package-free config
- Verify config loading/saving works without packages

### Step 5: Update Example Configs
- Update README examples to show package-free config
- Update test fixtures and mock configs

### Expected Breaking Changes
- Existing `plonk.yaml` files with package sections will need migration
- Commands that rely on config for package info will temporarily break
- This is acceptable since we're not maintaining backwards compatibility

### Success Criteria
- Config struct has no package fields
- Config service only handles settings and ignore patterns
- All config tests pass
- `plonk config show` displays clean, package-free output

## Phase 3 Detailed Plan

### Overview
Phase 3 focuses on re-enabling package commands to work with the lock file instead of the config file. The lock file infrastructure is already in place, so we need to integrate it with the command layer.

### Current State
- Package commands (`pkg add`, `pkg remove`) are temporarily disabled
- Lock file service (`YAMLLockService`) is implemented and tested
- Lock file adapter (`LockFileAdapter`) bridges lock service to existing interfaces
- Config no longer contains package information

### Step 1: Re-enable `pkg add` Command
**File**: `internal/commands/pkg_add.go`
**Goal**: Replace temporary error message with lock file integration

**Implementation**:
1. Remove the temporary error return
2. Initialize `YAMLLockService` with default config directory
3. For named package addition:
   - Add package to lock file using `AddPackage(manager, name, version)`
   - Install package using existing manager implementation
   - Handle installation errors (remove from lock file if install fails)
4. For bulk addition (`pkg add` with no args):
   - Use existing package discovery logic
   - Add untracked packages to lock file
5. Add proper error handling for lock file operations

**Key Considerations**:
- Use atomic operations (add to lock file, then install)
- Rollback lock file changes if installation fails
- Maintain existing command flags and behavior
- Use existing package manager integrations

### Step 2: Re-enable `pkg remove` Command  
**File**: `internal/commands/pkg_remove.go`
**Goal**: Replace temporary error message with lock file integration

**Implementation**:
1. Remove the temporary error return
2. Initialize `YAMLLockService` with default config directory
3. For package removal:
   - Remove package from lock file using `RemovePackage(manager, name)`
   - If `--uninstall` flag is set, also uninstall from system
   - Handle uninstall errors gracefully
4. Add proper error handling for lock file operations

**Key Considerations**:
- Lock file is always updated (even without `--uninstall`)
- Uninstall errors should not prevent lock file updates
- Maintain existing command flags and behavior

### Step 3: Update `pkg list` Command
**File**: `internal/commands/pkg_list.go`
**Goal**: Use lock file adapter instead of config adapter

**Implementation**:
1. Replace `ConfigAdapter` with `LockFileAdapter` 
2. Update package listing to read from lock file
3. Maintain existing output format and filtering
4. Add lock file status to output (managed packages)

### Step 4: Update `apply` Command
**File**: `internal/commands/apply.go`
**Goal**: Use lock file for package reconciliation

**Implementation**:
1. Update reconciler initialization to use `LockFileAdapter`
2. Ensure state reconciliation works with lock file
3. Test full apply workflow with lock file

### Step 5: Update `status` Command
**File**: `internal/commands/status.go`
**Goal**: Show package status from lock file

**Implementation**:
1. Update status display to read from lock file
2. Show managed vs untracked packages
3. Display lock file location and status

### Step 6: Update `doctor` Command
**File**: `internal/commands/doctor.go`
**Goal**: Check lock file health

**Implementation**:
1. Add lock file path to system checks
2. Verify lock file is readable/writable
3. Check lock file format validity
4. Update package count to read from lock file

### Step 7: Update Command Tests
**Files**: `internal/commands/*_test.go`
**Goal**: Test commands with lock file

**Implementation**:
1. Update test fixtures to use lock file
2. Create test helpers for lock file setup
3. Test error scenarios (corrupted lock file, etc.)
4. Test command interactions with lock file

### Success Criteria for Phase 3
- `pkg add` and `pkg remove` commands work with lock file
- `pkg list` shows packages from lock file
- `apply` command reconciles using lock file
- `status` command shows lock file-based package status
- `doctor` command validates lock file health
- All command tests pass with lock file integration
- No references to config file for package management remain

### Risk Mitigation
- Comprehensive testing of lock file operations
- Proper error handling for file I/O operations
- Atomic operations to prevent corrupted state
- Clear error messages for user debugging

## Phase 4 Detailed Plan

### Overview
Phase 4 focuses on updating the `apply` command to use the lock file for package reconciliation. This is the final piece needed for complete lock file integration. The apply command performs bulk operations (install missing packages, optionally remove untracked packages) and needs to work seamlessly with the lock file.

### Current State After Phase 3
- All individual package commands (`pkg add`, `pkg remove`, `pkg list`) work with lock file
- Lock file infrastructure is robust and well-tested
- Status and doctor commands provide visibility into lock file state
- Configuration is completely separated from package management

### Step 1: Update Apply Command Package Provider
**File**: `internal/commands/apply.go`
**Goal**: Replace config-based package provider with lock file-based provider

**Current Implementation Analysis**:
The apply command likely uses a similar pattern to status command:
```go
packageProvider, err := createPackageProvider(ctx, cfg)
```

**Required Changes**:
1. Update `createPackageProvider` call to use `configDir` instead of `cfg`
2. Ensure the package provider uses `LockFileAdapter` 
3. Update reconciler to use lock file for package state

**Key Considerations**:
- Apply command may have its own `createPackageProvider` function or reuse one from another command
- Need to ensure cargo manager support is included
- Maintain existing apply command behavior and flags

### Step 2: Test Apply Command Workflow
**Goal**: Verify that apply command works correctly with lock file-based package management

**Test Scenarios**:
1. **Missing Packages**: Packages in lock file but not installed on system
   - Apply should install these packages
   - Apply should update lock file timestamps if needed
2. **Untracked Packages**: Packages installed on system but not in lock file
   - Apply should report these (and optionally add them with `--add-untracked`)
3. **Managed Packages**: Packages both in lock file and installed
   - Apply should leave these alone
4. **Mixed State**: Combination of managed, missing, and untracked packages
   - Apply should handle all scenarios correctly

### Step 3: Error Handling and Edge Cases
**Goal**: Ensure robust error handling during apply operations

**Scenarios to Handle**:
1. **Lock File Corruption**: Apply should detect and report lock file issues
2. **Installation Failures**: If package installation fails, don't update lock file
3. **Partial Failures**: Some packages install successfully, others fail
4. **Permission Issues**: Handle lock file write permission errors
5. **Manager Unavailability**: Gracefully handle when package managers aren't available

### Step 4: Update Apply Command Tests
**Files**: `internal/commands/apply_test.go` (if exists)
**Goal**: Update tests to work with lock file instead of config

**Required Changes**:
1. Replace config fixtures with lock file fixtures
2. Test apply command with various lock file states
3. Test error scenarios (corrupted lock file, etc.)
4. Verify apply command respects lock file state correctly

### Integration Points
The apply command will integrate with:
- **Lock File Service**: For reading/writing package state
- **Package Managers**: For installing/removing packages (unchanged)
- **State Reconciler**: For determining what actions to take (unchanged)
- **Command Output**: For reporting results to user (unchanged)

### Success Criteria for Phase 4
- `plonk apply` installs packages from lock file (not config)
- Apply command handles missing, managed, and untracked packages correctly
- Lock file is updated appropriately during apply operations
- Error handling is robust for lock file operations
- All apply command tests pass with lock file integration
- Apply command maintains existing CLI interface and behavior

### Phase 4 Risks and Mitigation
**Risks**:
- Apply command is more complex than individual package commands
- Bulk operations have higher chance of partial failures
- Lock file consistency during multi-package operations

**Mitigation**:
- Use existing reconciler infrastructure (proven in status command)
- Implement atomic operations where possible
- Provide clear progress reporting and error messages
- Test thoroughly with various package states

## Open Questions

1. Should we track additional metadata (e.g., who installed, dependencies)?
2. Should lock file be in `.gitignore` by default?
3. Format for lock file - YAML vs JSON vs custom?

## Progress Tracking

- [x] Lock file interfaces defined (Phase 1)
- [x] Lock file service implemented (Phase 1)
- [x] Lock file adapter created (Phase 1)
- [x] Unit tests for lock service (Phase 1)
- [x] Config types updated (Phase 2 complete)
- [x] Core commands updated (Phase 3 complete)
- [ ] Apply command updated (Phase 4)
- [ ] Integration tests updated (Phase 5)
- [ ] Documentation updated (Phase 5)
- [ ] PR ready for review

## Lessons Learned from Phase 1

1. **Error Handling**: The existing error system (`errors.ConfigError`) requires error codes, operation names, and messages - not just simple error strings
2. **Atomic Writes**: The `AtomicFileWriter` is already available in the dotfiles package and takes no constructor arguments
3. **Testing Strategy**: Creating focused unit tests for each component helps validate the implementation incrementally
4. **Interface Adapters**: Creating adapters (like `LockFileAdapter`) helps bridge new components with existing interfaces without breaking changes

## Lessons Learned from Phase 2

1. **Complete Changes Over Incremental**: Making complete, cohesive changes is better than overly small incremental changes that leave the codebase in broken states
2. **Test-Driven Refactoring**: Updating tests in parallel with implementation helps catch issues early and ensures comprehensive coverage
3. **Validation Method Signatures**: The `ValidateConfig` method needed to return `*ValidationResult` instead of `error` to match the interface
4. **Default Configuration**: Adding a `GetDefaultConfig()` method helps with testing and initialization
5. **Consistent Error Messages**: Test assertions need to match the actual validation error messages (e.g., "min" instead of "timeout must be positive")
6. **Commit Completeness**: Commits should be "complete enough to pass tests and linter" - ensuring each commit maintains a working state

## Lessons Learned from Phase 3

1. **Interface Compatibility**: The `LockFileAdapter` implements `PackageConfigLoader` directly, so it can be used without wrapper adapters like `NewStatePackageConfigAdapter`
2. **Execution Order Matters**: Starting with read-only commands (`pkg list`) before write commands (`pkg add/remove`) reduces risk and validates the infrastructure
3. **Consistent Manager Support**: Adding cargo support consistently across all commands maintains feature parity and user expectations
4. **Atomic Operations**: Install first, then update lock file - this ensures system state remains consistent even if lock file operations fail
5. **Error Recovery**: Lock file operations should be designed to gracefully handle missing files (creating them as needed) vs. corrupted files (clear error messages)
6. **Command Output Structure**: Existing output structures can be extended (like adding `LockPath` to `StatusOutput`) without breaking compatibility
7. **Diagnostic Integration**: Adding lock file checks to `doctor` command provides essential debugging capabilities for users