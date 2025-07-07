# Contributing to Plonk

Thank you for your interest in contributing to Plonk! This guide covers the development workflow and standards for both human contributors and AI coding agents.

## Development Philosophy

Plonk follows a strict **Test-Driven Development (TDD)** approach. This ensures code quality, maintainability, and reliable functionality across all components.

## Required Development Workflow

All contributions MUST follow this TDD pattern:

### 1. 🔴 RED Phase
Write failing tests first. This ensures you understand the requirements before implementation.

### 2. 🟢 GREEN Phase  
Write the minimal code necessary to make the tests pass. Resist the urge to add extra functionality.

### 3. 🔵 REFACTOR Phase
Improve the code while keeping all tests green. This is where you enhance readability and performance.

### 4. 📝 COMMIT Phase
Commit your changes with clear, descriptive messages following the project conventions.

### 5. 📚 UPDATE Phase
Update relevant documentation:
- For humans: Update README.md, ARCHITECTURE.md, or other user-facing docs
- For AI agents: Update CLAUDE.md with implementation details
- For active work: Update TODO.md to reflect completed tasks

## Getting Started

### Prerequisites
- **Go 1.24.4 or later** - Install via:
  - [Official installer](https://golang.org/dl/) (recommended)
  - Package managers: `brew install go`, `apt install golang-go`, etc.
  - ASDF: `asdf install` (uses included .tool-versions)
- **Git** for version control

### Development Setup
```bash
# Clone the repository
git clone <repository-url>
cd plonk

# Install Go dependencies (includes all development tools)
go mod download

# Build the project
go run github.com/magefile/mage/mage build

# Run tests
go run github.com/magefile/mage/mage test

# Run the unified pre-commit checks (format, lint, test, security)
go run github.com/magefile/mage/mage precommit

# Individual development tasks
go run github.com/magefile/mage/mage format   # Format code
go run github.com/magefile/mage/mage lint     # Run linter
go run github.com/magefile/mage/mage security # Security checks
```

### Tool Versioning
**Pure Go project** - All tools managed via `go.mod`:
- **Go runtime** - Minimum version 1.24.4+ (specified in go.mod)
- **Development tools** - golangci-lint, gosec, govulncheck, goimports, mage

**Go Installation Options:**
- Any Go installation method works (official installer, package managers, version managers)
- `.tool-versions` provided for ASDF users convenience (optional)

### Installing from Source

Install plonk globally for development or testing:

```bash
# Standard installation from repository
go install ./cmd/plonk

# Verify installation
plonk --help
which plonk
```

**Notes:**
- `go install` automatically installs to the correct location based on your Go setup:
  - If `GOBIN` is set: installs to `$GOBIN/plonk`
  - Otherwise: installs to `$GOPATH/bin/plonk` (typically `~/go/bin/plonk`)
- ASDF users: GOBIN points to the current Go version's bin directory (already in PATH)
- Standard Go users: Ensure `~/go/bin` is in your PATH
- The binary will be named `plonk` and available globally after installation

## Making Changes

### Before Starting
1. Check TODO.md for active work items
2. Review ROADMAP.md for planned features
3. Ensure your change aligns with project goals

### Code Standards
- Follow Go idioms and best practices
- Use meaningful variable and function names
- Keep functions focused and testable
- Maintain consistent error handling patterns

### Testing Requirements
- All new features must have comprehensive tests
- Use table-driven tests where appropriate
- Mock external dependencies (don't execute real commands in tests)
- Aim for high test coverage

### Documentation
- Add inline comments for complex logic
- Update relevant documentation files

## Release Management

The project uses semantic versioning with automated release workflows:

```bash
# Prepare for release (analyze commits and suggest versions)
mage preparerelease

# Quick version suggestions
mage nextpatch    # Bug fixes (v1.0.1)
mage nextminor    # New features (v1.1.0)  
mage nextmajor    # Breaking changes (v2.0.0)

# Create a release (updates changelog, tags, commits)
mage release v1.0.0
```

**Release Guidelines:**
- Follow [semantic versioning](https://semver.org/) (MAJOR.MINOR.PATCH)
- Update CHANGELOG.md [Unreleased] section before release
- Use `mage preparerelease` to analyze commit history
- Test releases with pre-release versions (e.g., `v1.0.0-beta.1`)
- All releases automatically update changelog and create git tags

## Legal and Licensing

This project uses the MIT License for maximum compatibility and adoption:

- **LICENSE file** contains the full MIT License text
- **All Go files** include consistent license headers
- **New files** should include the standard header:
  ```go
  // Copyright (c) 2025 Rich Haase
  // Licensed under the MIT License. See LICENSE file in the project root for license information.
  ```
- **Use `mage addlicenseheaders`** to automatically add headers to new files
- Include examples in function documentation
- Keep CHANGELOG.md updated for significant changes

## Submitting Changes

### Pull Request Process
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the TDD workflow for all changes
4. Ensure all tests pass (`go run github.com/magefile/mage/mage test`)
5. Run the full development pipeline locally (`go run github.com/magefile/mage/mage precommit`)
6. Commit with clear messages (pre-commit hook will run automatically)
7. Push to your fork
8. Open a Pull Request with:
   - Clear description of changes
   - Link to related issues
   - Test coverage information
   - Documentation updates

### Commit Message Format
```
<type>: <subject>

<body>

<footer>
```

Types: feat, fix, docs, style, refactor, test, chore

## Working with AI Agents

If you're an AI coding agent:
1. Always read CLAUDE.md for project-specific guidance
2. Sync with TODO.md at the start of each session
3. Follow the TDD workflow strictly
4. Update TODO.md as you complete tasks
5. Refer to established patterns in the codebase

## Code Review Criteria

PRs will be reviewed for:
- [ ] Adherence to TDD workflow
- [ ] Test coverage and quality
- [ ] Code clarity and maintainability
- [ ] Consistent error handling
- [ ] Documentation completeness
- [ ] Performance considerations

## Getting Help

- Review existing code for patterns and examples
- Check ARCHITECTURE.md for system design decisions
- Consult CODEBASE_MAP.md for navigation help
- Open an issue for questions or clarifications

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Remember: **Quality over quantity**. A well-tested, clearly written small change is better than a large, untested addition.