# Phase 3: Simplification & Edge-case Fixes

## Objective
Aggressively simplify the codebase by removing abstractions, flattening implementations, and updating outdated code comments to reflect the new architecture.

## Timeline
Day 4-5 + ½ day buffer (20 hours total)

## Current State
- Resource abstraction implemented and working
- Orchestrator using generic Resources
- ~14,800 LOC (temporary increase from abstractions)
- Ready for aggressive simplification

## Target State
- StandardManager abstraction removed
- Manager implementations flattened and simplified
- State types simplified to single Item struct
- ErrorMatcher patterns removed
- Table output completed with tabwriter
- All code comments accurate and helpful

## Task Breakdown

### Task 3.1: Remove StandardManager Abstraction (3 hours)
**Agent Instructions:**
1. Analyze `internal/resources/packages/constructor.go` and identify StandardManager usage
2. For each package manager (homebrew, npm, pip, gem, cargo, goinstall):
   - Remove embedding of StandardManager
   - Copy only the methods actually used by each manager
   - Inline simple methods directly
   - Remove unused methods
3. Create `internal/resources/packages/helpers.go` with 3-4 truly shared functions:
   - Common command execution helpers
   - Version parsing utilities (if shared)
   - Error formatting helpers (if shared)
4. Delete constructor.go and any StandardManager-related code
5. Run tests to ensure no regression: `go test ./internal/resources/packages/...`
6. Commit: "refactor: remove StandardManager abstraction, flatten implementations"

**Expected simplifications:**
- Each manager should be self-contained
- No inheritance patterns
- Direct, readable code

**Validation:**
- All package manager tests pass
- No references to StandardManager remain
- Code is more direct and readable

### Task 3.2: Flatten Manager Implementations (4 hours)
**Agent Instructions:**
1. For each package manager file:
   - Remove unnecessary interfaces
   - Inline single-use helper functions
   - Simplify error handling (remove ErrorMatcher usage)
   - Use direct exec.Command calls
   - Remove defensive programming where Go's type system suffices
2. Specific simplifications:
   - Replace complex option structs with direct parameters
   - Remove builder patterns
   - Flatten nested error checks
   - Use simple string matching instead of error patterns
3. Ensure each manager is <300 LOC (currently ~600 each)
4. Run tests after each manager: `go test ./internal/resources/packages/... -run TestHomebrew`
5. Commit each manager separately for clean history:
   - "refactor: simplify homebrew manager implementation"
   - "refactor: simplify npm manager implementation"
   - etc.

**Validation:**
- Each manager file significantly smaller
- Tests still pass
- Code is more direct and obvious

### Task 3.3: Remove Error Matcher System (2 hours)
**Agent Instructions:**
1. Find all ErrorMatcher usage:
   ```bash
   grep -r "ErrorMatcher" --include="*.go" .
   grep -r "error.*pattern" --include="*.go" .
   ```
2. Replace ErrorMatcher patterns with simple error checks:
   ```go
   // Before:
   if matcher.Match(err) == ErrorTypeNotFound {

   // After:
   if strings.Contains(err.Error(), "not found") {
   ```
3. Delete `internal/resources/packages/error_matcher.go`
4. Delete `internal/resources/packages/error_matcher_test.go`
5. Update all error handling to use:
   - `errors.Is()` for known errors
   - `strings.Contains()` for error message matching
   - Direct error returns with context
6. Run all tests to verify error handling still works
7. Commit: "refactor: remove ErrorMatcher system, use simple error checks"

**Validation:**
- No ErrorMatcher references remain
- Error handling still functions correctly
- Tests pass

### Task 3.4: Simplify State Types (2 hours)
**Agent Instructions:**
1. Review `internal/resources/types.go`
2. Ensure Item struct only has essential fields:
   ```go
   type Item struct {
       Name  string
       Type  string
       State string
       Error error
       Meta  map[string]string
   }
   ```
3. Remove any complex methods on state types
4. Move any business logic to the resources that use it
5. Simplify any result types - prefer (value, error) returns
6. Update all usages to work with simplified types
7. Run tests: `go test ./...`
8. Commit: "refactor: simplify state types to essential fields only"

**Validation:**
- State types are simple data containers
- No business logic in types
- All tests pass

### Task 3.5: Complete Table Output Implementation (3 hours)
**Agent Instructions:**
1. Review `internal/output/tables.go`
2. Complete the table output implementation using `text/tabwriter`:
   ```go
   import "text/tabwriter"

   func WriteTable(w io.Writer, headers []string, rows [][]string) {
       tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
       defer tw.Flush()

       // Write headers
       fmt.Fprintln(tw, strings.Join(headers, "\t"))

       // Write rows
       for _, row := range rows {
           fmt.Fprintln(tw, strings.Join(row, "\t"))
       }
   }
   ```
3. Remove any complex table formatting logic
4. Ensure consistent output across all commands
5. Keep implementation under 250 LOC total
6. Test with various terminal widths
7. Commit: "feat: complete table output with simple tabwriter implementation"

**Validation:**
- Table output works for all commands
- Consistent formatting
- Simple implementation

### Task 3.6: Update Code Comments (4 hours)
**Agent Instructions:**
1. Search for outdated comments referencing deleted packages:
   ```bash
   grep -r "state\." --include="*.go" . | grep "//"
   grep -r "managers\." --include="*.go" . | grep "//"
   grep -r "SharedContext" --include="*.go" . | grep "//"
   grep -r "TODO" --include="*.go" .
   ```
2. For each file with outdated comments:
   - Update package references to new structure
   - Remove comments that describe the obvious
   - Ensure comments explain "why" not "what"
   - Update any architectural comments
   - Remove completed TODOs
   - Update interface documentation
3. Focus on high-value comments:
   - Complex algorithms
   - Non-obvious design decisions
   - Important constraints
   - Public API documentation
4. Remove noise comments like:
   ```go
   // GetName returns the name  <- Remove this
   func GetName() string {
   ```
5. Commit changes by package:
   - "docs: update comments in commands package"
   - "docs: update comments in resources package"
   - etc.

**Validation:**
- No references to deleted packages in comments
- Comments are accurate and helpful
- No obvious/redundant comments

### Task 3.7: Edge Case Fixes and Final Cleanup (2 hours)
**Agent Instructions:**
1. Run comprehensive tests and note any failures:
   ```bash
   go test ./... -v | grep -i "fail\|error"
   ```
2. Fix any edge cases discovered:
   - Path handling issues
   - Empty string handling
   - Nil pointer checks
   - Race conditions
3. Run linting and fix issues:
   ```bash
   golangci-lint run --fix
   go fmt ./...
   ```
4. Remove any unused imports or variables
5. Ensure all files follow Go conventions
6. Final test run: `go test ./...`
7. Commit: "fix: address edge cases and final cleanup"

**Validation:**
- All tests pass
- No linting warnings
- Clean, idiomatic code

## Risk Mitigations

1. **Breaking Changes**: Each simplification could break functionality
   - Run tests after every change
   - Commit frequently for easy rollback

2. **Over-simplification**: Removing too much could hurt maintainability
   - Keep truly shared code in helpers.go
   - Preserve necessary abstractions

3. **Comment Accuracy**: Updated comments might not match implementation
   - Review comments in context
   - Test examples in comments

## Success Criteria
- [ ] StandardManager abstraction completely removed
- [ ] Each manager implementation <300 LOC
- [ ] ErrorMatcher system removed
- [ ] State types simplified to data containers
- [ ] Table output completed with tabwriter
- [ ] All comments accurate and helpful
- [ ] All tests passing
- [ ] Significant LOC reduction achieved

## Notes for Agents
- Be aggressive with simplification
- When in doubt, choose the simpler approach
- Preserve functionality but not complexity
- Test frequently
- Keep commits atomic and well-described
- Focus on readability over cleverness
