# Plonk Command Documentation Project

## Summary of Setup Completed

Created documentation structure in `/docs/cmds/` with 7 files:
- **setup.md** - Initialize plonk or clone dotfiles repository
- **apply.md** - Install missing packages and deploy dotfiles
- **config.md** - Manage plonk configuration
- **status.md** - Show managed packages and dotfiles
- **doctor.md** - Check system health and configuration
- **package_management.md** - Commands for install/uninstall/search/info
- **dotfile_management.md** - Commands for add/rm

Each file has the following sections:
1. **Title** (completed)
2. **One-line summary** (completed)
3. **Description** (completed)
4. **Behavior** (completed)
5. **Implementation Notes** (in progress - adding implementation details)

## Progress Tracking

### Behavior Documentation Phase (Completed)
| Document | Status | Last Updated | Notes |
|----------|--------|--------------|-------|
| setup.md | ✅ Completed | 2025-07-28 | Documented dual modes and doctor integration |
| apply.md | ✅ Completed | 2025-07-28 | Documented reconciliation and resource states |
| config.md | ✅ Completed | 2025-07-28 | Documented show/edit subcommands |
| status.md | ✅ Completed | 2025-07-28 | Documented resource states and output formats |
| doctor.md | ✅ Completed | 2025-07-28 | Documented health checks and fix behavior |
| package_management.md | ✅ Completed | 2025-07-28 | Documented four commands with prefix syntax |
| dotfile_management.md | ✅ Completed | 2025-07-28 | Documented add/rm with filesystem-as-state |
| architecture.md | ✅ Completed | 2025-07-28 | Documented project architecture and design principles |
| why-plonk.md | ✅ Completed | 2025-07-28 | Explained project motivation and unique value proposition |

### Implementation Documentation Phase (Completed)
| Document | Status | Last Updated | Notes |
|----------|--------|--------------|-------|
| setup.md | ✅ Completed | 2025-07-28 | Added implementation section, found 5 behavior discrepancies |
| apply.md | ✅ Completed | 2025-07-28 | Added implementation section, found 4 behavior discrepancies |
| config.md | ✅ Completed | 2025-07-28 | Added implementation section, found 3 behavior discrepancies |
| status.md | ✅ Completed | 2025-07-28 | Added implementation section, found 3 behavior discrepancies |
| doctor.md | ✅ Completed | 2025-07-28 | Added implementation section, found 5 behavior discrepancies |
| package_management.md | ✅ Completed | 2025-07-28 | Added implementation section, found 4 behavior discrepancies |
| dotfile_management.md | ✅ Completed | 2025-07-28 | Added implementation section, found 3 behavior discrepancies |

**Total**: 27 behavior discrepancies identified across all commands

### Discrepancy Resolution Phase (Completed)
| Document | Total Discrepancies | Resolved | Remaining | Status |
|----------|-------------------|----------|-----------|--------|
| setup.md | 5 | 5 | 0 | ✅ Completed |
| apply.md | 4 | 4 | 0 | ✅ Completed |
| config.md | 3 | 3 | 0 | ✅ Completed |
| status.md | 3 | 3 | 0 | ✅ Completed |
| doctor.md | 5 | 5 | 0 | ✅ Completed |
| package_management.md | 4 | 4 | 0 | ✅ Completed |
| dotfile_management.md | 3 | 3 | 0 | ✅ Completed |

### Documentation Improvement Phase (Current)
For each documentation improvement item, we follow this process:
1. **Present Item**: Show the duplication/issue and proposed resolution(s)
2. **Query User**: Ask for additional context and resolution preferences
3. **Synthesize Response**: Refine resolution based on user feedback (repeat until approved)
4. **Apply Resolution**: Implement the user-approved solution
5. **Update Progress**: Mark item as resolved in tracking
6. **Repeat**: Continue with next item

#### Medium Priority Items (Eliminate Duplication)
| Item | Status | Description |
|------|--------|-------------|
| Package Manager Lists | ⏳ Pending | Consolidate repeated lists across README, cli.md, architecture.md |
| Command Syntax Duplication | ✅ Completed | Use cli.md as canonical reference, link from detailed docs |
| Configuration Examples | 🔄 In Progress | Single comprehensive example in configuration.md |
| Installation Instructions | ⏳ Pending | Unify setup guides with clear variations |
| Output Format Examples | ⏳ Pending | Standardize flag format and consolidate examples |
| Repository URL Formats | ⏳ Pending | Consolidate in setup.md, reference elsewhere |

#### Low Priority Items (Content Enhancement)
| Item | Status | Description |
|------|--------|-------------|
| Add Troubleshooting Guide | ⏳ Pending | Common errors, solutions, debugging tips |
| Create Migration Guide | ⏳ Pending | Version upgrade procedures and breaking changes |
| Document Performance | ⏳ Pending | Timeouts, limits, scaling considerations |
| Add Integration Examples | ⏳ Pending | Real-world dotfiles repository structures |
| Improve Structure | ⏳ Pending | Clear user vs developer sections |
| Add Cross-References | ⏳ Pending | Better linking between related topics |
| Add Visual Aids | ⏳ Pending | Diagrams for state model and architecture |
| Expand Examples | ⏳ Pending | More real-world scenarios and use cases |
| Add FAQ Section | ⏳ Pending | Common questions and answers |

## Documentation Process

### Phase 1: Behavior Documentation (Completed)
For each file, we followed this process:
1. **Review Progress**: Present user with documents needing updates and ask which to work on next
2. **User Interview**: Ask user to describe how the command should behave
3. **Clarification**: Ask clarifying questions to fully understand the behavior
4. **Write Content**: Complete the Description and Behavior sections (skip Implementation Notes)
5. **Refinement**: Work with user to improve the document
6. **Commit Document**: When complete, commit the documented file
7. **Update CLAUDE.md**: Update progress tracking status and document learnings
8. **Commit CLAUDE.md**: Save progress and learnings
9. **Repeat**: Continue with next document

### Phase 2: Implementation Documentation (Completed)
For each file, we followed this process:
1. **Select Document**: Choose which command documentation to work on
2. **Code Review**: Perform complete review of existing implementation code
3. **Behavior Comparison**: Compare documented behavior vs actual code behavior
   - Document any discrepancies as bugs to fix
   - Note any undocumented features or behaviors
4. **Generate Implementation Section**: Create brief, complete implementation notes
   - Focus on high-level architecture and flow
   - Avoid code examples unless absolutely necessary
   - Use ASCII diagrams for complex flows (ask when in doubt)
5. **Review & Approval**: Present implementation section for user approval
6. **Write to Document**: Add approved implementation section
7. **Update CLAUDE.md**: Document learnings and implementation patterns
8. **Commit**: Save the updated document
9. **Repeat**: Continue with next document

### Phase 3: Discrepancy Resolution (Current)
For each file, we will follow this process:
1. **Pick File**: Choose which command documentation to address
2. **Review Discrepancies**: Examine all documented discrepancies for the file
3. **For Each Discrepancy**:
   a. **Determine Resolution**: Decide whether to adjust code, documentation, or other solution
   b. **Plan Resolution**: Create specific plan for implementing the fix
   c. **Execute Resolution**: Implement the planned changes
   d. **Update Documentation**: Reflect changes in documentation as needed
   e. **Commit Fix**: Commit changes to docs, code, or both
   f. **Update Progress**: Mark discrepancy as resolved in tracking
4. **Repeat**: Continue until all discrepancies addressed for the file
5. **Next File**: Move to next file with discrepancies
6. **Complete**: All 27 discrepancies across 7 files resolved

## Documentation Guidelines and Learnings

### Style Preferences
- Technical audience assumed
- Brevity and clarity are key
- Use bullets and ASCII diagrams in Behavior section
- Cross-reference related commands
- Group related commands (e.g., install/uninstall/search/info)

### Known Component References
- `plonk doctor --fix` - Used by setup for health checks and package manager installation
- `plonk apply` - Automatically run by setup after cloning repository

### Implementation Documentation Guidelines
- Keep implementation notes brief and architecture-focused
- Avoid code examples unless absolutely necessary for clarity
- Focus on high-level flow and component interactions
- Use ASCII diagrams for complex flows (ask user first)
- Document discrepancies between documented and actual behavior
- Note any bugs found during code review but don't fix them

### Implementation Review Learnings
- **Config Command**: Found 3 discrepancies:
  1. Edit command validates and shows errors (not silently ignored as documented)
  2. Edit creates template config file, not actual defaults
  3. Documentation missing `$VISUAL` environment variable
- **Apply Command**: Found 4 discrepancies:
  1. `--packages` and `--dotfiles` are mutually exclusive (not redundant as documented)
  2. `--backup` flag is functional (not "under review" as documented)
  3. Hook execution not mentioned in documented flow
  4. Lock file updates during apply not documented
- **Setup Command**: Found 5 discrepancies:
  1. Does NOT delegate to `plonk doctor --fix` (uses custom implementation)
  2. Apply requires user confirmation (not automatic after clone)
  3. Checks for specific plonk files (not just "empty directory")
  4. Supports additional git:// protocol (undocumented)
  5. Missing PATH configuration guidance in docs
- **Doctor Command**: Found 5 discrepancies:
  1. Status terminology mismatch (pass/warn/fail vs PASS/WARN/ERROR)
  2. Overall status values differ (healthy/warning/unhealthy vs PASS/WARNING/ERROR)
  3. Package Manager Functionality identical to Availability (not differentiated)
  4. Uses emoji icons not color coding as documented
  5. Missing 30-second timeout in documentation
- **Status Command**: Found 3 discrepancies:
  1. JSON/YAML output ignores `--unmanaged` flag (always shows managed items)
  2. Flag combination behavior differs (`--packages --dotfiles` not redundant)
  3. Summary hidden completely for --unmanaged instead of showing untracked counts
- **Dotfile Management Commands**: Found 3 discrepancies:
  1. Non-functional `--force` flags defined but never used in implementation
  2. Help text mentions "preserve original files" but implementation always copies
  3. Inconsistent config loading between add (can fail) and remove (never fails)
- **Package Management Commands**: Found 4 discrepancies:
  1. Non-functional `--force` flag in uninstall command (defined but never used)
  2. Undocumented cargo search exclusion (hardcoded skip in implementation)
  3. Missing timeout documentation (5min install, 3min uninstall, 3sec search not mentioned)
  4. Go package name transformation undocumented (module path → binary name)
  - **Investigation Outcome**: Identified significant lock file format limitation where Go packages lose source path information (`golang.org/x/tools/cmd/gopls` → `gopls`). Proposed v2 lock format enhancement using metadata field to store both binary name and source path. Similar issue affects npm scoped packages.
- **Pattern**: Use "DISCREPANCY" to clearly mark behavior differences
- **Structure**: Organize by Command Structure, Key Details, Bugs Identified
- **Complex Commands**: Apply and Setup show layered architecture with orchestration patterns
- **Modular Design**: Setup shows good separation with dedicated files for git, tools, prompts
- **Consistent Issues**: Status terminology and timeout documentation frequently missing

### Other Learnings
- Pre-commit hooks are active and will auto-fix formatting issues (end-of-file, trailing whitespace)
- Include "Improvements" section in each document for future enhancement ideas
- Configuration uses zero-config approach with sensible defaults
- Output formats (table/json/yaml) have different field naming conventions:
  - JSON uses PascalCase
  - YAML uses snake_case
  - Table format is human-readable YAML-like
- Environment variables: PLONK_DIR controls config location, EDITOR controls edit command
- Resource states: managed (known + exists), missing (known + doesn't exist), unmanaged (unknown + exists)
- Apply command executes packages first, then dotfiles
- Plonk.lock absence is valid (dotfiles-only mode)
- Dotfile mapping: automatic dot-prefix handling ($PLONK_DIR/vimrc → $HOME/.vimrc)
- Features marked for review should be noted in documentation (e.g., --backup flag)
- Setup has dual modes: fresh initialization vs clone + apply workflow
- Package managers categorized as bootstrap (Homebrew, Cargo) vs language (npm, pip, gem, go)
- Required vs optional package managers: Homebrew is required, others are optional
- plonk.lock is only created by install/uninstall commands, never by setup
- Clone operations require empty $PLONK_DIR directory
- Doctor has three status levels: PASS, WARN, ERROR (overall status = worst found)
- Doctor checks six categories: System, Environment, Permissions, Configuration, Package Managers, Installation
- Fix behavior currently limited to package manager installation
- PATH configuration issues are informational only (not auto-fixable)
- Package manager "availability" and "functionality" checks are currently redundant
- Doctor output formats preserve different structures (table is hierarchical, json/yaml are flat arrays)
- Status command is NOT the default (plain `plonk` shows help)
- Status has alias `st` for convenience
- Status shows different columns for unmanaged dotfiles (simplified view)
- Current bug: --unmanaged flag doesn't affect JSON/YAML output
- Status uses "domain" terminology in structured output (dotfile/package)
- Always shows summary counts regardless of filters
- Dotfile state is filesystem-based: contents of $PLONK_DIR ARE the state
- File mapping removes/adds leading dot: ~/.zshrc ↔ $PLONK_DIR/zshrc
- Dotfiles within $PLONK_DIR (like .git) are ignored to prevent deployment as ..git
- Add command always overwrites (no warnings) - assumes git backup
- Remove only affects $PLONK_DIR, never touches files in $HOME
- --force flags are non-functional and should be removed
- Package management supports 6 managers with prefix syntax (brew:, npm:, etc.)
- Install gracefully handles already-installed packages (adds to management)
- Search has 3-second timeout per manager when searching all
- Info command has 3-tier priority: managed > installed > available
- Keep documentation concise and behavior-focused (not implementation-focused)

### Documentation Improvement Learnings
- **Critical vs Medium Priority**: User-facing inaccuracies (wrong commands, incorrect behavior) are critical; duplication and organization are medium priority
- **Single Source of Truth**: Eliminate duplication by designating canonical locations for information types
- **Cross-Reference Strategy**: Link between docs rather than repeat information
- **Process Iteration**: Present → Query → Synthesize → Apply → Repeat ensures user alignment
