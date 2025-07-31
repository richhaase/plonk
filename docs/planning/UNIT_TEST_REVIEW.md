# Unit Test Review Findings

**Date**: 2025-07-31
**Reviewer**: Assistant
**Scope**: Unit tests only (excluding integration and BATS tests)

## Executive Summary

The unit test review revealed several critical issues:
1. **External dependencies in tests**: Multiple unit tests make actual external calls
2. **Low coverage**: Average ~30% coverage with some packages as low as 4.5%
3. **Missing business logic tests**: Key business logic lacks unit tests
4. **Test organization**: Some tests are misclassified (integration tests in unit test files)

## Coverage Analysis

### Coverage by Package

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/lock` | 83.3% | ✅ Good |
| `internal/resources` | 58.5% | ⚠️ Moderate |
| `internal/resources/dotfiles` | 50.3% | ⚠️ Moderate |
| `internal/config` | 22.1% | ❌ Low |
| `internal/resources/packages` | 22.6% | ❌ Low |
| `internal/orchestrator` | 16.3% | ❌ Very Low |
| `internal/commands` | 4.5% | ❌ Critical |
| `internal/diagnostics` | 0.0% | ❌ No tests |
| `internal/output` | 0.0% | ❌ No tests |
| `internal/setup` | 0.0% | ❌ No tests |

### Files with No Coverage

- `internal/config/constants.go` - Configuration constants
- `internal/resources/resource.go` - Core resource abstraction
- `internal/resources/packages/interfaces.go` - Package manager interfaces
- `internal/lock/interfaces.go` - Lock service interfaces
- `internal/lock/types.go` - Lock file types

## Critical Finding: External Dependencies in Unit Tests

### 1. Package Manager Tests Making External Calls

**File**: `internal/resources/packages/capability_test.go`
- **Issue**: `TestCapabilityDiscoveryPattern` calls actual package managers
- **Lines**: 87-119 - Makes real calls to `IsAvailable()`, `Search()`
- **Impact**: Tests fail without package managers installed, slow execution

### 2. File System Operations

While file system operations are common in unit tests, many tests use real file I/O when they could use in-memory abstractions:
- Lock service tests use `os.MkdirTemp` and real file operations
- Config tests write actual files
- Dotfile tests perform real file operations

### 3. Command Execution

No direct `exec.Command` calls found in unit tests (good), but package manager tests indirectly execute commands through their interfaces.

## Business Logic Lacking Tests

### High-Complexity Files with Low/No Coverage

1. **`internal/diagnostics/health.go`** (730 lines, 0% coverage)
   - Critical health check logic
   - Platform-specific path detection
   - Package manager verification

2. **`internal/setup/setup.go`** (426 lines, 0% coverage)
   - Repository cloning logic
   - Git operations
   - Initial setup flow

3. **`internal/commands/status.go`** (524 lines, ~4.5% coverage)
   - Complex state reconciliation
   - Output formatting logic
   - Multi-format support (table/json/yaml)

4. **`internal/commands/apply.go`** (254 lines, ~4.5% coverage)
   - Core apply orchestration
   - Flag validation
   - Error handling

5. **`internal/orchestrator/apply.go`** (244 lines, ~16% coverage)
   - Dotfile deployment logic
   - Package installation orchestration
   - Hook execution

## Test Quality Issues

### 1. Misclassified Tests

**File**: `internal/orchestrator/integration_test.go`
- Despite the name, this is in the unit test suite
- Makes real file system operations
- Should be moved to integration tests or refactored

### 2. Missing Mock Usage

Most tests that interact with external systems don't use mocks:
- Package manager operations
- File system operations
- Git operations

### 3. Test Naming and Organization

- Test names are generally good (descriptive)
- Table-driven tests are used well where present
- Missing consistent use of test helpers

## Recommendations

### Immediate Actions

1. **Fix External Dependencies**
   - Refactor `capability_test.go` to use mocks
   - Create mock implementations for `PackageManager` interface
   - Use `afero` or similar for file system abstraction

2. **Add Critical Business Logic Tests**
   - Health diagnostics logic
   - Status command reconciliation
   - Apply command orchestration
   - Setup/clone operations

3. **Improve Test Organization**
   - Move `integration_test.go` to proper location
   - Create clear separation between unit and integration tests
   - Add build tags consistently

### Coverage Targets

While no specific target was set, recommended minimums for v1.0:
- Critical business logic: 80%+
- Commands: 60%+
- Core libraries (lock, config): 80%+
- Overall: 70%+

### Test Improvements

1. **Mock Creation**
   ```go
   // Example mock for PackageManager
   type MockPackageManager struct {
       mock.Mock
   }

   func (m *MockPackageManager) IsAvailable(ctx context.Context) (bool, error) {
       args := m.Called(ctx)
       return args.Bool(0), args.Error(1)
   }
   ```

2. **File System Abstraction**
   ```go
   // Use afero for file operations
   fs := afero.NewMemMapFs()
   // Use fs instead of os package
   ```

3. **Test Helpers**
   ```go
   // Create common test fixtures
   func createTestConfig(t *testing.T) *config.Config {
       t.Helper()
       // Return consistent test config
   }
   ```

## Summary

The unit tests need significant improvement before v1.0:
1. Remove all external dependencies
2. Increase coverage on critical business logic
3. Properly organize and classify tests
4. Add comprehensive mocking

Estimated effort: 2-3 days to address critical issues.
