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

**🎉 No setup required!** Plonk works immediately with zero configuration:

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

**Configuration is completely optional!** Plonk works with sensible defaults out of the box.

When you do want to customize settings, Plonk uses two files:

### Configuration File (`~/.config/plonk/plonk.yaml`) - Optional
Contains your custom settings and preferences (create with `plonk init`):

```yaml
default_manager: homebrew
operation_timeout: 300   # 5 minutes
package_timeout: 180     # 3 minutes
dotfile_timeout: 60      # 1 minute

expand_directories:
  - .config
  - .ssh
  - .aws
  - .kube
  - .docker
  - .gnupg
  - .local

ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"
  - "*.tmp"
  - "*.swp"
  - plonk.lock
```

### Lock File (`~/.config/plonk/plonk.lock`) - Automatic
Automatically created and managed file that tracks your packages (similar to package-lock.json):

```yaml
version: 1
packages:
  homebrew:
    - name: git
      installed_at: "2024-01-15T10:30:00Z"
      version: "2.43.0"
    - name: neovim
      installed_at: "2024-01-15T10:31:00Z"
      version: "0.9.5"
  npm:
    - name: typescript
      installed_at: "2024-01-15T10:32:00Z"
      version: "5.3.3"
  cargo:
    - name: ripgrep
      installed_at: "2024-01-15T10:35:00Z"
      version: "14.1.0"
```

**Note:** The lock file is automatically created and updated when you add/remove packages. You don't need to edit it manually.

### Zero-Config Defaults

When no configuration file exists, Plonk uses these sensible defaults:

```yaml
# Default settings (you can override these with `plonk init`)
default_manager: homebrew      # Primary package manager
operation_timeout: 300         # 5 minutes for overall operations
package_timeout: 180           # 3 minutes for package operations
dotfile_timeout: 60            # 1 minute for dotfile operations
expand_directories:            # Directories shown expanded in lists
  - .config
  - .ssh
  - .aws
  - .kube
  - .docker
  - .gnupg
  - .local

ignore_patterns:                 # Files ignored during dotfile discovery
  - .DS_Store
  - .git
  - "*.backup"
  - "*.tmp"
  - "*.swp"
```

**Dotfiles are auto-discovered** from your config directory:
- `~/.config/plonk/zshrc` → `~/.zshrc`
- `~/.config/plonk/config/nvim/` → `~/.config/nvim/`

## Getting Started Guide

### First Time Setup (Zero Configuration!)

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
# Show overall status
plonk status

# Apply all configuration
plonk apply --dry-run  # Preview changes
plonk apply           # Apply changes
plonk apply --backup  # Apply with backups of existing files

# Package management
plonk pkg list                    # List managed + missing + untracked count
plonk pkg list --verbose          # Show all packages including untracked
plonk pkg list --manager homebrew # Filter by package manager
plonk pkg add htop               # Add single package to lock file and install
plonk pkg add git neovim ripgrep htop  # Add multiple packages at once
plonk pkg add --manager npm typescript prettier  # Multiple packages with specific manager
plonk pkg add htop --dry-run     # Preview what would be added
plonk pkg remove htop            # Remove from lock file only
plonk pkg remove htop --dry-run  # Preview what would be removed
plonk pkg remove htop --uninstall # Remove from lock file and uninstall
plonk search git                 # Search for packages

# Dotfile management
plonk dot list           # List dotfiles (missing + managed + untracked count)
plonk dot list --verbose # Show all files including full untracked list
plonk dot add .vimrc     # Add single dotfile (flexible path resolution)
plonk dot add ~/.vimrc ~/.zshrc ~/.gitconfig  # Add multiple dotfiles at once
plonk dot add ~/.config/nvim/ ~/.tmux.conf    # Mix directories and files
plonk dot add ~/.config/nvim/init.lua  # Explicit path
plonk dot add init.lua   # Finds ./init.lua or ~/init.lua
plonk dot add --dry-run ~/.vimrc ~/.zshrc  # Preview multiple dotfile additions

# Configuration (optional - works without any config!)
plonk init            # Create config template with defaults
plonk config show     # Show effective config (defaults merged with user settings)
plonk config validate # Validate config syntax
plonk config edit     # Edit config file

# Diagnostics
plonk doctor          # Health check
plonk env             # Environment info
```

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
