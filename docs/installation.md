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
- **Go 1.23 or later** - Only if building from source (note: Go 1.24+ works as well)

### Optional Language Package Managers

Plonk can manage packages from these language-specific package managers:

- **Cargo** (Rust) - For Rust-based CLI tools, automatically installed by `plonk clone` when needed
- **npm** (Node.js) - For global JavaScript packages
- **pnpm** (Node.js) - For fast, disk-efficient JavaScript packages
- **pipx** (Python) - For Python applications in isolated environments
- **conda** (Python) - For scientific computing and data science packages
- **gem** (Ruby) - For Ruby gems
- **go** (Go) - For Go modules
- **uv** (Python) - For Python tools management
- **pixi** (Conda-forge) - For conda-forge packages
- **composer** (PHP) - For PHP global packages
- **dotnet** (.NET) - For .NET Global Tools

## Installation Methods

### Method 1: Homebrew (Recommended)

The easiest way to install plonk on macOS:

```bash
# Add the tap and install
brew tap richhaase/tap
brew install plonk

# Or install directly:
brew install richhaase/tap/plonk
```

**Benefits of Homebrew installation:**
- Automatic updates via `brew upgrade`
- Code signed and notarized for macOS (no security warnings)
- Managed installation/uninstallation
- Pre-built binaries for both Intel and Apple Silicon

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

# Using just (recommended if installed)
just build
# Binary will be in bin/plonk

# Or using go directly
go build -o plonk cmd/plonk/main.go

# Install to system
sudo mv plonk /usr/local/bin/  # Or add to your PATH
# Or from bin/ if using just
sudo mv bin/plonk /usr/local/bin/
```

### Method 4: Pre-built Releases

Download pre-built binaries from the [releases page](https://github.com/richhaase/plonk/releases).

```bash
# Example for macOS arm64
curl -L https://github.com/richhaase/plonk/releases/latest/download/plonk_Darwin_arm64.tar.gz | tar xz
sudo mv plonk /usr/local/bin/
```

## Verification

Verify your installation:

```bash
# Check plonk is installed and accessible
plonk --version

# Check system health and configuration
plonk doctor
```

The `plonk doctor` command will identify any missing dependencies. Use `plonk clone` to automatically install package managers needed by your managed packages.

## Uninstallation

### Uninstall via Homebrew

```bash
brew uninstall plonk
brew untap richhaase/tap  # Optional: remove the tap
```

### Manual Uninstallation

```bash
# Remove the binary
sudo rm /usr/local/bin/plonk
# Or if installed via go install
rm $(go env GOPATH)/bin/plonk

# Optionally remove configuration and dotfiles
# WARNING: This removes all your plonk-managed dotfiles!
rm -rf ~/.config/plonk
```

## System-Specific Setup

### macOS

```bash
# Install plonk via Homebrew (recommended)
brew install richhaase/tap/plonk

# Or via Go
go install github.com/richhaase/plonk/cmd/plonk@latest

# Verify system health
plonk doctor

# Clone your dotfiles (automatically installs package managers for managed packages)
plonk clone user/dotfiles      # GitHub shorthand for existing setup
```

**macOS Notes:**
- Homebrew must be installed first (https://brew.sh)
- Homebrew installation includes code signing and notarization
- Xcode Command Line Tools may be required for some packages
- System Integrity Protection (SIP) may affect some operations

### Linux

```bash
# Install plonk
go install github.com/richhaase/plonk/cmd/plonk@latest

# Verify system health
plonk doctor

# Clone your dotfiles (automatically installs package managers for managed packages)
plonk clone user/dotfiles      # GitHub shorthand
```

**Linux Notes:**
- Homebrew must be installed first (https://brew.sh)
- Homebrew on Linux installs to `/home/linuxbrew/.linuxbrew`
- Ensure Homebrew is in your PATH after installation
- Language package managers (npm, uv, etc.) work identically to macOS
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
- Update Go to 1.23 or later (1.24+ is also supported)
- Use `go version` to check current version
- The project uses Go 1.23.10 but newer versions work

**Permission denied errors**
- Check file permissions on `~/.config/plonk/`
- Run `plonk doctor` to diagnose permission issues

**Package manager not found**
- For Homebrew: Install manually from https://brew.sh (required prerequisite)
- For language package managers: Use `plonk clone` to automatically install those needed by your managed packages, or install manually
- See [Configuration Guide](configuration.md#package-manager-settings) for manual setup

### Getting Help

```bash
plonk --help                 # General help
plonk doctor                 # System health check with fix suggestions
```

For detailed command usage, see the [CLI Reference](cli.md).


## Next Steps

After successful installation:

1. **Clone existing setup** or **start tracking files**: `plonk clone user/dotfiles` or use `plonk add`
2. **Learn the commands**: See [CLI Reference](cli.md)
3. **Configure behavior**: See [Configuration Guide](configuration.md)
4. **Understand the architecture**: See [Architecture Overview](architecture.md)
