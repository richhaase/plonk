# IDEA: Consolidate Redundant Commands

## Current State

We have several commands that overlap in functionality:

1. **Status Display**:
   - `plonk` (no args) - Shows status
   - `plonk status` - Shows status
   - `plonk ls` - Lists installed packages

2. **Adding Items**:
   - `plonk add <files>` - Adds dotfiles
   - `plonk install <packages>` - Installs packages

3. **Removing Items**:
   - `plonk rm <files>` - Removes dotfiles
   - `plonk uninstall <packages>` - Uninstalls packages

## Issues to Consider

1. **User Confusion**: Multiple ways to do the same thing can confuse users
2. **Documentation Overhead**: Each command needs docs, examples, and tests
3. **Maintenance**: More commands = more code to maintain
4. **Discoverability**: Users might not find the "right" command

## Potential Solutions

### Option A: Keep Current Structure
- **Pros**:
  - Backwards compatibility
  - Multiple entry points for different user preferences
  - Clear separation between packages and dotfiles
- **Cons**:
  - Redundancy
  - More commands to maintain

### Option B: Consolidate to Single Commands
- Remove `plonk ls` in favor of `plonk status`
- Make `plonk` (no args) just an alias for `plonk status`
- Keep add/rm for dotfiles, install/uninstall for packages
- **Pros**:
  - Clearer mental model
  - Less redundancy
- **Cons**:
  - Breaking change for users relying on `ls`

### Option C: Resource-Based Commands
With the new Resource abstraction, we could move to:
- `plonk add <resource-type> <items>` (e.g., `plonk add package ripgrep`)
- `plonk remove <resource-type> <items>`
- `plonk list [resource-type]`
- **Pros**:
  - Extensible for AI Lab features
  - Consistent pattern
- **Cons**:
  - More verbose
  - Breaking change

### Option D: Context-Aware Commands
- `plonk add <item>` - Auto-detects if it's a file path (dotfile) or package name
- `plonk remove <item>` - Same auto-detection
- **Pros**:
  - Simpler UX
  - Fewer commands
- **Cons**:
  - Ambiguity (what if a package has the same name as a file?)
  - Magic behavior can be confusing

## Questions for Discussion

1. How important is backwards compatibility for existing users?
2. Do users appreciate having multiple ways to do things, or find it confusing?
3. Should we optimize for the current use case or future extensibility?
4. How do other successful CLIs handle this (kubectl, docker, git)?

## Recommendation Placeholder

_To be filled after discussion_
