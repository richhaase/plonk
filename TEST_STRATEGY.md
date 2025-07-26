# Plonk Test Strategy

## Current Issues with Integration Tests

### 1. Build Path Problems
- Tests are failing because they're building from wrong directory
- Need to run from project root or adjust paths

### 2. Command Changes
- Tests still reference old commands (e.g., `sync` instead of `apply`)
- Manager flags will change to prefix syntax
- `ls` command will be removed

### 3. Safety Concerns
- Tests install real packages on the system
- No isolation from user's actual environment
- Can interfere with developer's local setup
- Resource intensive (downloads, installs)

### 4. Platform Dependencies
- Tests assume certain package managers are available
- OS-specific behaviors not properly handled
- Network dependencies for package downloads

## Proposed Solutions

### Option 1: Mock-Based Unit Testing (Recommended)

**Approach:** Create comprehensive unit tests with mocked package managers

**Pros:**
- Fast execution (no real installs)
- Completely safe
- Deterministic results
- Can test edge cases easily
- No network dependencies

**Cons:**
- Doesn't test real package manager integration
- May miss OS-specific issues

**Implementation:**
```go
// Example mock manager
type MockPackageManager struct {
    InstalledPackages map[string]string
    AvailablePackages map[string]string
    ShouldFailOn      map[string]error
}

func (m *MockPackageManager) Install(ctx context.Context, pkg string) error {
    if err, shouldFail := m.ShouldFailOn[pkg]; shouldFail {
        return err
    }
    m.InstalledPackages[pkg] = "1.0.0"
    return nil
}
```

### Option 2: Docker-Based Integration Testing

**Approach:** Run tests in isolated Docker containers

**Pros:**
- Real package manager testing
- Isolated from host system
- Reproducible environment
- Can test multiple OS/distros

**Cons:**
- Requires Docker
- Slower than unit tests
- More complex setup

**Implementation:**
```dockerfile
# test.Dockerfile
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    golang-go \
    npm \
    python3-pip \
    cargo
COPY . /plonk
WORKDIR /plonk
CMD ["go", "test", "./tests/integration", "-tags=docker"]
```

### Option 3: Sandbox Mode with Test Prefixes

**Approach:** Add a test mode that prefixes all operations

**Pros:**
- Tests real commands
- Relatively safe (isolated directories)
- Can run on developer machines

**Cons:**
- Still installs real packages
- May leave artifacts
- Requires cleanup

**Implementation:**
```go
// In test mode, prefix all paths
if os.Getenv("PLONK_TEST_MODE") == "1" {
    config.PackagePrefix = ".plonk-test-"
    config.InstallPath = filepath.Join(tmpDir, "bin")
}
```

### Option 4: Fixture-Based Testing

**Approach:** Use pre-recorded command outputs and behaviors

**Pros:**
- Fast and deterministic
- No real installs needed
- Can test exact scenarios

**Cons:**
- Fixtures can become stale
- Doesn't catch new issues

**Implementation:**
```go
// Load fixture data
type Fixture struct {
    Command  string
    Args     []string
    Output   string
    ExitCode int
}

fixtures := LoadFixtures("testdata/fixtures.json")
```

### Option 5: Hybrid Approach (Recommended Long-term)

**Components:**
1. **Core Unit Tests** (80%)
   - Mock all external dependencies
   - Test business logic thoroughly
   - Fast, safe, run on every commit

2. **CLI Smoke Tests** (15%)
   - Test command parsing and basic flow
   - Use fixtures or mocks
   - Verify UX changes work correctly

3. **Real Integration Tests** (5%)
   - Run in CI only
   - Use Docker containers
   - Test against real package managers
   - Run on releases or major changes

## Immediate Action Plan

### Phase 1: Fix Current Tests (Quick Fix)
```bash
# Create a test runner script
#!/bin/bash
cd tests/integration
go build -o plonk ../../cmd/plonk
go test -tags=integration -v
```

### Phase 2: Create Safe CLI Tests
```go
// tests/cli/command_test.go
func TestCommandStructure(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantErr  bool
        contains []string
    }{
        {
            name:     "help command",
            args:     []string{"--help"},
            contains: []string{"plonk", "apply", "status"},
        },
        {
            name:     "unknown command",
            args:     []string{"unknown"},
            wantErr:  true,
            contains: []string{"unknown command"},
        },
    }
    // Run with mocked managers
}
```

### Phase 3: Mock Package Managers
```go
// internal/resources/packages/mock.go
type MockManager struct {
    name string
    packages map[string]PackageInfo
}

// Use in tests
func TestApplyCommand(t *testing.T) {
    mockBrew := &MockManager{
        name: "homebrew",
        packages: map[string]PackageInfo{
            "ripgrep": {Version: "14.0.0", Installed: true},
        },
    }
    // Test with mock
}
```

## Testing Matrix

| Test Type | What | How | When | Safety |
|-----------|------|-----|------|--------|
| Unit Tests | Business logic | Mocks | Every commit | ✅ Safe |
| CLI Tests | Command parsing | Fixtures | Every commit | ✅ Safe |
| Mock Integration | Full workflows | Mock managers | PR/Push | ✅ Safe |
| Docker Tests | Real managers | Containers | Release | ✅ Safe |
| Manual Tests | UX validation | Human testers | Major changes | ⚠️ Careful |

## Command-Specific Test Strategies

### For Phase 11-15 UX Changes:

1. **Command Consolidation (Phase 11)**
   - Test help output doesn't include `ls`
   - Test `st` alias works
   - Test no-args shows help

2. **Prefix Syntax (Phase 12)**
   - Test parsing "brew:package"
   - Test invalid prefixes
   - Test default manager fallback

3. **Search/Info (Phase 13)**
   - Mock parallel search results
   - Test timeout behavior
   - Test priority ordering

4. **Apply Continuation (Phase 14)**
   - Test partial failure scenarios
   - Mock some packages failing

5. **Output Formatting (Phase 15)**
   - Snapshot testing for outputs
   - Test all format flags

## Recommended Immediate Steps

1. **Create `tests/cli` directory** for safe command tests
2. **Implement mock managers** in `internal/resources/packages/testing`
3. **Convert dangerous tests** to use mocks
4. **Add GitHub Action** for Docker-based integration tests
5. **Document test patterns** for contributors

## Example Safe Test

```go
// tests/cli/apply_test.go
func TestApplyCommand(t *testing.T) {
    // Setup
    dir := t.TempDir()
    os.Setenv("PLONK_DIR", dir)

    // Create config
    config := `
version: 1
packages:
  homebrew:
    - ripgrep
`
    os.WriteFile(filepath.Join(dir, "plonk.yaml"), []byte(config), 0644)

    // Run with mocked manager
    output := runPlonkWithMocks(t, "apply", "--dry-run")

    // Verify
    assert.Contains(t, output, "Would install")
    assert.Contains(t, output, "ripgrep")
}
```

## Conclusion

The current integration tests are:
1. **Unsafe** - Installing real packages
2. **Brittle** - Dependent on network/OS
3. **Broken** - Need updates for UX changes

Recommendation: Implement Option 5 (Hybrid) starting with safe mocked tests for immediate needs, then add Docker-based tests for CI.
