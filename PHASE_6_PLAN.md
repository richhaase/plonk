# Phase 6: Final Structural Cleanup

## Objective
Complete remaining structural improvements, integrate Phase 5's hook and lock v2 systems, and prepare the codebase for final polish. Focus on simplicity and idiomatic Go patterns.

## Timeline
Day 9 (8 hours)

## Current State
- ~13,826 LOC after Phase 5
- 8 packages (commands, config, lock, orchestrator, output, resources + sub-packages)
- Resource abstraction complete
- Lock v2 and hooks implemented but NOT integrated
- Commands package still contains business logic

## Key Goals
- Wire up the new orchestrator with hooks and lock v2
- Extract business logic from commands package
- Simplify orchestrator to its coordination role
- Remove any remaining unnecessary abstractions
- Keep code simple and idiomatic

## Task Breakdown

### Task 6.1: Integrate New Orchestrator (2 hours)
**Agent Instructions:**
1. Update `internal/commands/sync.go` to use new orchestrator:
   ```go
   func runSync(cmd *cobra.Command, args []string) error {
       // ... flag parsing ...

       // Create new orchestrator
       orch := orchestrator.New(
           orchestrator.WithConfig(cfg),
           orchestrator.WithConfigDir(configDir),
           orchestrator.WithHomeDir(homeDir),
           orchestrator.WithDryRun(dryRun),
       )

       // Run sync with hooks and v2 lock
       result, err := orch.Sync(ctx)
       if err != nil {
           return fmt.Errorf("sync failed: %w", err)
       }

       // Format output
       return RenderOutput(result, format)
   }
   ```

2. Remove old sync functions:
   - `orchestrator.SyncPackages()`
   - `orchestrator.SyncDotfiles()`
   - Any other legacy sync code

3. Update output formatting to handle new result structure

4. Test the integration:
   ```bash
   # Test basic sync
   plonk sync --dry-run

   # Test with hooks
   echo "hooks:\n  pre_sync:\n    - command: 'echo pre-sync'" >> plonk.yaml
   plonk sync

   # Verify lock v2
   grep "version: 2" plonk.lock
   ```

5. Commit: "feat: integrate new orchestrator with CLI sync command"

**Validation:**
- Sync command works with new orchestrator
- Hooks execute properly
- Lock files are v2 format

### Task 6.2: Extract Business Logic from Commands (2.5 hours)
**Agent Instructions:**
1. Identify business logic in commands package:
   - Complex package operations
   - State calculations
   - Resource manipulations

2. Move logic to appropriate packages:
   ```go
   // Before (in commands/add.go):
   func calculatePackageState(manager, name string) (State, error) {
       // 50 lines of logic
   }

   // After (in resources/packages/state.go):
   func CalculateState(manager, name string) (State, error) {
       // Same logic, better location
   }

   // And in commands/add.go:
   state, err := packages.CalculateState(manager, name)
   ```

3. Commands should only:
   - Parse CLI flags
   - Call business logic
   - Format output
   - Handle errors

4. Target packages for logic:
   - `resources/packages/` - package operations
   - `resources/dotfiles/` - dotfile operations
   - `orchestrator/` - coordination logic
   - `config/` - configuration logic

5. Commit after each command is cleaned up

**Result:** Commands become thin CLI adapters

### Task 6.3: Simplify Orchestrator Package (1.5 hours)
**Agent Instructions:**
1. Review orchestrator package for non-coordination code:
   - Complex business logic
   - Resource-specific operations
   - Utility functions

2. Extract to appropriate locations:
   - Resource-specific logic → `resources/` package
   - General utilities → where they're used
   - Keep only coordination code

3. Orchestrator should only:
   - Load configuration
   - Initialize resources
   - Coordinate operations
   - Run hooks
   - Update lock file

4. Target structure:
   ```
   orchestrator/
   ├── orchestrator.go    # Main coordination
   ├── hooks.go          # Hook runner
   └── reconcile.go      # Generic reconciliation
   ```

5. Remove any code that doesn't fit this model

6. Commit: "refactor: simplify orchestrator to pure coordination"

### Task 6.4: Remove Unnecessary Abstractions (1.5 hours)
**Agent Instructions:**
1. Find interfaces with single implementations:
   ```bash
   # Look for interface definitions
   grep -r "type.*interface" internal/ | grep -v "_test.go"
   ```

2. Remove if:
   - Only one implementation exists
   - No plans for additional implementations
   - Not part of public API

3. Find overly generic code:
   - Generic handlers that only handle one type
   - Abstracted functions used in one place
   - Configuration layers that just pass through

4. Simplify to direct implementations:
   ```go
   // Before:
   type Runner interface {
       Run() error
   }
   type PackageRunner struct{}
   func (p *PackageRunner) Run() error {}

   // After (if only one implementation):
   func RunPackages() error {}
   ```

5. Keep abstractions that:
   - Have multiple implementations (Resource interface)
   - Enable testing (mock-able interfaces)
   - Provide clear architectural boundaries

6. Commit: "refactor: remove single-implementation interfaces"

### Task 6.5: Consolidate Related Code (1.5 hours)
**Agent Instructions:**
1. Look for split logic that belongs together:
   - Helper functions far from their usage
   - Related types in different files
   - Scattered constants

2. Consolidate within packages:
   ```go
   // If lock package has:
   // - types.go (5 lines)
   // - constants.go (3 lines)
   // - helpers.go (10 lines)
   // Consider merging into lock.go
   ```

3. Move misplaced code:
   - Config validation from commands → config package
   - Resource helpers from orchestrator → resources
   - Output utilities from commands → output

4. Guidelines:
   - Keep packages focused on one responsibility
   - Co-locate related functionality
   - Avoid tiny files unless they'll grow

5. Commit: "refactor: consolidate related code"

### Task 6.6: Final Integration Testing (1 hour)
**Agent Instructions:**
1. Full workflow test:
   ```bash
   # Clean start
   rm -f plonk.yaml plonk.lock

   # Initialize
   plonk init

   # Add packages and dotfiles
   plonk add brew jq
   plonk dotfiles add ~/.bashrc

   # Add hooks to config
   cat >> plonk.yaml << EOF
   hooks:
     pre_sync:
       - command: "echo 'Starting sync...'"
     post_sync:
       - command: "echo 'Sync complete!'"
   EOF

   # Sync with hooks
   plonk sync

   # Verify:
   # - Hooks executed
   # - Lock file is v2
   # - Resources tracked properly
   ```

2. Run all tests:
   ```bash
   go test ./...
   just test-ux
   ```

3. Check for issues:
   - Circular dependencies
   - Missing functionality
   - Performance problems

4. Create summary:
   ```
   PHASE_6_SUMMARY.md
   - Orchestrator integrated ✓
   - Commands simplified ✓
   - Abstractions removed: X
   - Code consolidated ✓
   - All tests passing ✓
   ```

5. Commit: "test: verify complete system integration"

## Design Principles

### Keep It Simple
- Direct function calls over interfaces when possible
- Clear names over clever abstractions
- Explicit over implicit

### Idiomatic Go
- Prefer functions over methods for stateless operations
- Use interfaces only when needed
- Keep error handling direct

### Clear Architecture
- Commands: CLI parsing and output
- Resources: Business logic for packages/dotfiles
- Orchestrator: Coordination only
- Config: Configuration management
- Lock: Lock file operations
- Output: Formatting utilities

## What NOT to Do

1. **Don't chase metrics**:
   - No arbitrary LOC targets
   - No forced package consolidation
   - Quality over numbers

2. **Don't over-abstract**:
   - No interfaces "for future use"
   - No generic solutions for specific problems
   - No layers that just pass through

3. **Don't break working code**:
   - Test after each change
   - Keep commits atomic
   - Preserve functionality

## Success Criteria
- [ ] New orchestrator integrated and working
- [ ] Hooks execute on sync
- [ ] Lock v2 files generated
- [ ] Commands contain only CLI logic
- [ ] Orchestrator is pure coordination
- [ ] Unnecessary abstractions removed
- [ ] All tests passing
- [ ] Code is simpler and more maintainable

## Notes for Agent
This is about making the code better, not hitting targets. Every change should make the code:
- Easier to understand
- Easier to modify
- More idiomatic
- More maintainable

If something is already good, leave it alone. Focus on real improvements.
