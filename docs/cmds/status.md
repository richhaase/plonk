# Status Command

The `plonk status` command shows managed packages and dotfiles.

## Description

The status command provides a comprehensive view of all plonk-managed resources, including packages and dotfiles. It displays the current state of each resource (managed, missing, or unmanaged), helping users understand what's tracked, what needs attention, and what exists outside of plonk's management. The command supports filtering by resource type and state, with multiple output formats for different use cases.

## Behavior

### Core Function

Status displays resources in three possible states:
- **✅ managed/deployed** - Resource is tracked by plonk and exists in the environment
- **❌ missing** - Resource is tracked by plonk but doesn't exist in the environment
- **⚠ unmanaged** - Resource exists in the environment but isn't tracked by plonk

### Default Display

Without flags, status shows all managed and missing resources in two sections:
- **PACKAGES**: Shows name, manager, and status
- **DOTFILES**: Shows source (in $PLONK_DIR), target (in $HOME), and status

Always displays summary counts at the end.

### Command Options

- `--packages` - Show only package status
- `--dotfiles` - Show only dotfile status
- `--unmanaged` - Show only unmanaged items
- `-o, --output` - Output format (table/json/yaml)

Alias: `st`

### Filter Behavior

Filters can be combined:
- `--packages --unmanaged` - Show only unmanaged packages
- `--dotfiles --unmanaged` - Show only unmanaged dotfiles

Note: Using `--packages --dotfiles` together is redundant (shows both, same as no flags).

### Table Format Display

**Packages Table**:
- NAME: Package name
- MANAGER: Package manager (brew, npm, cargo, pip, gem, go)
- STATUS: Current state with icon

**Dotfiles Table** (managed/missing):
- SOURCE: Path relative to `$PLONK_DIR`
- TARGET: Full path in `$HOME`
- STATUS: Current state with icon

**Dotfiles List** (unmanaged):
- Simple list of unmanaged dotfile paths (no source/target/status columns)

### JSON/YAML Output Structure

Structured output includes:
- Configuration file paths and validity
- Detailed summary with counts by domain (dotfile/package)
- Managed items in flat array with domain field
- Uses "domain" terminology instead of separate sections

Key difference: Unmanaged items not included even with --unmanaged flag.

**Bug**: Currently `plonk status --unmanaged -o json|yaml` outputs the same information as without the --unmanaged flag, showing only managed items instead of unmanaged ones.

### Summary Information

Always displays total counts:
- Managed: Resources tracked and present
- Missing: Resources tracked but not present

With `--unmanaged`, shows untracked counts in structured output.

### Special Behaviors

- No pagination for long lists (users can pipe to pager)
- Current sorting by package manager (not alphabetical)
- Missing dotfiles shown inline with managed ones
- Unmanaged view uses different column layout for clarity

### Relationship to Other Commands

- NOT the default command (plain `plonk` shows help)
- Works with resources tracked in `plonk.lock` and `$PLONK_DIR`
- Missing items can be resolved with `plonk apply`
- Unmanaged items can be added with `plonk install` or `plonk add`

## Implementation Notes

## Improvements

- Sort items alphabetically instead of by package manager
- Review flag combination behavior (e.g., --packages --dotfiles redundancy)
- Consider built-in pagination for very long lists
- Add --missing flag to show only missing resources
- Fix bug: Include unmanaged items in JSON/YAML output when --unmanaged is used
- Consider adding color coding to package manager column for visual grouping
