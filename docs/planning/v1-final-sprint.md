# V1.0 Final Sprint Action Plan

## Current Status (2025-07-30)

**Completed**: 5 out of 7 v1.0 features
**Remaining**: 2 tasks (3-5 days)
**Target**: v1.0.0 release

## Remaining Tasks

### 1. Linux Platform Testing (2-3 days)

**Purpose**: Validate plonk works identically on Linux as on macOS

**Test Environments**:
- Ubuntu 22.04 LTS (via Lima VM)
- Debian 12 (via Lima VM)
- WSL2 on Windows (Ubuntu)

**Test Plan**: See [linux-homebrew-testing-plan.md](linux-homebrew-testing-plan.md)

**Key Validation Points**:
1. Homebrew installation via `plonk doctor --fix`
2. All package managers work (brew, npm, cargo, pip, gem, go)
3. Dotfile management works identically
4. `plonk clone` full journey succeeds
5. Output formatting consistent with macOS

**Deliverables**:
- Test execution logs
- Bug fixes (if any)
- Linux-specific documentation updates

### 2. Documentation Updates & Release Prep (1-2 days)

**Critical Updates**:
- [x] Remove stability warning from README
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

## Stability Commitment
Core commands and behaviors are now stable. Future changes will maintain backwards compatibility.

## Installation
```bash
go install github.com/richhaase/plonk/cmd/plonk@v1.0.0
```

## Getting Started
See the [quickstart guide](https://github.com/richhaase/plonk#quick-start)
```

## Daily Plan

### Day 1: Linux Testing Setup & Basic Validation
- Set up Lima VMs for Ubuntu and Debian
- Run basic command suite
- Document any setup issues
- Fix critical bugs

### Day 2: Full Linux Testing
- Complete test scenarios from testing plan
- Test edge cases
- Verify WSL2 compatibility
- Document Linux-specific behaviors

### Day 3: Bug Fixes & Polish
- Fix any issues found during testing
- Update documentation based on findings
- Prepare release notes

### Day 4: Documentation & Release
- Remove stability warning
- Update version numbers
- Create comprehensive release notes
- Tag v1.0.0
- Create GitHub release with binaries

### Day 5: Buffer/Validation
- Final validation on fresh systems
- Address any last-minute issues
- Announce release

## Success Criteria

v1.0 ships when:
- [x] All critical features work reliably
- [ ] Linux testing shows parity with macOS
- [ ] Documentation reflects current behavior
- [ ] Version updated to 1.0.0
- [ ] Release notes explain stability
- [ ] Tagged and released on GitHub

## Not Blocking v1.0

These can wait:
- Additional package manager support
- Performance optimizations
- Verbose/debug modes
- Native Windows support
- Hook system implementation

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
