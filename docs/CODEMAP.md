# Plonk Codemap

## Overview
A quick reference guide to navigate the Plonk codebase - a unified package and dotfile manager.

## Project Structure

### Entry Points
- `cmd/plonk/main.go` - CLI entry point
- `internal/commands/root.go` - Root command and CLI structure

### Core Components

#### Commands Layer (`internal/commands/`)
- `root.go` - Root command setup and zero-argument status
- `add.go` - Intelligent package/dotfile addition with mixed operations
- `rm.go` - Intelligent package/dotfile removal with mixed operations
- `ls.go` - Smart listing with filtering (packages, dotfiles, managers)
- `sync.go` - Apply configuration (renamed from apply)
- `install.go` - Add and sync workflow command
- `link.go` - Explicit dotfile linking operations
- `unlink.go` - Explicit dotfile unlinking operations
- `dotfiles.go` - Dotfile-specific listing with enhanced detail
- `env.go` - Environment information for debugging
- `doctor.go` - Health checks and diagnostics
- `search.go` - Package search across managers
- `info.go` - Package information display
- `init.go` - Configuration template creation
- `pipeline.go` - CommandPipeline abstraction for unified processing
- `output.go` - Output formatting utilities
- `shared.go` - Shared utilities and detection logic

#### Configuration (`internal/config/`)
- `interfaces.go` - Core configuration interfaces
- `yaml_config.go` - YAML implementation
- `adapters.go` - Bridge between config and state
- `simple_validator.go` - Configuration validation
- `schema.go` - JSON schema generation (Phase 4)

#### State Management (`internal/state/`)
- `reconciler.go` - Core reconciliation engine
- `types.go` - State types (Managed, Missing, Untracked)
- `package_provider.go` - Package state provider
- `dotfile_provider.go` - Dotfile state provider
- `adapters.go` - State-config adapters

#### Package Managers (`internal/managers/`)
- `registry.go` - Manager registry with caching
- `homebrew.go` - Homebrew/Cask implementation
- `npm.go` - NPM global packages
- `common.go` - Shared manager utilities

#### Dotfile Operations (`internal/dotfiles/`)
- `operations.go` - Core dotfile manager
- `fileops.go` - File operations (copy, backup)
- `atomic.go` - Atomic file operations

#### Shared Operations (`internal/operations/`)
- `types.go` - OperationResult and shared utilities
- `progress.go` - Progress reporting infrastructure

#### Error Handling (`internal/errors/`)
- `types.go` - Structured error types

#### Runtime Infrastructure (`internal/runtime/`) - Phase 4
- `context.go` - Shared context singleton with caching
- `logging.go` - Industry-standard logging system

#### Interface Definitions (`internal/interfaces/`) - Phase 4
- `core.go` - Unified Provider interface
- `package_manager.go` - Unified PackageManager interface
- `mocks/` - Centralized mock generation

#### Test Infrastructure (`internal/testing/`) - Phase 4
- `helpers.go` - Test context and isolation utilities

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
4. **Registration:** Add to command layer in `internal/commands/`

#### Adding a Command
1. **Command template:** `internal/commands/status.go`
2. **Registration:** `internal/commands/root.go:rootCmd.AddCommand()`
3. **Output formatting:** Use `internal/commands/output.go`
4. **Error handling:** Use `internal/errors/types.go:PlonkError`

#### Understanding State Reconciliation
1. **Entry point:** `internal/state/reconciler.go:GetState()`
2. **Provider interface:** `internal/state/reconciler.go:Provider`
3. **State types:** `internal/state/types.go:ItemState`
4. **Package provider:** `internal/state/package_provider.go`
5. **Dotfile provider:** `internal/state/dotfile_provider.go`

#### Configuration Management
1. **Interface definitions:** `internal/config/interfaces.go`
2. **YAML implementation:** `internal/config/yaml_config.go`
3. **Validation:** `internal/config/simple_validator.go`

#### Error Handling Patterns
1. **Error types:** `internal/errors/types.go:PlonkError`
2. **Error codes:** `internal/errors/types.go:ErrorCode`
3. **Error domains:** `internal/errors/types.go:Domain`
