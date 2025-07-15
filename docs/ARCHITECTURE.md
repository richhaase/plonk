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
├── commands/    # CLI command implementations using CommandPipeline
├── config/      # Configuration with interfaces and validation
├── dotfiles/    # File operations and path management
├── errors/      # Structured error types and handling
├── interfaces/  # Unified interface definitions (Phase 4)
├── managers/    # Package manager implementations
├── operations/  # Shared utilities for batch operations
├── runtime/     # Shared context and logging system (Phase 4)
├── services/    # Application service layer orchestration
├── state/       # State reconciliation engine
└── testing/     # Test helpers and utilities (Phase 4)
```

### Component Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Commands  │────▶│   Services  │────▶│    State    │
│     (CLI)   │     │(Application)│     │ (Reconciler)│
└─────────────┘     └─────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  Providers  │────▶│  Managers   │
                    │  (Package,  │     │ (Homebrew,  │
                    │  Dotfile)   │     │    NPM)     │
                    └─────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Config    │     │  Dotfiles   │     │   Errors    │
│ (Interfaces)│     │   (File     │     │ (Structured │
└─────────────┘     │ Operations) │     │  Messages)  │
                    └─────────────┘     └─────────────┘
```

## Key Components

### 1. Configuration Layer (`internal/config/`)

Provides flexible configuration management with YAML file support, validation, and environment variable integration. See the [Configuration Guide](CONFIGURATION.md) for detailed configuration options and technical implementation details.

### 2. State Management (`internal/state/`)

**Core Concept:** The reconciler compares configured state with actual system state.

**Item States:**
- `Managed` - In configuration AND installed/present
- `Missing` - In configuration BUT NOT installed/present
- `Untracked` - Installed/present BUT NOT in configuration

**Provider Interface:**
- Key methods: `Domain()`, `GetConfiguredItems()`, `GetActualItems()`, `CreateItem()`

**Reconciliation Process:**
1. Load configured items from configuration
2. Discover actual items from system
3. Compare and categorize by state
4. Return unified view for operations

### 3. Package Management (`internal/managers/`)

**Unified Interface:**
- Key methods: `IsAvailable()`, `ListInstalled()`, `Install()`, `Uninstall()`, `IsInstalled()`, `Search()`, `Info()`
- All methods accept context for cancellation and timeout support

**Implementations:**
- `HomebrewManager` - Homebrew packages and casks (macOS/Linux)
- `NpmManager` - Global NPM packages (Node.js)
- `CargoManager` - Cargo packages (Rust ecosystem)
- `PipManager` - Pip packages (Python ecosystem)
- `GemManager` - Gem packages (Ruby ecosystem)
- `AptManager` - APT packages (Debian/Ubuntu Linux)
- `GoInstallManager` - Go Install packages (Go ecosystem)

**Features:**
- Context support for cancellation and timeout
- Comprehensive error handling with smart detection
- Differentiation between expected conditions and real errors
- Context-aware error messages with actionable suggestions
- Graceful handling of unavailable managers
- BaseManager pattern for 90% code reuse across implementations
- Mock-based unit testing with 100% test coverage
- Capability discovery for optional operations (search, etc.)

**Architecture Quality:**
- **BaseManager Pattern**: Extracted common functionality reducing code duplication by ~86%
- **ErrorMatcher System**: Consistent error detection across all package managers
- **CommandExecutor Interface**: Dependency injection enabling comprehensive unit testing
- **Parser Utilities**: Common parsing patterns for output processing
- **Capability Discovery**: Runtime detection of optional package manager features

### 4. Runtime Infrastructure (`internal/runtime/`) - Phase 4

**Shared Context:**
- Singleton pattern for expensive resource initialization
- Cached access to ManagerRegistry, Reconciler, and Config
- Manager availability caching with 5-minute TTL
- Configuration caching with intelligent fallback

**Logging System:**
- Industry-standard levels: Error, Warn, Info, Debug, Trace
- Domain-specific targeting: Command, Config, Manager, State, File, Lock
- Environment control: `PLONK_DEBUG=debug:manager,state` for targeted output
- Structured output with timestamps and domain labels

**Performance Optimizations:**
- 20-30% improvement in command startup times
- Eliminated 20+ redundant initialization calls
- Resource sharing across command invocations

### 5. Interface Consolidation (`internal/interfaces/`) - Phase 4

**Unified Definitions:**
- `Provider` interface for state reconciliation
- `PackageManager` interface for all package operations
- Centralized mock generation in `mocks/` subdirectory
- Type aliases for backward compatibility

**Benefits:**
- Eliminated duplicate interface definitions
- Simplified mock infrastructure
- Enhanced testability and maintainability

### 6. Test Infrastructure (`internal/testing/`) - Phase 4

**Test Helpers:**
- `TestContext` for isolated test environments
- Temporary directory management
- Environment variable isolation
- Cleanup automation

### 7. Application Services (`internal/services/`)

**Service Layer:**
- Orchestrates operations between commands and domain logic
- Encapsulates complex business workflows
- Provides high-level operations for commands to use

**Components:**
- `dotfile_operations.go` - ApplyDotfiles orchestration
- `package_operations.go` - ApplyPackages orchestration

**Responsibilities:**
- Coordinate multiple domain operations
- Handle transaction-like workflows
- Aggregate results for presentation layer
- Maintain operation consistency

### 8. Dotfile Management (`internal/dotfiles/`)

**Components:**
- `Manager` - Path resolution and directory expansion
- `FileOperations` - Copy, backup, and validate files

**Path Conventions:**
- `zshrc` → `~/.zshrc`
- `config/nvim/` → `~/.config/nvim/`
- `editorconfig` → `~/.editorconfig`

**Features:**
- Auto-discovery of dotfiles from config directory
- Configurable ignore patterns (gitignore-style)
- Smart directory expansion to individual files
- Environment variable support (`PLONK_DIR`)
- Context-aware file operations
- Path validation and normalization

### 9. Error Handling (`internal/errors/`)

Plonk uses a structured error system for consistent error handling across all operations. See the [Development Guide](DEVELOPMENT.md#error-handling) for detailed error handling patterns and implementation guidelines.

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
**Apply:** Unified command that reconciles all managed items (installs missing packages and deploys/updates dotfiles)
**Env:** Shows environment information for debugging and troubleshooting
**Config:** Manage configuration files (show effective config with defaults, validate, edit)
**List:** Shows items filtered by state or domain (available for both packages and dotfiles)
**Doctor:** Performs comprehensive health checks on system requirements, environment, and package managers
**Search:** Intelligent package search across available package managers with context-aware behavior
**Info:** Detailed package information including version, dependencies, and installation status

## Key Design Decisions

### 1. Unified State Model
Single reconciliation pattern for all domains (packages, dotfiles) enables consistency and extensibility.

**AI Agent Benefits:**
- Consistent state representation across all domains
- Predictable state transitions (Managed → Missing → Untracked)
- Uniform error handling patterns

### 2. Interface-Based Architecture
Configuration and providers use interfaces to prevent tight coupling and improve testability.

**Key Interfaces:**
- `ConfigReader`, `ConfigWriter`, `ConfigValidator`
- `PackageManager`, `Provider`
- `DotfileConfigReader`, `PackageConfigReader`

### 3. Context Throughout
All long-running operations accept context for cancellation and timeout support.

**Implementation Pattern:**
- All operations accept context for cancellation and timeout
- Check `ctx.Done()` before and during long-running operations
- Return `ctx.Err()` on cancellation

### 4. Adapter Architecture
Adapters enable cross-package communication while preventing circular dependencies. They are a fundamental part of plonk's architecture, not technical debt.

**Purpose:**
- Bridge package boundaries without direct imports
- Prevent circular dependencies between packages
- Enable type conversion between similar interfaces
- Maintain clean separation of concerns

**When to Use Adapters vs Type Aliases:**
- **Adapters**: When interfaces differ or cross-package communication needed
- **Type Aliases**: When interfaces are identical within same package boundary

**Implementation Pattern:**
```go
type SourceTargetAdapter struct {
    source SourceInterface
}

func (a *SourceTargetAdapter) TargetMethod() error {
    return a.source.SourceMethod() // Translate and delegate
}
```

**Best Practices:**
- Keep adapters thin - translation only, no business logic
- Always document why an adapter exists
- Add compile-time interface checks: `var _ TargetInterface = (*Adapter)(nil)`
- Test adapter translations thoroughly

**Current Adapters:**
- `StatePackageConfigAdapter` - config → state for packages
- `StateDotfileConfigAdapter` - config → state for dotfiles
- `ConfigAdapter` - config types → state interfaces
- `ManagerAdapter` - managers → state.ManagerInterface
- `LockFileAdapter` - lock service → state interfaces

### 5. Comprehensive Error Handling
PackageManager methods return (result, error) following Go best practices with smart detection of expected conditions vs real errors.

**Error Categories:**
- Expected conditions (package not found) - handled gracefully
- Real errors (network failures) - propagated with context

### 6. Structured Errors
PlonkError type provides user-friendly messages and debugging context.

**Error Structure:**
- Structured errors with codes, domains, and user-friendly messages
- Compatible with standard Go error handling patterns

### 7. Environment-Aware Configuration
Uses `PLONK_DIR` environment variable for config directory and `EDITOR` for editing.

**Configuration Resolution:**
1. `$PLONK_DIR/plonk.yaml` (if PLONK_DIR set)
2. `~/.config/plonk/plonk.yaml` (default)

### 8. Convention Over Configuration
Auto-discovery of dotfiles reduces configuration burden while maintaining customization through ignore patterns.

**File Discovery Pattern:**
- Scan config directory for files
- Apply ignore patterns (gitignore-style)
- Map to home directory paths

## Extension Points

### Adding Package Managers
1. Implement `PackageManager` interface with context support and comprehensive error handling
2. Follow error handling patterns: differentiate expected conditions vs real errors
3. Include context-aware error messages with actionable suggestions
4. Register in command layer
5. Add tests including context cancellation coverage

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
