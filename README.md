# Plonk

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)
[![Security](https://github.com/richhaase/plonk/workflows/Security%20Check/badge.svg)](https://github.com/richhaase/plonk/actions)
[![codecov](https://codecov.io/gh/richhaase/plonk/branch/main/graph/badge.svg)](https://codecov.io/gh/richhaase/plonk)

A unified package and dotfile manager for developers that maintains consistency across multiple machines.

## What is Plonk?

Plonk manages your development environment by tracking packages and dotfiles automatically. It uses state reconciliation to compare your desired state with your actual system state and applies the necessary changes.

**Key features:**
- **Unified management**: Packages (Homebrew, NPM, Cargo) and dotfiles tracked automatically
- **State reconciliation**: Automatically detects and applies missing configurations
- **Auto-discovery**: Finds dotfiles automatically with configurable ignore patterns
- **Directory expansion**: Smart expansion of configured directories in dot list output
- **AI-friendly**: Structured output formats and clear command syntax
- **Cross-platform**: Works on macOS, Linux, and Windows

## Quick Start

### Installation

Currently, plonk must be built from source:

```bash
git clone https://github.com/richhaase/plonk
cd plonk
just install
```

**Requirements:**
- Go 1.24.4+
- Just (command runner)
- Git

### Basic Usage

1. **Initialize configuration:**
```bash
plonk config edit
```

2. **Add your first package:**
```bash
plonk pkg add git
```

3. **Apply configuration:**
```bash
plonk apply
```

4. **Check status:**
```bash
plonk status
```

## Configuration

Plonk uses two files to manage your environment:

### Configuration File (`~/.config/plonk/plonk.yaml`)
Contains your settings and preferences:

```yaml
settings:
  default_manager: homebrew
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
```

### Lock File (`~/.config/plonk/plonk.lock`)
Automatically managed file that tracks your packages:

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

**Dotfiles are auto-discovered** from your config directory:
- `~/.config/plonk/zshrc` → `~/.zshrc`
- `~/.config/plonk/config/nvim/` → `~/.config/nvim/`

## Common Commands

```bash
# Show overall status
plonk status

# Apply all configuration
plonk apply --dry-run  # Preview changes
plonk apply           # Apply changes

# Package management
plonk pkg list                    # List managed + missing + untracked count
plonk pkg list --verbose          # Show all packages including untracked
plonk pkg list --manager homebrew # Filter by package manager
plonk pkg add htop               # Add package to lock file and install
plonk pkg add htop --dry-run     # Preview what would be added
plonk pkg remove htop            # Remove from lock file only
plonk pkg remove htop --dry-run  # Preview what would be removed
plonk pkg remove htop --uninstall # Remove from lock file and uninstall
plonk search git                 # Search for packages

# Dotfile management
plonk dot list           # List dotfiles (missing + managed + untracked count)
plonk dot list --verbose # Show all files including full untracked list
plonk dot add .vimrc     # Add dotfile (flexible path resolution)
plonk dot add ~/.config/nvim/init.lua  # Explicit path
plonk dot add init.lua   # Finds ./init.lua or ~/init.lua

# Configuration
plonk config show     # Show current config
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

### Releasing

```bash
# Get version suggestions
just release-version-suggest

# Create automated release
just release-auto v1.2.3
```