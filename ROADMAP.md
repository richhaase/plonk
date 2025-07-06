# Plonk Roadmap

This document is an ideation and planning tool for developers and AI agents working on Plonk.

## Ideas & Discussion

Features and concepts being considered:

### Core Features
- **Environment snapshots** - Create/restore complete environment snapshots
- **Cross-platform Windows support** - PowerShell profiles, Windows package managers
- **Package manager extensions** - Mac App Store (mas), additional Linux package managers
- **Advanced backup features** - Encryption, compression, remote sync capabilities
- **Configuration templates** - Predefined setups for common development environments
- **Multi-machine sync** - Synchronize configurations across multiple machines
- **Plugin system** - Custom package manager support

### Documentation Improvements
- **CONFIG.md** - Comprehensive configuration reference with validation rules
- **QUICKSTART.md** - 5-minute getting started guide
- **TROUBLESHOOTING.md** - Common issues and solutions
- **EXAMPLES.md** - Real-world configuration examples
- **API.md** - If plonk becomes a library
- **SECURITY.md** - Security policies and vulnerability reporting

### Process Improvements
- **Documentation testing** - Ensure examples actually work
- **Auto-generated docs** - CLI help → markdown generation
- **Documentation versioning** - Keep docs in sync with releases
- **README enhancements** - Prerequisites, "Why Plonk?" section
- **CONTRIBUTING.md updates** - Real repository URL, license details
- **CLAUDE.md enhancements** - User interaction patterns, AI troubleshooting

## Planned

Features ready for implementation, roughly in priority order:

### Code Quality & Maintenance
- **Organize imports** consistently across all files
- **Standardize function documentation** 
- **Convert remaining tests** to table-driven format

### Core Features
- **Diff command** - Show differences between config and reality
- **Import command** - Generate plonk.yaml from existing shell configs
- **Additional shell support** - Bash and Fish configuration generation
- **Watch mode** - Auto-apply changes when config files change

### Infrastructure
- **Integration testing** - End-to-end workflow tests
- **CI/CD setup** - Automated testing and releases

## In Progress

Active work items are tracked in [TODO.md](TODO.md).

## Completed

Major features completed (see [CHANGELOG.md](CHANGELOG.md) for details):

- ✅ Core CLI with 8 commands (status, install, apply, etc.)
- ✅ Package manager support (Homebrew, ASDF, NPM)
- ✅ YAML configuration system with validation
- ✅ Configuration drift detection
- ✅ ZSH and Git configuration generation
- ✅ Backup system with automatic cleanup
- ✅ Dry-run and preview capabilities
- ✅ TDD development infrastructure

## Parked

Ideas deferred or decided against:

- **Complex backup features** - Focus on core functionality first
- **Windows support** - Not priority during development phase
- **Package manager proliferation** - Keep focused on essential managers