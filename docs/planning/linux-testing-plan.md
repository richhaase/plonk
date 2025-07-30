# Linux Platform Testing Plan

## Overview
Since plonk currently only supports APT package manager, our Linux testing will focus exclusively on Debian-based distributions (Ubuntu and Debian).

## Testing Scope

### Supported Platforms
- **Ubuntu 22.04 LTS** (Jammy) - Primary target
- **Ubuntu 24.04 LTS** (Noble) - Latest LTS
- **Debian 12** (Bookworm) - Current stable
- **WSL2 Ubuntu** - Windows Subsystem for Linux

### Not Testing
- Fedora/RHEL (no dnf/yum support)
- Arch Linux (no pacman support)
- openSUSE (no zypper support)
- Other non-Debian distributions

## Test Categories

### 1. Installation Testing
- [ ] Install plonk via `go install`
- [ ] Verify PATH setup instructions work
- [ ] Test with different Go versions (1.22, 1.23)

### 2. Package Manager Testing
- [ ] Verify APT is detected as available
- [ ] Test `plonk doctor` shows APT correctly
- [ ] Test package operations:
  - [ ] Install (with sudo)
  - [ ] Uninstall (with sudo)
  - [ ] Search
  - [ ] Info
- [ ] Verify error messages for non-sudo operations
- [ ] Test with packages that require `--no-install-recommends`

### 3. Core Functionality Testing
- [ ] Clone dotfiles repository
- [ ] Add/remove dotfiles
- [ ] Apply dotfiles
- [ ] Drift detection
- [ ] Diff command
- [ ] Status command output

### 4. Platform-Specific Testing
- [ ] Path handling (~/. vs /home/user)
- [ ] Permission handling
- [ ] Default directories (~/.config/plonk)
- [ ] File system case sensitivity

### 5. Integration Testing
- [ ] Full setup flow: install plonk → clone → apply
- [ ] Mixed package managers (brew on Linux)
- [ ] Lock file compatibility
- [ ] Configuration handling

## Test Environments

### Local Testing (Optional)
- Docker containers for each distribution
- VirtualBox/VMware VMs
- WSL2 on Windows

### CI Testing (Primary)
- GitHub Actions ubuntu-latest (currently 22.04)
- Integration tests already in place
- Consider adding matrix for multiple Ubuntu versions

## Test Script

```bash
#!/bin/bash
# Basic Linux functionality test

# 1. Installation
go install github.com/richhaase/plonk/cmd/plonk@latest
export PATH="$PATH:$(go env GOPATH)/bin"

# 2. Verify installation
plonk --version
plonk doctor

# 3. Test APT operations
sudo plonk install apt:tree
plonk status --packages
plonk info apt:tree
sudo plonk uninstall apt:tree

# 4. Test dotfile operations
echo "test_config=true" > ~/.testrc
plonk add ~/.testrc
plonk status --dotfiles
plonk apply

# 5. Test drift detection
echo "modified=true" >> ~/.testrc
plonk status --dotfiles | grep drifted
plonk diff

# 6. Cleanup
plonk rm testrc
rm ~/.testrc
```

## Success Criteria

### Must Pass
- [ ] All commands work without errors on Ubuntu 22.04
- [ ] APT operations work with proper sudo handling
- [ ] Dotfile operations work correctly
- [ ] Error messages are clear and actionable

### Should Pass
- [ ] Works on Ubuntu 24.04 and Debian 12
- [ ] WSL2 compatibility verified
- [ ] Performance is acceptable

## Known Limitations
1. APT requires sudo - clear error messages provided
2. No automatic `apt update` - user must run manually
3. Package name differences from Homebrew not mapped

## Documentation Updates Needed
After testing:
1. Update installation guide with Linux-specific notes
2. Document any Linux-specific behaviors
3. Add troubleshooting section for common Linux issues

## Timeline
- Day 1: Set up test environments, run basic tests
- Day 2: Complete all test categories, document issues
- Day 3: Fix any critical issues, update documentation

## Notes
- Focus on Ubuntu as primary platform
- Debian and WSL2 are secondary but should work
- Don't over-test unsupported platforms
