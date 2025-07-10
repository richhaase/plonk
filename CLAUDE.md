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

### Phase 2: Config Refactoring
- [ ] Remove package fields from `Config` struct (homebrew, npm, cargo)
- [ ] Update `Settings` struct to only contain configuration settings
- [ ] Remove package-related methods from `ConfigService` and `ConfigAdapter`
- [ ] Update config validation to exclude package validation
- [ ] Update config tests to reflect package-free structure
- [ ] Ensure `plonk config show` only displays settings (not packages)

### Phase 3: Command Updates
- [ ] Update `pkg add` to write to lock file instead of config
- [ ] Update `pkg remove` to modify lock file instead of config
- [ ] Update `pkg list` to read from lock file (use existing adapter)
- [ ] Update `apply` command to use lock file for reconciliation
- [ ] Update `status` command to read from lock file
- [ ] Add lock file path to `doctor` command checks
- [ ] Update command tests to use lock file

### Phase 4: State Provider Updates
- [ ] Update reconciler initialization to use `LockFileAdapter` instead of `ConfigAdapter`
- [ ] Remove package-related methods from `ConfigAdapter`
- [ ] Update `MultiManagerPackageProvider` to use lock file adapter
- [ ] Test full reconciliation flow with lock file
- [ ] Update state provider tests

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

## Open Questions

1. Should we track additional metadata (e.g., who installed, dependencies)?
2. Should lock file be in `.gitignore` by default?
3. Format for lock file - YAML vs JSON vs custom?

## Progress Tracking

- [x] Lock file interfaces defined
- [x] Lock file service implemented  
- [x] Lock file adapter created
- [x] Unit tests for lock service
- [ ] Config types updated
- [ ] Commands updated
- [ ] Tests updated
- [ ] Documentation updated
- [ ] PR ready for review

## Lessons Learned from Phase 1

1. **Error Handling**: The existing error system (`errors.ConfigError`) requires error codes, operation names, and messages - not just simple error strings
2. **Atomic Writes**: The `AtomicFileWriter` is already available in the dotfiles package and takes no constructor arguments
3. **Testing Strategy**: Creating focused unit tests for each component helps validate the implementation incrementally
4. **Interface Adapters**: Creating adapters (like `LockFileAdapter`) helps bridge new components with existing interfaces without breaking changes