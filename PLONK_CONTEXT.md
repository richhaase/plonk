# Plonk Project Context

## Overview

**Plonk** is a unified package and dotfile manager for developers that maintains consistency across multiple machines. It uses state reconciliation to compare desired configuration with actual system state and applies necessary changes.

### Key Features
- **Unified management**: Packages (Homebrew, NPM) and dotfiles in one configuration
- **State reconciliation**: Automatically detects and applies missing configurations
- **Auto-discovery**: Finds dotfiles automatically with configurable ignore patterns
- **AI-friendly**: Structured output formats and clear command syntax
- **Cross-platform**: Works on macOS, Linux, and Windows
- **Context-aware**: Full cancellation and timeout support

## Project Architecture

### Core Principles
1. **State Reconciliation** - All operations reconcile configured vs actual state
2. **Provider Pattern** - Extensible architecture for different domains  
3. **Interface-Based Design** - Loose coupling through well-defined interfaces
4. **Context-Aware Operations** - Cancellable operations with configurable timeouts
5. **Structured Error Handling** - User-friendly errors with actionable guidance

### Directory Structure
```
cmd/plonk/           # CLI entry point
internal/
├── commands/        # CLI command implementations  
├── config/          # Configuration with interfaces and validation
├── dotfiles/        # File operations and path management
├── errors/          # Structured error types and handling
├── managers/        # Package manager implementations
└── state/           # State reconciliation engine
docs/                # Documentation
```

### Key Components

#### 1. Configuration Layer (`internal/config/`)
- **Interfaces**: `ConfigReader`, `ConfigWriter`, `ConfigValidator`, `DotfileConfigReader`, `PackageConfigReader`
- **Implementation**: `YAMLConfigService` - YAML file implementation of all interfaces
- **Features**: Flexible YAML unmarshaling, validation, environment variable support, auto-discovery

#### 2. State Management (`internal/state/`)
- **Core Concept**: Reconciler compares configured state with actual system state
- **Item States**: `Managed`, `Missing`, `Untracked`
- **Provider Interface**: Defines `Domain()`, `GetConfiguredItems()`, `GetActualItems()`, `CreateItem()`
- **Implementations**: `PackageProvider`, `DotfileProvider`, `MultiManagerPackageProvider`

#### 3. Package Management (`internal/managers/`)
- **Unified Interface**: `PackageManager` with context support and comprehensive error handling
- **Implementations**: `HomebrewManager`, `NpmManager`
- **Features**: Context support, comprehensive error handling, graceful manager unavailability

#### 4. Dotfile Management (`internal/dotfiles/`)
- **Components**: `Manager` (path resolution), `FileOperations` (copy, backup, validate)
- **Path Conventions**: `zshrc` → `~/.zshrc`, `config/nvim/` → `~/.config/nvim/`
- **Features**: Auto-discovery, configurable ignore patterns, atomic operations

#### 5. Error Handling (`internal/errors/`)
- **Structured Error System**: `PlonkError` with standardized fields
- **Error Codes**: Specific codes for different categories (`CONFIG_NOT_FOUND`, `PACKAGE_INSTALL`, etc.)
- **Domain Classification**: Errors grouped by functional domain
- **Exit Code Mapping**: Automatic mapping to appropriate exit codes (0=success, 1=user error, 2=system error)

## Configuration

### Configuration File
**Location**: `~/.config/plonk/plonk.yaml` (or `$PLONK_DIR/plonk.yaml`)

**Structure**:
```yaml
settings:
  default_manager: homebrew
  operation_timeout: 300
  package_timeout: 180
  dotfile_timeout: 60

ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"

homebrew:
  - git
  - curl
  - neovim

npm:
  - typescript
  - prettier
```

### Dotfile Auto-Discovery
- Automatically discovered from config directory
- Configurable ignore patterns (gitignore-style)
- Path mapping: `~/.config/plonk/zshrc` → `~/.zshrc`

## CLI Commands

### Core Commands
- **`plonk status`** - Show system state across all domains
- **`plonk apply`** - Apply configuration (install packages, deploy dotfiles)
- **`plonk env`** - Environment info for debugging
- **`plonk doctor`** - Health checks with actionable diagnostics
- **`plonk search <package>`** - Intelligent package search
- **`plonk info <package>`** - Detailed package information

### Package Management
- **`plonk pkg list [filter]`** - List packages by state
- **`plonk pkg add [package]`** - Add packages to configuration
- **`plonk pkg remove <package>`** - Remove from configuration

### Dotfile Management
- **`plonk dot list [filter]`** - List dotfiles by state
- **`plonk dot add <dotfile>`** - Add/update dotfile management

### Configuration Management
- **`plonk config show`** - Display current configuration
- **`plonk config validate`** - Validate configuration
- **`plonk config edit`** - Edit configuration file

### Global Options
- **`--output, -o`** - Output format: table|json|yaml (default: table)
- **`--version, -v`** - Show version information
- **`--help, -h`** - Show help

## Data Flow

### Reconciliation Process
1. **Load Configuration** - Via ConfigReader interface
2. **Create Providers** - With appropriate adapters
3. **Get Configured Items** - From configuration
4. **Get Actual Items** - From system (packages installed, files present)
5. **Compare States** - Categorize as Managed/Missing/Untracked
6. **Execute Operations** - Install packages, copy files, etc.
7. **Report Results** - With structured errors if needed

### State Types
- **Managed** - In configuration AND present on system
- **Missing** - In configuration BUT NOT present on system
- **Untracked** - Present on system BUT NOT in configuration

## Development

### Build System
- **Tool**: `just` command runner
- **Commands**: `just build`, `just test`, `just install`, `just lint`, `just precommit`
- **Requirements**: Go 1.24.4+, Just, Git

### Testing Strategy
- **Unit tests** with mocks for all components
- **Table-driven tests** for comprehensive coverage
- **Context cancellation tests** for all long-running operations
- **Structured error testing** with proper error codes and domains

### Error Handling Patterns
```go
// Create structured error
return errors.NewError(
    errors.ErrPackageInstall,    // Error code
    errors.DomainPackages,       // Domain
    "install",                   // Operation
    "failed to install package" // Message
)

// Wrap existing error
return errors.Wrap(err, code, domain, operation, message)

// Add item context
return errors.WrapWithItem(err, code, domain, operation, item, message)
```

### Key Interfaces

#### Package Manager
```go
type PackageManager interface {
    IsAvailable(ctx context.Context) (bool, error)
    ListInstalled(ctx context.Context) ([]string, error)
    Install(ctx context.Context, name string) error
    Uninstall(ctx context.Context, name string) error
    IsInstalled(ctx context.Context, name string) (bool, error)
    Search(ctx context.Context, query string) ([]string, error)
    Info(ctx context.Context, name string) (*PackageInfo, error)
}
```

#### State Provider
```go
type Provider interface {
    Domain() string
    GetConfiguredItems() ([]ConfigItem, error)
    GetActualItems(ctx context.Context) ([]ActualItem, error)
    CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
}
```

## Environment Variables
- **`PLONK_DIR`** - Config directory override (default: `~/.config/plonk`)
- **`EDITOR`** - Editor for `plonk config edit`
- **`PLONK_DEBUG`** - Enable detailed error information

## AI Agent Integration

### Structured Output
All commands support `--output json` for machine parsing:
```bash
plonk status --output json | jq '.packages.homebrew[] | select(.state == "missing")'
```

### Error Handling
- Consistent error codes across all commands
- Domain-based categorization
- User-friendly messages with troubleshooting steps
- Technical details available in debug mode

### Batch Operations
Commands can be chained for automated workflows:
```bash
plonk doctor --output json && plonk apply --dry-run
```

## Extension Points

### Adding Package Managers
1. Implement `PackageManager` interface with context support
2. Follow error handling patterns
3. Register in command layer
4. Add tests including context cancellation coverage

### Adding New Domains
1. Create `Provider` implementation
2. Define configuration interface
3. Create adapters if needed
4. Register with reconciler

## Release Process
- **Automated Release**: `just release-auto v1.2.3`
- **Pre-release validation**: tests, lint, security, build
- **Cross-platform binary building**
- **Version guidelines**: Semantic versioning

## System Requirements
- **Go 1.24.4+** (for building)
- **Just** (command runner)
- **Git** (version management)
- **Package managers**: Homebrew (macOS), NPM (optional)
- **Platform support**: macOS, Linux, Windows
- **Architecture support**: AMD64, ARM64