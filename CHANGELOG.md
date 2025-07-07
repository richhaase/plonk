# Changelog

All notable changes to the Plonk project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- **Complete Package Management UI** with `plonk pkg list [all|managed|missing|untracked]` and `plonk pkg status`
- **Complete Dotfiles Management UI** with `plonk dot list [all|managed|missing|untracked]` and `plonk dot status`
- **State Reconciliation Architecture** - StateReconciler, ConfigLoader, and VersionChecker for comparing config vs installed packages
- **Ultra-simple Dotfiles Implementation** - ~100 lines total using same patterns as package management
- **Machine-friendly Output Formats** - Global `--output/-o` flag supporting table (default), JSON, and YAML formats
- **Manager-specific Version Logic** - Homebrew ignores versions, ASDF/NPM require exact version matches
- **Unified CLI structure** with concept-specific commands (`plonk pkg`, `plonk dot`, future `plonk config`)
- **Justfile development workflow** replacing dev.go with just commands (build, test, lint, format, security, clean, install, precommit)
- **Simplified package managers** - Drastically simplified with minimal interface (IsAvailable, ListInstalled)

### Changed
- **CLI architecture** - Replaced monolithic status command with focused concept-specific commands
- **Development workflow** - Migrated from dev.go to justfile for simpler, standard development tasks
- **Package managers** - Simplified from 2,224 lines to ~150 lines, removing all abstraction layers
- **Dotfiles implementation** - Ultra-simplified to ~100 lines using same reconciliation patterns as packages
- **Project structure** - Moved pkg/* to internal/* following Go CLI conventions
- **State reconciliation** - Separated from package managers into dedicated reconciler service

### Fixed
- **Import command configuration path** - Fixed incorrect subdirectory usage that was inconsistent with other commands
- **Performance issues** - Removed slow state-aware methods that were causing hangs
- **Key mapping bug** - Fixed display name vs config name mapping (Homebrew vs homebrew) in reconciler
- **Test compilation** - Removed broken tests relying on deleted abstractions, added basic reconciler tests

### Removed
- **ZSH plugin management** - Removed ZSH plugin management functionality and related tests  
- **Monolithic command layer** - Removed entire internal/commands package (8400+ lines) for focused UI redesign
- **Mixed-concern status command** - Removed original status command that tried to display packages + dotfiles + drift in one view
- **Dev.go task runner** - Replaced with standard justfile for better tool ecosystem fit
- **Unused internal packages** - Removed internal/directories, internal/utils, internal/tasks after command layer removal
- **Over-engineered abstractions** - Removed CommandExecutor, CommandRunner from package managers
- **Complex package manager interface** - Simplified to minimal interface with just 2 methods
- **Broken test files** - Removed manager_test.go that relied on deleted abstractions

## [v0.3.0] - 2025-07-07

### Added
- **MIT License** with clear IP ownership documentation
- **Pure Go development workflow** with optional ASDF convenience
- **dev.go task runner** replacing Mage for simplified Go-native development

### Changed
- **Replaced Mage with pure Go dev.go task runner** for improved simplicity and maintainability
- **Comprehensive documentation reorganization** with clearer file purposes and reduced duplication
- **Updated copyright to Rich Haase** for clear intellectual property ownership
- **Streamlined development approach** focusing on Go toolchain without external dependencies

### Removed
- **ZSH and Git configuration generation** - Simplified dotfiles approach by removing automatic shell config generation
- **Mage task runner** - Replaced with lighter-weight Go-native solution
- **Documentation references** to removed ZSH/Git generation functionality

### Development
- **Temporarily disabled pre-commit hooks** for dogfooding phase to allow smooth testing workflow
- **Enhanced documentation structure** with clear role definitions for each documentation file
- **Improved project focus** on core package management without shell configuration complexity

## [v0.2.0] - 2025-07-06

### Added
- **Comprehensive versioning system** with Masterminds/semver library
  - Semantic versioning validation and parsing
  - Automated changelog updates following Keep a Changelog format
  - Git tagging workflow with release management commands
  - Version suggestion commands (mage nextpatch/nextminor/nextmajor)
  - Release preparation with commit history analysis
- **Professional CLI versioning** with --version/-v flags and version command
  - Build-time version injection via ldflags
  - Git commit hash and build date embedding
  - Support for tagged releases and development builds
- **Import command** - Generate plonk.yaml from existing shell environment
  - Discovers Homebrew packages via `brew list`
  - Discovers ASDF tools from `~/.tool-versions` (global only)
  - Discovers NPM global packages via `npm list -g`
  - Detects managed dotfiles (.zshrc, .gitconfig, .zshenv)
  - Generates clean YAML configuration file
  - Shows progress with emojis and summary statistics

### Changed
- **Migrated from Just to Mage** for Go-native task running
  - 33% performance improvement in build times
  - Eliminated shell script dependencies for better cross-platform support
  - Type-safe build logic with compile-time validation
  - Enhanced error handling in build processes
- **Standardized installation approach** to single opinionated method
  - Unified on `go install ./cmd/plonk` for all installations
  - Eliminated confusing multiple installation options
  - Updated all documentation for consistent installation experience
  - Fixed command examples to assume global installation

### Enhanced
- AsdfManager now supports `ListGlobalTools()` method for reading `~/.tool-versions`
- **Release management documentation** in CONTRIBUTING.md and CODEBASE_MAP.md
  - Complete workflow documentation for semantic versioning
  - Mage command examples for all release operations
  - Pre-release testing guidance and best practices

### Code Quality
- Fixed pre-commit hooks for reliable development workflow
- Standardized import organization using goimports with local prefixes
- Improved function documentation to follow Go idioms
- Added comprehensive package-level documentation for all packages
- Removed problematic errcheck and gocritic linters
- Enhanced API documentation readiness with godoc improvements

## [0.9.0] - 2025-01-06 - Documentation & Standards

### Documentation
- Created comprehensive documentation structure
- Added TODO.md for AI agent work tracking
- Added ROADMAP.md for future development plans
- Added CONTRIBUTING.md with TDD workflow guidelines
- Added ARCHITECTURE.md with system design details
- Refactored README.md to focus on user-facing content

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