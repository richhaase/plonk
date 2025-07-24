# Task 012: Simplify Config Package

## Objective
Eliminate the dual config system and Java-style patterns in the config package, reducing code by 65-70% (593 → 150-200 LOC) while preserving all functionality.

## Quick Context
- **Current**: Dual OldConfig/NewConfig system with 15 getters and 8 loading functions
- **Analysis**: TASK_011_CONFIG_ANALYSIS_REPORT.md shows migration debt can be eliminated
- **Impact**: 18 dependent files, but zero breaking changes to CLI or config files

## Work Required

### Phase 1: Eliminate OldConfig System
1. **Update all 18 consumers** to use `NewConfig` instead of `OldConfig`
2. **Remove pointer-based resolution** (`cfg.Resolve()` pattern)
3. **Delete old_config.go** entirely (118 LOC eliminated)
4. **Update type aliases** (`Config = NewConfig`)

### Phase 2: Remove Adapter Patterns  
1. **Replace ConfigAdapter usage** with direct config access
2. **Remove StateDotfileConfigAdapter** by moving interface to state package
3. **Simplify orchestrator config usage**
4. **Update state.DotfileProvider** to accept `*Config` directly

### Phase 3: Consolidate Loading Functions
1. **Keep only 2 loading functions**: `Load()` and `LoadWithDefaults()`
2. **Remove 6 compatibility aliases** and old system functions
3. **Remove ConfigManager pattern** → inline CRUD operations
4. **Rename NewConfig functions** to standard names

### Phase 4: Remove Java-Style Getters
1. **Update all cfg.GetX() calls** → direct field access `cfg.X`
2. **Delete all 15 getter methods**
3. **Update tests** for direct field access

## Files to Update (18 total)
**Commands**: add.go, config_edit.go, config_show.go, config_validate.go, doctor.go, dotfile_operations.go, info.go, install.go, search.go, status.go, sync.go, env.go, rm.go, uninstall.go
**Orchestrator**: reconcile.go, paths.go  
**Tests**: zero_config_test.go

## Preservation Requirements
- ✅ **YAML/JSON output** for automation (`config show`, `config validate`)
- ✅ **Zero-config behavior** (LoadWithDefaults functionality)
- ✅ **Validation functionality** (SimpleValidator)
- ✅ **CLI interface** (no command/flag changes)
- ✅ **Config file format** (same YAML structure)

## Expected Benefits
- **65-70% code reduction**: 593 → 150-200 LOC
- **Idiomatic Go patterns**: Direct field access instead of getters
- **Simplified architecture**: Single config type, 2 loading functions
- **Reduced cognitive load**: No dual system confusion
- **Faster development**: Less abstraction to navigate

## Success Criteria
1. ✅ **OldConfig system completely removed**
2. ✅ **All 15 getter methods eliminated** 
3. ✅ **8 loading functions → 2 functions**
4. ✅ **All 18 dependent files updated and compiling**
5. ✅ **YAML/JSON output preserved**
6. ✅ **All tests passing** (`go test ./...` and `just test-ux`)
7. ✅ **Zero CLI interface changes**

## Dependencies
- **No conflicts with Task 010**: Config and errors packages are independent
- **Can proceed immediately**: Analysis complete, plan detailed

## Completion Report
Create `TASK_012_COMPLETION_REPORT.md` with:
- **Before/after architecture comparison**
- **Code reduction metrics** (LOC eliminated per phase)
- **Breaking change verification** (should be zero)
- **Performance impact** (reduced abstraction layers)
- **Migration pattern examples** from each phase