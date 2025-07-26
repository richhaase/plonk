# Worker Context for UX Implementation

**IMPORTANT: Read this entire document before starting any work.**

## Overview

This document provides ground rules and guidance for workers implementing the UX improvements in Phases 9-14. The goal is to simplify plonk's user experience while maintaining all functionality and preparing for future AI Lab features.

## Core Principles

1. **Preserve Functionality**: All current features must continue to work
2. **Backward Compatibility**: Not required - we can make breaking changes
3. **Simplicity Over Features**: When in doubt, choose the simpler approach
4. **Consistency**: Similar operations should work similarly across commands
5. **Clear Errors**: Error messages should guide users to the solution

## Implementation Guidelines

### 1. Command Changes

When modifying commands:
- Remove old commands/flags completely (no deprecation needed)
- Update all help text to reflect new patterns
- Ensure command aliases are properly registered
- Test all output formats (-o table|json|yaml)

### 2. Code Organization

- Keep commands thin - business logic belongs in domain packages
- Avoid creating new abstractions unless absolutely necessary
- Prefer explicit code over clever abstractions
- Follow existing patterns in the codebase

### 3. Error Handling

- Use clear, actionable error messages
- Include what went wrong AND how to fix it
- For validation errors, show the invalid value and expected format
- Don't wrap errors multiple times

Example:
```go
// Bad
return fmt.Errorf("invalid input")

// Good
return fmt.Errorf("invalid package name %q: must not contain colons", input)
```

### 4. Testing

- Update existing tests rather than writing new ones where possible
- Focus on integration tests for command changes
- Don't test internal implementation details
- Ensure all tests pass before marking task complete

### 5. Output Formatting

- Maintain consistent table formatting across commands
- JSON output should be properly structured (not just dumped structs)
- YAML output should be human-readable
- Error output goes to stderr, data output to stdout

### 6. Prefix Syntax Implementation

For the new `manager:package` syntax:
- Parse prefix before the first colon only
- Validate manager name against known managers
- Provide helpful error if manager is unknown
- Fall back to default_manager if no prefix

Example:
```go
// "brew:ripgrep" -> manager: "brew", package: "ripgrep"
// "ripgrep" -> manager: default_manager, package: "ripgrep"
// "brew:some:package:with:colons" -> manager: "brew", package: "some:package:with:colons"
```

### 7. Search/Info Parallelization

When implementing parallel operations:
- Use goroutines with proper error handling
- Implement timeout (2-3 seconds max)
- Collect all results before displaying
- Handle partial failures gracefully

### 8. Documentation Updates

- Update command help text inline with changes
- Don't create separate documentation yet (that's Final Phase)
- Keep examples realistic and useful
- Remove references to old commands/patterns

## What NOT to Do

1. **Don't add new features** - Only implement what's specified
2. **Don't over-engineer** - Keep solutions simple
3. **Don't preserve old patterns** - Remove them completely
4. **Don't add configuration options** - Stick to the plan
5. **Don't refactor unrelated code** - Stay focused on the task

## Key Files to Modify

Based on the UX decisions, you'll primarily work with:
- `internal/commands/*.go` - Command definitions
- `internal/orchestrator/*.go` - For sync→apply rename
- `internal/config/*.go` - For validation changes
- `internal/resources/packages/*.go` - For manager selection

## Testing Your Changes

Before marking any task complete:
1. Run `go test ./...` - All unit tests must pass
2. Run `just test-ux` - Integration tests must pass
3. Test each command manually with -o table|json|yaml
4. Verify error messages are helpful

## Required Workflow

Every worker MUST follow these steps in order:

### 1. Validate Starting State
Run `just test-ux` to ensure you're starting from a clean state. All tests must pass before beginning work.

### 2. Identify Affected Tests
- Find which integration tests cover the components you'll modify
- Read these tests to understand current behavior
- Make note of tests that will need updates

### 3. Implement UX Changes
- Read existing code thoroughly before making changes
- Follow the guidelines in this document
- Make changes incrementally, testing as you go

### 4. Update Unit Tests
- Add/update unit tests for business logic changes
- Do NOT test system commands in unit tests
- Focus on testing the logic, not the CLI interaction

### 5. Update Integration Tests
- **IMPORTANT**: When tests involve package installation or file creation, you MUST:
  - List the packages/files you plan to use in tests
  - Ask the user: "I plan to use these packages/files in tests: [list]. Is this acceptable?"
  - Wait for approval before implementing
- Update existing integration tests to match new behavior
- Add new integration tests only if necessary

### 6. Verify All Tests Pass
- Run `go test ./...` for unit tests
- Run `just test-ux` for integration tests
- Fix any failures before proceeding

### 7. Write Completion Report
Create `PHASE_X_COMPLETION.md` (where X is your phase number) with:
- Summary of changes made
- List of files modified
- Test results (confirming all pass)
- Any decisions made during implementation
- Confirmation that all phase goals were met

## Questions?

If you encounter situations not covered here:
1. Check existing code for patterns
2. Choose the simpler solution
3. Document your decision in the commit message
4. Ask for clarification if truly blocked

Remember: The goal is a simpler, more intuitive CLI while maintaining all current functionality.
