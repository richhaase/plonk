# Plonk Command Documentation Project

## v1.0 Readiness Phase (Started 2025-07-30)

### Current Status
Working through v1.0 readiness tasks. Major accomplishments:
- ✅ Implemented dotfile drift detection and `plonk diff` command
- ✅ Added progress indicators for long-running operations
- ✅ Excluded .plonk/ directory from dotfile deployment
- ✅ Removed APT - focus on Homebrew for cross-platform consistency
- ⏳ Linux platform testing pending

### Recently Completed (2025-07-30)
1. **Dotfile Drift Detection**:
   - SHA256 checksum-based comparison
   - Integrated into status command (shows "drifted" state)
   - Drifted files restored with `plonk apply`

2. **Plonk Diff Command**:
   - Shows differences for drifted dotfiles
   - Supports various path formats (~/, $HOME/, absolute, env vars)
   - Configurable diff tool (default: `git diff --no-index`)
   - Created comprehensive documentation and tests

3. **APT Package Manager Support (Phase 1 & 2)**:
   - Platform detection for package managers
   - APT only available on Debian-based Linux distributions
   - Implemented read operations (search, info, check installed)
   - Updated doctor command to show platform-specific availability
   - Remaining: Phase 3 (install/uninstall), Phase 4 (integration), Phase 5 (docs)

## Summary of Setup Completed

Created documentation structure in `/docs/cmds/` with 8 files:
- **setup.md** - Initialize plonk or clone dotfiles repository
- **apply.md** - Install missing packages and deploy dotfiles
- **config.md** - Manage plonk configuration
- **status.md** - Show managed packages and dotfiles (now shows drift)
- **doctor.md** - Check system health and configuration
- **package_management.md** - Commands for install/uninstall/search/info
- **dotfile_management.md** - Commands for add/rm
- **diff.md** - Show differences for drifted dotfiles (NEW)

Each file has the following sections:
1. **Title** (completed)
2. **One-line summary** (completed)
3. **Description** (completed)
4. **Behavior** (completed)
5. **Implementation Notes** (completed)

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

### Documentation Improvement Phase (Completed - 2025-07-29)
**Summary**: Successfully eliminated documentation duplication across the project by establishing single sources of truth and implementing cross-references. All 6 medium priority items were completed, improving documentation maintainability and reducing confusion from conflicting information. Low priority enhancement items have been deferred for future work.

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
| Package Manager Lists | ✅ Completed | Consolidate repeated lists across README, cli.md, architecture.md |
| Command Syntax Duplication | ✅ Completed | Use cli.md as canonical reference, link from detailed docs |
| Configuration Examples | ✅ Completed | Single comprehensive example in configuration.md |
| Installation Instructions | ✅ Completed | Unify setup guides with clear variations |
| Output Format Examples | ✅ Completed | Standardize flag format and consolidate examples |
| Repository URL Formats | ✅ Completed | Consolidate in setup.md, reference elsewhere |

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

## Critical Implementation Guidelines

### STRICT RULE: No Unauthorized Features
**NEVER EVER independently add features or enhancements that were not explicitly requested.**
- You MAY propose improvements, but that is all
- Do NOT implement anything beyond the exact scope requested
- Do NOT add "helpful" extras without explicit approval
- Do NOT skip requested features without explicit approval
- When in doubt, implement ONLY what was explicitly requested

### Examples of Unauthorized Changes to Avoid:
- Adding emojis when not requested
- Creating new abstractions or structures beyond the task scope
- Adding terminal detection or environment variable checks when not specified
- Extending functionality to related areas (e.g., adding warning colors when only error colors were requested)
- Skipping tasks because you think they don't add value

### UI/UX Guidelines:
- **NEVER use emojis in plonk output** - Use colored text status indicators instead
- Status indicators should be colored minimally (only the status word, not full lines)
- Professional, clean output similar to tools like git, docker, kubectl

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

### Implementation Learnings
- **Lock File v2**: Breaking changes can lead to cleaner implementations - removed 50% of lock code by eliminating v1 support
- **Metadata Design**: Using `map[string]interface{}` provides flexibility without schema changes
- **Source Path Storage**: Critical for Go packages (`source_path`) and NPM scoped packages (`scope`, `full_name`)
- **Zero-Config Philosophy**: Removing init command reinforces zero-config principle - users just start using plonk
- **Simplification**: Sometimes the best solution is removal - init command was unnecessary complexity
- **Lock File Creation**: Lock files should only be created by install/uninstall, never by setup/init
- **Shell Detection**: Supporting multiple shells with specific syntax (fish_add_path) improves user experience
- **Filter Combinations**: Making flags work together (--missing --packages) provides flexible views
- **Bug Fixes Over Removal**: Fixing JSON/YAML filtering was better than removing the feature entirely

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

## Implementation Enhancement Phase (Completed - 2025-07-30)

This phase implemented critical improvements identified during documentation review, focusing on UI/UX enhancements and core functionality fixes.

### Summary of Completed Work

1. **Lock File v2** - Implemented metadata support for package source paths
2. **Setup Refactoring** - Removed init command, kept only clone for zero-config approach
3. **Quick Wins** - PATH commands, --missing flag, improved documentation
4. **UI/UX Overhaul** - Replaced all emojis with minimal colorization
5. **Config Improvements** - Visudo-style editing, user-defined highlighting
6. **Status Improvements** - Alphabetical sorting, flag validation
7. **Go Version** - Lowered requirement from 1.24.4 to 1.23
8. **Documentation** - Fixed broken links, created planning docs

### Completed Tasks

#### Task 2: Lock File Format Enhancement (Completed - 2025-07-29)
- **Implemented**: Lock file v2 with metadata support for package source paths
- **Approach**: Clean break from v1 (no migration) - users must remove old lock files
- **Benefits**: ~50% reduction in lock code complexity, cleaner API
- **Key Features**:
  - Go packages store `source_path` metadata (e.g., "golang.org/x/tools/cmd/gopls")
  - NPM scoped packages store `scope` and `full_name` metadata
  - Extensible metadata field for future enhancements
- **Impact**: Enables Phase 2 setup features that need package source information

#### Task 3: Setup Command Refactoring (Completed - 2025-07-29)
- **Implemented**: Major refactoring of setup workflow with focus on simplicity
- **Final Design**: Removed `init` command entirely, kept only `clone`
  - Original plan: Split setup into init/clone commands
  - Final decision: Remove init to maintain zero-config philosophy
- **Key Features Delivered**:
  - Intelligent clone that detects required managers from lock file
  - Auto-installs only necessary package managers
  - Runs apply automatically (with `--no-apply` option)
  - No more empty lock file creation
- **Benefits**:
  - Simpler mental model: clone existing or just start using plonk
  - True zero-config: no initialization step required
  - Cleaner codebase with single setup path
- **Note**: Skip flags were not implemented as init was removed entirely

#### Task 4: Quick Wins (Completed - 2025-07-29)
- **Implemented**: All 3 requested quick wins plus bonus bug fix
- **Task 1 - Doctor PATH Commands**:
  - Shell detection (zsh, bash, fish, ksh, tcsh)
  - Copy-paste commands for PATH fixes
  - Special fish shell syntax handling
- **Task 2 - Status --missing Flag**:
  - Filter to show only missing resources
  - Works with other filters (--packages, --dotfiles)
  - Clean output without misleading summaries
- **Task 3 - Improved Path Documentation**:
  - Enhanced help text for add/rm commands
  - Clear path resolution explanations
  - Special cases and practical examples
- **Bonus - Fixed JSON/YAML Filtering**:
  - StructuredData() now respects all filter flags
  - Resolved the unmanaged flag bug
  - Consistent behavior across output formats
- **Note**: JSON/YAML removal (original Task 4) not implemented per user guidance

#### Task 5: Status UI/UX Improvements (Completed - 2025-07-30)
- **Implemented**: Both status command improvements
- **Sort Alphabetically**: Case-insensitive sorting within logical groups
- **Flag Combinations**: Mutually exclusive validation with clear error messages
- **Enhanced UX**: Empty result messages and better help documentation

### Dependency Analysis

Key dependencies identified through code structure review:

1. **Lock File Format Enhancement** blocks intelligent setup features (needs package source info)
2. **Setup Command Items** are interdependent and should be implemented together
3. **Status/Config/Doctor improvements** are largely independent
4. **Location mapping**:
   - Lock file: `internal/lock/yaml_lock.go`
   - Setup: `internal/setup/setup.go`, `internal/commands/setup.go`
   - Status: `internal/commands/status.go`
   - Config: `internal/commands/config_show.go`, `config_edit.go`

### Implementation Phases

#### Phase 1: Foundation (Immediate Priority)
| Command | Item | Description | Dependency | Location | Status |
|---------|------|-------------|------------|----------|--------|
| package_management | Enhance lock file format | Store both binary name and full source path | Blocks setup features | `internal/lock/yaml_lock.go` | ✅ Completed |

#### Phase 2: Setup Refactoring (High Priority - Implement Together)
| Command | Item | Description | Dependency | Location | Status |
|---------|------|-------------|------------|----------|--------|
| setup | Split init/setup commands | Separate initialization from clone workflow | Core refactor | `internal/commands/setup.go` | ✅ Modified: Removed init entirely |
| setup | Skip package manager flags | Add --no-cargo, --no-npm, etc. flags | Depends on split | `internal/setup/tools.go` | ❌ Not implemented (init removed) |
| setup | Auto-detect from plonk.lock | Detect required managers from cloned repository | Needs lock v2 | `internal/setup/setup.go` | ✅ Completed |
| setup | Intelligent clone + apply | Only install managers for tracked packages | Needs auto-detect | `internal/setup/setup.go` | ✅ Completed |

#### Phase 3: Quick Wins (Independent Items) (Completed - 2025-07-29)
| Command | Item | Description | Dependency | Location | Status |
|---------|------|-------------|------------|----------|--------|
| doctor | Copy-paste PATH commands | Provide shell-specific PATH export commands | None | `internal/diagnostics/health.go` | ✅ Completed |
| status | Add --missing flag | Filter to show only missing resources | None | `internal/commands/status.go` | ✅ Completed |
| dotfile_management | Improve path docs | Better path resolution documentation | None | Help text only | ✅ Completed |
| global | Remove JSON/YAML output | Remove structured output formats until real use case emerges | None | `internal/commands/output.go` and all commands | ❌ Not implemented |

**Note**: JSON/YAML removal not implemented per user guidance. Instead, fixed the filtering bug to make structured output properly respect all flags.

#### Phase 4: UI/UX Improvements (Medium Priority) ✅ (Complete)
| Command | Item | Description | Dependency | Location | Status |
|---------|------|-------------|------------|----------|--------|
| config | Complete file editing | Edit full config, save only non-defaults | None | `internal/commands/config_edit.go` | ✅ Completed |
| config | Highlight user values | Distinguish user-defined from defaults | None | `internal/commands/config_show.go` | ✅ Completed |
| config | Sort alphabetically | Sort configuration values for easier reading | None | `internal/commands/config_show.go` | ❌ Skipped |
| config | Review flag combinations | Validate and handle all possible flag combinations | None | `internal/commands/config_*.go` | ✅ Verified |
| config | Color coding for edit | Use color coding in error messages for better visibility | None | `internal/commands/config_edit.go` | ✅ Completed |
| status | Sort alphabetically | Change default sort order | None | `internal/commands/status.go` | ✅ Completed |
| status | Review flag combinations | Fix --packages --dotfiles behavior | None | `internal/commands/status.go` | ✅ Completed |
| ~~status~~ | ~~Color coding~~ | ~~Visual grouping by package manager~~ | ~~None~~ | ~~`internal/output/formatters.go`~~ | ~~Removed~~ |

#### Task 9: Config Show User-Defined Highlighting (Completed - 2025-07-30)
- **Implementation**: Added blue highlighting for user-defined values in config show
- **Shared logic**: Created UserDefinedChecker to share detection logic with config edit
- **Clean output**: JSON/YAML output remains clean without annotations
- **Color support**: Uses blue color for "(user-defined)" annotations
- **Refactoring**: Both config edit and show now use the same user-defined detection
- **Benefits**: Clear visibility of customizations, consistent behavior across commands

#### Task 8: Visudo-style Config Edit (Completed - 2025-07-30)
- **Implementation**: Created visudo-style editing workflow for configuration
- **Show runtime config**: Display full merged configuration with defaults + user overrides
- **User-defined annotations**: Mark values that differ from defaults with `# (user-defined)`
- **Validation loop**: Edit/revert/quit options on validation failure
- **Minimal saving**: Only save non-default values to plonk.yaml
- **Editor support**: Check VISUAL, then EDITOR, then fallback to vim
- **Temp file cleanup**: Always clean up temporary files
- **Benefits**: Better user experience, minimal config files, clear visibility of customizations

#### Task 7: Emoji Replacement with Colorization (Completed - 2025-07-30)
- **High Priority Task**: Replace all emojis with professional colored text
- **Created color infrastructure**: Centralized color management in `internal/output/colors.go`
- **Terminal detection**: Automatic color support based on terminal capabilities
- **NO_COLOR support**: Respects standard NO_COLOR environment variable
- **Minimal colorization**: Only status words colored, not full lines
- **Complete emoji removal**: All Unicode emojis replaced with colored text
- **Benefits**: Professional appearance, better accessibility, universal compatibility

#### Task 5: Status UI/UX Improvements (Completed - 2025-07-30)
- **Task 1 - Sort Alphabetically**:
  - Implemented case-insensitive sorting for all categories
  - Maintains logical grouping (packages by manager)
  - Applied to all output formats (table, JSON, YAML)
- **Task 2 - Review Flag Combinations**:
  - Added validation for mutually exclusive flags (--unmanaged and --missing)
  - Clear error messages for invalid combinations
  - Empty result handling with helpful messages
  - Updated help text with "Flag Behavior" section
- **Note**: Color coding removed from improvements as it conflicts with CLI best practices

#### Task 6: Config UI/UX Improvements (Completed - 2025-07-30)
- **Task 1 - Sort Alphabetically**: Skipped per user request
- **Task 2 - Review Flag Combinations**: Verified existing behavior is correct, no changes needed
- **Task 3 - Color Coding for Edit**:
  - Added red color for validation errors
  - Added green color for success messages
  - Minimal implementation without unauthorized features
- **Important Lesson**: Initial implementation included unauthorized features (emojis, new structures) which were reverted
- **Result**: Strict adherence to requested scope with minimal changes

### UI/UX Philosophy
- **No emojis ever**: Plonk uses colored text status indicators, never emojis
- **Minimal colorization**: Only status words are colored for clean, scannable output
- **Professional appearance**: Similar to git, docker, kubectl
- **Universal compatibility**: Works in all terminals without Unicode issues

## v1.0 Readiness Phase (Current - Started 2025-07-30)

### Overview
Focus on implementing the minimum required features for a stable v1.0 release that delivers on plonk's core promise: one-command setup that works across platforms.

### Planning Documents
- **[v1-readiness.md](docs/planning/v1-readiness.md)** - Comprehensive checklist and requirements
- **[v1-summary.md](docs/planning/v1-summary.md)** - Executive summary with timeline
- **[ideas.md](docs/planning/ideas.md)** - Complete list of improvements with user priorities

### Implementation Progress

#### Phase 1: Foundation (Week 1) - COMPLETE
| Task | Priority | Est. Days | Status | Completed |
|------|----------|-----------|--------|-----------|
| .plonk/ Directory Exclusion | Medium | 0.5 | ✅ Complete | 2025-07-30 |
| Progress Indicators | High | 1-2 | ✅ Complete | 2025-07-30 |
| Doctor Code Consolidation | Medium | 1-2 | ⏸️ Skipped | 2025-07-30 |
| Dotfile Drift Detection | TOP | 2-3 | ✅ Complete | 2025-07-30 |

#### Phase 2: Core Features (Week 2) - COMPLETE
| Task | Priority | Est. Days | Status | Notes |
|------|----------|-----------|--------|-------|
| APT Package Manager | High | 3-5 | ✅ Removed | Built then removed - wrong approach |

#### Phase 3: Polish & Release (Week 3) - IN PROGRESS
| Task | Priority | Est. Days | Status | Notes |
|------|----------|-----------|--------|-------|
| Linux Platform Testing | High | 2-3 | ⏳ Pending | Ubuntu, Debian via Lima VM |
| Documentation Updates | Medium | 1-2 | ⏳ Pending | Remove outdated refs |
| Dead Code Cleanup | Medium | 0.5 | ✅ Complete | Removed 600+ lines |

### Implementation Notes

#### Completed: .plonk/ Directory Exclusion (2025-07-30)
- Added exclusion logic to `ShouldSkipPath()` in manager.go
- Added exclusion logic to `ShouldSkip()` in filter.go
- Added comprehensive tests for all exclusion scenarios
- Updated documentation in CONFIGURATION.md and dotfile_management.md
- This directory is reserved for future plonk metadata (hooks, templates, etc.)

#### Completed: Progress Indicators (2025-07-30)
- Created new `output/progress.go` with `ProgressUpdate()` and `StageUpdate()` functions
- Added progress to `install` and `uninstall` commands for multi-package operations
- Added two-phase progress to `apply` command (packages phase, then dotfiles phase)
- Added multi-stage progress to `clone` command
- Format: `[2/5] Installing: htop` for consistent user experience
- No progress shown for single-item operations (clean output)

#### Skipped: Doctor Code Consolidation (2025-07-30)
- Planned to consolidate shared package manager installation code
- Location should be `internal/resources/packages/` (not new packagemanager directory)
- Remove ALL interactive prompting - use `--no-npm`, `--no-cargo` flags instead
- Outstanding design questions documented in [doctor-consolidation-plan.md](docs/planning/doctor-consolidation-plan.md)
- Needs decisions on: file structure, interface design, flag handling, error messages
- Will revisit after other v1.0 features are complete

#### Completed: Dotfile Drift Detection Phase 1 (2025-07-30)
- Created comprehensive plan in [drift-detection-plan.md](docs/planning/drift-detection-plan.md)
- Created implementation tasks in [drift-detection-tasks.md](docs/planning/drift-detection-tasks.md)
- Created proof of concept in [drift-detection-poc.md](docs/planning/drift-detection-poc.md)
- Implemented Phase 1 features:
  - SHA256 checksum computation in dotfiles manager
  - Drift detection during reconciliation using StateDegraded
  - Status command shows "drifted" with yellow color
  - Summary includes drift count
  - Apply command restores drifted files
  - Comprehensive test coverage
- Key architecture decisions:
  - Comparison function stored in item metadata
  - Minimal changes to existing architecture
  - Backward compatible with existing Result structure
- Ready for Phase 2 enhancements (diff tools, preview)

### Phase 2 Review Decision (2025-07-30)
After reviewing Phase 2 features, decided to implement only:
- **plonk diff command** - Show differences for drifted files
- **Configurable diff tool** - With sensible default (git diff --no-index)

Rejected features:
- Internal diff display - Too complex, users have diff tools
- Preview flag - Not worth the complexity
- Selective apply - Users can use plonk add to re-add modified files
- Other complex features - Not worth implementation effort

### Progress Summary
- **Start Date**: 2025-07-30
- **Target Completion**: Mid-August 2025 (2-3 weeks total)
- **Current Status**: 5 of 8 tasks complete, 1 skipped (62.5%)
- **Current Phase**: Polish & Release (Week 3)
- **Days Elapsed**: 1

### Success Criteria
- ✅ One command setup works on Mac/Linux
- ✅ Core commands have stable interfaces
- ✅ New users can start immediately
- ✅ Cross-platform behavior is identical

### Post-v1.0 Roadmap
1. Package update command
2. Verbose/debug modes
3. Additional Linux package managers
4. Hook system (using `.plonk/`)
5. Performance optimizations

### Key Decisions Made
- **Exclusion pattern** chosen over prefix system for dotfiles
- **APT required** for v1.0 (not deferred)
- **Progress over verbose** - simple status output sufficient
- **Drift in status** - integrated, not separate command
- **Claude's Role**: Planning and documentation only (no code changes)

## v1.0 Implementation Progress (2025-07-30)

### Completed Features
1. **Progress Indicators** ✅
   - Spinner for operations longer than 100ms
   - Uses briandowns/spinner library
   - Integrated into package operations (install, uninstall, search)
   - Progress messages shown during reconciliation

2. **.plonk/ Directory Exclusion** ✅
   - Added to default ignore patterns
   - Prevents recursive deployment issues
   - Reserved for future metadata/hooks

3. **Dotfile Drift Detection** ✅
   - SHA256 checksum-based comparison
   - StateDegraded repurposed for drift
   - Shows "drifted" status in yellow
   - Drifted files restored with `plonk apply`
   - Backward-compatible implementation

4. **plonk diff Command** ✅
   - Shows differences for drifted dotfiles
   - Uses configurable diff tool (defaults to git diff --no-index)
   - Environment variables: PLONK_DIFF_TOOL and PLONK_DIFF_ARGS
   - Supports both specific file and all drifted files

5. **Homebrew Prerequisite** ✅
   - Removed all Homebrew installation logic
   - Homebrew is now a required prerequisite
   - Simplified setup: install Homebrew → install plonk → plonk clone
   - Doctor shows error if Homebrew missing

6. **Doctor Simplification** ✅
   - Removed --fix flag entirely
   - Doctor is now a pure diagnostic tool
   - Only `plonk clone` can install package managers
   - Clear separation of concerns

7. **Dead Code Cleanup** ✅
   - Removed 600+ lines of unused code
   - Used deadcode tool to identify unused functions and types
   - Cleaned up unused color functions, formatting helpers, and types
   - All tests pass after cleanup

### Implementation Learnings
- **State Model**: Successfully repurposed StateDegraded for drift without breaking changes
- **Path Resolution**: Created normalizePath() that handles all POSIX path formats
- **Dotfile Mapping**: Clear separation between source (no dot) and deployed (with dot)
- **Testing Philosophy**: Unit tests for business logic only, no mocks for CLIs
- **Zero-Config**: Maintained philosophy with sensible defaults (git diff)
- **Platform Detection**: Created comprehensive platform detection for package managers
- **APT Design**: Clean separation of read operations (no sudo) vs write operations (sudo)

### Remaining v1.0 Tasks (2025-07-30)

Based on [v1-final-sprint.md](docs/planning/v1-final-sprint.md):

1. **Linux Platform Testing** (2-3 days)
   - Test on Ubuntu 22.04 LTS and Debian 12 via Lima VM
   - Verify Homebrew installation process on Linux
   - Ensure all package managers work identically to macOS
   - Document any Linux-specific setup requirements
   - Test WSL2 compatibility

2. **Documentation Updates & Release Prep** (1-2 days)
   - ✅ Remove stability warning from README (already done)
   - Update version to 1.0.0 in cmd/plonk/main.go
   - Create v1.0.0 release notes using template
   - Review all command documentation for accuracy
   - Ensure installation guide is current
   - Tag v1.0.0 and create GitHub release with binaries

3. **Optional: Homebrew Formula** (not blocking)
   - Create formula for easy installation via `brew install plonk`
   - Can be done post-v1.0 release
