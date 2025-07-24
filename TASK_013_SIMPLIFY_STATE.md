# Task 013: Simplify State Package

## Objective
Dramatically simplify the state package by eliminating the over-abstracted provider pattern while preserving the essential Managed/Missing/Untracked reconciliation logic for AI Lab requirements.

## Quick Context
- **Current state package**: 1,011 LOC with complex provider abstractions
- **Analysis shows**: 60-70% reduction possible while keeping core functionality
- **Key preservation**: Managed/Missing/Untracked states for AI Lab extensibility
- **Major win**: Eliminate duplicate reconciliation logic and unnecessary abstractions

## Work Required

### Phase 1: Create Minimal State Types
1. **Create new minimal state/types.go** (~50-100 lines)
   - Keep only: `Item`, `ItemState`, `Result`, `Summary` types
   - Remove: Provider interfaces, ConfigItem/ActualItem split
   - Simplify: Flatten type hierarchy, reduce metadata complexity
2. **Move OperationResult types** to commands package where they're used

### Phase 2: Eliminate Provider Pattern
1. **Delete state/dotfile_provider.go** (275 lines)
2. **Create dotfiles/state.go** with simple functions:
   - `GetConfiguredDotfiles(homeDir, configDir string) ([]Item, error)`
   - `GetActualDotfiles(ctx context.Context, homeDir, configDir string) ([]Item, error)`
3. **Update managers package** to implement similar functions:
   - `GetConfiguredPackages(configDir string) ([]Item, error)`
   - `GetActualPackages(ctx context.Context) ([]Item, error)`
4. **Remove provider initialization** from all commands

### Phase 3: Unify Reconciliation Logic
1. **Create generic reconciliation in orchestrator/reconcile.go**:
   ```go
   func ReconcileItems(configured, actual []state.Item) []state.Item
   ```
2. **Delete duplicate implementations** in orchestrator
3. **Update ReconcileDotfiles/ReconcilePackages** to use generic function
4. **Simplify item creation** - pass domain-specific logic as needed

### Phase 4: Update All Consumers (15 files)
1. **Commands using providers** - update to call simple functions
2. **Orchestrator methods** - use new reconciliation approach
3. **UI formatters** - work with simplified types
4. **Tests** - update for new structure

## Implementation Strategy

### New Architecture
```
internal/
  state/
    types.go         # Core types only: Item, ItemState, Result (~50 lines)
  
  dotfiles/
    state.go         # GetConfigured/GetActual functions (~100 lines)
  
  managers/
    state.go         # Multi-manager state aggregation (~150 lines)
  
  orchestrator/
    reconcile.go     # Generic reconciliation + domain calls (~100 lines)
  
  commands/
    types.go         # OperationResult types for output (~64 lines)
```

### Migration Pattern Examples

**Before (with providers)**:
```go
// In command
cfg := config.LoadConfigWithDefaults(configDir)
dotfileConfigLoader := state.NewConfigBasedDotfileLoader(cfg.GetIgnorePatterns(), cfg.GetExpandDirectories())
provider := state.NewDotfileProvider(homeDir, configDir, dotfileConfigLoader)
configured, err := provider.GetConfiguredItems()
actual, err := provider.GetActualItems(ctx)
items := reconcileDotfileItems(configured, actual)
```

**After (direct functions)**:
```go
// In command
configured, err := dotfiles.GetConfiguredDotfiles(homeDir, configDir)
actual, err := dotfiles.GetActualDotfiles(ctx, homeDir, configDir)
items := orchestrator.ReconcileItems(configured, actual)
```

## Files to Update
**Major changes**: state/*, orchestrator/reconcile.go, dotfiles/*, managers/*
**Import updates**: add.go, rm.go, sync.go, status.go, doctor.go, info.go, ls.go, search.go, dotfiles.go, env.go, dotfile_operations.go, ui/progress.go, ui/format.go

## Expected Benefits
- **60-70% code reduction** in state package (1,011 → ~300-400 LOC)
- **Eliminate duplicate code** - single reconciliation implementation
- **Simpler mental model** - no provider abstraction to understand
- **Faster development** - direct function calls instead of provider setup
- **Preserved extensibility** - Managed/Missing/Untracked pattern intact

## AI Lab Preservation
✅ **Reconciliation semantics** - Managed/Missing/Untracked states preserved
✅ **Domain separation** - Dotfiles and packages remain distinct
✅ **Extensibility** - Easy to add new resource types with same pattern
✅ **Clean boundaries** - State types remain independent for future use

## Success Criteria
1. ✅ **State package reduced to core types only** (~50-100 lines)
2. ✅ **Provider pattern completely eliminated**
3. ✅ **Single generic reconciliation function**
4. ✅ **All 15 consumer files updated and working**
5. ✅ **Tests pass with simplified structure**
6. ✅ **60-70% code reduction achieved**
7. ✅ **AI Lab extensibility preserved**

## Dependencies
- **Can proceed after Task 012** completes (no direct config dependency)
- **Will affect many files** but changes are mechanical
- **No breaking CLI changes** - internal refactoring only

## Completion Report
Create `TASK_013_COMPLETION_REPORT.md` with:
- **Before/after architecture diagrams**
- **Code reduction metrics** per component
- **Migration examples** from each domain
- **Verification** of preserved AI Lab extensibility
- **Performance impact** (fewer abstraction layers)