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

#### 2. Info Command Shows Wrong Management Status
**Severity**: HIGH - Confuses users about package state
**Affects**: All platforms
**Description**:
- `plonk info <package>` shows "Installed (not managed)" for packages that ARE managed
- `plonk status` correctly shows them as "managed"
- Info command is not checking the lock file

**Test Case**:
```bash
plonk install ripgrep
plonk status | grep ripgrep    # Shows "managed" ✓
plonk info brew:ripgrep        # Shows "Installed (not managed)" ✗
```

**Expected**: Info should check lock file and show "Installed (managed)"
**Actual**: Always shows "not managed" for installed packages

**Code Location**: `internal/commands/info.go` - needs to check lock file

---

### 🟡 MEDIUM Priority - Should Fix for v1.0

#### 3. SOURCE Column Shows Incorrect Dotfile Paths
**Severity**: MEDIUM - Confusing display issue
**Affects**: All platforms
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

**Expected**: SOURCE should show actual path: `config/bat/config`
**Actual**: Shows deployment path with added dot

**Code Location**: Status display logic, possibly in `internal/commands/status.go`

---

#### 4. Apply Command Missing Progress Indicators
**Severity**: MEDIUM - Poor user experience
**Affects**: All platforms
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

**Expected**: Apply should show progress like install command
**Actual**: Silent operation until completion

**Code Location**: `internal/commands/apply.go` - needs progress indicators

---

#### 5. Apply Shows Useless Error Messages
**Severity**: MEDIUM - Debugging nightmare
**Affects**: All platforms
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

**Expected**: Show actual error from package manager
**Actual**: Generic "exit status 1" message

**Code Location**: Error handling in package operations

---

### 🟢 LOW Priority - Can Wait

#### 6. Doctor Shows macOS Homebrew Path on Linux
**Severity**: LOW - Cosmetic issue
**Affects**: Linux only
**Description**:
- Doctor checks `/opt/homebrew/bin` on Linux
- Should check `/home/linuxbrew/.linuxbrew/bin`

**Code Location**: `internal/diagnostics/health.go`

---

#### 7. Permission Errors Not Caught
**Severity**: LOW - Data integrity concern
**Affects**: All platforms
**Description**:
- Read-only lock file (chmod 444) doesn't prevent writes
- Operation succeeds when it should fail

**Test Case**:
```bash
chmod 444 ~/.config/plonk/plonk.lock
plonk install wget  # Succeeds but shouldn't
```

**Code Location**: Lock file operations

---

#### 8. Deployed Dotfiles Show as Missing
**Severity**: LOW - Display issue
**Affects**: All platforms
**Description**:
- After `plonk apply`, newly deployed files show as "missing"
- Status is correct after next reconciliation

**Code Location**: Status reconciliation logic

---

## Linux-Specific Issues

### ARM64 Homebrew Bottles
**Issue**: Many packages lack ARM64 Linux bottles
**Impact**: Packages fail to install on ARM64 Linux
**Solution**: Document limitation or add `--build-from-source` support
**Example**: fzf, gh, lazygit, docker, colima

## Summary

**Total Bugs Found**: 8
- 2 HIGH priority (blocks v1.0)
- 3 MEDIUM priority (should fix)
- 3 LOW priority (can defer)

**Key Finding**: Most bugs affect ALL platforms, not just Linux. The Linux testing was valuable for discovering these issues.

## Recommended Fix Order

1. Fix drift restoration in apply (CRITICAL)
2. Fix info command status (Quick fix)
3. Fix SOURCE column display
4. Add progress to apply
5. Improve error messages
6. Address remaining issues

## Testing Notes

All bugs were discovered during Linux platform testing but verified to affect macOS as well (except #6). Test cases are reproducible on any platform.
