# V1.0 Final Sprint Action Plan

## Current Status (2025-07-31)

**Completed**: All critical features and bug fixes
**Remaining**: Quality assurance and release preparation tasks
**Target**: v1.0.0 release

## Setup Philosophy Change

Plonk now assumes Homebrew is installed as a prerequisite. This simplifies our architecture and aligns with developer expectations:

**New Setup Flow**:
1. Install Homebrew (one-time, per machine)
2. Install plonk via `brew install plonk` (once Homebrew formula is available)
3. Either `plonk clone user/dotfiles` or just start using plonk

This removes complexity while maintaining the "zero-config" philosophy.

## Remaining Tasks

### 1. Comprehensive Test Review (1-2 days)

**Purpose**: Ensure test quality and coverage before v1.0

**Scope**:
- **Unit Tests**: Verify NO external calls, proper mocking
- **Integration Tests**: Review coverage of system interactions
- **BATS Tests**: Validate behavioral test completeness

**Key Questions**:
- Are unit tests truly isolated?
- What's our test coverage percentage?
- Are there critical paths without tests?
- Can we improve test organization or clarity?

**Deliverables**:
- Test coverage report
- List of tests that need fixing
- Recommendations for improvement

### 2. Code Complexity Review (1-2 days)

**Purpose**: Identify and reduce unnecessary complexity introduced during revisions

**Focus Areas**:
- Functions that have grown too large
- Duplicated logic that could be extracted
- Complex conditionals that could be simplified
- Error handling patterns that could be standardized

**Deliverables**:
- List of complexity hotspots
- Refactoring recommendations
- Any critical refactoring for v1.0

### 3. Critical Documentation Review (1 day)

**Purpose**: Ensure documentation is accurate, complete, and user-friendly

**Scope**:
- README.md accuracy and completeness
- Command documentation review
- Installation guide clarity
- Examples and tutorials
- API documentation (for library users)

**Key Questions**:
- What's missing that users need?
- What's outdated or incorrect?
- What's redundant and can be removed?
- Are examples clear and working?

**Deliverables**:
- Documentation gap analysis
- List of required updates
- Cleanup recommendations

### 4. Justfile and GitHub Actions Review (1 day)

**Purpose**: Ensure build system and CI/CD are robust for v1.0

**Justfile Review**:
- Are all recipes still relevant and working?
- Is the version management correct?
- Are there missing helpful recipes?
- Is the release process well-defined?

**GitHub Actions Review**:
- Is CI comprehensive enough?
- Are we testing on all target platforms?
- Is the release workflow ready for v1.0?
- Are there security concerns?

**Deliverables**:
- List of improvements needed
- Any critical fixes for v1.0
- Documentation of release process

### 5. Linux Platform Testing (COMPLETED)

**Test Environments**:
- Ubuntu 22.04 LTS (via Lima VM)
- Debian 12 (via Lima VM)
- WSL2 on Windows (Ubuntu)

**Test Plan**: See [linux-homebrew-testing-plan.md](linux-homebrew-testing-plan.md)

**Key Validation Points**:
1. Document Homebrew installation process on Linux
2. All package managers work (brew, npm, cargo, pip, gem, go)
3. Dotfile management works identically
4. `plonk clone` full journey succeeds
5. Output formatting consistent with macOS

**Deliverables**:
- Test execution logs
- Bug fixes (if any)
- Linux-specific documentation updates

### 6. Documentation Updates & Release Prep (1-2 days)

**Critical Updates**:
- [ ] Remove stability warning from README
- [ ] Update version to 1.0.0 in main.go
- [ ] Create v1.0.0 release notes
- [ ] Review all command documentation for accuracy
- [ ] Ensure installation guide is current

**Version Update Locations**:
- `cmd/plonk/main.go` - change `version = "dev"` to `version = "v1.0.0"`
- Tag commit with `v1.0.0`
- GitHub release with binaries

**Release Notes Template**:
```markdown
# Plonk v1.0.0 - Stable Release

The unified package and dotfile manager is ready for production use!

## Highlights
- 🎯 One command setup: `plonk clone user/dotfiles`
- 📦 Unified management of packages and dotfiles
- 🔄 Drift detection shows when configs have changed
- 🌍 Cross-platform support (macOS, Linux, WSL)
- 🚀 Progress indicators for long operations

## Prerequisites
- Homebrew (install from https://brew.sh)
- Git

## Installation
```bash
# Via Homebrew (coming soon)
brew install plonk

# Via Go
go install github.com/richhaase/plonk/cmd/plonk@v1.0.0
```

## Quick Start
```bash
# Clone existing dotfiles
plonk clone user/dotfiles

# Or just start using plonk
plonk add ~/.zshrc
plonk install ripgrep
```

## Stability Commitment
Core commands and behaviors are now stable. Future changes will maintain backwards compatibility.
```

## Updated Sprint Plan

### Phase 1: Quality Assurance (3-4 days)
**Day 1: Test Review**
- Analyze unit test coverage and isolation
- Review integration test completeness
- Validate BATS behavioral tests
- Create improvement plan

**Day 2: Code Complexity**
- Identify complexity hotspots
- Review error handling patterns
- Find duplicated logic
- Plan critical refactoring

**Day 3: Documentation & Build Review**
- Critical documentation review
- Justfile analysis
- GitHub Actions review
- Create documentation plan

**Day 4: Implement Critical Fixes**
- Fix any critical issues found
- Apply essential refactoring
- Update critical documentation

### Phase 2: Release Preparation (1-2 days)
**Day 5: Final Documentation**
- Remove stability warning
- Update all documentation
- Create release notes
- Final review

**Day 6: Release**
- Update version numbers
- Tag v1.0.0
- Create GitHub release
- Announce release

## Success Criteria

v1.0 ships when:
- [x] All critical features work reliably
- [x] Linux testing shows parity with macOS (with bugs documented)
- [x] Critical bugs fixed (ALL COMPLETED 2025-07-31)
- [ ] Documentation reflects current behavior
- [ ] Version updated to 1.0.0
- [ ] Release notes explain stability
- [ ] Tagged and released on GitHub

## All Critical Bugs Fixed (2025-07-31)

All critical bugs have been resolved:
1. ✅ Apply now restores drifted files
2. ✅ Info shows correct management status
3. ✅ SOURCE column displays correctly
4. ✅ Apply shows progress indicators
5. ✅ Apply error messages improved
6. ✅ Doctor Homebrew path fixed on Linux
7. ✅ Permission errors handled gracefully
8. ✅ config/test/ path reconciliation fixed
9. ✅ Non-functional --force flags removed
10. ✅ Error messages show actual package manager output
11. ✅ Usage no longer displayed on errors

See [v1-bugs-found.md](v1-bugs-found.md) for details.

## Implications of Homebrew Prerequisite

This simplification has several benefits:
1. **Removes complexity** - No need for plonk to manage Homebrew installation
2. **Cleaner error messages** - "Homebrew not found" vs complex installation failures
3. **Better user experience** - Homebrew's installer handles PATH setup correctly
4. **Simplifies doctor --fix** - May only need to install language package managers

### Completed for v1.0:
- ✅ Removed `doctor --fix` entirely - doctor is now a pure diagnostic tool
- ✅ Only `plonk clone` installs package managers (when needed for the repository)
- ✅ Updated all documentation to reflect Homebrew as prerequisite

## Not Blocking v1.0

These can wait:
- Additional package manager support
- Performance optimizations
- Verbose/debug modes
- Native Windows support
- Hook system implementation
- Homebrew formula creation (nice to have for v1.0 but not blocking)

## Risk Mitigation

**Linux Testing Risks**:
- Homebrew on Linux may have quirks → Document workarounds
- Path setup differences → Provide clear guidance
- Package availability → Note any limitations

**Documentation Risks**:
- Outdated examples → Review systematically
- Missing edge cases → Add as discovered

## Post-Release

After v1.0:
1. Monitor issue reports
2. Plan v1.1 based on feedback
3. Consider package update command as first new feature
