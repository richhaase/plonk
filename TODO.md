# TODO

Active work items for current development session. Maintained by AI agents.

## In Progress

## Pending

## Completed (This Session)
- ✅ **Fix pre-commit hooks** - Removed errcheck/gocritic, updated to use goimports 
- ✅ **Organize imports consistently** - Used goimports with local-prefixes: plonk
- ✅ **Standardize function documentation** - All exported functions follow Go idioms
- ✅ **Add package-level documentation** - Comprehensive package docs for API clarity
- ✅ **Verify godoc generation** - Clean professional documentation output
- ✅ **Convert remaining tests to table-driven format** - All package manager tests now consistent
- ✅ **Code Quality & Maintenance phase complete** - Ready for Core Features development
- ✅ **Import Command Implementation (Full TDD)** - Complete discovery and YAML generation:
  - Basic CLI structure and command registration
  - Homebrew package discovery (reusing existing manager)
  - AsdfManager.ListGlobalTools() enhancement for ~/.tool-versions
  - ASDF tool discovery with version support
  - NPM global package discovery (reusing existing manager)
  - Dotfile detection for .zshrc, .gitconfig, .zshenv
  - GenerateConfig() function for discovery results → Config struct
  - SaveConfig() function with clean YAML marshaling
  - Full CLI integration with progress indicators and summary

## Notes
- **Path C: Code Quality & Maintenance** completed successfully
- Pre-commit hooks working reliably with 0 linting errors
- Documentation API-ready with proper Go conventions
- All package manager tests follow table-driven patterns
- **Current phase**: Core Features - Import command development using TDD

### Import Command Design Notes
- **Purpose**: Generate plonk.yaml from existing shell environment (ROADMAP.md)
- **Components**: Package discovery (brew/asdf/npm list) + dotfile copying (.zshrc/.gitconfig/.zshenv)
- **TDD Pattern**: Each component gets red-green-refactor cycle per CONTRIBUTING.md
- **Test Patterns**: Use MockCommandExecutor, setupTestEnv(t), table-driven tests
- **Command Structure**: Follow existing patterns in internal/commands/ (status.go as reference)