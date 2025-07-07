# Plonk Architecture

This document describes the system architecture, design philosophy, and critical design decisions of Plonk.

## Design Philosophy

### Core Principles
- **Simplicity Over Complexity** - Focused on essential package managers rather than trying to support everything
- **Reliability Through Testing** - Test-driven development ensures stability across all components
- **Interface-Based Design** - Clean abstractions enable testing and extensibility
- **Single Responsibility** - Each component has a clear, focused purpose

### Design Goals
- **Cross-Machine Consistency** - Same environment setup across multiple machines
- **Zero-Configuration Workflow** - Sensible defaults with optional customization
- **Graceful Degradation** - Works even when some package managers are unavailable
- **Developer Experience** - Clear error messages and intuitive command structure

## System Architecture

### High-Level Design

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Config Layer   │    │ Managers Layer  │
│  (commands/)    │◄──►│   (config/)     │◄──►│  (managers/)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                        ┌─────────────────┐
                        │ Execution Layer │
                        │  (executor.go)  │
                        └─────────────────┘
```

### Core Components

#### Package Manager Abstraction
**Design Decision:** Interface-based design with shared execution layer
- **CommandExecutor Interface** - Enables dependency injection and testing
- **Common Interface** - All package managers implement identical operations
- **Shared CommandRunner** - Eliminates code duplication across managers

#### Configuration Management
**Design Decision:** Pure YAML with convention-based path mapping
- **Source-Target Convention** - `config/nvim/` → `~/.config/nvim/` automatically
- **Local Override Support** - `plonk.local.yaml` for machine-specific settings
- **Package-Centric Structure** - Organized by package manager for clarity

#### CLI Architecture
**Design Decision:** Cobra framework with hierarchical command structure
- **Subcommand Organization** - `plonk pkg list [manager]` pattern
- **Consistent Flag Patterns** - `--backup`, `--dry-run` across commands
- **Progressive Disclosure** - Simple commands for common tasks, detailed options available

## Critical Design Decisions

### 1. Go Language Choice
**Why:** Balance of simplicity, performance, and deployment ease
- Single binary distribution
- Excellent CLI tooling ecosystem (Cobra)
- Strong testing support
- Cross-platform compatibility

### 2. Limited Package Manager Scope
**Why:** Focus over feature creep
- **Homebrew** - Primary package management (macOS focus)
- **ASDF** - Version management for development tools
- **NPM** - JavaScript ecosystem packages not in Homebrew
- **Excluded:** apt, yum, pacman, etc. (complexity vs. value trade-off)

### 3. Test-Driven Development Mandate
**Why:** Reliability is critical for environment management
- MockCommandExecutor prevents side effects during testing
- Interface compliance tests ensure consistent behavior
- Red-Green-Refactor cycle enforced

### 4. Configuration Generation Removal
**Why:** Simplicity over convenience
- Originally generated `.zshrc`, `.zshenv`, `.gitconfig` from YAML
- **Decision:** Treat as regular dotfiles for simplicity
- Reduces complexity by 2,262 lines of code
- Easier to understand and debug

### 5. Pure Go Implementation
**Why:** Minimize external dependencies
- No shell script dependencies
- Git operations via Go libraries (not git CLI)
- Task runner implemented in Go (dev.go)
- Consistent tooling across platforms

## Error Handling Strategy

### Design Principles
- **Graceful Degradation** - Continue operation when individual managers fail
- **Context-Rich Messages** - Include actionable information in errors
- **Standardized Wrapping** - Consistent error patterns across components

### Error Categories
- **Configuration Errors** - YAML parsing, validation failures
- **Installation Errors** - Package manager operation failures  
- **File System Errors** - Backup, apply, and restore operations
- **Network Errors** - Git operations, package downloads

## Testing Architecture

### Mock Strategy
**Design Decision:** Interface-based mocking at the command execution level
- **MockCommandExecutor** - Intercepts all external commands
- **Predictable Test Scenarios** - No network or file system dependencies
- **Interface Compliance** - All managers tested against same interface

### Test Categories
- **Unit Tests** - Individual component behavior
- **Integration Tests** - Cross-component workflows  
- **Interface Compliance** - Consistent manager behavior
- **Configuration Validation** - YAML parsing and validation

## Security Considerations

### File System Safety
- **Backup Before Overwrite** - Always preserve existing configurations
- **Path Validation** - Prevent directory traversal attacks
- **Permission Preservation** - Maintain file permissions during operations

### Command Execution
- **Input Sanitization** - Validate all external command inputs
- **Privilege Minimization** - No unnecessary elevated permissions
- **Audit Trail** - Clear logging of all system modifications

## Extension Points

### Adding Package Managers
1. Implement `Manager` interface in `pkg/managers/`
2. Add configuration structure to `pkg/config/yaml_config.go`
3. Register in CLI commands
4. Add comprehensive test coverage

### Adding Commands
1. Create command file in `internal/commands/`
2. Implement Cobra command structure
3. Add to root command registration
4. Follow existing error handling patterns

See [CONTRIBUTING.md](CONTRIBUTING.md) for development workflow and [CODEBASE_MAP.md](CODEBASE_MAP.md) for implementation details.