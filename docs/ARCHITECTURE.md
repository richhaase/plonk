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

### Resource Abstraction

Plonk treats packages and dotfiles as "resources" with a common interface:

```go
type Resource interface {
    ID() string
    Desired() []Item
    Actual(ctx context.Context) []Item
    Apply(ctx context.Context, item Item) error
}
```

This abstraction allows for future resource types (services, configurations, etc.) without major architectural changes.

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
- `package_validation.go` - Package validation utilities

### 2. Orchestration Layer (`internal/orchestrator/`)

The orchestrator coordinates complex operations across multiple resource types:
- Manages the overall apply workflow through the coordinator
- Handles hooks execution
- Coordinates between package and dotfile reconciliation
- Manages orchestrator options and types
- Ensures proper error handling

### 3. Resource Layer (`internal/resources/`)

Resources are organized by type:

#### Packages (`internal/resources/packages/`)
- Package manager interfaces and implementations
- Registry for dynamic package manager discovery
- Dependency resolution system with topological sorting for correct installation order
- Reconciliation and apply logic for package state
- Package specification parsing (`spec.go`)
- Command execution abstraction (`executor.go`)
- Operations for install, uninstall, search, info, upgrade
- Health checking and self-installation capabilities
- Outdated package detection

Supported package managers:
- Homebrew (brew) - macOS and Linux packages
- NPM (npm) - Node.js global packages
- PNPM (pnpm) - Fast, disk-efficient Node.js packages
- Cargo (cargo) - Rust packages
- Pipx (pipx) - Python applications in isolated environments
- Conda (conda) - Scientific computing and data science packages
- Gem (gem) - Ruby packages
- Go (go install) - Go modules and tools
- UV (uv) - Python tool manager
- Pixi (pixi) - Conda-forge package manager
- Composer (composer) - PHP global packages
- .NET (dotnet) - .NET Global Tools


#### Dotfiles (`internal/resources/dotfiles/`)
- Dotfile scanning and discovery
- Directory scanning utilities
- File copying and deployment operations
- Atomic file operations
- Path resolution and validation
- File comparison utilities
- Configuration file handling
- Apply operations for dotfile state

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
- Automatic package manager installation via SelfInstall()
- Dependency-aware installation that resolves package manager dependencies and installs in correct order
- Tool installation

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

### 2. Package Manager Abstraction

Each package manager implements a common interface, allowing:
- Consistent behavior across different tools
- Easy addition of new package managers
- Capability detection (search, info, etc.)
- Self-health checking and diagnostics
- Automatic self-installation during environment setup
- Package upgrade management

### 3. State-Based Management

Instead of tracking operations (install X, remove Y), Plonk tracks desired state. This enables:
- Idempotent operations
- Easy sharing of configurations
- Clear understanding of system state

#### State Storage Model

Plonk uses two distinct storage mechanisms:

**Package State** (`plonk.lock`):
- YAML file containing detailed package information
- Updated atomically with each install/uninstall operation
- Tracks name, version, and installation timestamp
- Example structure:
  ```yaml
  version: 1
  packages:
    brew:
      - name: ripgrep
        installed_at: 2025-07-27T11:01:03.519704-06:00
        version: 14.1.1
      - name: neovim
        installed_at: 2025-07-27T11:00:51.028708-06:00
        version: 0.11.3
    npm:
      - name: '@google/gemini-cli'
        installed_at: 2025-07-28T15:11:08.74692-06:00
        version: 0.1.14
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

1. Implement the `PackageManager` interface in `internal/resources/packages/`
2. Implement required operations (Install, Uninstall, ListInstalled, etc.)
3. Implement optional capabilities through interfaces (search, info)
4. Implement health checking via CheckHealth() method
5. Implement self-installation via SelfInstall() method
6. Implement package upgrade capabilities via Upgrade() and Outdated() methods

### Adding a New Resource Type

1. Implement the `Resource` interface
2. Add reconciliation logic
3. Update orchestrator to handle the new type
4. Add corresponding commands

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
                                      |
                                      -> Execute Hooks
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
- No automatic execution of downloaded scripts
- Git operations use standard git binary

## Future Considerations

The architecture is designed to support:
- Plugin system for custom package managers
- Remote state storage
- Team/organization configurations
- Service management (Docker, systemd, etc.)
- Configuration templating
