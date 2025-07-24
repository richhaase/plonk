# Task 008: Transform Runtime to Orchestrator Package

## Objective
Transform the `runtime` package into a minimal `orchestrator` package (~200-300 LOC) that preserves essential coordination functionality while eliminating complex caching and singleton patterns.

## Quick Context
- Current `runtime` package: 370+ LOC with complex SharedContext singleton, caching, context pooling
- Over-engineered for a CLI tool with singleton pattern that provides no real benefit
- Contains essential reconciliation logic (Managed/Missing/Untracked) that must be preserved for AI Lab extensibility
- Logging functionality should be retained but simplified

## Current Issues with Runtime Package
1. **Singleton SharedContext**: Anti-pattern for a CLI tool, adds unnecessary complexity
2. **Caching**: Manager availability cache and config caching provide no performance benefit for CLI
3. **Context Pooling**: Unnecessary optimization that adds complexity
4. **Dual Methods**: Both complex and "Simplified" versions exist - the simplified ones prove the complex ones are wrong

## Essential Functions to Preserve
### Must Keep (Core Orchestration):
- `SimplifiedReconcileDotfiles()` - Core reconciliation logic for dotfiles
- `SimplifiedReconcilePackages()` - Core reconciliation logic for packages
- `SimplifiedReconcileAll()` - Coordinates all domain reconciliation
- `reconcileDotfileItems()` - Core Managed/Missing/Untracked logic
- `reconcilePackageItems()` - Core Managed/Missing/Untracked logic

### Must Keep (Utilities):
- All logging functions (Error, Warn, Info, Debug, Trace) and configuration
- Directory path helpers (HomeDir, ConfigDir access)

### Must Remove (Over-Engineering):
- SharedContext singleton pattern
- Manager availability caching (`IsManagerAvailable`, `InvalidateManagerCache`)
- Context pooling (`AcquireContext`, `ReleaseContext`, `contextWithCancel`)
- Config caching (`Config()`, `ConfigWithDefaults()`, `InvalidateConfig`)
- Complex reconciliation methods that just call simplified versions

## Target Orchestrator Package Structure
```go
// orchestrator/reconcile.go - Core reconciliation logic (~100 LOC)
func ReconcileDotfiles(ctx context.Context, homeDir, configDir string) (state.Result, error)
func ReconcilePackages(ctx context.Context, configDir string) (state.Result, error)
func ReconcileAll(ctx context.Context, homeDir, configDir string) (map[string]state.Result, error)

// orchestrator/paths.go - Simple path utilities (~30 LOC)
func GetHomeDir() string
func GetConfigDir() string

// orchestrator/logging.go - Existing logging (keep as-is, ~140 LOC)
// All existing logging functions and configuration
```

## Work Required

### Phase 1: Create New Orchestrator Package
1. Create `internal/orchestrator/` directory
2. Move logging.go as-is to `orchestrator/logging.go`
3. Create simplified reconciliation functions without SharedContext
4. Create simple path utility functions

### Phase 2: Update All Consumers
Update all 13 command files to use orchestrator instead of runtime:
- `add.go`, `rm.go`, `sync.go`, `install.go`, `uninstall.go`
- `doctor.go`, `status.go`, `ls.go`, `env.go`, `info.go`, `search.go`
- `helpers.go`, `shared.go`

Changes needed:
```go
// Before:
sharedCtx := runtime.GetSharedContext()
homeDir := sharedCtx.HomeDir()
result, err := sharedCtx.SimplifiedReconcileDotfiles(ctx)

// After:
homeDir := orchestrator.GetHomeDir()
configDir := orchestrator.GetConfigDir()
result, err := orchestrator.ReconcileDotfiles(ctx, homeDir, configDir)
```

### Phase 3: Remove Runtime Package
1. Delete `internal/runtime/` directory completely
2. Verify no remaining imports
3. Run full test suite

## Key Architectural Changes
1. **No Singleton**: Direct function calls instead of SharedContext methods
2. **No Caching**: Create fresh instances when needed (CLI doesn't need caching)
3. **Explicit Parameters**: Pass homeDir/configDir as parameters instead of hidden state
4. **Preserve Reconciliation**: Keep exact Managed/Missing/Untracked logic for AI Lab compatibility

## Expected Code Reduction
- **Before**: 370+ LOC across 3 files with complex patterns
- **After**: ~270 LOC across 3 files with simple functions
- **Net Reduction**: ~100+ LOC eliminated
- **Package Count**: 13 → 13 (transformation, not deletion)

## Success Criteria
1. ✅ All command functions work identically
2. ✅ Reconciliation logic preserves Managed/Missing/Untracked semantics
3. ✅ Logging functionality unchanged
4. ✅ No performance regression (should be faster without caching overhead)
5. ✅ All tests pass: `go test ./...` and `just test-ux`
6. ✅ No remaining `runtime` package imports

## Critical Preservation Points
- **Reconciliation Semantics**: Managed/Missing/Untracked pattern must work identically
- **Logging**: All debug/trace functionality preserved
- **AI Lab Readiness**: Clean interfaces for future extensions

## Completion Report
Create `TASK_008_COMPLETION_REPORT.md` with:
- Summary of transformation approach
- List of all files modified
- Before/after code metrics
- Test results confirmation
- Verification that reconciliation semantics are preserved
