# Task 014: Remove BaseManager Inheritance Pattern

## Objective
Eliminate the Java-style BaseManager inheritance pattern in the managers package, replacing it with idiomatic Go composition and helper functions to achieve ~30% code reduction.

## Quick Context
- **Current state**: 4,402 LOC with BaseManager embedded in all managers
- **Target**: ~3,000 LOC with standalone manager implementations
- **Key principle**: Prefer composition over inheritance (Go way)
- **Preserve**: All functionality, just restructure implementation

## Current Anti-Pattern Analysis

### BaseManager Embedding
```go
// Current: Java-style inheritance
type NpmManager struct {
    *BaseManager  // Embedding used as inheritance
}

// Methods like Install() call b.executeCommand() on BaseManager
```

### Problems with Current Approach
1. **Unnecessary abstraction**: BaseManager adds complexity without benefit
2. **Hidden dependencies**: Hard to see what each manager actually needs
3. **Testing complexity**: Requires mocks and complex setup
4. **Non-idiomatic**: Go developers expect composition, not inheritance

## Work Required

### Phase 1: Analyze BaseManager Usage
1. **Identify common functionality** in base.go:
   - Command execution helpers
   - Error matching/parsing
   - Availability checking
   - Output parsing helpers

2. **Categorize by actual need**:
   - What's truly shared vs. what's forced to be shared
   - What could be simple helper functions
   - What belongs in each manager

### Phase 2: Create Helper Functions
1. **Create managers/helpers.go** with common functions:
   ```go
   // Instead of BaseManager methods, use simple functions
   func ExecuteCommand(name string, args ...string) ([]byte, error)
   func ParsePackageList(output []byte, parser func([]byte) ([]string, error)) []string
   func CheckCommandAvailable(name string) error
   ```

2. **Move error matching** to dedicated file:
   - Keep error_matcher.go as is (already well-structured)
   - Managers can use it directly

### Phase 3: Refactor Each Manager
For each manager (homebrew, npm, cargo, pip, gem, goinstall):

1. **Remove BaseManager embedding**
2. **Add only needed fields directly**:
   ```go
   type NpmManager struct {
       name string  // If needed at all
   }
   ```

3. **Implement interface methods directly**:
   - Call helper functions where needed
   - Inline simple operations
   - Remove unnecessary abstraction layers

4. **Simplify command execution**:
   ```go
   // Before: m.executeCommand("npm", "install", pkg)
   // After: ExecuteCommand("npm", "install", pkg)
   ```

### Phase 4: Update Tests
1. **Remove BaseManager test file**
2. **Simplify manager tests**:
   - Remove mock setups
   - Test actual command parsing logic
   - Use test helpers for command execution

### Phase 5: Clean Up
1. **Delete base.go** completely
2. **Update registry.go** to work with simplified managers
3. **Ensure all managers still implement the interface**
4. **Run all tests** to verify functionality

## Implementation Strategy

### Example: NPM Manager Transformation

**Before (with BaseManager)**:
```go
type NpmManager struct {
    *BaseManager
}

func NewNpmManager() *NpmManager {
    return &NpmManager{
        BaseManager: NewBaseManager("npm", "npm"),
    }
}

func (m *NpmManager) Install(packageName string) error {
    output, err := m.executeCommand(m.managerName, "install", "-g", packageName)
    if err != nil {
        return errors.NewPackageInstallError(packageName, m.name, err.Error())
    }
    return m.parseInstallResult(output, packageName)
}
```

**After (direct implementation)**:
```go
type NpmManager struct{}

func NewNpmManager() *NpmManager {
    return &NpmManager{}
}

func (m *NpmManager) Name() string {
    return "npm"
}

func (m *NpmManager) Install(packageName string) error {
    output, err := ExecuteCommand("npm", "install", "-g", packageName)
    if err != nil {
        return fmt.Errorf("npm install %s: %w", packageName, err)
    }

    // Direct parsing logic here or helper function
    if !strings.Contains(string(output), "added") {
        return fmt.Errorf("npm install %s: no packages added", packageName)
    }
    return nil
}
```

## Files to Update
- **Delete**: base.go, base_test.go
- **Create**: helpers.go
- **Modify**: homebrew.go, npm.go, cargo.go, pip.go, gem.go, goinstall.go
- **Update**: All corresponding test files
- **Adjust**: registry.go (minor updates)

## Success Criteria
1. ✅ BaseManager completely eliminated
2. ✅ All managers work independently
3. ✅ ~30% code reduction achieved (4,402 → ~3,000 LOC)
4. ✅ All tests pass
5. ✅ No functionality lost
6. ✅ More idiomatic Go code
7. ✅ Simpler command execution pattern

## Risk Mitigation
- **Test continuously**: Run tests after each manager refactor
- **One manager at a time**: Don't try to do all at once
- **Preserve behavior**: The CLI interface must not change
- **Keep error messages**: Users rely on specific error formats

## Completion Report
Create `TASK_014_COMPLETION_REPORT.md` with:
- LOC reduction metrics per file
- List of all deleted/created/modified files
- Verification that all managers still work
- Any behavioral differences noted
- Performance impact (should be minimal or positive)
