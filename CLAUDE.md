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

### Core Commands
```bash
./plonk status                     # Package manager availability and drift detection
./plonk pkg list [manager]         # List installed packages
./plonk install                    # Install packages from config
./plonk apply [--backup] [--dry-run] # Apply configurations
./plonk clone <repo>               # Clone dotfiles repository
./plonk setup                      # Install foundational tools
./plonk repo <repo>                # Complete setup workflow
```

## Current Status

**Phase**: Maintenance & Code Quality
- **Linter issues reduced**: 1245 → ~60 (95% improvement)
- **Active task**: Configure errcheck filtering and fix 3 real validation issues
- **Next priority**: Task 52h - Organize imports consistently across all files

## Project History

See **CHANGELOG.md** for complete development history, including:
- 40+ completed features with TDD methodology
- Version history from v0.1.0 (core features) to v0.8.0 (development infrastructure)
- Technical evolution and architectural improvements

## Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/go-git/go-git/v5` - Git operations
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/go-playground/validator/v10` - Configuration validation

## Critical Files

- **`internal/commands/root.go`** - CLI structure and command registration
- **`pkg/managers/common.go`** - Core interfaces and patterns
- **`pkg/config/yaml_config.go`** - Configuration structure and parsing
- **`internal/commands/error_handling.go`** - Standardized error patterns
- **`internal/commands/test_helpers.go`** - Test utilities and patterns