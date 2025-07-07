# Plonk - Shell Environment Lifecycle Manager

## Project Overview

Plonk is a CLI tool for managing shell environments across multiple machines. It helps you manage package installations and environment switching using a focused set of package managers:

- **Homebrew** - Primary package installation
- **ASDF** - Programming language tools and versions
- **NPM** - Packages not available via Homebrew (like claude-code)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and workflow requirements.

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for system design and technical details.


## Usage

### Quick Installation
```bash
git clone <repository-url>
cd plonk
go install ./cmd/plonk
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development setup.

### Commands
```bash
plonk --help                     # Show main help
plonk status                     # Package manager availability and counts
plonk pkg list                   # List packages from all managers
plonk pkg list brew              # List only Homebrew packages
plonk pkg list asdf              # List only ASDF tools
plonk pkg list npm               # List only NPM packages

# Foundational setup
plonk setup                      # Install Homebrew, ASDF, and Node.js/NPM

# Git operations
plonk clone <repo>               # Clone dotfiles repository
plonk pull                       # Pull updates to existing repository

# Import existing environment
plonk import                     # Generate plonk.yaml from current environment

# Package and configuration management
plonk install                    # Install packages from config
plonk apply                      # Apply all configuration files
plonk apply <package>            # Apply configuration for specific package
plonk apply --backup             # Apply all configurations with backup
plonk apply --dry-run            # Show what would be applied without making changes
plonk apply --backup --dry-run   # Preview what would be applied with backup

# Backup operations
plonk backup                     # Backup all files that apply would overwrite
plonk backup ~/.zshrc ~/.vimrc   # Backup specific files

# Convenience commands
plonk repo <repo>                # Complete setup: clone + install + apply
plonk <repo>                     # Same as above (convenience syntax)

# Environment variable
PLONK_DIR=~/my-dotfiles plonk clone <repo>    # Clone to custom location
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

