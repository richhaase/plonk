# Fixes and Improvements

Actionable improvements identified from comprehensive code review.

## Critical Fixes

### 1. ✅ Fix pipx home directory resolution bug (COMPLETED)
**Priority**: High (Bug)
**File**: `internal/resources/packages/pipx.go`
**Status**: ✅ Completed in Phase 1 (commit 10e9002)

**Problem**: `getBinDirectory` uses `filepath.Abs("~")` which doesn't resolve home directory correctly - it creates a path relative to CWD, not `$HOME`.

**Fix Applied**:
```go
// Replaced in getBinDirectory fallback:
home, err := os.UserHomeDir()
if err != nil {
    return "", fmt.Errorf("failed to get user home directory: %w", err)
}
return filepath.Join(home, ".local", "bin"), nil
```

### 2. ~~Security: Remote install script verification~~ (SUPERSEDED by #4)
**Status**: No longer needed - removing self-install functionality entirely per #4

### 3. ✅ Fix MultiPackageResource concurrency safety (COMPLETED)
**Priority**: Medium (Potential race condition)
**File**: `internal/resources/packages/resource.go`
**Status**: ✅ Completed in Phase 1 (commit 10e9002)

**Problem**: `MultiPackageResource.Actual` directly iterates over `m.registry.managers` map, breaking encapsulation and risking data races if managers are ever registered dynamically.

**Fix Applied**:
```go
// In MultiPackageResource.Actual, replaced:
// for managerName := range m.registry.managers

// With:
for _, managerName := range m.registry.GetAllManagerNames() {
    manager, err := m.registry.GetManager(managerName)
    if err != nil {
        continue
    }
    // ... rest of logic
}
```

## High-Priority Improvements

### 4. ✅ Remove self-install functionality for package managers (COMPLETED)
**Priority**: High (Security & Architecture)
**Status**: ✅ Completed in Phase 1 (commit 10e9002)

**Files Changed**: 26 files (10 managers + 2 commands + 2 interfaces + 12 tests)
**Files Deleted**: `internal/resources/packages/install_helpers.go`

**Problem**: Self-install via shell scripts (`curl | sh`) posed security risks and added complexity.

**Changes Applied**:
- ✅ Deleted `PackageSelfInstaller` interface from `interfaces.go`
- ✅ Deleted `SupportsSelfInstall()` capability check function
- ✅ Deleted entire `install_helpers.go` file (executeInstallScript, executeInstallCommand, etc.)
- ✅ Removed `SelfInstall()` methods from all 10 package managers:
  - homebrew.go, cargo.go, npm.go, pnpm.go, pipx.go
  - gem.go, pixi.go, uv.go, goinstall.go, conda.go
- ✅ Updated `install.go` to return clear error directing users to manual installation
- ✅ Updated `clone/setup.go` to skip unavailable managers with helpful messages
- ✅ Updated all test files to remove SelfInstall stubs

**Results**:
- 🔒 Eliminates remote script execution security vulnerability
- 🧹 Removed ~326 lines of complex installation code
- 📝 Clearer user expectations and better error messages
- 🏥 `plonk doctor` is now the authoritative source for installation help

**New User Flow**:
```bash
# User tries to install package but manager is missing
$ plonk install pipx:black
Error: Package manager 'pipx' is not available.
Install it manually or via another package manager, then try again.
Run 'plonk doctor' for installation instructions.

# User runs doctor to see how to install
$ plonk doctor
### Pipx Package Manager
**Status**: WARN
**Suggestions:**
- Install pipx via pip: pip3 install --user pipx
- Or via Homebrew: brew install pipx
```

### 5. ✅ Standardize manager registration to V2 (COMPLETED)
**Priority**: High (Architecture consistency)
**Status**: ✅ Completed in Phase 3 (commit 66e5531)

**Files**:
- `internal/resources/packages/uv.go`
- `internal/resources/packages/goinstall.go`
- `internal/resources/packages/pixi.go`

**Problem**: Some managers still use V1 `RegisterManager` and don't accept injected executors, while most use V2. This reduces testability and creates inconsistency.

**Changes Applied**:
- Converted all three managers to `RegisterManagerV2` with factory pattern
- Each now accepts injected `CommandExecutor` for testability
- Consistent with all other package managers (brew, npm, pnpm, cargo, pipx, conda, gem)

**Benefits**: Complete testability, consistent execution across all 10 package managers

### 6. ✅ Parallelize manager operations (COMPLETED)
**Priority**: Medium (Performance)
**File**: `internal/resources/packages/resource.go`
**Status**: ✅ Completed in Phase 3 (commit 66e5531)

**Problem**: `MultiPackageResource.Actual` runs `IsAvailable` + `ListInstalled` for each manager sequentially, which is slow.

**Changes Applied**:
- Implemented `errgroup.WithContext` for parallel execution
- Concurrency limited to `min(runtime.GOMAXPROCS(0), 4)` for safety
- Thread-safe with `sync.Mutex` for item collection
- Honors context cancellation properly
- Added imports: `golang.org/x/sync/errgroup`, `sync`, `runtime`

**Benefits**: Significant speedup for `plonk status` and `plonk apply` operations, especially with many managers

### 7. ✅ Switch Homebrew Info to JSON parsing (COMPLETED)
**Priority**: Medium (Robustness)
**File**: `internal/resources/packages/homebrew.go`
**Status**: ✅ Completed in Phase 3 (commit 66e5531)

**Problem**: `Info` method scrapes plain text output, which is brittle and locale-dependent.

**Changes Applied**:
- Changed from `brew info <name>` to `brew info --json=v2 <name>`
- Parses structured JSON response (formulae and casks)
- Handles both installed and stable versions
- Robust, locale-independent implementation

**Benefits**: More reliable package info, easier to maintain, consistent with ListInstalled implementation

### 7. Centralize package key generation
**Priority**: Low (Code quality)
**File**: `internal/resources/packages/reconcile.go`

**Current**: Uses `manager:name` as key in multiple places.

**Improvement**: Create helper function to ensure consistency:
```go
// Add to internal/resources/packages/types.go or reconcile.go
func KeyForPackage(managerName, packageName string) string {
    return fmt.Sprintf("%s:%s", managerName, packageName)
}

// Use everywhere instead of inline fmt.Sprintf
```

## UX/Display Improvements

### 8. ✅ Fix duplicate listing of drifted dotfiles (COMPLETED)
**Priority**: Medium (UX bug)
**Command**: `plonk status`
**Status**: ✅ Completed (commit 55e9249)

**Problem**: When a dotfile is in drifted status, it appears twice in the output.

**Fix Applied**:
- Removed misleading comment about handling drifted items separately
- Drifted files (State==StateDegraded) already in Managed list now display once with "drifted" status
- File: `internal/output/status_formatter.go`

### 9. ✅ Improve dotfile column headers in status (COMPLETED)
**Priority**: Medium (UX improvement)
**Command**: `plonk status`
**Status**: ✅ Completed (commit 55e9249, refined in 66e5531)

**Problem**: Column headers "source" and "target" are confusing for dotfiles.

**Fix Applied**:
- Changed headers from "SOURCE", "TARGET", "STATUS" to "$HOME", "$PLONK_DIR", "STATUS"
- Reordered columns: $HOME (deployed location), $PLONK_DIR (source), STATUS
- Updated all AddRow calls to match new column order
- Corrected $PLONKDIR → $PLONK_DIR throughout codebase and docs
- File: `internal/output/status_formatter.go`

### 10. ✅ Fix diff output column ordering (COMPLETED)
**Priority**: Medium (UX improvement)
**Command**: `plonk diff`
**Status**: ✅ Completed (commit 55e9249)

**Problem**: Current ordering is inconsistent with user mental model.

**Fix Applied**:
- Swapped diff arguments from `source, dest` to `dest, source`
- Now shows $HOME (deployed) on left and $PLONKDIR (source) on right
- Matches standard diff conventions (current state vs. source)
- File: `internal/commands/diff.go`

## Feature Additions

### 11. ✅ Add `plonk add -y` to sync drifted files back to $PLONKDIR (COMPLETED)
**Priority**: Medium (Feature enhancement)
**Command**: `plonk add`
**Status**: ✅ Completed (commit 55e9249)

**Feature**: Add a `-y` or `--sync-drifted` flag to automatically copy all drifted files from $HOME back to $PLONKDIR (reverse of `apply`).

**Implementation**:
```bash
plonk add -y                 # Sync all drifted files
plonk add -y --dry-run       # Preview what would be synced
```

**Changes Applied**:
- Added `--sync-drifted` flag (short: `-y`) to add command
- Finds all files with State==StateDegraded (drifted)
- Copies them from $HOME back to $PLONKDIR
- Shows summary of synced files
- Works with `--dry-run` for preview
- Shows appropriate message when no files are drifted
- File: `internal/commands/add.go`

### 12. ✅ Add selective file deployment to `plonk apply` (COMPLETED)
**Priority**: Medium (Feature enhancement)
**Command**: `plonk apply`
**Status**: ✅ Completed (commit 55e9249)

**Feature**: Allow `plonk apply <file1> <file2> ...` to selectively deploy only specified files from $PLONKDIR to $HOME.

**Implementation**:
```bash
plonk apply ~/.vimrc ~/.zshrc    # Apply only specified files
plonk apply                       # Apply all (original behavior)
```

**Changes Applied**:
- Modified command to accept optional file arguments: `apply [files...]`
- Validates that specified files are managed by plonk before proceeding
- Shows clear error if file not found or not managed
- Prevents combining file arguments with `--packages` or `--dotfiles` flags
- Updated help text with examples
- File: `internal/commands/apply.go`

## Nice-to-Haves

### 13. Remove unused code
**Files**:
- `internal/commands/install.go`: `managerInstallCount` computed but never used
- `internal/resources/reconcile.go`: `ReconcileItemsDeprecated` if marked safe to remove

### 14. Dotfile directory permissions
**File**: `internal/resources/dotfiles/manager.go`

**Current**: Uses `0750` for parent directories in `AddSingleFile`.

**Consideration**: Use `0755` for better cross-platform compatibility unless there's a security reason for `0750`. Make it configurable if needed.

### 15. Registry isolation for tests
**File**: `internal/resources/packages/registry.go`

**Improvement**: Add `NewIsolatedRegistry()` that returns a fresh registry instance (not the singleton) for test isolation:
```go
func NewIsolatedRegistry() *ManagerRegistry {
    reg := &ManagerRegistry{
        managers:   make(map[string]PackageManager),
        factoriesV2: make(map[string]ManagerFactoryV2),
    }
    // Copy registrations from defaultRegistry
    for name, factory := range defaultRegistry.factoriesV2 {
        reg.factoriesV2[name] = factory
    }
    return reg
}
```

### 16. ✅ Add comprehensive tests (PARTIALLY COMPLETED)
**Status**: ✅ Added drift and sync tests (commit 5800789, f2646a0)

**Completed**:
- ✅ Created `tests/bats/behavioral/10-drift-and-sync.bats` with 11 new tests
- ✅ Tests for duplicate drifted dotfiles, column headers, diff ordering
- ✅ Tests for `plonk add -y` sync functionality
- ✅ Tests for selective `plonk apply <files>`
- ✅ Integration test for complete drift workflow

**Remaining**:
1. Concurrency smoke test for parallel manager reconciliation
2. Symlink traversal tests for dotfile operations

### 17. Documentation updates
**Files**: README.md

**Tasks**:
1. Update architecture.md to reflect ManagerRegistry V2 and reconciliation keying
2. Document that package managers must be installed manually or via other supported package managers

### 18. Performance: Cache ListInstalled within reconciliation
**Files**: Various package managers

**Idea**: Cache `ListInstalled` results per manager within a single command execution to avoid duplicate subprocess calls (e.g., upgrade calls before/after). Could be a simple map keyed by manager name with TTL of the command duration.

### 19. Per-package version lookups optimization
**Files**:
- `internal/resources/packages/pnpm.go`
- `internal/resources/packages/pipx.go`

**Current**: Fetch full package lists to determine single package version.

**Improvement**: If per-package commands exist (like npm's `npm list -g <name>`), prefer them to avoid parsing large lists repeatedly.

## Implementation Priority

### Phase 1: Critical ✅ COMPLETED
1. ✅ Fix pipx home directory bug (#1)
2. ✅ Fix MultiPackageResource concurrency (#3)
3. ✅ Remove self-install functionality (#4)

**Completed**: 2025-01-06 in commits 619924f, 10e9002
**Results**: 3 critical fixes, 26 files changed, 1 file deleted, -293 LOC, all tests passing

### Phase 2: UX Improvements ✅ COMPLETED
4. ✅ Fix duplicate drifted dotfiles in status (#8)
5. ✅ Improve dotfile column headers in status (#9)
6. ✅ Fix diff output column ordering (#10)
7. ✅ Add `plonk add -y` to sync drifted files (#11)
8. ✅ Add selective file deployment to `plonk apply` (#12)

**Completed**: 2025-01-06 in commit 55e9249
**Results**: 5 UX improvements, 4 files changed, +190 LOC, all tests passing

### Phase 3: Architecture & Performance ✅ COMPLETED
9. ✅ Standardize V2 registration (#5)
10. ✅ Parallelize manager operations (#6)
11. ✅ Switch Homebrew to JSON (#7)

**Completed**: 2025-01-06 in commit 66e5531
**Results**: 4 improvements (incl. $PLONK_DIR fix), 15 files changed, +126/-53 LOC, all tests passing, significant performance gains

### Phase 4: Polish
12. Centralize package keying (#7)
13. Remove unused code (#13)
14. Add comprehensive tests (#16)
