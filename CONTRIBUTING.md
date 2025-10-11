# Contributing to Plonk

Thank you for your interest in contributing to Plonk! This guide will help you get started with development, understand the codebase, and make meaningful contributions.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Contributions](#making-contributions)
- [Testing](#testing)
- [Documentation](#documentation)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Community Guidelines](#community-guidelines)

## Getting Started

### Prerequisites

- **Go 1.23+** (Go 1.24+ also works)
- **Homebrew** (required prerequisite for plonk)
- **Git**
- **just** (recommended for build tasks) - `brew install just`

### Quick Start

```bash
# Clone the repository
git clone https://github.com/richhaase/plonk.git
cd plonk

# Set up development environment (installs dependencies and tools)
just dev-setup

# Run tests to ensure everything works
go test ./...

# Build and install locally for testing
just build
# Binary will be available in bin/plonk

# Or build and install to system
just install
```

## Development Setup

### Using just (Recommended)

The project uses `just` for common development tasks:

```bash
just dev-setup    # Install development dependencies and tools
just build        # Build the binary to bin/plonk
just install      # Build and install to system
just test         # Run all tests
just lint         # Run linters
just clean        # Clean build artifacts
```

### Manual Setup

If you prefer not to use `just`:

```bash
# Build
go build -o bin/plonk ./cmd/plonk

# Install
go install ./cmd/plonk

# Test
go test ./...
```

## Project Structure

Understanding the codebase architecture will help you contribute effectively:

```
plonk/
├── cmd/plonk/              # Main application entry point
├── internal/
│   ├── commands/           # CLI command implementations
│   ├── resources/          # Core resource management
│   │   ├── packages/       # Package manager implementations
│   │   └── dotfiles/       # Dotfile management
│   ├── orchestrator/       # Coordination and reconciliation
│   ├── config/             # Configuration management
│   ├── lock/               # Lock file handling
│   ├── clone/              # Repository cloning logic
│   ├── diagnostics/        # Health checks
│   └── output/             # Output formatting
├── docs/                   # Documentation
└── tests/                  # Integration tests
```

### Key Concepts

- **Resources**: Packages and dotfiles are treated as resources with common interfaces
- **State Reconciliation**: Plonk compares desired state vs actual state
- **Package Managers**: Abstracted through common interfaces for extensibility
- **Lock File**: Tracks package state in `plonk.lock`
- **Filesystem as State**: Dotfile state is represented by `$PLONK_DIR` structure

For detailed architecture information, see [docs/architecture.md](docs/architecture.md) and [docs/code-map.md](docs/code-map.md).

## Making Contributions

### Types of Contributions

We welcome various types of contributions:

1. **Bug Fixes** - Fix issues in existing functionality
2. **New Features** - Add new commands, package managers, or capabilities
3. **Documentation** - Improve docs, add examples, or fix typos
4. **Testing** - Add tests, improve test coverage, or fix flaky tests
5. **Performance** - Optimize existing code or improve resource usage
6. **Developer Experience** - Improve build tools, CI/CD, or development workflow

### Finding Work

- Check [GitHub Issues](https://github.com/richhaase/plonk/issues) for open tasks
- Look for issues labeled `good first issue` for beginner-friendly work
- Check the [project roadmap](docs/why-plonk.md#goals) for larger initiatives
- Browse `TODO` comments in the codebase for improvement opportunities

### Adding New Package Managers

Plonk supports 8 built-in package managers, plus the ability to add custom managers via YAML configuration. Most package managers can be added through YAML without writing Go code.

#### Adding a Manager via YAML Configuration

The easiest way to add a new package manager is by defining it in `plonk.yaml`:

```yaml
managers:
  apt:
    binary: apt
    list:
      command: [apt, list, --installed]
      parse: regex
    install:
      command: [sudo, apt, install, -y]
    upgrade:
      command: [sudo, apt, upgrade, -y]
    upgrade_all:
      command: [sudo, apt, upgrade, -y]
    uninstall:
      command: [sudo, apt, remove, -y]
```

**Required Fields:**
- `binary`: Binary name (for health checks)
- `list.command`: Command to list installed packages
- `list.parse`: Parser type (`regex`, `simple`, or `json`)
- `install.command`: Command to install packages (package name appended)
- `uninstall.command`: Command to uninstall packages

**Optional Fields:**
- `upgrade.command`: Per-package upgrade command
- `upgrade_all.command`: Command to upgrade all packages
- `install.idempotent_errors`: Error patterns that indicate package already installed
- `uninstall.idempotent_errors`: Error patterns that indicate package already removed
- `list.json_field`: For JSON output, which field contains package names

**Parser Types:**
- `simple`: Extract first word from each line (e.g., `npm list`)
- `regex`: Extract package names from formatted output (e.g., `brew list`)
- `json`: Parse JSON output using specified field path

**Contributing YAML Manager Definitions:**

To contribute a new YAML manager configuration:

1. **Test locally** by adding the manager config to your `plonk.yaml`
2. **Verify functionality**: Test install, uninstall, list, and upgrade operations
3. **Add to default managers** in `internal/config/default_managers.go`
4. **Submit PR** with the manager definition and example usage

#### When to Write Go Code

You only need to write Go code for a package manager if:

1. **Complex parsing** - Manager output requires custom parsing logic beyond regex/JSON
2. **Special operations** - Manager needs unique installation/upgrade workflows
3. **Binary dependencies** - Manager requires detecting or installing its own binary programmatically
4. **Advanced features** - Manager needs capabilities not supported by YAML config

For Go implementations:

1. **Implement PackageManager interface** in `internal/resources/packages/`
2. **Add core operations**: Install, Uninstall, ListInstalled, Upgrade, CheckHealth
3. **Register the manager**: Use `RegisterManagerV2` with executor injection pattern
4. **Write comprehensive tests** for all functionality
5. **Update documentation** with examples and supported operations

See `internal/resources/packages/` directory for existing Go implementations (brew, npm, pnpm, cargo, pipx, conda, gem, uv).

### Adding New Commands

To add a new command:

1. **Create command file** in `internal/commands/`
2. **Implement the command logic** following existing patterns
3. **Add output formatting** support (table, JSON, YAML)
4. **Register with root command** in `root.go`
5. **Add command completion** if applicable
6. **Write tests** for the command
7. **Add documentation** in `docs/cmds/`
8. **Update CLI reference** in `docs/cli.md`

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/resources/packages

# Run with coverage
go test -cover ./...
```

### Test Structure

- **Unit tests**: Alongside implementation files (`*_test.go`)
- **Integration tests**: In `tests/` directory
- **Test helpers**: In `internal/testutil/`

### Writing Tests

- Follow Go testing conventions
- Use table-driven tests where appropriate
- Mock external dependencies
- Test both success and error cases
- Include edge cases and boundary conditions

Example test structure:
```go
func TestPackageManager_Install(t *testing.T) {
    tests := []struct {
        name        string
        packageName string
        wantErr     bool
    }{
        {"valid package", "ripgrep", false},
        {"invalid package", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Documentation

### Types of Documentation

1. **Code documentation**: Inline comments and Go doc strings
2. **Command documentation**: Detailed docs in `docs/cmds/`
3. **Architecture documentation**: High-level design docs
4. **User guides**: Installation, configuration, and usage guides

### Writing Documentation

- Use clear, concise language
- Include practical examples
- Update relevant docs when changing functionality
- Follow existing documentation structure and style
- Test all code examples to ensure they work

### Required Documentation Updates

When making changes, update:
- Command documentation if adding/changing commands
- Architecture docs if changing core design
- Configuration docs if adding new settings
- CLI reference for new flags or options

## Code Style

### Go Style Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (runs automatically in most editors)
- Follow Go naming conventions
- Write clear, self-documenting code
- Prefer explicit error handling over panics

### Project-Specific Conventions

- **Error handling**: Return structured results with per-item status
- **Context usage**: Pass context through all layers for cancellation
- **Output formatting**: Support table, JSON, and YAML formats
- **Configuration**: Use sensible defaults, make everything configurable
- **Resource abstraction**: Implement common interfaces for extensibility

### Code Organization

- Keep packages focused and cohesive
- Use interfaces to define contracts
- Separate business logic from CLI concerns
- Put shared utilities in appropriate packages
- Follow the established directory structure

## Submitting Changes

### Before Submitting

1. **Run tests**: Ensure all tests pass (`go test ./...`)
2. **Run linters**: Fix any linting issues (`just lint` if available)
3. **Test manually**: Verify your changes work as expected
4. **Update documentation**: Include relevant doc updates
5. **Check formatting**: Ensure code is properly formatted

### Pull Request Process

1. **Fork the repository** and create a feature branch
2. **Make your changes** following the guidelines above
3. **Write clear commit messages** describing what and why
4. **Push to your fork** and create a pull request
5. **Describe your changes** in the PR description:
   - What the change does
   - Why it's needed
   - How to test it
   - Any breaking changes

### Pull Request Template

```
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Other (please describe)

## Testing
- [ ] Tests pass locally
- [ ] Added tests for new functionality
- [ ] Manually tested the changes

## Documentation
- [ ] Updated relevant documentation
- [ ] Added command examples if applicable

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added where needed
```

### Commit Message Guidelines

Use clear, descriptive commit messages:

```
feat: add support for Nix package manager

- Implement NixPackageManager with Install/Uninstall operations
- Add health checks and self-installation support
- Include tests and documentation updates

Resolves #123
```

Format: `type: brief description`

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `chore`

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive in all interactions
- Focus on constructive feedback and collaboration
- Help newcomers get started with contributing
- Follow the [Go Community Code of Conduct](https://golang.org/conduct)

### Getting Help

- **Issues**: Use GitHub Issues for bugs and feature requests
- **Discussions**: Use GitHub Discussions for general questions
- **Code Review**: Engage constructively in pull request reviews

### AI-Friendly Development

Plonk is built with AI-assisted development in mind:
- Clear interfaces and minimal magic
- Straightforward patterns throughout the codebase
- Rich documentation and context
- Well-structured code that's easy to understand and extend

This makes it easier for both humans and AI to contribute effectively.

## Additional Resources

- [Architecture Documentation](docs/architecture.md)
- [Code Map](docs/code-map.md)
- [CLI Reference](docs/cli.md)
- [Configuration Guide](docs/configuration.md)
- [Why Plonk?](docs/why-plonk.md)

## Questions?

If you have questions about contributing, please:
1. Check existing documentation
2. Search GitHub Issues for similar questions
3. Create a new GitHub Discussion
4. Mention maintainers in your issue if urgent

Thank you for contributing to Plonk! 🚀
