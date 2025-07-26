# Phase 14: Additional UX Improvements

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase focuses on integration, polish, and ensuring all the UX improvements from Phases 9-13 work well together. It also addresses any remaining items from the Phase 8 UX review and ensures consistent behavior across all commands.

## Objectives

1. Ensure apply command continues on partial failures
2. Update all remaining help text and examples
3. Fix any integration issues between phases
4. Add command aliases where helpful
5. Ensure consistent error handling across commands
6. Implement BATS test suite for CLI validation

## Key Areas to Address

### 1. Apply Command Failure Handling

Based on user feedback: "apply should report failures and move on"

**In orchestrator or apply command:**
- Continue processing all packages/dotfiles even if some fail
- Collect all errors and report at the end
- Exit with non-zero code if any failures occurred
- Show clear summary of what succeeded and what failed

Example output:
```
Applying configuration...

✓ Installed: brew:ripgrep
✗ Failed: brew:unknown-package (package not found)
✓ Installed: npm:typescript
✓ Linked: ~/.gitconfig

Summary: 3 succeeded, 1 failed
```

### 2. Command Examples Update

Update all command examples to use new patterns:
- No more `plonk sync` → use `plonk apply`
- No more `--brew` flags → use `brew:package`
- Show `plonk st` as valid alias
- Remove any `plonk ls` references

### 3. Help Text Consistency

Review and update:
- Root command help to reflect all changes
- Subcommand help for consistency
- Remove any outdated examples
- Ensure prefix syntax is explained clearly

### 4. Integration Verification

Test scenarios combining multiple phases:
- `plonk apply` with invalid config (Phase 9 + 10)
- `plonk install brew:pkg npm:pkg` (Phase 12)
- `plonk info` on managed package after `plonk apply` (Phase 13)
- Ensure `plonk` shows help, not status (Phase 11)

### 5. Error Message Consistency

Ensure all error messages follow the pattern established in WORKER_CONTEXT.md:
- What went wrong
- How to fix it
- Consistent formatting

### 6. Hook Name Updates

If not already done in Phase 10:
- Ensure hooks use `pre_apply` and `post_apply`
- Update any example configs
- Clear error if old hook names are used

### 7. BATS Test Suite Implementation

Create comprehensive CLI tests using BATS (Bash Automated Testing System):

**Setup:**
- Create `tests/bats/` directory structure
- Add README with clear warnings about system modifications
- Implement test_helper.bash for common functions

**Test Coverage:**
```bash
tests/bats/
├── 01-basic-commands.bats    # Help, version, init
├── 02-config-commands.bats   # Config operations
├── 03-apply-command.bats     # Apply with partial failures
├── 04-package-ops.bats       # Install/uninstall with prefix syntax
├── 05-search-info.bats       # Parallel search, info priority
├── 06-status-commands.bats   # Status and st alias
├── 07-output-formats.bats    # JSON/YAML/table formats
└── 99-cleanup.bats          # Clean up test artifacts
```

**Key Tests:**
- Verify `plonk` (no args) shows help, not status
- Confirm `plonk ls` returns "unknown command"
- Test `plonk st` alias works correctly
- Validate prefix syntax: `plonk install brew:jq`
- Ensure apply continues on partial failures
- Test all output formats work consistently

**Safety Measures:**
- Document that tests modify real system
- Use small, common packages (jq, cowsay)
- Provide cleanup instructions
- Add environment variable to skip dangerous tests

## Testing Requirements

### Integration Tests
1. Full workflow test:
   - Create config with multiple packages
   - Run `plonk apply` with one invalid package
   - Verify partial success behavior
   - Check that valid packages were installed

2. Command interaction tests:
   - Install with prefix, then check with info
   - Apply config, then check status
   - Search, then install from results

3. Error handling tests:
   - Invalid prefix in install
   - Config validation failures
   - Network timeouts in search

### Manual Testing
- Run through typical user workflows
- Try to break things with edge cases
- Verify all error messages are helpful

## Expected Changes

This phase is more about polish than major changes:

1. **Modified files:**
   - Various command files for help text updates
   - Orchestrator for apply failure handling
   - Any files with outdated examples

2. **Documentation updates:**
   - Command help text
   - Example configs
   - Any inline documentation

3. **Bug fixes:**
   - Issues discovered during integration
   - Edge cases in prefix parsing
   - Timeout handling improvements

## Validation Checklist

Before marking complete:
- [ ] Apply continues on partial failures with clear reporting
- [ ] All help text uses new command patterns
- [ ] No references to old commands/flags remain
- [ ] Integration between phases works smoothly
- [ ] Error messages are consistent and helpful
- [ ] Example configs use new patterns
- [ ] BATS test suite implemented and passing
- [ ] All existing tests updated
- [ ] Manual testing confirms good UX

## Notes

- This phase is intentionally flexible to address issues found during implementation
- Focus on polish and consistency rather than new features
- If major issues are found, document them but don't expand scope
- The goal is a cohesive, simplified UX across all commands
- Any output format issues should be noted for Phase 15

Remember to create `PHASE_14_COMPLETION.md` when finished!
