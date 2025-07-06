# Plonk - Shell Environment Lifecycle Manager

## Project Overview

Plonk is a CLI tool for managing shell environments across multiple machines. It helps you manage package installations and environment switching using a focused set of package managers:

- **Homebrew** - Primary package installation
- **ASDF** - Programming language tools and versions
- **NPM** - Packages not available via Homebrew (like claude-code)

## Development Approach

This project was developed using **Test-Driven Development (TDD)** with Red-Green-Refactor cycles throughout the implementation.

### REQUIRED TDD WORKFLOW
**IMPORTANT**: All changes to this codebase MUST follow the TDD pattern:
1. **RED**: Write failing tests first
2. **GREEN**: Write minimal code to make tests pass
3. **REFACTOR**: Improve code while keeping tests green
4. **COMMIT**: Commit the changes
5. **UPDATE MEMORY**: Update CLAUDE.md to reflect completed work

This is a strict requirement for maintaining code quality and consistency.

## Architecture

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
./plonk apply --dry-run          # Show what would be applied without making changes
./plonk apply --backup --dry-run # Preview what would be applied with backup

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

33. **Complete Phase 1 codebase cleanup (Task 52a-52c)** - ✅ Completed
    - **Task 52a**: Removed legacy TOML configuration system entirely (130+ lines of dead code)
    - **Task 52b**: Consolidated package manager interfaces into single location (pkg/managers/common.go)
    - **Task 52c**: Renamed YAML config types to be primary (YAMLConfig → Config, LoadYAMLConfig → LoadConfig)
    - Eliminated duplicate PackageManager interface definitions between command files
    - Updated all function signatures and references across codebase
    - YAML is now the pure configuration format after TOML removal
    - Improved code organization and eliminated architectural duplication

34. **Create centralized DirectoryManager for all path operations (Task 52d)** - ✅ Completed
    - Built centralized DirectoryManager with caching for performance optimization
    - Replaced scattered directory functions (getPlonkDir, getRepoDir, getBackupsDir, expandHomeDir) 
    - Added Reset() method for proper test isolation with environment variables
    - Updated all commands to use directories.Default instead of individual functions
    - Removed redundant internal/commands/directory.go file entirely
    - Fixed test failures by implementing proper cleanup in test environments
    - Followed strict TDD methodology: Red-Green-Refactor cycle with comprehensive tests

35. **Consolidate test helper functions to eliminate repetitive patterns (Task 52f)** - ✅ Completed
    - Created centralized test setup helpers in internal/commands/test_helpers.go
    - setupTestEnv(t) for basic HOME environment isolation with cleanup
    - setupTestEnvWithPlonkDir(t, dir) for tests requiring custom PLONK_DIR
    - Updated 33+ test functions across 10 files to use helpers instead of 6-8 lines of repetitive setup
    - Eliminated repetitive patterns: t.TempDir(), os.Getenv("HOME"), defer cleanup, os.Setenv()
    - Improved maintainability while preserving test isolation and functionality
    - All tests continue to pass with cleaner, more consistent code organization

36. **Refactor validation system with unified error reporting (Task 47a7)** - ✅ Completed
    - Replaced complex custom validation system with go-playground/validator library
    - Added struct validation tags to all Config types for declarative validation
    - Created SimpleValidator with custom validators for package_name and file_path
    - Simplified ValidationResult to just Errors and Warnings (removed complex types)
    - Integrated validation seamlessly into LoadConfig function
    - Reduced codebase from 280+ lines of complex code to ~160 lines using standard library
    - All tests passing with improved maintainability and industry-standard patterns

37. **Add dry-run flag support to apply command (Tasks 47c1-47c2)** - ✅ Completed
    - **Task 47c1**: Created focused unit tests for --dry-run flag parsing (Red phase)
    - **Task 47c2**: Implemented comprehensive dry-run functionality (Green phase)
    - Added --dry-run flag to apply command alongside existing --backup flag
    - Implemented preview functions for all configuration types:
      - previewAllConfigurations for full config preview
      - previewPackageConfiguration for package-specific preview
      - previewDotfiles, previewZSHConfiguration, previewGitConfiguration
      - previewPackageConfigurations for all package configs
    - Used clear visual indicators: ✨ (new files), 📝 (overwrites), 📁 (directories), ⚠️ (warnings)
    - Dry-run mode prevents actual file modifications while showing what would happen
    - Supports both full config apply and package-specific preview
    - All tests passing with proper unit and integration test coverage

38. **Refactor package installation logic to eliminate duplication (Task 52e)** - ✅ Completed
    - Followed strict TDD workflow: Red-Green-Refactor-Commit-Update Memory
    - Created helper functions to reduce code duplication in package installation:
      - `extractInstalledPackages()` - combines package lists from all managers
      - `shouldInstallPackage()` - consistent install checks across managers
      - `getPackageDisplayName()` - unified package name formatting (handles NPM scoped packages)
      - `getPackageConfig()` - extracts config path from any package type
      - `getPackageName()` - gets base package name for config tracking
    - Updated all install functions (Homebrew, ASDF, NPM) to use helpers consistently
    - Reduced code duplication while maintaining same functionality
    - All tests passing with comprehensive test coverage for new helpers

39. **Standardize error handling patterns across commands (Task 52g)** - ✅ Completed
    - Followed strict TDD workflow: Red-Green-Refactor-Commit-Update Memory
    - **Standardized Error Wrapping Functions:**
      - `WrapConfigError()` - consistent configuration loading error messages
      - `WrapPackageManagerError()` - standardized package manager availability errors
      - `WrapInstallError()` - uniform package installation error handling
      - `WrapFileError()` - consistent file operation error messages
    - **Standardized Argument Validation:**
      - `ValidateNoArgs()` - commands that take no arguments
      - `ValidateExactArgs()` - commands requiring specific argument count
      - `ValidateMaxArgs()` - commands with maximum argument limits
      - Proper singular/plural grammar handling in error messages
    - **Refactored All Commands:** install, apply, setup, pull, repo, clone
    - **Benefits:** Consistent error format, proper error chaining, improved debugging
    - All tests passing with comprehensive error handling test coverage

40. **Setup development infrastructure for code quality (Maintenance Phase)** - ✅ Completed
    - **Development Tools Management:**
      - Created `.tool-versions` with golang 1.24.4, golangci-lint 2.2.1, just 1.41.0
      - Setup ASDF-based tool management for reproducible development environment
    - **Linting and Formatting Infrastructure:**
      - Configured golangci-lint v2 with proper formatters and linters
      - Automated import organization via goimports integration
      - Found 59 errcheck issues ready for systematic fixing
    - **Task Runner Implementation:**
      - Created `justfile` with development workflow commands
      - Automated Task 52h (import organization) via `just format`
      - Implemented `just dev` (format+lint+test) and `just ci` (full pipeline)
    - **Updated Documentation:**
      - Enhanced CODEBASE_MAP.md with development infrastructure
      - Documented justfile commands and workflow
    - **Benefits:** Automated code quality, reproducible development environment, simplified workflows

### 🔄 Current Pending Tasks (Reconsidered by Value, Complexity, Dependencies)

**Prioritization Strategy**: Balanced approach considering implementation simplicity, user value, dependencies, and project impact. Focus on quick wins that improve codebase quality and provide immediate user benefits.

---

## 🎯 **TIER 1: HIGH VALUE + LOW COMPLEXITY + NO DEPENDENCIES** 
*Quick wins that immediately improve user experience*

### **Group A: Code Quality & Cleanup (Simple Infrastructure)**
**Value**: 🟢 **High** | **Complexity**: 🟢 **Low** | **Dependencies**: 🟢 **None**

52h. **Organize imports consistently across all files** - 🎯 **NEXT PRIORITY**
52i. **Standardize function documentation**
52j. **Convert remaining tests to table-driven format**

**Why Tier 1**: Simple refactoring tasks that improve maintainability with minimal risk. Can be done independently and make future development easier.

### **Group B: Diff Command (Builds on Existing Infrastructure)**
**Value**: 🟢 **High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

47e1. **Add tests for diff command structure (Red phase)**
47e2. **Implement basic diff command (Green phase)**
47e3. **Add tests for file content comparison (Red phase)**
47e4. **Implement file diff logic (Green phase)**
47e5. **Add tests for config state comparison (Red phase)**
47e6. **Implement config vs reality diff (Green phase)**
47e7. **Refactor with colored diff output (Refactor phase)**

**Why Tier 1**: Builds directly on existing drift detection and dry-run work. High user value for seeing configuration differences. Well-defined scope.

---

## 🚀 **TIER 2: HIGH VALUE + MEDIUM COMPLEXITY + SOME DEPENDENCIES**
*Significant user value with manageable implementation*

### **Group C: Integration Testing (Foundation)**
**Value**: 🟢 **High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟡 **Should complete Tier 1 first**

46a. **Add integration tests for end-to-end workflows (Red phase)**
46b. **Implement comprehensive integration test suite (Green phase)**
46c. **Refactor integration tests with CI/CD support (Refactor phase)**

**Why Tier 2**: Critical for stability but needs existing codebase to be clean first. High value for preventing regressions.

### **Group D: Additional Shell Support (Natural Extension)**
**Value**: 🟡 **Medium-High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

49a. **Add Bash shell config generation tests (Red phase)**
49b. **Implement Bash shell config generation functionality (Green phase)**
49c. **Add Fish shell config generation tests (Red phase)**
49d. **Implement Fish shell config generation functionality (Green phase)**
49e. **Refactor shell config generation with multi-shell support (Refactor phase)**

**Why Tier 2**: Natural extension of existing ZSH work. Clear user value for non-ZSH users. Well-defined patterns to follow.

### **Group E: Import Command (User Onboarding)**
**Value**: 🟡 **Medium-High** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

38a. **Add shell config parsing tests for common formats (Red phase)**
38b. **Implement basic .zshrc/.bashrc parsing functionality (Green phase)**
38c. **Add tests for plonk.yaml generation from parsed configs (Red phase)**
38d. **Implement plonk import command with YAML suggestion (Green phase)**
38e. **Refactor import command with support for multiple shell types (Refactor phase)**

**Why Tier 2**: High value for user adoption. One-time use but critical for migration to plonk.

---

## ⚡ **TIER 3: COMPLEX BUT HIGH IMPACT**
*Advanced features requiring significant implementation*

### **Group F: Watch Mode (Complex File Operations)**
**Value**: 🟡 **Medium** | **Complexity**: 🔴 **High** | **Dependencies**: 🟡 **Should have stable foundation**

47g1. **Add tests for watch command structure (Red phase)**
47g2. **Implement basic watch command (Green phase)**
47g3. **Add tests for file change detection (Red phase)**
47g4. **Implement file watcher (Green phase)**
47g5. **Add tests for auto-apply on change (Red phase)**
47g6. **Implement auto-apply logic (Green phase)**
47g7. **Refactor with debouncing and error handling (Refactor phase)**

**Why Tier 3**: Complex file watching, debouncing, error handling. High complexity with moderate value. Needs stable base.

### **Group G: Repository Infrastructure (DevOps Setup)**
**Value**: 🟡 **Medium** | **Complexity**: 🔴 **High** | **Dependencies**: 🔴 **Needs stable codebase**

44a. **Add pre-commit hook tests for Go formatting (Red phase)**
44b. **Implement pre-commit hooks for Go formatting (Green phase)**
44c. **Add linting tests with golangci-lint (Red phase)**
44d. **Implement golangci-lint configuration and hooks (Green phase)**
44e. **Refactor code quality setup with development workflow integration (Refactor phase)**
44f. **Add development workflow tests (Red phase)**
44g. **Implement development workflow tool (Green phase)**
44h. **Add test coverage enforcement tests (Red phase)**
44i. **Implement test coverage tooling (Green phase)**
44j. **Refactor development workflow with documentation and optimization (Refactor phase)**

**Why Tier 3**: Important for project health but complex setup. Should wait until core functionality is stable.

---

## 🎁 **TIER 4: NICE-TO-HAVE ENHANCEMENTS**
*Lower priority features with specific use cases*

### **Group H: Advanced Backup Features**
**Value**: 🟡 **Low-Medium** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

48a. **Add selective restore tests for granular file restoration (Red phase)**
48b. **Implement selective restore functionality (Green phase)**
48c. **Add backup compression tests for space optimization (Red phase)**
48d. **Implement backup compression functionality (Green phase)**
48e. **Add remote backup tests for cloud sync (Red phase)**
48f. **Implement remote backup sync functionality (Green phase)**
48g. **Add backup encryption tests for sensitive data protection (Red phase)**
48h. **Implement backup encryption functionality (Green phase)**
48i. **Refactor advanced backup features with unified management (Refactor phase)**

### **Group I: Cross-Platform Support**  
**Value**: 🟡 **Low** | **Complexity**: 🔴 **High** | **Dependencies**: 🟢 **None**

51a. **Add Windows PowerShell profile tests for cross-platform support (Red phase)**
51b. **Implement Windows PowerShell profile generation (Green phase)**
51c. **Add Linux distribution package manager tests (Red phase)**
51d. **Implement Linux distribution package manager support (Green phase)**
51e. **Refactor cross-platform support with unified configuration (Refactor phase)**

### **Group J: Package Manager Extensions**
**Value**: 🔴 **Low** | **Complexity**: 🟡 **Medium** | **Dependencies**: 🟢 **None**

50a. **Add mas command support tests for Mac App Store integration (Red phase)**
50b. **Implement mas command functionality for App Store apps (Green phase)**
50c. **Refactor package manager integration with mas support (Refactor phase)**

### **Group K: Environment Snapshots**
**Value**: 🔴 **Low** | **Complexity**: 🔴 **High** | **Dependencies**: 🟢 **None**

43a. **Add full environment snapshot tests (Red phase)**
43b. **Implement plonk snapshot create functionality (Green phase)**
43c. **Add snapshot restoration tests (Red phase)**
43d. **Implement plonk snapshot restore functionality (Green phase)**
43e. **Add snapshot management tests (list, delete) (Red phase)**
43f. **Implement plonk snapshot list/delete functionality (Green phase)**
43g. **Refactor snapshot system with metadata and cross-platform support (Refactor phase)**

### **Group L: Optional Enhancements (Lower Priority)**
47c3. **Add tests for dry-run preview output formatting (Red phase)** - 🟡 Optional (core functionality complete)
47c4. **Implement enhanced dry-run preview logic (Green phase)** - 🟡 Optional (core functionality complete)
47c5. **Refactor with improved preview formatting (Refactor phase)** - 🟡 Optional (core functionality complete)
47i1. **Refactor all developer experience features with unified CLI patterns** - 🟡 Pending

---

## 🎯 **RECOMMENDED EXECUTION ORDER**

### **Phase 1: Foundation (Tier 1) - 3-5 weeks**
1. **Group A**: Code Quality & Cleanup (5 tasks) - 1-2 weeks
2. **Group B**: Diff Command (7 tasks) - 2-3 weeks

**Rationale**: Quick wins that improve codebase quality and provide immediate user value.

### **Phase 2: Core Extensions (Tier 2) - 6-8 weeks** 
3. **Group C**: Integration Testing (3 tasks) - 2 weeks
4. **Group D**: Additional Shell Support (5 tasks) - 2-3 weeks  
5. **Group E**: Import Command (5 tasks) - 2-3 weeks

**Rationale**: Builds on stable foundation to extend core value proposition.

### **Phase 3+: Advanced Features (Tier 3+)**
6. **Group F**: Watch Mode - if user demand exists
7. **Group G**: Repository Infrastructure - when codebase is mature
8. **Groups H-K**: Nice-to-have features based on user feedback

## 🔄 **CURRENT PHASE: MAINTENANCE & CODEBASE IMPROVEMENT**

**Phase Objective**: Improve codebase maintainability, navigation, and developer experience before continuing with new features.

**Maintenance Tasks (Following TDD: Red→Green→Refactor→Commit→Update Memory):**
1. **Create CODEBASE_MAP.md** - Navigation aid for large codebase ✅ **COMPLETED**
2. **Infrastructure-First Code Quality Bundle:**
   - **Setup local development tools with asdf (.tool-versions)** ✅ **COMPLETED**
   - **Setup golangci-lint configuration** ✅ **COMPLETED**  
   - **Create justfile for common tasks** ✅ **COMPLETED**
   - **Add pre-commit hooks for Go formatting** ⚡ **NEXT**
   - **Task 52i**: Standardize function documentation  
   - **Task 52j**: Convert remaining tests to table-driven format
3. **Development utilities**: Create helper functions for codebase analysis
4. **Key files reference**: Document critical files and their purposes

**Infrastructure Status:**
- ✅ golangci-lint v2.2.1 configured and working (found 59 errcheck issues)
- ✅ justfile with dev tasks: build, test, lint, format, ci, dev workflow
- ✅ .tool-versions: golang 1.24.4, golangci-lint 2.2.1, just 1.41.0
- 🔧 Task 52h (import organization) automated via `just format` command

**Rationale**: With 39 completed tasks and 2000+ lines of code across multiple packages, the codebase needs better organization and documentation to maintain development velocity.

**Next Phase**: Return to feature development with improved maintainability.

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