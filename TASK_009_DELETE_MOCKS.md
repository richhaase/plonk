# Task 009: Delete Mocks Package

## Objective
Delete the `mocks` package and replace generated mocks with simple test doubles, eliminating ~300 LOC of generated complexity.

## Quick Context
- Current mocks package: 3 generated files with brittle mock expectations
- Only used in `internal/testing/helpers.go` - very contained impact
- Generated mocks add complexity without benefit for a CLI tool
- Simple test doubles are more reliable and easier to understand

## Current Mocks Package Analysis
```
internal/mocks/
├── config_mocks.go      (~100 LOC generated)
├── core_mocks.go        (~150 LOC generated)
└── package_manager_mocks.go (~80 LOC generated)
```

**Only Usage**: `internal/testing/helpers.go` - 1 file to update

## Problems with Current Mocks
1. **Generated Complexity**: 330+ LOC of generated code that's hard to understand
2. **Brittle Tests**: Mock expectations break easily when implementation changes
3. **Over-Engineering**: CLI tool doesn't need sophisticated mocking framework
4. **Maintenance Burden**: Generated code requires mockgen tooling

## **CRITICAL REQUIREMENT: Protect Developer Environment**

**The mocks serve a crucial safety function**: They prevent unit tests from running actual package manager commands that could modify a developer's system.

**Current Safety Pattern**:
```go
// internal/testing/helpers.go MockManagerSetup prevents real commands:
func (ms *MockManagerSetup) WithAvailability(available bool) *MockManagerSetup {
    ms.Manager.EXPECT().IsAvailable(gomock.Any()).Return(available, nil).AnyTimes()
    return ms  // Returns mock, NOT real manager that would run brew/npm/etc
}
```

**This is essential** - without mocks, a developer running `go test ./...` could accidentally:
- Install/uninstall packages on their system
- Modify their dotfiles
- Break their development environment

## Target Simple Test Doubles
Replace generated mocks with simple, safe test implementations:

```go
// Safe test doubles that NEVER run real commands:
type TestPackageManager struct {
    Name         string
    Available    bool
    Installed    []string
    InstallError error
    // NEVER calls real brew/npm/cargo/etc commands
}

func (t *TestPackageManager) IsAvailable(ctx context.Context) (bool, error) {
    return t.Available, nil  // Returns test data, no real system calls
}

func (t *TestPackageManager) Install(ctx context.Context, pkg string) error {
    return t.InstallError    // Simulates result, no real installation
}

func (t *TestPackageManager) ListInstalled(ctx context.Context) ([]string, error) {
    return t.Installed, nil  // Returns test data, no real package queries
}
```

## Work Required

### Phase 1: Analyze Current Usage
1. Review `internal/testing/helpers.go` to understand exactly how mocks are used
2. Identify the 3-4 test scenarios that need simple test doubles
3. Document the interface contracts needed

### Phase 2: Create Safe Test Doubles
Replace generated mocks with simple, SAFE structs in `internal/testing/helpers.go`:
1. **SafeTestPackageManager** - Returns test data, NEVER runs real package managers
2. **SafeTestConfigManager** - Operates on test directories only
3. **SafeTestDotfileProvider** - Works with test files only, never touches real dotfiles

**Critical Safety Requirements**:
- ✅ NEVER call exec.Command() with real package manager binaries
- ✅ NEVER touch files outside test temp directories
- ✅ NEVER modify real dotfiles or system packages
- ✅ All test doubles return controlled, predictable test data

### Phase 3: Update Tests with Safety Verification
1. Replace mock expectations with direct field assignments
2. Use simple assertions instead of complex mock verifications
3. **VERIFY** that no tests can touch the real system
4. Add comments documenting the safety of each test double

### Phase 4: Delete Mocks Package
1. Remove `internal/mocks/` directory completely
2. Remove mockgen references from build files/justfile
3. Verify no remaining imports

## Expected Code Changes
- **Before**: 330+ LOC of generated mocks + complex test setup
- **After**: ~50 LOC of simple test doubles with clear assertions
- **Net Reduction**: ~280 LOC eliminated
- **Package Count**: 13 → 12

## Success Criteria
1. ✅ **SAFETY FIRST**: Unit tests NEVER run real package manager commands or modify system
2. ✅ All tests pass with simple, safe test doubles
3. ✅ Test doubles only return controlled test data, never touch real files/packages
4. ✅ No generated mock code remains
5. ✅ Tests are more readable and maintainable
6. ✅ No mockgen tooling dependencies
7. ✅ Mocks package completely deleted

## **SAFETY VERIFICATION CHECKLIST**
Before completing Task 009, verify:
- [ ] No test calls real `brew`, `npm`, `cargo`, `pip`, `gem`, or `go install` commands
- [ ] No test modifies files outside of `t.TempDir()` directories
- [ ] No test can break a developer's dotfiles or installed packages
- [ ] All TestPackageManager methods return hardcoded test data only
- [ ] Running `go test ./...` is completely safe on any developer machine

## Completion Report
Create `TASK_009_COMPLETION_REPORT.md` with:
- Analysis of mock usage patterns
- Before/after test code comparison
- List of simplified test cases
- Verification that all tests pass
- Code reduction metrics
