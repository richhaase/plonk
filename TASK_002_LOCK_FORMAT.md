# Task Context: Enhance Lock File Format

## Task ID: TASK_002_LOCK_FORMAT
## Phase: 1 - Foundation
## Priority: High (Blocks Setup Features)
## Estimated Effort: Medium

## Problem Statement
The current lock file format (v1) stores only the binary name for Go packages, losing the source path information. This makes it impossible to reinstall packages correctly and blocks intelligent setup features that need to know which package managers are required.

### Current Problem Examples
- `golang.org/x/tools/cmd/gopls` → stored as `gopls` (loses source path)
- `@npmcli/arborist` → stored as `arborist` (loses scope)
- Cannot reinstall Go packages without user re-specifying the full path
- Cannot auto-detect which package managers are needed from lock file

### Solution
Implement lock file format v2 that stores both binary name and source path in metadata, while maintaining backward compatibility.

## Technical Context

### Key Files
- **Lock Types**: `internal/lock/types.go` - Defines lock file structures
- **Lock Service**: `internal/lock/yaml_lock.go` - Implements reading/writing
- **Package Operations**: `internal/resources/packages/operations.go` - Uses lock service
- **Go Install**: `internal/resources/packages/goinstall.go` - Contains `ExtractBinaryNameFromPath()`

### Current Implementation
The code already has scaffolding for v2:
- `LockV2` struct exists with `Resources` array and `Metadata` field
- Version detection logic exists (`readV1` vs `readV2`)
- Migration path partially implemented

### Lock File Versions
- **v1**: Current format with `packages` map (manager → array of packages)
- **v2**: New format with `resources` array and metadata support

## Implementation Requirements

### 1. Complete v2 Format Implementation
```yaml
# Example v2 format
version: 2
resources:
  - type: package
    id: brew:htop
    state: managed
    metadata:
      manager: brew
      name: htop
      version: "3.2.1"
    installed_at: "2024-01-15T10:30:00Z"

  - type: package
    id: go:gopls
    state: managed
    metadata:
      manager: go
      name: gopls
      source_path: golang.org/x/tools/cmd/gopls
      version: "0.14.2"
    installed_at: "2024-01-15T10:31:00Z"
```

### 2. Migration Strategy
- Auto-migrate v1 to v2 on first write operation
- Preserve all existing data
- For Go packages without source_path, mark as "legacy" in metadata
- Continue to support reading v1 files

### 3. Update Package Operations
- Modify `InstallPackage()` to store source path in metadata
- Update `UninstallPackage()` to use metadata when available
- Ensure backward compatibility for packages without metadata

### 4. Testing Requirements
- Test v1 → v2 migration
- Test new installs create v2 format
- Test reading both v1 and v2 formats
- Test Go package reinstallation with source path

## Implementation Steps

### Step 1: Complete v2 Write Implementation
- Implement `writeV2()` method in YAMLLockService
- Update `AddPackage()` to accept and store metadata
- Ensure proper YAML formatting

### Step 2: Update Package Operations
- Modify `InstallPackage()` to pass source path as metadata
- Special handling for Go packages to preserve full module path
- Similar handling for npm scoped packages

### Step 3: Implement Migration
- Add `migrateV1ToV2()` method
- Call migration on first write operation
- Preserve all v1 data in v2 format

### Step 4: Update Read Operations
- Ensure `FindPackage()` works with both formats
- Update package listing to use v2 data when available

## Success Criteria
1. New installations create v2 lock files with full source paths
2. Existing v1 lock files are auto-migrated on first write
3. Go packages can be reinstalled using stored source path
4. Setup commands can detect required managers from lock file
5. No breaking changes for existing users

## Deliverables
1. Implemented v2 lock file format with metadata support
2. Auto-migration from v1 to v2
3. Updated package operations to use new format
4. Test coverage for migration and compatibility
5. Summary document explaining:
   - Implementation approach
   - Migration strategy
   - How metadata is used
   - Any design decisions made

## Important Considerations
- **Backward Compatibility**: Must continue reading v1 files
- **Forward Compatibility**: Design metadata to be extensible
- **Migration Safety**: Never lose data during migration
- **Performance**: Migration should be fast even for large lock files

## Example Metadata Uses
```yaml
# Go package with source
metadata:
  manager: go
  name: gopls
  source_path: golang.org/x/tools/cmd/gopls

# NPM scoped package
metadata:
  manager: npm
  name: arborist
  scope: "@npmcli"
  full_name: "@npmcli/arborist"

# Legacy migrated package (no source)
metadata:
  manager: go
  name: staticcheck
  legacy: true  # Indicates migrated without source path
```
