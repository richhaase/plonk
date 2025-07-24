# Task 010: Delete Errors Package

## Objective
Delete the `internal/errors` package completely and replace all custom error handling with idiomatic Go error patterns, eliminating 766 LOC of over-engineering.

## Quick Context
- Current errors package: 766 LOC (321 implementation + 445 tests)
- Over-engineered system with domains, error codes, severity levels, metadata
- **Anti-pattern for Go CLIs** - most successful CLIs use standard library only
- Complex 6-parameter error constructors that add no value for CLI users

## Why Delete Rather Than Simplify?

### **Go CLI Ecosystem Analysis**
Popular Go CLIs **don't have dedicated error packages**:
- **kubectl**: Uses `fmt.Errorf()` and standard error wrapping
- **docker CLI**: Simple error messages with context
- **gh (GitHub CLI)**: Standard Go error patterns
- **hugo**: Minimal error helpers, mostly `fmt.Errorf()`

### **Current Anti-Idiomatic Patterns**
```go
// Current: 6-parameter nightmare
return errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages,
    "install", packageName, "failed to get package manager")

// Idiomatic Go: Clear and concise
return fmt.Errorf("install %s: package manager unavailable: %w", packageName, err)
```

### **Unused Complexity**
- **Error Codes**: No consumer checks `if err.Code == ErrFileNotFound`
- **Domains**: CLI users don't care about DomainPackages vs DomainDotfiles
- **Severity**: Meaningless for CLI - all errors are fatal to the operation
- **Metadata**: Unused structured data that complicates error handling

## Current Errors Package Analysis
```
internal/errors/
├── types.go          (321 LOC) - Complex PlonkError with 10+ fields
└── types_test.go     (445 LOC) - Tests for over-engineered system
```

**Total**: 766 LOC to be eliminated entirely

## Impact Analysis
Files that import `internal/errors`:
- Most command files: `add.go`, `install.go`, `uninstall.go`, `sync.go`, etc.
- Domain packages: `dotfiles`, `managers`, `ui`
- All use the over-engineered error constructors

## Work Required

### Phase 1: Audit All Error Usage
1. **Find all `errors.` calls** in the codebase
2. **Categorize error patterns** (wrap existing, create new, with context)
3. **Plan migration strategy** for each pattern type
4. **Identify any legitimate error handling** that needs preservation

### Phase 2: Replace with Standard Go Patterns
**Migration patterns**:

```go
// Pattern 1: Wrapping existing errors
// Before:
return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to copy file")
// After:
return fmt.Errorf("copy file: %w", err)

// Pattern 2: Creating new errors with context
// Before:
return errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "add", "dotfile not found")
// After:
return fmt.Errorf("add dotfile: file not found")

// Pattern 3: Complex contextual errors
// Before:
return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages,
    "install", packageName, "failed to install").WithMetadata("manager", mgr)
// After:
return fmt.Errorf("install %s via %s: %w", packageName, mgr, err)
```

### Phase 3: Update All Consumers
**Files to update** (estimate ~15-20 files):
- **Commands**: `add.go`, `install.go`, `uninstall.go`, `sync.go`, `rm.go`, etc.
- **Domain packages**: `dotfiles/`, `managers/`, `config/`
- **UI package**: Remove PlonkError type assertions
- **Any error checking code**: Replace with standard patterns

### Phase 4: Delete Package and Cleanup
1. **Remove `internal/errors/` directory** completely
2. **Remove all `internal/errors` imports**
3. **Verify no error code checking** remains
4. **Update any error matching** to use standard string methods

## Expected Benefits
- **766 LOC eliminated** (entire package deleted)
- **Package count**: 13 → 12
- **More idiomatic Go code** following community standards
- **Simpler error messages** that are easier to write and understand
- **Better developer experience** - no complex error constructors
- **Reduced cognitive load** - standard patterns everyone knows

## Error Message Quality
**Maintain or improve error message quality**:

```go
// Current complex but good message:
"failed to install package 'htop': manager 'brew' is not available"

// Idiomatic with same clarity:
fmt.Errorf("install htop: brew manager not available")

// Can add suggestions in context:
fmt.Errorf("install htop: brew not available (run: brew --version)")
```

## Success Criteria
1. ✅ **No `internal/errors` package remains**
2. ✅ **All error handling uses standard Go patterns**
3. ✅ **Error messages remain user-friendly and informative**
4. ✅ **No error codes, domains, or severity levels**
5. ✅ **All imports removed and files compile**
6. ✅ **All tests pass with new error patterns**
7. ✅ **766 LOC of complexity eliminated**

## Testing Strategy
- **Error message tests**: Verify error strings contain expected information
- **Error wrapping tests**: Ensure `errors.Is()` and `errors.As()` work with wrapped errors
- **No complex error matching**: Remove tests that check error codes/types

## Dependencies
- **Minimal overlap with Task 008**: Runtime package uses some error utilities
- **Can proceed after Task 008**: Orchestrator will use simplified error handling

## Completion Report
Create `TASK_010_COMPLETION_REPORT.md` with:
- **Before/after error handling examples** from each major command
- **Migration patterns used** for different error types
- **Error message quality comparison**
- **Code reduction metrics** (766 LOC eliminated)
- **Verification that all error scenarios still provide good UX**
