# Plonk Roadmap

This document is an ideation and planning tool for developers and AI agents working on Plonk.

## Ideas & Discussion

Features and concepts being considered:

### Core Features
- **Additional shell support** - Bash and Fish configuration generation
- **Advanced backup features** - Encryption, compression, remote sync capabilities
- **Configuration templates** - Predefined setups for common development environments
- **Container testing support** - Docker/VM integration for bulletproof isolation during development
- **Cross-platform Windows support** - PowerShell profiles, Windows package managers
- **Developer deployment** - Enable `go install` for plonk installation and dogfooding during development
- **Diff command** - Show differences between config and reality
- **Enhanced dry-run capabilities** - Show exactly what files would be modified and what package commands would run across all operations
- **Environment snapshots** - Create/restore complete environment snapshots
- **Homebrew formula** - Package plonk for homebrew installation (separate tap repository)
- **Multi-machine sync** - Synchronize configurations across multiple machines
- **Package manager extensions** - Mac App Store (mas), additional Linux package managers
- **Plugin system** - Custom package manager support
- **Test isolation mode** - Sandboxed testing environment using separate directories and mocked package managers
- **Watch mode** - Auto-apply changes when config files change

### Documentation
- **API.md** - If plonk becomes a library
- **Auto-generated docs** - CLI help → markdown generation
- **CLAUDE.md enhancements** - User interaction patterns, AI troubleshooting
- **CONFIG.md** - Comprehensive configuration reference with validation rules
- **CONTRIBUTING.md updates** - Real repository URL, license details
- **Documentation testing** - Ensure examples actually work
- **Documentation versioning** - Keep docs in sync with releases
- **EXAMPLES.md** - Real-world configuration examples
- **QUICKSTART.md** - 5-minute getting started guide
- **README enhancements** - Prerequisites, "Why Plonk?" section
- **SECURITY.md** - Security policies and vulnerability reporting
- **TROUBLESHOOTING.md** - Common issues and solutions

### Infrastructure & Process
- **CI/CD setup** - GitHub Actions for automated testing and releases
- **Dependency validation** - Clean go.sum, license compliance checking
- **Integration testing** - End-to-end workflow tests
- **Security remediation** - Address gosec findings, configure suppressions for false positives
- **Public GitHub repository** - Create public repository and initial release

## In Progress

Active work items are tracked in [TODO.md](TODO.md).

## Completed

Major features completed (see [CHANGELOG.md](CHANGELOG.md) for details):

- ✅ Core CLI with 9 commands (status, install, apply, import, etc.)
- ✅ Import command - Generate plonk.yaml from existing environment
- ✅ Package manager support (Homebrew, ASDF, NPM)
- ✅ YAML configuration system with validation
- ✅ Configuration drift detection
- ✅ ZSH and Git configuration generation
- ✅ Backup system with automatic cleanup
- ✅ Dry-run and preview capabilities
- ✅ TDD development infrastructure
- ✅ Pure Go task runner - Migrated from Mage to dev.go + internal/tasks/ for zero external dependencies
- ✅ Versioning support/integration - Professional version management with semantic versioning, automated changelog updates, and release workflow
- ✅ License headers - MIT License with consistent headers across all Go files for legal compliance
- ✅ All-Go development workflow - Unified tooling via go.mod, single quality gate, simplified git hooks, cross-platform Go-native development
- ✅ Pre-commit safety - Removed dangerous `git add .`, improved hook reliability with dev.go integration
- ✅ Security scanning - Added govulncheck, gosec for vulnerability detection in development pipeline

## Parked

Ideas deferred or decided against:

- **Complex backup features** - Focus on core functionality first
- **Windows support** - Not priority during development phase
- **Package manager proliferation** - Keep focused on essential managers