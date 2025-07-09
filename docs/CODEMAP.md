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
- `plonk.yaml` - Main configuration (location: `$PLONK_DIR` or `~/.config/plonk`)
- Auto-discovered dotfiles from config directory
- Configurable ignore patterns for dotfile discovery

### Quick Navigation for AI Agents

#### Adding a Package Manager
1. **Interface location:** `internal/managers/common.go:PackageManager`
2. **Implementation template:** `internal/managers/homebrew.go`
3. **Complete interface specification:** `docs/api/managers.md`
4. **Registration:** Add to command layer in `internal/commands/`

#### Adding a Command
1. **Command template:** `internal/commands/status.go`
2. **Registration:** `internal/commands/root.go:rootCmd.AddCommand()`
3. **Output formatting:** Use `internal/commands/output.go`
4. **Error handling:** Use `internal/errors/types.go:PlonkError`
5. **Command API details:** `docs/api/commands.md`

#### Understanding State Reconciliation
1. **Entry point:** `internal/state/reconciler.go:GetState()`
2. **Provider interface:** `internal/state/reconciler.go:Provider`
3. **State types:** `internal/state/types.go:ItemState`
4. **Package provider:** `internal/state/package_provider.go`
5. **Dotfile provider:** `internal/state/dotfile_provider.go`
6. **Complete API specification:** `docs/api/state.md`

#### Configuration Management
1. **Interface definitions:** `internal/config/interfaces.go`
2. **YAML implementation:** `internal/config/yaml_config.go`
3. **Validation:** `internal/config/simple_validator.go`
4. **Complete API specification:** `docs/api/config.md`

#### Error Handling Patterns
1. **Error types:** `internal/errors/types.go:PlonkError`
2. **Error codes:** `internal/errors/types.go:ErrorCode`
3. **Error domains:** `internal/errors/types.go:Domain`
4. **Complete API specification:** `docs/api/errors.md`