# Phase 6: Code Quality & Naming

## Objective
Improve code quality through consistent naming, removal of unused code, and addressing linter issues. Focus on maintainability over metrics.

## Timeline
Day 9 (8 hours)

## Current State
- ~13,826 LOC after Phase 5
- 8 packages (down from 9)
- Resource abstraction implemented
- Lock v2 and hooks implemented but NOT integrated into CLI
- Good architectural foundation, needs polish

## Important Context
**The lock v2 and hook systems from Phase 5 are not yet wired to the CLI commands.** Static analysis tools will correctly identify some of this code as "unused" because:
- The new `Orchestrator.Sync()` method exists but isn't called by `cmd/sync.go`
- Hook execution code is implemented but not reachable from user commands
- Lock v2 migration code exists but current commands still create v1 locks

**This is intentional** - integration is deferred to Phase 7 to avoid rework. When you see "unused" warnings for these systems, note them but don't remove the code.

## Task Breakdown

### Task 6.1: Identify Unused Code (2 hours)
**Agent Instructions:**
1. Run static analysis tools:
   ```bash
   # Find unused functions/types
   staticcheck -checks U1000 ./...

   # Find unused struct fields
   staticcheck -checks U1001 ./...

   # Check for unnecessary dependencies
   go mod why -m all | grep -B1 "# .* is not used"
   ```

2. Create a report categorizing findings:
   ```
   PHASE_6_UNUSED_CODE.md

   ## Safe to Remove
   - Functions never called anywhere
   - Old helper functions replaced by better versions
   - Test utilities no longer used

   ## Keep - Integration Pending (Phase 7)
   - orchestrator.Sync() - new method pending CLI integration
   - HookRunner methods - pending CLI integration
   - Lock v2 migration - pending CLI integration

   ## Keep - Exported API
   - Public functions that might be used by external code
   - Interface methods (even if only one implementation)
   ```

3. Remove only the "Safe to Remove" category

4. Expected reduction: 200-500 lines

5. Commit: "cleanup: remove genuinely unused code"

**Important**: If unsure whether something is part of Phase 5's pending integration, keep it.

### Task 6.2: Standardize Function Naming (2 hours)
**Agent Instructions:**
1. Fix inconsistent getter patterns:
   ```go
   // Inconsistent:
   GetHomeDir()        // has Get prefix
   ConfigDir()         // no Get prefix

   // Choose one pattern (prefer no prefix in Go):
   HomeDir()
   ConfigDir()
   ```

2. Fix manager/manager confusion:
   ```go
   // Confusing:
   type Manager interface {}     // package interface
   type BrewManager struct {}    // specific implementation
   manager := managers.Get()     // which kind of manager?

   // Clearer:
   type PackageManager interface {}
   type BrewManager struct {}
   mgr := packages.GetManager()
   ```

3. Clarify ambiguous names:
   - `Process()` → `ProcessPackages()` or `ProcessDotfiles()`
   - `Load()` → `LoadConfig()` or `LoadLock()`
   - `items` → `packages`, `dotfiles`, or `resources`

4. Run on each package separately, commit per package

5. Expected changes: ~50-100 renamed items

**Validation**: Code still compiles and tests pass after each rename

### Task 6.3: Improve Variable Names (1.5 hours)
**Agent Instructions:**
1. Replace single-letter names (except in obvious loops):
   ```go
   // Bad:
   func Process(c *Config, m Manager, p []Package) error {
       for _, i := range p {

   // Better:
   func Process(config *Config, pkgManager Manager, packages []Package) error {
       for _, pkg := range packages {
   ```

2. Replace generic names with specific ones:
   - `data` → `configData`, `lockData`, `packageData`
   - `result` → `syncResult`, `applyResult`
   - `info` → `packageInfo`, `systemInfo`
   - `item` → `package`, `dotfile`, `resource`

3. Fix abbreviations:
   - `mgr` → `manager` (unless very locally scoped)
   - `cfg` → `config`
   - `pkg` is OK for `package` (common Go convention)
   - `ctx` is OK for `context.Context` (standard)

4. Commit: "refactor: improve variable naming clarity"

### Task 6.4: Fix Linter Issues (1.5 hours)
**Agent Instructions:**
1. Run comprehensive linting:
   ```bash
   # Run golangci-lint with all checks
   golangci-lint run ./...

   # Run go vet
   go vet ./...

   # Check formatting
   gofmt -d -s .
   ```

2. Fix issues in priority order:
   - Errors (must fix)
   - Warnings about correctness
   - Style issues that improve readability
   - Skip pedantic style issues that hurt readability

3. Common fixes:
   - Add missing error checks
   - Fix error strings (lowercase, no punctuation)
   - Remove redundant type declarations
   - Simplify slice/map initialization

4. Commit fixes by category:
   - "fix: add missing error checks"
   - "style: fix error string formatting"
   - "refactor: simplify initialization"

### Task 6.5: Improve Package Documentation (1 hour)
**Agent Instructions:**
1. Ensure each package has a clear package comment:
   ```go
   // Package orchestrator coordinates resource operations across
   // package managers and dotfiles. It provides the main sync
   // logic for the plonk CLI.
   package orchestrator
   ```

2. Add missing function documentation:
   - Public functions must have comments
   - Focus on "why" not "what"
   - Document non-obvious behavior

3. Fix comment style:
   ```go
   // Bad:
   // Gets the home directory
   func GetHomeDir() string {

   // Better:
   // HomeDir returns the user's home directory, using the
   // HOME environment variable with fallback to os.UserHomeDir.
   func HomeDir() string {
   ```

4. Document pending integration:
   ```go
   // Sync orchestrates resource synchronization with hook support.
   // NOTE: This method is not yet integrated into the CLI commands.
   // Integration is planned for Phase 7 of the refactor.
   func (o *Orchestrator) Sync() error {
   ```

5. Commit: "docs: improve package and function documentation"

### Task 6.6: Final Code Quality Check (1 hour)
**Agent Instructions:**
1. Run complexity analysis:
   ```bash
   # Check cyclomatic complexity
   gocyclo -over 15 ./...
   ```

2. Simplify complex functions if found (split into helpers)

3. Check for:
   - Functions over 50 lines → split if possible
   - Deeply nested code → early returns
   - Complex conditionals → extract to well-named functions

4. Final test run:
   ```bash
   go test ./...
   ```

5. Create summary report:
   ```
   PHASE_6_SUMMARY.md
   - Unused code removed: X lines
   - Functions renamed: Y
   - Linter issues fixed: Z
   - Documentation improved: N packages
   - Complex functions simplified: M
   ```

6. Commit: "chore: final code quality improvements"

## Important Notes

### What NOT to Remove
1. **Phase 5 Integration Code**:
   - New Orchestrator.Sync() method
   - HookRunner and all hook-related code
   - Lock v2 migration functions
   - New ResourceEntry types

2. **Exported API**:
   - Public functions/types (might be used externally)
   - Interface definitions (even with single implementation)

3. **Future Extensibility**:
   - Meta fields in structs
   - "degraded" state (reserved for future use)
   - Resource abstraction methods

### What TO Remove
1. **Obviously Dead Code**:
   - Old implementations replaced by new ones
   - Helper functions no longer called
   - Test utilities not used by any tests
   - Commented-out code

2. **Redundant Code**:
   - Duplicate helper functions
   - Unused error types
   - Empty interfaces

## Success Criteria
- [ ] No genuinely dead code remains
- [ ] Naming is consistent across packages
- [ ] All public APIs are documented
- [ ] Linter issues addressed (where sensible)
- [ ] Code is more readable and maintainable
- [ ] Phase 5 integration code preserved for Phase 7

## Risk Mitigation
- **Over-deletion**: When in doubt, keep the code
- **Breaking changes**: Run tests after each change
- **Lost functionality**: Keep clear notes about what's pending integration

Remember: This phase is about code quality and maintainability, not hitting specific metrics. A well-named, documented codebase is more valuable than arbitrary LOC reduction.
