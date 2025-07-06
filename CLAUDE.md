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

### Configuration Management (`pkg/config/`)
- **YAML-First Design** - Primary config format with TOML legacy support
- **Package-Centric Structure** - Organized by package manager with dotfiles section
- **Source-Target Convention** - Automatic path mapping (config/nvim/ → ~/.config/nvim/)
- **Local Overrides** - Support for plonk.local.yaml for machine-specific settings

### CLI (`internal/commands/`)
- **Cobra Framework** - Professional CLI with help, autocompletion, and subcommands
- **Status Command** - Shows availability and package counts for all managers
- **Pkg Command** - Modular package listing with `plonk pkg list [manager]` structure
- **Git Operations** - Clone and pull commands with configurable locations
- **Package Management** - Install command with automatic config application
- **Configuration Deployment** - Apply command for dotfiles and package configs
- **Foundational Setup** - Setup command for installing core tools (Homebrew/ASDF/NPM)
- **Convenience Commands** - Repository-based workflow and direct repo syntax

### Testing
- **Comprehensive Test Coverage** - All components tested with TDD approach
- **Mock Command Executor** - Enables testing without actual command execution
- **Interface Compliance Tests** - Ensures consistent behavior across managers
- **Config Validation Tests** - YAML/TOML parsing and validation coverage

## File Structure

```
plonk/
├── cmd/plonk/main.go              # CLI entry point
├── internal/commands/             # CLI commands
│   ├── root.go                   # Root command definition
│   ├── status.go                 # Status command implementation
│   ├── pkg.go                    # Package listing commands
│   ├── clone.go                  # Clone command with git operations
│   ├── pull.go                   # Pull command for updates
│   ├── install.go                # Install command for packages from config
│   ├── apply.go                  # Apply command for configuration deployment
│   ├── setup.go                  # Setup command for foundational tools
│   ├── repo.go                   # Repository convenience command
│   ├── test_helpers.go          # Shared testing utilities
│   ├── status_test.go           # Status command tests
│   ├── clone_test.go            # Clone command tests
│   ├── pull_test.go             # Pull command tests
│   ├── install_test.go          # Install command tests
│   ├── apply_test.go            # Apply command tests
│   ├── setup_test.go            # Setup command tests
│   ├── repo_test.go             # Repository convenience command tests
│   ├── backup.go                # Backup functionality with configurable location
│   └── backup_test.go           # Backup functionality tests
├── pkg/managers/                 # Package manager implementations
│   ├── common.go                 # CommandExecutor interface & CommandRunner
│   ├── executor.go               # Real command execution for production
│   ├── homebrew.go              # Homebrew package manager
│   ├── asdf.go                  # ASDF tool manager
│   ├── npm.go                   # NPM global package manager
│   └── manager_test.go          # Comprehensive test suite
├── pkg/config/                   # Configuration management
│   ├── config.go                 # Legacy TOML config support
│   ├── config_test.go           # TOML config tests
│   ├── yaml_config.go           # Primary YAML config implementation
│   ├── yaml_config_test.go      # YAML config tests
│   ├── zsh_generator.go         # ZSH configuration file generation
│   └── zsh_generator_test.go    # ZSH generator tests
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
./plonk status                   # Package manager availability and counts
./plonk pkg list                 # List packages from all managers
./plonk pkg list brew            # List only Homebrew packages
./plonk pkg list asdf            # List only ASDF tools
./plonk pkg list npm             # List only NPM packages

# Foundational setup
./plonk setup                    # Install Homebrew, ASDF, and Node.js/NPM

# Git operations
./plonk clone <repo>             # Clone dotfiles repository
./plonk pull                     # Pull updates to existing repository

# Package and configuration management
./plonk install                  # Install packages from config
./plonk apply                    # Apply all configuration files
./plonk apply <package>          # Apply configuration for specific package
./plonk apply --backup           # Apply all configurations with backup

# Backup operations
./plonk backup                   # Backup all files that apply would overwrite
./plonk backup ~/.zshrc ~/.vimrc # Backup specific files

# Convenience commands
./plonk repo <repo>              # Complete setup: clone + install + apply
./plonk <repo>                   # Same as above (convenience syntax)

# Environment variable
PLONK_DIR=~/my-dotfiles ./plonk clone <repo>  # Clone to custom location
```

### Example Output
```
Package Manager Status
=====================

## Homebrew
✅ Available
📦 139 packages installed

## ASDF
✅ Available
📦 8 packages installed

## NPM
✅ Available
📦 6 packages installed
```

## Configuration Format

The new YAML-based configuration supports both simple and complex package definitions:

```yaml
settings:
  default_manager: homebrew

# Standalone config files (no package install needed)
dotfiles:
  - zshrc                    # -> ~/.zshrc
  - zshenv                   # -> ~/.zshenv
  - plugins.zsh              # -> ~/.plugins.zsh
  - dot_gitconfig            # -> ~/.gitconfig

homebrew:
  brews:
    - aichat                 # Simple package
    - aider
    - name: neovim           # Package with config
      config: config/nvim/   # -> ~/.config/nvim/
    - name: mcfly
      config: config/mcfly/  # -> ~/.config/mcfly/
  
  casks:
    - font-hack-nerd-font
    - google-cloud-sdk

asdf:
  - name: nodejs
    version: "24.2.0"
    config: config/npm/      # -> ~/.config/npm/
  - name: python
    version: "3.13.2"
  - name: golang
    version: "1.24.4"

npm:
  - "@anthropic-ai/claude-code"
  - name: some-tool
    package: "@scope/different-name"
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
    - Streamlined to preferred toolchain

13. **Implement pkg list command structure to replace --all flag** - ✅ Completed
    - Added modular `plonk pkg list [manager]` command structure
    - Supports individual manager listing and all managers
    - Correctly handles NPM scoped packages like @anthropic-ai/claude-code

14. **Design package-centric config with default_manager and simplified npm handling** - ✅ Completed
    - Created package-centric TOML configuration structure
    - Added default manager support to reduce repetition

15. **Implement TOML config parsing with package definitions** - ✅ Completed
    - Built TOML parsing with package validation
    - Added local config override support (plonk.local.toml)

16. **Create config package struct and validation logic** - ✅ Completed
    - Implemented comprehensive validation for all package managers
    - Added ASDF version requirement validation

17. **Refactor config to use YAML with simplified source->target convention** - ✅ Completed
    - Migrated from TOML to YAML for better nested structure support
    - Added dotfiles section for standalone configuration files
    - Implemented source-to-target path convention (config/nvim/ -> ~/.config/nvim/)
    - Created mixed simple/complex package definitions within manager lists
    - Added comprehensive test coverage following TDD methodology

18. **Implement separate plonk clone and plonk pull commands with configurable location** - ✅ Completed
    - Separated clone and pull functionality for Unix-like simplicity
    - Added go-git integration for pure Go git operations
    - Implemented configurable clone location via PLONK_DIR environment variable
    - Created mockable GitInterface for comprehensive testing
    - Built clean separation: clone always clones, pull always pulls

19. **Implement plonk install command (install packages from config)** - ✅ Completed
    - Created package installation from YAML config using existing package managers
    - Added automatic configuration application for newly installed packages
    - Implemented graceful handling when package managers are unavailable
    - Built comprehensive test coverage with TDD methodology

20. **Add plonk apply command (deploy config files)** - ✅ Completed
    - Implemented dotfile deployment using source->target convention
    - Added support for both global dotfiles and package-specific configurations
    - Created package-specific application (plonk apply <package>)
    - Built file and directory copying functionality with proper error handling

21. **Create plonk setup command for foundational tool installation** - ✅ Completed
    - Built setup command that installs Homebrew → ASDF → Node.js/NPM in sequence
    - Added platform detection and prerequisite checking
    - Implemented graceful handling when tools are already installed
    - Created clear user guidance for foundational vs repository-based setup

22. **Add plonk repo command (convenience: clone/pull + install + apply)** - ✅ Completed
    - Renamed previous setup to repo command for repository-based setup
    - Implemented complete workflow: git operations → package installation → config application
    - Added root command support for `plonk <repo>` convenience syntax
    - Built intelligent clone vs pull detection based on existing repository state

23. **Add config drift detection (Red-Green-Refactor)** - ✅ Completed
    - Implemented config drift detection tests (Red phase)
    - Built drift detection in status command (Green phase)
    - Refactored status command for better drift reporting (Refactor phase)
    - Added SHA256 file comparison for detecting configuration changes

24. **Add branch support for clone command (Red-Green-Refactor)** - ✅ Completed
    - Added branch support tests for clone command (Red phase)
    - Implemented branch support in git operations (Green phase)
    - Refactored clone command with flag and URL syntax (Refactor phase)
    - Supports both `--branch` flag and `repo#branch` URL syntax

25. **Design and implement ZSH configuration management** - ✅ Completed
    - Designed ZSH configuration structure with plugins, env vars, aliases, and functions
    - Implemented ZSH plugin manager with auto-clone and update functionality
    - Replaced auto-detection with explicit initialization and completion commands in config
    - Generated complete .zshrc and .zshenv files from plonk configuration using TDD

26. **Integrate ZSH management into apply command** - ✅ Completed
    - Added ZSH configuration file generation to apply command workflow
    - Integrated .zshrc and .zshenv generation with existing dotfiles and package config application
    - Built comprehensive test coverage for ZSH config integration scenarios

27. **Implement backup functionality with configurable location and count-based cleanup** - ✅ Completed
    - Created BackupExistingFile() for individual file backups with timestamp
    - Implemented BackupConfigurationFiles() with configurable backup directory
    - Added BackupConfig structure to YAML config with location and keep_count settings
    - Built automatic cleanup of old backups to maintain configured count limit
    - Defaults to ~/.config/plonk/backups/ with 5 backup retention
    - Comprehensive TDD implementation with edge case testing

28. **Implement apply --backup flag and standalone backup command with TDD** - ✅ Completed
    - Added --backup flag to apply command for automated backups before applying (Red-Green-Refactor)
    - Created standalone `plonk backup` command for manual backup operations
    - Implemented smart backup detection that only backs up files that will be overwritten
    - Added comprehensive test coverage for both backup scenarios (apply --backup and standalone backup)
    - Supports backup of dotfiles, ZSH configs, and package-specific configurations
    - Integrated with existing configurable backup location and cleanup system

29. **Add gitconfig management functionality with full TDD cycle** - ✅ Completed
    - Created GitConfig struct supporting all standard Git configuration sections (user, core, aliases, etc.)
    - Implemented GenerateGitconfig function for creating .gitconfig files from YAML configuration
    - Added Git configuration generation to apply command workflow (follows ZSH pattern)
    - Integrated Git configuration backup support for apply --backup flag
    - Support for local config overrides (plonk.yaml + plonk.local.yaml merging)
    - Comprehensive test coverage including unit tests and integration tests
    - Enables users to manage .gitconfig declaratively through plonk.yaml

30. **Implement YAML syntax validation with TDD cycle (Task 47a1-47a2)** - ✅ Completed
    - Created comprehensive YAML syntax validation tests (Red phase)
    - Implemented ValidateYAML function for config validation (Green phase)
    - Handles empty content, comments, and various YAML syntax errors
    - Uses gopkg.in/yaml.v3 decoder for robust validation
    - Provides specific error messages for common syntax issues
    - All tests passing with proper edge case handling

31. **Implement package name validation with TDD cycle (Task 47a3-47a4)** - ✅ Completed
    - Added tests for package name validation across all manager sections (Red phase)
    - Implemented ValidatePackageName and ValidateConfigContent functions (Green phase)
    - Support for homebrew (brews/casks), npm, and asdf package validation
    - Character validation without regex - allows letters, numbers, hyphens, underscores, dots, @, /
    - Prevents empty names, whitespace, and invalid patterns (starting/ending with hyphens)
    - Validates both simple string packages and complex objects with name fields
    - Clear error messages with section and index information for debugging

32. **Implement file path validation with TDD cycle (Task 47a5-47a6)** - ✅ Completed
    - Added comprehensive file path validation tests for config content (Red phase)
    - Implemented ValidateFilePath function and integrated with ValidateConfigContent (Green phase)
    - Validation of config paths in package objects across all package managers
    - Validation of standalone dotfiles section file paths
    - Prevents absolute paths, empty paths, and paths with problematic characters
    - File path rules: relative paths only, no special characters, clear error messages
    - Supports both package-specific config paths and standalone dotfile paths

### 🔄 Current Pending Tasks (Re-prioritized)

**Prioritization Rationale**: Foundation & daily value over new features. Prioritize what makes plonk safer and more reliable for daily use.

**HIGH PRIORITY (Foundation & Daily Value):**

**1. Task Group 47 - Developer Experience Enhancements (Tasks 47a-47i):**
- Config validation prevents broken deploys
- Dry-run mode gives confidence before changes  
- Diff mode shows drift (builds on existing drift detection)
- These make plonk safer and more predictable daily
- *Needs detailed specs*

47a. **Design config validation system with YAML syntax and content checks (Red phase)** - 🟡 Pending
47b. **Implement config validation functionality (Green phase)** - 🟡 Pending
47c. **Add dry-run mode tests for preview functionality (Red phase)** - 🟡 Pending
47d. **Implement dry-run mode for apply command (Green phase)** - 🟡 Pending
47e. **Add diff mode tests for showing configuration differences (Red phase)** - 🟡 Pending
47f. **Implement diff mode functionality (Green phase)** - 🟡 Pending
47g. **Add watch mode tests for auto-apply on config changes (Red phase)** - 🟡 Pending
47h. **Implement watch mode functionality (Green phase)** - 🟡 Pending
47i. **Refactor developer experience features with unified CLI (Refactor phase)** - 🟡 Pending

**2. Task Group 46 - Integration Tests (Tasks 46a-46c):**
- Validates all existing functionality works together
- Critical before adding more features
- Prevents regressions as complexity grows

46a. **Add integration tests for end-to-end workflows (Red phase)** - 🟡 Pending
46b. **Implement comprehensive integration test suite (Green phase)** - 🟡 Pending
46c. **Refactor integration tests with CI/CD support (Refactor phase)** - 🟡 Pending

**3. Task Group 44 - Repository Infrastructure (Tasks 44a-44j):**
- Code quality tools prevent bugs
- Pre-commit hooks ensure consistency
- Foundation for sustainable development

44a. **Add pre-commit hook tests for Go formatting (Red phase)** - 🟡 Pending
44b. **Implement pre-commit hooks for Go formatting (Green phase)** - 🟡 Pending
44c. **Add linting tests with golangci-lint (Red phase)** - 🟡 Pending
44d. **Implement golangci-lint configuration and hooks (Green phase)** - 🟡 Pending
44e. **Refactor code quality setup with development workflow integration (Refactor phase)** - 🟡 Pending
44f. **Add development workflow tests (Red phase)** - 🟡 Pending
44g. **Implement development workflow tool (Green phase)** - 🟡 Pending
44h. **Add test coverage enforcement tests (Red phase)** - 🟡 Pending
44i. **Implement test coverage tooling (Green phase)** - 🟡 Pending
44j. **Refactor development workflow with documentation and optimization (Refactor phase)** - 🟡 Pending

**MEDIUM PRIORITY (Extend Core Value):**

**4. Task Group 49 - Additional Shell Support (Tasks 49a-49e):**
- Bash/Fish users can't fully use plonk without this
- Natural extension of existing ZSH support
- *Needs detailed specs*

49a. **Add Bash shell config generation tests (Red phase)** - 🟡 Pending
49b. **Implement Bash shell config generation functionality (Green phase)** - 🟡 Pending
49c. **Add Fish shell config generation tests (Red phase)** - 🟡 Pending
49d. **Implement Fish shell config generation functionality (Green phase)** - 🟡 Pending
49e. **Refactor shell config generation with multi-shell support (Refactor phase)** - 🟡 Pending

**5. Task Group 38 - Import Command (Tasks 38a-38e):**
- Helps users migrate TO plonk
- One-time use but high value for adoption

38a. **Add shell config parsing tests for common formats (Red phase)** - 🟡 Pending
38b. **Implement basic .zshrc/.bashrc parsing functionality (Green phase)** - 🟡 Pending
38c. **Add tests for plonk.yaml generation from parsed configs (Red phase)** - 🟡 Pending
38d. **Implement plonk import command with YAML suggestion (Green phase)** - 🟡 Pending
38e. **Refactor import command with support for multiple shell types (Refactor phase)** - 🟡 Pending

**6. Task Group 48 - Advanced Backup/Restore Features (Tasks 48a-48i):**
- Nice-to-have enhancements over basic backup
- Selective restore most valuable, encryption/compression less so
- *Needs detailed specs*

48a. **Add selective restore tests for granular file restoration (Red phase)** - 🟡 Pending
48b. **Implement selective restore functionality (Green phase)** - 🟡 Pending
48c. **Add backup compression tests for space optimization (Red phase)** - 🟡 Pending
48d. **Implement backup compression functionality (Green phase)** - 🟡 Pending
48e. **Add remote backup tests for cloud sync (Red phase)** - 🟡 Pending
48f. **Implement remote backup sync functionality (Green phase)** - 🟡 Pending
48g. **Add backup encryption tests for sensitive data protection (Red phase)** - 🟡 Pending
48h. **Implement backup encryption functionality (Green phase)** - 🟡 Pending
48i. **Refactor advanced backup features with unified management (Refactor phase)** - 🟡 Pending

**LOW PRIORITY (Nice to Have):**

**7. Task Group 50 - Package Manager Extensions (Tasks 50a-50c):**
- mas is niche (App Store apps)
- Current package managers cover most needs
- *Needs detailed specs*

50a. **Add mas command support tests for Mac App Store integration (Red phase)** - 🟡 Pending
50b. **Implement mas command functionality for App Store apps (Green phase)** - 🟡 Pending
50c. **Refactor package manager integration with mas support (Refactor phase)** - 🟡 Pending

**8. Task Group 43 - Full Environment Snapshots (Tasks 43a-43g):**
- Complex feature with unclear use cases
- Current config management may be sufficient

43a. **Add full environment snapshot tests (Red phase)** - 🟡 Pending
43b. **Implement plonk snapshot create functionality (Green phase)** - 🟡 Pending
43c. **Add snapshot restoration tests (Red phase)** - 🟡 Pending
43d. **Implement plonk snapshot restore functionality (Green phase)** - 🟡 Pending
43e. **Add snapshot management tests (list, delete) (Red phase)** - 🟡 Pending
43f. **Implement plonk snapshot list/delete functionality (Green phase)** - 🟡 Pending
43g. **Refactor snapshot system with metadata and cross-platform support (Refactor phase)** - 🟡 Pending

**9. Task Group 51 - Cross-Platform Support (Tasks 51a-51e):**
- macOS focus seems primary
- Can add later if user base expands
- *Needs detailed specs*

51a. **Add Windows PowerShell profile tests for cross-platform support (Red phase)** - 🟡 Pending
51b. **Implement Windows PowerShell profile generation (Green phase)** - 🟡 Pending
51c. **Add Linux distribution package manager tests (Red phase)** - 🟡 Pending
51d. **Implement Linux distribution package manager support (Green phase)** - 🟡 Pending
51e. **Refactor cross-platform support with unified configuration (Refactor phase)** - 🟡 Pending

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
- **Git Operations** - Pure Go git operations with mockable interface for testing
- **Configuration Management** - Automatic application of package-specific configurations
- **Foundational Setup** - Automated installation of prerequisite tools
- **Intelligent Workflows** - Smart detection of existing repositories and installed packages
- **ZSH Configuration Generation** - Complete .zshrc and .zshenv file generation from YAML config
- **Shell Best Practices** - Proper separation of environment variables (.zshenv) and interactive config (.zshrc)
- **Automated Backup System** - Smart backup detection with --backup flag and standalone backup command
- **Configurable Backup Management** - Timestamped backups with automatic cleanup and retention policies