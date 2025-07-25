# Phase 4: AI Lab-Compatible Code Reduction

## Objective
Achieve 3,000-4,000 LOC reduction through duplication removal and internal simplification while preserving ALL functionality required for AI Lab features.

## Timeline
Day 6-7 (16 hours)

## Current State
- ~14,300 LOC after Phase 3
- All 6 package managers functional and needed
- Resource abstraction in place for AI Lab
- Significant duplication identified in Phase 3.5 analysis

## Target State
- ~10,000-11,000 LOC
- Same functionality, better organized
- Common patterns extracted
- Cleaner, more maintainable code

## Task Breakdown

### Task 4.1: Extract Command Boilerplate (4 hours)
**Agent Instructions:**
1. Create `internal/commands/builder.go` with common command patterns:
   ```go
   package commands

   import "github.com/spf13/cobra"

   type CommandBuilder struct {
       Use   string
       Short string
       Long  string
   }

   func (b *CommandBuilder) Build(run RunFunc) *cobra.Command {
       cmd := &cobra.Command{
           Use:   b.Use,
           Short: b.Short,
           Long:  b.Long,
           RunE:  wrapRunFunc(run),
       }

       // Add common flags
       cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

       return cmd
   }

   func wrapRunFunc(run RunFunc) func(*cobra.Command, []string) error {
       return func(cmd *cobra.Command, args []string) error {
           // Common setup
           format, err := parseOutputFormat(cmd)
           if err != nil {
               return err
           }

           // Call the actual command logic
           return run(cmd, args, format)
       }
   }
   ```

2. Refactor commands to use builder:
   - Extract common output format parsing (12+ files)
   - Extract config directory retrieval
   - Standardize error handling patterns
   - Remove duplicate flag parsing

3. Expected reduction:
   - 30-40 lines per command × 20 commands = 600-800 lines
   - But builder adds ~100 lines
   - Net reduction: 300-400 lines

4. Commit: "refactor: extract command boilerplate with builder pattern"

**Validation:**
- All commands still work
- Less duplicate code
- Consistent behavior

### Task 4.2: Consolidate Flag Definitions (3 hours)
**Agent Instructions:**
1. Create `internal/commands/flags.go`:
   ```go
   package commands

   func AddPackageManagerFlags(cmd *cobra.Command) {
       cmd.Flags().Bool("brew", false, "Use Homebrew")
       cmd.Flags().Bool("npm", false, "Use NPM")
       cmd.Flags().Bool("pip", false, "Use pip")
       cmd.Flags().Bool("gem", false, "Use gem")
       cmd.Flags().Bool("cargo", false, "Use Cargo")
       cmd.Flags().Bool("go", false, "Use go install")
   }

   func AddCommonFlags(cmd *cobra.Command) {
       cmd.Flags().Bool("dry-run", false, "Preview changes without applying")
       cmd.Flags().Bool("force", false, "Force operation without confirmation")
   }

   func GetSelectedManager(cmd *cobra.Command) (string, error) {
       managers := []string{"brew", "npm", "pip", "gem", "cargo", "go"}
       for _, mgr := range managers {
           if val, _ := cmd.Flags().GetBool(mgr); val {
               return mgr, nil
           }
       }
       return "", nil
   }
   ```

2. Update commands to use shared flags:
   - Replace 18 package manager flag definitions
   - Replace 5 dry-run definitions
   - Replace 4 force definitions
   - Simplify ParseSimpleFlags in helpers.go

3. Expected reduction:
   - Package manager flags: 108 lines
   - Common flags: 60 lines
   - Total: 150-200 lines

4. Commit: "refactor: consolidate duplicate flag definitions"

**Validation:**
- Same flags available in all commands
- Flag parsing still works
- No behavior changes

### Task 4.3: Create Error Handling Utilities (2 hours)
**Agent Instructions:**
1. Create `internal/commands/errors.go`:
   ```go
   package commands

   import "fmt"

   func WrapCommandError(cmd string, err error) error {
       return fmt.Errorf("%s: %w", cmd, err)
   }

   func HandleCommonErrors(err error) error {
       // Common error transformations
       // e.g., permission errors, file not found, etc.
       return err
   }
   ```

2. Create `internal/resources/errors.go` for resource errors:
   ```go
   package resources

   func IsPackageNotFound(err error) bool {
       // Common patterns across managers
       patterns := []string{
           "not found",
           "404",
           "No available formula",
           "Unable to find",
       }
       // Check error against patterns
   }
   ```

3. Update error handling across codebase:
   - Replace 44 instances of `return fmt.Errorf`
   - Standardize error messages
   - Extract common error checking

4. Expected reduction: 200+ lines

5. Commit: "refactor: extract common error handling patterns"

### Task 4.4: Simplify Dotfiles Package (3 hours)
**Agent Instructions:**
Based on Phase 3.5 findings:

1. Consolidate path validation:
   - Merge 3 overlapping validation functions into 1
   - Remove over-defensive checks
   - Savings: 60 lines

2. Simplify atomic operations:
   - Use standard library for most operations
   - Keep atomic only where truly needed
   - Savings: 80 lines

3. Consolidate types:
   - Merge similar dotfile representations
   - Use single type throughout
   - Savings: 30 lines

4. Simplify directory walking:
   - Use single implementation
   - Remove duplicate logic
   - Savings: 120 lines

5. Reduce test redundancy:
   - Consolidate similar test cases
   - Remove redundant assertions
   - Target: 300-500 lines

6. Commit each simplification separately

**Total expected reduction: 500-800 lines**

### Task 4.5: Merge Doctor into Status (2 hours)
**Agent Instructions:**
1. Analyze `doctor.go` and `status.go`:
   - Identify overlapping functionality
   - Determine what's unique to doctor

2. Add doctor functionality to status:
   ```go
   // In status.go
   cmd.Flags().Bool("check-health", false, "Include health checks (doctor mode)")
   ```

3. Migrate doctor-specific checks:
   - Manager availability checks
   - Configuration validation
   - Path validation

4. Remove doctor.go entirely

5. Update help text and documentation

6. Expected reduction: ~200 lines (doctor.go size minus additions to status)

7. Commit: "refactor: merge doctor command into status with --check-health"

**Validation:**
- `plonk status --check-health` provides doctor functionality
- No functionality lost
- Single command for system state

### Task 4.6: Extract Cross-Package Utilities (2 hours)
**Agent Instructions:**
1. Create `internal/common/` package for truly shared utilities:
   ```
   internal/common/
   ├── paths.go      # Path manipulation utilities
   ├── strings.go    # String utilities
   ├── files.go      # File operation helpers
   └── exec.go       # Command execution utilities
   ```

2. Move utilities identified in Phase 3.5:
   - Path/string utilities: 100 lines
   - File operation helpers: 80 lines
   - Command execution patterns: 100 lines
   - Context helpers: 50 lines

3. Update imports across codebase

4. Remove duplicate implementations

5. Expected reduction: 500-800 lines

6. Commit: "refactor: extract cross-package utilities to common"

**Note:** Be selective - only extract truly common code used in 3+ places

### Task 4.7: Final Cleanup and Validation (2 hours)
**Agent Instructions:**
1. Run analysis tools:
   ```bash
   # Find remaining duplication
   dupl -threshold 10 ./internal/

   # Check for unused code
   staticcheck -checks U1000 ./...

   # Verify no circular dependencies
   go list -f '{{ join .Imports "\n" }}' ./...
   ```

2. Remove any dead code found

3. Run all tests:
   ```bash
   go test ./...
   just test-ux
   ```

4. Measure final LOC:
   ```bash
   find internal/ -name "*.go" -not -path "*/test/*" | xargs wc -l
   ```

5. Create summary report

6. Commit: "chore: final cleanup for Phase 4"

## Risk Mitigations

1. **Breaking Changes**: Test after each refactoring
2. **Over-extraction**: Only extract code used in 3+ places
3. **Complexity**: Keep utilities simple and focused

## Success Criteria
- [ ] 3,000-4,000 LOC reduction achieved
- [ ] All tests passing
- [ ] All 6 package managers still work
- [ ] Commands behavior unchanged
- [ ] No loss of functionality
- [ ] Code is more maintainable

## Notes for Agents
- Focus on real duplication, not superficial similarity
- Preserve all functionality - this is about organization, not removal
- Keep commits atomic for easy rollback
- Test frequently
- Don't over-abstract - some duplication is OK
