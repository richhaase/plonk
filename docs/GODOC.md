# API Documentation Guide

Go documentation generation and usage for plonk's internal APIs.

## Overview

Plonk uses Go's built-in documentation tools to generate comprehensive API documentation from source code comments. This documentation is essential for contributors and AI agents working with the codebase.

## Generating Documentation

### Local Documentation Server

```bash
# Start local godoc server
godoc -http=:6060

# Open in browser
open http://localhost:6060/pkg/github.com/your-username/plonk/
```

### Command Line Documentation

```bash
# Show package documentation
go doc ./internal/config

# Show specific function documentation
go doc ./internal/config Config.ReadConfig

# Show all exported items
go doc -all ./internal/managers
```

## Documentation Structure

### Package-Level Documentation

Each package should have comprehensive package documentation:

```go
// Package config provides configuration management for plonk.
//
// This package implements the core configuration interfaces for loading,
// validating, and managing plonk configuration files. It supports YAML
// configuration with environment variable overrides and auto-discovery
// of dotfiles.
//
// Key interfaces:
//   - ConfigReader: Load configuration from storage
//   - ConfigWriter: Save configuration to storage
//   - ConfigValidator: Validate configuration correctness
//
// Example usage:
//
//     config, err := yaml_config.NewYAMLConfigService()
//     if err != nil {
//         return err
//     }
//
//     cfg, err := config.ReadConfig()
//     if err != nil {
//         return err
//     }
//
package config
```

### Function Documentation

All exported functions should have documentation:

```go
// ReadConfig loads the plonk configuration from the default location.
//
// It first checks for configuration at $PLONK_DIR/plonk.yaml, then falls
// back to ~/.config/plonk/plonk.yaml. Returns an error if the configuration
// file cannot be found or parsed.
//
// The returned Config struct contains all configuration sections with
// defaults applied for optional fields.
func (y *YAMLConfigService) ReadConfig() (*Config, error) {
    // implementation
}
```

### Interface Documentation

Interfaces should clearly document their contract:

For complete interface specifications and usage examples, see the generated API documentation in `docs/api/managers.md`.

## API Documentation by Package

Complete API documentation is auto-generated and available in `docs/api/`:

- **`docs/api/config.md`** - Configuration management interfaces and types
- **`docs/api/state.md`** - State reconciliation engine and provider interfaces
- **`docs/api/managers.md`** - Package manager implementations and interfaces
- **`docs/api/dotfiles.md`** - Dotfile operations and file management
- **`docs/api/errors.md`** - Structured error handling types and codes
- **`docs/api/commands.md`** - CLI command implementations

**Key package relationships:**
- `config` provides configuration loading and validation
- `state` provides unified reconciliation across domains
- `managers` implements package manager interfaces
- `dotfiles` handles file operations and path management
- `errors` provides structured error types across all packages
- `commands` orchestrates operations using other packages

## Documentation Best Practices

### Writing Good Documentation

1. **Start with purpose** - What does this function/type do?
2. **Describe parameters** - What are the inputs and their constraints?
3. **Explain return values** - What is returned and under what conditions?
4. **Document errors** - What errors can occur and why?
5. **Provide examples** - Show typical usage patterns
6. **Mention context** - How does this fit into the larger system?

### Example Documentation Template

```go
// FunctionName does X by performing Y operation.
//
// It takes parameter A which must be non-nil and parameter B which
// should be a valid string. The function returns result C which
// represents the processed data.
//
// Context handling:
//   - Respects context cancellation
//   - Uses context timeout for operations
//   - Returns context.Canceled on cancellation
//
// Error conditions:
//   - Returns ErrInvalidInput for invalid parameters
//   - Returns ErrNotFound if resource doesn't exist
//   - Returns context.Canceled if context is cancelled
//
// Example usage:
//
//     ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//     defer cancel()
//     
//     result, err := FunctionName(ctx, validInput, "parameter")
//     if err != nil {
//         return err
//     }
//     
//     // Use result...
//
func FunctionName(ctx context.Context, a InputType, b string) (ResultType, error) {
    // implementation
}
```

## Integration with Development

### Documentation in Development Workflow

1. **Write documentation** as you write code
2. **Update documentation** when changing APIs
3. **Review documentation** in code reviews
4. **Generate documentation** before releases
5. **Validate examples** in documentation

### Documentation Commands

```bash
# Generate all API documentation (automated)
just generate-docs

# Generate documentation for specific package
go doc -all ./internal/config > docs/api/config.md

# Check documentation coverage
go doc -all ./... | grep -c "^func\|^type"

# Validate documentation examples
go test -run=ExampleFunction ./internal/config
```

### Documentation Testing

Go supports testable examples:

```go
// ExampleYAMLConfigService_ReadConfig demonstrates loading configuration.
func ExampleYAMLConfigService_ReadConfig() {
    service := yaml_config.NewYAMLConfigService()
    config, err := service.ReadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Default manager: %s\n", config.Settings.DefaultManager)
    // Output: Default manager: homebrew
}
```

## AI Agent Integration

### Structured Documentation Access

AI agents can access documentation through:

```bash
# Get interface documentation
go doc -all ./internal/managers | grep -A 10 "type PackageManager"

# Get function signatures
go doc ./internal/config | grep "func"

# Get package overview
go doc ./internal/config
```

### Documentation Parsing

Documentation follows consistent patterns for AI parsing:

- **Function signatures** - Standard Go format
- **Error documentation** - Structured error conditions
- **Example usage** - Consistent code examples
- **Interface contracts** - Clear behavioral expectations

## See Also

- [DEVELOPMENT.md](DEVELOPMENT.md) - Development setup and workflows
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [CODEMAP.md](CODEMAP.md) - Code navigation guide
- [TESTING.md](TESTING.md) - Testing infrastructure