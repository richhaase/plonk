# Plonk Internals

Architecture and code organization for contributors.

## Overview

Plonk is a state reconciliation engine. It compares desired state (lock file, config directory) with actual state (installed packages, deployed files) and applies changes to reconcile them.

## Directory Structure

```
plonk/
├── cmd/plonk/main.go           # Entry point
├── internal/
│   ├── commands/               # CLI commands
│   │   ├── root.go             # Root command, global flags
│   │   ├── track.go            # Package tracking
│   │   ├── untrack.go          # Package untracking
│   │   ├── add.go              # Dotfile addition
│   │   ├── rm.go               # Dotfile removal
│   │   ├── apply.go            # State application
│   │   ├── status.go           # Status display
│   │   ├── diff.go             # Drift display
│   │   ├── clone.go            # Repository cloning
│   │   ├── doctor.go           # Health checks
│   │   └── config*.go          # Configuration commands
│   ├── packages/               # Package management
│   │   ├── manager.go          # Manager interface
│   │   ├── registry.go         # Manager lookup
│   │   ├── apply.go            # Package application
│   │   ├── brew.go             # Homebrew
│   │   ├── cargo.go            # Cargo
│   │   ├── go.go               # Go
│   │   ├── pnpm.go             # PNPM
│   │   └── uv.go               # UV
│   ├── dotfiles/               # Dotfile management
│   │   ├── dotfiles.go         # Manager + operations
│   │   ├── reconcile.go        # State reconciliation
│   │   ├── apply.go            # Selective apply
│   │   ├── types.go            # Dotfile/Status types
│   │   └── fs.go               # FileSystem abstraction
│   ├── orchestrator/           # Coordination
│   │   ├── coordinator.go      # Apply coordination
│   │   └── reconcile.go        # Cross-domain reconciliation
│   ├── config/                 # Configuration
│   │   └── config.go           # Config loading/defaults
│   ├── lock/                   # Lock file
│   │   ├── v3.go               # V3 format + migration
│   │   └── types.go            # Lock types
│   ├── clone/                  # Clone operations
│   │   ├── setup.go            # Clone + apply
│   │   └── git.go              # Git operations
│   ├── diagnostics/            # Health checks
│   │   └── health.go           # System checks
│   └── output/                 # Output formatting
│       ├── formatters.go       # Table/JSON/YAML
│       └── colors.go           # Terminal colors
└── tests/bats/                 # Integration tests
```

## Key Interfaces

### Manager (packages)

```go
type Manager interface {
    IsInstalled(ctx context.Context, name string) (bool, error)
    Install(ctx context.Context, name string) error
}
```

That's it. Two methods per package manager.

### Lock Service

```go
type LockV3Service interface {
    Read() (*LockV3, error)
    Write(lock *LockV3) error
}
```

## State Model

### Package State

Stored in `plonk.lock`:
```yaml
version: 3
packages:
  brew: [fd, ripgrep]
  cargo: [bat]
```

### Dotfile State

The filesystem IS the state. Files in `$PLONK_DIR` (excluding `plonk.yaml`, `plonk.lock`) are managed dotfiles.

### Resource States

- **managed** - Tracked and exists
- **missing** - Tracked but doesn't exist
- **drifted** - Exists but modified (dotfiles only)
- **unmanaged** - Exists but not tracked

## Data Flow

### Track Flow
```
User → track command → Verify installed → Update lock file
```

### Apply Flow
```
Lock file → List tracked → Check installed → Install missing
Config dir → List files → Check deployed → Deploy missing
```

### Status Flow
```
Lock file + system state → Reconcile → Display differences
```

## Adding a Package Manager

1. Create `internal/packages/newmanager.go`:

```go
type NewManagerSimple struct{}

func NewNewManagerSimple() *NewManagerSimple {
    return &NewManagerSimple{}
}

func (m *NewManagerSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
    // Check if package is installed
}

func (m *NewManagerSimple) Install(ctx context.Context, name string) error {
    // Install package
}
```

2. Register in `internal/packages/registry.go`:

```go
case "newmanager":
    return NewNewManagerSimple(), nil
```

3. Add to `SupportedManagers` in `internal/packages/manager.go`

4. Add tests in `tests/bats/behavioral/03-package-managers.bats`

## Testing

### Unit Tests

```bash
go test ./...
go test -v ./internal/packages/...
```

### BATS Integration Tests

```bash
bats tests/bats/behavioral/
```

BATS tests call the real CLI and real package managers. Use the safe package list in `tests/bats/config/safe-packages.list`.

## Error Handling

- Commands return structured results with per-item status
- Partial failures supported (some succeed, some fail)
- Non-zero exit codes for scripting

## Output Formatting

All commands support `-o table|json|yaml`. Table is default for humans, JSON/YAML for scripting.
