# Lima Testing Setup for Plonk

## Quick Start with Lima

### 1. Create Ubuntu Test VMs

```bash
# Ubuntu 22.04 LTS (Primary target)
limactl create --name=plonk-ubuntu-22 --vm-type=vz --mount-writable --cpus=2 --memory=4 template://ubuntu-22.04

# Ubuntu 24.04 LTS (Latest)
limactl create --name=plonk-ubuntu-24 --vm-type=vz --mount-writable --cpus=2 --memory=4 template://ubuntu-24.04

# Start the VM
limactl start plonk-ubuntu-22
```

### 2. Access the VM

```bash
# Shell into the VM
limactl shell plonk-ubuntu-22

# Or use lima shortcut
lima-plonk-ubuntu-22
```

### 3. Install Prerequisites in VM

```bash
# Update package lists
sudo apt update

# Install Go (if not present)
sudo apt install -y golang-go git

# Verify Go version
go version

# Set up Go path
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
```

### 4. Install and Test Plonk

```bash
# Install plonk
go install github.com/richhaase/plonk/cmd/plonk@latest

# Verify installation
plonk --version
plonk doctor
```

### 5. Run Test Suite

Use the test script from linux-testing-plan.md, or test interactively:

```bash
# Test APT detection
plonk doctor | grep apt

# Test package operations
sudo plonk install apt:tree
plonk status --packages
sudo plonk uninstall apt:tree

# Test error handling (should fail with sudo message)
plonk install apt:htop
```

## Things to Watch For

1. **Go Version**: Ubuntu 22.04 might have older Go, may need manual install
2. **PATH Issues**: Ensure ~/go/bin is in PATH
3. **Sudo Behavior**: Test both with and without sudo
4. **Performance**: Note any slow operations

## Cleanup

```bash
# Stop VM
limactl stop plonk-ubuntu-22

# Delete VM when done
limactl delete plonk-ubuntu-22
```

## Reporting Back

When testing, please note:
- Which Ubuntu version
- Any error messages
- Unexpected behaviors
- Performance issues
- Success cases

This will help me update documentation and fix any Linux-specific issues!
