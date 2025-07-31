# Linux Platform Test Results

**Date**: 2025-07-30
**Platform**: Ubuntu 24.10 (arm64) via Lima VM
**Plonk Version**: v0.8.10-dev

## Test Environment

- **VM**: Lima on macOS (Apple Silicon)
- **OS**: Ubuntu 24.10 (oracular)
- **Architecture**: aarch64/arm64
- **Homebrew**: 4.5.13 (installed to `/home/linuxbrew/.linuxbrew`)
- **Go**: 1.24.5 (installed via Homebrew)

## Test Results Summary

### ✅ Successful Tests

1. **Environment Setup**
   - Ubuntu VM creation via Lima
   - Homebrew installation (with arm64 warning but functional)
   - Go installation via Homebrew
   - Plonk build and installation from source

2. **Core Package Management**
   - Single package installation (`plonk install ripgrep`)
   - Multiple package installation (`plonk install fd bat`)
   - Package installation with brew: prefix (`plonk install brew:jq brew:htop`)
   - Package uninstallation (`plonk uninstall brew:jq`)
   - Package search (`plonk search brew:rust`)
   - Progress indicators display correctly
   - Lock file creation and updates

3. **Language Package Managers**
   - Node.js installation via Homebrew
   - npm package management (`npm:prettier`, `npm:typescript`)
   - Python installation via Homebrew
   - pip package management (`pip:black`, `pip:httpie`)
   - All packages properly tracked in status

4. **Dotfile Management**
   - Adding dotfiles (`plonk add`)
   - Directory structure preservation (`.config/test/`)
   - Dotfile deployment status tracking
   - Missing file restoration via `plonk apply`
   - Drift detection (shows "drifted" status)
   - Diff command functionality

5. **Error Handling**
   - Non-existent package shows clear error
   - Uninstalled package info shows "Available" status

### ⚠️ Issues Found

1. **Doctor Command PATH Issue**
   - Shows `/opt/homebrew/bin` (macOS path) instead of `/home/linuxbrew/.linuxbrew/bin`
   - Minor issue - doesn't affect functionality

2. **Info Command Bug**
   - Shows managed packages as "Installed (not managed)"
   - `plonk status` correctly shows them as managed
   - Info command not checking lock file properly

3. **Apply Command Bug**
   - Only restores "missing" files, ignores "drifted" files
   - Drift detection works but restoration doesn't
   - Critical bug for drift recovery feature

4. **Permission Error Not Shown**
   - Read-only lock file (chmod 444) didn't prevent package installation
   - Expected permission error but operation succeeded
   - Potential data integrity issue

5. **Terminal Warning**
   - `plonk diff` shows "WARNING: terminal is not fully functional"
   - Still works but requires pressing RETURN

## Performance Observations

- Package operations are fast and responsive
- Progress indicators helpful for multi-package operations
- Search operations complete quickly
- No noticeable performance difference from macOS

## Linux-Specific Findings

1. **Homebrew Location**: `/home/linuxbrew/.linuxbrew` (not `/opt/homebrew`)
2. **PATH Setup**: Required manual configuration in `.bashrc`
3. **Architecture**: arm64 warning from Homebrew but everything works
4. **Home Directory**: Lima uses `/home/username.linux` pattern

## Bugs to Fix

### High Priority
1. **Apply doesn't restore drifted files** - Only handles missing files
2. **Info command shows wrong management status** - Not checking lock file

### Medium Priority
3. **Doctor shows wrong Homebrew path on Linux** - Cosmetic but confusing
4. **Permission errors not caught** - Lock file modifications succeed when they shouldn't

### Low Priority
5. **Terminal warning in diff command** - Works but shows warning

## Recommendations

1. **Fix Apply Command**: Should restore both missing AND drifted files
2. **Fix Info Command**: Should check lock file for management status
3. **Update Doctor**: Add Linux-specific Homebrew path (`/home/linuxbrew/.linuxbrew`)
4. **Add Permission Checks**: Ensure lock file writes fail appropriately
5. **Documentation Updates**:
   - Add Linux-specific PATH setup instructions
   - Note Homebrew installation location difference
   - Add troubleshooting for terminal warnings

## Next Steps

1. Test clone workflow on fresh VM (pending)
2. Test concurrent operations
3. Test WSL2 compatibility
4. Fix identified bugs
5. Update documentation for Linux users

## Overall Assessment

Plonk works well on Linux with Homebrew as the package manager. The core functionality is solid with only a few bugs that need fixing. The main issues are:
- Drift restoration not working
- Info command status bug
- Minor path and permission issues

Once these are fixed, plonk will have full feature parity between macOS and Linux.
