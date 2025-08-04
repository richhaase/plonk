# Plonk Development Rules for AI Agents

This document contains critical rules that MUST be followed when working on the Plonk codebase. These rules exist to protect user systems and maintain code quality. Violating these rules can cause serious harm to developer machines.

## 1. Scope Control Rules

### NEVER Add Unrequested Features
- **FORBIDDEN**: Implementing features, enhancements, or improvements that were not explicitly requested
- **ALLOWED**: Proposing improvements through comments or suggestions
- **REQUIRED**: When given a task, implement EXACTLY what was asked - nothing more, nothing less
- **EXAMPLE**: If asked to "fix the JSON output bug", do NOT also "improve error messages" or "add helpful logging"

### File Creation Restrictions
- **FORBIDDEN**: Creating new files unless absolutely necessary for the requested task
- **FORBIDDEN**: Creating documentation files (*.md) or README files unless explicitly requested
- **REQUIRED**: Always prefer editing existing files over creating new ones
- **EXAMPLE**: If fixing a bug, modify the existing file rather than creating a new helper file

## 2. User Interface Rules

### No Emojis in Output
- **FORBIDDEN**: Using emojis (🎉, ✅, ❌, etc.) in any plonk command output
- **REQUIRED**: Use colored text for status indicators instead
- **REQUIRED**: Color only the status word itself, not entire lines
- **EXAMPLE**: Use `[green]installed[/green]` not `✅ Package installed successfully! 🎉`

### Professional Output Standards
- **REQUIRED**: Output must be clean and professional like git, docker, or kubectl
- **FORBIDDEN**: Chatty, conversational, or "friendly" output messages
- **EXAMPLE**: Use `Installing package...` not `Let's install this package for you!`

## 3. Testing Safety Rules

### The Golden Rule of Testing
**UNIT TESTS MUST NEVER MODIFY THE HOST SYSTEM**

This is the most critical rule in the entire codebase. Tests that modify system state put developer machines at risk.

### Forbidden Test Operations
Tests MUST NEVER:
- Call `Apply()` methods that could install real packages
- Execute real package manager commands (`brew install`, `apt-get`, `npm install`, etc.)
- Run hooks or shell commands that affect the system
- Write to any paths outside of temporary test directories created with `os.MkdirTemp()`
- Create or modify dotfiles in the user's home directory
- Modify ANY aspect of the developer's machine

### Safe Testing Practices
- **ALLOWED**: Testing pure functions that only manipulate data
- **ALLOWED**: Testing business logic that doesn't touch the filesystem
- **ALLOWED**: Using `os.MkdirTemp()` to create temporary directories for test files
- **ALLOWED**: Using the `CommandExecutor` interface to mock system commands
- **REQUIRED**: Integration tests that need real system interaction MUST run in Docker containers

### The Safety Check Question
Before adding ANY test, ask yourself: "Could this test modify the real system?"
- If YES → DO NOT add the test
- If UNSURE → DO NOT add the test
- If NO → Proceed with caution

### Integration Testing Rules
- **REQUIRED**: Integration tests MUST run in Docker containers using testcontainers-go
- **FORBIDDEN**: Running integration tests directly on developer machines
- **REQUIRED**: Integration tests must have no side effects outside their container

## 4. Code Architecture Rules

### Commands Package Testing
- **FACT**: Commands package orchestration functions are intentionally not unit testable
- **FORBIDDEN**: Attempting to unit test CLI command orchestration
- **ALLOWED**: Testing extracted business logic from command handlers
- **REQUIRED**: Accept that some code paths will have low coverage for safety

### Test Coverage Philosophy
- **PRINCIPLE**: Safety > Coverage
- **FORBIDDEN**: Adding unsafe tests to increase coverage metrics
- **ALLOWED**: Having lower coverage if it means keeping tests safe
- **REMINDER**: "No tests is better than tests that break developer machines"

## 5. Examples of Rule Violations

### ❌ WRONG: Adding Unrequested Features
```go
// Task: "Fix the install command error handling"
// WRONG: Also added progress bar, emoji output, and new --verbose flag
func installCommand() {
    showProgressBar() // <- NOT REQUESTED
    fmt.Println("🚀 Starting installation!") // <- EMOJIS FORBIDDEN
    if verbose { // <- NEW FLAG NOT REQUESTED
        // ...
    }
}
```

### ❌ WRONG: Unsafe Test
```go
// WRONG: This test will install real packages on the developer's machine!
func TestInstallCommand(t *testing.T) {
    cmd := exec.Command("brew", "install", "wget") // <- MODIFIES REAL SYSTEM
    cmd.Run() // <- DANGER: ACTUALLY INSTALLS PACKAGE
}
```

### ✅ CORRECT: Safe Test
```go
// CORRECT: Uses mock executor, doesn't touch real system
func TestInstallLogic(t *testing.T) {
    executor := &MockCommandExecutor{} // <- MOCK, NOT REAL
    result := processInstallRequest("wget", executor)
    assert.Equal(t, "would install wget", result)
}
```

## Remember

These rules exist because:
1. User trust is paramount - we must never harm their systems
2. Scope creep makes code harder to review and can introduce bugs
3. Professional tools have professional output
4. Safety is more important than metrics

When in doubt, err on the side of caution. It's better to do less safely than more dangerously.

## 6. CLAUDE.md Usage Rules

### This File is for Development Rules Only
- **REQUIRED**: CLAUDE.md must contain ONLY development rules and guidelines
- **FORBIDDEN**: Using CLAUDE.md to store project status, todo lists, future plans, or any other context
- **FORBIDDEN**: Adding sections about current progress, version information, or feature tracking
- **REQUIRED**: Store project status, plans, and tracking information in appropriate files (e.g., docs/planning/*.md, TODO.md, etc.)
- **EXAMPLE**: Development rules belong here, but "Current sprint goals" belong in a separate planning document

This restriction ensures CLAUDE.md remains a clear, focused reference for development rules without becoming cluttered with transient information.
