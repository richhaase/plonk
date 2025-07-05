# Plonk - Shell Environment Lifecycle Manager

## Project Overview

Plonk is a CLI tool for managing shell environments across multiple machines. It helps you manage package installations and environment switching using a focused set of package managers:

- **Homebrew** - Primary package installation
- **ASDF** - Programming language tools and versions
- **NPM** - Packages not available via Homebrew (like claude-code)

## Development Approach

This project was developed using **Test-Driven Development (TDD)** with Red-Green-Refactor cycles throughout the implementation.

## Architecture

### Package Managers (`pkg/managers/`)
- **CommandExecutor Interface** - Abstraction for command execution (supports dependency injection)
- **CommandRunner** - Shared command execution logic to eliminate code duplication
- **Individual Managers** - Homebrew, ASDF, NPM with consistent interfaces

### CLI (`internal/commands/`)
- **Cobra Framework** - Professional CLI with help, autocompletion, and subcommands
- **Status Command** - Shows availability and package counts for all managers
- **`--all` Flag** - Option to show complete package lists vs. truncated view

### Testing
- **47 Total Tests** - Comprehensive test coverage across all components
- **Mock Command Executor** - Enables testing without actual command execution
- **Interface Compliance Tests** - Ensures consistent behavior across managers

## File Structure

```
plonk/
├── cmd/plonk/main.go              # CLI entry point
├── internal/commands/             # CLI commands
│   ├── root.go                   # Root command definition
│   ├── status.go                 # Status command implementation
│   └── status_test.go            # Status command tests
├── pkg/managers/                 # Package manager implementations
│   ├── common.go                 # CommandExecutor interface & CommandRunner
│   ├── executor.go               # Real command execution for production
│   ├── homebrew.go              # Homebrew package manager
│   ├── asdf.go                  # ASDF tool manager
│   ├── npm.go                   # NPM global package manager
│   └── manager_test.go          # Comprehensive test suite
├── go.mod                       # Go module definition
└── CLAUDE.md                    # This documentation
```

## Usage

### Build and Install
```bash
go build ./cmd/plonk
```

### Commands
```bash
./plonk --help                   # Show main help
./plonk status                   # Quick package overview (first 5 per manager)
./plonk status --all            # Complete package lists
./plonk status --help           # Status command help
```

### Example Output
```
Package Manager Status
=====================

## Homebrew
✅ Available
📦 139 packages installed:
   - aichat
   - aider
   - ansible
   - ansible-lint
   - asdf
   ... and 134 more (use --all to show all packages)

## ASDF
✅ Available
📦 8 packages installed:
   - golang
   - nodejs
   - opentofu
   - python
   - terraform-docs
   ... and 3 more (use --all to show all packages)

## NPM
✅ Available
📦 5 packages installed:
   - lib
   - corepack
   - eslint
   - prettier
   - typescript
```

## Todo List History

### ✅ Completed Tasks

1. **Redesign config structure for environment profiles (home/work)** - ✅ Completed
   - Pivoted to shell environment lifecycle manager focus

2. **Add package management lifecycle (install/update/drift detection)** - ✅ Completed  
   - Implemented comprehensive package manager abstractions

3. **Complete package manager trait implementations for all managers** - ✅ Completed
   - Built Homebrew, ASDF, NPM, Pip, Cargo managers with full CRUD operations

4. **Create CommandExecutor trait for testability** - ✅ Completed
   - Implemented dependency injection pattern for command execution

5. **Add testing dependencies to Cargo.toml** - ✅ Completed
   - Transitioned from Rust to Go, added Cobra CLI framework

6. **Write unit tests with mocked commands** - ✅ Completed
   - 47 comprehensive tests with MockCommandExecutor

7. **Write integration tests with real commands** - ✅ Completed  
   - RealCommandExecutor for production use

8. **Add Pip package manager implementation with TDD** - ✅ Completed
   - Full implementation with user-level package management

9. **Add Cargo package manager implementation with TDD** - ✅ Completed
   - Complete Rust package management with binary installation

10. **Create CLI status command that uses all package managers** - ✅ Completed
    - Professional CLI with Cobra framework and comprehensive status reporting

11. **Update package managers to use explicit global flags for global-only package management** - ✅ Completed
    - Focused approach: removed Pip/Cargo, kept Homebrew/ASDF/NPM

12. **Remove Pip and Cargo managers, keep only Homebrew, ASDF, and NPM** - ✅ Completed
    - Streamlined to preferred toolchain with --all flag for detailed views

## Development Timeline

- **Language Evolution**: Started with Rust → Python → Go (perfect for CLI tools)
- **TDD Approach**: Consistent Red-Green-Refactor cycles throughout
- **Package Manager Abstraction**: Built reusable patterns for easy extension
- **CLI Implementation**: Professional-grade command interface with Cobra
- **Focused Scope**: Refined to essential package managers for shell environment management

## Key Design Decisions

1. **Go Over Rust/Python** - Better balance of simplicity and power for CLI tools
2. **Test-Driven Development** - Ensures reliability and maintainability
3. **CommandRunner Abstraction** - Eliminates code duplication across managers
4. **Interface-Based Design** - Easy to add new package managers or mock for testing
5. **Focused Package Managers** - Homebrew + ASDF + NPM covers most shell environment needs
6. **Cobra CLI Framework** - Professional CLI with built-in help, completion, and extensibility

## Technical Highlights

- **Dependency Injection** - CommandExecutor interface enables testing without side effects
- **Output Parsing** - Handles different package manager output formats correctly
- **Error Handling** - Graceful degradation when package managers are unavailable
- **Scoped Package Support** - Correctly handles NPM scoped packages (@vue/cli)
- **Version Management** - ASDF integration for language tool versioning
- **Global Package Focus** - Avoids local/project-specific package management complexity