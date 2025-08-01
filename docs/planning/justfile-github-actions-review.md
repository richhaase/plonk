# Justfile and GitHub Actions Review

Date: 2025-08-01

## Executive Summary

This review analyzes the build and CI/CD setup for plonk, identifying opportunities for simplification and improvement. The current setup is well-structured but has some unused features and minor inconsistencies that can be cleaned up for v1.0.

## Current State Analysis

### Justfile (350 lines)
**Strengths:**
- Well-organized with clear task grouping
- Good use of helper functions to reduce duplication
- Comprehensive development workflow support
- Clear documentation and examples

**Issues Found:**
- References non-existent `generate-mocks` recipe
- Contains unused manual release commands
- Some redundancy with CI workflows

### GitHub Actions
**Strengths:**
- Modular design with composite actions
- Cross-platform testing (Ubuntu, macOS)
- Matrix testing for multiple Go versions
- Integrated security scanning

**Issues Found:**
- Go version matrix (1.22, 1.23) doesn't match go.mod requirement (1.23.0)
- Security workflow only runs on PRs, missing main branch
- No coverage enforcement (intentional for now)

### Testing Structure
**Current:**
- Unit tests (standard Go tests)
- BATS behavioral tests (shell-based)
- Integration tests (Go with build tag)

**Issue:** Dual integration test systems (BATS + Go integration) create maintenance overhead

## Proposed Improvements

### 1. Immediate Fixes for v1.0

#### Justfile Cleanup
```diff
# In dev-setup recipe:
- @echo "  • Generating test mocks..."
- just generate-mocks

# Remove entire release section:
- release VERSION:
- release-snapshot:
- release-check:
- release-version-suggest:
- _require-clean-git:
- _check-main-branch:
```

#### Security Workflow Update
```diff
# .github/workflows/security.yml
- on: [pull_request]
+ on:
+   push:
+     branches: [main]
+   pull_request:
+     branches: [main]
```

#### Go Version Alignment
**Update**: Cannot lower to Go 1.22 due to tool dependencies requiring 1.23:
- `rogpeppe/go-internal` (used by golangci-lint)
- `honnef.co/go/tools` (static analysis tools)

Recommended approach:
- Keep `go 1.23.0` in go.mod (required by dependencies)
- Update CI matrix to test `['1.23', '1.24']` to match actual requirements
- Remove toolchain directive if not needed

### 2. Test Consolidation (Hybrid Approach)

**Move to Go integration tests:**
- Package manager operations
- File system operations
- Lock file management
- Most command testing

**Keep minimal BATS tests for:**
- Full `plonk clone` workflow
- End-to-end setup scenarios
- Shell integration verification
- Cross-command workflows

**Benefits:**
- Reduces maintenance of two systems
- Leverages existing Go testing knowledge
- Better IDE support and debugging
- Keeps critical user-journey validation

### 3. Post-v1.0 Improvements

Document these for future consideration:

#### Coverage Enforcement
- Add coverage ratcheting (never decrease)
- Or set minimum threshold (e.g., 60%)
- Use codecov.yml configuration

#### Release Automation
- Consider semantic-release for automated versioning
- Add changelog generation from conventional commits
- Automate release notes

#### Performance Testing
- Add benchmark tests for critical paths
- Monitor binary size growth
- Profile memory usage

## Implementation Priority

### Must Do Before v1.0
1. Remove `generate-mocks` reference from Justfile
2. Update security workflow to include main branch
3. Lower go.mod requirement to 1.22
4. Remove unused release commands from Justfile

### Should Do Soon (Post-v1.0)
1. Consolidate to hybrid test approach
2. Document coverage enforcement options
3. Consider build caching improvements

### Nice to Have (Future)
1. Automated dependency updates
2. Performance benchmarking
3. Release automation improvements

## Risk Assessment

All proposed changes are low risk:
- Justfile changes only remove unused code
- Security workflow change adds coverage, doesn't break anything
- Go version change matches already-tested versions
- Test consolidation can be done incrementally

## Conclusion

The build and CI/CD setup is fundamentally sound. The proposed improvements focus on:
1. Removing unused complexity
2. Aligning version requirements with reality
3. Consolidating duplicate test systems
4. Improving security coverage

These changes will make the system easier to maintain while preserving all current functionality.
