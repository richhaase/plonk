# TODO

Active work items for current development session. Maintained by AI agents.

## In Progress

## Pending

### CLI Interface Improvements (High Priority)
- **Add global --dry-run flag** for all applicable commands

### CLI Interface Improvements (Medium Priority)  
- **Implement command aliases** (ls→pkg list, sync→repo, check→status)
- **Enhance pkg subcommands** with search/info/update functionality
- **Add --verbose/--quiet global flags** for output control
- **Test dogfooding workflow**: plonk status, apply --dry-run, apply

### CLI Interface Improvements (Low Priority)
- **Enhance status command** with --detailed, --drift-only, --json flags

## Completed (This Session)
- **Validate plonk installation** ✅ - check which plonk and plonk --help
- **Research plonk CLI interface design** ✅ - provided comprehensive suggestions
- **Standardize argument patterns** ✅ - install and apply commands now accept optional [package] argument

## Notes

### Session Context - Dogfooding Implementation
- **Enhanced import completed** ✅ - Now parses dotfiles into rich ZSH/Git configs
- **Import workflow** ✅ - `plonk import` creates `/Users/rdh/.config/plonk/repo/plonk.yaml` 
- **GOBIN in PATH** ✅ - Added `$(go env GOBIN)` to ~/.zshrc for dogfooding
- **Justfile fixed** ✅ - Updated to use `$(go env GOBIN)` instead of `$GOPATH/bin`
- **Installation completed** ✅ - `just install` ran successfully, installed to GOBIN

### After Shell Restart - Resume Here:
1. **Validate plonk installation**: Run `which plonk` and `plonk --help`
2. **Test dogfooding workflow**: 
   - `plonk status` - check current state
   - `plonk apply --dry-run` - see what would be applied  
   - `plonk apply` - apply rich configs to generate dotfiles
3. **Complete dogfooding loop**: Edit code → `just install` → test with plonk

### Technical Details:
- **GOBIN**: `/Users/rdh/.asdf/installs/golang/1.24.4/bin` 
- **Rich configs**: 10 env vars, 44 aliases (ZSH), user settings + 8 aliases (Git)
- **Import is pure**: No automatic modifications, faithful read operation