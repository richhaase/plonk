# Task 015: Improve Integration Test Implementation

## Objective
Refactor the integration test implementation to be more Go idiomatic and maintainable while preserving all existing functionality and test coverage.

## Quick Context
- **Current state**: Tests work perfectly but have implementation issues
- **Goal**: Clean up code without breaking any tests
- **Priority**: Code quality improvement, not functionality change
- **Critical**: All existing test coverage must be preserved

## Issues to Fix

### 1. Repeated Manager Switch Logic (High Priority)
**Problem**: Lines 184-197 and 306-318 have identical switch statements
```go
// Current: Duplicated in multiple places
switch availableManager {
case "npm":
    installArgs = append(installArgs, "--npm")
case "pip":
    installArgs = append(installArgs, "--pip")
// ... repeated 6 times
}
```

**Solution**: Extract to helper function
```go
func addManagerFlag(args []string, manager string) []string {
    return append(args, "--"+manager)
}

// Usage: installArgs = addManagerFlag(installArgs, availableManager)
```

### 2. Unchecked JSON Error (High Priority)
**Problem**: Line 108 ignores json.Unmarshal error
```go
// Current: Error ignored
json.Unmarshal([]byte(doctorJSON), &doctorData)

// Fixed: Check error
if err := json.Unmarshal([]byte(doctorJSON), &doctorData); err != nil {
    t.Fatalf("Failed to parse doctor JSON: %v", err)
}
```

### 3. Ignored File Read Errors (High Priority)
**Problem**: Lines 504, 551, 563 ignore os.ReadFile errors
```go
// Current: Error ignored but content used
lockContent, _ := os.ReadFile(filepath.Join(testDir, "plonk.lock"))

// Fixed: Check error
lockContent, err := os.ReadFile(filepath.Join(testDir, "plonk.lock"))
if err != nil {
    t.Fatalf("Failed to read lock file: %v", err)
}
```

### 4. Extract Constants (Medium Priority)
**Problem**: Magic strings scattered throughout
```go
// Add at top of file
const (
    PlonkBinary   = "./plonk"
    LockFileName  = "plonk.lock"
    ConfigFileName = "plonk.yaml"
)

// Manager constants
const (
    ManagerNPM   = "npm"
    ManagerPip   = "pip"
    ManagerCargo = "cargo"
    ManagerGem   = "gem"
    ManagerBrew  = "brew"
    ManagerGo    = "go"
)
```

### 5. Simplify JSON Parsing (Medium Priority)
**Problem**: Lines 111-139 have complex nested type assertions
```go
// Current: Complex nested assertions
if checks, ok := doctorData["checks"].([]interface{}); ok {
    for _, check := range checks {
        if c, ok := check.(map[string]interface{}); ok {
            // ... more nesting
        }
    }
}

// Better: Extract to helper function
func parseManagersFromDoctor(doctorJSON string) ([]string, error) {
    var doctorData map[string]interface{}
    if err := json.Unmarshal([]byte(doctorJSON), &doctorData); err != nil {
        return nil, err
    }

    var availableManagers []string
    // Simplified parsing logic here
    return availableManagers, nil
}
```

### 6. Define TestPackage Type (Low Priority)
**Problem**: Anonymous struct used for test packages
```go
// Current: Anonymous struct
testPackages := map[string]struct {
    install  string
    search   string
    nonexist string
}{...}

// Better: Named type
type TestPackage struct {
    Install  string
    Search   string
    NonExist string
}

var testPackages = map[string]TestPackage{...}
```

### 7. Break Up Long Test Function (Low Priority)
**Problem**: TestCompleteUserExperience is 435 lines
**Solution**: Keep structure but extract complex subtests to separate functions

## Implementation Strategy

### Phase 1: High Priority Fixes (Must Do)
1. **Fix error handling**: Add proper error checks for JSON parsing and file operations
2. **Extract manager flag helper**: Remove duplicated switch statements
3. **Test thoroughly**: Ensure all tests still pass

### Phase 2: Medium Priority Improvements (Should Do)
1. **Add constants**: Replace magic strings with named constants
2. **Simplify JSON parsing**: Extract complex parsing to helper function
3. **Improve helper consistency**: Make helper function APIs more consistent

### Phase 3: Low Priority Polish (Nice to Have)
1. **Add TestPackage type**: Structure test data better
2. **Extract complex subtests**: Break up long functions if beneficial

## Files to Modify
- `tests/integration/ux_complete_test.go` (primary)
- `tests/integration/error_messages_test.go` (minor constants update)

## Success Criteria
1. ✅ All integration tests still pass (`just test-ux`)
2. ✅ No duplicated manager switch logic
3. ✅ All JSON parsing and file operations check errors
4. ✅ Magic strings replaced with constants
5. ✅ Code is more maintainable and readable
6. ✅ No functionality changes or test coverage loss

## Testing Strategy
1. **Before changes**: Run `just test-ux` to establish baseline
2. **After each phase**: Run `just test-ux` to ensure no regressions
3. **Final verification**: Run full test suite to ensure no side effects

## Risk Mitigation
- **Make small, incremental changes**: Don't refactor everything at once
- **Test after each change**: Catch regressions immediately
- **Focus on non-functional changes**: Don't alter test logic, just implementation
- **Preserve all debug output**: Keep existing t.Logf statements for debugging

## Completion Report
Create `TASK_015_COMPLETION_REPORT.md` with:
- List of all changes made
- Before/after code examples for major improvements
- Confirmation that all tests still pass
- Any implementation notes or decisions made
