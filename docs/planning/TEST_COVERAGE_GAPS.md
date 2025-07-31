# Test Coverage Gap Analysis

**Date**: 2025-07-31
**Current Coverage**: Unit tests: 9.3%, Integration tests: 28.3%
**Combined Estimate**: ~30-35% (with overlap)

## Executive Summary

The current test coverage is critically low for a v1.0 release. With less than 35% combined coverage, we have significant blind spots that pose risks for:
- Undetected regressions
- Difficult refactoring
- Poor confidence in releases
- Higher bug escape rate

## Coverage by Package

### 1. Commands Package (6.6% unit coverage)
**Critical Gaps:**
- `clone.go` - 0% coverage (major user-facing feature)
- `install.go` - 0% coverage (core functionality)
- `uninstall.go` - 0% coverage (core functionality)
- `search.go` - 0% coverage
- `info.go` - 0% coverage
- `diff.go` - 0% coverage
- `config.go` - Minimal coverage (only edit path tested)
- `apply.go` - Only scope determination tested
- `status.go` - Only sorting functions tested

**What IS tested:**
- Basic output utilities (6.6%)
- Some helper functions

### 2. Clone Package (0% coverage) - formerly setup
**Complete blind spot:**
- Git URL parsing
- Repository cloning
- Package manager detection from lock file
- Package manager installation
- User prompts and interaction
- Tool installation (cargo, etc.)

**Risk**: This is a critical user journey with zero test coverage.

### 3. Resources Package
#### Dotfiles (partial coverage)
**Gaps:**
- Error handling paths
- Edge cases (symlinks, permissions)
- Conflict resolution

#### Packages (partial coverage)
**Gaps:**
- Individual package manager implementations
- Install/uninstall actual execution
- Error handling and recovery
- Package name parsing edge cases

### 4. Config Package (38.4% unit coverage)
**What's tested:**
- Default value application
- Basic loading

**Gaps:**
- Config validation
- Config saving
- Error scenarios
- Migration between versions

### 5. Orchestrator Package (16.3% coverage)
**Critical Gaps:**
- Apply orchestration error paths
- Hook execution
- State management edge cases
- Rollback scenarios
- Progress reporting

### 6. Diagnostics Package (13.7% unit coverage)
**What's tested:**
- Basic health calculations
- PATH generation

**Gaps:**
- Actual diagnostic checks
- Fix operations
- Platform-specific logic

### 7. Lock Package (minimal coverage)
**Gaps:**
- Lock file corruption handling
- Version migration
- Concurrent access
- Write operations

### 8. Output Package (0% coverage)
**Complete blind spot:**
- Spinner functionality
- Progress indicators
- Error formatting

## Integration Test Gaps

### Missing Command Tests:
1. **plonk clone** - Entire user onboarding flow untested
2. **plonk install** - Package installation scenarios
3. **plonk uninstall** - Package removal scenarios
4. **plonk search** - Search functionality across managers
5. **plonk info** - Package information lookup
6. **plonk diff** - Drift visualization
7. **plonk config edit** - Configuration editing

### Missing Scenario Tests:
1. **Error scenarios** - Network failures, permission issues, missing tools
2. **Concurrent operations** - Multiple plonk instances
3. **State corruption** - Invalid lock files, config files
4. **Platform differences** - macOS vs Linux behavioral differences
5. **Package manager failures** - When brew/npm/etc fail
6. **Rollback scenarios** - Partial failures during apply

## Critical User Journeys Without Tests

### 1. New User Onboarding (0% coverage)
```
plonk clone user/dotfiles
```
- No tests for git operations
- No tests for manager detection
- No tests for automatic apply

### 2. Package Management Flow (0% coverage)
```
plonk search ripgrep
plonk info brew:ripgrep
plonk install ripgrep
plonk uninstall ripgrep
```

### 3. Dotfile Drift Management (partial coverage)
```
plonk status
plonk diff .zshrc
plonk apply --dotfiles
```
- Status and apply have some tests
- Diff has no tests

### 4. Configuration Management (minimal coverage)
```
plonk config
plonk config edit
```

## Risk Assessment

### High Risk Areas (0% coverage):
1. **Clone command** - First user experience, completely untested
2. **Package installation** - Core functionality, no tests
3. **Error handling** - Most error paths untested
4. **Output/Progress** - User feedback mechanisms untested

### Medium Risk Areas (<20% coverage):
1. **Orchestrator** - Complex coordination logic
2. **Diagnostics** - System health checks
3. **Lock file operations** - Data integrity

### Lower Risk Areas (>30% coverage):
1. **Config defaults** - Basic functionality tested
2. **Resource types** - Some coverage through integration tests

## Recommendations

### Immediate (Pre-v1.0):
1. Add integration tests for clone command (mock git operations)
2. Add integration tests for install/uninstall commands
3. Add unit tests for critical business logic in commands
4. Add error scenario tests

### Short-term (v1.1):
1. Achieve 60%+ unit test coverage
2. Add comprehensive integration test suite
3. Add BATS tests for complex user workflows
4. Set up coverage gates in CI

### Long-term:
1. Target 80%+ combined coverage
2. Add mutation testing
3. Add performance benchmarks
4. Add fuzz testing for parsers

## Technical Debt Impact

The low test coverage creates significant technical debt:
- **Refactoring Risk**: Hard to change code safely
- **Bug Risk**: High probability of regressions
- **Onboarding Risk**: New contributors can't understand expected behavior
- **Maintenance Cost**: More time spent on manual testing

## Conclusion

The current test coverage of ~30-35% is insufficient for a v1.0 release. Critical user-facing features like `clone`, `install`, and `uninstall` have zero test coverage. This poses significant risks for users and maintainers.

At minimum, we should add integration tests for the primary user journeys before v1.0. The ideal target would be 60%+ coverage before release, with a plan to reach 80%+ post-release.
