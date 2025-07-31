# Critical Fixes Plan for plonk v1.0

## Overview
Linux validation testing on 2025-07-31 revealed 3 critical issues that must be fixed before v1.0 release.

## Critical Issues

### 1. Remove Non-functional --force Flags
**Priority**: CRITICAL - User-facing broken functionality
**Effort**: 30 minutes

#### Problem
- --force flags are defined in install, uninstall, add, and rm commands
- These flags appear in help text but do nothing
- Users expect them to work

#### Fix Plan
1. Remove flag definitions from:
   - `internal/commands/install.go` (line 46)
   - `internal/commands/uninstall.go`
   - `internal/commands/add.go`
   - `internal/commands/rm.go`
2. Remove any references to `force` variable in command implementations
3. Update help text and examples if needed
4. Test all affected commands

#### Verification
```bash
plonk install -h  # Should not show --force
plonk add -h      # Should not show --force
plonk uninstall -h # Should not show --force
plonk rm -h       # Should not show --force
```

---

### 2. Fix config/test/ Path Status Bug
**Priority**: CRITICAL - Broken core functionality
**Effort**: 1-2 hours

#### Problem
- ANY file under `config/test/` directory ALWAYS shows as "missing"
- Files exist and are deployed correctly
- Only affects status reconciliation
- Case-sensitive: `config/TEST/` works fine

#### Investigation Steps
1. Search for special handling of "test" in status/reconciliation code
2. Check if there's test file filtering logic
3. Review dotfile path comparison logic
4. Check for any ignore patterns affecting "test"

#### Fix Plan
1. Locate the bug in status reconciliation logic
2. Remove or fix special handling for "test" paths
3. Ensure path comparison is consistent
4. Add test case to prevent regression

#### Verification
```bash
mkdir -p ~/.config/test
echo "content" > ~/.config/test/anyfile
plonk add ~/.config/test/anyfile
plonk status --dotfiles  # Should show as "deployed" not "missing"
```

---

### 3. Fix Error Message Capture
**Priority**: HIGH - Poor user experience
**Effort**: 1-2 hours

#### Problem
- Package manager errors show generic "✗ failed" message
- Actual error output (e.g., "no bottle available") is not captured
- Users can't diagnose installation failures
- Bug #5 was marked as fixed but isn't working

#### Investigation
1. Review error handling in all package managers
2. Check if CombinedOutput() is capturing stderr
3. Verify error messages are being propagated

#### Fix Plan
1. Ensure all package managers capture full error output
2. Include stderr in error messages
3. Trim and format error messages for readability
4. Test with known failing packages

#### Verification
```bash
# Test with non-existent package
plonk install this-does-not-exist-xyz
# Should show: "No available formula with the name..."

# Test with ARM64 bottle issue
plonk install fzf  # on Linux ARM64
# Should show: "no bottle available!"
```

---

## Implementation Order

1. **--force flags removal** (30 min)
   - Simplest fix
   - High user visibility
   - No complex logic

2. **Error message capture** (1-2 hours)
   - Important for user experience
   - May already be partially implemented
   - Clear testing approach

3. **config/test/ path bug** (1-2 hours)
   - Most complex investigation
   - Core functionality issue
   - Requires understanding reconciliation logic

## Testing Plan

### Pre-fix Testing
1. Document current broken behavior for each issue
2. Create test scripts for each scenario

### Post-fix Testing
1. Run test scripts on macOS and Linux
2. Full regression test of all commands
3. Verify no new issues introduced

### Final Validation
1. Fresh Lima instance test
2. Complete bug validation suite
3. User acceptance testing

## Timeline
- Total effort: 3-5 hours
- Can be completed in one focused session
- Must be done before v1.0 release

## Success Criteria
- [ ] All commands show correct help text (no --force)
- [ ] config/test/ files show correct status
- [ ] Error messages show actual package manager output
- [ ] All existing tests pass
- [ ] No regressions in fixed bugs
