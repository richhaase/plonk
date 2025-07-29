# Setup Command Refactoring Summary

**UPDATE**: The `plonk setup` command has been completely removed as requested. The refactoring is now complete with only `plonk init` and `plonk clone` commands.

## Overview

Successfully refactored the `plonk setup` command into two distinct commands: `plonk init` and `plonk clone`. This provides better user control and intelligent package manager detection.

## Changes Made

### 1. New Commands Created

#### `plonk init`
- **Purpose**: Initialize fresh plonk configuration with manual control
- **Location**: `internal/commands/init.go`
- **Key Features**:
  - Creates default plonk.yaml and empty lock file
  - Supports skip flags for each package manager (--no-homebrew, --no-cargo, etc.)
  - Installs all managers by default unless explicitly skipped
  - Non-interactive mode with --yes flag

#### `plonk clone`
- **Purpose**: Clone repository with intelligent auto-detection
- **Location**: `internal/commands/clone.go`
- **Key Features**:
  - Clones repository into PLONK_DIR
  - Automatically detects required managers from lock file
  - Only installs managers that are actually needed
  - No skip flags - intelligence replaces manual control
  - Optional --no-apply flag to skip running apply

### 2. Core Implementation Updates

#### Setup Package Enhancements
- Added `SkipManagers` struct for init command control
- Added `DetectRequiredManagers()` function for lock file analysis
- Added `installDetectedManagers()` for targeted installation
- Updated `Config` struct with new fields: SkipManagers, NoApply
- Modified `CheckAndInstallToolsFromReport()` to respect skip flags

#### Detection Logic
- Reads v2 lock file format
- Extracts managers from resource metadata
- Falls back to ID prefix parsing if needed
- Returns empty list if no lock file exists

### 3. Complete Removal of Setup Command

The original `plonk setup` command has been:
- Completely removed from the codebase
- Documentation deleted (setup.md)
- All references updated in other documentation
- No backward compatibility maintained

### 4. Documentation Updates

- Created `docs/cmds/init.md` - comprehensive init command docs
- Created `docs/cmds/clone.md` - comprehensive clone command docs
- Deleted `docs/cmds/setup.md` - removed entirely
- Updated `docs/cli.md` - added new commands, removed setup
- Updated `docs/installation.md` - replaced setup with init/clone
- Updated `docs/why-plonk.md` - replaced setup example with clone
- Updated `docs/cmds/doctor.md` - updated integration references
- All documentation follows existing format conventions

## Design Decisions

### Why No Skip Flags for Clone?

The whole point of clone is intelligent detection. Adding skip flags would:
- Defeat the purpose of auto-detection
- Add unnecessary complexity
- Confuse users about when to use which flags

### Lock File Detection Strategy

1. Try metadata field first (v2 format)
2. Fall back to ID prefix parsing
3. Build unique list of managers
4. Handle missing lock files gracefully

### Error Handling Philosophy

- Partial success is acceptable (some managers install, others fail)
- Provide manual installation instructions for failures
- Don't block entire operation for optional tools
- Clone continues even without lock file

## Migration Guide for Users

### Previous Commands (No Longer Available)
```bash
# Old way - initialize
plonk setup

# Old way - clone
plonk setup user/dotfiles
```

### New Commands (Required)
```bash
# New way - initialize with control
plonk init
plonk init --no-cargo --no-gem

# New way - clone with intelligence
plonk clone user/dotfiles
plonk clone user/repo --no-apply
```

**Note**: The `plonk setup` command has been completely removed. Users must use the new commands.

## Technical Benefits

1. **Clearer Intent**: Command names match their purpose
2. **Better Control**: Init allows fine-grained manager selection
3. **Smarter Defaults**: Clone only installs what's needed
4. **Cleaner Code**: Separate commands with focused responsibilities
5. **Future Proof**: Easy to extend each command independently

## Testing Recommendations

1. Test init with various skip flag combinations
2. Test clone with repositories containing different managers
3. Test clone with missing or invalid lock files
4. Test deprecation warnings on old setup command
5. Test --no-apply flag on clone command
6. Test non-interactive mode for both commands

## Future Enhancements

1. Consider adding --only-<manager> flags as alternative to skip flags
2. Add progress indicators during manager installation
3. Consider parallel manager installation for speed
4. Add dry-run support to preview what would be installed
