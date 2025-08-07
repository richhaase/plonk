# Plonk

[![CI](https://github.com/richhaase/plonk/workflows/CI/badge.svg)](https://github.com/richhaase/plonk/actions)
[![Security](https://github.com/richhaase/plonk/workflows/Security%20Check/badge.svg)](https://github.com/richhaase/plonk/actions)

**The unified package and dotfile manager for developers who tinker.** One command to set up your entire development environment.

```bash
# Clone your dotfiles and set up everything
plonk clone user/dotfiles
# Done. Seriously.
```

## What is Plonk?

Plonk is the missing link between dotfile managers and package managers. Born from the frustration of maintaining separate tools for configuration files and installed packages, plonk unifies both in a single, simple tool.

**Why another dotfile manager?** [→ Read the full story](docs/why-plonk.md)

After trying bash scripts, symlink farms, [dotter](https://github.com/SuperCuber/dotter), and [chezmoi](https://www.chezmoi.io/), I wanted something that:
- Manages packages as a first-class concern alongside dotfiles
- Has zero configuration complexity (no templates, no YAML manifestos)
- Adapts quickly to my constantly changing toolset
- Just works

**Key innovations:**
- **Package Manager Manager™**: One interface for 10 package managers (brew, npm, cargo, pip, gem, go, uv, pixi, composer, dotnet)
- **Filesystem as truth**: Your dotfiles directory IS the state - no sync issues
- **Copy, don't symlink**: Cleaner, simpler, and more compatible
- **State-based**: Track what should exist, not what commands were run
- **Drift detection**: Know when deployed dotfiles have been modified (`plonk diff`)
- **AI-friendly**: Built with and for AI coding assistants

**For developers who:**
- Set up new machines/VMs regularly
- Experiment with new CLI/TUI tools constantly
- Want their environment manager to keep up with their tinkering
- Value simplicity over features

## Core Philosophy

**Just works** - Zero configuration required
**Unified** - Packages and dotfiles together
**Simple** - Your filesystem IS the state
**Fast** - One command from fresh OS to ready

## Quick Start

### Prerequisites

1. **Install Homebrew** (if not already installed):
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

2. **Install plonk**:
   ```bash
   # Via Homebrew
   brew install richhaase/tap/plonk

   # Via Go (requires Go 1.23+)
   go install github.com/richhaase/plonk/cmd/plonk@latest
   ```

**Requirements:** Homebrew, Git

**📖 [Complete Installation Guide →](docs/installation.md)**

### Basic Usage

```bash
# Track your existing setup
plonk add ~/.zshrc ~/.vimrc ~/.config/nvim/    # Add dotfiles
plonk install ripgrep fd bat                   # Install & track packages

# See what plonk manages
plonk status                                   # Show all resources

# Replicate on a new machine
plonk clone your-github/dotfiles              # Clone and apply everything
```

The beauty is in what you don't need to do:
- No configuration files to write
- No symlinks to manage
- No separate package lists to maintain
- No complex templating languages to learn

## Setting Up a New Machine

The fastest way to set up a new development machine:

```bash
# Clone existing dotfiles and set up environment
plonk clone user/dotfiles                # GitHub shorthand
plonk clone https://github.com/user/dotfiles.git

# Or start fresh - just start using plonk
plonk add ~/.zshrc                       # Start tracking your dotfiles
```

The `clone` command:
1. Clones your dotfiles repository
2. Installs missing language package managers needed by your dotfiles
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
plonk status                          # Show all managed items (including drift)
plonk apply                           # Sync system to desired state
plonk doctor                          # Check system health

# Configuration
plonk config show                     # View current settings
plonk config edit                     # Edit configuration
```

## Supported Package Managers

Plonk supports 10 package managers across multiple language ecosystems:

### Core Package Managers
- **Homebrew** (brew) - macOS/Linux packages and system tools
- **NPM** (npm) - Node.js packages (global)
- **Cargo** (cargo) - Rust packages
- **Pip** (pip) - Python packages (user installs)
- **Gem** (gem) - Ruby packages
- **Go** (go) - Go packages via `go install`

### Extended Language Support
- **UV** (uv) - Fast Python tool manager with isolated environments
- **Pixi** (pixi) - Cross-platform package manager using conda-forge ecosystem
- **Composer** (composer) - PHP global packages and CLI tools
- **.NET Global Tools** (dotnet) - .NET CLI tools and utilities

Package manager prefixes in commands:
```bash
# Core managers
plonk install brew:wget npm:prettier cargo:ripgrep
plonk install pip:black gem:rubocop go:golangci-lint

# Extended support
plonk install uv:ruff pixi:jupyter composer:php-cs-fixer
plonk install dotnet:dotnetsay
```

## Configuration

Plonk stores its data in `~/.config/plonk/`:
- **`plonk.lock`** - Automatically maintained list of packages
- **`plonk.yaml`** - Optional configuration (only create if needed)
- **Dotfiles** - Stored directly in the config directory (e.g., `zshrc`, `vimrc`)

The lock file is automatically updated when you:
- Install/uninstall packages via plonk

Dotfiles are managed by the filesystem itself - files in `$PLONK_DIR` are your tracked dotfiles

Share your `plonk.lock` file (e.g., in a dotfiles repo) to replicate your environment on other machines.

## Output Formats

All commands support multiple output formats:

```bash
plonk status --output json       # JSON output for scripts/tools
plonk status --output yaml       # YAML output
plonk status --output table      # Human-readable table (default)
```

## Output

Plonk uses minimal colorization for status indicators:
- **Green**: Success/managed/available states
- **Red**: Error/missing/failed states
- **Yellow**: Warning/unmanaged states
- **Blue**: Informational annotations

Respects the standard `NO_COLOR` environment variable for color-free output.

## Development

For contributors and developers:

```bash
# Clone and set up development environment
git clone https://github.com/richhaase/plonk
cd plonk
just dev-setup          # Install dependencies and tools

# Run tests
go test ./...           # Unit tests

# Build and install locally
just install
```

## Documentation

### Core Documentation
- **[Why Plonk?](docs/why-plonk.md)** - The journey that led to plonk and what makes it different
- **[Architecture](docs/architecture.md)** - Technical design, state model, and implementation details

### Command Documentation
- **[Clone](docs/cmds/clone.md)** - Clone and set up existing dotfiles
- **[Apply](docs/cmds/apply.md)** - Sync your system to desired state
- **[Status](docs/cmds/status.md)** - View managed packages and dotfiles
- **[Package Management](docs/cmds/package-management.md)** - install, uninstall, search, info
- **[Dotfile Management](docs/cmds/dotfile-management.md)** - add, rm
- **[Config](docs/cmds/config.md)** - Manage plonk configuration
- **[Doctor](docs/cmds/doctor.md)** - Check system health

## Requirements

- Homebrew (required prerequisite)
- Git (for cloning repositories)
- Go 1.23+ (only if installing via `go install`)
- macOS, Linux, or Windows (WSL)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

See the codebase for examples and patterns.
