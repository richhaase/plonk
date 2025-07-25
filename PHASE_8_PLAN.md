# Phase 8: Testing & Documentation

## Objective
Finalize the refactoring with comprehensive testing, documentation updates, and verification that all systems work correctly together.

## Timeline
Day 11 (8 hours)

## Current State (Expected after Phase 6-7)
- New orchestrator integrated with CLI
- Hooks and lock v2 working
- Business logic extracted from commands
- Code quality improved and naming consistent
- All structural refactoring complete

## Task Breakdown

### Task 8.1: Update Tests for New Structure (2 hours)
**Agent Instructions:**
1. Review and update test files:
   - Update imports for moved code
   - Fix tests broken by refactoring
   - Remove tests for deleted code
   - Add tests for new functionality

2. Ensure test coverage for:
   - New orchestrator integration
   - Hook execution
   - Lock v2 migration
   - Resource abstraction

3. Run tests and fix failures:
   ```bash
   go test ./...
   ```

4. Check test coverage:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

5. Commit: "test: update tests for refactored structure"

### Task 8.2: Add Integration Tests (2 hours)
**Agent Instructions:**
1. Create integration tests for key workflows:
   ```go
   // tests/integration/hooks_test.go
   func TestSyncWithHooks(t *testing.T) {
       // Create config with hooks
       // Run sync
       // Verify hooks executed
       // Verify lock v2 created
   }

   func TestLockV1Migration(t *testing.T) {
       // Create v1 lock file
       // Run sync
       // Verify migration to v2
       // Verify data preserved
   }
   ```

2. Test error scenarios:
   - Hook timeout
   - Hook failure with continue_on_error
   - Invalid configuration
   - Missing resources

3. Performance benchmarks:
   ```go
   func BenchmarkSync(b *testing.B) {
       // Measure sync performance
   }
   ```

4. Commit: "test: add integration tests for hooks and migration"

### Task 8.3: Create ARCHITECTURE.md (1.5 hours)
**Agent Instructions:**
1. Create `docs/ARCHITECTURE.md`:
   ```markdown
   # Plonk Architecture

   ## Overview
   Plonk uses a resource-based architecture...

   ## Core Concepts
   ### Resources
   All managed items (packages, dotfiles) implement the Resource interface...

   ### Orchestrator
   Coordinates operations across resources...

   ### Lock File
   Tracks state using versioned schema...

   ## Package Structure
   - `commands/` - CLI command handlers
   - `config/` - Configuration management
   - `lock/` - Lock file operations
   - `orchestrator/` - Resource coordination
   - `output/` - Output formatting
   - `resources/` - Resource implementations
     - `packages/` - Package managers
     - `dotfiles/` - Dotfile management

   ## Adding a New Resource Type
   1. Implement the Resource interface
   2. Register with orchestrator
   3. Add configuration support
   4. Update lock schema if needed

   ## Adding a New Package Manager
   1. Create manager in resources/packages/
   2. Implement PackageManager interface
   3. Add to manager registry
   4. Add tests
   ```

2. Include:
   - Architectural decisions
   - Data flow diagrams
   - Extension points
   - Future considerations

3. Commit: "docs: create architecture documentation"

### Task 8.4: Update README and Examples (1.5 hours)
**Agent Instructions:**
1. Update README.md:
   - Fix any outdated commands
   - Update feature list
   - Add hook examples
   - Update installation instructions

2. Create example configurations:
   ```yaml
   # examples/basic-plonk.yaml
   packages:
     homebrew:
       - jq
       - ripgrep

   # examples/with-hooks.yaml
   packages:
     homebrew:
       - jq
   hooks:
     pre_sync:
       - command: "echo 'Starting sync...'"
     post_sync:
       - command: "./notify.sh"
         continue_on_error: true
   ```

3. Update quick-start guide:
   - Reflect new features
   - Show common workflows
   - Include troubleshooting

4. Commit: "docs: update README and add examples"

### Task 8.5: Performance Optimization (1 hour)
**Agent Instructions:**
1. Profile the application:
   ```bash
   go test -bench=. -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   ```

2. Identify bottlenecks:
   - Slow package manager operations
   - Inefficient file operations
   - Unnecessary allocations

3. Optimize if reasonable:
   - Parallel operations where safe
   - Reduce allocations
   - Cache expensive computations

4. Measure improvements:
   ```bash
   # Before and after benchmarks
   go test -bench=. -benchmem
   ```

5. Only optimize if:
   - Significant improvement (>20%)
   - Doesn't complicate code
   - Maintains correctness

6. Commit: "perf: optimize critical paths"

### Task 8.6: Final Verification (1 hour)
**Agent Instructions:**
1. Full system test:
   ```bash
   # Clean environment
   docker run -it golang:latest

   # Install plonk
   go install github.com/yourusername/plonk@latest

   # Test all commands
   plonk init
   plonk add brew jq
   plonk dotfiles add .bashrc
   plonk sync
   plonk status
   ```

2. Verify all features:
   - [ ] Package management works
   - [ ] Dotfile management works
   - [ ] Hooks execute properly
   - [ ] Lock v2 migration works
   - [ ] All commands function
   - [ ] Output formats work (table/json/yaml)

3. Check metrics:
   ```bash
   # Line count
   find internal/ -name "*.go" | xargs wc -l

   # Package count
   find internal/ -type d -maxdepth 1 | wc -l

   # Test execution time
   time go test ./...
   ```

4. Create final summary:
   ```
   REFACTOR_COMPLETE.md
   - Starting: 9 packages, ~13,536 LOC
   - Ending: X packages, Y LOC
   - Major features added:
     - Resource abstraction
     - Hook system
     - Lock v2 with migration
   - All tests passing
   - Documentation updated
   ```

5. Commit: "docs: final refactor summary"

## Success Criteria
- [ ] All tests passing
- [ ] Integration tests cover key workflows
- [ ] Architecture documented
- [ ] README updated with current information
- [ ] Examples provided for new features
- [ ] Performance acceptable
- [ ] All features working correctly

## Notes for Agent
This is the final phase - focus on:
1. Ensuring everything works correctly
2. Documenting what was built
3. Making it easy for others to understand and extend
4. Verifying the refactor achieved its goals

Don't introduce new features or major changes at this stage.
