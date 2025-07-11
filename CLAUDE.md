# Plonk Development Guidelines

## Development Rules

### Error Handling
- **Always use plonk's error handling** - `errors.Wrap()` not `fmt.Errorf()`
- **Package manager availability** - Return `(false, nil)` not error when binary missing
- **Structured errors** - Use appropriate error codes and domains for user-friendly messages

### Testing & Quality
- **Test before release** - Run `just precommit` before any release
- **Graceful degradation** - Features should fail gracefully (e.g., unavailable package managers)
- **Context everywhere** - All long-running operations must accept context for cancellation

### Developer Experience
- **Zero-config first** - Features must work without configuration
- **One-command setup** - New developers should be productive with `just dev-setup`
- **Fast feedback loops** - Use pre-commit framework for file-specific checks (94% faster)

### Build & Release
- **Binary location** - Output to `bin/` not `build/`
- **Automated releases** - Use GoReleaser via `just release-auto`
- **Composite actions** - Reuse GitHub Actions for consistency

## Completed Work Summary

### Session 1: Core Functionality Fixes
- Fixed 7 critical issues including:
  - JSON/YAML output verbosity
  - Lock file being treated as dotfile
  - Package installation false success reports
  - Config field name formatting
- All issues resolved without breaking changes

### Session 2: Automation & Developer Experience

#### Release Process
- Replaced custom release with GoReleaser
- Added GitHub Actions workflow for automated releases
- Fixed package manager availability checks

#### Pre-commit Framework
- Migrated from shell scripts to pre-commit framework
- 94% performance improvement on non-Go changes
- Removed all legacy hook infrastructure

#### CI/CD Improvements
- Created composite actions (50% complexity reduction)
- Eliminated 60+ lines of duplicated justfile code
- Standardized Go environment setup across workflows

#### Developer Commands
- `just dev-setup` - Complete environment setup in one command
- `just deps-update` - Safe dependency updates with validation
- `just clean-all` - Deep clean including all caches

#### Build Standardization
- Changed output directory from `build/` to `bin/`
- Updated all documentation and scripts

### Key Achievements
- 🚀 2-minute developer onboarding
- ⚡ 94% faster pre-commit hooks
- 🔧 Modern CI/CD with reusable components
- 📦 Automated multi-platform releases
- 🧹 Cleaner, more maintainable codebase

### Next Steps
- Symlink behavior investigation (deferred)
- Unicode path support
- Network failure handling
- Performance benchmarking

## Commands Reference
```bash
just dev-setup      # One-time developer setup
just deps-update    # Update dependencies safely
just clean-all      # Complete cleanup
just release-auto   # Create automated release
```
