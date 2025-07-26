# Phase 15: Output Standardization

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase standardizes output formatting across all commands to ensure a consistent, professional user experience. We'll establish patterns for tables, JSON, and YAML output while maintaining the existing `-o` flag functionality.

## Objectives

1. Standardize table formatting across all commands
2. Ensure consistent JSON/YAML structure
3. Standardize error message formats
4. Create output guidelines for consistency
5. Address info command output design

## Output Format Principles

### General Rules
- Keep the `-o table|json|yaml` flag on all commands
- Default to table format for human readability
- JSON/YAML should be properly structured, not just dumped structs
- Error output always goes to stderr
- Data output always goes to stdout
- Tables should be readable on 80-character terminals where possible

### Table Format Standards

Use consistent column headers and alignment:
```
PACKAGE     MANAGER    VERSION    STATUS
ripgrep     homebrew   14.0.3     installed
typescript  npm        5.2.0      missing
```

Guidelines:
- Left-align text columns
- Right-align numeric columns
- Use consistent spacing (tabwriter)
- Show most important info first
- Avoid truncation where possible

### JSON Format Standards

Ensure consistent structure:
```json
{
  "packages": [
    {
      "name": "ripgrep",
      "manager": "homebrew",
      "version": "14.0.3",
      "status": "installed"
    }
  ],
  "summary": {
    "total": 1,
    "installed": 1,
    "missing": 0
  }
}
```

### YAML Format Standards

Human-readable YAML:
```yaml
packages:
  - name: ripgrep
    manager: homebrew
    version: 14.0.3
    status: installed

summary:
  total: 1
  installed: 1
  missing: 0
```

## Command-Specific Standards

### 1. Status Command

Table format:
```
PACKAGES (2 managed, 1 missing)
NAME        MANAGER    STATUS
ripgrep     homebrew   installed
typescript  npm        missing

DOTFILES (3 managed, 3 linked)
PATH                TARGET                          STATUS
~/.gitconfig       ~/dotfiles/git/gitconfig        linked
~/.zshrc           ~/dotfiles/zsh/zshrc           linked
~/.config/nvim     ~/dotfiles/nvim                linked
```

### 2. Info Command

Based on package state, show relevant information:

**For managed packages:**
```
Package: ripgrep
Status: Managed by plonk
Manager: homebrew
Installed: 14.0.3
Latest: 14.0.3
Description: Recursively search directories for patterns
```

**For installed (not managed):**
```
Package: wget
Status: Installed (not managed)
Manager: homebrew
Version: 1.21.4
Description: Internet file retriever
Note: Run 'plonk install brew:wget' to manage this package
```

**For available:**
```
Package: jq
Status: Available
Manager: homebrew
Version: 1.7 (latest)
Description: Command-line JSON processor
Install: plonk install brew:jq
```

### 3. Search Command

Unified table showing all results:
```
Searching for "grep"...

PACKAGE     MANAGER    VERSION    DESCRIPTION
ripgrep     homebrew   14.0.3     Fast grep replacement
grep        homebrew   3.11       GNU grep
ripgrep     cargo      14.0.3     Recursively search directories
```

### 4. Apply Command

Progress and summary:
```
Applying configuration...

Packages:
✓ brew:ripgrep (already installed)
✓ npm:typescript (installed)
✗ brew:unknown (not found)

Dotfiles:
✓ ~/.gitconfig (linked)
✓ ~/.zshrc (linked)

Summary: 4 succeeded, 1 failed
```

### 5. Error Messages

Consistent format across all commands:
```
Error: <what went wrong>
<how to fix it>

Example:
Error: unknown package manager "hombrew"
Valid managers: homebrew, npm, pip, cargo, go, gem
```

## Implementation Details

### 1. Create Output Helpers

Consider creating shared formatting utilities:
- Table builder for consistent spacing
- JSON/YAML serialization helpers
- Status icon helpers (✓, ✗, ...)
- Column width calculators

### 2. Update Each Command

Review and update output for:
- status
- info
- search
- apply
- install/uninstall
- add/rm
- config show

### 3. Test Output Formats

Ensure each command properly supports:
- Default table output
- `-o json` with proper structure
- `-o yaml` with readable format
- Error output to stderr

## Testing Requirements

### Output Tests
1. Verify table alignment with various data
2. Test JSON is valid and structured
3. Test YAML is valid and readable
4. Verify error output goes to stderr
5. Test output on narrow terminals (80 chars)

### Integration Tests
- Test all commands with all output formats
- Verify consistency across commands
- Test error handling in each format

## Validation Checklist

Before marking complete:
- [ ] All commands use consistent table formatting
- [ ] JSON output is properly structured (not raw structs)
- [ ] YAML output is human-readable
- [ ] Error messages follow consistent format
- [ ] Info command shows appropriate detail by state
- [ ] Search results are clearly presented
- [ ] Apply command shows clear progress and summary
- [ ] All output formats tested
- [ ] Guidelines documented for future commands

## Notes

- This phase is about consistency and professionalism
- Don't change the -o flag behavior, just standardize the output
- Focus on making default (table) output excellent
- Consider terminal width but don't over-optimize
- Clear, consistent output improves the overall UX significantly

Remember to create `PHASE_15_COMPLETION.md` when finished!
