# Plonk v1.0 Readiness Checklist

This document defines the requirements for plonk v1.0 - a stable release with core features ready for users and a solid foundation for future development.

## Vision Alignment

Per [why-plonk.md](../why-plonk.md), v1.0 must deliver on the core promise:
- **One command setup**: `plonk clone user/dotfiles` sets up an entire development environment
- **Zero configuration**: Works out of the box with sensible defaults
- **Unified management**: Packages and dotfiles together
- **Cross-platform**: Same experience on macOS and Linux

## Success Criteria

v1.0 is ready when:
1. ✅ A user can set up a new Mac/Linux machine with minimal effort (install homebrew, then plonk, then `plonk clone`)
2. ✅ Core commands work reliably without surprises
3. ✅ New users can start using plonk immediately after installation (zero-config)
4. ✅ The tool works equally well on macOS and Linux (including WSL)

## Required Features for v1.0

### 🔴 Critical - Must Have

#### 1. Dotfile Drift Detection
**Status**: ✅ Completed (2025-07-30)
**Priority**: TOP - User identified as "critical gap"
- Added "drifted" state to reconciliation system
- Shows in `plonk status` when deployed dotfiles differ from source
- Implemented `plonk diff` command to show changes
- SHA256 checksum-based comparison for performance

#### 2. Linux Support via Homebrew
**Status**: ✅ Completed (2025-07-30)
**Priority**: HIGH - Required for Linux support
- Removed APT in favor of Homebrew on Linux
- Homebrew provides consistent cross-platform experience
- No sudo required - true user-space package management
- Same dotfiles work on macOS and Linux

#### 3. Progress Indicators
**Status**: ✅ Completed (2025-07-30)
**Priority**: HIGH - Critical for user feedback
- Spinner-based progress for operations > 100ms
- Shows progress for: install, apply, search operations
- Clean implementation using briandowns/spinner library

#### 4. Linux Platform Testing & Support
**Status**: ✅ Completed (2025-07-30)
**Priority**: HIGH - Core platform requirement
- Tested on Ubuntu 24.10 ARM64 via Lima
- All commands work identically to macOS
- Homebrew on Linux verified and documented
- Linux-specific bugs found and fixed
- ARM64 bottle limitations documented

### 🟡 Important - Should Have

#### 5. Add `.plonk/` Directory Exclusion
**Status**: ✅ Completed (2025-07-30)
**Priority**: MEDIUM - Future flexibility
- Exclude `.plonk/` directory from dotfile deployment
- Reserve for future plonk metadata (hooks, templates, etc.)
- Simple change: add to existing exclusion logic
- Document as "reserved for future use"
- Only create directory when actually needed

#### 6. Doctor Code Consolidation
**Status**: ⏸️ Skipped - Needs design decisions (2025-07-30)
**Priority**: MEDIUM - Technical debt
- Extract shared logic from clone/init and doctor
- Use same internal functions (not shelling out)
- Reduces maintenance and ensures consistency
- See: [doctor-consolidation-plan.md](doctor-consolidation-plan.md) for details

#### 7. Documentation Updates
**Status**: In Progress
**Priority**: MEDIUM - User experience
- [x] Remove all references to defunct `setup` command
- [ ] Update installation guide with current release info
- [ ] Review and update README stability warning
- [x] Ensure examples use current commands
- [x] Document `.plonk/` directory as reserved

### 🟢 Nice to Have - Can Wait

#### 8. Package Update Command
**Status**: Not implemented
**Priority**: LOW - Has workaround
- Users can uninstall/install for now
- Important feature but not blocking v1

#### 9. Verbose/Debug Modes
**Status**: Needs discussion
**Priority**: LOW - Needs design work
- Balance between information and UI cleanliness
- May implement post-v1 based on user feedback

## Stability Commitments for v1.0

### Will Not Change (Stable APIs)
1. **Core command interfaces**: install, uninstall, add, rm, status, apply, etc.
2. **Command flags**: Existing flags remain compatible
3. **Basic behavior**: Commands continue to work as documented

### Can Evolve (With Compatibility)
1. **Configuration format**: Additive changes only (new fields OK)
2. **Lock file format**: Must auto-upgrade transparently
3. **Output formats**: Can be enhanced but not broken
4. **New commands**: Can be added without breaking existing ones

## Known Limitations for v1.0

Acceptable limitations that don't block v1:
1. No native Windows support (WSL only)
2. No retry on network failures
3. Doctor --fix limited to package managers
4. No built-in verbose mode (yet)
5. Basic error messages (enhancement later)

## Pre-Release Checklist

Before tagging v1.0.0:

- [x] Implement dotfile drift detection
- [x] Implement Linux support (via Homebrew)
- [x] Implement progress indicators
- [x] Complete Linux platform testing
- [x] Add `.plonk/` directory exclusion
- [⏸️] Consolidate doctor/setup shared code (skipped - needs design decisions)
- [ ] Update all documentation
- [ ] Remove stability warning from README
- [x] Document `.plonk/` as reserved directory
- [x] Test full user journey on fresh macOS
- [x] Test full user journey on fresh Ubuntu
- [ ] Test full user journey on WSL
- [x] Review all error messages for clarity
- [x] Ensure `plonk clone` handles edge cases:
  - [x] No lock file (dotfiles only) - works correctly
  - [x] Empty repository - works correctly
  - [x] Existing destination directory - shows clear error

## Post-v1.0 Roadmap

High-value features to implement after v1.0:
1. Package update command
2. Verbose/debug modes
3. Doctor --fix for more issues
4. Better error messages with remediation hints
5. Performance optimizations
6. Hook system (using `.plonk/hooks/`)
7. Native Windows support (beyond WSL)

## Risk Assessment

### Completed Items (No Longer Risks)
1. **Drift detection** - Implemented with SHA256 checksums, performs well
2. **Progress indicators** - Completed using spinner library
3. **Linux support** - Simplified by using Homebrew only

### Remaining Low Risk Items
1. **Linux testing** - Mostly validation work
2. **Documentation updates** - Final cleanup only
3. **Edge case handling** - May discover issues during testing

## Definition of Done for v1.0

- [x] All critical features implemented and tested
- [x] No known data loss bugs
- [x] Commands work identically on macOS and Linux
- [x] Can set up new machine with published documentation
- [x] Core user journeys work without reading source code
- [ ] Version changed from 0.x to 1.0.0
- [ ] Release notes explain stability commitments
- [ ] Tagged and released with pre-built binaries

## Remaining Work

Based on completed work and remaining tasks:
- ✅ Dotfile drift detection: COMPLETE
- ✅ Linux support (Homebrew): COMPLETE
- ✅ Progress indicators: COMPLETE
- ✅ `.plonk/` exclusion: COMPLETE
- ✅ Linux testing: COMPLETE
- ✅ All critical bugs: FIXED
- [ ] Test quality review: Unit test isolation, coverage analysis
- [ ] Code complexity review: Identify and reduce complexity
- [ ] Documentation review: Critical analysis and cleanup
- [ ] Build system review: Justfile and GitHub Actions
- [ ] Version update & release: After quality assurance

**Remaining estimate**: 4-6 days including quality assurance phase

---

*Note: This document represents the minimum viable v1.0. Many valuable features from ideas.md are intentionally deferred to maintain focus and ship a solid foundation.*
