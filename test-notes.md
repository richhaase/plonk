# Manual Test Notes

## Doctor Command Issues

### Issue: Misleading Package Manager Warning
The doctor command shows:
- "⚠️ Some package managers are not available"
- "Consider installing additional package managers for better compatibility"

This is misleading on macOS where `apt` cannot be installed. The warning should be platform-aware and only suggest installing package managers that are actually available for the current OS.

### Doctor Output (macOS)
# Plonk Doctor Report

🟡 Overall Status: WARNING
   2 warnings found

## System
✅ **System Requirements**: System requirements met
   • Go version: go1.24.4
   • OS: darwin
   • Architecture: arm64

## Environment
✅ **Environment Variables**: Environment variables configured
   • HOME: /Users/rdh
   • PATH entries: 24

✅ **PATH Configuration**: PATH is configured correctly
   • PATH contains 24 directories

## Permissions
✅ **File Permissions**: File permissions are correct
   • Config directory: /Users/rdh/src/plonk/manual-test
   • Config directory is writable

## Configuration
✅ **Configuration File**: Configuration file is accessible
   • Config file: /Users/rdh/src/plonk/manual-test/plonk.yaml
   • Size: 1400 bytes
   • Modified: 2025-07-21 11:12:05
   • Content length: 1400 characters

✅ **Configuration Validity**: Configuration is valid
   • Default manager: homebrew
   • Configured packages: 0
   • Auto-discovered dotfiles: 1

⚠️ **Lock File**: Lock file does not exist (will be created when packages are added)
   • Lock file path: /Users/rdh/src/plonk/manual-test/plonk.lock
   • Lock file will be automatically created when you add packages

 **Lock File Validity**: Lock file is valid but contains no packages
   • Total managed packages: 0
   • Lock file version: 1

## Package Managers
⚠️ **Package Manager Availability**: Some package managers are not available
   • pip: ✅ Available
   • gem: ✅ Available
   • go: ✅ Available
   • apt: ❌ Not available
   • homebrew: ✅ Available
   • npm: ✅ Available
   • cargo: ✅ Available
   Suggestions:
   💡 Consider installing additional package managers for better compatibility

✅ **Package Manager Functionality**: Package managers are functional
   • npm: Listed 5 installed packages
   • cargo: Listed 2 installed packages
   • pip: Listed 15 installed packages
   • gem: Listed 82 installed packages
   • go: Listed 7 installed packages
   • homebrew: Listed 124 installed packages

## Installation
✅ **Executable Path**: Executable is accessible
   • Executable: /Users/rdh/src/plonk/bin/plonk
   • Found in PATH: /Users/rdh/.asdf/shims/plonk



## Already Installed Package Behavior

### Command: `plonk install jq` (when already installed)

```
✗ jq - already managed

Summary: 0 added, 0 updated, 1 skipped, 0 failed
```

**Issue**: The output shows a red ✗ symbol next to "jq - already managed", which is misleading. The red X typically indicates an error or failure, but in this case the package is correctly managed and was properly skipped. This should use a different symbol (perhaps yellow ⚠ or blue ℹ) to indicate the package was skipped because it's already managed, not because of an error.


## Info Command Issue

### Problem: Info command doesn't respect installed package manager

When a package is installed via a non-default manager, the info command still shows information from the default manager instead of the actual installed package.

### Test Case:

1. Installed ripgrep via cargo:
```
$ plonk install --cargo ripgrep
✓ ripgrep@14.1.1 (cargo)

Summary: 1 added, 0 updated, 0 skipped, 0 failed
```

2. Verified installation:
```
$ which rg && rg --version
/Users/rdh/.cargo/bin/rg
ripgrep 14.1.1
```

3. Checked lock file (shows ripgrep under cargo):
```yaml
cargo:
    - name: ripgrep
      installed_at: 2025-07-21T11:32:07.032673-06:00
      version: 14.1.1
```

4. But info command shows homebrew (default) instead:
```
$ plonk info ripgrep
📦 Package 'ripgrep' available in homebrew (default)

Name: ripgrep
Homepage: https://github.com/Homebrew/homebrew-core/blob/HEAD/Formula/r/ripgrep.rb
Manager: homebrew
Installed: false
```

### Expected Behavior:
- If package is installed, show info from the manager that installed it
- If package is not installed, show info from default manager (or first available)
- The info should show "Installed: true" and the correct manager (cargo)


## Gem Install Error - Already Installed Package

### Problem: Gem install fails with generic error when package is already installed

When attempting to install a gem that's already installed, plonk fails with a generic error instead of recognizing it as already installed.

### Test Case:

1. Check that bundler is already installed:
```
$ gem list bundler
bundler (default: 1.17.2)
```

2. Try to install via plonk:
```
$ plonk install --gem bundler
Error: plonk install packages: failed to process 1 item(s)
...
Error: Failed to install package:
```

### Expected Behavior:
Should show something like:
```
✗ bundler - already installed

Summary: 0 added, 0 updated, 1 skipped, 0 failed
```

### Note:
This suggests the gem package manager's error detection may not be properly handling the "already installed" case, unlike other managers (npm, pip, etc.) which correctly identify and skip already installed packages.


## Gem Install Error - New Package Installation Fails

### Problem: Gem install fails with generic error for new packages

When attempting to install a gem that's NOT already installed, plonk still fails with a generic error.

### Test Case:

1. Verify solargraph is NOT installed:
```
$ gem list | grep solargraph
(no output - package not installed)
```

2. Try to install via plonk:
```
$ plonk install --gem solargraph
Error: plonk install packages: failed to process 1 item(s)
Usage:
  plonk install <packages...> [flags]

Flags:
      --brew      Use Homebrew package manager
      --cargo     Use Cargo package manager
  -n, --dry-run   Show what would be installed without making changes
  -f, --force     Force installation even if already managed
      --gem       Use gem package manager
      --go        Use go install package manager
  -h, --help      help for install
      --npm       Use NPM package manager
      --pip       Use pip package manager

Global Flags:
  -o, --output string   Output format (table|json|yaml) (default "table")

Error: Failed to install package:
```

### Expected Behavior:
Should successfully install the package and show:
```
✓ solargraph@X.X.X (gem)

Summary: 1 added, 0 updated, 0 skipped, 0 failed
```

### Note:
This indicates a broader issue with the gem package manager implementation - it appears to be failing for both already installed AND new packages. The error message provides no useful information about what went wrong.


## Uninstall Summary Bug

### Problem: Uninstall shows incorrect summary

When uninstalling a package, the summary line shows "0 added, 0 updated, 0 skipped, 0 failed" which doesn't reflect that a package was actually removed.

### Test Case:

```
$ plonk uninstall --pip six
✓ six (pip) - removed from configuration

Summary: 0 added, 0 updated, 0 skipped, 0 failed
```

### Expected Behavior:
The summary should indicate the removal, perhaps:
```
✓ six (pip) - removed from configuration

Summary: 0 added, 0 updated, 0 skipped, 0 failed, 1 removed
```

Or at minimum, the line "removed from configuration" should be sufficient without a misleading summary that suggests nothing happened.

### Note:
This same pattern was observed with all uninstall operations (npm, homebrew, etc.). The individual package line correctly shows the removal, but the summary doesn't account for removals.


## Lock File Update Bug

### Problem: Uninstall doesn't remove package from lock file

When uninstalling a package, it reports success but the package remains in the lock file.

### Test Case:

After uninstalling six:
```
$ plonk uninstall --pip six
✓ six (pip) - removed from configuration
```

The lock file still contains:
```yaml
pip:
    - name: six
      installed_at: 2025-07-21T11:24:41.26639-06:00
      version: 1.17.0
```

### Expected Behavior:
The package should be removed from the lock file when uninstalled.


## Unavailable Package Manager Error

### Problem: No helpful error when using unavailable package manager

When trying to use a package manager that's not available on the current OS, plonk shows a generic "unknown flag" error instead of explaining that the package manager is not available.

### Test Case:

On macOS, trying to use apt:
```
$ plonk install --apt vim
Error: unknown flag: --apt
Usage:
  plonk install <packages...> [flags]

Flags:
      --brew      Use Homebrew package manager
      --cargo     Use Cargo package manager
  -n, --dry-run   Show what would be installed without making changes
  -f, --force     Force installation even if already managed
      --gem       Use gem package manager
      --go        Use go install package manager
  -h, --help      help for install
      --npm       Use NPM package manager
      --pip       Use pip package manager

Global Flags:
  -o, --output string   Output format (table|json|yaml) (default "table")

Error: unknown flag: --apt
```

### Expected Behavior:
Should show a helpful message like:
```
Error: APT package manager is not available on macOS

APT is a Linux-only package manager used on Debian/Ubuntu systems.
On macOS, consider using Homebrew instead:
  plonk install --brew vim
```

### Note:
The doctor command correctly identifies that apt is not available, but the install command doesn't provide this context when the flag is used. The flag shouldn't be hidden - it should be available but provide helpful guidance when used on incompatible systems.


## Summary of Bugs Found During Manual Testing

1. **Doctor Command - Misleading Package Manager Suggestion**: Suggests installing package managers that aren't available on the current OS (e.g., apt on macOS)

2. **Already Installed Package - Misleading Icon**: Shows red ✗ for "already managed" packages, suggesting error when it's actually correct behavior

3. **Info Command - Wrong Manager**: Shows default manager info instead of the actual manager that installed a package

4. **Gem Manager - Complete Failure**: Both new installations and already-installed packages fail with generic error

5. **Uninstall Summary - Missing Count**: Summary line doesn't reflect removals (shows 0 for all categories)

6. **Lock File - Not Updated on Uninstall**: Packages remain in lock file after being uninstalled

7. **Unavailable Package Manager - Poor Error**: Shows "unknown flag" instead of explaining the manager isn't available on this OS

### Testing Status

✅ **Working Correctly**:
- homebrew: install/info/ls/uninstall
- npm: install/info/ls/uninstall
- pip: install/info/ls/uninstall
- cargo: install/ls/uninstall (info shows wrong manager)
- go: install/info/ls/uninstall

❌ **Not Working**:
- gem: install fails for all packages
- All managers: uninstall doesn't update lock file properly

### Integration Test Validation

The integration test we created covers the core workflows well, but these bugs show areas where the test assertions could be strengthened:
- Verify lock file is updated after uninstall
- Check that info shows the correct manager for installed packages
- Test gem manager functionality (currently broken)
- Validate error messages are helpful, not generic
