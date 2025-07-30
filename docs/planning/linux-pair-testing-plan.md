# Linux Platform Testing - Pair Programming Plan

## Overview
This is a structured pair programming test plan where:
- **You (Driver)**: Execute commands on the Lima VM
- **Me (Navigator)**: Guide tests, interpret results, suggest fixes

## Test Environment Setup

### Step 1: Create Fresh Ubuntu VM
```bash
# Create Ubuntu 22.04 VM
limactl create --name=plonk-test-ubuntu --vm-type=vz --cpus=2 --memory=4 template://ubuntu-22.04
limactl start plonk-test-ubuntu
limactl shell plonk-test-ubuntu
```

**Expected**: VM starts successfully
**I need to see**: VM creation output, any errors

### Step 2: Install Prerequisites
```bash
# Update system and install basics
sudo apt update
sudo apt install -y build-essential curl git

# Check installations
git --version
gcc --version
```

**Expected**: Git and build tools installed
**I need to see**: Version outputs

### Step 3: Install Homebrew
```bash
# Install Homebrew (our prerequisite)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# IMPORTANT: Configure PATH for Linux
echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

# Verify Homebrew
brew --version
which brew
echo $PATH | grep linuxbrew
```

**Expected**: Homebrew installed to `/home/linuxbrew/.linuxbrew`
**I need to see**: Installation output, PATH confirmation

### Step 4: Install Plonk
```bash
# Option A: Install via Go (for testing latest)
brew install go  # Install Go via Homebrew
go install github.com/richhaase/plonk/cmd/plonk@latest

# Add Go bin to PATH if needed
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify plonk
plonk --version
which plonk
```

**Expected**: Plonk installed and accessible
**I need to see**: Version output, installation path

## Core Functionality Tests

### Test 1: Doctor Command
```bash
# Basic doctor check
plonk doctor

# Check output format
plonk doctor --output json | jq .

# Save output for analysis
plonk doctor > doctor-output.txt
cat doctor-output.txt
```

**Expected**:
- Homebrew shows as available (PASS)
- No sudo prompts
- Clean formatting

**I need to see**: Full doctor output

### Test 2: Basic Package Operations
```bash
# Test install with single package
plonk install ripgrep

# Check installation
which rg
rg --version
plonk status --packages

# Test install with multiple packages
plonk install fd bat

# Verify
plonk status
ls -la ~/.config/plonk/plonk.lock
```

**Expected**:
- Progress indicators show
- Packages install successfully
- Lock file created/updated

**I need to see**: Install output, status output, lock file content

### Test 3: Package Manager Prefixes
```bash
# Test explicit brew prefix
plonk install brew:jq brew:htop

# Test search with prefix
plonk search brew:rust

# Test info
plonk info brew:ripgrep
```

**Expected**: All operations work with brew: prefix
**I need to see**: Command outputs

### Test 4: Dotfile Management
```bash
# Create test dotfiles
echo "# Test bashrc" > ~/.bashrc_test
echo "set number" > ~/.vimrc_test
mkdir -p ~/.config/test
echo "test: true" > ~/.config/test/settings.yml

# Add to plonk
plonk add ~/.bashrc_test ~/.vimrc_test ~/.config/test/

# Check status
plonk status --dotfiles
ls -la ~/.config/plonk/

# Test deployment
rm ~/.bashrc_test
plonk apply --dotfiles
test -f ~/.bashrc_test && echo "SUCCESS: File restored"
```

**Expected**: Dotfiles managed correctly
**I need to see**: Status output, directory listings

### Test 5: Drift Detection
```bash
# Modify a managed dotfile
echo "# Modified" >> ~/.bashrc_test

# Check drift detection
plonk status --dotfiles | grep drifted

# Test diff command
plonk diff
plonk diff ~/.bashrc_test

# Restore original
plonk apply --dotfiles
plonk status --dotfiles
```

**Expected**: Drift detected and shown in yellow
**I need to see**: Status showing drift, diff output

### Test 6: Language Package Managers
```bash
# Install Node.js first
plonk install node

# Test npm packages
plonk install npm:prettier npm:typescript

# Check installation
npm list -g
plonk status | grep npm

# Test Python packages
plonk install python
plonk install pip:black pip:ruff

# Verify
pip list | grep -E "(black|ruff)"
plonk status | grep pip
```

**Expected**: Language packages work via Homebrew-installed interpreters
**I need to see**: Installation outputs, package listings

### Test 7: Clone Workflow
```bash
# Create test repository
mkdir ~/test-dotfiles
cd ~/test-dotfiles
git init

# Add plonk content
echo "export TEST_VAR=linux" > bashrc
echo "brew:tree" > plonk.packages
cat > plonk.yaml << 'EOF'
default_manager: brew
operation_timeout: 300
EOF

git add -A
git commit -m "Test dotfiles"

# Test clone in new directory
cd ~
rm -rf ~/.config/plonk
plonk clone ~/test-dotfiles

# Verify results
plonk status
test -f ~/.bashrc && echo "SUCCESS: bashrc deployed"
which tree && echo "SUCCESS: tree installed"
```

**Expected**: Clone sets up complete environment
**I need to see**: Clone output, final status

## Edge Cases & Error Handling

### Test 8: Missing Homebrew
```bash
# This test requires fresh VM or PATH manipulation
# Temporarily hide brew
OLD_PATH=$PATH
export PATH=/usr/bin:/bin

plonk doctor
plonk install test-package

# Restore
export PATH=$OLD_PATH
```

**Expected**: Clear error about Homebrew being required
**I need to see**: Error messages

### Test 9: Permission Issues
```bash
# Create read-only config
mkdir -p ~/.config/plonk
touch ~/.config/plonk/test.lock
chmod 444 ~/.config/plonk/test.lock

# Try operation that needs write
plonk install brew:wget

# Cleanup
chmod 644 ~/.config/plonk/test.lock
rm ~/.config/plonk/test.lock
```

**Expected**: Clear permission error
**I need to see**: Error output

### Test 10: Concurrent Operations
```bash
# Test lock handling
plonk install brew:tmux &
PID1=$!
plonk install brew:vim &
PID2=$!

wait $PID1 $PID2
echo "Exit codes: $? $?"

plonk status --packages | grep -E "(tmux|vim)"
```

**Expected**: One waits for the other, both succeed
**I need to see**: Any lock messages, final status

## Performance Comparison

### Test 11: Operation Timing
```bash
# Time various operations
time plonk status
time plonk search rust
time plonk install brew:curl
time plonk doctor
```

**Expected**: Reasonable performance (compare to macOS)
**I need to see**: Timing results

## Linux-Specific Checks

### Test 12: Path and Environment
```bash
# Check PATH handling
echo $PATH | tr ':' '\n' | grep -E "(brew|plonk|go)"

# Check plonk env
plonk env
plonk env --shell

# Test in new shell
bash -c 'plonk --version'
```

**Expected**: PATH correctly configured
**I need to see**: PATH entries, env output

### Test 13: Different Package Availability
```bash
# Test packages that might differ on Linux
plonk search docker
plonk search mac
plonk info brew:mas  # Mac App Store CLI
```

**Expected**: Some packages may not be available
**I need to see**: Search results, error messages

## Final Validation

### Test 14: Clean System Test
```bash
# Remove plonk directory
rm -rf ~/.config/plonk

# Start fresh
plonk status
plonk add ~/.bashrc_test
plonk install ripgrep fd bat
plonk status

# Output formats
plonk status --output json > status.json
plonk status --output yaml > status.yaml
cat status.json | jq .
```

**Expected**: Clean start works properly
**I need to see**: All outputs

## Reporting Checklist

After each test, please provide:

1. **Command executed** (exact)
2. **Output received** (full text or screenshot)
3. **Any errors or warnings**
4. **Unexpected behavior**
5. **Performance observations**

## Success Criteria

- [ ] All basic commands work without errors
- [ ] Homebrew integration works smoothly
- [ ] Dotfile operations identical to macOS
- [ ] Package operations identical to macOS
- [ ] Clear error messages for Linux-specific issues
- [ ] No sudo required (except system package installation)
- [ ] Performance acceptable
- [ ] Lock files compatible with macOS

## Quick Test Script

For rapid testing, save this as `quick-test.sh`:

```bash
#!/bin/bash
set -e

echo "=== Quick Plonk Linux Test ==="
echo "1. Version check"
plonk --version

echo -e "\n2. Doctor check"
plonk doctor | grep -E "(PASS|WARN|ERROR)"

echo -e "\n3. Package test"
plonk install brew:jq
plonk status --packages | grep jq

echo -e "\n4. Dotfile test"
echo "test=1" > ~/.plonktest
plonk add ~/.plonktest
plonk status --dotfiles | grep plonktest
plonk rm plonktest
rm ~/.plonktest

echo -e "\n✓ Quick test passed!"
```

Ready to begin testing! Please start with Step 1 and share the output as we go.
