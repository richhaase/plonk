# Plonk v1.0 Work Summary

## Executive Summary

To reach v1.0, plonk needs 5 critical features and 3 important improvements. The total effort is estimated at **2-3 weeks** of focused development.

## Priority Overview

### 🔴 Critical (Must Have)
1. **Dotfile Drift Detection** - Show when deployed files differ from source
2. **APT Package Manager** - Linux support via apt
3. **Progress Indicators** - User feedback during operations
4. **Linux Platform Testing** - Ensure cross-platform compatibility

### 🟡 Important (Should Have)
5. **`.plonk/` Directory Exclusion** - Future-proof metadata storage
6. **Doctor Code Consolidation** - Reduce code duplication
7. **Documentation Updates** - Remove outdated references

## Proposed Execution Order

### Phase 1: Foundation (Week 1)
1. **`.plonk/` Directory Exclusion** (0.5 days)
   - Simplest change, enables future features
   - No breaking changes

2. **Progress Indicators** (1-2 days)
   - Improves UX immediately
   - Helps debug subsequent work

3. **Doctor Code Consolidation** (1-2 days)
   - Reduces technical debt
   - Makes APT implementation cleaner

### Phase 2: Core Features (Week 2)
4. **APT Package Manager Support** (3-5 days)
   - Most complex feature (sudo handling)
   - Enables Linux support

5. **Dotfile Drift Detection** (2-3 days)
   - Top user priority
   - Requires new reconciliation state

### Phase 3: Polish & Release (Week 3)
6. **Linux Platform Testing** (2-3 days)
   - Test on Ubuntu, Fedora, Arch
   - Fix platform-specific issues

7. **Documentation Updates** (1-2 days)
   - Update all references
   - Document new features
   - Remove stability warnings

## Level of Effort Summary

| Task | Estimated Days |
|------|----------------|
| `.plonk/` Directory Exclusion | 0.5 |
| Progress Indicators | 1.5 |
| Doctor Code Consolidation | 1.5 |
| APT Package Manager Support | 4.0 |
| Dotfile Drift Detection | 2.5 |
| Linux Platform Testing | 2.5 |
| Documentation Updates | 1.5 |
| **TOTAL** | **14 days** |

**Total Timeline: 2-3 weeks** (accounting for context switching, review, and testing)

## Why This Order?

1. **Quick Wins First**: `.plonk/` exclusion and progress indicators provide immediate value with low risk

2. **Technical Debt Before Features**: Consolidating doctor code makes APT implementation cleaner

3. **Complex Features Mid-Sprint**: APT and drift detection when focus is highest

4. **Testing & Polish Last**: Ensures all features are complete before final validation

## Success Metrics

v1.0 ships when:
- ✅ User can set up Linux/Mac with one command
- ✅ All features work identically across platforms
- ✅ Core commands have no breaking changes
- ✅ New users can start immediately

## Not Included (Post-v1.0)

- Package update command
- Verbose/debug modes
- Additional package managers (yum, pacman)
- Hook system implementation
- Performance optimizations

---

*This focused scope ensures v1.0 delivers on plonk's core promise: one-command setup that just works.*
