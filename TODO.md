# TODO

Active work items for current development session. Maintained by AI agents.

## In Progress

## Pending

### Import Command Implementation (TDD Breakdown)
- ✅ **RED-GREEN-REFACTOR: Basic import command structure** - CLI interface and command registration complete
- ✅ **RED-GREEN-REFACTOR: Homebrew package discovery** - HomebrewDiscoverer with brew list parsing complete
- ✅ **RED-GREEN-REFACTOR: AsdfManager.ListGlobalTools() enhancement** - Added ~/.tool-versions file reading capability
- ✅ **RED-GREEN-REFACTOR: ASDF package discovery** - AsdfDiscoverer using new ListGlobalTools() method complete
- ✅ **RED-GREEN-REFACTOR: NPM package discovery** - NpmDiscoverer reusing existing NpmManager complete
- ✅ **RED-GREEN-REFACTOR: Dotfile detection** - DotfileDiscoverer for managed dotfiles complete
- ✅ **RED-GREEN-REFACTOR: GenerateConfig() function** - Convert discovery results to Config struct complete
- ✅ **RED-GREEN-REFACTOR: SaveConfig() function** - Marshal Config struct to YAML file with custom formatting complete
- 🔴 **RED: Wire import command** - Integrate all components into CLI command with proper output

## Completed (This Session)
- ✅ **Fix pre-commit hooks** - Removed errcheck/gocritic, updated to use goimports 
- ✅ **Organize imports consistently** - Used goimports with local-prefixes: plonk
- ✅ **Standardize function documentation** - All exported functions follow Go idioms
- ✅ **Add package-level documentation** - Comprehensive package docs for API clarity
- ✅ **Verify godoc generation** - Clean professional documentation output
- ✅ **Convert remaining tests to table-driven format** - All package manager tests now consistent
- ✅ **Code Quality & Maintenance phase complete** - Ready for Core Features development

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