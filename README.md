# Plonk

> ⚠️ **WARNING: This project is under active development.** APIs, commands, and configuration formats may change without notice. Use at your own risk in production environments.

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)
[![Security](https://github.com/richhaase/plonk/workflows/Security%20Check/badge.svg)](https://github.com/richhaase/plonk/actions)
[![codecov](https://codecov.io/gh/richhaase/plonk/branch/main/graph/badge.svg)](https://codecov.io/gh/richhaase/plonk)

A unified package and dotfile manager that simplifies developer environment setup and maintenance across multiple machines.

## What is Plonk?

Plonk is a development environment manager that unifies package installation and dotfile management. It enables one-command setup of new development machines by tracking your packages and configuration files in a simple YAML format.

**Why Plonk?**
- **One tool, everything managed**: No more separate dotfile managers and package lists
- **True zero-config**: Works immediately without any setup files
- **Smart state reconciliation**: Knows what's missing and only installs what's needed
- **Cross-machine consistency**: Keep multiple development environments in sync
- **AI-friendly operations**: Structured outputs and predictable command patterns

**Key features:**
- **Unified management**: Handles packages (Homebrew, NPM, Cargo, Pip, Gem, Go) and dotfiles together
- **Automatic tracking**: `plonk install` both installs and tracks packages in one step
- **State-based approach**: Compares desired vs actual state, only changes what's needed
- **Simple commands**: Unix-style `add`, `rm`, `status`, `apply` that just work
- **Cross-platform**: Supports macOS, Linux, and Windows with appropriate package managers

## Quick Start

### Installation

```bash
go install github.com/richhaase/plonk/cmd/plonk@latest
```

**Requirements:**
- Go 1.24.4 or later
- One or more package managers (Homebrew, NPM, etc.)

### Basic Usage

```bash
# Install and track packages in one command
plonk install ripgrep fd bat              # Install with default manager (Homebrew)
plonk install npm:prettier cargo:exa      # Install with specific managers

# Manage your dotfiles
plonk add ~/.vimrc ~/.zshrc              # Start tracking dotfiles
plonk add ~/.config/nvim/                # Track entire directories

# Check what's being managed
plonk status                             # See all packages and dotfiles
plonk                                    # Same as status (like git)

# Set up a new machine
plonk apply                              # Install missing packages, deploy dotfiles
```

## Setting Up a New Machine

The fastest way to set up a new development machine:

```bash
# Clone existing dotfiles and set up environment
plonk setup github-user/dotfiles         # Uses GitHub shorthand
plonk setup https://github.com/user/dotfiles.git

# Or start fresh
plonk setup                              # Initialize plonk, install missing tools
```

The `setup` command:
1. Clones your dotfiles repository (if provided)
2. Installs missing package managers (Homebrew, Cargo, etc.)
3. Runs `plonk apply` to install all packages and deploy dotfiles
4. Gets your machine ready for development in minutes

## Key Commands

```bash
# Package management
plonk install ripgrep fd              # Install and track packages
plonk uninstall ripgrep               # Uninstall and stop tracking
plonk search ripgrep                  # Search across all package managers
plonk info ripgrep                    # Show package details

# Dotfile management
plonk add ~/.vimrc ~/.zshrc           # Start tracking dotfiles
plonk rm ~/.vimrc                     # Stop tracking (doesn't delete file)

# System state
plonk status                          # Show all managed items
plonk apply                           # Sync system to desired state
plonk doctor                          # Check system health

# Configuration
plonk config show                     # View current settings
plonk config edit                     # Edit configuration
```

## Supported Package Managers

Plonk supports the following package managers:
- **Homebrew** (brew) - macOS/Linux packages
- **NPM** (npm) - Node.js packages (global)
- **Cargo** (cargo) - Rust packages
- **Pip** (pip) - Python packages
- **Gem** (gem) - Ruby packages
- **Go** (go) - Go packages via `go install`

Package manager prefixes in commands:
```bash
plonk install brew:wget npm:prettier cargo:ripgrep
plonk install pip:black gem:rubocop go:golangci-lint
```

## Configuration

Plonk stores its data in `~/.config/plonk/`:
- **`plonk.lock`** - Automatically maintained list of packages and dotfiles
- **`plonk.yaml`** - Optional configuration (only create if needed)

The lock file is automatically updated when you:
- Install/uninstall packages via plonk
- Add/remove dotfiles from management

Share your `plonk.lock` file (e.g., in a dotfiles repo) to replicate your environment on other machines.

## Output Formats

All commands support multiple output formats:

```bash
plonk status --output json       # JSON output for scripts/tools
plonk status --output yaml       # YAML output
plonk status --output table      # Human-readable table (default)
```

## Development

For contributors and developers:

```bash
# Clone and set up development environment
git clone https://github.com/richhaase/plonk
cd plonk
just dev-setup          # Install dependencies and tools

# Run tests
go test ./...           # Unit tests
just test-bats          # Behavioral tests (installs real packages!)

# Build and install locally
just install
```

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - Technical design and implementation details
- **[Why Plonk?](docs/why-plonk.md)** - Project motivation and goals

## Requirements

- Go 1.24.4+ (for installation)
- Git (for setup with repositories)
- At least one supported package manager
- macOS, Linux, or Windows

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

See the codebase for examples and patterns.
