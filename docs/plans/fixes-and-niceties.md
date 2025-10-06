# Fixes and Improvements

Actionable improvements identified from comprehensive code review.

## Critical Fixes

### 1. Fix pipx home directory resolution bug
**Priority**: High (Bug)
**File**: `internal/resources/packages/pipx.go`

**Problem**: `getBinDirectory` uses `filepath.Abs("~")` which doesn't resolve home directory correctly - it creates a path relative to CWD, not `$HOME`.

**Fix**:
```go
// Replace in getBinDirectory fallback:
home, err := os.UserHomeDir()
if err != nil {
    return "", fmt.Errorf("failed to get user home directory: %w", err)
}
return filepath.Join(home, ".local", "bin"), nil
```

### 2. ~~Security: Remote install script verification~~ (SUPERSEDED by #4)
**Status**: No longer needed - removing self-install functionality entirely per #4

### 3. Fix MultiPackageResource concurrency safety
**Priority**: Medium (Potential race condition)
**File**: `internal/resources/packages/resource.go`

**Problem**: `MultiPackageResource.Actual` directly iterates over `m.registry.managers` map, breaking encapsulation and risking data races if managers are ever registered dynamically.

**Fix**:
```go
// In MultiPackageResource.Actual, replace:
// for managerName := range m.registry.managers

// With:
for _, managerName := range m.registry.GetAllManagerNames() {
    manager, err := m.registry.GetManager(managerName)
    if err != nil {
        continue // or handle error appropriately
    }
    // ... rest of logic
}
```

## High-Priority Improvements

### 4. Remove self-install functionality for package managers
**Priority**: High (Security & Architecture)
**Files**:
- `internal/resources/packages/homebrew.go` (SelfInstall)
- `internal/resources/packages/uv.go` (SelfInstall)
- `internal/resources/packages/pnpm.go` (installViaStandaloneScript)
- `internal/resources/packages/cargo.go` (SelfInstall)
- `internal/resources/packages/pixi.go` (SelfInstall)
- All other managers implementing SelfInstaller interface

**Problem**: Self-install via shell scripts (`curl | sh`) poses security risks and adds complexity.

**New approach**:
- Package managers should only be installed via other supported package managers
- For managers with no installation path (like Homebrew itself), report in `plonk doctor` that manual installation is required
- Remove SelfInstaller interface or make it return "not supported" errors with helpful messages

**Benefits**:
- Eliminates remote script execution security risk
- Simpler, more maintainable code
- Clearer user expectations
- `plonk doctor` becomes the source of truth for missing prerequisites

**Example flow**:
```bash
# User tries to install cargo (Rust)
plonk install cargo
# Error: cargo package manager not found. Install via: brew install rust

# Or in doctor:
plonk doctor
# ✗ Homebrew not installed - manual installation required: https://brew.sh
# ✗ Cargo not found - install via: brew install rust
```

### 5. Standardize manager registration to V2
**Priority**: High (Architecture consistency)
**Files**:
- `internal/resources/packages/uv.go`
- `internal/resources/packages/goinstall.go`
- `internal/resources/packages/pixi.go`

**Problem**: Some managers still use V1 `RegisterManager` and don't accept injected executors, while most use V2. This reduces testability and creates inconsistency.

**Fix**: Update each to use `RegisterManagerV2` with `NewXxxManagerWithExecutor` constructors:
```go
// Example for uv.go:
func init() {
    RegisterManagerV2("uv", NewUvManagerWithExecutor)
}

func NewUvManagerWithExecutor(exec CommandExecutor) PackageManager {
    return &UvManager{executor: exec}
}
```

**Benefits**: Complete testability, consistent execution across all managers

### 5. Parallelize manager operations
**Priority**: Medium (Performance)
**File**: `internal/resources/packages/resource.go`

**Problem**: `MultiPackageResource.Actual` runs `IsAvailable` + `ListInstalled` for each manager sequentially, which is slow.

**Fix**: Use `errgroup.WithContext` to parallelize across managers with bounded concurrency:
```go
import "golang.org/x/sync/errgroup"

func (m *MultiPackageResource) Actual(ctx context.Context) ([]Item, error) {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(runtime.GOMAXPROCS(0)) // or fixed limit like 4

    mu := sync.Mutex{}
    var allItems []Item

    for _, managerName := range m.registry.GetAllManagerNames() {
        managerName := managerName // capture for goroutine
        g.Go(func() error {
            manager, _ := m.registry.GetManager(managerName)
            // ... IsAvailable + ListInstalled logic

            mu.Lock()
            allItems = append(allItems, items...)
            mu.Unlock()
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return allItems, nil
}
```

**Benefits**: Significant speedup for status/apply operations, especially with many managers

### 6. Switch Homebrew Info to JSON parsing
**Priority**: Medium (Robustness)
**File**: `internal/resources/packages/homebrew.go`

**Problem**: `Info` method scrapes plain text output, which is brittle and locale-dependent.

**Fix**: Use `brew info --json=v2 <name>` and parse structured JSON (similar to `getInstalledPackagesInfo`):
```go
func (h *HomebrewManager) Info(ctx context.Context, name string) (*PackageInfo, error) {
    output, err := h.executor.ExecuteWithTimeout(
        ctx, h.config.Timeout, "brew", "info", "--json=v2", name,
    )
    if err != nil {
        return nil, err
    }

    // Parse brewInfoV2JSON for both formulae and casks
    // Return structured PackageInfo
}
```

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

### 8. Fix duplicate listing of drifted dotfiles
**Priority**: Medium (UX bug)
**Command**: `plonk status`

**Problem**: When a dotfile is in drifted status, it appears twice in the output.

**Fix**: Ensure drifted files only appear once in the status output with their drift state clearly indicated.

### 9. Improve dotfile column headers in status
**Priority**: Medium (UX improvement)
**Command**: `plonk status`

**Problem**: Column headers "source" and "target" are confusing for dotfiles.

**Fix**:
- Use actual expanded paths like `/Users/username` and `/Users/username/.config/plonk` instead of generic "source/target"
- Reorder columns to: `$HOME`, `$PLONKDIR`, `STATUS`
- Makes it immediately clear which file is where

### 10. Fix diff output column ordering
**Priority**: Medium (UX improvement)
**Command**: `plonk diff`

**Problem**: Current ordering is inconsistent with user mental model.

**Fix**: Display `$HOME` (deployed location) on the left and `$PLONKDIR` (source) on the right, matching standard diff conventions where "original" is on left and "modified" is on right.

## Feature Additions

### 11. Add `plonk add -y` to sync drifted files back to $PLONKDIR
**Priority**: Medium (Feature enhancement)
**Command**: `plonk add`

**Feature**: Add a `-y` or `--sync-drifted` flag to automatically copy all drifted files from $HOME back to $PLONKDIR (reverse of `apply`).

**Use case**: When you've edited dotfiles in $HOME and want to quickly capture all changes back to your managed config directory.

**Behavior**:
```bash
plonk add -y
# Finds all files with "drifted" status
# Copies them from $HOME to $PLONKDIR
# Updates timestamps/hashes
```

**Implementation considerations**:
- Should show which files will be synced before copying (or require `-y` for non-interactive)
- Consider `--dry-run` option to preview changes
- Respect `.plonkignore` or similar patterns if implemented

### 12. Add selective file deployment to `plonk apply`
**Priority**: Medium (Feature enhancement)
**Command**: `plonk apply`

**Feature**: Allow `plonk apply <file1> <file2> ...` to selectively deploy only specified files from $PLONKDIR to $HOME.

**Use case**: When you've updated specific dotfiles and only want to deploy those without touching others.

**Behavior**:
```bash
plonk apply ~/.vimrc ~/.zshrc
# Only deploys vimrc and zshrc from $PLONKDIR to $HOME
# Leaves other managed dotfiles untouched
```

**Implementation considerations**:
- Accept both $HOME paths (`.vimrc`) and $PLONKDIR paths
- Validate files are actually managed by plonk
- Provide clear error if file not found or not managed
- Still show before/after status for specified files

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

### 16. Add comprehensive tests
**Recommended test additions**:
1. Self-install path tests per manager (verify correct installer commands via mock executor)
2. Concurrency smoke test for parallel manager reconciliation
3. Symlink traversal tests for dotfile operations

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

### Phase 1: Critical (Do First)
1. Fix pipx home directory bug (#1)
2. Fix MultiPackageResource concurrency (#3)
3. Remove self-install functionality (#4)

### Phase 2: UX Improvements
4. Fix duplicate drifted dotfiles in status (#8)
5. Improve dotfile column headers in status (#9)
6. Fix diff output column ordering (#10)
7. Add `plonk add -y` to sync drifted files (#11)
8. Add selective file deployment to `plonk apply` (#12)

### Phase 3: Architecture & Performance
9. Standardize V2 registration (#5)
10. Parallelize manager operations (#6)
11. Switch Homebrew to JSON (#7)

### Phase 4: Polish
12. Centralize package keying (#7)
13. Remove unused code (#13)
14. Add comprehensive tests (#16)
