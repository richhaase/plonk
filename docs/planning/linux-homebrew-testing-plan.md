# Linux Testing Plan - Homebrew Focus

## Overview
Test plonk on Linux using Homebrew as the primary package manager, ensuring feature parity with macOS.

## Prerequisites
- Homebrew must be installed before plonk (this is now a requirement)
- Git must be installed

## Testing Scope

### Supported Platforms
- **Ubuntu 22.04 LTS** (Primary target)
- **Ubuntu 24.04 LTS** (Latest LTS)
- **WSL2 Ubuntu** (Windows Subsystem for Linux)

### Key Differences from macOS
- Homebrew installs to `/home/linuxbrew/.linuxbrew` on Linux
- Different PATH setup required
- Some formulae may not be available on Linux

## Test Categories

### 1. Fresh Linux Setup with Plonk
Test the complete new user experience on a fresh Linux VM.

```bash
# Start fresh Ubuntu VM
limactl create --name=plonk-linux-test --vm-type=vz --cpus=2 --memory=4 template://ubuntu-22.04
limactl start plonk-linux-test
limactl shell plonk-linux-test

# Install prerequisites
sudo apt update
sudo apt install -y build-essential curl git

# Install Homebrew (required prerequisite)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Add Homebrew to PATH (Linux specific)
echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

# Install Go (if testing via go install)
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install plonk
go install github.com/richhaase/plonk/cmd/plonk@latest

# Verify installation
plonk --version
```

### 2. Doctor Command Testing
Test that doctor correctly identifies Homebrew as available.

```bash
# Doctor check (should show Homebrew available)
plonk doctor

# Check specific messages
plonk doctor | grep -E "(brew|Homebrew)"

# Test doctor --fix flow
plonk doctor --fix --yes

# Verify Homebrew was installed
which brew
brew --version

# Re-run doctor to confirm
plonk doctor
```

### 3. Package Management Testing
Test all package operations using Homebrew on Linux.

```bash
# Test package installation
plonk install ripgrep fd bat jq

# Verify packages installed
plonk status --packages
which rg fd bat jq

# Test search functionality
plonk search brew:rust
plonk search brew:neovim

# Test info command
plonk info brew:ripgrep
plonk info brew:bat

# Test uninstall
plonk uninstall brew:jq
plonk status --packages | grep -v jq
```

### 4. Language Package Manager Testing
Verify language-specific package managers work identically to macOS.

```bash
# Install Node.js via Homebrew first
plonk install brew:node

# Test npm packages
plonk install npm:prettier npm:eslint
plonk status --packages | grep npm

# Install Python via Homebrew
plonk install brew:python@3.12

# Test pip packages
plonk install pip:black pip:ruff
plonk status --packages | grep pip

# Test other language managers as available
```

### 5. Dotfile Management Testing
Full dotfile workflow should work identically to macOS.

```bash
# Create test dotfiles
echo "export PLONK_TEST=true" > ~/.testrc
echo "set number" > ~/.testvimrc
mkdir -p ~/.config/test && echo "test: true" > ~/.config/test/config.yml

# Add dotfiles
plonk add ~/.testrc ~/.testvimrc ~/.config/test

# Check status
plonk status --dotfiles

# Test apply
rm ~/.testrc
plonk apply --dotfiles
test -f ~/.testrc && echo "✓ Dotfile restored"

# Test drift detection
echo "modified=true" >> ~/.testrc
plonk status --dotfiles | grep drifted
plonk diff

# Test removal
plonk rm testrc
plonk status --dotfiles | grep -v testrc
```

### 6. Clone Workflow Testing
Test the primary use case: cloning a dotfiles repo on a new machine.

```bash
# Create a test dotfiles repository
mkdir ~/test-dotfiles
cd ~/test-dotfiles
git init

# Add some test content
echo "export TEST=1" > zshrc
echo "brew:ripgrep" > packages.txt
git add .
git commit -m "Initial dotfiles"

# Push to a test repo (or use local path)
cd ~

# Test clone with local path
plonk clone ~/test-dotfiles

# Verify everything was set up
plonk status
test -f ~/.zshrc && echo "✓ Dotfiles deployed"
which rg && echo "✓ Packages installed"
```

### 7. Edge Cases and Error Handling

```bash
# Test behavior without Homebrew
# (Would need to uninstall/use fresh VM)

# Test with read-only directories
chmod 444 ~/.config/plonk
plonk install brew:tree
# Should show clear error

# Test with missing PATH
PATH=/usr/bin:/bin plonk doctor
# Should show PATH configuration needed

# Test concurrent operations
plonk install brew:htop &
plonk install brew:curl &
wait
# Should handle lock contention gracefully
```

### 8. Performance Testing
Compare performance between Linux and macOS.

```bash
# Time package operations
time plonk search rust

# Time large package installation
time plonk install brew:neovim

# Time status with many packages
plonk install brew:git brew:wget brew:tree brew:jq brew:fd
time plonk status
```

## Success Criteria

### Must Work
- [ ] Homebrew installation via `plonk doctor --fix`
- [ ] All package operations (install/uninstall/search/info)
- [ ] All dotfile operations (add/rm/apply/diff)
- [ ] Clone workflow for new machine setup
- [ ] Status and doctor commands
- [ ] Drift detection and diff

### Should Work
- [ ] Performance comparable to macOS
- [ ] Clear error messages for Linux-specific issues
- [ ] PATH configuration guidance
- [ ] Lock file compatibility between platforms

### Known Differences
- Homebrew location: `/home/linuxbrew/.linuxbrew` vs `/opt/homebrew`
- Some formulae may be macOS-only
- PATH setup is more manual on Linux

## Test Script
Create `test-linux.sh` for automated testing:

```bash
#!/bin/bash
set -e

echo "=== Plonk Linux Test Suite ==="

# 1. Check environment
echo "→ Testing environment..."
plonk --version
plonk doctor

# 2. Install Homebrew if needed
if ! command -v brew &> /dev/null; then
    echo "→ Installing Homebrew..."
    plonk doctor --fix --yes
fi

# 3. Test package operations
echo "→ Testing package management..."
plonk install brew:ripgrep brew:fd
plonk status --packages
plonk uninstall brew:fd

# 4. Test dotfile operations
echo "→ Testing dotfile management..."
echo "test=1" > ~/.plonk-test
plonk add ~/.plonk-test
plonk status --dotfiles
plonk rm plonk-test
rm -f ~/.plonk-test

# 5. Summary
echo "✓ All tests passed!"
```

## Lima Helper Script
Quick setup for testing:

```bash
#!/bin/bash
# setup-lima-test.sh
VM_NAME="plonk-linux-${1:-test}"

# Create and start VM
limactl create --name="$VM_NAME" --vm-type=vz --cpus=2 --memory=4 template://ubuntu-22.04
limactl start "$VM_NAME"

# Copy this test plan into VM
limactl copy linux-homebrew-testing-plan.md "$VM_NAME":~/

# Enter VM
limactl shell "$VM_NAME"
```

## Reporting Template
When testing, please note:

1. **Environment**:
   - Linux distribution and version
   - Go version
   - Fresh install or existing system

2. **Homebrew Installation**:
   - Did doctor detect missing Homebrew?
   - Did doctor --fix work smoothly?
   - Any PATH issues?

3. **Package Operations**:
   - Install/uninstall success rate
   - Search performance
   - Any Linux-specific package issues

4. **Dotfile Operations**:
   - All operations working?
   - Drift detection accurate?
   - Performance acceptable?

5. **Overall**:
   - Any Linux-specific error messages needed?
   - Documentation improvements needed?
   - Feature parity with macOS achieved?

This comprehensive testing ensures plonk works as well on Linux as it does on macOS!
