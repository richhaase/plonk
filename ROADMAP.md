# Plonk Roadmap

Strategic planning document for Plonk development. See [TODO.md](TODO.md) for tactical session management.

## Ideas & Discussion

Features and concepts being considered:

### Core Features
- **Advanced backup features** - Encryption, compression, remote sync capabilities
- **Configuration templates** - Predefined setups for common development environments
- **Container testing support** - Docker/VM integration for bulletproof isolation during development
- **Cross-platform Windows support** - PowerShell profiles, Windows package managers
- **Diff command** - Show differences between config and reality
- **Environment snapshots** - Create/restore complete environment snapshots
- **Multi-machine sync** - Synchronize configurations across multiple machines
- **Package manager extensions** - Mac App Store (mas), additional Linux package managers
- **Plugin system** - Custom package manager support
- **Test isolation mode** - Sandboxed testing environment using separate directories and mocked package managers

### Documentation
- **API.md** - If plonk becomes a library
- **Auto-generated docs** - CLI help → markdown generation
- **CONFIG.md** - Comprehensive configuration reference with validation rules
- **Documentation testing** - Ensure examples actually work
- **Documentation versioning** - Keep docs in sync with releases
- **QUICKSTART.md** - 5-minute getting started guide

### Infrastructure & Process
- **Dependency validation** - Clean go.sum, license compliance checking

## Current Phase: Pre-Launch Validation

**🐕 Dogfooding Phase (6-8 hours estimated)**

Real-world validation using Rich's complete development environment before public release.

### Stage 1: End-to-End Workflow Validation (~2 hours)
Validate complete install → import → status workflow with fresh installation.
- Fresh plonk installation test
- Complete setup workflow validation
- Import functionality testing
- Status and listing command accuracy

### Stage 2: Real Environment Migration (~2-3 hours)  
Migrate Rich's complete development environment under plonk management.
- Comprehensive plonk.yaml creation
- Dotfiles management testing
- Package manager migration (Homebrew, ASDF, NPM)
- Backup and restore functionality validation

### Stage 3: Integration Testing & Documentation (~1-2 hours)
Create integration tests and document edge cases discovered during dogfooding.
- Integration test creation from real scenarios
- Edge case documentation
- Multi-machine testing
- Error condition validation

### Stage 4: UX Refinement & Polish (~1-2 hours)
Improve user experience based on real usage pain points.
- Error message improvements
- Command help text enhancement
- Workflow optimization
- User feedback improvements

**🎯 GitHub Launch Phases (8-10 hours estimated)**

Post-dogfooding activities to prepare for public release.

### Phase 1: Security & Quality (~2 hours)
Address security findings and establish CI/CD pipeline.
- Security findings resolution (34 gosec issues)
- Re-enable pre-commit hooks (currently disabled for dogfooding)
- CI/CD pipeline setup
- Quality gate establishment

### Phase 2: User Experience (~5-6 hours)
Create essential user documentation and improve error handling.
- Essential documentation creation
- Error handling improvements
- User onboarding optimization

### Phase 3: Robustness (~5-6 hours)
Enhanced validation, testing, and documentation improvements.
- Validation enhancements
- Testing coverage expansion
- Create DEVELOPER.md with detailed technical patterns and conventions
- Documentation improvements

### Phase 4: Launch (~2 hours)
GitHub-specific setup and repository creation.
- GitHub project setup
- Release automation
- Repository creation and first release

## Historical Context

For completed features and development history, see [CHANGELOG.md](CHANGELOG.md).

## Parked

Ideas deferred or decided against:
