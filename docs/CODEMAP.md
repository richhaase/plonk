# Plonk Codemap

## Overview
A quick reference guide to navigate the Plonk codebase - a unified package and dotfile manager.

## Project Structure

### Entry Points
- `cmd/plonk/main.go` - CLI entry point
- `internal/commands/root.go` - Root command and CLI structure

### Core Components

#### Commands Layer (`internal/commands/`)
- `root.go` - Root command setup
- `status.go` - Show system state
- `apply.go` - Unified apply command (packages and dotfiles)
- `env.go` - Show environment information for debugging
- `doctor.go` - Comprehensive health checks and diagnostics
- `search.go` - Intelligent package search across package managers
- `info.go` - Detailed package information display
- `pkg_*.go` - Package management commands (add, remove, list)
- `dot_*.go` - Dotfile management commands (add, list, re-add)
- `config.go` - Configuration base command
- `config_show.go` - Display configuration content
- `config_validate.go` - Validate configuration syntax and structure
- `config_edit.go` - Edit configuration in preferred editor
- `output.go` - Output formatting utilities
- `errors.go` - Command-level error handling

#### Configuration (`internal/config/`)
- `interfaces.go` - Core configuration interfaces
- `yaml_config.go` - YAML implementation
- `adapters.go` - Bridge between config and state
- `simple_validator.go` - Configuration validation

#### State Management (`internal/state/`)
- `reconciler.go` - Core reconciliation engine
- `types.go` - State types (Managed, Missing, Untracked)
- `package_provider.go` - Package state provider
- `dotfile_provider.go` - Dotfile state provider
- `adapters.go` - State-config adapters

#### Package Managers (`internal/managers/`)
- `homebrew.go` - Homebrew/Cask implementation
- `npm.go` - NPM global packages
- `common.go` - Shared manager utilities

#### Dotfile Operations (`internal/dotfiles/`)
- `operations.go` - Core dotfile manager
- `fileops.go` - File operations (copy, backup)
- `atomic.go` - Atomic file operations

#### Error Handling (`internal/errors/`)
- `types.go` - Structured error types

### Testing
- `test/integration/` - Integration test suite
- `*_test.go` files - Unit tests alongside implementations

### Key Interfaces

#### Configuration
- `ConfigReader` - Load configuration
- `ConfigWriter` - Save configuration  
- `ConfigValidator` - Validate config
- `DotfileConfigReader` - Dotfile-specific config
- `PackageConfigReader` - Package-specific config

#### State Management
- `Provider` - State provider interface
- `PackageManager` - Package manager interface

### Data Flow
1. Commands → State Reconciler → Providers
2. Providers → Configuration & System State
3. Reconciler → Package Managers & Dotfile Operations

### Configuration Files
- `plonk.yaml` - Main configuration
- `plonk.local.yaml` - Local overrides (gitignored)

### Quick Navigation

#### Adding a Package Manager
Start at: `internal/managers/` → implement `PackageManager` interface

#### Adding a Command
Start at: `internal/commands/` → create command file → register in `root.go`

#### Understanding State Reconciliation
Start at: `internal/state/reconciler.go` → follow to providers

#### Configuration Format
Start at: `internal/config/yaml_config.go` → see example structures

#### Error Handling
Start at: `internal/errors/types.go` → see structured error approach