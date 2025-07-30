# Outstanding Improvements and Ideas

This document catalogs all improvements, enhancements, and ideas mentioned throughout the plonk documentation. Items are organized by source and priority.

## Verification Summary (2025-07-30)

Based on review of CLAUDE.md, recent commits, and code verification:

### ✅ COMPLETED Items (verified):
1. **Lock file v2 with metadata** - Completed 2025-07-29
2. **Status alphabetical sorting** - Completed 2025-07-30
3. **Status flag combinations review** - Completed 2025-07-30
4. **Status --missing flag** - Completed 2025-07-29
5. **Doctor PATH copy-paste commands** - Completed 2025-07-29

### ❌ NOT COMPLETED (despite user thinking they were):
1. **Doctor --fix for all issues** - Still limited to package managers only
2. **Doctor PATH auto-fix** - Only copy-paste commands, no auto-fix

### 🔴 Critical Issue Found:
- **Setup command no longer exists** - Was removed and replaced with init/clone
- Documentation still references setup command in multiple places

## From Command Documentation (docs/cmds/)

### Dotfile Management (dotfile_management.md)

**Improvements Section:**
- Improve path resolution documentation in help text
  - Needs to review this logic and ensure that it is valid.
- Add verbose output option to show ignore pattern matches
  - Yes.  This is needed.
- Consider warning when re-adding files that differ from current version
  - This should be a difference in reporting success.  It's not a warning, but it would be helpful to output a listing of what was added v. re-added especially in the case of re-adding a directory or subdirectory.
- Add drift detection system to identify when deployed dotfiles differ from source
  - Yes.  This is a critical gap in dofile management today. This is the top priority from this section.

### Package Management (package_management.md)

**Improvements Section:**
- **Enhance lock file format**: Store both binary name and full source path for Go packages and npm scoped packages
  - Current limitation: Go packages lose source path information (e.g., `golang.org/x/tools/cmd/gopls` → `gopls`)
  - Proposed v2 lock format enhancement using metadata field
    - I thought we already implemented this?  We need to review this an ensure it is setup.
- Add verbose search mode showing descriptions and versions
  - I think this should actually be the default search behavior.
- Support version pinning in install command
  - This is an intersting idea, we should plan it, but it's not high priority.
- Add update command to upgrade managed packages
  - Yes.  This is definitely a gap in the packge manager manager`
- Show installation progress for long-running operations
  - Yes.  We need to report status more quickly especially when adding multiple packages.
- Add --all flag to uninstall all packages from a manager
  - This is worth considering.
- Consider showing dependencies in info output
  - This is interesting, but it should be handle by the individual package managers, maybe we can surface what the package managers know about dependencies?  This could be complicated.

**Implementation Notes Section:**
- Future Enhancement: v2 lock format could store both binary name and source path in metadata
  - I was pretty sure we did made this change yesterday, if we didnt't this is top priority for this section.

### Apply Command (apply.md)

**Improvements Section:**
- Consider adding progress indicators for large apply operations
  - Yes.
- Add veerbose mode for detailed operation logging
  - Yes.
- Add support for selective dotfile deployment based on patterns
  - This is interesting, but it seems like pretty advanced usecase, we should consider this low priority.

### Status Command (status.md)

**Improvements Section:**
- Sort items alphabetically instead of by package manager
  - I'm pretty sure we already did this.  the plonk output looks sorted alphabetically.
- Review flag combination behavior (e.g., --packages --dotfiles redundancy)
  - We already did this.
- Consider built-in pagination for very long lists
  - Interesesting by low priority.
- Add --missing flag to show only emissing resources
  - Pretty sure we already did this.

### Doctor Command (doctor.md)

**Improvements Section:**
- Extend `--fix` to address all fixable issues, not just package managers
  - I'm pretty sure we already did this
- Review and standardize all health check behaviors
  - Yes.
- Revisit check categories for better organization
  - yes.
- Extend package manager checks to include actual functionality testing beyond availability
  - consider this, but at medium to low priority.  if the binaries exist there is a very good chance they are "healthy", extra checks would be nice, but shouldn't be necessary for our initial target users.
- Consider having setup directly call doctor instead of duplicating code
  - Yes.  I don't like having two paths for plonk to perform the same operation.  This shouldn't shell a new plonk to run doctor, but SHOULD use the same internal logic
- Add auto-fix capabilities for PATH configuration issues
  - I thought we added this already?
- Provide copy-paste ready PATH export commands for user's specific shell
  - I thought we added this already?

## From Main Documentation

### Project Status (README.md)
- Currently marked with warning: "This project is under active development. APIs, commands, and configuration formats may change without notice."  - We are currently stablizing core features and interface.  Add a note to consider what we need to accomplish before we can "freeze" the core functionality as largely stable.
- No specific improvements mentioned, but indicates project is not yet stable

### Installation Guide (installation.md)
- States "Pre-built binaries will be available for major platforms in future releases"
  - Note: This appears outdated as releases already exist (v0.8.9)
    - We should update this document to reflect the actual install pattern right now, but this requires discussion.

## Categorization by Priority

### High Priority (Core Functionality & Data Integrity)

1. **Enhanced Lock File Format**
   - Problem: Go packages lose source information
   - Impact: Cannot reinstall packages correctly
   - Solution: v2 lock format with metadata field

2. **Version Pinning Support**
   - Allow specifying exact versions during install
   - Critical for reproducible environments

3. **Package Update Command**
   - Currently no way to upgrade managed packages
   - Essential for maintenance

### Medium Priority (User Experience)

1. **Progress Indicators**
   - Long operations provide no feedback
   - Affects: apply, install, search commands

2. **Alphabetical Sorting in Status**
   - Current grouping by package manager is less intuitive
   - Simple fix with high UX impact

3. **Verbose Modes**
   - Debugging is difficult without detailed output
   - Affects: apply, search, dotfile operations

4. **Doctor PATH Auto-Fix**
   - Currently only suggests manual fixes
   - Could generate shell-specific commands

5. **Missing Resources Flag**
   - `plonk status --missing` to show only what needs attention
   - Useful for CI/CD scenarios

### Low Priority (Nice to Have)

1. **Dotfile Drift Detection**
   - Identify when deployed files differ from source
   - Helpful for debugging unexpected behavior

2. **Dependency Information**
   - Show what else would be affected by package changes
   - Complex to implement across package managers

3. **Built-in Pagination**
   - For users with many managed resources
   - Terminal pagers work as workaround

4. **Selective Dotfile Deployment**
   - Pattern-based apply for dotfiles
   - Edge case for most users

## Technical Debt & Inconsistencies

1. **Doctor/Setup Code Duplication**
   - Setup reimplements doctor functionality <- RED FLAG, we removed setup yesterday, only init and clone exist now.
   - Should delegate to shared code <- Yes

2. **Flag Combination Behavior**
   - Some commands have redundant flag combinations  <- Requires review I think we took care of this yesterday
   - Needs consistency review

3. **Health Check Standardization**
   - Doctor checks could be better organized
   - Some checks only test availability, not functionality

## Documentation Updates Needed

1. **Installation Guide**
   - Update to reflect existing releases
   - Add download instructions for pre-built binaries

2. **Project Stability Warning**
   - Consider if still needed with v0.8.9
   - May discourage adoption unnecessarily

## Revised Priority List Based on User Feedback

### Critical Priority - Must Have
1. **Dotfile Drift Detection** (NEW TOP PRIORITY)
   - User: "This is a critical gap... top priority"
   - Compare deployed vs source files
   - Essential for debugging

2. **Package Update Command**
   - User: "This is definitely a gap in the package manager manager"
   - No current way to upgrade packages
   - Essential for maintenance

3. **Progress Indicators**
   - User: "We need to report status more quickly"
   - Multiple "Yes" responses
   - Critical for user feedback during long operations

### High Priority - Core Features
1. **Verbose Modes**
   - User: Multiple "Yes" responses
   - Default search should show descriptions/versions
   - Essential for debugging

2. **Doctor Code Consolidation**
   - User: "I don't like having two paths for plonk to perform the same operation"
   - Clone/init should use shared doctor logic internally
   - Reduces maintenance burden

3. **Standardize Doctor Health Checks**
   - User: "Yes" to both review and reorganization
   - Better categories and consistent behavior

### Medium Priority - Nice to Have
1. **Add/Re-add Reporting**
   - User: "helpful to output a listing of what was added v. re-added"
   - Better success reporting for directory operations

2. **Version Pinning**
   - User: "interesting idea... not high priority"
   - For reproducible environments

3. **Uninstall --all Flag**
   - User: "This is worth considering"
   - Convenience feature

4. **Doctor Functionality Testing**
   - User: "medium to low priority... binaries exist = probably healthy"
   - Beyond basic availability checks

### Low Priority - Future Considerations
1. **Selective Dotfile Deployment**
   - User: "pretty advanced usecase... low priority"
   - Pattern-based apply

2. **Built-in Pagination**
   - User: "Interesting but low priority"
   - Terminal pagers work as alternative

3. **Dependency Information**
   - User: "This could be complicated"
   - Surface package manager dependency info

## Items Requiring Clarification

1. **Doctor --fix Status**
   - User thought this was complete for all issues
   - Reality: Only fixes package managers
   - Need to verify user expectations

2. **PATH Auto-Fix**
   - User thought this was complete
   - Reality: Only provides copy-paste commands
   - May need to clarify distinction

3. **Setup Command References**
   - Setup was removed, replaced with init/clone
   - Documentation needs systematic update

## General Platform Improvements

1. **Linux Support**
   - Ensure full compatibility across Linux distributions
   - Test on major distros (Ubuntu, Fedora, Arch, etc.)
   - Handle Linux-specific package managers and paths
   - Address any Linux-specific issues or limitations

## Next Steps

1. **Immediate Actions**
   - Implement dotfile drift detection (critical gap)
   - Add package update command
   - Add progress indicators

2. **Documentation Cleanup**
   - Remove all setup command references
   - Update installation guide for current releases
   - Consider removing/updating stability warning

3. **Technical Debt**
   - Consolidate doctor logic for reuse
   - Complete flag combination review
   - Standardize health check implementation
