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
- [ ] Remove package fields from `Config` struct
- [ ] Update `Settings` to only contain configuration settings
- [ ] Remove package-related methods from `ConfigService`
- [ ] Update config validation to exclude package validation

### Phase 3: Command Updates
- [ ] Update `pkg add` to write to lock file instead of config
- [ ] Update `pkg remove` to modify lock file
- [ ] Update `pkg list` to read from lock file
- [ ] Update `apply` command to use lock file for reconciliation
- [ ] Update `status` command to read from lock file

### Phase 4: State Provider Updates
- [ ] Create `LockFilePackageConfigLoader` implementing `PackageConfigLoader`
- [ ] Update `PackageProvider` to use lock file loader
- [ ] Update reconciler initialization to use lock file
- [ ] Remove `ConfigAdapter` or update to only handle settings

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

## Open Questions

1. Should we track additional metadata (e.g., who installed, dependencies)?
2. Should lock file be in `.gitignore` by default?
3. Format for lock file - YAML vs JSON vs custom?

## Progress Tracking

- [ ] Lock file interfaces defined
- [ ] Lock file service implemented
- [ ] Config types updated
- [ ] Commands updated
- [ ] Tests updated
- [ ] Documentation updated
- [ ] PR ready for review