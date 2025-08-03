# Consolidated Test Coverage Improvement Plan

**Date**: 2025-08-03
**Current Coverage**: 34.6% (up from 32.7%)
**Target Coverage**: 50% (minimum for v1.0)
**Timeline**: 2-3 weeks
**Progress**: Phase 1 Foundation complete

## Executive Summary

After reviewing all test improvement documentation and verifying actual coverage, this plan consolidates the approach into practical, low-risk improvements that can achieve 50% coverage for v1.0. The strategy focuses on:

1. **No coverage reporting bugs found** - Config (38.4%) and dotfiles (50.3%) already have decent coverage
2. **Quick wins** - Add tests to packages with 0% coverage (output, clone)
3. **High-impact packages** - Focus on commands package (4,577 LOC at 9.2%)
4. **Pragmatic targets** - Need +15.4% to reach 50% total (was +17.3%)

## Current State Analysis

### Coverage by Priority

| Package | Current | LOC | Realistic Target | Coverage Gain | Impact | Status |
|---------|---------|-----|------------------|---------------|--------|--------|
| commands | 9.2% | 4,577 | 40% | +10-12% | **HIGHEST** | Pending |
| output | **80%** | 224 | 80% | **+1.9%** | **COMPLETE** | ✅ |
| clone | 0% | 436 | 30% | +2-3% | **QUICK WIN** | Pending |
| orchestrator | 0.7% | 490 | 40% | +2% | **MEDIUM** | Pending |
| diagnostics | 13.7% | 798 | 40% | +2-3% | **MEDIUM** | Pending |
| config | 38.4% | 855 | 50% | +1% | **LOW** | Pending |
| dotfiles | 50.3% | 2,769 | 60% | +2% | **LOW** | Pending |
| testutil | **100%** | 46 | - | - | **FOUNDATION** | ✅ |

### Key Findings

1. **No Coverage Reporting Issues**: Config (38.4%) and dotfiles (50.3%) already have decent coverage
2. **Common Pattern**: Most packages need the existing `CommandExecutor` pattern extended
3. **No Tests**: Clone and output packages have 0% coverage
4. **Well-Tested**: lock (83.3%), packages (61.7%), resources (58.5%) exceed our targets

## Phase 1: Foundation & Quick Wins (Week 1) ✅ COMPLETE

### Add Minimal Test Infrastructure (2 days) ✅

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

## Phase 2: High-Impact Improvements (Week 2)

### Commands Package (5 days)

**Current**: 9.2% (421/4,577 statements)
**Target**: 40% (~1,830 statements)
**Approach**: Test pure functions and add minimal mocks

1. **Test all flag parsing** (quick win)
2. **Test validation functions**
3. **Test output formatting**
4. **Mock CommandExecutor for basic flows**
5. **Skip complex orchestration initially**

**Example improvements**:
```go
// Easy to test - pure functions
func TestValidatePackageName(t *testing.T)
func TestFormatPackageList(t *testing.T)
func TestParseFlags(t *testing.T)

// Medium - use MockExecutor
func TestInstallCommand_DryRun(t *testing.T)
func TestStatusCommand_Output(t *testing.T)
```

### Output Package (2 days)

**Current**: 0%
**Target**: 80%
**Approach**: Add writer interface (already planned)

1. Implement BufferWriter from testutil
2. Test all status words and formatting
3. Test progress output

**Expected gain**: +2% overall coverage

## Phase 3: Medium Priority (Week 3)

### Diagnostics Package (2 days)

**Current**: 13.7%
**Target**: 40%
**Approach**: Test pure functions first

1. Test shell detection logic
2. Test PATH manipulation
3. Test health check aggregation
4. Mock system calls only where necessary

### Orchestrator Package (2 days)

**Current**: 0%
**Target**: 40%
**Approach**: Mock external dependencies

1. Test option functions (easy)
2. Test result conversion
3. Mock package/dotfile operations
4. Skip complex Apply method initially

### Clone Package (1 day)

**Current**: 0%
**Target**: 30%
**Approach**: Test only pure functions

1. Test Git URL parsing (pure function)
2. Test configuration generation
3. Skip actual git/network operations

## Implementation Guidelines

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
| Overall | 32.7% | 50% | ✓ |
| commands | 9.2% | 40% | ✓ |
| output | 0% | 80% | ✓ |
| diagnostics | 13.7% | 40% | ✓ |
| clone | 0% | 30% | Nice to have |
| orchestrator | 0.7% | 40% | Nice to have |
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

1. **Week 1**: Add testutil package + output package tests
2. **Week 2**: Commands package improvements
3. **Week 3**: Diagnostics + remaining quick wins
4. **Review**: Verify 50% target is met

Total effort: 10-15 days to reach 50% coverage for v1.0 (reduced since no reporting bugs exist).
