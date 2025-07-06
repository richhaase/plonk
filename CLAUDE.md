# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Plonk** is a CLI tool for managing shell environments across multiple machines using Homebrew, ASDF, and NPM package managers. Written in Go with strict Test-Driven Development (TDD) practices.

## Development Commands

### Primary Workflow (via justfile)
```bash
just dev          # format + lint + test (main development workflow)
just ci           # format + lint + test + build (CI pipeline)
just build        # Build plonk binary
just test         # Run all tests
just lint         # Run golangci-lint
just format       # Format code with goimports/gofmt
```

### Testing Commands
```bash
go test ./...                          # Run all tests
go test ./pkg/config -v               # Run specific package tests
go test ./internal/commands -run TestStatus  # Run specific test
```

### Tool Management
- **ASDF tools**: Go 1.24.4, golangci-lint 2.2.1, just 1.41.0 (`.tool-versions`)
- **Pre-commit hooks**: Automatic formatting, linting, and testing before commits
- **Linting**: golangci-lint v2 with errcheck, gocritic, govet enabled

## Architecture

### Core Interfaces
```go
// pkg/managers/common.go
type CommandExecutor interface {
    Execute(name string, args ...string) *exec.Cmd
}

type PackageManager interface {
    IsAvailable() bool
    ListInstalled() ([]string, error)
}
```

### Package Structure
- **`cmd/plonk/`** - CLI entry point
- **`internal/commands/`** - CLI command implementations (status, install, apply, etc.)
- **`internal/directories/`** - Centralized path management
- **`pkg/managers/`** - Package manager abstractions (Homebrew, ASDF, NPM)
- **`pkg/config/`** - YAML configuration system with validation

### Configuration System
- **YAML-based** with validation using `go-playground/validator`
- **Source→Target convention**: `config/nvim/` → `~/.config/nvim/`
- **Local overrides** via `plonk.local.yaml`
- **Generators** for ZSH and Git configurations

## Required TDD Workflow

**CRITICAL**: All changes MUST follow this pattern:
1. **RED**: Write failing tests first
2. **GREEN**: Write minimal code to make tests pass
3. **REFACTOR**: Improve code while keeping tests green
4. **COMMIT**: Commit the changes
5. **UPDATE MEMORY**: Update CLAUDE.md to reflect completed work

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