# Plonk - Shell Environment Lifecycle Manager

## Project Overview

Plonk is a CLI tool for managing shell environments across multiple machines. It helps you manage package installations and environment switching using a focused set of package managers:

- **Homebrew** - Primary package installation
- **ASDF** - Programming language tools and versions
- **NPM** - Packages not available via Homebrew (like claude-code)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and workflow requirements.

## Architecture

### Package Managers (`pkg/managers/`)
- **CommandExecutor Interface** - Abstraction for command execution (supports dependency injection)
- **CommandRunner** - Shared command execution logic to eliminate code duplication
- **Individual Managers** - Homebrew, ASDF, NPM with consistent interfaces

### Configuration Management (`pkg/config/`)
- **YAML Configuration** - Pure YAML config format (TOML support removed in cleanup)
- **Package-Centric Structure** - Organized by package manager with dotfiles section
- **Source-Target Convention** - Automatic path mapping (config/nvim/ → ~/.config/nvim/)
- **Local Overrides** - Support for plonk.local.yaml for machine-specific settings
- **Config Validation** - Comprehensive YAML syntax and content validation

### CLI (`internal/commands/`)
- **Cobra Framework** - Professional CLI with help, autocompletion, and subcommands
- **Status Command** - Shows availability and package counts for all managers
- **Pkg Command** - Modular package listing with `plonk pkg list [manager]` structure
- **Git Operations** - Clone and pull commands with configurable locations
- **Package Management** - Install command with automatic config application
- **Configuration Deployment** - Apply command for dotfiles and package configs with --backup and --dry-run support
- **Foundational Setup** - Setup command for installing core tools (Homebrew/ASDF/NPM)
- **Convenience Commands** - Repository-based workflow and direct repo syntax

### Testing
- **Comprehensive Test Coverage** - All components tested with TDD approach
- **Mock Command Executor** - Enables testing without actual command execution
- **Interface Compliance Tests** - Ensures consistent behavior across managers
- **Config Validation Tests** - YAML/TOML parsing and validation coverage

## File Structure

```
plonk/
├── cmd/plonk/main.go              # CLI entry point
├── internal/commands/             # CLI commands
│   ├── root.go                   # Root command definition
│   ├── status.go                 # Status command implementation
│   ├── pkg.go                    # Package listing commands
│   ├── clone.go                  # Clone command with git operations
│   ├── pull.go                   # Pull command for updates
│   ├── install.go                # Install command for packages from config
│   ├── apply.go                  # Apply command for configuration deployment
│   ├── setup.go                  # Setup command for foundational tools
│   ├── repo.go                   # Repository convenience command
│   ├── test_helpers.go          # Shared testing utilities
│   ├── status_test.go           # Status command tests
│   ├── clone_test.go            # Clone command tests
│   ├── pull_test.go             # Pull command tests
│   ├── install_test.go          # Install command tests
│   ├── apply_test.go            # Apply command tests
│   ├── setup_test.go            # Setup command tests
│   ├── repo_test.go             # Repository convenience command tests
│   ├── backup.go                # Backup functionality with configurable location
│   └── backup_test.go           # Backup functionality tests
├── pkg/managers/                 # Package manager implementations
│   ├── common.go                 # CommandExecutor interface & CommandRunner
│   ├── executor.go               # Real command execution for production
│   ├── homebrew.go              # Homebrew package manager
│   ├── asdf.go                  # ASDF tool manager
│   ├── npm.go                   # NPM global package manager
│   └── manager_test.go          # Comprehensive test suite
├── pkg/config/                   # Configuration management
│   ├── config.go                 # Legacy TOML config support
│   ├── config_test.go           # TOML config tests
│   ├── yaml_config.go           # Primary YAML config implementation
│   ├── yaml_config_test.go      # YAML config tests
│   ├── zsh_generator.go         # ZSH configuration file generation
│   └── zsh_generator_test.go    # ZSH generator tests
├── go.mod                       # Go module definition
└── CLAUDE.md                    # This documentation
```

## Usage

### Build and Install
```bash
go build ./cmd/plonk
```

### Commands
```bash
./plonk --help                   # Show main help
./plonk status                   # Package manager availability and counts
./plonk pkg list                 # List packages from all managers
./plonk pkg list brew            # List only Homebrew packages
./plonk pkg list asdf            # List only ASDF tools
./plonk pkg list npm             # List only NPM packages

# Foundational setup
./plonk setup                    # Install Homebrew, ASDF, and Node.js/NPM

# Git operations
./plonk clone <repo>             # Clone dotfiles repository
./plonk pull                     # Pull updates to existing repository

# Package and configuration management
./plonk install                  # Install packages from config
./plonk apply                    # Apply all configuration files
./plonk apply <package>          # Apply configuration for specific package
./plonk apply --backup           # Apply all configurations with backup
./plonk apply --dry-run          # Show what would be applied without making changes
./plonk apply --backup --dry-run # Preview what would be applied with backup

# Backup operations
./plonk backup                   # Backup all files that apply would overwrite
./plonk backup ~/.zshrc ~/.vimrc # Backup specific files

# Convenience commands
./plonk repo <repo>              # Complete setup: clone + install + apply
./plonk <repo>                   # Same as above (convenience syntax)

# Environment variable
PLONK_DIR=~/my-dotfiles ./plonk clone <repo>  # Clone to custom location
```

### Example Output
```
Package Manager Status
=====================

## Homebrew
✅ Available
📦 139 packages installed

## ASDF
✅ Available
📦 8 packages installed

## NPM
✅ Available
📦 6 packages installed
```

## Configuration Format

The new YAML-based configuration supports both simple and complex package definitions:

```yaml
settings:
  default_manager: homebrew

# Standalone config files (no package install needed)
dotfiles:
  - zshrc                    # -> ~/.zshrc
  - zshenv                   # -> ~/.zshenv
  - plugins.zsh              # -> ~/.plugins.zsh
  - dot_gitconfig            # -> ~/.gitconfig

homebrew:
  brews:
    - aichat                 # Simple package
    - aider
    - name: neovim           # Package with config
      config: config/nvim/   # -> ~/.config/nvim/
    - name: mcfly
      config: config/mcfly/  # -> ~/.config/mcfly/
  
  casks:
    - font-hack-nerd-font
    - google-cloud-sdk

asdf:
  - name: nodejs
    version: "24.2.0"
    config: config/npm/      # -> ~/.config/npm/
  - name: python
    version: "3.13.2"
  - name: golang
    version: "1.24.4"

npm:
  - "@anthropic-ai/claude-code"
  - name: some-tool
    package: "@scope/different-name"
```


## Development Timeline

- **TDD Approach**: Consistent Red-Green-Refactor cycles throughout
- **Package Manager Abstraction**: Built reusable patterns for easy extension
- **CLI Implementation**: Professional-grade command interface with Cobra
- **Focused Scope**: Refined to essential package managers for shell environment management

## Key Design Decisions

1. **Go Over Rust/Python** - Better balance of simplicity and power for CLI tools
2. **Test-Driven Development** - Ensures reliability and maintainability
3. **CommandRunner Abstraction** - Eliminates code duplication across managers
4. **Interface-Based Design** - Easy to add new package managers or mock for testing
5. **Focused Package Managers** - Homebrew + ASDF + NPM covers most shell environment needs
6. **Cobra CLI Framework** - Professional CLI with built-in help, completion, and extensibility

## Technical Highlights

- **Dependency Injection** - CommandExecutor interface enables testing without side effects
- **Output Parsing** - Handles different package manager output formats correctly
- **Error Handling** - Graceful degradation when package managers are unavailable
- **Scoped Package Support** - Correctly handles NPM scoped packages (@vue/cli)
- **Version Management** - ASDF integration for language tool versioning
- **Global Package Focus** - Avoids local/project-specific package management complexity
- **Git Operations** - Pure Go git operations with mockable interface for testing
- **Configuration Management** - Automatic application of package-specific configurations
- **Foundational Setup** - Automated installation of prerequisite tools
- **Intelligent Workflows** - Smart detection of existing repositories and installed packages
- **ZSH Configuration Generation** - Complete .zshrc and .zshenv file generation from YAML config
- **Shell Best Practices** - Proper separation of environment variables (.zshenv) and interactive config (.zshrc)
- **Automated Backup System** - Smart backup detection with --backup flag and standalone backup command
- **Configurable Backup Management** - Timestamped backups with automatic cleanup and retention policies
