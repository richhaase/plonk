# Phase 4: Idiomatic Go Simplification

## Objective
Achieve meaningful code reduction through idiomatic Go patterns and genuine simplification, avoiding abstractions that hide clarity.

## Timeline
Day 6-7 (16 hours)

## Current State
- ~14,300 LOC after Phase 3
- All 6 package managers functional and needed
- Resource abstraction in place for AI Lab
- Significant duplication identified in Phase 3.5 analysis

## Target State
- ~11,000-12,000 LOC (more realistic target)
- Clearer, more maintainable code
- Genuine simplifications, not hidden complexity
- Preserved Go idioms and explicit behavior

## Task Breakdown

### Task 4.1: Simplify Error Handling (3 hours)
**Agent Instructions:**
1. Remove error wrapping where it adds no value:
   - Find patterns like `fmt.Errorf("failed to X: %w", err)` where the wrapper adds no context
   - Return the original error directly
   - Keep wrapping only where it adds specific, actionable context

2. Simplify error checks:
   ```go
   // Before:
   if err != nil {
       return fmt.Errorf("npm list: %w", err)
   }

   // After (if context is obvious):
   if err != nil {
       return err
   }
   ```

3. Remove defensive error checks for impossible conditions:
   - Standard library functions that don't actually return errors in practice
   - Over-checking after successful operations

4. Expected reduction: 400-600 lines

5. Commit: "refactor: simplify error handling, remove redundant wrapping"

**Validation:**
- Error messages still useful
- No loss of debugging context where needed
- Tests still pass

### Task 4.2: Consolidate Package Manager Tests (4 hours)
**Agent Instructions:**
1. Identify truly redundant test cases across package managers:
   - Same test logic repeated for each manager
   - Mock setup that could be shared
   - Common assertion patterns

2. Create test helpers in each package (not shared):
   ```go
   // In homebrew_test.go
   func setupBrewTest(t *testing.T) (*BrewManager, *testutil.TempDir) {
       // Common setup for brew tests
   }
   ```

3. Reduce test redundancy WITHOUT creating test abstractions:
   - Keep tests explicit and readable
   - Share setup/teardown where obvious
   - Don't create generic test frameworks

4. Expected reduction: 800-1,200 lines

5. Commit: "test: reduce redundancy in package manager tests"

**Validation:**
- All tests still pass
- Tests remain readable and debuggable
- Coverage maintained

### Task 4.3: Simplify Dotfiles Package (3 hours)
**Agent Instructions:**
Based on Phase 3.5 findings, remove over-engineering:

1. Simplify path operations:
   - Use filepath.Clean() instead of custom validation
   - Remove redundant safety checks
   - Trust Go's standard library

2. Remove unnecessary atomic operations:
   ```go
   // If not truly needed for safety:
   // Before: Complex atomic write with temp file
   // After: ioutil.WriteFile or os.WriteFile
   ```

3. Consolidate file walking:
   - Use filepath.Walk or filepath.WalkDir consistently
   - Remove custom directory traversal code

4. Simplify types:
   - If multiple structs represent the same thing, use one
   - Remove unnecessary type aliases

5. Expected reduction: 400-600 lines

6. Commit: "refactor: simplify dotfiles package, trust standard library"

**Validation:**
- Dotfile operations still work correctly
- No race conditions introduced
- Tests pass

### Task 4.4: Merge Doctor into Status (2 hours)
**Agent Instructions:**
1. Add health check functionality to status command:
   ```go
   // In status.go
   if checkHealth {
       // Include doctor checks inline
   }
   ```

2. Move doctor-specific checks directly into status:
   - Don't create abstractions
   - Keep the logic visible and clear
   - Add --health or --check flag

3. Remove doctor.go entirely

4. Expected reduction: ~200 lines

5. Commit: "refactor: merge doctor functionality into status command"

**Validation:**
- `plonk status --health` provides doctor functionality
- No functionality lost

### Task 4.5: Remove Unused Code (2 hours)
**Agent Instructions:**
1. Use tools to find dead code:
   ```bash
   # Find unused functions/types
   staticcheck -checks U1000 ./...

   # Find unused struct fields
   staticcheck -checks U1001 ./...
   ```

2. Remove genuinely unused code:
   - Unused helper functions
   - Dead branches
   - Unused struct fields
   - Commented-out code

3. Be conservative - if unsure, keep it

4. Expected reduction: 200-400 lines

5. Commit: "cleanup: remove unused code identified by static analysis"

### Task 4.6: Inline Trivial Helpers (2 hours)
**Agent Instructions:**
1. Find single-use helper functions that add no value:
   ```go
   // Before:
   func isBrewInstalled() bool {
       return checkCommandExists("brew")
   }

   // After (at call site):
   if checkCommandExists("brew") {
   ```

2. Inline where it improves clarity:
   - Single-line functions used once
   - Wrappers that just rename
   - Getters that just return a field

3. Keep helpers that:
   - Are used multiple times
   - Encapsulate complex logic
   - Improve readability

4. Expected reduction: 300-500 lines

5. Commit: "refactor: inline trivial single-use helpers"

### Task 4.7: Final Cleanup and Validation (2 hours)
**Agent Instructions:**
1. Format and organize imports:
   ```bash
   gofmt -w -s ./...
   goimports -w ./...
   ```

2. Run linters and address issues:
   ```bash
   golangci-lint run
   ```

3. Verify no regressions:
   ```bash
   go test ./...
   just test-ux
   ```

4. Measure final LOC:
   ```bash
   find internal/ -name "*.go" -not -path "*/test/*" | xargs wc -l
   ```

5. Create summary report

6. Commit: "chore: final cleanup and formatting"

## Key Principles

1. **Explicit is better than clever** - Don't hide behavior behind abstractions
2. **Duplication is sometimes OK** - Especially in tests and command definitions
3. **Trust the standard library** - Don't wrap what already works well
4. **Keep related code together** - Avoid generic "common" packages
5. **Make the zero value useful** - Simplify initialization where possible

## What We're NOT Doing

- Creating builder patterns or factories
- Making generic "utility" packages
- Over-abstracting to reduce LOC
- Creating complex test frameworks
- Hiding flag definitions behind helpers

## Risk Mitigations

1. **Breaking Changes**: Test after each task
2. **Over-simplification**: Keep error context where valuable
3. **Readability**: Prefer clear code over clever code

## Success Criteria
- [ ] 2,000-3,000 LOC reduction achieved (realistic)
- [ ] All tests passing
- [ ] No new abstractions added
- [ ] Code is more idiomatic Go
- [ ] No loss of functionality
- [ ] Improved clarity and maintainability

## Notes for Agents
- When in doubt, choose explicit over implicit
- Don't create abstractions just to reduce line count
- Keep Go idioms: error handling, struct initialization, etc.
- Test frequently
- Small, atomic commits
