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

Note: Using `--packages --dotfiles` together explicitly selects both domains (same effect as no flags).

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

**Bug**: JSON/YAML output formats do not support the `--unmanaged` flag and will always show managed items only, regardless of the flag. Use table format to view unmanaged items.

### Summary Information

Always displays total counts:
- Managed: Resources tracked and present
- Missing: Resources tracked but not present

With `--unmanaged`, summary is hidden entirely in table format. Structured output (JSON/YAML) does not support unmanaged filtering.

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

The status command provides comprehensive resource state management through a unified reconciliation system:

**Command Structure:**
- Entry point: `internal/commands/status.go`
- Orchestration layer: `internal/orchestrator/orchestrator.go`
- Domain reconciliation: `internal/resources/dotfiles/reconcile.go`, `internal/resources/packages/reconcile.go`
- Core reconciliation engine: `internal/resources/reconcile.go`

**Key Implementation Flow:**

1. **Command Processing:**
   - Parses filter flags (`--packages`, `--dotfiles`, `--unmanaged`)
   - Implementation: If neither `--packages` nor `--dotfiles` is set, both are enabled
   - Using `--packages --dotfiles` together explicitly selects both domains (same effect as default)
   - Loads configuration gracefully (continues even if config is invalid)

2. **State Reconciliation:**
   - Calls `orchestrator.ReconcileAll()` with 30-second implicit timeout via context
   - Reconciles all domains (packages and dotfiles) in parallel
   - Uses unified reconciliation engine with three states:
     - `StateManaged`: Resource in config AND present in environment
     - `StateMissing`: Resource in config BUT not present in environment
     - `StateUntracked`: Resource present in environment BUT not in config

3. **Domain-Specific Reconciliation:**
   - **Packages**: Uses composite keys (`manager:name`) from plonk.lock file
   - **Dotfiles**: Uses filesystem-as-state from $PLONK_DIR contents
   - Each domain handles its own discovery and state comparison
   - Metadata merging preserves actual resource details

4. **Output Formatting:**
   - Table format shows hierarchical sections (PACKAGES, DOTFILES)
   - Implementation: StructuredData() method doesn't support ShowUnmanaged flag
   - JSON/YAML formats only show managed items regardless of --unmanaged flag
   - Structured output includes summary counts and managed items only

5. **Table Display Logic:**
   - Displays packages sorted alphabetically within each manager group
   - Shows different column layouts for managed vs unmanaged dotfiles
   - Implementation: Summary is completely hidden when `--unmanaged` is used in table format
   - Colors are applied to status words only, not full lines
   - Respects NO_COLOR environment variable for accessibility

6. **Configuration and File Status:**
   - Checks existence and validity of `plonk.yaml` and `plonk.lock`
   - Gracefully handles missing or invalid configuration files
   - Missing plonk.lock is treated as valid (dotfiles-only mode)

7. **Filter Processing:**
   - `--packages`: Shows only package-related results
   - `--dotfiles`: Shows only dotfile-related results
   - `--unmanaged`: Shows only untracked items (changes table format)
   - Multiple filters can be combined (e.g., `--packages --unmanaged`)

**Architecture Patterns:**
- Uses functional options pattern for orchestrator configuration
- Implements unified resource interface across domains
- Employs state-based reconciliation with explicit item categorization
- Graceful degradation when components are missing or invalid

**Error Conditions:**
- Continues operation even with invalid configuration
- Missing lock file treated as informational, not error
- Reconciliation errors are propagated but don't prevent partial results

**Integration Points:**
- Shares reconciliation engine with apply command
- Uses same configuration loading as other commands
- Output formatting consistent with other commands

**Bugs Identified:**
1. **JSON/YAML --unmanaged flag limitation**: `StructuredData()` method doesn't support `ShowUnmanaged` flag, always returns managed items only

## Improvements

- Sort items alphabetically instead of by package manager
- Review flag combination behavior (e.g., --packages --dotfiles redundancy)
- Consider built-in pagination for very long lists
- Add --missing flag to show only missing resources
- Consider adding color coding to package manager column for visual grouping
- Remove JSON/YAML output format support until there's a real use case
