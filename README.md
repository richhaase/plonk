# Plonk

> ⚠️ **WARNING: This project is under active development.** APIs, commands, and configuration formats may change without notice. Use at your own risk in production environments.

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)
[![Security](https://github.com/richhaase/plonk/workflows/Security%20Check/badge.svg)](https://github.com/richhaase/plonk/actions)
[![codecov](https://codecov.io/gh/richhaase/plonk/branch/main/graph/badge.svg)](https://codecov.io/gh/richhaase/plonk)

A unified package and dotfile manager for developers that maintains consistency across multiple machines.

## What is Plonk?

Plonk manages your development environment by tracking packages and dotfiles automatically. It uses state reconciliation to compare your desired state with your actual system state and applies the necessary changes.

**🚀 Zero Configuration Required** - Start using Plonk immediately without any setup!

**Key features:**
- **Zero-config**: Works immediately with sensible defaults - no setup required
- **Intelligent detection**: Automatically recognizes packages vs dotfiles
- **Unified management**: Packages (Homebrew, NPM, Cargo, Pip, Gem, APT, Go Install) and dotfiles in one tool
- **State reconciliation**: Compares desired vs actual state and applies changes
- **Auto-discovery**: Finds dotfiles automatically with configurable ignore patterns
- **Shell completion**: Tab completion for commands, package names, and file paths
- **AI-friendly**: Structured output formats (JSON/YAML) and clear command syntax
- **Cross-platform**: Works on macOS, Linux, and Windows

**🚀 CLI Benefits:**
- **Unix-style commands**: Familiar `add`, `rm`, `status`, `apply` commands
- **Smart operations**: Mixed package/dotfile operations in single commands
- **Zero-argument status**: Just type `plonk` for system overview (like git)
- **One-command workflows**: `plonk install` = install package + add to management

## Quick Start

### Installation

The easiest way to install plonk is using Go's built-in package manager:

```bash
go install github.com/richhaase/plonk/cmd/plonk@latest
```

**Requirements:**
- Go 1.24.4+

#### Alternative: Build from source

For development or if you need a specific version:

```bash
git clone https://github.com/richhaase/plonk
cd plonk
just dev-setup  # Sets up development environment
just install    # Installs plonk globally
```

**Additional requirements for building from source:**
- Just (command runner)
- Git
- Pre-commit (optional, for development)

### Basic Usage

**🎉 No setup required!** Plonk works immediately with zero-config:

```bash
# Check your current environment
plonk status             # Show managed packages and dotfiles
plonk st                 # Short alias for status

# Install packages (automatically adds to management)
plonk install ripgrep    # Install with default manager
plonk install brew:htop  # Install with specific manager

# Manage dotfiles
plonk add ~/.vimrc       # Add dotfile to management
plonk rm ~/.vimrc        # Remove dotfile from management

# Apply all changes (reinstall missing packages, restore dotfiles)
plonk apply              # Apply everything
plonk apply --dry-run    # See what would happen

# Get help
plonk --help             # Show all commands
plonk help install       # Get help for specific command
```

See [CLI Reference](docs/CLI.md) for complete command documentation.

### Shell Completion

Enable tab completion for enhanced productivity:

```bash
# Bash
source <(plonk completion bash)

# Zsh
source <(plonk completion zsh)

# Fish
plonk completion fish > ~/.config/fish/completions/plonk.fish

# PowerShell
plonk completion powershell | Out-String | Invoke-Expression
```

For permanent installation, see the [CLI Reference](docs/CLI.md#shell-completion).

## Configuration

**🚀 Zero-config required!** Plonk works immediately with sensible defaults.

**Optional customization:** Create a config file only when you need to change defaults:

```bash
plonk init           # Create config template
plonk config show    # View current settings (including defaults)
plonk config edit    # Edit your settings
```

**Two file types:**
- **`plonk.yaml`** - Your optional settings (timeouts, preferences)
- **`plonk.lock`** - Automatic package tracking (like package-lock.json)

> **📖 Complete Configuration Guide**: See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for detailed configuration options, examples, and file formats.

## Getting Started Guide

### First Time Setup

1. **Install Plonk** (see Installation section above)

2. **Set up your environment:**
```bash
plonk setup              # Initialize plonk and install missing tools
plonk doctor             # Check system health
```

3. **Check what's managed:**
```bash
plonk status             # Shows managed packages and dotfiles
plonk st                 # Short alias
```

3. **Start managing packages and dotfiles:**
```bash
# Install packages
plonk install git neovim ripgrep htop

# Install packages with specific manager
plonk install npm:typescript         # NPM packages
plonk install cargo:ripgrep          # Rust/Cargo packages
plonk install pip:black              # Python/Pip packages
plonk install gem:bundler            # Ruby/Gem packages
plonk install apt:htop               # APT packages (Linux)
plonk install go:golangci-lint       # Go packages

# Add dotfiles to management
plonk add ~/.vimrc ~/.zshrc ~/.gitconfig

# Add entire directories of dotfiles
plonk add ~/.config/nvim/

# See what's available to manage
plonk status --verbose            # See everything including untracked
```

4. **Check system health:**
```bash
plonk doctor
# Verifies:
# - Package managers are working
# - File permissions are correct
# - Configuration is valid
```

5. **Optional: Customize settings** (only if needed)
```bash
plonk config show    # Show current configuration
plonk config edit    # Edit configuration file
```

### Daily Workflow

```bash
plonk status           # Check managed packages and dotfiles
plonk apply            # Reinstall missing packages, restore dotfiles
plonk install <pkg>    # Install new packages
plonk add <dotfile>    # Add dotfiles to management
plonk uninstall <pkg>  # Uninstall packages
plonk rm <dotfile>     # Remove dotfiles from management
plonk doctor           # System health check
```

## Common Commands

```bash
# Essential workflows
plonk status                                      # Check managed packages and dotfiles
plonk st                                          # Short alias for status
plonk apply                                       # Apply all changes

# Package management
plonk install git neovim ripgrep                 # Install packages and add to management
plonk install npm:typescript                     # Install with specific manager
plonk install pip:black                          # Python packages
plonk install gem:bundler                        # Ruby packages
plonk install apt:htop                           # APT packages (Linux)
plonk install go:golangci-lint                   # Go packages
plonk uninstall htop                             # Uninstall package and remove from management

# Dotfile management
plonk add ~/.vimrc ~/.zshrc ~/.gitconfig         # Add dotfiles to management
plonk add ~/.config/nvim/                        # Add directory of dotfiles
plonk rm ~/.vimrc                                # Remove dotfile from management

# Search and info
plonk search ripgrep                              # Search all package managers
plonk search brew:ripgrep                         # Search specific manager
plonk info ripgrep                                # Show package details

# System overview
plonk status                                      # Show all managed items
plonk status --packages                           # Show packages only
plonk status --dotfiles                           # Show dotfiles only
plonk doctor                                      # System health check
plonk doctor --fix                                # Fix common issues
```

> **📖 Complete Command Reference**: See [docs/CLI.md](docs/CLI.md) for comprehensive command documentation with examples and options.

## Output Formats

All commands support structured output for AI agents:

```bash
plonk status --output json
plonk status --output yaml
plonk status --output table  # default
```

## Environment Variables

- `PLONK_DIR` - Config directory (default: `~/.config/plonk`)
- `EDITOR` - Editor for `plonk config edit`

## Testing

Plonk uses two testing approaches:

1. **Go Unit Tests**: Test internal logic and functions
   ```bash
   go test ./...
   ```

2. **BATS Behavioral Tests**: Test CLI behavior and user experience
   ```bash
   just test-bats
   ```

⚠️ **WARNING**: BATS tests install real packages and create real files. See `tests/bats/README.md` for details.

## Documentation

- **[CLI Reference](docs/CLI.md)** - Complete command documentation
- **[Configuration Guide](docs/CONFIGURATION.md)** - Configuration file format
- **[Architecture](ARCHITECTURE.md)** - Technical architecture and extension guide
- **[Development](docs/DEVELOPMENT.md)** - Contributing and development setup

## System Requirements

- **Go 1.24.4+** (for building)
- **Just** (command runner)
- **Git** (version management)
- **Package managers**: Homebrew (macOS), NPM, Cargo (Rust), Pip (Python), Gem (Ruby), APT (Linux), Go Install
- **Platform support**: macOS, Linux, Windows
- **Architecture support**: AMD64, ARM64

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for development setup and contributing guidelines.

### Development

Key automation features for contributors:

```bash
# One-time setup for new developers
just dev-setup

# Update all dependencies safely
just deps-update

# Complete cleanup for troubleshooting
just clean-all

# Create a release
just release-auto v1.2.3
```

See [Development Guide](docs/DEVELOPMENT.md) for complete details.
