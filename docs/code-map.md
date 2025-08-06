# Plonk Code Map

This document provides a comprehensive map of the plonk codebase to aid in implementation documentation.

## Command Entry Points

### Main Entry Point
- `cmd/plonk/main.go` - Application entry point, initializes root command

### Command Implementations (`internal/commands/`)

#### Core Commands
- `root.go` - Root command setup, global flags, version info
- `clone.go` - Clone dotfiles repository and apply it
- `apply.go` - Reconcile system state with desired configuration
- `status.go` - Display managed packages and dotfiles
- `doctor.go` - System health checks and fixes

#### Package Management Commands
- `install.go` - Install and track packages
- `uninstall.go` - Remove packages from system and tracking
- `search.go` - Search for packages across managers
- `info.go` - Display package information

#### Dotfile Management Commands
- `add.go` - Add dotfiles to plonk management
- `rm.go` - Remove dotfiles from plonk management
- `dotfiles.go` - Shared dotfile command utilities

#### Configuration Commands
- `config.go` - Parent command for configuration management
- `config_show.go` - Display configuration values
- `config_edit.go` - Edit configuration file

#### Utility Commands
- `env.go` - Display environment information (hidden command)

#### Command Support Files
- `helpers.go` - Shared command utilities and helpers
- `output.go` - Output formatting infrastructure
- `output_utils.go` - Output formatting utilities
- `output_types.go` - Output type definitions
- `shared.go` - Shared command logic

## Core Components

### Configuration (`internal/config/`)
- `config.go` - Configuration loading and management
- `constants.go` - Configuration constants and defaults
- `compat.go` - Configuration compatibility handling

### Lock File Management (`internal/lock/`)
- `interfaces.go` - Lock file interfaces
- `types.go` - Lock file type definitions
- `yaml_lock.go` - YAML lock file implementation

### Orchestration (`internal/orchestrator/`)
- `orchestrator.go` - Main orchestration logic
- `apply.go` - Apply command orchestration
- `hooks.go` - Pre/post hook management

### Resources (`internal/resources/`)

#### Base Resource Infrastructure
- `resource.go` - Resource interface and base types
- `types.go` - Resource type definitions
- `reconcile.go` - Resource reconciliation logic

#### Package Resources (`internal/resources/packages/`)
- `resource.go` - Package resource implementation
- `reconcile.go` - Package reconciliation logic
- `operations.go` - Package operations (install, uninstall, etc.)
- `interfaces.go` - Package manager interfaces
- `registry.go` - Package manager registry
- `helpers.go` - Package utilities

##### Package Manager Implementations
- `homebrew.go` - Homebrew package manager
- `npm.go` - NPM package manager
- `cargo.go` - Cargo package manager
- `pip.go` - Pip package manager
- `gem.go` - Gem package manager
- `goinstall.go` - Go install package manager

##### Package Parsing
- `parsers/parsers.go` - Package string parsing utilities

#### Dotfile Resources (`internal/resources/dotfiles/`)
- `resource.go` - Dotfile resource implementation
- `manager.go` - Dotfile management operations
- `reconcile.go` - Dotfile reconciliation logic
- `scanner.go` - Dotfile discovery and scanning
- `filter.go` - Dotfile filtering logic
- `fileops.go` - File operations (copy, remove, etc.)
- `atomic.go` - Atomic file operations
- `expander.go` - Path expansion utilities

### Setup Utilities (`internal/setup/`)
- `setup.go` - Setup command implementation
- `git.go` - Git operations for cloning
- `tools.go` - Tool installation logic
- `prompts.go` - Interactive prompt utilities

### Diagnostics (`internal/diagnostics/`)
- `health.go` - System health check implementations

### Output Formatting (`internal/output/`)
- `formatters.go` - Output formatting implementations

## Command-to-Implementation Mapping

### setup command
- Entry: `internal/commands/setup.go`
- Logic: `internal/setup/setup.go`
- Git operations: `internal/setup/git.go`
- Tool installation: `internal/setup/tools.go`
- Uses: `internal/diagnostics/health.go` (via doctor --fix for language package managers only)

### apply command
- Entry: `internal/commands/apply.go`
- Orchestration: `internal/orchestrator/apply.go`
- Package reconciliation: `internal/resources/packages/reconcile.go`
- Dotfile reconciliation: `internal/resources/dotfiles/reconcile.go`
- Hook execution: `internal/orchestrator/hooks.go`

### status command
- Entry: `internal/commands/status.go`
- Resource reconciliation: `internal/resources/reconcile.go`
- Package state: `internal/resources/packages/resource.go`
- Dotfile state: `internal/resources/dotfiles/resource.go`

### doctor command
- Entry: `internal/commands/doctor.go`
- Health checks: `internal/diagnostics/health.go`
- Tool installation: `internal/setup/tools.go` (for --fix)

### config show/edit commands
- Entry: `internal/commands/config_show.go`, `config_edit.go`
- Config management: `internal/config/config.go`

### Package management commands (install/uninstall/search/info)
- Entries: `internal/commands/install.go`, `uninstall.go`, `search.go`, `info.go`
- Operations: `internal/resources/packages/operations.go`
- Manager registry: `internal/resources/packages/registry.go`
- Lock file updates: `internal/lock/yaml_lock.go`

### Dotfile management commands (add/rm)
- Entries: `internal/commands/add.go`, `rm.go`
- Operations: `internal/resources/dotfiles/manager.go`
- File operations: `internal/resources/dotfiles/fileops.go`
- Path filtering: `internal/resources/dotfiles/filter.go`

## Key Interfaces

### Resource Interface
Location: `internal/resources/resource.go`
- Implemented by: Package and Dotfile resources
- Used by: Reconciliation and orchestration

### PackageManager Interface
Location: `internal/resources/packages/interfaces.go`
- Implemented by: All package managers
- Defines: Install, Uninstall, List, Search, Info operations

### Lock Interface
Location: `internal/lock/interfaces.go`
- Implemented by: YAMLLock
- Defines: Read, Write, Update operations

## Data Flow Patterns

### Command Execution Flow
1. `cmd/plonk/main.go` → Initialize root command
2. `internal/commands/{command}.go` → Parse flags, validate input
3. Component layer → Execute business logic
4. Output formatting → Format results based on --output flag

### Resource Reconciliation Flow
1. Load desired state (lock file or filesystem)
2. Query actual state (system packages or deployed files)
3. Compare states to identify missing/unmanaged
4. Apply changes to reach desired state

### Package Operation Flow
1. Parse package specification (name, manager prefix)
2. Look up package manager in registry
3. Execute operation via manager interface
4. Update lock file if state changed

### Dotfile Operation Flow
1. Expand and validate file paths
2. Apply filtering rules
3. Execute file operations
4. No lock file update (filesystem is state)

## Common Patterns and Conventions

### Error Handling
- Commands return structured results with per-item status
- Partial failures are supported (some operations succeed, others fail)
- Non-zero exit codes for scripting compatibility

### Output Formatting
- All commands support -o/--output flag (table, json, yaml)
- Table format is default for human readability
- JSON uses PascalCase field names
- YAML uses snake_case field names

### Context Usage
- Commands accept context.Context for cancellation
- Context flows through all layers
- Used for timeouts in search operations

### Testing Patterns
- Unit tests alongside implementation files (*_test.go)
- Test helpers in test_helpers.go or helpers_test.go
- Integration tests in integration_test.go files

### Configuration Patterns
- Zero-config by default with sensible defaults
- Configuration loaded once and passed down
- Environment variables override config file

### File Organization
- Commands in internal/commands/
- Business logic in internal/resources/, internal/orchestrator/
- Shared utilities in internal/ subdirectories
- Public API (if any) would be in pkg/
