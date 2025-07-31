# Safe Testing Strategy for System-Modifying Tools

## The Core Problem

Plonk is inherently dangerous to test because it:
- Installs real packages on the system (brew, npm, pip, etc.)
- Modifies user dotfiles in $HOME
- Changes system state that affects the developer's environment
- Can break the developer's machine if tests go wrong

Current tests use "safe" packages like `cowsay` and `figlet`, but even these:
- Take time to download and install
- Consume disk space
- Modify package manager databases
- Can conflict with user's actual needs

## Current Unsafe Practices

### BATS Tests
```bash
@test "install single brew package" {
  run plonk install brew:cowsay  # INSTALLS ON REAL SYSTEM!
  # ...
}
```

### Integration Tests
```go
// Tests check CI=true to avoid running on developer machines
if os.Getenv("CI") != "true" {
    t.Skip("Integration tests only run in CI")
}
```

This is inadequate because:
- Developers can't run tests locally
- CI environments are different from user environments
- Still modifies the CI system

## Safe Testing Approaches

### 1. Mock Mode for Package Managers
Add a test/mock mode that simulates operations without executing them:

```bash
PLONK_MOCK_MODE=true plonk install brew:vim
# Would simulate the install without actually running brew
```

Implementation:
- Add mock implementations of each package manager
- Return realistic output without system changes
- Verify correct commands would be executed

### 2. Containerized Testing
Run all tests in disposable containers:

```yaml
# docker-compose.test.yml
services:
  test-macos:
    image: sickcodes/docker-osx
    volumes:
      - .:/plonk
    command: make test

  test-ubuntu:
    image: ubuntu:22.04
    volumes:
      - .:/plonk
    command: make test
```

Benefits:
- Complete isolation
- Test on multiple OSes
- No impact on developer systems
- Reproducible environments

### 3. Fixture-Based Testing
Instead of real operations, use fixtures:

```go
func TestInstallPackage(t *testing.T) {
    // Override exec.Command to capture calls
    execCommand = mockExecCommand
    defer func() { execCommand = exec.Command }()

    // Test runs but captures commands instead of executing
    output := runPlonk("install", "brew:vim")

    // Verify the right commands would be called
    assert.Contains(t, capturedCommands, "brew install vim")
}
```

### 4. Separate Test Commands
Add explicit test-only commands that are safe:

```bash
plonk test-install brew:fake-package
# Uses a mock registry, doesn't touch real brew
```

### 5. VM-Based Testing (Current Lima Approach)
Enhance the Lima approach with automation:

```bash
# Create ephemeral VM for each test run
lima create --name=plonk-test-$$ template.yaml
lima shell plonk-test-$$ make test
lima delete plonk-test-$$
```

## Recommended Solution: Layered Approach

### Layer 1: Unit Tests (Local Safe)
- Test pure functions only
- No system calls
- Run on every commit

### Layer 2: Mock Integration Tests (Local Safe)
- Use dependency injection for package managers
- Mock all system interactions
- Verify correct commands would be executed
- Run on every commit

### Layer 3: Container Tests (Local Optional)
- Full integration tests in Docker
- Real package installations in isolated environment
- Developer can opt-in with `make test-docker`

### Layer 4: VM Tests (CI Only)
- Full system tests in Lima/GitHub Actions
- Test on real macOS and Linux
- Only run on CI to avoid developer impact

## Implementation Priority

### Phase 1: Add Mock Mode (Critical for v1.0)
```go
type PackageManager interface {
    Install(ctx context.Context, name string) error
}

type MockPackageManager struct {
    RecordedCalls []string
}

func (m *MockPackageManager) Install(ctx context.Context, name string) error {
    m.RecordedCalls = append(m.RecordedCalls, "install "+name)
    return nil
}
```

### Phase 2: Dockerize Tests
- Create Dockerfiles for each supported OS
- Add docker-compose.test.yml
- Update Makefile with docker-test target

### Phase 3: Improve CI
- Use matrix builds for multiple OS versions
- Cache package manager state between runs
- Add smoke tests that verify basics without installing

## Example Safe Test Pattern

```go
func TestInstallCommand(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Use test directory, not real home
    tmpHome := t.TempDir()
    t.Setenv("HOME", tmpHome)
    t.Setenv("PLONK_DIR", filepath.Join(tmpHome, ".plonk"))

    // Use mock package manager
    t.Setenv("PLONK_MOCK_MODE", "true")

    // Now safe to run
    output := runPlonk("install", "brew:vim")

    // Verify mock was called correctly
    assert.Contains(t, output, "Would install: vim")
}
```

## Conclusion

The current testing approach is fundamentally unsafe. We need to:

1. **Never modify developer systems** during tests
2. **Provide mock modes** for all system interactions
3. **Use containers/VMs** for real integration tests
4. **Make local testing safe by default**

This is critical for v1.0 to ensure developers can contribute without fear of breaking their systems.
