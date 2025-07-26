# Phase 13.75: Replace Init with Setup Command

## Objective
Replace the limited `plonk init` command with a comprehensive `plonk setup` command that handles both initial configuration and cloning existing dotfile repositories, making plonk much more useful for new users and machine setup.

## Current State
- `plonk init` only creates default config and lock files
- No automated way to clone and set up existing plonk repositories
- No automated tool installation from doctor findings
- Users must manually create PLONK_DIR and run multiple commands

## Target State
- `plonk init` command removed entirely
- New `plonk setup` command that:
  - Without args: creates PLONK_DIR, config files, installs tools
  - With git repo: clones into PLONK_DIR, detects config, installs tools, runs apply
- `plonk doctor --fix` flag to install missing tools without full setup
- Streamlined onboarding experience for new machines

## Implementation Tasks

### 1. Create Setup Command Structure
- [ ] Create `internal/commands/setup.go`
- [ ] Define command with optional git repo argument
- [ ] Add `--yes` flag for non-interactive mode
- [ ] Register command in root

### 2. Implement Core Setup Logic
- [ ] Create setup package in `internal/setup/` for business logic
- [ ] Implement `setupWithoutRepo()` function:
  - [ ] Check if PLONK_DIR exists, prompt for overwrite if needed
  - [ ] Create PLONK_DIR
  - [ ] Create default plonk.yaml config
  - [ ] Create empty plonk.lock file
  - [ ] Run doctor checks and install missing tools (with prompts)
- [ ] Implement `setupWithRepo()` function:
  - [ ] Validate git URL format (support HTTPS, SSH, GitHub shorthand)
  - [ ] Convert GitHub shorthand to HTTPS URL
  - [ ] Check if PLONK_DIR exists, prompt to delete if needed
  - [ ] Clone repository into PLONK_DIR
  - [ ] Check for existing plonk.yaml
  - [ ] If no plonk.yaml, create default files
  - [ ] Run doctor checks and install missing tools
  - [ ] If plonk.yaml existed, run `plonk apply`

### 3. Add Doctor Fix Flag
- [ ] Add `--fix` flag to doctor command
- [ ] Create shared tool installation code in `internal/setup/tools.go`
- [ ] Implement tool installation logic:
  - [ ] Detect which package managers are missing
  - [ ] Prompt user before installing each (skip with --yes)
  - [ ] Install using appropriate method for each OS
  - [ ] Re-run doctor to verify installation
- [ ] Both `plonk doctor --fix` and `plonk setup` use same installation code

### 4. Git Operations
- [ ] Implement git URL parsing and validation
- [ ] Support formats:
  - [ ] HTTPS: `https://github.com/user/repo.git`
  - [ ] SSH: `git@github.com:user/repo.git`
  - [ ] GitHub shorthand: `user/repo` → `https://github.com/user/repo.git`
- [ ] Handle git clone with proper error handling
- [ ] Clean up on failure (remove partial clone)

### 5. Interactive Prompts
- [ ] Implement confirmation prompts for:
  - [ ] Overwriting existing PLONK_DIR (with trash/delete)
  - [ ] Installing each missing tool
  - [ ] Running plonk apply after clone
- [ ] Respect --yes flag to skip all prompts
- [ ] Clear error messages for each failure scenario
- [ ] Implement OS-specific trash behavior:
  - [ ] macOS: Move to ~/.Trash
  - [ ] Linux: Try XDG trash spec, fallback to rm
  - [ ] Show user where files were moved

### 6. Remove Init Command
- [ ] Delete `internal/commands/init.go`
- [ ] Remove init command registration
- [ ] Update any references in documentation

### 7. Update Help Text
- [ ] Write comprehensive help for setup command
- [ ] Include examples for both use cases
- [ ] Update root command help to mention setup
- [ ] Add setup to getting started docs

### 8. Error Handling
- [ ] PLONK_DIR already exists scenarios
- [ ] Git clone failures (network, auth, invalid URL)
- [ ] Tool installation failures
- [ ] Apply command failures after clone
- [ ] Partial state recovery strategies

## Testing Requirements

### Unit Tests
- [ ] URL parsing and validation
- [ ] GitHub shorthand conversion
- [ ] Setup logic with mocked file operations
- [ ] Doctor fix logic with mocked installers

### Integration Tests
- [ ] Setup without repo creates correct structure
- [ ] Setup with invalid URLs fails appropriately
- [ ] Doctor --fix installs tools correctly
- [ ] Prompts work correctly in interactive mode
- [ ] --yes flag bypasses all prompts

### Manual Testing Checklist
- [ ] `plonk setup` creates new config
- [ ] `plonk setup user/dotfiles` clones and configures
- [ ] `plonk setup https://...` works with full URLs
- [ ] Existing PLONK_DIR handling works correctly
- [ ] Tool installation prompts work
- [ ] Error messages are helpful
- [ ] Cleanup happens on failures

## Success Criteria
- Init command completely removed
- Setup provides smooth onboarding experience
- Git repository cloning works reliably
- Tool installation is automated but safe
- Error handling guides users to resolution
- --yes flag enables full automation

## Risk Mitigation
- Move to trash instead of permanent deletion (OS-specific)
- Test git operations thoroughly (network issues, auth)
- Ensure cleanup on partial failures
- Clear documentation on expected behavior
- Consider dry-run mode for testing?

## Example Usage

```bash
# First time setup on new machine
$ plonk setup
Creating plonk directory at ~/.plonk...
Creating default configuration...
Checking system requirements...

Missing package managers:
- homebrew (required for brew packages)
- cargo (required for rust packages)

Install homebrew? [y/N]: y
Installing homebrew...
✓ Homebrew installed successfully

Install cargo? [y/N]: y
Installing cargo...
✓ Cargo installed successfully

✓ Setup complete! Run 'plonk status' to see current state.

# Setup from existing dotfiles repo
$ plonk setup richhaase/dotfiles
Cloning https://github.com/richhaase/dotfiles.git into ~/.plonk...
✓ Repository cloned successfully
✓ Found existing plonk.yaml configuration

Checking system requirements...
Missing package managers:
- homebrew (required for brew packages)

Install homebrew? [y/N]: y
Installing homebrew...
✓ Homebrew installed successfully

Running 'plonk apply' to configure your system...
✓ Applied 15 packages and 8 dotfiles

✓ Setup complete! Your dotfiles are now managed by plonk.

# Non-interactive setup
$ plonk setup richhaase/dotfiles --yes
Cloning https://github.com/richhaase/dotfiles.git into ~/.plonk...
✓ Repository cloned successfully
✓ Found existing plonk.yaml configuration
Installing missing tools...
✓ Homebrew installed
✓ Applied configuration
✓ Setup complete!

# Just fix missing tools
$ plonk doctor --fix
Missing package managers:
- npm (required for node packages)

Install npm? [y/N]: y
Installing Node.js and npm...
✓ npm installed successfully
✓ All tools available!
```

## Notes
- This significantly improves the new user experience
- Makes sharing dotfiles with plonk much easier
- The doctor --fix flag is useful beyond just setup
- Consider adding progress indicators for long operations (git clone, tool install)
- May want to add --verbose flag for debugging
