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
1. ✅ A user can set up a new Mac/Linux machine with minimal effort (install git + homebrew/apt, then `plonk clone`)
2. ✅ Core commands work reliably without surprises
3. ✅ New users can start using plonk immediately after installation
4. ✅ The tool works equally well on macOS and Linux (including WSL)

## Required Features for v1.0

### 🔴 Critical - Must Have

#### 1. Dotfile Drift Detection
**Status**: Not implemented
**Priority**: TOP - User identified as "critical gap"
- Add "out-of-sync" state to reconciliation system
- Show in `plonk status` when deployed dotfiles differ from source
- Provide clear indication of what will change on `plonk apply`
- Future: Same concept for package updates

#### 2. APT Package Manager Support
**Status**: Not implemented
**Priority**: HIGH - Required for Linux support
- Add apt manager to package manager registry
- Handle sudo requirements gracefully
- Ensure uniform experience across brew/apt
- Test on Ubuntu, Debian at minimum

#### 3. Progress Indicators
**Status**: ✅ Completed (2025-07-30)
**Priority**: HIGH - Critical for user feedback
- Periodic status output: "Installing package 2 of 5..."
- Most urgent for: install, apply, search operations
- Simple implementation sufficient for v1

#### 4. Linux Platform Testing & Support
**Status**: Partial - needs verification
**Priority**: HIGH - Core platform requirement
- Test on major distributions (Ubuntu, Fedora, Arch)
- Ensure all commands work identically to macOS
- Handle Linux-specific paths and conventions
- Document any Linux-specific setup requirements

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
**Status**: Partially complete
**Priority**: MEDIUM - User experience
- Remove all references to defunct `setup` command
- Update installation guide with current release info
- Review and update README stability warning
- Ensure examples use current commands
- Document `.plonk/` directory as reserved

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

- [ ] Implement dotfile drift detection
- [ ] Implement APT package manager support
- [x] Implement progress indicators
- [ ] Complete Linux platform testing
- [x] Add `.plonk/` directory exclusion
- [⏸️] Consolidate doctor/setup shared code (skipped - needs design decisions)
- [ ] Update all documentation
- [ ] Remove stability warning from README
- [ ] Document `.plonk/` as reserved directory
- [ ] Test full user journey on fresh macOS
- [ ] Test full user journey on fresh Ubuntu
- [ ] Test full user journey on WSL
- [ ] Review all error messages for clarity
- [ ] Ensure `plonk clone` handles edge cases:
  - [ ] No lock file (dotfiles only)
  - [ ] Empty repository
  - [ ] Existing destination directory

## Post-v1.0 Roadmap

High-value features to implement after v1.0:
1. Package update command
2. Verbose/debug modes
3. Additional Linux package managers (yum, pacman, dnf)
4. Doctor --fix for more issues
5. Better error messages with remediation hints
6. Performance optimizations
7. Hook system (using `.plonk/hooks/`)

## Risk Assessment

### High Risk Items
1. **APT implementation** - Different security model than brew
   - Mitigation: Careful design of sudo handling

2. **Drift detection** - Performance on large dotfile sets
   - Mitigation: Start simple, optimize later

### Low Risk Items
1. Progress indicators - Straightforward to implement
2. Documentation updates - No code changes
3. Linux testing - Mostly validation work

## Definition of Done for v1.0

- [ ] All critical features implemented and tested
- [ ] No known data loss bugs
- [ ] Commands work identically on macOS and Linux
- [ ] Can set up new machine with published documentation
- [ ] Core user journeys work without reading source code
- [ ] Version changed from 0.x to 1.0.0
- [ ] Release notes explain stability commitments
- [ ] Tagged and released with pre-built binaries

## Timeline Estimate

Based on the scope above:
- Dotfile drift detection: 2-3 days
- APT support: 3-5 days (including testing)
- Progress indicators: 1-2 days
- Linux testing: 2-3 days
- `.plonk/` exclusion: 0.5 days
- Documentation & cleanup: 1-2 days

**Total estimate**: 2-3 weeks of focused development

---

*Note: This document represents the minimum viable v1.0. Many valuable features from ideas.md are intentionally deferred to maintain focus and ship a solid foundation.*
