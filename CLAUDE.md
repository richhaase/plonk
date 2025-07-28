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
| setup.md | Not Started | - | |
| apply.md | ✅ Completed | 2025-07-28 | Documented reconciliation and resource states |
| config.md | ✅ Completed | 2025-07-28 | Documented show/edit subcommands |
| status.md | Not Started | - | |
| doctor.md | Not Started | - | |
| package_management.md | Not Started | - | |
| dotfile_management.md | Not Started | - | |

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
- TBD as we document

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
