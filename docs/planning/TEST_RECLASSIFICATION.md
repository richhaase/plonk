# Test Reclassification Report

**Date**: 2025-07-31
**Task**: Move tests with external dependencies to integration tests

## Summary

Successfully reclassified tests to ensure unit tests are truly isolated with no external dependencies.

## Files Moved

### 1. Package Manager Capability Test
**From**: `internal/resources/packages/capability_test.go`
**To**: `tests/integration/packages/capability_test.go`
**Reason**: Makes real calls to package managers (`IsAvailable()`, `Search()`)

### 2. Orchestrator Integration Test
**From**: `internal/orchestrator/integration_test.go`
**To**: `tests/integration/orchestrator/integration_test.go`
**Reason**: Already named as integration test, performs file I/O operations

## Changes Made

1. **Added build tags** to moved files:
   ```go
   //go:build integration
   // +build integration
   ```

2. **Updated package declarations**:
   - Changed from `package packages` to `package integration_test`
   - Added proper imports for packages being tested

3. **Fixed type references**:
   - Updated `PackageManager` to `packages.PackageManager`
   - Updated constructor calls to include package prefix

## Test Results

### Unit Tests (after reclassification)
```bash
$ go test ./internal/... -tags='!integration'
ok  	github.com/richhaase/plonk/internal/commands	0.405s
ok  	github.com/richhaase/plonk/internal/config	0.245s
ok  	github.com/richhaase/plonk/internal/lock	0.598s
ok  	github.com/richhaase/plonk/internal/orchestrator	0.826s
ok  	github.com/richhaase/plonk/internal/resources	0.974s
ok  	github.com/richhaase/plonk/internal/resources/dotfiles	0.820s
ok  	github.com/richhaase/plonk/internal/resources/packages	0.842s
```

### Integration Tests
```bash
$ go test ./tests/integration/... -tags=integration
ok  	github.com/richhaase/plonk/tests/integration	0.179s
ok  	github.com/richhaase/plonk/tests/integration/orchestrator	0.470s
ok  	github.com/richhaase/plonk/tests/integration/packages	10.535s
```

## Remaining Pure Unit Tests

The following test files remain as unit tests because they test pure logic:

### Package Manager Tests (Parse-only)
- `cargo_test.go` - Tests `parseListOutput`, `parseSearchOutput`, `parseInfoOutput`
- `gem_test.go` - Tests output parsing functions
- `homebrew_test.go` - Tests output parsing functions
- `npm_test.go` - Tests output parsing functions
- `pip_test.go` - Tests output parsing functions
- `goinstall_test.go` - Tests configuration and module path parsing

### Other Pure Tests
- `config_test.go` - Configuration loading and validation
- `lock/yaml_lock_test.go` - Lock file operations (uses temp files)
- `resources/reconcile_test.go` - Pure reconciliation logic
- `dotfiles/*_test.go` - File operations (uses temp directories)

## Benefits

1. **Faster unit tests**: No external dependencies means tests run quickly
2. **Reliable CI**: Unit tests won't fail due to missing package managers
3. **Clear separation**: Easy to run just unit tests or just integration tests
4. **Better organization**: Tests are where they logically belong

## Next Steps

1. Add more pure unit tests for business logic
2. Focus on testing parsing, validation, and state management
3. Leave external behavior testing to integration tests
