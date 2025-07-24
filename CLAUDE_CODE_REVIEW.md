# Comprehensive Code Review: Plonk CLI

## Executive Summary

Plonk is a package and dotfile manager that has accumulated significant complexity during rapid development. The codebase contains **22 packages** and **~26,000 lines of code** for what is fundamentally a wrapper around package managers and a file synchronization tool. This review proposes a **ruthless simplification**, reducing the codebase by **~70%** while maintaining all core user-facing functionality.

**Key Finding**: The codebase must be reduced from 22 packages to 5-6 packages, eliminating approximately 18,000 lines of code. Half-measures have failed - radical simplification is required.

**Critical Insight**: A simple operation like `plonk add ~/.vimrc` currently traverses 9 layers of abstraction. After refactoring, it will be ~30 lines of direct, readable code.

## Current State Analysis

### What Plonk Actually Does
1. **Package Management**: Wraps brew, npm, cargo, pip, gem, and go install commands
2. **Dotfile Management**: Copies files between home directory and config directory
3. **State Tracking**: Maintains YAML files (plonk.yaml, plonk.lock) to track what's managed
4. **Synchronization**: Ensures configured state matches actual state

### Package Sprawl (22 packages for simple operations)

```
internal/
├── cli/           (1 file)   - Simple helpers that belong in commands
├── commands/      (26 files) - Reasonable, but contains too much logic
├── config/        (5 files)  - Reasonable, but has dual config system
├── constants/     (2 files)  - Should be inlined where used
├── core/          (3 files)  - Thin abstraction over operations
├── dotfiles/      (11 files) - Reasonable core functionality
├── errors/        (2 files)  - Over-engineered error system
├── executor/      (1 file)   - Unnecessary wrapper around exec.Command
├── interfaces/    (4 files)  - Java-style interface definitions
├── lock/          (5 files)  - Reasonable, but over-abstracted
├── managers/      (24 files) - Reasonable, but inheritance pattern
├── mocks/         (2 files)  - Generated mocks add complexity
├── operations/    (4 files)  - Unnecessary abstraction layer
├── paths/         (5 files)  - Over-engineered path handling
├── runtime/       (3 files)  - Complex singleton with no benefit
├── services/      (2 files)  - Pure pass-through layer
├── state/         (5 files)  - Over-abstracted provider pattern
├── testing/       (1 file)   - Test helpers
├── types/         (1 file)   - Type aliases that add confusion
└── ui/            (2 files)  - Simple formatting helpers
```

### Tracing a Simple Operation

Let's trace `plonk add ~/.vimrc` through the current architecture:

1. `cmd/plonk/main.go` → Cobra command setup
2. `internal/commands/add.go` → Command handler (200+ lines)
3. `internal/runtime/context.go` → SharedContext singleton
4. `internal/core/dotfiles.go` → AddSingleDotfile
5. `internal/paths/resolver.go` → Path resolution
6. `internal/config/adapter.go` → Config adaptation
7. `internal/dotfiles/operations.go` → File operations
8. `internal/dotfiles/fileops.go` → Actual file copy
9. `internal/operations/types.go` → Result wrapping

**That's 9 layers for: copy file from A to B and remember we did it**

## Non-Idiomatic Go Patterns

### 1. Java-Style Getters (103 instances found)
```go
// internal/config/config.go:103-131
func (c *NewConfig) GetDefaultManager() string {
    return c.DefaultManager  // Just returning a public field!
}
```
**Fix**: Remove all getters, use direct field access

### 2. Result Types with Embedded Errors
```go
// internal/services/package_operations.go:40-44
type PackageResult struct {
    Name   string
    Status string
    Error  string  // Error as string in struct
}
```
**Fix**: Return `(result, error)` like idiomatic Go

### 3. Interfaces in Wrong Place
```go
// internal/interfaces/package_manager.go
type PackageManager interface {
    // 15+ methods defined before any implementation exists
}
```
**Fix**: Define interfaces where they're used (consumer side)

### 4. Unnecessary Abstraction Layers
```go
// The existence of these methods proves the abstraction failed:
func (sc *SharedContext) SimplifiedReconcileDotfiles(...)
func (sc *SharedContext) SimplifiedReconcilePackages(...)
```
**Fix**: Remove the complex versions entirely

### 5. Over-Engineered Error System
```go
type PlonkError struct {
    Code      ErrorCode
    Domain    Domain
    Severity  Severity
    Operation string
    Message   string
    // ... 10 more fields
}
```
**Fix**: Use simple error wrapping with context

### 6. Context Misuse
```go
// Storing contexts in structs
type contextWithCancel struct {
    ctx    context.Context
    cancel context.CancelFunc
}
```
**Fix**: Pass context as parameter

### 7. Inheritance via Embedding
```go
type NpmManager struct {
    *BaseManager  // Using embedding as inheritance
}
```
**Fix**: Use composition or simple functions

## Proposed Simplified Architecture

### Package Structure (5-6 packages total)
```
plonk/
├── cmd/
│   └── plonk/
│       └── main.go        # Entry point only
├── internal/
│   ├── cli/               # Command handlers and output formatting
│   │   ├── commands.go    # All command implementations
│   │   ├── output.go      # Human and JSON output
│   │   └── root.go        # Cobra setup
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config struct and loading
│   │   └── defaults.go    # Default values
│   ├── dotfiles/          # Dotfile operations
│   │   ├── dotfiles.go    # Add, remove, sync operations
│   │   └── backup.go      # Backup handling
│   ├── managers/          # Package managers
│   │   ├── manager.go     # Common interface (minimal)
│   │   ├── brew.go        # Homebrew implementation
│   │   ├── npm.go         # NPM implementation
│   │   └── ...            # Other managers
│   └── lock/              # Lock file management
│       └── lock.go        # Read/write plonk.lock
```

### Simplified Type System
```go
// config/config.go
type Config struct {
    DefaultManager    string
    OperationTimeout  time.Duration
    IgnorePatterns    []string
    ExpandDirectories []string
}

// No getters, direct field access:
timeout := config.OperationTimeout

// lock/lock.go
type Lock struct {
    Version  int
    Packages map[string][]Package
}

type Package struct {
    Name        string
    InstalledAt time.Time
}

// managers/manager.go
type Manager interface {
    Name() string
    Install(pkg string) error
    Remove(pkg string) error
    List() ([]string, error)
    IsInstalled(pkg string) bool
}

// No complex error types, just:
return fmt.Errorf("install %s: %w", pkg, err)
```

### Example: Simplified Add Command

**Before**: 200+ lines across 9 packages
```go
// internal/commands/add.go (current implementation)
func runAdd(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    sharedCtx := ctx.Value(common.SharedContextKey).(*runtime.SharedContext)
    cfg, err := sharedCtx.GetConfig()
    if err != nil {
        return err
    }
    provider := state.NewDotfileProvider(...)
    reconciler := operations.NewReconciler(...)
    // ... 150+ more lines of abstraction
}
```

**After**: ~30 lines of direct code
```go
// internal/cli/add.go (proposed implementation)
func runAdd(cmd *cobra.Command, args []string) error {
    for _, path := range args {
        src, err := filepath.Abs(path)
        if err != nil {
            return fmt.Errorf("resolve %s: %w", path, err)
        }

        dst := filepath.Join(configDir, "dotfiles", filepath.Base(src))
        if err := copyFile(src, dst); err != nil {
            return fmt.Errorf("copy %s: %w", path, err)
        }

        fmt.Printf("Added %s\n", filepath.Base(src))
    }
    return nil
}
```

### Example: Simplified Install Command
```go
func runInstall(cmd *cobra.Command, args []string) error {
    // Load config directly
    cfg, err := config.Load()
    if err != nil {
        return err
    }

    // Get manager directly
    mgr := managers.Get(cfg.DefaultManager)
    if mgr == nil {
        return fmt.Errorf("manager %s not found", cfg.DefaultManager)
    }

    // Install directly
    for _, pkg := range args {
        if err := mgr.Install(pkg); err != nil {
            return fmt.Errorf("install %s: %w", pkg, err)
        }

        // Update lock file directly
        lock.AddPackage(mgr.Name(), pkg)
    }

    return nil
}
```

### Example: Simplified Sync Command
```go
func runSync(cmd *cobra.Command, args []string) error {
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    // Sync packages
    lock := lock.Load()
    for mgrName, packages := range lock.Packages {
        mgr := managers.Get(mgrName)
        if mgr == nil {
            continue
        }

        installed := mgr.List()
        for _, pkg := range packages {
            if !contains(installed, pkg.Name) {
                if dryRun {
                    fmt.Printf("Would install %s\n", pkg.Name)
                } else {
                    mgr.Install(pkg.Name)
                }
            }
        }
    }

    // Sync dotfiles
    dotfiles, _ := filepath.Glob(filepath.Join(configDir, "dotfiles", "*"))
    for _, src := range dotfiles {
        dst := filepath.Join(homeDir, filepath.Base(src))
        if dryRun {
            fmt.Printf("Would copy %s to %s\n", src, dst)
        } else {
            copyFile(src, dst)
        }
    }

    return nil
}
```

## Specific Recommendations

### 1. Eliminate These Packages Entirely
- ✅ `cli` → merge 1 file into `commands` (COMPLETED 2025-07-24)
- ✅ `constants` → inline where used (COMPLETED 2025-07-24)
- `core` → merge into domain packages
- `executor` → use exec.Command directly
- `interfaces` → define interfaces where used
- `operations` → unnecessary abstraction
- `paths` → use filepath package directly
- `runtime` → eliminate SharedContext entirely
- `services` → pure pass-through, delete it
- ✅ `types` → confusing aliases, delete it (COMPLETED 2025-07-24)
- `mocks` → use simple test doubles instead

### 2. Simplify These Packages
- `config` → Remove dual config system, remove getters
- `errors` → Replace with standard error wrapping
- `state` → Remove provider pattern, use direct functions
- `managers` → Remove BaseManager inheritance
- `commands` → Move business logic to domain packages

### 3. Idiomatic Go Changes
- Remove all 103 getter methods
- Replace Result types with `(value, error)` returns
- Move interfaces to consumer packages
- Replace struct embedding with composition
- Use error wrapping: `fmt.Errorf("context: %w", err)`
- Remove the context pooling
- Replace complex mocks with simple test doubles

### 4. Feature Simplifications
- Merge `doctor` command into `status`
- Remove YAML output format (keep JSON only)
- Remove progress indicators
- Simplify error matching to basic checks
- Remove SharedContext caching (no performance benefit)
- Make dry-run only available for sync command

### 5. State Management Simplification
Current: Complex provider/reconciler pattern with 6 different item types
Proposed: Simple comparison of config files vs actual state

```go
// Current: Over-abstracted
provider := state.NewMultiManagerPackageProvider(sharedCtx)
configured := provider.GetConfiguredPackages()
actual := provider.GetActualPackages(ctx)
items := state.ReconcilePackages(configured, actual)

// Proposed: Direct and simple
lock := lock.Load()
for mgr, pkgs := range lock.Packages {
    installed := managers.Get(mgr).List()
    // Compare directly
}
```

## Migration Path

**Important Note**: This refactoring will be temporarily messy. The codebase will be harder to navigate during the transition, but this is necessary to expose and eliminate the unnecessary complexity.

### Two Migration Approaches

#### Approach A: "Rip and Replace" (Aggressive, 3 weeks)
Best for: Small teams, good test coverage, tolerance for temporary instability

#### Approach B: "Gradual Migration" (Conservative, 4-5 weeks)
Best for: Larger teams, limited test coverage, need for continuous stability

### Phase 1: De-layering (Week 1)

#### Approach A: Rip and Replace
**Day 1-2**: Delete packages where we have clear migration paths:
- `services` → calls move directly to domain packages
- `types` → replace with concrete types
- `cli` → merge single file into `commands`

**Critical**: Only delete packages where you can trace every usage and have a clear replacement strategy. Do NOT delete and hope for the best.

**Day 3-4**: Delete more complex packages:
- `runtime` → BUT FIRST map every SharedContext usage to its replacement
- `operations` → merge logic into calling packages
- `interfaces` → will be recreated where actually needed

#### Approach B: Gradual Migration
**Day 1-3**: Mark packages for deletion, redirect calls:
- Add deprecation comments
- Create temporary forwarding functions
- Update one command at a time to bypass deprecated layers

**Day 4-5**: Remove deprecated packages once all calls are redirected

### Phase 2: Consolidation and Restructuring (Week 2)
**Goal**: Create the new 5-package structure

1. **Create new structure** (both approaches):
   ```
   internal/
   ├── cli/       # Commands + output (merge commands + ui)
   ├── config/    # Config management (simplified)
   ├── dotfiles/  # Dotfile operations (merge dotfiles + paths)
   ├── lock/      # Lock file handling (simplified)
   └── managers/  # Package managers (remove inheritance)
   ```

2. **Migrate systematically**:
   - Start with leaf packages (no dependencies)
   - Move tests alongside code
   - Ensure each package compiles before moving to next

3. **Key consolidations**:
   - `paths` → merge into `dotfiles` (they're tightly coupled)
   - `errors` → replace with standard Go error wrapping
   - `state` → split between `lock` and direct implementations
   - Constants → move to packages that use them

### Phase 3: Testing and Refinement (Week 3)
**Goal**: Ensure quality and correct behavior

1. **Replace mock-based tests with real tests**:
   ```go
   // Instead of mocking file operations:
   func TestAdd(t *testing.T) {
       tempDir := t.TempDir()
       // Actually create files
       // Run the add command
       // Check files were copied correctly
   }
   ```

2. **Testing strategy**:
   - Test actual file system operations
   - Test actual command execution (exec.Command)
   - Use test fixtures for package manager outputs
   - Integration tests for full workflows

3. **Final validation**:
   - Every CLI command must be manually tested
   - Run against real dotfiles and packages
   - Verify config file compatibility
   - Performance should improve (fewer layers = faster)

### Risk Mitigation

1. **Before starting**:
   - Tag current version
   - Ensure CI/CD is working
   - Document current behavior with integration tests

2. **During refactoring**:
   - Commit after each successful phase
   - Keep a migration log of what was moved where
   - Run tests continuously

3. **Rollback plan**:
   - Each phase should be independently revertable
   - Keep the old packages until Phase 3 is complete
   - Maintain a mapping of old code locations to new

## Benefits of Simplification

1. **Reduced Cognitive Load**: 5 packages instead of 22
2. **Faster Development**: Less abstraction to navigate
3. **Better Performance**: Direct calls instead of 9 layers
4. **Easier Testing**: Simple functions instead of complex mocks
5. **More Idiomatic**: Follows Go community patterns
6. **AI-Friendly**: Simpler codebase easier for AI assistants

## Risks and Mitigations

1. **Risk**: Breaking existing configs
   **Mitigation**: Config loading remains backward compatible

2. **Risk**: Lost functionality
   **Mitigation**: Comprehensive CLI tests before refactoring

3. **Risk**: Performance regression
   **Mitigation**: The current caching provides no real benefit for a CLI

## Conclusion

The current codebase suffers from premature abstraction and non-idiomatic patterns that make it harder to understand and maintain than necessary. The proposed simplification would reduce the codebase by ~70% while maintaining all essential functionality and improving developer experience. The key insight is that Plonk is fundamentally simple: it copies files and runs package manager commands. The implementation should reflect this simplicity.

## Progress Tracking

### Completed Refactoring
1. **2025-07-24**: Merged `cli` package into `commands`
   - Moved 4 helper functions from `internal/cli/helpers.go` to `internal/commands/helpers.go`
   - Updated 7 files to remove cli imports
   - Package count: 22 → 21

2. **2025-07-24**: Deleted `types` package
   - Moved Result and Summary structs to `internal/state/types.go`
   - Removed type aliases (Item, ItemState)
   - Updated 5 files to use state package directly
   - Package count: 21 → 20

3. **2025-07-24**: Deleted `constants` package
   - Moved constants to their domain packages (config, lock, managers)
   - Updated 7 files to use domain-specific constants
   - Constants now co-located with their logic
   - Package count: 20 → 19

### Remaining Work
- 8 packages still to eliminate
- 5 packages to simplify
- ~18,000 lines of code to remove
