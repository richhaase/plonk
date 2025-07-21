# Bug Fix Action Plan

## Executive Summary

Based on our comprehensive testing, we've identified 7 bugs ranging from simple UI fixes to structural issues. This plan prioritizes fixes based on user impact, implementation complexity, and architectural integrity.

## Bug Categorization

### Quick Wins (1-2 hours each)
These can be fixed immediately without architectural changes:

1. **Bug #2: Error Symbol for "Already Managed"**
   - **Impact**: High (confusing UX)
   - **Fix**: Change symbol from `✗` to `ℹ` or `⚠` in status output
   - **Location**: Likely in `internal/operations/` or command output formatting
   - **Risk**: None - purely cosmetic

2. **Bug #5: Uninstall Summary Missing Count**
   - **Impact**: Medium (incomplete feedback)
   - **Fix**: Add "removed" count to summary line
   - **Location**: `internal/operations/types.go` - extend OperationResult
   - **Risk**: None - additive change only

3. **Bug #7: Unhelpful Unavailable Manager Error**
   - **Impact**: Medium (poor UX)
   - **Fix**: Add OS-aware error messages for unavailable managers
   - **Location**: Command flag registration or error handling
   - **Risk**: None - improves error messaging

### Medium Complexity (4-6 hours each)
Require careful implementation but no major restructuring:

4. **Bug #3: Info Shows Wrong Manager**
   - **Impact**: High (incorrect information)
   - **Fix**: Check lock file first, then fall back to default manager
   - **Location**: `internal/commands/info.go`
   - **Approach**:
     - Read lock file to find which manager installed the package
     - Only query default manager if not in lock file
   - **Risk**: Low - follows existing patterns

5. **Bug #6: Lock File Not Updated on Uninstall**
   - **Impact**: Critical (state inconsistency)
   - **Fix**: Ensure uninstall operation updates lock file
   - **Location**: `internal/state/` or uninstall command
   - **Approach**:
     - Verify state reconciliation is called after uninstall
     - Ensure lock file writer removes uninstalled packages
   - **Risk**: Medium - needs careful testing

### Structural Issues (1-2 days each)
Require deeper investigation and potential refactoring:

6. **Bug #4: Gem Manager Completely Broken**
   - **Impact**: Critical (feature unavailable)
   - **Fix**: Debug gem command execution and error handling
   - **Location**: `internal/managers/gem.go`
   - **Investigation needed**:
     - Command syntax differences
     - Error detection patterns
     - Version parsing issues
   - **Risk**: Medium - isolated to one manager

7. **Bug #1: Doctor Platform-Unaware Suggestions**
   - **Impact**: Low (misleading advice)
   - **Fix**: Make suggestions OS-aware
   - **Location**: `internal/commands/doctor.go`
   - **Approach**:
     - Add OS compatibility map for managers
     - Filter suggestions based on current platform
   - **Risk**: Low - additive logic

## Recommended Implementation Order

### Phase 1: Quick Wins (Day 1)
Start with high-impact, low-risk fixes:
1. Fix Bug #2 (error symbol) - immediate UX improvement
2. Fix Bug #5 (uninstall summary) - complete the feedback loop
3. Fix Bug #7 (unavailable manager error) - better error experience

### Phase 2: State Management (Day 2)
Fix critical state consistency issues:
1. Fix Bug #6 (lock file updates) - critical for correctness
2. Fix Bug #3 (info command) - depends on correct lock file

### Phase 3: Manager-Specific (Day 3-4)
1. Fix Bug #4 (gem manager) - restore functionality
2. Fix Bug #1 (doctor suggestions) - polish

## Testing Strategy

After each fix:
1. Run targeted integration test: `go test -tags=integration ./tests/integration -run TestCompleteUserExperience/AllPackageManagers/{manager}`
2. Run manual verification for the specific scenario
3. Ensure no regressions in other managers

## Architecture Preservation

### Principles to Maintain:
1. **State Reconciliation**: All changes must go through the reconciler
2. **Error Handling**: Use structured errors with proper domains
3. **Manager Interface**: Don't break the PackageManager contract
4. **Lock File Format**: Maintain backward compatibility

### Patterns to Follow:
- Use existing error wrapping: `errors.Wrap()`
- Follow operation result patterns in `internal/operations/`
- Maintain manager abstraction boundaries
- Keep UI formatting separate from business logic

## Additional Discoveries

### Go Package Naming Issue
- Lock file stores "goimports" but searches use full path
- Needs consistent normalization strategy
- Consider separate fix or document as known limitation

### Cargo Uninstall Failure
- Requires investigation - may be environment-specific
- Add to backlog if not consistently reproducible

## Success Metrics

1. All integration tests pass
2. Manual testing confirms fixes
3. No regressions in other commands
4. Code follows existing patterns
5. Performance unchanged

## Next Steps

1. Start with Bug #2 as proof of concept
2. Set up branch for bug fixes: `git checkout -b fix/integration-test-bugs`
3. Fix one bug per commit for easy review
4. Update integration tests to remove TODO comments as bugs are fixed
