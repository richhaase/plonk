# Plan: Remove Hook System from Plonk

**Generated:** 2025-09-02
**Implemented:** 2025-09-03
**Status:** ✅ COMPLETED
**Goal:** Removing the hook system from plonk

## Goal & Constraints
- **Goal**: Remove the hook system entirely from plonk to eliminate the security vulnerability identified in the assessment (command injection via arbitrary shell execution)
- **Constraints**:
  - Follow project development rules (exact scope only, no unnecessary features, prefer editing over creating)
  - Keep diffs small and focused
  - Update related documentation and tests
  - No system modification in unit tests

## Implementation Summary
All hook system components successfully removed:
- ✅ `/Users/rdh/src/plonk/internal/orchestrator/hooks.go` - HookRunner implementation (deleted)
- ✅ `/Users/rdh/src/plonk/internal/orchestrator/hooks_test.go` - Hook tests (deleted)
- ✅ `/Users/rdh/src/plonk/internal/orchestrator/coordinator.go:23,32,58,84` - Hook integration points (removed)
- ✅ `/Users/rdh/src/plonk/internal/config/config.go:23,27-38` - Hook configuration structs (removed)
- ✅ `/Users/rdh/src/plonk/internal/commands/config_edit.go:215-218` - Config editing for hooks (removed)
- ✅ `/Users/rdh/src/plonk/internal/config/user_defined_test.go` - Hook-related user config tests (removed)
- ✅ `/Users/rdh/src/plonk/internal/orchestrator/orchestrator_test.go:29,277` - Test references (removed)
- ✅ `/Users/rdh/src/plonk/internal/config/user_defined.go` - Hook validation rules (removed)

## Options & Trade-offs
**Recommendation:** Complete removal - eliminates security risk, simple deletion of unused speculative feature.

## Design Sketch (recommended)
- **Interfaces/contracts**: Remove `HookRunner` type, `Hook` config struct, `Hooks` config struct
- **Data/schema**: Remove `hooks` field from Config struct, eliminate hook validation
- **Error & boundaries**: Remove hook-related error handling in coordinator
- **Compatibility**: No breaking changes since nothing uses it

## Implementation Steps (completed)
1. ✅ **Remove hook execution from coordinator** - Eliminated hook runner calls in apply operations
2. ✅ **Delete hook implementation files** - Removed hooks.go and hooks_test.go
3. ✅ **Clean config structures** - Removed Hook/Hooks types from config package
4. ✅ **Update config editing** - Removed hook handling from config_edit command
5. ✅ **Clean orchestrator construction** - Removed hookRunner field and initialization
6. ✅ **Update tests** - Removed hook-related test assertions and mocks
7. ✅ **Validation cleanup** - Removed hook validation rules from user_defined.go

## Test Plan
- **Unit**: Verify orchestrator tests pass without hook runner (orchestrator_test.go), config tests validate without hooks (config_test.go)
- **Integration**: Run `just test-bats` to ensure CLI behavior unchanged
- **Regression**: No regression risk since feature unused
- **Performance**: No impact expected since removing unused functionality

## Observability & Rollout
- **Rollout strategy**: Direct deployment - no user impact since feature unused

## Docs & Comms
- **Decision Log**: "Removed unused hook system due to command injection security vulnerability identified in assessment"

## Risks & Mitigations
- **Risk**: Minimal - removing unused speculative code
- **Mitigation**: Standard testing validates no impact

## Backout Plan
Not needed - removing unused speculative feature with no production usage.

## Acceptance Checklist
- ✅ `just build` succeeds without hook-related symbols
- ✅ `just test` passes with all hook tests removed
- ✅ `just lint` passes without hook-related code
- ✅ No references to "hook" remain in codebase (only 3 unrelated Git hooks references)

## Results
- **Security vulnerability eliminated**: Command injection via arbitrary shell execution is no longer possible
- **No breaking changes**: Feature was unused, so no user impact
- **Clean codebase**: All hook-related code successfully removed
- **Tests passing**: All unit tests continue to pass after removal
- **Build successful**: Project builds without errors or warnings
