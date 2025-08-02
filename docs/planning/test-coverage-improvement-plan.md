# Test Coverage Improvement Plan

**Date**: 2025-08-02
**Current Coverage**: 30%
**Target for v1.0**: 50-60%
**Gap**: 20-30% coverage needed

## Executive Summary

Analysis of the plonk codebase reveals that reaching acceptable test coverage for v1.0 requires focused effort on a few high-impact packages. The `internal/commands` package alone represents over 4,500 lines of code with only 9.2% coverage - improving this single package would have the largest impact on overall coverage.

## Current Coverage by Package

| Package | LOC | Current Coverage | Impact Score* | Priority |
|---------|-----|-----------------|---------------|----------|
| internal/commands | 4,577 | 9.2% | 4,157 | **HIGHEST** |
| internal/diagnostics | 798 | 13.7% | 689 | **HIGH** |
| internal/orchestrator | 490 | 0.7% | 487 | **MEDIUM** |
| internal/config | 855 | 38.4% | 527 | **MEDIUM** |
| internal/clone | 436 | 0% | 436 | **LOW-MEDIUM** |
| internal/resources/dotfiles | 2,769 | 50.3% | 1,376 | **LOW** |
| internal/output | 224 | 0% | 224 | **LOW** |

*Impact Score = LOC × (1 - Coverage%)

### Well-Covered Packages ✅
- internal/lock: 83.3%
- internal/resources/packages: 61.7%
- internal/resources: 58.5%

## Deep Inspection Commands

### 1. Analyze internal/commands (Highest Priority)

```bash
# List all command files
ls -la internal/commands/*.go | grep -v _test

# See which commands have tests
ls -la internal/commands/*_test.go

# Check complexity of each command
scc internal/commands/*.go --by-file --no-cocomo

# Find untested exported functions
go doc -all ./internal/commands | grep "^func"

# See test coverage by function
go test -coverprofile=cover.out ./internal/commands
go tool cover -func=cover.out | grep "internal/commands"
```

**Key Questions:**
- Which commands are completely untested?
- Are there shared helper functions that would give broad coverage?
- Can we test command logic without the cobra CLI layer?

### 2. Analyze internal/diagnostics (High Priority)

```bash
# Analyze the diagnostics package structure
tree internal/diagnostics/

# Check what's being tested
grep -n "func Test" internal/diagnostics/*_test.go

# Look for mockable interfaces
grep -n "type.*interface" internal/diagnostics/*.go

# Find external dependencies that need mocking
grep -E "brew|npm|cargo|pip" internal/diagnostics/*.go
```

**Key Questions:**
- Is the doctor command testing actual system calls or using mocks?
- What are the main diagnostic checks that need coverage?
- Can we mock package manager checks effectively?

### 3. Analyze internal/orchestrator (Medium Priority)

```bash
# Understand the orchestrator's role
grep -n "type.*Orchestrator" internal/orchestrator/*.go
grep -n "func.*Orchestrator" internal/orchestrator/*.go

# Check for existing test infrastructure
ls internal/orchestrator/*_test.go

# Find integration points
grep -E "resources\.|lock\.|config\." internal/orchestrator/*.go
```

**Key Questions:**
- Is this mostly coordination logic or business logic?
- What are the main workflows it orchestrates?
- Are the dependencies already mockable?

### 4. Analyze internal/config (Medium Priority)

```bash
# See what configuration aspects exist
grep -n "type.*Config" internal/config/*.go

# Check file I/O operations
grep -E "os\.|ioutil\.|filepath\." internal/config/*.go

# Look for validation logic
grep -n "Validate\|Valid\|Check" internal/config/*.go
```

**Key Questions:**
- How much is file I/O vs business logic?
- Is there complex validation that needs testing?
- Can we use temp directories for file-based tests?

## Effort Estimation Framework

For each package, assess these factors:

### Mockability (1-5 scale)
- 5: Interfaces already exist, easy to mock
- 3: Some refactoring needed for testability
- 1: Tight coupling, major refactoring required

### Complexity (1-5 scale)
- 5: Complex business logic, many edge cases
- 3: Mix of simple and complex logic
- 1: Simple CRUD or pass-through logic

### Dependencies (1-5 scale)
- 5: Many external dependencies (OS, network, etc.)
- 3: Some external calls, mostly internal
- 1: Pure functions, minimal dependencies

**Quick LOE Formula**: Hours = (LOC / 100) × (Complexity + Dependencies) / Mockability

## Recommended Implementation Plan

### Phase 1: Commands Package (Week 1)
Target: Add 15-20% overall coverage

1. Identify 3-4 most critical commands:
   - `status` - Shows system state
   - `install` - Package installation
   - `apply` - Applies configuration
   - `diff` - Shows drift

2. Test approach:
   - Focus on business logic, not cobra setup
   - Mock external interfaces (already exist via CommandExecutor)
   - Test error paths and edge cases

### Phase 2: Diagnostics Package (Week 2)
Target: Add 3-5% overall coverage

1. Focus on testable diagnostic checks
2. Mock package manager interfaces
3. Test diagnostic result aggregation

### Phase 3: Config & Orchestrator (Week 3)
Target: Add 2-3% overall coverage

1. Config: Test validation and parsing logic
2. Orchestrator: Consider if integration tests are more appropriate

## Quick Wins

Look for these easy coverage gains:

1. **Utility Functions**
   - Path manipulation helpers
   - String formatting utilities
   - Validation functions

2. **Error Handling**
   - Error creation and wrapping
   - Error type checking
   - Recovery paths

3. **Pure Functions**
   - Sorting/filtering logic
   - Data transformations
   - Business rule calculations

## Success Metrics

- [ ] Overall coverage reaches 50%+
- [ ] `internal/commands` reaches 40%+ coverage
- [ ] `internal/diagnostics` reaches 50%+ coverage
- [ ] No critical business logic remains untested
- [ ] All quick wins implemented

## Notes

- Consider integration tests for packages like `clone` and `orchestrator` where unit tests provide limited value
- The existing mock infrastructure (CommandExecutor) should make testing commands straightforward
- Focus on testing business logic, not I/O operations
