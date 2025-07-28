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
3. **Description** (empty - optional expansion of summary)
4. **Behavior** (empty - main content explaining expected behavior)
5. **Implementation Notes** (empty - high-level implementation details)

## Progress Tracking

| Document | Status | Last Updated | Notes |
|----------|--------|--------------|-------|
| setup.md | ✅ Completed | 2025-07-28 | Documented dual modes and doctor integration |
| apply.md | ✅ Completed | 2025-07-28 | Documented reconciliation and resource states |
| config.md | ✅ Completed | 2025-07-28 | Documented show/edit subcommands |
| status.md | ✅ Completed | 2025-07-28 | Documented resource states and output formats |
| doctor.md | ✅ Completed | 2025-07-28 | Documented health checks and fix behavior |
| package_management.md | ✅ Completed | 2025-07-28 | Documented four commands with prefix syntax |
| dotfile_management.md | ✅ Completed | 2025-07-28 | Documented add/rm with filesystem-as-state |

## Documentation Process

For each file, we will follow this process:

1. **Review Progress**: Present user with documents needing updates and ask which to work on next
2. **User Interview**: Ask user to describe how the command should behave
3. **Clarification**: Ask clarifying questions to fully understand the behavior
4. **Write Content**: Complete the Description and Behavior sections (skip Implementation Notes)
5. **Refinement**: Work with user to improve the document
6. **Commit Document**: When complete, commit the documented file
7. **Update CLAUDE.md**:
   - Update progress tracking status
   - Document any learnings (style preferences, component references, etc.)
8. **Commit CLAUDE.md**: Save progress and learnings
9. **Repeat**: Continue with next document

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
