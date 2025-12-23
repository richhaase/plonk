# Plonk Architecture

This document describes the high-level architecture and design decisions of Plonk.

## Overview

Plonk is a unified package and dotfile manager built with a clear separation of concerns and a focus on extensibility. The architecture follows domain-driven design principles with distinct layers for commands, business logic, and resource management.

## Core Concepts

### State Reconciliation

At its heart, Plonk is a state reconciliation engine. It operates on three key concepts:

1. **Desired State**: What packages and dotfiles should be present (stored in `plonk.lock`)
2. **Actual State**: What is currently installed/deployed on the system
3. **Reconciliation**: The process of comparing states and applying changes

This approach ensures idempotent operations - running `plonk apply` multiple times is safe and only changes what needs to be changed.

### Domain-Specific Implementations

Plonk implements packages and dotfiles as separate domains, each with their own reconciliation logic:

- **Packages**: Uses lock file as desired state, queries package managers for actual state
- **Dotfiles**: Uses config directory as desired state, scans home directory for actual state

Both domains share common data types (`resources.Item`, `resources.Result`) and the reconciliation algorithm (`resources.ReconcileItems`), but implement their own domain-specific logic for fetching state and applying changes.

## Architecture Layers

### 1. Command Layer (`internal/commands/`)

The command layer handles:
- CLI parsing and validation
- User interaction and output formatting
- Calling into the orchestrator or resource layers

Key components:
- `root.go` - Main command setup and global flags
- Individual command files (`install.go`, `add.go`, `clone.go`, `diff.go`, etc.)
- `shared.go` - Shared command logic
 - `helpers.go` - Command utilities

### 2. Orchestration Layer (`internal/orchestrator/`)

The orchestrator coordinates complex operations across multiple resource types:
- Manages the overall apply workflow through the coordinator
- Coordinates between package and dotfile reconciliation
- Manages orchestrator options and types
- Ensures proper error handling

### 3. Resource Layer (`internal/resources/`)

Resources are organized by type:

#### Packages (`internal/packages/`)
- Hardcoded package manager implementations with consistent interfaces
- **ManagerRegistry** provides access to all registered package managers
- Reconciliation and apply logic for package state using `manager:name` as unique keys
- Package specification parsing (`spec.go`)
- Command execution abstraction (`executor.go`) with injected executors
- Operations for install, uninstall, upgrade
- Parallel manager operations using errgroup for performance

Built-in managers: brew, npm, pnpm, cargo, pipx, conda, gem, uv, bun, go.


#### Dotfiles (`internal/dotfiles/`)
- Dotfile scanning and discovery
- Directory scanning utilities
- File copying and deployment operations
- Atomic file operations
- Path resolution and validation
- **Drift detection** via `DriftComparator` interface - compares source files in `$PLONK_DIR` with deployed files in home directory
- File comparison utilities (byte-level comparison for detecting modifications)
- Configuration file handling
- Apply operations for dotfile state
- Integration with external diff tools (configurable via `diff_tool` setting)

### 4. Configuration Layer (`internal/config/`)

Manages two types of configuration:
- `plonk.yaml` - User preferences and settings
- `plonk.lock` - State tracking (packages and dotfiles)

The lock file uses a versioned format for future compatibility.

### 5. Utility Layers

#### Lock Management (`internal/lock/`)
- Interfaces for reading/writing lock files
- YAML lock file implementation
- Version compatibility handling
- Atomic file operations

#### Diagnostics (`internal/diagnostics/`)
- System health checks
- Package manager availability
- Permission verification
- Orchestrates health checks by calling CheckHealth() on each package manager

#### Clone (`internal/clone/`)

- Clone command implementation
- Git repository cloning
- Package and dotfile deployment from cloned repositories
- Package manager detection and reporting based on the cloned lock file

#### Output Formatting (`internal/output/`)
- Multiple formatters for different commands
- Color output utilities
- Progress display and writer
- Rendering and print utilities
- Support for table, JSON, and YAML formats

#### Test Utilities (`internal/testutil/`)
- Test helpers and utilities for the codebase

## Key Design Decisions

### 1. Zero Configuration

Plonk works without any configuration files by using sensible defaults:
- Default package manager: Homebrew
- Default config location: `~/.config/plonk`
- Automatic dotfile discovery with smart filtering

### 2. Package Manager Architecture

Package managers are implemented as Go structs with consistent interfaces. This provides:
- Type-safe, testable implementations
- Consistent behavior across different tools
- Self-health checking and diagnostics
- Package upgrade management

Note: Package managers must be installed manually or via other supported managers before use. The `plonk doctor` command provides installation instructions for missing managers.

### 3. State-Based Management

Instead of tracking operations (install X, remove Y), Plonk tracks desired state. This enables:
- Idempotent operations
- Easy sharing of configurations
- Clear understanding of system state

#### State Storage Model

Plonk uses two distinct storage mechanisms:

**Package State** (`plonk.lock`):
- YAML file containing package information
- Updated atomically with each install/uninstall operation
- Tracks manager, name, and installation timestamp (no per-package version tracking)
- Example structure (v2 schema):
  ```yaml
  version: 2
  resources:
    - type: package
      id: brew:ripgrep
      metadata:
        manager: brew
        name: ripgrep
      installed_at: "2025-07-27T11:01:03-06:00"
    - type: package
      id: npm:@google/gemini-cli
      metadata:
        manager: npm
        name: "@google/gemini-cli"
      installed_at: "2025-07-28T15:11:08-06:00"
  ```

**Dotfile State** (filesystem-based):
- The `$PLONK_DIR` filesystem IS the state
- No separate tracking file needed
- Directory structure mirrors home directory without leading dots:
  ```
  $PLONK_DIR/
  ├── plonk.yaml        # Configuration (not a dotfile)
  ├── plonk.lock        # Package state (not a dotfile)
  ├── zshrc             # From ~/.zshrc
  ├── gitconfig         # From ~/.gitconfig
  └── config/
      └── nvim/
          └── init.lua  # From ~/.config/nvim/init.lua
  ```

#### Resource States

All resources exist in one of three states:
- **Managed**: Known to plonk AND exists in environment
- **Missing**: Known to plonk BUT doesn't exist in environment
- **Unmanaged**: Unknown to plonk BUT exists in environment

The reconciliation process identifies missing resources and applies them.

### 4. Structured Output

All commands support multiple output formats (table, JSON, YAML) to support:
- Human readability (table format)
- Scripting and automation (JSON/YAML)
- AI/LLM tool integration

## Extension Points

### Adding a New Package Manager

To add a new package manager, implement the `PackageManager` interface in Go:

1. Create a new manager struct in `internal/packages/`
2. Implement required methods: `Install`, `Uninstall`, `ListInstalled`, `Upgrade`, `CheckHealth`
3. Register the manager in the registry
4. Add tests for all operations

See existing managers (e.g., `brew.go`, `npm.go`) for examples.

### Adding a New Resource Type

1. Create a new domain package in `internal/resources/`
2. Implement functions to get desired and actual state
3. Use `resources.ReconcileItems` for state comparison
4. Implement apply logic for the domain
5. Update orchestrator to call the new domain
6. Add corresponding commands

### Adding a New Command

1. Create command file in `internal/commands/`
2. Register with root command
3. Implement output formatting
4. Add to command completion

## Data Flow

### Install Flow
```
User -> CLI Command -> Package Spec Parser -> Package Manager -> Lock File
                                           |
                                           -> System (actual install)
```

### Apply Flow
```
Lock File -> Orchestrator/Coordinator -> Reconcile Resources -> Apply Changes
```

### Status Flow
```
Lock File -> Reconciler -> Compare with System State -> Format Output
```

## Error Handling

Plonk uses structured error handling:
- Operations return detailed results with per-item status
- Partial failures are supported (some packages install, others fail)
- Clear error messages with actionable fixes
- Non-zero exit codes for scripting

## Security Considerations

- No elevated privileges required for user packages
- Atomic file operations prevent corruption
- Backup files created before modifications
- Git operations use standard git binary
- Package managers must be installed separately before use

## Future Considerations

The architecture is designed to support:
- Plugin system for custom package managers
- Remote state storage
- Team/organization configurations
- Service management (Docker, systemd, etc.)
- Configuration templating

## Removed Features in V2

The following features were removed in the v2 architecture to simplify the system:

- **Search and Info commands**: Removed to reduce complexity. Users can search packages using their package manager's native tools.
- **Dependency resolution**: Package managers handle their own dependencies; Plonk no longer performs topological sorting or installation ordering.
- **Per-package version tracking**: Lock file only tracks that a package is managed, not its specific version.
