# Consolidated Test Coverage Improvement Plan

## 🚨🚨🚨 CRITICAL SAFETY WARNING 🚨🚨🚨

**UNIT TESTS MUST NEVER MODIFY THE REAL SYSTEM STATE**

**This is an ABSOLUTE, NON-NEGOTIABLE RULE. Tests that install packages, modify dotfiles, or execute system commands are FORBIDDEN and DANGEROUS.**

**We have had MULTIPLE incidents of tests attempting to modify developer machines. This CANNOT happen again.**

**NO TESTS IS BETTER THAN TESTS THAT BREAK DEVELOPER MACHINES**

---

**Date**: 2025-08-03
**Current Coverage**: 37.6% (up from 32.7%)
**Target Coverage**: 50% (minimum for v1.0)
**Timeline**: 2-3 weeks
**Progress**: Phase 1 & 2 complete

## Executive Summary

After reviewing all test improvement documentation and verifying actual coverage, this plan consolidates the approach into practical, low-risk improvements that can achieve 50% coverage for v1.0. The strategy focuses on:

1. **No coverage reporting bugs found** - Config (38.4%) and dotfiles (50.3%) already have decent coverage
2. **Quick wins** - ✅ Output package now at 80% coverage (Phase 1)
3. **High-impact packages** - ✅ Commands package improved from 9.2% to 14.1% (Phase 2)
4. **Pragmatic targets** - Need +12.4% more to reach 50% total

## Current State Analysis

### Coverage by Priority

| Package | Current | LOC | Realistic Target | Coverage Gain | Impact | Status |
|---------|---------|-----|------------------|---------------|--------|--------|
| commands | **14.1%** | 4,577 | 40% | +8-10% | **HIGHEST** | Phase 2 ✅ |
| output | **80%** | 224 | 80% | **+1.9%** | **COMPLETE** | Phase 1 ✅ |
| clone | 0% | 436 | 30% | +2-3% | **QUICK WIN** | Pending |
| orchestrator | 0.7% | 490 | 40% | +2% | **MEDIUM** | Pending |
| diagnostics | 13.7% | 798 | 40% | +2-3% | **MEDIUM** | Pending |
| config | 38.4% | 855 | 50% | +1% | **LOW** | Pending |
| dotfiles | 50.3% | 2,769 | 60% | +2% | **LOW** | Pending |
| testutil | **100%** | 46 | - | - | **FOUNDATION** | Phase 1 ✅ |

### Key Findings

1. **No Coverage Reporting Issues**: Config (38.4%) and dotfiles (50.3%) already have decent coverage
2. **Common Pattern**: Most packages need the existing `CommandExecutor` pattern extended
3. **No Tests**: Clone and output packages have 0% coverage
4. **Well-Tested**: lock (83.3%), packages (61.7%), resources (58.5%) exceed our targets

## Phase 1: Foundation & Quick Wins ✅ COMPLETE

### Add Minimal Test Infrastructure ✅

**Created `internal/testutil` package** with common test helpers:

```go
// internal/testutil/executor.go
package testutil

import "context"

// MockExecutor provides common command execution mocking
type MockExecutor struct {
    Commands []Command
    Results  map[string]Result
}

type Command struct {
    Name string
    Args []string
}

type Result struct {
    Output []byte
    Error  error
}

func (m *MockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
    m.Commands = append(m.Commands, Command{Name: name, Args: args})
    key := name + " " + strings.Join(args, " ")
    if r, ok := m.Results[key]; ok {
        return r.Output, r.Error
    }
    return nil, nil
}
```

```go
// internal/testutil/writer.go
package testutil

import "bytes"

// BufferWriter captures output for testing
type BufferWriter struct {
    Buffer bytes.Buffer
}

func (b *BufferWriter) Printf(format string, args ...interface{}) {
    fmt.Fprintf(&b.Buffer, format, args...)
}

func (b *BufferWriter) String() string {
    return b.Buffer.String()
}
```

**Benefits**:
- Shared test infrastructure
- Consistent mocking patterns
- Reduces duplication across packages

## Phase 2: Commands Package Pure Functions ✅ COMPLETE

### Commands Package

**Result**: 9.2% → 14.1% (+4.9% coverage)
**Approach**: Tested pure functions without mocking

**Completed tests**:
- ✅ `TruncateString` - string truncation with edge cases
- ✅ `GetMetadataString` - safe metadata extraction
- ✅ `CompleteDotfilePaths` - shell completion suggestions
- ✅ `NewStandardTableBuilder` - table formatting
- ✅ `formatVersion` - version string formatting
- ✅ `completeOutputFormats` - output format completion
- ✅ `getEditor` - editor detection from env vars
- ✅ `getConfigPath` - config file path construction
- ✅ `convertApplyResult` - apply result conversion

**Key achievement**: Improved coverage without complex mocking

## Phase 3: Commands Package Architecture Decision ✅ COMPLETE

### Decision: Do Not Unit Test Command Orchestration Functions

After critical analysis, we determined that command orchestration functions (`runXXX`) are not suitable for unit testing because:
- They directly instantiate dependencies with no injection points
- They are integration points by design, not business logic containers
- They mix multiple concerns (parsing, orchestration, output)
- Testing them would require significant production code changes

**Result**: Focus shifted to packages with actual business logic. Commands package will remain at ~15% coverage (pure functions only).

See [Architecture Decision Document](commands-testing-architecture-decision.md) for full rationale.

## Phase 4: High Priority Business Logic Packages

### Orchestrator Package (3 days) - NEW HIGH PRIORITY

**Current**: 0.7%
**Target**: 40%
**Approach**: This package contains core business logic and is highly testable

1. Test option functions (easy wins)
2. Test ReconcileAll with mocked packages/dotfiles
3. Test Apply with various scenarios
4. Test hook execution with mocked commands
5. Mock package and dotfile reconciliation

### Diagnostics Package (2 days)

**Current**: 13.7%
**Target**: 40%
**Approach**: Test pure functions and add minimal system mocks

1. Test shell detection logic
2. Test PATH manipulation
3. Test health check aggregation
4. Create minimal SystemInterface for external calls

## Phase 5: Additional Coverage

### Clone Package (1 day)

**Current**: 0%
**Target**: 30%
**Approach**: Test pure functions

1. Test Git URL parsing (pure function)
2. Test configuration generation
3. Skip actual git/network operations

## Implementation Guidelines

### ⚠️ MANDATORY SAFETY CHECKS ⚠️

Before writing ANY test, ask yourself:
1. **Could this test install a package?** If yes, DO NOT WRITE IT
2. **Could this test modify a dotfile?** If yes, DO NOT WRITE IT
3. **Could this test execute a shell command?** If yes, DO NOT WRITE IT
4. **Could this test write outside temp directories?** If yes, DO NOT WRITE IT
5. **Could this test change ANYTHING on the system?** If yes, DO NOT WRITE IT

### Principles

1. **Don't over-engineer** - Use simple mocks, not frameworks
2. **Test behavior, not implementation** - Focus on inputs/outputs
3. **Start simple** - Test pure functions first
4. **Reuse patterns** - Extend CommandExecutor pattern where it exists
5. **Preserve behavior** - No changes to CLI/UX

### Anti-Patterns to Avoid

❌ Complex mocking frameworks
❌ Testing third-party libraries
❌ 100% coverage goals
❌ Refactoring working code just for tests
❌ Breaking changes to add testability

### Test Patterns to Use

✅ Table-driven tests
✅ Temporary directories for file operations
✅ Package-level test hooks (existing pattern)
✅ Simple struct-based mocks
✅ Focus on error paths and edge cases

## Success Metrics

### Minimum Viable Coverage (v1.0)

| Package | Current | Target | Must Have |
|---------|---------|--------|-----------|
| Overall | 37.6% | 50% | ✓ |
| commands | 14.1% | ~15% | ✅ Pure functions only |
| output | 80% | 80% | ✅ Complete |
| orchestrator | 0.7% | 40% | ✓ HIGH PRIORITY |
| diagnostics | 13.7% | 40% | ✓ HIGH PRIORITY |
| clone | 0% | 30% | Nice to have |
| config | 38.4% | 50% | Nice to have |
| dotfiles | 50.3% | 60% | Nice to have |

### Risk Mitigation

1. **No reporting bugs found** - Coverage data is accurate
2. **Test incrementally** - Merge improvements daily
3. **Focus on stability** - Don't refactor working code
4. **Time-box efforts** - 2-3 weeks maximum
5. **Accept "good enough"** - 50% is sufficient for v1.0

## Post-v1.0 Improvements

These can wait:
- Integration tests for complex flows
- Full mock infrastructure
- 80%+ coverage goals
- Refactoring for better testability

## Next Steps

1. **Phase 1** ✅: testutil package + output package tests (Complete)
2. **Phase 2** ✅: Commands package pure functions (Complete)
3. **Phase 3** ✅: Architecture decision - commands not unit testable (Complete)
4. **Phase 4**: Orchestrator package - HIGH PRIORITY (40% target)
5. **Phase 5**: Diagnostics package improvements (40% target)
6. **Phase 6**: Clone package basic tests (30% target)

**Progress**: 32.7% → 37.6% (+4.9% achieved, need +12.4% more)

**Revised Strategy**: Focus on business logic packages (orchestrator, diagnostics) rather than CLI adapter layer (commands).

Total remaining effort: 6-8 days to reach 50% coverage for v1.0.
