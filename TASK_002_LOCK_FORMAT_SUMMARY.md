# Lock File Format v2 Implementation Summary

## Overview
Successfully implemented lock file format v2 with metadata support for storing package source paths. The implementation opted for a complete breaking change instead of migration support, dramatically simplifying the codebase.

## Implementation Approach

### Breaking Change Decision
- Removed all v1 format support and migration logic
- Simplified API by eliminating dual read/write methods
- Clear error message for users with old lock files: "unsupported lock file version 1 (expected 2). Please remove plonk.lock and reinstall your packages"

### Key Changes

1. **Type Simplification**
   - Removed: `LockFile`, `PackageEntry`, `Package`, `LockData` types
   - Kept only: `Lock` and `ResourceEntry` types
   - Removed version detection and migration code

2. **Interface Simplification**
   - Single `Read()` and `Write()` methods (removed `Load()`/`Save()`)
   - `AddPackage()` now accepts metadata parameter
   - All methods work with `ResourceEntry` instead of mixed types

3. **Metadata Implementation**
   - Go packages store `source_path` (e.g., "golang.org/x/tools/cmd/gopls")
   - NPM scoped packages store `scope` and `full_name` (e.g., "@npmcli/arborist")
   - All packages store `manager`, `name`, and `version` in metadata

## Lock File Format Example

```yaml
version: 2
resources:
  - type: package
    id: go:gopls
    metadata:
      manager: go
      name: gopls
      source_path: golang.org/x/tools/cmd/gopls
      version: v0.14.2
    installed_at: "2025-07-29T12:04:50-06:00"

  - type: package
    id: npm:arborist
    metadata:
      full_name: '@npmcli/arborist'
      manager: npm
      name: arborist
      scope: '@npmcli'
      version: 6.2.0
    installed_at: "2025-07-29T12:04:50-06:00"
```

## Design Decisions

1. **No State Field**: Removed redundant `state: managed` field since being in the lock file inherently means the resource is managed.

2. **Resource ID Format**: Kept simple `manager:name` format for compatibility with existing code patterns.

3. **Metadata Flexibility**: Used `map[string]interface{}` for metadata to allow future extensibility without schema changes.

4. **Breaking Change Benefits**:
   - Cleaner codebase with ~50% less lock-related code
   - No complex migration logic to maintain
   - Single code path for all operations
   - Easier to test and reason about

## How Metadata is Used

1. **Go Packages**: The `source_path` metadata allows reinstallation using the full module path instead of just the binary name.

2. **NPM Scoped Packages**: The `scope` and `full_name` metadata preserve the complete package identity for reinstallation.

3. **Future Extensibility**: The metadata field can store additional information like:
   - Installation flags or options
   - Dependency information
   - Custom user annotations

## Files Modified

- `internal/lock/types.go` - Simplified to only v2 types
- `internal/lock/interfaces.go` - Simplified interface with metadata support
- `internal/lock/yaml_lock.go` - Complete rewrite for v2-only support
- `internal/resources/packages/operations.go` - Updated to pass metadata when installing
- `internal/resources/packages/reconcile.go` - Updated to read v2 format
- `internal/diagnostics/health.go` - Updated lock file validity check
- `internal/lock/yaml_lock_test.go` - Complete test rewrite for v2

## Testing

All tests pass successfully:
- Lock file read/write operations
- Package addition with metadata
- Package removal
- Package queries (GetPackages, FindPackage, HasPackage)
- Unsupported version error handling
- NPM scoped package handling
- Go package source path preservation

## User Impact

Users with existing v1 lock files will need to:
1. Remove their `plonk.lock` file
2. Reinstall packages using `plonk install` commands
3. Future installations will automatically use v2 format with proper metadata

This one-time inconvenience enables a much cleaner and more maintainable codebase going forward.
