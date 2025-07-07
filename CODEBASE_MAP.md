# Plonk Codebase Navigation Map

This document provides a comprehensive guide for navigating and understanding the Plonk codebase.

## 📊 Project Statistics

- **Total Go Files**: ~30+ files
- **Estimated Lines**: 2000+ lines
- **Test Coverage**: High (all major components have tests)
- **Packages**: 5 main packages
- **Commands**: 8 CLI commands

## 🏗️ Directory Structure

```
plonk/
├── cmd/plonk/                     # 📍 Entry Point
│   └── main.go                    # CLI application entry point
├── internal/                      # 🔒 Private Application Code
│   ├── commands/                  # 🎯 CLI Command Implementations
│   ├── directories/               # 📁 Directory Management
│   └── utils/                     # 🛠️ Utility Functions
├── pkg/                          # 📦 Public Packages
│   ├── config/                   # ⚙️ Configuration Management
│   └── managers/                 # 📋 Package Manager Abstractions
├── .tool-versions                # 🔧 ASDF development tools
├── .golangci.yml                 # 🔍 Linting configuration
├── magefile.go                   # ⚡ Mage task runner (Go-native build tool)
├── go.mod                        # 🔧 Go Module Definition
├── go.sum                        # 🔒 Dependency Lock File
├── CLAUDE.md                     # 📖 Project Documentation
└── CODEBASE_MAP.md              # 🗺️ This Navigation Guide
```

## 🎯 Key Files by Function

### 🚀 **Entry Points & Core**
- `cmd/plonk/main.go` - Application entry point
- `internal/commands/root.go` - Root command and CLI structure

### 🎮 **CLI Commands** (internal/commands/)
- `status.go` - Show package manager availability and counts
- `pkg.go` - List packages (`plonk pkg list`)
- `clone.go` - Clone dotfiles repositories 
- `pull.go` - Pull repository updates
- `install.go` - Install packages from config
- `apply.go` - Apply configuration files
- `setup.go` - Install foundational tools
- `repo.go` - Complete repository setup workflow
- `backup.go` - Backup functionality

### 🔧 **Core Infrastructure**
- `internal/commands/error_handling.go` - Standardized error patterns
- `internal/commands/package_installer.go` - Installation helpers
- `internal/directories/manager.go` - Centralized path management
- `internal/commands/test_helpers.go` - Shared test utilities

### ⚡ **Development Infrastructure**
- `magefile.go` - Mage task runner commands (build, test, lint, format, clean)
- `.tool-versions` - ASDF development tools specification
- `.golangci.yml` - Linting and formatting configuration

### ⚙️ **Configuration System** (pkg/config/)
- `yaml_config.go` - Primary YAML configuration parsing
- `validator.go` - Configuration validation
- `zsh_generator.go` - ZSH configuration file generation
- `git_generator.go` - Git configuration file generation

### 📋 **Package Managers** (pkg/managers/)
- `common.go` - Shared interfaces and command runner
- `executor.go` - Real command execution
- `homebrew.go` - Homebrew package manager
- `asdf.go` - ASDF version manager
- `npm.go` - NPM package manager
- `zsh.go` - ZSH shell management

### 🧪 **Testing**
- `*_test.go` - Test files (following Go convention)
- Most components have comprehensive test coverage

## 🔍 Quick File Finder

### Looking for...
- **CLI command logic**: `internal/commands/{command_name}.go`
- **Package manager code**: `pkg/managers/{manager_name}.go`
- **Configuration parsing**: `pkg/config/yaml_config.go`
- **Error handling**: `internal/commands/error_handling.go`
- **Test helpers**: `internal/commands/test_helpers.go`
- **Installation logic**: `internal/commands/package_installer.go`

### Common Tasks
- **Add new CLI command**: Create `internal/commands/new_command.go` + tests
- **Add package manager**: Implement in `pkg/managers/` + update interfaces
- **Modify config format**: Update `pkg/config/yaml_config.go` + validation
- **Fix error messages**: Check `internal/commands/error_handling.go`

## 🔗 Key Interfaces & Types

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

### Configuration Types
```go
// pkg/config/yaml_config.go
type Config struct {
    Settings Settings `yaml:"settings"`
    Dotfiles []string `yaml:"dotfiles"`
    Homebrew HomebrewConfig `yaml:"homebrew"`
    ASDF []ASDFTool `yaml:"asdf"`
    NPM []NPMPackage `yaml:"npm"`
    // ... other fields
}
```

## 🧭 Navigation Patterns

### 1. **Understanding a Command**
1. Start with `internal/commands/{command}.go`
2. Look at corresponding `*_test.go` for examples
3. Check if it uses package managers (`pkg/managers/`)
4. See if it loads config (`pkg/config/`)

### 2. **Understanding Package Management**
1. Start with `pkg/managers/common.go` for interfaces
2. Look at specific manager (e.g., `homebrew.go`)
3. Check `internal/commands/install.go` for usage
4. Review `internal/commands/package_installer.go` for helpers

### 3. **Understanding Configuration**
1. Start with `pkg/config/yaml_config.go` for structure
2. Check `validator.go` for validation rules
3. Look at `internal/commands/apply.go` for usage
4. Review generators (`zsh_generator.go`, `git_generator.go`)

### 4. **Understanding Tests**
1. Look for `*_test.go` files alongside implementation
2. Check `internal/commands/test_helpers.go` for common patterns
3. Review `pkg/managers/manager_test.go` for comprehensive examples

## 🎨 Common Patterns

### Error Handling
```go
// Standardized error wrapping
return WrapConfigError(err)
return WrapInstallError(packageName, err)
return WrapPackageManagerError("homebrew", err)
```

### Argument Validation
```go
// Consistent validation patterns
if err := ValidateNoArgs("command", args); err != nil {
    return err
}
```

### Test Setup
```go
// Consistent test environment
tempHome, cleanup := setupTestEnv(t)
defer cleanup()
```

### Package Installation
```go
// Consistent package handling
displayName := getPackageDisplayName(pkg)
if shouldInstallPackage(name, isInstalled) {
    // install logic
}
```

## 🚀 Getting Started Guide

### For New Contributors
1. **Read**: `CLAUDE.md` for project overview
2. **Understand**: This `CODEBASE_MAP.md` for navigation
3. **Explore**: Start with `internal/commands/status.go` (simple command)
4. **Run Tests**: `go test ./...` to verify environment
5. **Build**: `go build ./cmd/plonk` to create binary

### For Adding Features
1. **Plan**: Follow TDD pattern (Red-Green-Refactor-Commit-Update Memory)
2. **Test First**: Create `*_test.go` with failing tests
3. **Implement**: Write minimal code to pass tests
4. **Refactor**: Improve while keeping tests green
5. **Document**: Update relevant documentation

### For Debugging
1. **Error Messages**: Check `internal/commands/error_handling.go`
2. **Package Issues**: Look at `pkg/managers/{manager}.go`
3. **Config Problems**: Check `pkg/config/yaml_config.go` and `validator.go`
4. **Test Failures**: Review test files and `test_helpers.go`

## 🔧 Development Tools

### Mage Commands (Primary)
```bash
# See all available commands
mage -l

# Core development tasks
mage build        # Build the plonk binary
mage test         # Run all tests
mage lint         # Run linter
mage format       # Format code (gofmt)
mage clean        # Clean build artifacts

# Release management
mage preparerelease  # Analyze commits and suggest versions
mage nextpatch       # Suggest next patch version
mage nextminor       # Suggest next minor version
mage nextmajor       # Suggest next major version
mage release v1.0.0  # Create release with changelog and git tag

# Development workflow (manual)
mage format && mage lint && mage test
```

### Traditional Go Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/commands/...

# Build the application
go build ./cmd/plonk

# Format code
go fmt ./...

# Static analysis
go vet ./...
```

### File Structure Queries
```bash
# Find all Go files
find . -name "*.go" | grep -v vendor

# Find test files
find . -name "*_test.go"

# Count lines of code
find . -name "*.go" -not -path "./vendor/*" | xargs wc -l

# Find function definitions
grep -r "func " --include="*.go" .
```

## 📝 Notes for Maintainers

- **TDD Required**: All changes must follow Red-Green-Refactor pattern
- **Test Coverage**: Maintain high test coverage for reliability  
- **Error Consistency**: Use standardized error handling functions
- **Documentation**: Keep CLAUDE.md updated with completed tasks
- **Code Quality**: Follow established patterns and conventions

---

*This map is maintained as the codebase evolves. Last updated: Maintenance Phase 2024*