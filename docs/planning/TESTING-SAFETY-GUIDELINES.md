# CRITICAL TESTING SAFETY GUIDELINES

## 🚨 ABSOLUTE RULE: NEVER MODIFY SYSTEM STATE IN UNIT TESTS 🚨

**This document exists because this rule has been violated multiple times by both human developers and AI agents.**

## The Golden Rule

**NO UNIT TEST MAY EVER:**
- Install or uninstall packages
- Modify dotfiles
- Execute shell commands that affect the system
- Write to any path outside of temporary test directories
- Change ANY aspect of the developer's machine

**VIOLATION OF THIS RULE PUTS DEVELOPERS AT RISK OF HAVING THEIR MACHINES BROKEN**

## Why This Rule Exists

We have had multiple incidents where:
1. Tests attempted to run `brew install` commands
2. Tests tried to modify real dotfiles in home directories
3. Tests executed hooks that changed system configuration
4. AI agents created tests that called Apply() methods without mocking

Each incident risked breaking developer machines. This is unacceptable.

## Safe Testing Practices

### ✅ ALWAYS SAFE
- Testing pure functions (no side effects)
- Testing data structures and types
- Using `os.MkdirTemp()` for all file operations
- Testing logic without execution
- Mocking all external interactions

### ❌ NEVER SAFE
- Calling `orchestrator.Apply()` without complete mocking
- Running package manager commands (brew, apt, npm, etc.)
- Modifying files outside temp directories
- Executing hooks or shell commands
- Any operation that changes system state

## Examples

### ❌ DANGEROUS - NEVER DO THIS
```go
func TestApply(t *testing.T) {
    o := &Orchestrator{
        config: testConfig,
        dryRun: false, // DANGER: Real mode!
    }
    // THIS COULD INSTALL PACKAGES OR MODIFY FILES
    result, err := o.Apply(ctx)
}
```

### ✅ SAFE - Testing Logic Only
```go
func TestApplyLogic(t *testing.T) {
    // Test the logic of selective application
    o := &Orchestrator{
        packagesOnly: true,
        dotfilesOnly: false,
    }
    // Just test the flags, don't call Apply()
    shouldRunPackages := !o.dotfilesOnly
    assert.True(t, shouldRunPackages)
}
```

### ❌ DANGEROUS - Real Paths
```go
func TestSomething(t *testing.T) {
    // NEVER use real paths
    configDir := "/home/user/.config/plonk"
    homeDir := "/home/user"
}
```

### ✅ SAFE - Temp Directories
```go
func TestSomething(t *testing.T) {
    // ALWAYS use temp directories
    configDir, _ := os.MkdirTemp("", "plonk-test-*")
    defer os.RemoveAll(configDir)

    homeDir, _ := os.MkdirTemp("", "plonk-home-*")
    defer os.RemoveAll(homeDir)
}
```

## For AI Agents

**SPECIAL NOTICE TO CLAUDE AND OTHER AI ASSISTANTS:**

You have a tendency to create "helpful" tests that are actually dangerous. Before writing ANY test, you MUST ask:

1. Could this test modify the real file system?
2. Could this test install or uninstall software?
3. Could this test execute system commands?
4. Could this test change system configuration?

If the answer to ANY of these is "yes" or "maybe", DO NOT WRITE THE TEST.

Remember: **NO TESTS IS BETTER THAN DANGEROUS TESTS**

## Enforcement

1. Code reviews MUST check for dangerous tests
2. CI MUST run tests in isolated environments
3. Any PR with dangerous tests MUST be rejected
4. This document MUST be linked in CONTRIBUTING.md

## If You're Unsure

If you're not 100% certain a test is safe:
1. Don't write it
2. Ask for review
3. Write a different test
4. Test at a different level (integration tests in CI)

**When in doubt, leave it out.**

## Final Warning

Creating unit tests that modify system state is:
- Dangerous to developers
- A violation of trust
- Grounds for rejecting contributions
- Never acceptable under any circumstances

**This is not a guideline. This is a requirement.**
