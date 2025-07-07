# TODO

Session-level work items and progress tracking. Maintained by AI agents for tactical execution.

## Current Session Focus

**Primary Goal:** Execute Stage 1 dogfooding - End-to-End Workflow Validation
**Session Date:** July 7, 2025
**Estimated Time:** 2 hours (as per ROADMAP.md)

## Active Work Items

### Completed This Session
- [x] Documentation cleanup for ZSH/Git generation removal ✅
- [x] Documentation file organization and duplication cleanup ✅
- [x] Release v0.3.0 with changelog update ✅
- [x] Review ROADMAP.md for dogfooding scope ✅

### Stage 1 Dogfooding Tasks (In Progress)

#### 1. Fresh Installation Test
- [ ] **Clean Environment Setup** - Create test scenario in isolated directory
- [x] **Installation Process** - Test `go install ./cmd/plonk` with proper version injection ✅
- [x] **Version Verification** - Confirm `plonk version` shows v0.3.0 with correct git commit ✅
- [ ] **Help System Test** - Verify `plonk --help` and `plonk <command> --help` work correctly
- [ ] **Global Flags Test** - Test `--dry-run` flag functionality across commands

#### 2. Import Functionality Test (Rich's Real Environment)
- [ ] **Pre-Import Baseline** - Document current environment (139 brew, 10 asdf, 6 npm packages)
- [ ] **Import Execution** - Run `plonk import` and verify plonk.yaml generation
- [ ] **Package Discovery Accuracy** - Verify all discovered packages match actual installed packages
  - Compare `plonk import` output vs `brew list`, `asdf list`, `npm list -g`
- [ ] **Dotfile Detection** - Verify .zshrc, .gitconfig, .zshenv are detected correctly
- [ ] **YAML Structure Validation** - Ensure generated plonk.yaml follows expected schema
- [ ] **Import Dry-Run Test** - Test `plonk import --dry-run` shows what would be discovered
- [ ] **Complex Environment Handling** - Test import with Rich's extensive dotfiles setup

#### 3. Status Command Accuracy
- [ ] **Status Before Import** - Document current drift detection (3 missing config files)
- [ ] **Status After Import** - Verify status changes appropriately after import
- [ ] **Package Count Validation** - Confirm `plonk status` counts match actual installed packages
- [ ] **Drift Detection Accuracy** - Test config file drift detection with various scenarios
- [ ] **Package List Commands** - Test `plonk pkg list` for each manager (brew, asdf, npm)
- [ ] **List Command Aliases** - Verify `plonk ls` alias works correctly

#### 4. Setup Workflow Validation (Clean Environment)
- [ ] **Baseline Check** - Document current system state before setup
- [ ] **Setup Command Test** - Run `plonk setup` and verify it installs missing tools
- [ ] **Dependency Validation** - Confirm Homebrew, ASDF, and Node.js/NPM are available post-setup
- [ ] **Idempotency Test** - Run `plonk setup` again to ensure no errors when tools already exist
- [ ] **Error Handling** - Test behavior when some tools are already installed vs missing

#### 5. End-to-End Workflow Test
- [ ] **Complete New User Flow** - Simulate fresh user: setup → import → status → pkg list
- [ ] **Configuration Application** - Test `plonk apply` with generated configuration
- [ ] **Backup Integration** - Test `plonk apply --backup` functionality
- [ ] **Restore Functionality** - Test `plonk restore` if backups created
- [ ] **Error Recovery** - Test behavior when operations fail midway

#### 6. Edge Cases & Error Scenarios
- [ ] **Missing Package Managers** - Test behavior when tools aren't installed
- [ ] **Permission Issues** - Test behavior with file permission problems
- [ ] **Invalid Configuration** - Test behavior with malformed plonk.yaml
- [ ] **Network Issues** - Test behavior when package managers can't reach repositories
- [ ] **Partial State Recovery** - Test recovery from interrupted operations

#### 7. Document Workflow Issues
- [ ] **Usability Pain Points** - Record any confusing or inefficient workflows
- [ ] **Error Message Quality** - Evaluate error messages for clarity and helpfulness
- [ ] **Performance Issues** - Note any slow operations or optimization opportunities
- [ ] **Missing Features** - Identify functionality gaps discovered during testing
- [ ] **Success Metrics** - Document what worked well and exceeded expectations

### Critical Success Criteria
- Complete workflow works from scratch
- Import captures Rich's environment accurately  
- No major usability blockers identified

## Dogfooding Phase Overview (6-8 hours total)

**Stage 1:** End-to-End Workflow Validation (~2 hours) - **CURRENT FOCUS**
**Stage 2:** Real Environment Migration (~2-3 hours)
**Stage 3:** Integration Testing & Documentation (~1-2 hours)  
**Stage 4:** UX Refinement & Polish (~1-2 hours)

## Current Blockers

- **Security Findings** - 34 gosec issues need resolution before public release  
- **Missing Integration Tests** - Need to create based on dogfooding scenarios

## Temporary Changes

- **Pre-commit Hooks Disabled** - Temporarily disabled to allow smooth dogfooding without security check failures

## Context for Next Session

- **Ready for Stage 1** - Dogfooding plan established, documentation clean
- **Key Decision** - Real-world validation before GitHub launch
- **Security Note** - Dogfooding can proceed while security fixes are planned

## Session Notes

**Working Patterns for This Session:**
1. **Focus Area**: Status command accuracy dogfooding - testing existing `plonk status` implementation against real environment
2. **Documentation Updates**: Update *.md docs as we go to ensure accuracy (design decisions → ARCHITECTURE.md)
3. **Code Changes**: Use TDD red→green→refactor cycles, small changes, frequent commits
4. **Scope Boundaries**: Plonk does only two things - manage packages from multiple package managers, and manage configuration files (dotfiles)

## Notes

*For session history, see git log and CHANGELOG.md. For strategic planning, see ROADMAP.md.*