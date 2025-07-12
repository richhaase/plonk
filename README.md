# Plonk

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)
[![Security](https://github.com/richhaase/plonk/workflows/Security%20Check/badge.svg)](https://github.com/richhaase/plonk/actions)
[![codecov](https://codecov.io/gh/richhaase/plonk/branch/main/graph/badge.svg)](https://codecov.io/gh/richhaase/plonk)

A unified package and dotfile manager for developers that maintains consistency across multiple machines.

## What is Plonk?

Plonk manages your development environment by tracking packages and dotfiles automatically. It uses state reconciliation to compare your desired state with your actual system state and applies the necessary changes.

**🚀 Zero Configuration Required** - Start using Plonk immediately without any setup!

**Key features:**
- **Zero-config**: Works immediately with sensible defaults - no setup required
- **Unified management**: Packages (Homebrew, NPM, Cargo) and dotfiles tracked automatically
- **State reconciliation**: Automatically detects and applies missing configurations
- **Auto-discovery**: Finds dotfiles automatically with configurable ignore patterns
- **Optional customization**: Create configuration only when you need to customize defaults
- **AI-friendly**: Structured output formats and clear command syntax
- **Cross-platform**: Works on macOS, Linux, and Windows

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

1. **Check your current environment:**
```bash
plonk status
```

2. **Add your first package:**
```bash
plonk pkg add git
```

3. **Apply configuration:**
```bash
plonk apply
```

4. **Get system health check:**
```bash
plonk doctor
```

**Optional:** Create a configuration file if you want to customize settings:
```bash
plonk init    # Creates helpful config template
```

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

### First Time Setup (Zero-Config!)

1. **Install Plonk** (see Installation section above)

2. **Check what's already on your system:**
```bash
plonk status
# Shows:
# - Untracked packages (already installed but not managed by Plonk)
# - Auto-discovered dotfiles
# - Overall system health
```

3. **Start managing packages:**
```bash
# Add multiple packages at once
plonk pkg add git neovim ripgrep htop

# Add packages with specific manager
plonk pkg add --manager npm typescript prettier eslint

# Or discover and add untracked packages
plonk pkg list --untracked    # See what's installed but not tracked
plonk apply --add-untracked   # Add them to Plonk management
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
plonk init           # Creates config template with helpful comments
plonk config show    # Show effective config (defaults merged with user settings)
plonk config edit    # Edit configuration file
```

**Note:** `plonk config show` always displays the complete effective configuration that plonk is using, including all defaults merged with any user overrides.

### Daily Workflow

```bash
plonk status         # Check what needs attention
plonk apply          # Install missing packages, sync dotfiles
plonk pkg add <name> # Add new packages
plonk doctor         # Health check when something seems wrong
```

## Common Commands

```bash
# Essential workflows
plonk status                                      # Check system state
plonk apply                                       # Apply all changes
plonk pkg add git neovim ripgrep                 # Add multiple packages
plonk dot add ~/.vimrc ~/.zshrc ~/.gitconfig     # Add multiple dotfiles
plonk doctor                                      # Health check
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

## Documentation

- **[CLI Reference](docs/CLI.md)** - Complete command documentation
- **[Configuration Guide](docs/CONFIGURATION.md)** - Configuration file format
- **[Architecture](docs/ARCHITECTURE.md)** - Technical architecture
- **[Development](docs/DEVELOPMENT.md)** - Contributing and development setup

## System Requirements

- **Go 1.24.4+** (for building)
- **Just** (command runner)
- **Git** (version management)
- **Package managers**: Homebrew (macOS), NPM (optional)
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
