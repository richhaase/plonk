# BATS vs Go Integration Tests Analysis

## Current State

### BATS Tests (10 files)
Located in `tests/bats/behavioral/`:
- 00-smoke.bats - Basic command availability
- 01-basic-commands.bats - Core commands (status, config, etc.)
- 02-output-formats.bats - JSON/YAML/table output
- 03-package-install.bats - Package installation flows
- 04-package-uninstall.bats - Package removal flows
- 05-dotfile-add.bats - Dotfile management
- 06-dotfile-rm.bats - Dotfile removal
- 07-apply-behavior.bats - Apply command behavior
- 99-cleanup-all.bats - Test cleanup
- test-go-debug.bats - Development/debug tests

### Go Integration Tests (5 files)
Located in `tests/integration/`:
- crossplatform_test.go - Cross-platform behavior
- dotfiles_test.go - Dotfile operations
- helpers_test.go - Test utilities
- package_capability_test.go - Package manager capabilities
- package_managers_test.go - Package manager detection

## Coverage Comparison

### BATS Coverage
**Strengths:**
- Tests actual user workflows end-to-end
- Tests real package installations (brew:cowsay, npm:lodash)
- Tests CLI output and error messages
- Tests interactive features (prompts, editor)
- Easy to write and understand
- Good for acceptance testing

**What BATS Tests Well:**
- Command-line interface behavior
- Output formatting
- Error messages and user feedback
- Multi-step workflows
- Shell integration (environment variables, PATH)

### Go Integration Test Coverage
**Strengths:**
- Faster execution (can mock external commands)
- Better isolation (temp directories, controlled environment)
- Type-safe test assertions
- Can test internal state
- Better debugging capabilities
- Native Go tooling (coverage, profiling)

**What Go Tests Well:**
- API contracts
- Edge cases and error conditions
- Performance characteristics
- Concurrent operations
- Internal state management

## Pros and Cons

### BATS Tests

**Pros:**
- Simple to write (shell-like syntax)
- Tests exactly what users experience
- Good for documentation (tests read like tutorials)
- Tests shell integration naturally
- No compilation needed

**Cons:**
- Slower (spawns real processes)
- Harder to isolate (affects real system)
- Limited assertion capabilities
- No type safety
- Harder to debug failures
- Can't measure code coverage
- Requires BATS runtime

### Go Integration Tests

**Pros:**
- Fast execution
- Full isolation capability
- Rich assertion libraries (testify)
- Code coverage metrics
- Native debugging
- Type safety
- Better CI integration
- Can test internal APIs

**Cons:**
- More complex to write
- Need to spawn processes for CLI testing
- Harder to test interactive features
- More code to maintain
- Less readable for non-developers

## Current Test Redundancy

Significant overlap exists:
- Both test package installation
- Both test dotfile operations
- Both test output formats
- Both test error conditions

## Recommendation

### Short Term (v1.0)
Keep both, but with clear separation:
- **BATS**: User acceptance tests, happy paths, documentation
- **Go**: Edge cases, error conditions, performance

### Long Term (v2.0+)
Migrate to Go integration tests as primary:

1. **Port BATS tests to Go** using a pattern like:
```go
func TestInstallPackage(t *testing.T) {
    cmd := exec.Command(plonkBinary, "install", "brew:htop")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)
    require.Contains(t, string(output), "added")
}
```

2. **Keep minimal BATS suite** for:
   - Smoke tests
   - User acceptance criteria
   - Documentation examples

3. **Benefits of Go-primary approach:**
   - Single test framework
   - Unified coverage reporting
   - Better maintainability
   - Faster test execution
   - Better CI integration

## Migration Strategy

1. **Phase 1**: Identify unique BATS test scenarios
2. **Phase 2**: Port non-redundant tests to Go
3. **Phase 3**: Create test helpers for common patterns
4. **Phase 4**: Deprecate redundant BATS tests
5. **Phase 5**: Keep curated BATS suite for acceptance

## Coverage Gaps in Both

Neither BATS nor Go integration tests cover:
- `plonk clone` command (most critical gap)
- Error recovery scenarios
- Concurrent command execution
- Performance under load
- Large-scale operations (1000+ packages)
- Network failure handling
- Disk space exhaustion

## Conclusion

BATS provides better user-perspective testing but Go integration tests offer superior maintainability and tooling. The ideal approach is to migrate primary testing to Go while keeping a focused BATS suite for user acceptance testing.
