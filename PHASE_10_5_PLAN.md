# Phase 10.5: Fix Apply Command Hanging Issue

**IMPORTANT: Read WORKER_CONTEXT.md before starting any work.**

## Overview

During Phase 10 testing, it was discovered that `plonk apply --dry-run` hangs and does not complete execution. This critical issue must be resolved before proceeding with further UX improvements. This phase focuses on debugging and fixing the hanging behavior.

## Objectives

1. Identify the root cause of the hanging behavior
2. Fix the issue without breaking existing functionality
3. Ensure all apply command variations work correctly
4. Add tests to prevent regression

## Symptoms

- `plonk apply --dry-run` hangs indefinitely
- No error message or timeout occurs
- Process must be manually terminated
- Other commands appear to work normally

## Investigation Strategy

### 1. Reproduce the Issue

First, confirm the hanging behavior:
```bash
# Try different variations to isolate the issue
plonk apply --dry-run
plonk apply --dry-run --packages
plonk apply --dry-run --dotfiles
plonk apply  # without dry-run
```

### 2. Add Debug Logging

Add temporary debug statements to trace execution:
- Entry/exit of main apply function
- Config loading steps
- Hook execution points
- Package/dotfile processing loops
- Context creation and cancellation

Example:
```go
fmt.Fprintf(os.Stderr, "DEBUG: Entering Apply function\n")
// ... code ...
fmt.Fprintf(os.Stderr, "DEBUG: Config loaded successfully\n")
```

### 3. Check Common Causes

**Context Issues:**
- Missing context initialization
- Context not being passed to goroutines
- Missing context cancellation

**Infinite Loops:**
- Circular dependencies in renamed code
- Recursive function calls
- Loop conditions that never terminate

**Blocking Operations:**
- Synchronous operations without timeouts
- Channel operations without buffering
- Mutex deadlocks

**Hook Execution:**
- Pre/post apply hooks running without timeout
- Hook commands hanging
- Hook error handling issues

### 4. Use Debugging Tools

```bash
# Run with race detector
go run -race cmd/plonk/main.go apply --dry-run

# Generate goroutine dump when hanging
# In another terminal while plonk is hanging:
kill -QUIT <pid>  # Dumps goroutines to stderr

# Use delve debugger
dlv debug cmd/plonk/main.go -- apply --dry-run
```

## Likely Problem Areas

Based on the rename from sync to apply, check these specific areas:

### 1. Command Registration

Verify the command is properly registered:
```go
// In internal/commands/apply.go
func init() {
    rootCmd.AddCommand(applyCmd)  // Ensure this exists
}
```

### 2. Dry Run Logic

Check if dry-run flag handling was affected:
```go
// Look for conditions like:
if !dryRun {
    // Actual apply logic
}
// Ensure there's a code path for dry-run
```

### 3. Output Handling

The apply command likely uses channels or goroutines for output:
```go
// Check for unbuffered channels
ch := make(chan Result)  // Could block if no reader

// Or missing channel closes
defer close(ch)
```

### 4. Resource Manager Calls

The renamed functions might have initialization issues:
```go
// Check if managers are properly initialized
if manager == nil {
    return errors.New("manager not initialized")
}
```

## Fix Implementation

Once the root cause is identified:

1. **Implement the minimal fix** - Don't refactor unrelated code
2. **Add comments** explaining what was fixed and why
3. **Test thoroughly** - All apply command variations
4. **Add regression test** - Ensure this specific issue doesn't recur

## Testing Requirements

### Manual Testing
1. `plonk apply --dry-run` completes successfully
2. `plonk apply` works normally (without dry-run)
3. All flag combinations work:
   - `--packages` only
   - `--dotfiles` only
   - `--force`
   - Multiple flags combined
4. Hook execution works (if configured)
5. Error cases handled properly

### Automated Testing
- Add integration test specifically for dry-run
- Test with timeout to catch hanging
- Verify output is correct in dry-run mode

Example test:
```go
func TestApplyDryRunCompletes(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "plonk", "apply", "--dry-run")
    err := cmd.Run()

    if ctx.Err() == context.DeadlineExceeded {
        t.Fatal("apply --dry-run command timed out")
    }
}
```

## Expected Resolution

The fix will likely be one of:
1. **Missing initialization** - Add proper setup code
2. **Blocking channel** - Add buffering or proper goroutine
3. **Missing code path** - Add logic for dry-run case
4. **Context issue** - Properly propagate context
5. **Renamed reference** - Fix a missed rename

## Validation Checklist

Before marking complete:
- [ ] Root cause identified and documented
- [ ] Fix implemented with minimal changes
- [ ] All apply command variations tested manually
- [ ] Regression test added
- [ ] No new issues introduced
- [ ] Code comments explain the fix
- [ ] All existing tests still pass

## Notes

- This is a blocking issue that must be resolved before Phase 11
- Keep the fix focused - don't refactor unrelated code
- Document the root cause in the completion report
- If the issue is complex, break it into smaller debugging steps
- Consider that the issue might be environment-specific

Remember to create `PHASE_10_5_COMPLETION.md` when finished!
