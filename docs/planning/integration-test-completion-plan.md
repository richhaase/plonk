# Integration Test Completion Plan

**Created**: 2025-08-04
**Purpose**: Add happy path integration tests for all plonk commands
**Approach**: One test file per command, testing real behavior in Docker

## Current State

✅ **Completed**:
- Infrastructure setup (testcontainers-go, Dockerfile)
- First test working (TestInstallPackage in container_test.go)
- JSON output parsing validated
- Stderr/stdout separation fixed

## Test Files to Create

Each command gets its own test file in `tests/integration/`:

### 1. `install_test.go` ✅
- Already exists as part of container_test.go
- Move TestInstallPackage here for consistency

### 2. `uninstall_test.go`
- Install a package (brew:jq)
- Verify it's in the lock file
- Uninstall it
- Verify it's removed from lock file
- Verify `brew list` doesn't show it

### 3. `status_test.go`
- Install 2-3 packages (brew:jq, brew:tree, brew:wget)
- Run status command
- Verify all packages appear in managed_items
- Verify counts are correct

### 4. `apply_test.go`
**Test apply with packages:**
- Manually create a plonk.lock with 2 packages
- Run `plonk apply`
- Verify packages get installed via brew
- Verify status shows them as managed

**Test apply with dotfiles:**
- Create test dotfiles (.testrc, .testconf)
- Add them with `plonk add`
- Run `plonk apply`
- Verify symlinks are created in home directory

### 5. `add_test.go`
- Create test dotfiles in container
- Add them to plonk management
- Verify they appear in lock file
- Verify source files exist in config directory

### 6. `rm_test.go`
- Add a dotfile
- Remove it with `plonk rm`
- Verify it's removed from lock file
- Verify source file is deleted
- Verify symlink is removed

### 7. `search_test.go`
- Search for a known package (e.g., "git")
- Verify results contain expected package
- Test with different package managers if specified

### 8. `list_test.go`
- Install multiple packages
- Add multiple dotfiles
- Run `plonk list`
- Verify all items appear with correct states

### 9. `diff_test.go`
- Add a dotfile
- Modify the deployed version
- Run `plonk diff`
- Verify it detects the drift

### 10. `clone_test.go`
- Create a test git repo with dotfiles
- Clone it with `plonk clone`
- Verify plonk directory is set up
- Verify lock file is read correctly

### 11. `config_test.go`
- Test `config show` returns valid JSON
- Test modifying config affects behavior

## Test Pattern

Each test should follow this structure:

```go
//go:build integration

package integration_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCommandName(t *testing.T) {
    env := NewTestEnv(t)

    // Setup (if needed)

    // Execute command
    var result struct {
        // Expected JSON structure
    }
    err := env.RunJSON(&result, "command", "args...")
    require.NoError(t, err)

    // Verify results
    assert.Equal(t, expected, actual)

    // Verify side effects (brew list, file existence, etc.)
}
```

## Implementation Order

1. Start with package commands (uninstall, status, list, search)
2. Then dotfile commands (add, rm, diff)
3. Then orchestration commands (apply, clone)
4. Finally config commands

## Common Test Helpers Needed

Add to `container_test.go`:

- `CreateTestDotfile(name, content string)` - Creates a dotfile in test home
- `FileExists(path string) bool` - Checks if file exists in container
- `IsSymlink(path string) bool` - Checks if path is a symlink
- `GetLockFileContent() (LockFile, error)` - Reads and parses lock file

## Success Criteria

- [x] All commands have at least one happy path test
- [x] Tests use real packages and real operations
- [x] Tests validate both JSON output and actual system state
- [x] Each test is independent and can run in isolation

## Implementation Status (2025-08-04)

All integration tests have been implemented:

1. **`install_test.go`** ✅ - Tests package installation (part of container_test.go)
2. **`uninstall_test.go`** ✅ - Tests package uninstallation
3. **`status_test.go`** ✅ - Tests status with multiple packages
4. **`list_test.go`** ✅ - Tests listing packages
5. **`search_test.go`** ✅ - Tests package search functionality
6. **`apply_test.go`** ✅ - Tests applying packages and dotfiles
7. **`add_test.go`** ✅ - Tests adding dotfiles
8. **`rm_test.go`** ✅ - Tests removing dotfiles
9. **`diff_test.go`** ✅ - Tests diffing dotfiles
10. **`clone_test.go`** ✅ - Tests cloning repositories
11. **`config_test.go`** ✅ - Tests config show and custom values

### Key Implementation Details

- Added `WriteFile` helper method to TestEnv for creating files in containers
- Fixed JSON structure mismatches (e.g., add command returns single object, not array)
- All tests verify both command output and actual system state
- Tests use real packages (jq, tree, wget, curl, htop) and real operations
