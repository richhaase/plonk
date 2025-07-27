# Plonk Architecture

## Overview

Plonk is a unified package and dotfile manager built with a Resource abstraction at its core. This document describes the internal architecture and how to extend plonk with new resource types.

## Directory Structure

```
internal/
├── commands/      # CLI command handlers
├── config/        # Configuration loading and types
├── orchestrator/  # Coordination layer for resources
├── lock/          # Lock file management (v2 schema)
├── output/        # Table/JSON/YAML formatting
└── resources/     # Resource implementations
    ├── resource.go    # Core Resource interface
    ├── reconcile.go   # Shared reconciliation logic
    ├── packages/      # Package manager implementations
    │   ├── manager.go     # PackageManager interface
    │   ├── homebrew.go    # Homebrew/Linuxbrew
    │   ├── npm.go         # NPM packages
    │   ├── pip.go         # Python packages
    │   ├── gem.go         # Ruby gems
    │   ├── go.go          # Go modules
    │   └── cargo.go       # Rust/Cargo packages
    └── dotfiles/      # Dotfile management
        └── manager.go     # Dotfile operations
```

## Core Concepts

### Resource Interface

The Resource interface is the foundation for all managed entities in plonk:

```go
type Resource interface {
    ID() string                    // Unique identifier for the resource type
    Desired(ctx context.Context) []Item   // Items that should exist (from config/lock)
    Actual(ctx context.Context) []Item    // Items that currently exist on system
    Apply(ctx context.Context, item Item) error  // Apply a single item
}
```

### Item Structure

Items represent individual resources (packages, dotfiles, etc.):

```go
type Item struct {
    Name   string              // Resource name (e.g., "ripgrep", ".zshrc")
    Type   string              // Resource type (e.g., "package", "dotfile")
    State  string              // Current state ("managed", "missing", "untracked")
    Error  error               // Any error associated with this item
    Meta   map[string]string   // Additional metadata
}
```

### States

- **managed**: Resource is tracked by plonk and exists on the system
- **missing**: Resource is tracked by plonk but doesn't exist on the system
- **untracked**: Resource exists on the system but isn't tracked by plonk

## Adding a New Resource Type

To add a new resource type (e.g., Docker containers, systemd services):

### 1. Implement the Resource Interface

Create a new package under `internal/resources/`:

```go
package containers

import (
    "context"
    "github.com/richhaase/plonk/internal/resources"
)

type Manager struct {
    configDir string
    // ... other fields
}

func New(configDir string) *Manager {
    return &Manager{configDir: configDir}
}

func (m *Manager) ID() string {
    return "containers"
}

func (m *Manager) Desired(ctx context.Context) []resources.Item {
    // Load from lock file or config
    // Return items that should exist
}

func (m *Manager) Actual(ctx context.Context) []resources.Item {
    // Query Docker API or docker CLI
    // Return items that currently exist
}

func (m *Manager) Apply(ctx context.Context, item resources.Item) error {
    // Start/stop container based on item.State
    // Handle "missing" → start container
    // Handle "untracked" → stop/remove container
}
```

### 2. Register with Orchestrator

Update the orchestrator to include your new resource:

```go
// In orchestrator.New() or similar initialization
resources := []resources.Resource{
    packages.NewMultiManager(config, configDir),
    dotfiles.New(configDir, homeDir),
    containers.New(configDir),  // Your new resource
}
```

### 3. Update Lock File Schema

The lock file v2 schema supports generic resources:

```yaml
version: 2
packages:           # Legacy section for backward compatibility
  homebrew:
    - name: ripgrep
      installed_at: ...
resources:          # Generic resources section
  - type: container
    id: postgres-dev
    name: postgres:15
    state: running
    meta:
      port: "5432"
      volume: "pgdata:/var/lib/postgresql/data"
```

### 4. Add CLI Commands (Optional)

If your resource needs specific commands beyond the standard apply/status:

```go
// In internal/commands/
func NewContainerCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "container",
        Short: "Manage Docker containers",
    }

    cmd.AddCommand(
        newContainerListCommand(),
        newContainerStartCommand(),
        // ... other subcommands
    )

    return cmd
}
```

## Package Managers

Package managers implement the Resource interface with additional capabilities:

### PackageManager Interface

```go
type PackageManager interface {
    resources.Resource
    Name() string
    Search(ctx context.Context, query string) ([]resources.Item, error)
    Info(ctx context.Context, name string) (*resources.Item, error)
    Install(ctx context.Context, names ...string) error
    Uninstall(ctx context.Context, names ...string) error
}
```

### Adding a New Package Manager

1. Create a new file in `internal/resources/packages/`
2. Implement the PackageManager interface
3. Register in `MultiManager` (in `packages/multi.go`)
4. Add to the `GetManager()` switch statement

Example structure:
```go
package packages

type AptManager struct {
    *BaseFields  // Common fields (homeDir, configDir, etc.)
}

func NewAptManager(base *BaseFields) *AptManager {
    return &AptManager{BaseFields: base}
}

func (m *AptManager) Name() string { return "apt" }

// ... implement other PackageManager methods
```

## Lock File

The lock file tracks all managed resources:

### V2 Schema

```go
type LockV2 struct {
    Version   int                          `yaml:"version"`
    Packages  map[string][]PackageEntry   `yaml:"packages"`
    Resources []ResourceEntry             `yaml:"resources,omitempty"`
}

type ResourceEntry struct {
    Type  string            `yaml:"type"`
    ID    string            `yaml:"id"`
    Name  string            `yaml:"name"`
    State string            `yaml:"state"`
    Meta  map[string]string `yaml:"meta,omitempty"`
}
```

### Migration

The lock package handles automatic migration from v1 to v2:
- Reader accepts both v1 and v2 formats
- Writer always outputs the latest version
- Migration is logged during apply operations

## Hooks

Plonk supports pre/post hooks for apply operations:

```yaml
# In plonk.yaml
hooks:
  pre_apply:
    - command: "echo Starting apply..."
      timeout: 30s
      continue_on_error: false
  post_apply:
    - command: "./scripts/notify.sh"
```

Hooks are executed by the orchestrator with:
- Default timeout: 10 minutes
- Default behavior: fail-fast
- Optional: continue_on_error flag

## Output System

The output package provides consistent formatting across all commands:

- **Table**: Human-readable columnar output (default)
- **JSON**: Machine-readable JSON format
- **YAML**: Machine-readable YAML format

All commands support the `-o/--output` flag to choose format.

## Testing

Plonk uses two testing approaches:

1. **Go Unit Tests**: Test internal logic and functions
   ```bash
   go test ./...
   ```

2. **BATS Behavioral Tests**: Test CLI behavior and user experience
   ```bash
   just test-bats
   ```

## Future Extensions

The Resource abstraction is designed to support:
- Docker Compose stacks
- Systemd services
- Cloud resources (AWS, GCP, etc.)
- Configuration files with templating
- Any entity that can be desired, detected, and applied

Each new resource type follows the same pattern: implement Resource interface, register with orchestrator, and optionally add CLI commands.
