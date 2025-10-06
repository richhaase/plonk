# Phase 1 Implementation Summary

**Status**: ✅ COMPLETE
**Date**: 2025-01-06

## Overview

Phase 1 focused on critical fixes identified in the comprehensive code review, addressing security vulnerabilities, bugs, and architectural issues.

## Changes Implemented

### 1. Fixed pipx home directory bug (#1)

**Problem**: `pipx.go` used `filepath.Abs("~")` which doesn't resolve home directory correctly - it creates a path relative to CWD, not `$HOME`.

**Solution**:
- Added `import "os"` to pipx.go
- Replaced `filepath.Abs("~")` with `os.UserHomeDir()`
- Added proper error handling for home directory lookup failure

**Files Modified**:
- `internal/resources/packages/pipx.go`

**Impact**: pipx binary directory detection now works correctly across all environments.

---

### 2. Fixed MultiPackageResource concurrency safety (#3)

**Problem**: `MultiPackageResource.Actual()` directly accessed `m.registry.managers` map, breaking encapsulation and risking data races.

**Solution**:
- Changed from `for managerName := range m.registry.managers`
- To `for _, managerName := range m.registry.GetAllManagerNames()`
- Uses proper registry API instead of accessing internal fields

**Files Modified**:
- `internal/resources/packages/resource.go`

**Impact**: Eliminates potential race conditions and improves encapsulation.

---

### 3. Removed self-install functionality (#4) 🔐

**Problem**: Self-install via remote shell scripts (`curl | sh`) posed significant security risks:
- No hash/signature verification
- Remote code execution from arbitrary URLs
- Supply chain attack vector
- Added complexity to codebase

**Solution - Complete Removal**:

#### Interface Changes
- Deleted `PackageSelfInstaller` interface from `interfaces.go`
- Deleted `SupportsSelfInstall()` capability check function
- Deleted entire `install_helpers.go` file containing:
  - `executeInstallScript()` - ran bash/sh scripts
  - `executeInstallCommand()` - wrapped install commands
  - `checkPackageManagerAvailable()` - availability checker

#### Package Manager Changes
Removed `SelfInstall()` methods and helper functions from all 10 managers:
1. **homebrew.go** - Removed `SelfInstall()` (was: remote install script)
2. **cargo.go** - Removed `SelfInstall()` (was: rustup script)
3. **npm.go** - Removed `SelfInstall()` and `installViaHomebrew()`
4. **pnpm.go** - Removed `SelfInstall()` and `installViaStandaloneScript()`
5. **pipx.go** - Removed `SelfInstall()` and `installViaHomebrew()`
6. **gem.go** - Removed `SelfInstall()` and `installViaHomebrew()`
7. **pixi.go** - Removed `SelfInstall()`
8. **uv.go** - Removed `SelfInstall()`
9. **goinstall.go** - Removed `SelfInstall()` and `installViaHomebrew()`
10. **conda.go** - Removed `SelfInstall()` and `installMicromambaViaHomebrew()`

#### Command Logic Updates
- **install.go**: Updated `handleManagerSelfInstall()` to return clear error:
  ```
  Package manager 'X' is not available.
  Install it manually or via another package manager, then try again.
  Run 'plonk doctor' for installation instructions.
  ```

- **clone/setup.go**: Updated manager installation logic to skip unavailable managers with helpful messages instead of attempting self-install

#### Test Updates (12 files)
Removed `SelfInstall()` stub implementations from all fake/mock managers:
- `operations_injection_test.go`
- `apply_flow_more_test.go`
- `dependencies_cycle_test.go`
- `operations_error_paths_test.go`
- `operations_flow_test.go`
- `operations_metadata_test.go`
- `operations_more_test.go`
- `operations_uninstall_paths_test.go`
- `operations_version_err_test.go`
- `timeout_test.go`
- `capabilities_test.go` - Removed `SupportsSelfInstall()` test
- `clone/setup_more_test.go` - Updated to verify self-install now returns error

#### What Remains
- ✅ `Dependencies()` methods on all managers - still used by `plonk doctor`
- ✅ `CheckHealth()` methods provide installation instructions
- ✅ All package manager detection and usage functionality

**Files Modified**: 26 files (10 managers + 2 commands + 2 interfaces + 12 tests)
**Files Deleted**: 1 file (`install_helpers.go`)

**New User Flow**:
```bash
# User tries to install package but manager is missing
$ plonk install pipx:black
Error: Package manager 'pipx' is not available.
Install it manually or via another package manager, then try again.
Run 'plonk doctor' for installation instructions.

# User runs doctor to see how to install
$ plonk doctor
...
### Pipx Package Manager
**Status**: WARN
**Message**: pipx is not available

**Suggestions:**
- Install pipx via pip: pip3 install --user pipx
- Or via Homebrew: brew install pipx
- After installation, ensure pipx is in your PATH
```

**Impact**:
- 🔒 Eliminates remote script execution security vulnerability
- 🧹 Removes 300+ lines of complex installation code
- 📝 Clearer user expectations and better error messages
- 🏥 `plonk doctor` becomes the authoritative source for installation help

---

## Verification

### Tests
```bash
$ go test ./...
ok      github.com/richhaase/plonk/internal/clone
ok      github.com/richhaase/plonk/internal/commands
ok      github.com/richhaase/plonk/internal/config
ok      github.com/richhaase/plonk/internal/diagnostics
ok      github.com/richhaase/plonk/internal/lock
ok      github.com/richhaase/plonk/internal/orchestrator
ok      github.com/richhaase/plonk/internal/output
ok      github.com/richhaase/plonk/internal/resources
ok      github.com/richhaase/plonk/internal/resources/dotfiles
ok      github.com/richhaase/plonk/internal/resources/packages
ok      github.com/richhaase/plonk/internal/testutil
```

### Lint
```bash
$ just lint
Running linter...
go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m
✓ No issues found
```

### BATS Integration Tests
```bash
$ cd tests && bats bats/behavioral/09-error-scenarios.bats -f "package manager unavailable"
ok 1 install handles package manager unavailable gracefully
```

### Manual Testing
```bash
$ plonk doctor
Overall Status: WARNING (if any managers missing)

## Package Managers
### Homebrew Manager
**Status**: PASS
**Message**: Homebrew is available and properly configured
...

### Pipx Package Manager
**Status**: WARN
**Message**: pipx is not available
**Suggestions:**
- Install pipx via pip: pip3 install --user pipx
- Or via Homebrew: brew install pipx
```

---

## Metrics

- **Files Changed**: 26
- **Files Deleted**: 1
- **Lines Removed**: ~326
- **Lines Added**: ~33
- **Net Change**: -293 lines (simpler codebase)
- **Test Coverage**: Maintained (all tests pass)
- **Security Issues Fixed**: 1 critical (remote script execution)
- **Bugs Fixed**: 2 (pipx home dir, concurrency)

---

## Breaking Changes

### For Users
⚠️ **Package managers can no longer be auto-installed**

**Before**:
```bash
$ plonk install pnpm:typescript
# Would automatically install pnpm via remote script if missing
```

**After**:
```bash
$ plonk install pnpm:typescript
Error: Package manager 'pnpm' is not available.
Install it manually or via another package manager, then try again.
Run 'plonk doctor' for installation instructions.

$ plonk doctor
# Shows how to install pnpm

$ brew install pnpm
$ plonk install pnpm:typescript
✓ Success
```

### For Developers
- `PackageSelfInstaller` interface removed
- `SupportsSelfInstall()` function removed
- `install_helpers.go` deleted
- All `SelfInstall()` methods removed from managers

---

## Next Steps

Phase 1 is complete. Ready to proceed to:
- **Phase 2**: UX Improvements (duplicate dotfiles, column headers, diff ordering, add -y, selective apply)
- **Phase 3**: Architecture & Performance (V2 registration, parallelization, Homebrew JSON)
- **Phase 4**: Polish (cleanup, tests, documentation)

---

## Commits

1. `619924f` - docs: add comprehensive review findings and improvement plan
2. `10e9002` - feat: implement Phase 1 critical fixes
