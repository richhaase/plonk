# Architecture Decision: Commands Package Testing Strategy

## 🛑 SAFETY NOTICE: THIS DECISION PREVENTS DANGEROUS TESTS 🛑

**This architecture decision helps enforce the CRITICAL RULE: NEVER modify system state in unit tests.**

**Command orchestration functions directly call package managers and file operations. Testing them would risk breaking developer machines.**

---

**Date**: 2025-08-03
**Decision**: Do not unit test command orchestration functions
**Status**: Accepted

## Context

During Phase 3 of test improvement efforts, we attempted to plan unit testing for the commands package orchestration functions (the `runXXX` functions). Analysis revealed fundamental architectural constraints that make this inappropriate.

## Decision

We will NOT unit test the command orchestration functions in the commands package. Instead, we will:
1. Test pure functions within commands (already completed in Phase 2)
2. Focus testing efforts on packages with actual business logic
3. Accept that commands are integration points, not unit-testable components

## Rationale

### Why Commands Resist Unit Testing

1. **Direct Dependency Instantiation**
   ```go
   // Commands directly create their dependencies
   cfg := config.LoadWithDefaults(configDir)
   orch := orchestrator.New(...)
   result, err := orch.ReconcileAll(ctx)
   ```
   - No dependency injection
   - No way to provide test doubles
   - Tightly coupled to real implementations

2. **Global State Dependencies**
   ```go
   // Commands rely on global functions
   configDir := config.GetConfigDir()
   homeDir := config.GetHomeDir()
   ```
   - File system paths from environment
   - No abstraction layer
   - Real system interaction expected

3. **Integration by Design**
   - Commands are the CLI adapter layer
   - They orchestrate multiple subsystems
   - Their purpose is integration, not isolated logic

4. **Mixing Concerns**
   Each command function handles:
   - Flag parsing
   - Dependency creation
   - Business logic orchestration
   - Output formatting
   - Error handling

### Architectural Reality

Commands are essentially "main" functions for each CLI operation. They are:
- **Integration points** by design
- **Orchestration layers** not business logic layers
- **CLI adapters** that translate user input to system operations

Trying to unit test them is like trying to unit test a `main()` function - it's the wrong abstraction level.

## Consequences

### Positive
- Focuses testing effort where it provides value
- Respects the architectural design
- Avoids complex test infrastructure for little benefit
- Maintains clean production code

### Negative
- Commands package will have lower coverage (~15-20%)
- Some error paths in commands won't be tested
- Integration tests become more important

### Mitigation
- Thorough testing of underlying packages (orchestrator, packages, dotfiles)
- Integration tests for critical user workflows
- Manual testing of CLI interactions

## Alternative Approaches Considered

1. **Dependency Injection** - Would require significant production code changes
2. **Interface Everything** - Would add complexity without clear benefit
3. **Test Helpers with Mocks** - Still requires production changes for injection points
4. **Subcutaneous Testing** - This is effectively what we're doing by testing the layers below

## Lessons Learned

1. **Not everything needs unit tests** - Integration points are better tested as integrations
2. **Respect architectural boundaries** - Don't force unit tests where they don't belong
3. **Test at the right level** - Business logic in packages, integration at command level
4. **Coverage isn't everything** - Quality and appropriate testing matters more than percentages

## Future Considerations

If commands need to become more testable:
1. Extract complex logic to testable packages
2. Keep commands as thin as possible
3. Consider integration tests for critical paths
4. Use temporary file systems for file-based operations

## Reference

This decision affects:
- Test coverage goals (adjusted expectations for commands package)
- Testing strategy (focus on business logic packages)
- Development practices (where to put complex logic)
