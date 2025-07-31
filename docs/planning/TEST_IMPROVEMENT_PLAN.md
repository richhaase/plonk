# Test Coverage Improvement Plan

**Goal**: Achieve 60%+ unit test coverage
**Current**: 9.3% unit, 28.3% integration
**Timeline**: 2-3 days

## Strategy

### 1. Focus on Unit Tests (Primary Goal)
Unit tests should test business logic in isolation without external dependencies. They're fast, reliable, and provide the best ROI for coverage.

### 2. Reduce Test Redundancy
- BATS tests and Go integration tests overlap significantly
- Keep BATS for user acceptance testing
- Use Go integration tests for complex scenarios BATS can't handle
- Focus effort on unit tests

### 3. Test Pyramid
```
         /\
        /  \  BATS (5-10 tests) - User acceptance
       /    \
      /      \  Integration (10-20 tests) - Complex scenarios
     /        \
    /          \  Unit Tests (100+ tests) - Business logic
   /____________\
```

## High-Value Unit Test Targets

### Phase 1: Commands Package (Target: 50%+ coverage)
Each command should have unit tests for:
- Flag validation
- Business logic
- Error cases
- Output formatting

**Priority Files**:
1. `install.go` - Test package parsing, validation, error handling
2. `uninstall.go` - Test removal logic, state management
3. `clone.go` - Test URL parsing, validation (mock git operations)
4. `apply.go` - Test scope logic, orchestration setup
5. `search.go` - Test query validation, result aggregation
6. `info.go` - Test package lookup logic
7. `config.go` - Test show/edit logic
8. `diff.go` - Test diff generation logic

**Test Pattern**:
```go
func TestInstallCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        flags   map[string]interface{}
        wantErr bool
        mock    func(*mockPackageManager)
    }{
        // Test cases for different scenarios
    }
}
```

### Phase 2: Clone Package (Target: 40%+ coverage)
Mock external dependencies:
- Git operations
- User prompts
- Package manager checks

**Key Functions**:
1. `parseGitURL()` - URL validation and parsing
2. `DetectRequiredManagers()` - Lock file analysis
3. `createDefaultConfig()` - Config generation
4. Error handling paths

### Phase 3: Resources Package (Target: 60%+ coverage)
**Dotfiles**:
- Path resolution
- State calculation
- Deployment logic (mock filesystem)

**Packages**:
- Each manager's parsing logic
- State reconciliation
- Error handling

### Phase 4: Core Packages (Target: 70%+ coverage)
1. **Config**: Validation, loading, saving
2. **Lock**: Reading, writing, migration
3. **Orchestrator**: Coordination logic
4. **Output**: Formatting, progress

## Implementation Approach

### Day 1: Command Unit Tests
1. Create test helpers and mocks
2. Add comprehensive tests for install/uninstall
3. Add tests for apply/status
4. Target: Commands package to 50%+ coverage

### Day 2: Clone and Resources
1. Mock git operations and prompts
2. Test clone package thoroughly
3. Test resource business logic
4. Target: Overall to 40%+ coverage

### Day 3: Remaining Gaps
1. Test config/lock packages
2. Test orchestrator logic
3. Fill remaining critical gaps
4. Target: Overall to 60%+ coverage

## Testing Patterns to Use

### 1. Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        // cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### 2. Mock Interfaces
```go
type mockPackageManager struct {
    mock.Mock
}

func (m *mockPackageManager) Install(ctx context.Context, pkg string) error {
    args := m.Called(ctx, pkg)
    return args.Error(0)
}
```

### 3. Test Fixtures
```go
func TestWithConfig(t *testing.T) {
    cfg := &config.Config{
        DefaultManager: "brew",
        // test config
    }
    // test with known config
}
```

## What NOT to Unit Test
- External command execution (use mocks)
- File I/O (use interfaces)
- Network calls (use mocks)
- User interaction (use mocks)

## Success Metrics
- Unit test coverage: 60%+
- Fast test execution: < 5 seconds for all unit tests
- No flaky tests
- Clear test names that document behavior

## Next Steps
1. Set up mock infrastructure
2. Start with high-value command tests
3. Measure coverage after each file
4. Adjust plan based on progress
