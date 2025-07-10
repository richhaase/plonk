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

### Phase 3: Command Updates
- [ ] Re-enable `pkg add` command with lock file integration
- [ ] Re-enable `pkg remove` command with lock file integration
- [ ] Update `pkg list` to read from lock file (use existing adapter)
- [ ] Update `apply` command to use lock file for reconciliation
- [ ] Update `status` command to read from lock file
- [ ] Add lock file path to `doctor` command checks
- [ ] Update command tests to use lock file
- [ ] Ensure proper error handling for lock file operations

### Phase 4: State Provider Updates
- [ ] Update reconciler initialization to use `LockFileAdapter` instead of `ConfigAdapter`
- [ ] Update `MultiManagerPackageProvider` to use lock file adapter
- [ ] Test full reconciliation flow with lock file
- [ ] Update state provider tests
- [ ] Verify end-to-end package reconciliation workflow
- [ ] Update any remaining references to config-based package management

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

## Open Questions

1. Should we track additional metadata (e.g., who installed, dependencies)?
2. Should lock file be in `.gitignore` by default?
3. Format for lock file - YAML vs JSON vs custom?

## Progress Tracking

- [x] Lock file interfaces defined
- [x] Lock file service implemented  
- [x] Lock file adapter created
- [x] Unit tests for lock service
- [x] Config types updated (Phase 2 complete)
- [ ] Commands updated (Phase 3)
- [ ] State provider updated (Phase 4)
- [ ] Tests updated (Phase 5)
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