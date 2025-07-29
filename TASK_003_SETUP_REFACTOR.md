# Task Context: Setup Command Refactoring

## Task ID: TASK_003_SETUP_REFACTOR
## Phase: 2 - Setup Refactoring
## Priority: High (All items implement together)
## Estimated Effort: Large (4 interconnected features)

## Overview
Refactor the setup command to provide better user control and intelligence. This involves splitting the command into two distinct use cases (`init` vs `clone`), adding package manager skip flags, implementing auto-detection from lock files, and making the setup process intelligent about which tools to install.

## Current State Analysis

### Existing Setup Command
- Single command with dual behavior based on arguments
- `plonk setup` → Initialize new configuration
- `plonk setup user/repo` → Clone repository and apply
- Always attempts to install ALL package managers
- No way to skip specific managers
- Doesn't read cloned lock file to determine needs

### Key Files
- **Command Entry**: `internal/commands/setup.go`
- **Core Logic**: `internal/setup/setup.go`
- **Tool Installation**: `internal/setup/tools.go`
- **Git Operations**: `internal/setup/git.go`
- **Prompts**: `internal/setup/prompts.go`

## Requirements

### 1. Split into init/clone Commands
Transform the current dual-mode setup into two clear commands:

#### plonk init
- Initialize new plonk configuration
- Create default plonk.yaml and empty lock file
- Check system requirements
- Install package managers (with skip flags)
- Clear single purpose: "start fresh"

#### plonk clone <repo>
- Clone dotfiles repository into PLONK_DIR
- Read plonk.lock to detect required managers
- Install only necessary package managers
- Run apply automatically after setup
- Clear single purpose: "restore from repo"

### 2. Skip Package Manager Flags
Add flags to skip specific package manager installations:
- `--no-homebrew` - Skip Homebrew installation
- `--no-cargo` - Skip Cargo/Rust installation
- `--no-npm` - Skip npm/Node.js installation
- `--no-pip` - Skip pip/Python installation
- `--no-gem` - Skip gem/Ruby installation
- `--no-go` - Skip Go installation
- `--all` - Install all managers (current behavior)

These flags apply to both `init` and `clone` commands.

### 3. Auto-detect from plonk.lock
When running `plonk clone`:
1. Clone the repository first
2. Read the cloned plonk.lock file
3. Detect which package managers are needed:
   - Check resource entries with type "package"
   - Extract manager from metadata or ID prefix
   - Build list of required managers
4. Only prompt to install detected managers
5. Honor skip flags to override auto-detection

### 4. Intelligent Clone + Apply
Make the clone operation smarter:
1. Clone repository
2. Detect required managers from lock file
3. Install only required managers (minus skipped)
4. Run apply to install packages and deploy dotfiles
5. Show summary of what was set up

## Implementation Plan

### Step 1: Command Structure Refactoring
1. Create new `init.go` and `clone.go` command files
2. Move initialization logic from setup.go to init command
3. Move clone logic from setup.go to clone command
4. Update command help text and examples
5. Keep `setup` as deprecated alias (shows deprecation message)

### Step 2: Add Skip Flags
1. Define flags in both init and clone commands
2. Create `SkipManagers` struct to pass skip preferences
3. Update `installMissingManagers()` to respect skip flags
4. Modify prompts to exclude skipped managers

### Step 3: Implement Lock File Detection
1. Add `DetectRequiredManagers(lockPath string) ([]string, error)`
2. Use the new v2 lock format to read metadata
3. Extract unique manager list from resources
4. Handle missing/invalid lock files gracefully

### Step 4: Integrate Intelligence
1. In clone command, call DetectRequiredManagers after clone
2. Intersect detected managers with non-skipped managers
3. Only attempt to install the resulting set
4. Pass this filtered list to installMissingManagers

### Step 5: Update Apply Integration
1. Ensure apply runs after successful clone setup
2. Add --no-apply flag if users want to skip it
3. Show clear progress messages during the process

## Example Usage

```bash
# Initialize fresh plonk setup
plonk init
plonk init --no-cargo --no-gem  # Skip specific managers

# Clone existing dotfiles
plonk clone richhaase/dotfiles
plonk clone user/repo --no-npm  # Clone but skip npm even if needed
plonk clone user/repo --no-apply # Clone and setup but don't apply

# Deprecated (shows warning)
plonk setup  # → "Use 'plonk init' instead"
plonk setup user/repo  # → "Use 'plonk clone' instead"
```

## Technical Considerations

### Lock File Reading
With v2 format, managers are stored in metadata:
```yaml
resources:
  - type: package
    id: go:gopls
    metadata:
      manager: go
      name: gopls
      source_path: golang.org/x/tools/cmd/gopls
```

### Manager Detection Logic
```go
func DetectRequiredManagers(lockPath string) ([]string, error) {
    // Read lock file
    // Extract unique managers from resources
    // Return sorted list
}
```

### Skip Flags Structure
```go
type SkipManagers struct {
    Homebrew bool
    Cargo    bool
    NPM      bool
    Pip      bool
    Gem      bool
    Go       bool
}
```

### Backward Compatibility
- Keep `plonk setup` working with deprecation notice
- Suggest appropriate new command based on usage
- Remove in future version after transition period

## Success Criteria
1. ✅ Two distinct commands: `init` and `clone`
2. ✅ Skip flags work for all package managers
3. ✅ Clone auto-detects required managers from lock file
4. ✅ Only necessary managers are installed
5. ✅ Apply runs automatically after clone (unless skipped)
6. ✅ Clear deprecation path for old setup command
7. ✅ All existing functionality preserved

## Testing Requirements
1. Test init with various skip flag combinations
2. Test clone with repositories containing different managers
3. Test auto-detection with v2 lock files
4. Test backward compatibility warnings
5. Test error handling (bad repo, missing lock file, etc.)

## Breaking Changes
- `plonk setup` command is deprecated (but still works)
- Users need to choose between `init` and `clone`
- This is acceptable as there's only one user

## Deliverables
1. Refactored command structure (init.go, clone.go)
2. Skip flags implementation
3. Lock file detection functionality
4. Intelligent setup that reads requirements
5. Updated documentation and help text
6. Summary document explaining:
   - How commands were split
   - How detection works
   - Any design decisions made
   - Migration guide for users
