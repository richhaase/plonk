# Task 019: Merge Paths into Dotfiles and Simplify

## Objective
Merge the `paths` package (1,067 LOC) into `dotfiles` package and review the combined result for duplication, non-idiomatic Go patterns, and opportunities for simplification.

## Background
The `paths` package is tightly coupled with `dotfiles` - it exists primarily to support dotfile operations. After our successful simplification of other packages, we should:
1. Eliminate the artificial boundary between these packages
2. Review the combined code for further simplification opportunities
3. Ensure the result follows idiomatic Go patterns

## Current State
- `paths` package: 1,067 LOC (7.3% of codebase)
  - resolver.go: 204 LOC
  - validator.go: 114 LOC
  - Test files: 809 LOC
- `dotfiles` package: 2,245 LOC (15.4% of codebase)
  - Already a core domain, reasonable size
  - May have duplication or complexity to address

## Scope of Work

### Phase 1: Analyze Dependencies
1. Map all imports from dotfiles → paths
2. Map all imports from other packages → paths
3. Identify any circular dependency risks
4. Document which path functionality is dotfile-specific vs general-purpose

### Phase 2: Merge Implementation
1. Move path resolution logic into dotfiles package
2. Move path validation logic into dotfiles package
3. Update all imports across the codebase
4. Ensure tests continue to pass after each file move
5. Delete the empty paths package

### Phase 3: Review and Simplify Combined Package
1. **Duplication Analysis**
   - Look for similar validation patterns
   - Find repeated error handling
   - Identify overlapping functionality

2. **Idiomaticity Review**
   - Remove any Java-style patterns
   - Ensure errors follow Go conventions
   - Check for unnecessary abstractions
   - Verify interface placement (consumer-side)

3. **Simplification Opportunities**
   - Can resolver and validator be combined?
   - Are all validation rules necessary?
   - Can we reduce the number of types?
   - Is the symlink handling over-engineered?

### Phase 4: Specific Areas to Investigate

1. **Path Resolution Complexity**
   ```go
   // Current: Complex resolution with many edge cases
   // Review if all cases are actually needed
   ```

2. **Validation Rules**
   ```go
   // Are we over-validating?
   // Which rules provide real value vs defensive programming?
   ```

3. **Test Coverage**
   - 809 LOC of tests for 305 LOC of implementation (2.6:1 ratio)
   - Review if all test cases provide value
   - Consolidate redundant test scenarios

4. **Error Handling**
   - Standardize error messages
   - Use simple error wrapping
   - Remove any custom error types

## Expected Outcomes
1. Package count: 9 → 8 (achieving our 7-9 package target)
2. Combined dotfiles package: ~3,300 LOC initially
3. After simplification: Target ~2,500-2,800 LOC (20-25% reduction)
4. Cleaner architecture with no artificial boundaries
5. More maintainable code following Go idioms

## Success Criteria
- [ ] All tests pass after merge
- [ ] No imports of paths package remain
- [ ] Combined package has clear, single responsibility
- [ ] Code follows Go idioms throughout
- [ ] At least 20% reduction in combined LOC
- [ ] No loss of functionality
- [ ] Improved code clarity

## Implementation Notes
1. Move files incrementally to avoid breaking changes
2. Run tests after each file move
3. Update imports in small batches
4. Keep commit history clean for easy rollback
5. Document any behavioral changes discovered

## Risk Mitigation
- Review orchestrator package for any hidden dependencies on paths
- Ensure config loading doesn't depend on path validation
- Check if any commands directly use paths package
- Preserve all security-related path validations

## Task Priority
HIGH - This is the final structural change to achieve our target architecture of 7-9 well-defined domain packages.
