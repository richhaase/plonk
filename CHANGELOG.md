# Changelog

All notable changes to the Plonk project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### In Progress
- Configure errcheck filtering and fix 3 real validation issues (infra-8)

### Planned
- Organize imports consistently across all files (52h)
- Convert remaining tests to table-driven format (52j)

## [0.8.0] - 2025-01-06 - Development Infrastructure

### Added
- Development tools management with `.tool-versions` (golang 1.24.4, golangci-lint 2.2.1, just 1.41.0)
- Task runner implementation with `justfile` for development workflow commands
- Pre-commit hook integration for automatic code formatting and quality checks
- Comprehensive linting configuration with golangci-lint v2

### Changed
- Reduced linter issues from 1245 to ~60 (95% improvement)
- Automated code formatting via go fmt integration
- Enhanced development workflow with `just dev` and `just ci` commands

## [0.7.0] - 2025-01-05 - Code Quality & Standardization

### Added
- Standardized error handling patterns across all commands
- Centralized error wrapping functions (WrapConfigError, WrapInstallError, etc.)
- Standardized argument validation (ValidateNoArgs, ValidateExactArgs, etc.)
- Package installation helper functions to eliminate code duplication

### Changed
- Refactored all CLI commands to use consistent error handling
- Consolidated test helper functions to eliminate repetitive patterns
- Improved code organization with centralized utilities

### Removed
- Repetitive error handling patterns across command files
- Duplicate test setup code (33+ test functions updated)

## [0.6.0] - 2025-01-04 - Advanced Configuration Features

### Added
- Dry-run flag support for apply command with comprehensive preview
- Configuration validation system using go-playground/validator library
- YAML syntax validation with specific error messages
- Package name validation across all manager sections
- File path validation for configuration paths

### Changed
- Simplified validation system from 280+ lines to ~160 lines using standard library
- Enhanced apply command with visual indicators for dry-run mode
- Improved validation error messages with field context

## [0.5.0] - 2025-01-03 - Codebase Architecture Cleanup

### Added
- Centralized DirectoryManager for all path operations with caching
- Phase 1 codebase cleanup removing legacy systems

### Changed
- Consolidated package manager interfaces into single location
- Renamed YAML config types to be primary (YAMLConfig → Config)
- Updated all commands to use centralized directory management

### Removed
- Legacy TOML configuration system (130+ lines of dead code)
- Duplicate PackageManager interface definitions
- Redundant internal/commands/directory.go file

## [0.4.0] - 2025-01-02 - Configuration Generation

### Added
- Git configuration management with GenerateGitconfig function
- Support for all standard Git configuration sections (user, core, aliases, etc.)
- Git configuration generation integrated into apply command workflow
- Local config override support (plonk.yaml + plonk.local.yaml merging)

### Changed
- Enhanced apply command to generate both ZSH and Git configurations
- Improved backup support for generated configuration files

## [0.3.0] - 2025-01-01 - Backup System

### Added
- Comprehensive backup functionality with configurable location and cleanup
- `--backup` flag for apply command with automated backups
- Standalone `plonk backup` command for manual backup operations
- Smart backup detection that only backs up files that will be overwritten
- Configurable backup retention with automatic cleanup

### Changed
- Enhanced apply command with backup integration
- Improved file operation safety with backup support

## [0.2.0] - 2024-12-31 - Shell Configuration Management

### Added
- ZSH configuration management with complete .zshrc and .zshenv generation
- ZSH configuration structure supporting plugins, env vars, aliases, and functions
- Integration of ZSH management into apply command workflow
- Comprehensive test coverage for ZSH config integration scenarios

### Changed
- Enhanced apply command to generate shell configurations from YAML
- Improved configuration deployment with shell-specific file generation

## [0.1.0] - 2024-12-30 - Core Features & Commands

### Added
- CLI infrastructure with Cobra framework and professional command interface
- Core package manager abstractions (Homebrew, ASDF, NPM)
- YAML-based configuration system with validation
- Source→Target convention for configuration file deployment
- Eight core CLI commands:
  - `status` - Package manager availability and drift detection
  - `pkg list` - Modular package listing with manager-specific support
  - `clone` - Git repository cloning with branch support
  - `pull` - Git repository updates
  - `install` - Package installation from configuration
  - `apply` - Configuration file deployment
  - `setup` - Foundational tool installation
  - `repo` - Complete setup workflow
- Configuration drift detection with SHA256 file comparison
- MockCommandExecutor and RealCommandExecutor for comprehensive testing
- Git operations with go-git integration and mockable interfaces
- Package-specific configuration application
- Local configuration override support (plonk.local.yaml)

### Changed
- Migrated from TOML to YAML for better nested structure support
- Focused package manager support to Homebrew, ASDF, and NPM only
- Implemented Test-Driven Development workflow throughout

### Technical Features
- Dependency injection pattern with CommandExecutor interface
- Interface-based design for easy extension and testing
- Comprehensive test coverage with both unit and integration tests
- Error handling with graceful degradation when package managers unavailable
- NPM scoped package support (@anthropic-ai/claude-code)
- Configurable clone locations via PLONK_DIR environment variable
- Intelligent clone vs pull detection based on repository state