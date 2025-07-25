# Phase 7: Code Quality & Naming

## Objective
Improve code quality through consistent naming, removal of unused code, and addressing linter issues. Focus on maintainability and clarity.

## Timeline
Day 10 (8 hours)

## Current State
- ~13,800 LOC after Phase 6
- 8 packages (commands, config, lock, orchestrator, output, resources + sub-packages)
- Resource abstraction implemented and working
- Lock v2 and hooks fully integrated into CLI (Phase 6 completed this)
- Clean architectural boundaries established
- Import cycles resolved

## Important Context
Phase 6 successfully integrated the orchestrator, so all infrastructure is now connected and functional. Any "unused" code findings should be genuine dead code, not pending integrations.

## Task Breakdown

### Task 7.1: Identify and Remove Unused Code (2 hours)
**Agent Instructions:**
1. Run static analysis tools:
   ```bash
   # Find unused functions/types
   staticcheck -checks U1000 ./...

   # Find unused struct fields
   staticcheck -checks U1001 ./...

   # Check for unnecessary dependencies
   go mod why -m all | grep -B1 "# .* is not used"

   # Use deadcode tool for comprehensive analysis
   go install golang.org/x/tools/cmd/deadcode@latest
   deadcode ./...
   ```

2. Create a report categorizing findings:
   ```
   PHASE_7_UNUSED_CODE.md

   ## Safe to Remove
   - Functions never called anywhere
   - Old helper functions replaced by better versions
   - Test utilities no longer used
   - Unused struct fields
   - Dead constants and variables

   ## Keep - Exported API
   - Public functions that might be used externally
   - Interface methods (even if only one implementation)
   - Struct fields used in serialization (json/yaml tags)
   ```

3. Remove only genuinely dead code

4. Expected reduction: 100-300 lines

5. Commit: "cleanup: remove dead code identified by static analysis"

### Task 7.2: Standardize Function Naming (2 hours)
**Agent Instructions:**
1. Fix inconsistent getter patterns:
   ```go
   // Current mixed patterns in config package:
   GetHomeDir()          // has Get prefix
   GetConfigDir()        // has Get prefix
   LoadWithDefaults()    // no Get prefix

   // Standardize to idiomatic Go (no Get prefix):
   HomeDir()
   ConfigDir()
   LoadWithDefaults()    // already correct
   ```

2. Clarify ambiguous resource names:
   ```go
   // Current vague names:
   Item              // Too generic for a type
   Result            // Which kind of result?

   // Consider more specific names:
   ResourceItem      // If it's the main resource type
   ReconcileResult   // If it's reconciliation specific
   ```

3. Fix package/type stuttering:
   ```go
   // Avoid:
   resources.ResourceItem  // stutters

   // Prefer:
   resources.Item         // cleaner
   ```

4. Run on each package separately, commit per package

5. Expected changes: ~30-50 renamed items

**Validation**: Code still compiles and tests pass after each rename

### Task 7.3: Improve Variable and Parameter Names (1.5 hours)
**Agent Instructions:**
1. Replace vague names with specific ones:
   ```go
   // Bad:
   func Process(cfg *Config, mgr Manager, items []Item) error {
       for _, i := range items {

   // Better:
   func ProcessPackages(config *Config, pkgManager Manager, packages []Item) error {
       for _, pkg := range packages {
   ```

2. Fix abbreviated names (except well-known conventions):
   ```go
   // Keep these common abbreviations:
   ctx  → context.Context (standard)
   pkg  → package (common in Go)
   cfg  → config (acceptable if local)
   err  → error (standard)

   // Expand unclear abbreviations:
   mgr  → manager
   rsc  → resource
   svc  → service
   ```

3. Ensure consistency across codebase:
   - If using `package` in one place, don't use `pkg` elsewhere
   - If using `manager` in one place, don't use `mgr` elsewhere

4. Commit: "refactor: improve variable and parameter naming"

### Task 7.4: Fix Linter Issues (1.5 hours)
**Agent Instructions:**
1. Run comprehensive linting:
   ```bash
   # Run golangci-lint with most checks enabled
   golangci-lint run ./...

   # Run go vet
   go vet ./...

   # Check formatting
   gofmt -d -s .
   goimports -d .
   ```

2. Fix issues in priority order:
   - Correctness issues (must fix)
   - Error handling issues
   - Simplification opportunities
   - Style consistency

3. Common fixes needed:
   - Add missing error checks
   - Fix error strings (lowercase, no punctuation)
   - Simplify slice/map initialization
   - Remove redundant type conversions
   - Fix receiver names consistency

4. Skip pedantic rules that hurt readability

5. Commit fixes by category

### Task 7.5: Improve Documentation (1.5 hours)
**Agent Instructions:**
1. Ensure package documentation is clear:
   ```go
   // Package orchestrator coordinates resource operations across
   // package managers and dotfiles, providing the main sync
   // logic with hook support and lock file management.
   package orchestrator
   ```

2. Document all exported types and functions:
   ```go
   // Item represents a single managed resource (package or dotfile)
   // that can be in one of several states (managed, missing, untracked).
   type Item struct {
   ```

3. Add missing godoc comments focusing on:
   - Why something exists (not just what it does)
   - Important behavior or side effects
   - Usage examples for complex functions

4. Document important internal functions if behavior is non-obvious

5. Fix comment formatting:
   - Start with function/type name
   - Use proper sentences
   - Keep concise but complete

6. Commit: "docs: improve package and function documentation"

### Task 7.6: Final Code Quality Check (1 hour)
**Agent Instructions:**
1. Run final quality checks:
   ```bash
   # Cyclomatic complexity
   gocyclo -over 15 ./...

   # Cognitive complexity
   gocognit -over 20 ./...

   # Final test run
   go test ./...
   ```

2. Address any remaining issues:
   - Split complex functions if needed
   - Simplify deeply nested code
   - Extract helper functions for clarity

3. Measure current state:
   ```bash
   # Line count
   scc --include-lang Go internal/

   # Package structure
   find internal/ -type d -maxdepth 1 | wc -l
   ```

4. Create summary report:
   ```
   PHASE_7_SUMMARY.md
   - Dead code removed: X lines
   - Functions renamed: Y
   - Variables improved: Z
   - Linter issues fixed: N
   - Documentation added: M functions/types
   - Current LOC: ~X,XXX
   - All tests passing: ✓
   ```

5. Commit: "chore: final code quality improvements"

## Important Notes

### Naming Conventions
- Use idiomatic Go patterns (no Get prefix for getters)
- Be consistent throughout the codebase
- Prefer clarity over brevity
- Avoid stuttering (package.PackageThing)

### What NOT to Change
- Well-established names that would break compatibility
- Names that are clear and widely used in the codebase
- Test function names (they have specific patterns)

### Quality Over Metrics
- Don't rename just to rename
- Don't remove code that might be useful
- Focus on genuine improvements

## Success Criteria
- [ ] No genuinely dead code remains
- [ ] Naming is consistent across packages
- [ ] All exported APIs are documented
- [ ] Linter issues addressed (where sensible)
- [ ] Code is more readable and maintainable
- [ ] All tests still pass

## Risk Mitigation
- Test after each major change
- Keep commits focused and atomic
- Don't break public APIs
- Preserve working functionality

This phase focuses on polish and consistency. Every change should make the code easier to understand and maintain.
