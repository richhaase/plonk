# Plonk Development Context

## Project Overview
Plonk is a Go-based state management tool for development environments that provides uniform deployment of laptop configurations across multiple machines. The system implements declarative state management for packages, dotfiles, and configuration with a focus on consistency and reconciliation.

## Core Concepts
1. **Package Management**: Manage packages across Homebrew and NPM with consistent interfaces
2. **Dotfile Management**: Deploy and synchronize dotfiles with path mapping and conventions  
3. **Configuration Management**: YAML-based configuration with validation and multi-file support
4. **State Management**: Declarative state reconciliation comparing desired (config) vs actual (installed/present)

## Architectural Principles
- **Declarative State**: Configuration defines desired state, reconcilers determine actions
- **Domain Separation**: Clear boundaries between package, dotfile, config, and state domains
- **Provider Pattern**: Pluggable providers for different domains (packages, dotfiles)
- **Interface-Driven**: Clean abstractions enable extensibility and testing

## Current Architecture

### Domain-Driven Structure
```
internal/commands/     # CLI interface layer
├── pkg*.go           # Package management commands
├── dot*.go           # Dotfile management commands  
├── status.go         # Cross-domain status command
└── output.go         # Multi-format output support

internal/state/       # Core state management domain
├── types.go          # Universal state types (Item, Result, Summary)
├── reconciler.go     # Universal state reconciliation engine
├── package_provider.go  # Package domain state provider
├── dotfile_provider.go  # Dotfile domain state provider
└── adapters.go       # Bridge existing code to new state system

internal/managers/    # Package domain implementation
├── common.go         # PackageManager interface
├── homebrew.go       # Homebrew implementation
├── npm.go           # NPM implementation
└── reconciler.go     # Legacy reconciler (being phased out)

internal/config/      # Configuration domain
├── yaml_config.go    # Configuration parsing and validation
└── simple_validator.go  # Configuration validation rules
```

### Key Design Patterns
- **Provider Pattern**: Pluggable state providers for different domains
- **Universal State Types**: Common Item/Result/Summary across all domains
- **Interface-Driven**: Clean abstractions via Provider and Manager interfaces
- **State Reconciliation**: Unified reconciliation engine for all domains
- **Command Pattern**: Self-contained CLI commands with consistent interfaces

## Current Status
- ✅ Package management complete and production-ready
- ✅ Dotfile management complete and production-ready  
- ✅ Configuration management with enhanced dotfile format
- ✅ State management foundation implemented
- ✅ ASDF support removed (simplified to Homebrew + NPM only)
- 🟡 Legacy reconciler code needs migration to new state system
- 🟡 Some command code duplication opportunities remain

## State Management System

### Universal State Types
- **Item**: Universal representation of any managed item (package, dotfile)
- **ItemState**: Managed/Missing/Untracked classification
- **Result**: State reconciliation results for a domain
- **Summary**: Aggregate state across all domains

### Provider Pattern
- **Provider Interface**: Common abstraction for all domains
- **PackageProvider**: Handles package state via manager APIs
- **DotfileProvider**: Handles dotfile state via filesystem
- **MultiManagerPackageProvider**: Aggregates multiple package managers

### Reconciliation Engine
- **Universal Reconciler**: Domain-agnostic state reconciliation
- **ConfigItem/ActualItem**: Input types for reconciliation
- **Provider.CreateItem()**: Domain-specific item creation

### Current Implementation
```go
// Register providers
reconciler := state.NewReconciler()
reconciler.RegisterProvider("package", packageProvider)
reconciler.RegisterProvider("dotfile", dotfileProvider)

// Reconcile all domains
summary, err := reconciler.ReconcileAll()
```

## Package Management Improvements Needed

### Phase 1: Extract Common Patterns

#### 1. Create Manager Registry Pattern
**Problem**: Hard-coded manager lists scattered across commands
**Location**: All `pkg_*.go` files repeat manager initialization

**Solution**: Create `internal/managers/registry.go`
```go
type ManagerRegistry struct {
    managers map[string]ManagerInfo
}

type ManagerInfo struct {
    Key         string  // "homebrew", "npm"
    DisplayName string  // "Homebrew", "NPM" 
    Manager     PackageManager
}

func NewManagerRegistry() *ManagerRegistry
func (r *ManagerRegistry) GetAvailableManagers() map[string]PackageManager
func (r *ManagerRegistry) GetManager(name string) (PackageManager, bool)
func (r *ManagerRegistry) ValidateManagerName(name string) error
```

#### 2. Extract Command Helpers
**Problem**: Duplicate initialization code across commands
**Location**: `pkg_apply.go`, `pkg_list.go`, `status.go` all repeat similar setup

**Solution**: Create `internal/commands/pkg_common.go`
```go
func initializeManagerRegistry() *managers.ManagerRegistry
func createReconciler(configDir string) (*managers.StateReconciler, error)
func getDefaultConfigDir() (string, error)
func handleManagerError(managerName string, err error, format OutputFormat)
```

#### 3. Standardize Output Structures
**Problem**: Similar but inconsistent output structures across commands
**Location**: `pkg_apply.go` (ApplyOutput), `pkg_list.go` (PackageListOutput)

**Solution**: Create `internal/commands/pkg_output.go`
```go
type BasePackageOutput struct {
    Managers []ManagerResult `json:"managers"`
}

type ManagerResult struct {
    Name     string    `json:"name"`
    Status   string    `json:"status,omitempty"`
    Count    int       `json:"count,omitempty"`
    Packages []Package `json:"packages,omitempty"`
    Error    string    `json:"error,omitempty"`
}
```

### Phase 2: Fix Inconsistencies

#### 4. Manager Name Translation
**Problem**: Manual manager name translation in multiple places
**Location**: `pkg_list.go` lines 103-108, similar in other files

**Current Code**:
```go
switch mgr.name {
case "Homebrew":
    managerKey = "homebrew"
case "NPM":
    managerKey = "npm"
}
```

**Fix**: Use ManagerInfo struct to eliminate translation

#### 5. Config Directory Handling
**Problem**: Config directory discovery repeated in every command
**Location**: All `pkg_*.go` files, `status.go`

**Fix**: Centralize in pkg_common.go helper functions

#### 6. Error Handling Consistency
**Problem**: Similar but slightly different error handling patterns
**Location**: All commands handle manager errors differently

**Fix**: Create standard error handling helper in pkg_common.go

### Phase 3: Add Robustness

#### 7. Manager Validation
**Problem**: No validation that config references valid managers
**Solution**: Add validation in config loading and command execution

#### 8. Transaction Safety
**Problem**: No rollback if batch operations fail partway
**Solution**: Consider implementing transaction-like patterns for multi-package operations

## Current Working Commands
- `plonk status` - Overall system status
- `plonk pkg list [filter]` - List packages (all/managed/missing/untracked)  
- `plonk pkg apply [--dry-run]` - Install missing packages
- `plonk pkg add <package> [--manager]` - Add to config and install
- `plonk pkg remove <package> [--uninstall]` - Remove from config (optionally uninstall)

## Known Issues
1. **Duplicate Package Bug**: `plonk pkg remove` only removes from first manager found when package exists in multiple managers
2. **Limited Test Coverage**: ~15-20% coverage, missing integration tests
3. **Dotfile Management**: Incomplete implementation (detection works, deployment doesn't)

## Todo List Context
The current todo list includes:
- Research and design 'unmanage' functionality
- Implement dotfile deployment functionality
- Add --prune flag to plonk pkg apply
- Fix duplicate package handling in remove command

## Development Environment
- Go 1.24.4
- Uses justfile for build automation (`just build`, `just test`, `just lint`)
- Git repository on branch `dogfooding/stage-1`
- No pre-commit hooks currently active

## Testing
- Unit tests in `*_test.go` files
- Mock implementations available for testing
- Run tests with `just test` or `go test ./...`
- Build with `just build`

## Notes for Future Development
- Package management system is production-ready for basic use
- Focus should shift to dotfile deployment functionality
- Code improvements listed above are polish, not critical
- System is designed for macOS primarily (Homebrew dependency)