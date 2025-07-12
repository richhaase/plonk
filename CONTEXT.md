# Implementation Context for Multiple Add Features

## Overview

This document provides essential context for AI agents implementing the multiple package add and multiple dotfile add features. It covers overlapping needs, pre-work requirements, and essential knowledge about the plonk codebase.

## Overlapping Implementation Needs

### 1. **Common Result Types and Progress Patterns**
Both package and dotfile implementations need similar result structures and progress reporting:

```go
// Common pattern for both packages and dotfiles
type OperationResult struct {
    Name     string    // Package name or file path
    Status   string    // "added", "updated", "skipped", "failed", "would-add"
    Error    error
    // ... domain-specific fields
}

// Common progress reporting interface
type ProgressReporter interface {
    ShowProgress(result OperationResult)
    ShowSummary(results []OperationResult)
}
```

### 2. **Common Error Handling Patterns**
Both implementations need identical "continue on failure" logic and error suggestion formatting:

```go
// Shared utility functions
func formatErrorWithSuggestion(err error, itemName string, itemType string) string
func determineExitCode(results []OperationResult) error
func shouldContinueOnError(err error) bool
```

### 3. **Context Management**
Both use identical timeout and cancellation patterns:

```go
// Common context setup utility
func createOperationContext(timeout time.Duration) (context.Context, context.CancelFunc)
func checkCancellation(ctx context.Context) error
```

### 4. **Summary Display Logic**
Both need the same counting and summary formatting:

```go
// Shared summary utilities
func countByStatus(results []OperationResult, status string) int
func formatSummaryLine(added, updated, skipped, failed int) string
```

## Pre-work Requirements

### 1. **Enhanced PackageManager Interface**
The package implementation needs version information:

```go
// Add to internal/managers/common.go
type PackageManager interface {
    // Existing methods...
    GetInstalledVersion(ctx context.Context, name string) (string, error) // NEW
}
```

**Implementation Notes:**
- Update all package manager implementations (homebrew, npm, cargo)
- Update mock generation for new interface method
- Add tests for version retrieval functionality

### 2. **Common Operation Utilities Package**
Create `internal/operations/` for shared utilities:

```go
// internal/operations/common.go
type BatchProcessor interface {
    ProcessItems(ctx context.Context, items []string) ([]OperationResult, error)
}

type ProgressReporter interface {
    ShowItemProgress(result OperationResult)
    ShowBatchSummary(results []OperationResult)
}

// internal/operations/progress.go
func FormatErrorWithSuggestion(err error, itemName string, itemType string) string
func DetermineExitCode(results []OperationResult) error
func CountByStatus(results []OperationResult, status string) int
```

### 3. **Enhanced Error Message System**
Extend the existing error system with suggestion patterns:

```go
// Add to internal/errors/types.go
type ErrorSuggestion struct {
    Message string
    Command string
}

func (e *PlonkError) WithSuggestion(suggestion ErrorSuggestion) *PlonkError
```

## Essential Codebase Knowledge

### **Justfile Commands**
The project uses `just` as a command runner. Essential commands for development:

```bash
# Development setup
just dev-setup          # Complete environment setup (dependencies, hooks, mocks, tests)
just generate-mocks     # Regenerate mocks after interface changes
just deps-update        # Update dependencies with safety checks

# Development workflow
just build              # Build binary with version info
just test               # Run unit tests
just test-coverage      # Run tests with coverage report
just precommit          # Run all quality checks (format, lint, test, security)

# Code quality
just format             # Format code and organize imports
just lint               # Run golangci-lint
just security           # Run govulncheck and gosec

# Release process
just release-check      # Validate GoReleaser configuration
just release-snapshot   # Test release build (no publishing)
just release v1.2.3     # Create actual release
```

### **Key Libraries and Patterns**

1. **Mock Generation**: Uses `go.uber.org/mock/mockgen` (configured in justfile)
2. **Error Handling**: Structured errors with `internal/errors/types.go`
3. **Context Pattern**: All operations accept `context.Context` for cancellation
4. **Testing Pattern**: Table-driven tests with mocks
5. **CLI Framework**: Uses `cobra` (seen in command files)
6. **Configuration**: Interface-based with YAML implementation

### **Existing Code Patterns to Follow**

1. **Command Structure**: Follow `internal/commands/pkg_add.go` pattern
   - Use cobra.Command with proper argument validation
   - Handle dry-run mode consistently
   - Support structured output (JSON/YAML)

2. **Error Creation**: Use structured error system
   ```go
   // Create new errors
   errors.NewError(code, domain, operation, message)

   // Wrap existing errors
   errors.Wrap(err, code, domain, operation, message)

   // Add item context
   errors.WrapWithItem(err, code, domain, operation, item, message)
   ```

3. **Interface Usage**: Mock interfaces for testing
   - All external dependencies are mockable
   - Use dependency injection for testability
   - See `mock_manager.go` for patterns

4. **Config Access**: Use existing config interfaces
   - Never access YAML directly
   - Use `ConfigReader`, `ConfigWriter` interfaces
   - Support zero-config defaults

5. **Output Formatting**: Follow `internal/commands/output.go` patterns
   - Support table, JSON, YAML output formats
   - Consistent formatting across commands

### **Testing Infrastructure**

1. **Existing Mocks**:
   - `MockPackageManager` (internal/managers/)
   - `MockConfigReader` (internal/config/)
   - `MockProvider` (internal/state/)

2. **Test Helpers**:
   - `t.TempDir()` for isolated test environments
   - Table-driven test patterns throughout codebase
   - Mock setup utilities in existing test files

3. **Context Testing**:
   - Examples in existing tests for cancellation scenarios
   - Timeout testing patterns
   - Error propagation testing

4. **Test Organization**:
   - Tests alongside source code as `*_test.go`
   - Comprehensive error scenario testing
   - Backward compatibility verification

### **Quality Assurance Infrastructure**

#### **Pre-commit Hooks**
The project uses the pre-commit framework for automated quality checks:

- **Go formatting**: `goimports` for code formatting and import organization
- **Linting**: `golangci-lint` for comprehensive static analysis
- **Testing**: Automated unit test execution
- **Security**: `govulncheck` for vulnerability scanning
- **File checks**: YAML/TOML syntax, trailing whitespace, end-of-file fixing

**Benefits:**
- 94% faster on non-Go file changes (file-specific filtering)
- Automatic hook updates via `just deps-update`
- Rich error reporting with file context
- Prevents common issues before commit

#### **GitHub Actions**
Automated CI/CD pipeline includes:

- **Cross-platform testing**: Linux, macOS, Windows
- **Multiple Go versions**: Ensures compatibility
- **Security scanning**: `gosec` and `govulncheck`
- **Release automation**: GoReleaser for multi-platform binaries
- **Code coverage**: Integration with coverage reporting

**Key workflows:**
- Pull request validation
- Automated releases on tag push
- Security vulnerability scanning
- Dependency updates with safety checks

### **Development Environment**

#### **Required Tools**
- **Go 1.24.4+** (see `.tool-versions`)
- **Just** (command runner)
- **Git** (version control)
- **Pre-commit** (for git hooks)

#### **Optional but Recommended**
- **GoReleaser** (for release testing)
- **golangci-lint** (for local linting)

#### **IDE Integration**
The project includes configurations for:
- Go module structure with proper imports
- Consistent code formatting via `goimports`
- Mock generation integration
- Test discovery and execution

## Implementation Order Recommendations

### **Phase 1: Pre-work (Required First)**
1. **Add `GetInstalledVersion()` to PackageManager interface**
   - Update `internal/managers/common.go`
   - Implement in all package managers (homebrew, npm, cargo)
   - Update mocks with `just generate-mocks`
   - Add tests for version retrieval

2. **Create shared utilities in `internal/operations/`**
   - Common result types and interfaces
   - Progress reporting utilities
   - Error suggestion formatting
   - Summary display logic

3. **Extend error system**
   - Add suggestion support to PlonkError
   - Create helper functions for common error scenarios

### **Phase 2: Implementation**
1. **Package Implementation** (simpler, establishes patterns)
   - Implement multiple package add functionality
   - Establish testing patterns
   - Validate shared utilities

2. **Dotfile Implementation** (builds on package patterns)
   - Leverage established patterns
   - Add file-specific functionality (glob expansion, attribute preservation)

### **Phase 3: Documentation**
1. **CLI Documentation**: Update CLI.md with multiple add examples
2. **Command Help**: Update cobra command descriptions
3. **Examples**: Update README.md with new capabilities
4. **Architecture**: Document shared utilities in ARCHITECTURE.md

## Testing Strategy

### **Unit Tests**
- Test all new functionality with mocks
- Maintain existing test coverage standards
- Use table-driven tests for multiple scenarios
- Test error conditions and edge cases

### **Integration Tests**
- Verify backward compatibility
- Test with real (but isolated) file systems
- Validate context cancellation behavior
- Ensure proper mock usage

### **Quality Checks**
- Run `just precommit` before any commit
- Verify no regressions in existing functionality
- Test with different output formats (table, JSON, YAML)
- Validate structured error handling

## Documentation Updates Required

Both implementations will need to update:

1. **CLI.md**: Add multiple add examples and syntax documentation
2. **Command help text**: Update `Use` and `Long` descriptions in cobra commands
3. **README.md**: Update quick start examples to show multiple add capability
4. **ARCHITECTURE.md**: Document new shared utilities and patterns

## Success Criteria

1. **Functionality**: Both commands support multiple arguments with identical UX patterns
2. **Compatibility**: All existing single-argument usage continues to work unchanged
3. **Quality**: All tests pass, no linting issues, security checks clear
4. **Documentation**: Complete and accurate documentation for new features
5. **Performance**: Efficient sequential processing with proper error handling
