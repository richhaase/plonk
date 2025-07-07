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
- Go 1.24.4 or later
- ASDF for tool version management (optional but recommended)
- Git for version control

### Development Setup
```bash
# Clone the repository
git clone <repository-url>
cd plonk

# Install development tools (if using ASDF)
asdf install

# Build the project
mage build

# Run tests
mage test

# Run the full development cycle
mage format && mage lint && mage test
```

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
- Include examples in function documentation
- Keep CHANGELOG.md updated for significant changes

## Submitting Changes

### Pull Request Process
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the TDD workflow for all changes
4. Ensure all tests pass (`mage test`)
5. Run the full development pipeline locally (`mage format && mage lint && mage test`)
6. Commit with clear messages
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