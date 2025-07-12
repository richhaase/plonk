# Multiple Add Interface Enhancement

## Overview

This document explores ways to enhance the plonk interface to make it easier to add configuration files and packages, particularly for new users onboarding to plonk. All proposals maintain backward compatibility with existing CLI commands and internal code.

## Current State

### Current Package Add Interface
```bash
# Single package
plonk pkg add git
plonk pkg add typescript --manager npm

# Add all untracked (exists but not documented well)
plonk pkg add
```

### Current Dotfile Add Interface
```bash
# Single file
plonk dot add ~/.vimrc
plonk dot add ~/.config/nvim/

# No bulk add capability
```

## Proposed Enhancements

### 1. Multiple Package Add Support ✅ HIGH PRIORITY

**Enhancement:** Allow multiple packages in a single command
```bash
# Add multiple packages at once
plonk pkg add git neovim ripgrep htop

# With specific managers
plonk pkg add --manager npm typescript prettier eslint

# Mixed managers (uses default manager for unspecified)
plonk pkg add git neovim --npm typescript prettier --cargo ripgrep
```

**Implementation approach:**
- Modify `pkg add` command to accept variadic args
- Process packages in batch for better performance
- Show progress for each package
- Continue on failure with summary at end

**Example output:**
```
Adding packages...
✓ git (homebrew)
✓ neovim (homebrew)
✗ ripgrep (homebrew) - already managed
✓ htop (homebrew)

Summary: 3 added, 1 skipped
```

### 2. Multiple Dotfile Add Support ✅ HIGH PRIORITY

**Enhancement:** Allow multiple dotfiles in a single command
```bash
# Add multiple dotfiles
plonk dot add ~/.vimrc ~/.zshrc ~/.gitconfig

# Glob pattern support
plonk dot add ~/.config/nvim/* ~/.ssh/config

# Directory and file mix
plonk dot add ~/.vimrc ~/.config/nvim/ ~/.ssh/
```

**Implementation approach:**
- Modify `dot add` command to accept variadic args
- Expand glob patterns before processing
- Handle mix of files and directories
- Show progress and summary

**Example output:**
```
Adding dotfiles...
✓ ~/.vimrc → vimrc
✓ ~/.zshrc → zshrc
✓ ~/.gitconfig → gitconfig
✓ ~/.config/nvim/ → config/nvim/

Summary: 4 files added
```

### 3. Interactive Add Mode 🤔 FUTURE CONSIDERATION

**Enhancement:** Interactive selection for untracked items
```bash
# Interactive package selection
plonk pkg add --interactive
# Shows checklist of untracked packages, user selects with space, confirms with enter

# Interactive dotfile selection
plonk dot add --interactive
# Shows tree view of untracked dotfiles, allows multi-select
```

**Implementation approach:**
- Add `--interactive` flag to add commands
- Use existing list functionality to get untracked items
- Leverage a TUI library (like bubbletea) for selection
- Batch process selected items

**Considerations:**
- Adds dependency on TUI library
- More complex testing requirements
- Could significantly improve UX for users with many untracked items

### 4. Smart Suggestions ⚠️ CHALLENGING

**Enhancement:** Suggest commonly managed items

**Modified approach (to address maintenance concerns):**
```bash
# Instead of curated lists, use heuristics
plonk suggest
# Analyzes installed packages and suggests based on:
# - Package popularity (if available from package manager)
# - Common patterns (dev tools, shell configs)
# - User's existing managed items

# Output:
# Based on your system, you might want to add:
# Packages: git (version control), neovim (editor), ripgrep (search)
# Dotfiles: ~/.zshrc, ~/.gitconfig
# Run: plonk pkg add git neovim ripgrep
```

**Alternative: Frequency-based suggestions**
- Track anonymous usage statistics locally
- Suggest packages that are frequently managed together
- No external maintenance required

### 5. Discovery Commands 🎯 SIMPLIFIED VERSION

**Enhancement:** Focus on truly universal items only

```bash
# Show only the most common untracked items
plonk discover
# Output:
# Essential packages not yet managed:
#   git - Version control (found in homebrew)
#   curl - HTTP client (found in homebrew)
#
# Essential dotfiles not yet managed:
#   ~/.zshrc - Shell configuration
#   ~/.bashrc - Shell configuration
#   ~/.gitconfig - Git configuration
```

**Implementation approach:**
- Very minimal curated list (< 10 items)
- Focus on truly universal tools
- Only suggest if actually present on system
- Self-contained, no external maintenance

## Usage Scenarios

### Scenario 1: New Developer Onboarding
```bash
# Quick setup with multiple adds
plonk pkg add git neovim curl wget
plonk dot add ~/.zshrc ~/.gitconfig ~/.vimrc

# Or use discovery for suggestions
plonk discover
# Then add suggested items
plonk pkg add git curl
plonk dot add ~/.zshrc ~/.gitconfig
```

### Scenario 2: Migrating Existing Setup
```bash
# Add all untracked packages at once
plonk pkg add

# Add multiple dotfiles
plonk dot add ~/.zshrc ~/.bashrc ~/.vimrc ~/.gitconfig ~/.config/nvim/
```

### Scenario 3: Quick Multi-Package Setup
```bash
# Add multiple development tools at once
plonk pkg add git neovim tmux fzf ripgrep bat

# Add Node.js tools
plonk pkg add --manager npm typescript prettier eslint jest
```

## Implementation Priority

### Phase 1: Core Enhancements (Immediate)
1. **Multiple package add support** - Essential for better UX
2. **Multiple dotfile add support** - Essential for better UX

### Phase 2: Discovery (Next)
3. **Simplified discovery command** - Minimal curated list of universal tools

### Phase 3: Advanced Features (Future)
4. **Interactive mode** - Consider based on user feedback
5. **Smart suggestions** - Only if sustainable approach found

## Benefits

1. **Faster Onboarding**: Add multiple items at once instead of one-by-one
2. **Better Discoverability**: Users learn what they can manage
3. **Reduced Friction**: Common operations become single commands
4. **Maintains Simplicity**: Enhancements are optional, basic usage unchanged
5. **Progressive Disclosure**: Simple commands still work, power features available when needed

## Implementation Notes

### Multiple Add Implementation Details

**Package Add Changes:**
- Update `pkg_add.go` to accept `[]string` args instead of single arg
- Process each package sequentially (parallel could cause conflicts)
- Collect results and show summary at end
- Exit code 0 if any succeed, 1 only if all fail

**Dotfile Add Changes:**
- Update `dot_add.go` to accept `[]string` args
- Expand globs using Go's `filepath.Glob()`
- Process each file/directory sequentially
- Show clear mapping of source → destination

**Error Handling:**
- Continue processing on individual failures
- Show clear error for each failed item
- Summary shows counts of success/failure
- Return appropriate exit code

### Discovery Command Implementation

**Minimal Curated List:**
```go
var essentialPackages = []string{
    "git",      // Version control
    "curl",     // HTTP client
    "wget",     // HTTP client
    "vim",      // Text editor
    "neovim",   // Text editor
}

var essentialDotfiles = []string{
    ".zshrc",     // Zsh config
    ".bashrc",    // Bash config
    ".gitconfig", // Git config
    ".vimrc",     // Vim config
}
```

## Next Steps

1. Implement multiple package add support
2. Implement multiple dotfile add support
3. Update documentation and examples
4. Consider minimal discovery command
5. Gather user feedback before pursuing interactive mode
