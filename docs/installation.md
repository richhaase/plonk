# Installation Guide

This guide covers installing plonk and setting up your development environment.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
- [Verification](#verification)
- [System-Specific Setup](#system-specific-setup)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required

- **Homebrew** - The primary package manager (install from https://brew.sh)
- **Git** - For cloning dotfiles repositories
- **Go 1.23 or later** - Only if installing via `go install`

### Optional Language Package Managers

Plonk can manage packages from these language-specific package managers:

- **Cargo** (Rust) - For Rust-based CLI tools, can be installed by `plonk doctor --fix`
- **npm** (Node.js) - For global JavaScript packages
- **pip** (Python) - For Python packages
- **gem** (Ruby) - For Ruby gems
- **go** (Go) - For Go modules

## Installation Methods

### Method 1: Homebrew (Coming Soon)

Once the Homebrew formula is available:

```bash
brew install plonk
```

### Method 2: Direct Go Install

Install the latest version directly from source:

```bash
go install github.com/richhaase/plonk/cmd/plonk@latest
```

### Method 3: Clone and Build

For development or if you need to modify the source:

```bash
git clone https://github.com/richhaase/plonk.git
cd plonk
go build -o plonk cmd/plonk/main.go
sudo mv plonk /usr/local/bin/  # Or add to your PATH
```

### Method 4: Pre-built Releases (Future)

Pre-built binaries will be available for major platforms in future releases.

## Verification

Verify your installation:

```bash
# Check plonk is installed and accessible
plonk --version

# Check system health and configuration
plonk doctor
```

The `plonk doctor` command will identify any missing dependencies and can automatically install supported package managers.

## System-Specific Setup

### macOS

```bash
# Install plonk
go install github.com/richhaase/plonk/cmd/plonk@latest

# Verify system health
plonk doctor

# Clone your dotfiles (installs needed package managers)
plonk clone user/dotfiles      # Clone existing setup
```

**macOS Notes:**
- Homebrew must be installed first (https://brew.sh)
- Xcode Command Line Tools may be required for some packages
- System Integrity Protection (SIP) may affect some operations

### Linux

```bash
# Install plonk
go install github.com/richhaase/plonk/cmd/plonk@latest

# Verify system health
plonk doctor

# Clone your dotfiles (installs needed package managers)
plonk clone user/dotfiles
```

**Linux Notes:**
- Homebrew must be installed first (https://brew.sh)
- Homebrew on Linux installs to `/home/linuxbrew/.linuxbrew`
- Ensure Homebrew is in your PATH after installation
- Language package managers (npm, pip, etc.) work identically to macOS
- Ensure your PATH includes `$GOPATH/bin` (usually `~/go/bin`)
- Docker/container environments are supported

### Windows

Plonk has limited Windows support. Use WSL2 for best experience:

```bash
# In WSL2
go install github.com/richhaase/plonk/cmd/plonk@latest
plonk doctor
```

## Environment Setup

### PATH Configuration

Ensure plonk is in your PATH. Add to your shell configuration:

```bash
# For bash (~/.bashrc) or zsh (~/.zshrc)
export PATH="$PATH:$(go env GOPATH)/bin"

# Reload your shell
source ~/.bashrc  # or ~/.zshrc
```

### Configuration Directory

Plonk uses `~/.config/plonk/` by default. Override with:

```bash
export PLONK_DIR="/path/to/your/config"
```

## Quick Start After Installation

### Option 1: Clone Existing Dotfiles

If you have an existing dotfiles repository:

```bash
plonk clone user/dotfiles                    # GitHub shorthand
plonk clone https://github.com/user/repo.git # Full URL
```

This will:
1. Clone your repository to `~/.config/plonk/`
2. Install missing language package managers (if needed)
3. Install all packages from `plonk.lock`
4. Deploy all dotfiles to your home directory

### Option 2: Start Fresh

To begin tracking your current setup:

```bash
# Create plonk directory
mkdir -p ~/.config/plonk

# Add your existing dotfiles
plonk add ~/.zshrc ~/.vimrc ~/.gitconfig

# Add packages you want to track
plonk install ripgrep fd bat exa

# Check status
plonk status

# Apply to deploy everything
plonk apply
```

## Troubleshooting

### Common Issues

**"plonk: command not found"**
- Ensure `$(go env GOPATH)/bin` is in your PATH
- Try `~/go/bin/plonk` directly to verify installation

**"Go version too old"**
- Update Go to 1.23 or later
- Use `go version` to check current version

**Permission denied errors**
- Check file permissions on `~/.config/plonk/`
- Run `plonk doctor` to diagnose permission issues

**Package manager not found**
- For Homebrew: Install manually from https://brew.sh (required prerequisite)
- For language package managers: Install manually or use `plonk clone` for automatic installation
- See [Configuration Guide](CONFIGURATION.md#package-manager-settings) for manual setup

### Getting Help

```bash
plonk --help                 # General help
plonk doctor                 # System health check with fix suggestions
```

For detailed command usage, see the [CLI Reference](cli.md).

### Debug Mode

Set environment variable for verbose output:

```bash
export PLONK_DEBUG=1
plonk doctor
```

## Next Steps

After successful installation:

1. **Clone existing setup** or **start tracking files**: `plonk clone user/dotfiles` or use `plonk add`
2. **Learn the commands**: See [CLI Reference](cli.md)
3. **Configure behavior**: See [Configuration Guide](CONFIGURATION.md)
4. **Understand the architecture**: See [Architecture Overview](architecture.md)
