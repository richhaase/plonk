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
├── state/       # State reconciliation engine
└── testing/     # Test helpers and utilities (Phase 4)
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
- Environment variable support (`PLONK_DIR` for config directory)
- Auto-discovery of dotfiles with configurable ignore patterns

**Configuration Example:**
```yaml
default_manager: homebrew
operation_timeout: 600  # 10 minutes
package_timeout: 300    # 5 minutes
dotfile_timeout: 60     # 1 minute

ignore_patterns:
  - .DS_Store
  - .git
  - "*.backup"
  - "*.tmp"
  - "*.swp"

homebrew: [git, neovim, firefox]
npm: [typescript, prettier]
```

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
- `HomebrewManager` - Homebrew packages and casks
- `NpmManager` - Global NPM packages

**Features:**
- Context support for cancellation and timeout
- Comprehensive error handling with smart detection
- Differentiation between expected conditions and real errors
- Context-aware error messages with actionable suggestions
- Graceful handling of unavailable managers

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

### 7. Dotfile Management (`internal/dotfiles/`)

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

### 5. Error Handling (`internal/errors/`)

**Structured Error System:**
Plonk implements a comprehensive structured error system that provides consistent error handling across all commands and operations.

**Core Features:**
- **Structured Error Types** - All errors use `PlonkError` with standardized fields
- **Error Codes** - Specific codes for different error categories
- **Domain Classification** - Errors grouped by functional domain
- **User-Friendly Messages** - Clear, actionable error messages
- **Debug Mode Support** - Detailed technical information when needed
- **Context Preservation** - Original error causes maintained through wrapping

**Error Codes:**
```go
// Configuration errors
ErrConfigNotFound      // Configuration file missing
ErrConfigParseFailure  // Configuration syntax error
ErrConfigValidation    // Configuration validation failed

// Package management errors
ErrPackageInstall      // Package installation failed
ErrManagerUnavailable  // Package manager not available

// File operation errors
ErrFileIO             // General file I/O error
ErrFilePermission     // Permission denied
ErrFileNotFound       // File not found

// User input errors
ErrInvalidInput       // Invalid command arguments

// System errors
ErrInternal           // Internal system error
ErrReconciliation     // State reconciliation failed
```

**Error Domains:**
```go
DomainConfig          // Configuration-related operations
DomainPackages        // Package management operations
DomainDotfiles        // Dotfile operations
DomainCommands        // Command-level operations
DomainState           // State reconciliation
```

**Error Creation Patterns:**
```go
// Create new structured error
errors.NewError(code, domain, operation, message)

// Wrap existing error with context
errors.Wrap(err, code, domain, operation, message)

// Wrap with item context (package name, file path, etc.)
errors.WrapWithItem(err, code, domain, operation, item, message)
```

**Exit Code Mapping:**
- `0` - Success
- `1` - User error (config, input validation)
- `2` - System error (permissions, unavailable managers)

**User Experience:**
- **Table Format** - Human-readable error messages with troubleshooting steps
- **JSON Format** - Structured error data for programmatic handling
- **Debug Mode** - Technical details via `PLONK_DEBUG=1` environment variable

**Integration:**
- All commands use consistent error handling patterns
- Error messages include actionable guidance
- Automatic exit code determination based on error type
- Compatible with standard Go error handling (`errors.Is`, `errors.As`)

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

### 4. Comprehensive Error Handling
PackageManager methods return (result, error) following Go best practices with smart detection of expected conditions vs real errors.

**Error Categories:**
- Expected conditions (package not found) - handled gracefully
- Real errors (network failures) - propagated with context

### 5. Structured Errors
PlonkError type provides user-friendly messages and debugging context.

**Error Structure:**
- Structured errors with codes, domains, and user-friendly messages
- Compatible with standard Go error handling patterns

### 6. Environment-Aware Configuration
Uses `PLONK_DIR` environment variable for config directory and `EDITOR` for editing.

**Configuration Resolution:**
1. `$PLONK_DIR/plonk.yaml` (if PLONK_DIR set)
2. `~/.config/plonk/plonk.yaml` (default)

### 7. Convention Over Configuration
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
