# Phase 12: Package Manager Selection

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

This phase replaces the current flag-based package manager selection (--brew, --npm, etc.) with a more intuitive prefix syntax (brew:package, npm:package). This makes commands shorter and more readable.

## Objectives

1. Implement prefix syntax parsing for package names
2. Remove all --<manager> flags from install/uninstall commands
3. Use default_manager from config when no prefix is provided
4. Update help text and examples to show new syntax
5. Ensure clear error messages for invalid prefixes

## Current State

- Package managers are selected using flags: `plonk install --brew ripgrep`
- Multiple flags exist: --brew, --npm, --pip, --cargo, --go, --gem
- Default manager is used when no flag is specified
- Search and info commands also use these flags (to be addressed in Phase 13)

## Implementation Details

### 1. Implement Prefix Parsing

Create a helper function to parse the prefix syntax:

```go
// ParsePackageSpec splits "manager:package" into (manager, package)
// Returns ("", package) if no prefix is found
func ParsePackageSpec(spec string) (manager, packageName string) {
    parts := strings.SplitN(spec, ":", 2)
    if len(parts) == 2 {
        return parts[0], parts[1]
    }
    return "", spec
}
```

### 2. Update Install Command

**In `internal/commands/install.go`:**
- Remove all manager flag definitions (--brew, --npm, etc.)
- Update the Run function to parse prefix from package names
- Validate manager names against known managers
- Use default_manager when no prefix

Example flow:
```go
// For each package argument:
manager, pkgName := ParsePackageSpec(arg)
if manager == "" {
    manager = cfg.DefaultManager
}
if !IsValidManager(manager) {
    return fmt.Errorf("unknown package manager %q. Valid managers: %s",
        manager, strings.Join(GetValidManagers(), ", "))
}
```

### 3. Update Uninstall Command

**In `internal/commands/uninstall.go`:**
- Apply the same changes as install command
- Remove manager flags
- Parse prefix syntax
- Validate and use default_manager

### 4. Package Resolution Logic

**Important:** Package names can contain colons (e.g., some npm packages)
- Only split on the FIRST colon
- Everything after the first colon is the package name
- Example: `npm:@types/node:test` → manager: "npm", package: "@types/node:test"

### 5. Update Help Text

Update command help to show new syntax:
```
Examples:
  plonk install ripgrep              # Uses default manager
  plonk install brew:wget npm:typescript
  plonk install cargo:ripgrep pip:black

  plonk uninstall brew:wget
  plonk uninstall typescript         # Uses default manager
```

### 6. Error Messages

Provide clear errors for common mistakes:
- Unknown manager: "unknown package manager 'hombrew'. Did you mean 'homebrew'?"
- Empty prefix: "invalid package specification ':package'"
- No package name: "invalid package specification 'brew:'"

## Testing Requirements

### Unit Tests
- Test ParsePackageSpec with various inputs
- Test validation of manager names
- Test error messages for invalid specs

### Integration Tests
1. Install with prefix: `plonk install brew:jq`
2. Install without prefix uses default_manager
3. Install with invalid manager shows helpful error
4. Multiple packages with mixed prefixes: `plonk install brew:jq npm:typescript`
5. Uninstall with same prefix patterns
6. Old flag syntax (--brew) returns unknown flag error

### Manual Testing
- Try various prefix combinations
- Verify default_manager behavior
- Test error cases
- Ensure packages with colons in names work

## Expected Changes

1. **Modified files:**
   - `internal/commands/install.go` - Remove flags, add prefix parsing
   - `internal/commands/uninstall.go` - Same changes
   - Possibly create `internal/commands/utils.go` for shared parsing function

2. **Test updates:**
   - Update integration tests that use --manager flags
   - Add new tests for prefix syntax
   - Remove tests for flag-based selection

3. **Behavior changes:**
   - `plonk install --brew ripgrep` → error: unknown flag
   - `plonk install brew:ripgrep` → installs with brew
   - `plonk install ripgrep` → uses default_manager

## Validation Checklist

Before marking complete:
- [ ] Prefix syntax works for all package managers
- [ ] Default manager used when no prefix
- [ ] All --<manager> flags removed from install/uninstall
- [ ] Clear error messages for invalid prefixes
- [ ] Package names with colons handled correctly
- [ ] Help text updated with new examples
- [ ] All tests updated and passing
- [ ] No references to old flag syntax in code or help

## Notes

- This change affects install/uninstall only in this phase
- Search and info commands will be updated in Phase 13
- This is a breaking change - old flag syntax will no longer work
- The prefix syntax is more intuitive and saves typing
- Aligns with common tools like docker (ubuntu:latest) and go modules

Remember to create `PHASE_12_COMPLETION.md` when finished!
