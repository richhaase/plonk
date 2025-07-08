# Plonk Architecture

## Overview

Plonk manages packages and dotfiles through unified state reconciliation. It compares desired state (configuration) with actual state (system) and reconciles differences.

## Core Principles

1. **State Reconciliation** - All operations reconcile configured vs actual state
2. **Provider Pattern** - Extensible architecture for different domains  
3. **Interface-Based Design** - Loose coupling through well-defined interfaces
4. **Context-Aware Operations** - Cancellable operations with configurable timeouts
5. **Structured Error Handling** - User-friendly errors with actionable guidance

## Component Architecture

### Directory Structure

```
internal/
├── commands/    # CLI command implementations  
├── config/      # Configuration with interfaces and validation
├── dotfiles/    # File operations and path management
├── errors/      # Structured error types and handling
├── managers/    # Package manager implementations
└── state/       # State reconciliation engine
```

### Component Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Commands  │────▶│    State    │────▶│  Managers   │
│     (CLI)   │     │ (Reconciler)│     │ (Homebrew,  │
└─────────────┘     └─────────────┘     │    NPM)     │
       │                    │            └─────────────┘
       │                    │                     
       ▼                    ▼                    
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Config    │     │  Providers  │────▶│  Dotfiles   │
│ (Interfaces)│────▶│  (Package,  │     │   (File     │
└─────────────┘     │  Dotfile)   │     │ Operations) │
                    └─────────────┘     └─────────────┘
                            │                     
                            ▼                    
                    ┌─────────────┐              
                    │   Errors    │              
                    │ (Structured │              
                    │  Messages)  │              
                    └─────────────┘              
```

## Key Components

### 1. Configuration Layer (`internal/config/`)

**Interfaces:**
- `ConfigReader` - Load configuration from various sources
- `ConfigWriter` - Save configuration back to storage
- `ConfigValidator` - Validate configuration correctness
- `DotfileConfigReader` - Domain-specific dotfile config access
- `PackageConfigReader` - Domain-specific package config access

**Implementation:**
- `YAMLConfigService` - YAML file implementation of all interfaces
- `ConfigAdapter` - Bridges Config struct to domain interfaces
- State adapters for type conversion between packages

**Features:**
- Flexible YAML unmarshaling (simple strings or complex objects)
- Validation with custom rules and timeout settings
- Support for `plonk.yaml` and `plonk.local.yaml` overrides

**Configuration Example:**
```yaml
settings:
  default_manager: homebrew
  operation_timeout: 600  # 10 minutes

dotfiles:
  - zshrc                 # Simple form
  - source: config/nvim/  # Directory form
    destination: ~/.config/nvim/

homebrew:
  brews: [git, neovim]
  casks: [firefox]
```

### 2. State Management (`internal/state/`)

**Core Concept:** The reconciler compares configured state with actual system state.

**Item States:**
- `Managed` - In configuration AND installed/present
- `Missing` - In configuration BUT NOT installed/present  
- `Untracked` - Installed/present BUT NOT in configuration

**Provider Interface:**
```go
type Provider interface {
    Domain() string
    GetConfiguredItems() ([]ConfigItem, error)
    GetActualItems(ctx context.Context) ([]ActualItem, error)
    CreateItem(name, state, configured, actual) Item
}
```

**Reconciliation Process:**
1. Load configured items from configuration
2. Discover actual items from system
3. Compare and categorize by state
4. Return unified view for operations

### 3. Package Management (`internal/managers/`)

**Unified Interface:**
```go
type PackageManager interface {
    IsAvailable(ctx context.Context) bool
    ListInstalled(ctx context.Context) ([]string, error)
    Install(ctx context.Context, name string) error
    Uninstall(ctx context.Context, name string) error
    IsInstalled(ctx context.Context, name string) bool
}
```

**Implementations:**
- `HomebrewManager` - Homebrew packages and casks
- `NpmManager` - Global NPM packages

**Features:**
- Context support for cancellation
- Graceful handling of unavailable managers
- Clear error reporting with exit codes

### 4. Dotfile Management (`internal/dotfiles/`)

**Components:**
- `Manager` - Path resolution and directory expansion
- `FileOperations` - Copy, backup, and validate files

**Path Conventions:**
- `zshrc` → `~/.zshrc`
- `config/nvim/` → `~/.config/nvim/`  
- `dot_gitconfig` → `~/.gitconfig`

**Features:**
- Smart directory expansion to individual files
- Backup creation before modifications
- Context-aware file operations
- Path validation and normalization

### 5. Error Handling (`internal/errors/`)

**Structured Error Type:**
```go
type PlonkError struct {
    Code      ErrorCode    // Specific error type
    Domain    Domain       // Where error occurred
    Operation string       // What was being done
    Message   string       // Technical details
    Cause     error        // Original error
}
```

**Error Domains:**
- Config, Dotfiles, Packages, State, Commands

**Features:**
- User-friendly messages with solutions
- Error wrapping with context preservation
- Compatible with standard Go error handling
- Metadata for debugging

## Data Flow

### Reconciliation Process

1. **Load Configuration** - Via ConfigReader interface
2. **Create Providers** - With appropriate adapters
3. **Get Configured Items** - From configuration
4. **Get Actual Items** - From system (packages installed, files present)
5. **Compare States** - Categorize as Managed/Missing/Untracked
6. **Execute Operations** - Install packages, copy files, etc.
7. **Report Results** - With structured errors if needed

### Command Examples

**Status:** Shows all items across all domains with their states
**Apply:** Reconciles missing items (installs packages, copies dotfiles)
**List:** Shows items filtered by state or domain

## Key Design Decisions

1. **Unified State Model** - Single reconciliation pattern for all domains enables consistency and extensibility.

2. **Interface-Based Architecture** - Configuration and providers use interfaces to prevent tight coupling and improve testability.

3. **Context Throughout** - All long-running operations accept context for cancellation and timeout support.

4. **Structured Errors** - PlonkError type provides user-friendly messages and debugging context.

5. **Separate File Operations** - Dotfile operations extracted from state management for clarity and reusability.

## Extension Points

### Adding Package Managers
1. Implement `PackageManager` interface with context support
2. Register in command layer
3. Add tests

### Adding New Domains  
1. Create `Provider` implementation
2. Define configuration interface
3. Create adapters if needed
4. Register with reconciler

### Adding Configuration Formats
1. Implement config interfaces (Reader, Writer, Validator)
2. Add format-specific validation
3. Update command layer

## Testing & Quality

### Testing Strategy
- **Unit tests** with mocks for all components
- **Test isolation** using `t.TempDir()` - no system dependencies
- **Table-driven tests** for comprehensive coverage
- **Context cancellation tests** - Complete coverage for all long-running operations including package managers, file operations, and state reconciliation

### Security
- Path validation prevents directory traversal
- Backup creation before modifications
- No arbitrary code execution
- Home directory boundaries enforced

### Performance
- Lazy package manager initialization
- Efficient directory traversal
- Minimal command execution
- Future: Concurrent reconciliation

## Development

**Build:** `just` command runner with common tasks  
**Dependencies:** Minimal - only essential Go packages  
**Contributing:** Extend via interfaces, maintain test coverage