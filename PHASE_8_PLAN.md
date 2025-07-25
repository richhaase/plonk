# Phase 8: Testing & Documentation

## Objective
Finalize the refactoring with comprehensive testing, documentation updates, and verification that all systems work correctly together.

## Timeline
Day 11 (8 hours)

## Current State (Expected after Phase 7)
- Clean, consistent codebase with good naming
- ~13,500-13,700 LOC (after dead code removal)
- 8 well-organized packages
- All infrastructure integrated and working:
  - New orchestrator with hooks
  - Lock v2 with automatic migration
  - Resource abstraction for extensibility
- No known architectural issues

## Task Breakdown

### Task 8.1: Update and Enhance Tests (2.5 hours)
**Agent Instructions:**
1. Review existing test coverage:
   ```bash
   # Generate coverage report
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out -o coverage.html

   # Focus on packages with low coverage
   go test -cover ./...
   ```

2. Add missing unit tests for Phase 5-6 features:
   ```go
   // Test hook execution
   func TestHookRunner_Timeout(t *testing.T) {
       // Test that hooks timeout correctly
   }

   // Test lock migration
   func TestLockV1ToV2Migration(t *testing.T) {
       // Test automatic migration preserves data
   }

   // Test new orchestrator
   func TestOrchestrator_WithOptions(t *testing.T) {
       // Test functional options pattern
   }
   ```

3. Update broken tests from refactoring:
   - Fix import paths
   - Update for new function names
   - Adjust for new behavior

4. Ensure critical paths have tests:
   - Resource reconciliation
   - Hook execution with continue_on_error
   - Lock file migration
   - Config validation

5. Commit: "test: add missing tests for new features"

### Task 8.2: Add Integration Tests (2 hours)
**Agent Instructions:**
1. Create end-to-end tests for key workflows:
   ```go
   // tests/integration/full_sync_test.go
   func TestFullSyncWithHooks(t *testing.T) {
       // Create config with packages, dotfiles, and hooks
       // Run sync
       // Verify hooks executed
       // Verify lock v2 created
       // Verify resources applied
   }

   func TestLockMigrationOnSync(t *testing.T) {
       // Create v1 lock file
       // Run sync
       // Verify automatic migration
       // Verify data preserved
   }
   ```

2. Test error scenarios:
   - Hook failures (with and without continue_on_error)
   - Invalid configurations
   - Missing resources
   - Permission errors

3. Test resource extensibility:
   ```go
   // Verify the Resource interface works for future types
   type mockResource struct{}
   func (m *mockResource) ID() string { return "mock" }
   // ... implement interface
   ```

4. Performance benchmarks:
   ```go
   func BenchmarkReconciliation(b *testing.B) {
       // Measure reconciliation performance
   }

   func BenchmarkSync(b *testing.B) {
       // Measure full sync performance
   }
   ```

5. Commit: "test: add comprehensive integration tests"

### Task 8.3: Create Architecture Documentation (1.5 hours)
**Agent Instructions:**
1. Create `docs/ARCHITECTURE.md`:
   ```markdown
   # Plonk Architecture

   ## Overview
   Plonk uses a resource-based architecture that provides extensibility
   for future resource types while maintaining simplicity for current
   package and dotfile management.

   ## Core Concepts

   ### Resources
   The Resource interface is the foundation for extensibility:
   ```go
   type Resource interface {
       ID() string
       Desired() []Item
       Actual(ctx context.Context) []Item
       Apply(ctx context.Context, item Item) error
   }
   ```

   ### Reconciliation
   Generic reconciliation pattern compares desired vs actual state...

   ### Orchestrator
   Coordinates operations across resources, manages hooks and lock files...

   ## Package Structure
   ```
   internal/
   ├── commands/      # CLI command handlers (thin layer)
   ├── config/        # Configuration management
   ├── diagnostics/   # Health checks and diagnostics
   ├── lock/          # Lock file v2 with migration
   ├── orchestrator/  # Resource coordination and hooks
   ├── output/        # Output formatting (table/json/yaml)
   └── resources/     # Resource implementations
       ├── dotfiles/  # Dotfile management
       └── packages/  # Package manager abstraction
   ```

   ## Data Flow
   1. User runs command → commands package
   2. Command loads config → config package
   3. Creates orchestrator → orchestrator package
   4. Orchestrator coordinates resources → resources package
   5. Results formatted → output package

   ## Adding a New Resource Type
   To add a new resource type (e.g., Docker Compose):

   1. Create package: `resources/compose/`
   2. Implement Resource interface
   3. Register in orchestrator
   4. Add config schema support
   5. Update lock file schema if needed

   ## Adding a New Package Manager
   1. Create file: `resources/packages/newmanager.go`
   2. Implement PackageManager interface
   3. Register in package registry
   4. Add tests

   ## Hook System
   Hooks run at pre_sync and post_sync phases...

   ## Lock File Format
   Version 2 supports generic resources...
   ```

2. Include diagrams where helpful (mermaid or ASCII)

3. Document design decisions and trade-offs

4. Commit: "docs: create comprehensive architecture documentation"

### Task 8.4: Update User Documentation (1.5 hours)
**Agent Instructions:**
1. Update `README.md`:
   - Add hook examples
   - Document lock file migration
   - Update installation instructions
   - Add troubleshooting section

2. Create example configurations:
   ```yaml
   # examples/basic.yaml
   packages:
     homebrew:
       - ripgrep
       - jq
   dotfiles:
     - ~/.gitconfig

   # examples/with-hooks.yaml
   packages:
     homebrew:
       - ripgrep
   hooks:
     pre_sync:
       - command: "echo 'Starting sync...'"
       - command: "./backup.sh"
         timeout: "2m"
     post_sync:
       - command: "./notify.sh"
         continue_on_error: true

   # examples/multi-manager.yaml
   packages:
     homebrew:
       - ripgrep
     npm:
       - typescript
       - prettier
     pip:
       - black
       - mypy
   ```

3. Document new features:
   - Hook configuration and behavior
   - Lock file v2 benefits
   - Resource extensibility

4. Update CLI help text if needed

5. Commit: "docs: update user documentation for new features"

### Task 8.5: Performance Verification (0.5 hours)
**Agent Instructions:**
1. Run benchmarks:
   ```bash
   # Run all benchmarks
   go test -bench=. -benchmem ./...

   # Profile if needed
   go test -cpuprofile=cpu.prof -bench=BenchmarkSync
   go tool pprof cpu.prof
   ```

2. Verify acceptable performance:
   - Reconciliation should be fast (<100ms for typical configs)
   - No obvious memory leaks
   - Reasonable memory usage

3. Document any performance considerations

4. Commit only if optimizations are made

### Task 8.6: Final Verification (1 hour)
**Agent Instructions:**
1. Full system test on clean environment:
   ```bash
   # Use Docker for clean environment
   docker run -it --rm golang:1.22 bash

   # In container:
   git clone <repo>
   cd plonk
   go install ./cmd/plonk

   # Test all major workflows
   plonk init
   plonk add brew ripgrep
   plonk dotfiles add ~/.bashrc
   plonk sync
   plonk status --health
   ```

2. Verify all features work:
   - [ ] Package management (all 6 managers)
   - [ ] Dotfile management
   - [ ] Hooks execute properly
   - [ ] Lock v2 migration automatic
   - [ ] All output formats (table/json/yaml)
   - [ ] Health checks function

3. Check final metrics:
   ```bash
   # Line count
   scc --include-lang Go internal/

   # Test execution time
   time go test ./...

   # Build time
   time go build ./cmd/plonk
   ```

4. Create final summary:
   ```markdown
   # Refactor Complete Summary

   ## Metrics
   - Starting: 9 packages, ~13,536 LOC
   - Ending: 8 packages, ~X,XXX LOC
   - Test coverage: XX%
   - Test execution: Xs

   ## Major Achievements
   - Resource abstraction for extensibility
   - Hook system for automation
   - Lock v2 with automatic migration
   - Clean architectural boundaries
   - Idiomatic Go throughout

   ## Breaking Changes
   - None (full backward compatibility maintained)

   ## Ready for AI Lab
   - Resource interface supports future types
   - Hook system enables automation
   - Clean architecture for extensions
   ```

5. Commit: "docs: final refactor summary and verification"

## Success Criteria
- [ ] Test coverage >70% for critical packages
- [ ] All integration tests pass
- [ ] Architecture clearly documented
- [ ] User docs updated with examples
- [ ] Performance acceptable
- [ ] All features working correctly
- [ ] No regressions from refactor

## Notes for Agent

### Testing Philosophy
- Test behavior, not implementation
- Focus on critical paths and edge cases
- Integration tests verify the full system works
- Keep tests readable and maintainable

### Documentation Philosophy
- Write for future maintainers
- Include "why" not just "what"
- Use examples liberally
- Keep it up to date

### Final Checks
- Ensure nothing is broken
- Verify backward compatibility
- Confirm extensibility goals met
- Leave codebase better than you found it

This is the final phase - make sure everything is polished, tested, and ready for long-term maintenance and AI Lab integration.
