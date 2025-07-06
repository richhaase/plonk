# Plonk Architecture

This document describes the system architecture, design decisions, and technical implementation details of Plonk.

## System Overview

Plonk is a CLI tool for managing shell environments across multiple machines. It provides a unified interface for package management and configuration deployment using Homebrew, ASDF, and NPM.

## Architecture Components

### Package Managers (`pkg/managers/`)
- **CommandExecutor Interface** - Abstraction for command execution (supports dependency injection)
- **CommandRunner** - Shared command execution logic to eliminate code duplication
- **Individual Managers** - Homebrew, ASDF, NPM with consistent interfaces

### Configuration Management (`pkg/config/`)
- **YAML Configuration** - Pure YAML config format (TOML support removed in cleanup)
- **Package-Centric Structure** - Organized by package manager with dotfiles section
- **Source-Target Convention** - Automatic path mapping (config/nvim/ → ~/.config/nvim/)
- **Local Overrides** - Support for plonk.local.yaml for machine-specific settings
- **Config Validation** - Comprehensive YAML syntax and content validation

### CLI (`internal/commands/`)
- **Cobra Framework** - Professional CLI with help, autocompletion, and subcommands
- **Status Command** - Shows availability and package counts for all managers
- **Pkg Command** - Modular package listing with `plonk pkg list [manager]` structure
- **Git Operations** - Clone and pull commands with configurable locations
- **Package Management** - Install command with automatic config application
- **Configuration Deployment** - Apply command for dotfiles and package configs with --backup and --dry-run support
- **Foundational Setup** - Setup command for installing core tools (Homebrew/ASDF/NPM)
- **Convenience Commands** - Repository-based workflow and direct repo syntax

### Testing Architecture
- **Comprehensive Test Coverage** - All components tested with TDD approach
- **Mock Command Executor** - Enables testing without actual command execution
- **Interface Compliance Tests** - Ensures consistent behavior across managers
- **Config Validation Tests** - YAML/TOML parsing and validation coverage

## Key Design Decisions

### 1. Go Over Rust/Python
Better balance of simplicity and power for CLI tools. Go provides excellent CLI tooling, simple deployment (single binary), and good performance.

### 2. Test-Driven Development
Ensures reliability and maintainability. Every feature is developed following the Red-Green-Refactor cycle.

### 3. CommandRunner Abstraction
Eliminates code duplication across package managers by providing a shared execution layer.

### 4. Interface-Based Design
Makes it easy to add new package managers or mock components for testing. All package managers implement the same interface.

### 5. Focused Package Managers
Homebrew + ASDF + NPM covers most shell environment needs without overwhelming complexity.

### 6. Cobra CLI Framework
Professional CLI with built-in help, completion, and extensibility. Industry-standard for Go CLIs.

## Technical Implementation Details

### Dependency Injection
CommandExecutor interface enables testing without side effects. Tests can use MockCommandExecutor while production uses RealCommandExecutor.

### Output Parsing
Each package manager has custom parsing logic to handle different output formats correctly.

### Error Handling
- Graceful degradation when package managers are unavailable
- Standardized error wrapping functions for consistent user experience
- Detailed error messages with context

### Package Management Features
- **Scoped Package Support** - Correctly handles NPM scoped packages (@vue/cli)
- **Version Management** - ASDF integration for language tool versioning
- **Global Package Focus** - Avoids local/project-specific package management complexity

### Git Integration
Pure Go git operations with mockable interface for testing. No dependency on git CLI.

### Configuration Features
- **Automatic Application** - Package-specific configurations applied after installation
- **Foundational Setup** - Automated installation of prerequisite tools
- **Intelligent Workflows** - Smart detection of existing repositories and installed packages

### Shell Configuration
- **ZSH Configuration Generation** - Complete .zshrc and .zshenv file generation from YAML config
- **Shell Best Practices** - Proper separation of environment variables (.zshenv) and interactive config (.zshrc)

### Backup System
- **Smart Backup Detection** - Only backs up files that will be overwritten
- **Automated Backup** - --backup flag for apply command
- **Configurable Management** - Timestamped backups with automatic cleanup and retention policies

## Development Timeline

- **Language Evolution** - Started with Rust → Python → Go (perfect for CLI tools)
- **TDD Approach** - Consistent Red-Green-Refactor cycles throughout
- **Package Manager Abstraction** - Built reusable patterns for easy extension
- **CLI Implementation** - Professional-grade command interface with Cobra
- **Focused Scope** - Refined to essential package managers for shell environment management

## Directory Structure

See [CODEBASE_MAP.md](CODEBASE_MAP.md) for detailed file structure and navigation guide.

## Future Architecture Considerations

See [ROADMAP.md](ROADMAP.md) for planned architectural enhancements and new features.