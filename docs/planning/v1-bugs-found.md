# V1.0 Bugs Found During Linux Testing

**Date**: 2025-07-30
**Testing Platform**: Ubuntu 24.10 ARM64 via Lima
**Plonk Version**: v0.8.10-dev

## Overview

Linux platform testing revealed multiple bugs that affect ALL platforms (macOS and Linux). These bugs need to be fixed before v1.0 release.

## Bug List by Priority

### 🔴 HIGH Priority - Blocks v1.0

#### 1. Apply Command Doesn't Restore Drifted Files
**Severity**: CRITICAL - Breaks core drift recovery feature
**Affects**: All platforms
**Description**:
- `plonk apply` only processes "missing" files, completely ignores "drifted" files
- Drift detection works correctly (shows "drifted" status)
- But apply doesn't restore the original content

**Test Case**:
```bash
echo "modified" >> ~/.bashrc_test
plonk status --dotfiles | grep drifted  # Shows as drifted ✓
plonk apply --dotfiles                  # Should restore but doesn't
cat ~/.bashrc_test                      # Still has modified content
```

**Expected**: Apply should restore BOTH missing AND drifted files
**Actual**: Only missing files are processed

**Code Location**: Likely in `internal/orchestrator/apply.go` or `internal/resources/dotfiles/operations.go`

---

#### 2. ✅ FIXED - Info Command Shows Wrong Management Status
**Severity**: HIGH - Confuses users about package state
**Affects**: All platforms
**Status**: Fixed in commit 5b58ed9
**Description**:
- `plonk info <package>` shows "Installed (not managed)" for packages that ARE managed
- `plonk status` correctly shows them as "managed"
- Info command was not checking the lock file when using prefix syntax

**Test Case**:
```bash
plonk install ripgrep
plonk status | grep ripgrep    # Shows "managed" ✓
plonk info brew:ripgrep        # Shows "Installed (not managed)" ✗
```

**Root Cause**: The `getInfoFromSpecificManager` function (used for prefix syntax) wasn't checking the lock file
**Fix**: Added lock file check to properly identify managed packages for both prefix and non-prefix syntax

---

### 🟡 MEDIUM Priority - Should Fix for v1.0

#### 3. ✅ FIXED - SOURCE Column Shows Incorrect Dotfile Paths
**Severity**: MEDIUM - Confusing display issue
**Affects**: All platforms
**Status**: Fixed in commit a2754ce
**Description**:
- Status shows SOURCE as `.config/bat/config`
- Actual source is `config/bat/config` (no leading dot)
- The leading dot is only added during deployment

**Test Case**:
```bash
plonk status --dotfiles
# Shows: SOURCE: .config/bat/config
ls ~/.config/plonk/config/bat/config  # File exists without dot
```

**Root Cause**: The status command was using `item.Name` which contains the deployment path
**Fix**: Updated to use the `source` field from metadata which contains the actual source path

---

#### 4. ✅ PARTIALLY FIXED - Apply Command Missing Progress Indicators
**Severity**: MEDIUM - Poor user experience
**Affects**: All platforms
**Status**: Partially fixed in commit 920d370
**Description**:
- Individual install shows progress: `[1/3] Installing: package`
- Apply shows nothing during operation
- Users don't know what's happening

**Test Case**:
```bash
plonk apply --packages
# No output during installation
# vs
plonk install pkg1 pkg2 pkg3
# Shows: [1/3] Installing: pkg1
```

**Root Cause**: Two issues found:
1. ProgressUpdate skipped output for single items (total <= 1) - FIXED
2. Package managers use CombinedOutput() which doesn't stream real-time output - NOT FIXED

**Partial Fix**: Now shows "Installing: package" for all operations, including single items
**Remaining Issue**: Still no real-time output from package managers during installation

---

#### 5. ✅ FIXED - Apply Shows Useless Error Messages
**Severity**: MEDIUM - Debugging nightmare
**Affects**: All platforms
**Status**: Fixed in commits f886c6c and f219198
**Description**:
- Errors show as "exit code 1: exit status 1"
- Actual error from package manager is not captured
- Makes troubleshooting impossible

**Test Case**:
```bash
plonk apply
# Shows: ✗ fzf: package installation failed (exit code 1): exit status 1
brew install fzf
# Shows: Error: fzf: no bottle available!
```

**Root Cause**: Package managers were not including command output in error messages
**Fix**: All package managers now include actual command output (trimmed to 500 chars) in error messages

---

### 🟢 LOW Priority - Can Wait

#### 6. ✅ FIXED - Doctor Shows macOS Homebrew Path on Linux
**Severity**: LOW - Cosmetic issue
**Affects**: Linux only
**Status**: Fixed in commit 7ccc314
**Description**:
- Doctor checks `/opt/homebrew/bin` on Linux
- Should check `/home/linuxbrew/.linuxbrew/bin`

**Root Cause**: Hardcoded macOS path was used for all platforms
**Fix**: Added getHomebrewPath() function to detect OS and architecture:
- macOS ARM64: /opt/homebrew/bin
- macOS Intel: /usr/local/bin
- Linux: /home/linuxbrew/.linuxbrew/bin

---

#### 7. ❌ NOT A BUG - Permission Errors Not Caught
**Severity**: N/A
**Affects**: N/A
**Status**: Determined not to be a bug
**Description**:
- Read-only lock file (chmod 444) doesn't prevent writes
- Operation succeeds when it should fail

**Resolution**: This is not a bug because:
- PLONK_DIR is in the user's home directory
- Users own their config files and have control over them
- Making your own config files read-only is a self-inflicted issue
- No real-world scenario where this would be a problem
- The test case is artificial and doesn't represent actual usage

---

#### 8. ❌ CRITICAL BUG FOUND - Specific Path Breaks Status
**Severity**: CRITICAL - Broken functionality
**Affects**: All platforms
**Found**: During Linux validation testing
**Description**:
- ANY file under `config/test/` directory ALWAYS shows as "missing"
- Files are correctly deployed and exist in both locations
- Apply command knows they need deployment
- Status reconciliation is specifically broken for this exact path
- Case-sensitive: `config/TEST/` works fine, only lowercase `config/test/` is broken

**Test Case**:
```bash
mkdir -p ~/.config/test
echo "content" > ~/.config/test/anyfile
plonk add ~/.config/test/anyfile
plonk status --dotfiles  # Shows as "missing" even though it exists
```

**Evidence**:
- `config/test/config` → always missing
- `config/test/otherfile` → always missing
- `config/TEST/file` → works correctly
- `config/testing/file` → works correctly
- `config/mydir/config` → works correctly

**Code Location**: Status reconciliation logic, likely special handling for "test" directories

---

## Linux-Specific Issues

### ARM64 Homebrew Bottles
**Issue**: Many packages lack ARM64 Linux bottles
**Impact**: Packages fail to install on ARM64 Linux
**Solution**: Document limitation or add `--build-from-source` support
**Example**: fzf, gh, lazygit, docker, colima

### 🔴 CRITICAL Priority - Must Fix Before v1.0

#### 9. Non-functional --force Flags Still Present
**Severity**: CRITICAL - User-facing broken functionality
**Affects**: All platforms
**Found**: During Linux validation testing
**Description**:
- --force flags are still defined in install, uninstall, add, and rm commands
- These flags were identified as non-functional but never removed
- Users see these in help text and expect them to work

**Test Case**:
```bash
plonk install -h  # Shows --force flag
plonk add -h      # Shows --force flag
```

**Code Locations**:
- internal/commands/install.go (line 46)
- internal/commands/uninstall.go (needs checking)
- internal/commands/add.go (needs checking)
- internal/commands/rm.go (needs checking)

---

## Summary

**Total Bugs Found**: 9
- 4 CRITICAL/HIGH priority (2 fixed, 2 remaining)
- 3 MEDIUM priority (3 fixed, 0 remaining - 1 partial)
- 2 LOW priority (1 fixed, 0 remaining, 1 not a bug)

**Bugs Fixed**: 6/9 (1 partial)
- ✅ Bug #1: Apply command drift restoration (HIGH)
- ✅ Bug #2: Info command management status (HIGH)
- ✅ Bug #3: SOURCE column dotfile display (MEDIUM)
- ⚠️  Bug #4: Apply command progress indicators (MEDIUM - partially fixed)
- ✅ Bug #5: Apply error messages (MEDIUM)
- ✅ Bug #6: Doctor Homebrew path on Linux (LOW)
- ❌ Bug #7: Not a bug - permission errors
- 🆕 Bug #8: CRITICAL - config/test/ path breaks status
- 🆕 Bug #9: CRITICAL - Non-functional --force flags

**Key Finding**: Most bugs affect ALL platforms, not just Linux. The Linux testing was valuable for discovering these issues.

## Recommended Fix Order

1. ✅ Fix drift restoration in apply (CRITICAL) - DONE
2. ✅ Fix info command status (Quick fix) - DONE
3. ✅ Fix SOURCE column display - DONE
4. ⚠️  Add progress to apply - PARTIALLY DONE
5. ✅ Improve error messages - DONE
6. Address remaining issues (3 LOW priority bugs)

## Testing Notes

All bugs were discovered during Linux platform testing but verified to affect macOS as well (except #6). Test cases are reproducible on any platform.
