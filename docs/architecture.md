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
// Simplified for documentation - see internal/resources/resource.go for exact signatures
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

### 2. Orchestration Layer (`internal/orchestrator/`)

The orchestrator coordinates complex operations across multiple resource types:
- Manages the overall apply workflow through the coordinator
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


#### Dotfiles (`internal/resources/dotfiles/`)
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

Each package manager implements a common interface following the Interface Segregation Principle (ISP). This allows:
- Consistent behavior across different tools
- Easy addition of new package managers
- Optional capability detection via type assertions (search, info, upgrade, health checks, self-install)
- Managers implement only the capabilities they support
- Self-health checking and diagnostics
- Automatic self-installation during environment setup
- Package upgrade management

See [Capability Usage Examples](#capability-usage-examples) for implementation patterns.

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
- Example structure (v2 schema):
  ```yaml
  version: 2
  resources:
    - type: package
      id: brew:ripgrep
      metadata:
        manager: brew
        name: ripgrep
        version: 14.1.1
      installed_at: "2025-07-27T11:01:03-06:00"
    - type: package
      id: npm:@google/gemini-cli
      metadata:
        manager: npm
        name: "@google/gemini-cli"
        version: 0.1.14
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

1. Implement the `PackageManager` interface in `internal/resources/packages/`
2. Implement required operations (Install, Uninstall, ListInstalled, etc.)
3. Implement optional capabilities through interfaces (search, info, upgrade, health, self-install)
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
- Package manager self-installation may execute remote scripts (opt-in via SelfInstall capability)
- Git operations use standard git binary
- Users should review package manager installation scripts before using automatic installation

## Future Considerations

The architecture is designed to support:
- Plugin system for custom package managers
- Remote state storage
- Team/organization configurations
- Service management (Docker, systemd, etc.)
- Configuration templating
### Capability Usage Examples

Optional capabilities are exposed via small interfaces. Callers must perform a type assertion before invoking optional methods, or use the helper predicates for readability.

Type assertions:
```
// Info (PackageInfoProvider)
if infoProvider, ok := mgr.(packages.PackageInfoProvider); ok {
    info, err := infoProvider.Info(ctx, name)
    _ = info; _ = err
}

// Search (PackageSearcher)
if searcher, ok := mgr.(packages.PackageSearcher); ok {
    results, err := searcher.Search(ctx, query)
    _ = results; _ = err
}

// Upgrade (PackageUpgrader)
if upgrader, ok := mgr.(packages.PackageUpgrader); ok {
    _ = upgrader.Upgrade(ctx, []string{name})
}

// Health (PackageHealthChecker)
if hc, ok := mgr.(packages.PackageHealthChecker); ok {
    _, _ = hc.CheckHealth(ctx)
}

// Self-install (PackageSelfInstaller)
if si, ok := mgr.(packages.PackageSelfInstaller); ok {
    _ = si.SelfInstall(ctx)
}
```

Helper predicates:
```
if packages.SupportsInfo(mgr) { /* safe to assert PackageInfoProvider */ }
if packages.SupportsSearch(mgr) { /* safe to assert PackageSearcher */ }
if packages.SupportsUpgrade(mgr) { /* safe to assert PackageUpgrader */ }
if packages.SupportsHealthCheck(mgr) { /* safe to assert PackageHealthChecker */ }
if packages.SupportsSelfInstall(mgr) { /* safe to assert PackageSelfInstaller */ }
```

These examples demonstrate the intent of the capability model: managers may omit optional features; callers should not assume availability without checking.
