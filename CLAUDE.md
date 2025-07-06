# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Plonk** is a CLI tool for managing shell environments across multiple machines using Homebrew, ASDF, and NPM package managers. Written in Go with strict Test-Driven Development (TDD) practices.

## Development Environment

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and workflow commands.

## Architecture Overview

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed system design and [CODEBASE_MAP.md](CODEBASE_MAP.md) for navigation.

## Required TDD Workflow

Follow the Test-Driven Development workflow described in [CONTRIBUTING.md](CONTRIBUTING.md).

**AI Agent Specific Requirements**:
- Sync with TODO.md at session start and maintain it throughout
- Reference ARCHITECTURE.md for system design decisions
- Use CODEBASE_MAP.md for navigation
- Follow established patterns found in existing code

### Testing Patterns
- **MockCommandExecutor** for unit tests (avoid actual command execution)
- **setupTestEnv(t)** helper for test isolation
- **Table-driven tests** for comprehensive coverage
- **Interface compliance** tests for new package managers

## Standardized Patterns

### Error Handling
```go
// internal/commands/error_handling.go
WrapConfigError(err)                    // Configuration loading errors
WrapInstallError(packageName, err)      // Package installation errors
WrapPackageManagerError("homebrew", err) // Package manager availability errors
```

### Command Structure
All CLI commands follow consistent patterns:
- Cobra command structure in `internal/commands/`
- Argument validation using `ValidateNoArgs()`, `ValidateExactArgs()`
- Error wrapping for consistent user experience
- Comprehensive test coverage with mocks

### File Operations
- Use `internal/directories.Default` for all path operations
- Backup functionality available via `internal/commands/backup.go`
- Configuration generation via `pkg/config/*_generator.go`

## CLI Commands

For complete command reference, see README.md. Key commands:
- `status` - Package manager availability and drift detection
- `install` - Install packages from config  
- `apply` - Apply configurations (supports --backup, --dry-run)
- `clone`/`pull` - Git repository operations

## Critical Files

- **`internal/commands/root.go`** - CLI structure and command registration
- **`pkg/managers/common.go`** - Core interfaces and patterns
- **`pkg/config/yaml_config.go`** - Configuration structure and parsing
- **`internal/commands/error_handling.go`** - Standardized error patterns
- **`internal/commands/test_helpers.go`** - Test utilities and patterns