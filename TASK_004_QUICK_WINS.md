# Task Context: Phase 3 Quick Wins

## Task ID: TASK_004_QUICK_WINS
## Phase: 3 - Quick Wins (Independent Items)
## Priority: Mixed (See individual tasks)
## Estimated Effort: Small to Medium

## Overview
This phase contains 4 independent improvements that can be implemented in any order. Each provides immediate value without dependencies on other work. These are "quick wins" that improve user experience and code cleanliness.

## Task 1: Doctor Copy-Paste PATH Commands

### Priority: Medium
### Location: `internal/diagnostics/health.go`
### Function: `checkPathConfiguration()`

#### Current State
The doctor command detects missing PATH directories and provides generic suggestions:
```
Suggestions:
- Add missing directories to your PATH in your shell configuration file (~/.zshrc, ~/.bashrc, etc.)

For example, add these lines to your shell config:
  export PATH="/path/to/dir:$PATH"
```

#### Required Enhancement
Detect the user's shell and provide exact, copy-pasteable commands:

1. **Detect Shell**: Use `$SHELL` environment variable or fall back to common shells
2. **Provide Shell-Specific Commands**:
   ```bash
   # For zsh (detected):
   echo 'export PATH="/Users/john/.local/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc

   # For bash:
   echo 'export PATH="/Users/john/.local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc

   # For fish:
   fish_add_path /Users/john/.local/bin
   ```

3. **Implementation Hints**:
   - Check `$SHELL` environment variable
   - Common shells: `/bin/zsh`, `/bin/bash`, `/usr/local/bin/fish`
   - Format commands so users can copy and run directly
   - Include both the config file update AND the reload command

#### Success Criteria
- Shell detection works for zsh, bash, fish at minimum
- Commands are formatted for direct copy-paste
- Instructions include both persisting and immediate application

## Task 2: Status --missing Flag

### Priority: Low
### Location: `internal/commands/status.go`

#### Current State
Status command has flags:
- `--packages` - Show only packages
- `--dotfiles` - Show only dotfiles
- `--unmanaged` - Show only unmanaged items

Missing resources (tracked but not installed) are shown mixed with managed items.

#### Required Enhancement
Add `--missing` flag to show only missing resources:

1. **Add Flag**:
   ```go
   statusCmd.Flags().Bool("missing", false, "Show only missing resources")
   ```

2. **Filter Logic**:
   - Missing = resources that are tracked (in lock file or plonk dir) but not installed/deployed
   - Works with existing filters: `--missing --packages` shows only missing packages
   - Updates table output to show only missing items

3. **Implementation Details**:
   - Add to existing filter logic in `runStatus()`
   - Missing packages: in lock file but not installed
   - Missing dotfiles: in $PLONK_DIR but not deployed to $HOME
   - Preserve existing output format

#### Success Criteria
- `plonk status --missing` shows only missing resources
- Combines with other filters correctly
- Summary counts remain accurate
- Table output clearly indicates these are missing

## Task 3: Improve Dotfile Path Documentation

### Priority: Low
### Location: Help text in `internal/commands/add.go` and `rm.go`

#### Current State
Path resolution is documented but could be clearer:
```
Path Resolution:
- Absolute paths: /home/user/.vimrc
- Tilde paths: ~/.vimrc
- Relative paths: First tries current directory, then home directory
```

#### Required Enhancement
Expand documentation with examples and edge cases:

1. **Enhanced Help Text**:
   ```
   Path Resolution:
   Plonk accepts paths in multiple formats and intelligently resolves them:

   - Absolute paths: /home/user/.vimrc → Used as-is
   - Tilde paths: ~/.vimrc → Expands to /home/user/.vimrc
   - Relative paths: .vimrc → Tries:
     1. Current directory: /current/dir/.vimrc
     2. Home directory: /home/user/.vimrc
   - Plain names: vimrc → Tries:
     1. Current directory: /current/dir/vimrc
     2. Home with dot: /home/user/.vimrc

   Special Cases:
   - Directories: Recursively processes all files (add only)
   - Symlinks: Follows links and copies target file
   - Hidden files: Automatically handled (dot removed in plonk dir)

   Examples:
     plonk add vimrc          # Finds ~/.vimrc automatically
     plonk add .config/nvim   # Adds entire nvim config directory
     plonk add ../myfile      # Relative to current directory
   ```

2. **Also Update**:
   - Similar improvements to `rm` command help
   - Ensure consistency between commands

#### Success Criteria
- Help text clearly explains all path resolution behaviors
- Examples cover common use cases
- Edge cases are documented
- Users understand how plonk finds their files

## Task 4: Remove JSON/YAML Output Support

### Priority: Medium (Simplification)
### Location: Multiple files

#### Current State
All commands support `-o json` and `-o yaml` flags, but:
- Limited use case for a dotfile manager
- Adds maintenance burden
- Bug surface area (as seen with --unmanaged flag issue)

#### Required Changes

1. **Remove Output Formats**:
   - Keep only table output (default)
   - Remove `OutputJSON` and `OutputYAML` constants
   - Remove JSON/YAML rendering code

2. **Files to Modify**:
   - `internal/commands/output.go` - Remove JSON/YAML support
   - `internal/commands/output_types.go` - Remove structured data interfaces
   - All command files - Remove `-o/--output` flag registration
   - Remove `StructuredData()` methods from output types

3. **Simplification Benefits**:
   - Remove `TableOutput()` and `StructuredData()` interface
   - Commands directly print table output
   - No more field name case conversions (PascalCase/snake_case)
   - Simpler error messages (no format parsing)

4. **User Impact**:
   - Breaking change but acceptable (single user)
   - Simpler command interface
   - Cleaner help text

#### Success Criteria
- All JSON/YAML code removed
- Commands still produce clear table output
- Help text no longer mentions output formats
- Tests updated to remove format testing

## Implementation Notes

### General Guidelines
1. Each task is independent - implement in any order
2. Keep changes focused on the specific task
3. Update tests as needed
4. Don't over-engineer - these are "quick wins"

### Testing Approach
- Task 1: Test with different shells and PATH scenarios
- Task 2: Test filter combinations and edge cases
- Task 3: Documentation only - review for clarity
- Task 4: Ensure all commands still work after removal

### Breaking Changes
Only Task 4 (JSON/YAML removal) is a breaking change. This is acceptable as there's only one user.

## Deliverables
For each task completed:
1. Implementation matching requirements
2. Updated tests if applicable
3. Brief summary of changes made
4. Any design decisions or trade-offs

The implementer can choose to do all 4 tasks or any subset based on time/interest.
