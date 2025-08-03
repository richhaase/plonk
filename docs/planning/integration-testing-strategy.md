# Plonk Integration Testing Strategy

**Created**: 2025-08-03
**Updated**: 2025-08-03
**Status**: APPROVED
**Target**: v1.1+ (post-v1.0 release)

## Executive Summary

This document outlines a comprehensive integration testing strategy for Plonk that complements our existing unit tests (45.1% coverage) by testing the actual system interactions we explicitly avoided in unit tests. The strategy uses Docker containers to enable safe local testing while using real packages to ensure realistic behavior.

## Key Decisions

1. **Docker Required**: Yes - provides complete isolation for safe local testing
2. **Test Packages**: Real packages (jq, ripgrep, etc.) with network error handling
3. **Local Testing**: Enabled - Docker makes it safe for developers to run locally
4. **Scope**: Start with basic coverage, defer advanced scenarios
5. **Platform Coverage**: Test both macOS and Linux environments

## Research Findings: Modern Go Integration Testing (2025)

Based on current best practices, the following approaches are recommended:

### 1. **Testcontainers for Go**
- Dynamic port configuration to avoid conflicts
- Automatic cleanup after test completion
- Cross-platform compatibility (Docker Desktop, OrbStack, Podman)
- Wait strategies for container readiness
- Lifecycle hooks for post-startup configuration

### 2. **Subprocess Testing with os/exec**
- Standard approach for CLI testing
- Coverage collection via GOCOVERDIR environment variable
- Handling programs that call os.Exit()
- Security improvements in Go 1.19+ (no relative path resolution)

### 3. **Modern Testing Frameworks**
- efficientgo/e2e for complex workload scenarios
- ory/dockertest for database/service dependencies
- testcli for minimal CLI testing wrapper
- testify suite for test organization with setup/teardown

## Current Testing State

### Unit Test Coverage Analysis

From our unit testing efforts, we achieved 45.1% coverage with clear boundaries:

| Package | Coverage | Limitation |
|---------|----------|------------|
| parsers | 100% | Pure business logic ✅ |
| config | 95.4% | Pure business logic ✅ |
| resources | 89.8% | Utility functions only |
| diagnostics | 70.6% | Health checks with temp dirs |
| packages | 62.1% | Limited by actual installations |
| dotfiles | 50.5% | Limited by file operations |
| clone | 28.9% | Limited by git/network ops |
| orchestrator | 17.6% | Limited by system operations |
| commands | 14.6% | CLI orchestration layer |

### Key Finding: The Testing Gap

Our unit tests revealed that **54.9% of the codebase** involves system interactions that cannot be safely unit tested:
- Package installation/removal
- Dotfile symlink management
- Git operations
- File system modifications
- Shell command execution

This is precisely where integration tests excel.

## Integration Testing Philosophy

### Core Principles

1. **Complete Isolation**: All tests run in disposable environments
2. **Real Operations**: Test actual package installations, file operations, git clones
3. **CI-First Design**: Optimized for GitHub Actions, but runnable locally
4. **Explicit Opt-In**: Developers must explicitly run integration tests
5. **Fast Feedback**: Parallel execution where possible

### Testing Boundaries

| Test Type | What It Tests | Where It Runs |
|-----------|--------------|---------------|
| Unit Tests | Business logic, pure functions | Developer machines, CI |
| Integration Tests | System interactions, CLI commands | CI only (opt-in locally) |
| E2E Tests | Full user workflows | Dedicated test environments |

## Proposed Architecture

### 1. Docker-Based Test Environments

```go
// internal/testintegration/environment.go
type TestEnvironment struct {
    container     testcontainers.Container
    tempDir       string
    plonkBinary   string
    homeDir       string
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
    ctx := context.Background()

    // Create container with necessary tools
    req := testcontainers.ContainerRequest{
        Image: "plonk-test:latest", // Custom image with brew, npm, etc.
        Cmd: []string{"tail", "-f", "/dev/null"}, // Keep alive
        Mounts: []string{
            // Mount compiled plonk binary
            fmt.Sprintf("%s:/usr/local/bin/plonk", getBinaryPath()),
        },
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started: true,
    })
    require.NoError(t, err)

    return &TestEnvironment{
        container: container,
        tempDir: "/tmp/plonk-test",
        homeDir: "/home/testuser",
    }
}
```

### 2. Test Categories

#### Package Manager Tests
```go
// tests/integration/packages_test.go
func TestPackageInstallation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    env := NewTestEnvironment(t)
    defer env.Cleanup()

    tests := []struct {
        name    string
        manager string
        package string
    }{
        {"brew formula", "brew", "jq"},
        {"npm global", "npm", "prettier"},
        {"pip package", "pip", "black"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test installation
            output := env.RunPlonk("install", fmt.Sprintf("%s:%s", tt.manager, tt.package))
            assert.Contains(t, output, "Successfully installed")

            // Verify in lock file
            lock := env.ReadLockFile()
            assert.Contains(t, lock, tt.package)

            // Test removal
            output = env.RunPlonk("uninstall", fmt.Sprintf("%s:%s", tt.manager, tt.package))
            assert.Contains(t, output, "Successfully removed")
        })
    }
}
```

#### Dotfile Management Tests
```go
// tests/integration/dotfiles_test.go
func TestDotfileSymlinking(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    // Create test dotfiles
    env.WriteFile(".bashrc", "# Test bashrc")
    env.WriteFile(".gitconfig", "[user]\nname = Test")

    // Add dotfiles
    output := env.RunPlonk("dot", "add", ".bashrc", ".gitconfig")
    assert.Contains(t, output, "Added 2 dotfiles")

    // Apply dotfiles
    output = env.RunPlonk("apply", "--dotfiles-only")
    assert.Contains(t, output, "Created 2 symlinks")

    // Verify symlinks
    assert.True(t, env.IsSymlink(".bashrc"))
    assert.True(t, env.IsSymlink(".gitconfig"))
}
```

#### Clone and Setup Tests
```go
// tests/integration/clone_test.go
func TestCloneAndSetup(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    // Use a minimal test repository
    output := env.RunPlonk("clone", "plonk-test/minimal-setup", "--no-apply")
    assert.Contains(t, output, "Repository cloned successfully")
    assert.Contains(t, output, "Detected required managers: brew, npm")

    // Verify config creation
    assert.True(t, env.FileExists(".config/plonk/plonk.yaml"))
}
```

### 3. GitHub Actions Integration

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on:
  pull_request:
    types: [opened, synchronize, reopened]
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        test-suite: [packages, dotfiles, clone, full]

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.23'

    - name: Build plonk binary
      run: go build -o plonk ./cmd/plonk

    - name: Build test container
      run: docker build -t plonk-test:latest -f tests/integration/Dockerfile .

    - name: Run integration tests
      run: |
        go test -v -tags=integration \
          -run "^Test.*/${{ matrix.test-suite }}" \
          ./tests/integration/...
      env:
        PLONK_TEST_BINARY: ./plonk
```

### 4. Local Development Support

```bash
# Justfile additions
test-integration: build
	@echo "🐳 Running integration tests in Docker containers (safe for local development)"
	PLONK_TEST_BINARY=./plonk go test -v -tags=integration ./tests/integration/...

test-integration-quick: build
	# Run only fast integration tests
	PLONK_TEST_BINARY=./plonk go test -v -short -tags=integration ./tests/integration/...
```

## Test Data Management

### 1. Fixture Repository
Create `github.com/plonk-test/fixtures` with:
```
fixtures/
├── minimal/
│   ├── .bashrc
│   ├── .gitconfig
│   └── plonk.lock
├── complex/
│   ├── .config/
│   ├── .vim/
│   └── plonk.lock
└── broken/
    └── plonk.lock  # Invalid format for error testing
```

### 2. Test Package Registry
Set up a local package registry for predictable testing:
- Docker registry for container images
- Verdaccio for npm packages
- Local file server for Homebrew formulas

## Questions for Discussion

### 1. Test Environment Strategy
**Option A**: Use Docker containers for complete isolation
- ✅ Pro: Perfect isolation, consistent environment
- ❌ Con: Requires Docker, slower startup

**Option B**: Use VMs (e.g., via Vagrant)
- ✅ Pro: More realistic environment
- ❌ Con: Much slower, resource intensive

**Option C**: Use separate user accounts on host
- ✅ Pro: Fast, no containerization overhead
- ❌ Con: Less isolation, potential for system contamination

**Recommendation**: Option A for CI, with Option C as a fast local alternative?

### 2. Test Data Approach
Should we:
- Use real packages (jq, ripgrep) that are stable and small?
- Create minimal test packages that we control?
- Use a mix based on what we're testing?

### 3. Performance Considerations
Integration tests will be slow. Should we:
- Run full suite only on main branch merges?
- Run targeted tests based on changed files in PRs?
- Implement test sharding across multiple containers?

### 4. Error Injection
How should we test error scenarios?
- Network failures during package downloads
- Disk full conditions
- Permission errors
- Corrupted lock files

### 5. Platform Coverage
Current unit tests assume Unix-like systems. Should integration tests:
- Focus only on Linux (our primary CI environment)?
- Include macOS-specific tests (Homebrew behavior)?
- Test WSL scenarios for Windows users?

## Success Metrics

1. **Coverage Target**: Combined unit + integration tests reach 80%+ coverage
2. **Execution Time**: Full suite runs in < 10 minutes on CI
3. **Reliability**: < 1% flaky test rate
4. **Developer Experience**: Can run targeted tests in < 30 seconds locally

## Implementation Timeline

### Phase 1: Foundation (Week 1-2)
- Set up Docker test environment
- Create basic test harness
- Implement first package installation test

### Phase 2: Core Features (Week 3-4)
- Package manager tests (all supported managers)
- Dotfile management tests
- Basic apply/remove workflows

### Phase 3: Advanced Scenarios (Week 5-6)
- Clone and setup tests
- Error handling and recovery
- Lock file corruption/recovery

### Phase 4: CI Integration (Week 7-8)
- GitHub Actions workflow
- Test result reporting
- Performance optimization

## Proposed Hybrid Approach

Based on our research and the 54.9% of code that requires system interaction testing, we propose a hybrid strategy:

### 1. **Subprocess Tests (Fast, Local-Friendly)**
For basic CLI behavior and command flow testing:
```go
// tests/integration/cli_test.go
func TestCLICommands(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    binary := buildTestBinary(t)

    tests := []struct {
        name     string
        args     []string
        wantExit int
        wantOut  string
    }{
        {"help command", []string{"--help"}, 0, "Plonk helps you"},
        {"invalid command", []string{"invalid"}, 1, "unknown command"},
        {"status without config", []string{"status"}, 0, "No packages or dotfiles"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command(binary, tt.args...)
            cmd.Env = append(os.Environ(),
                "HOME=" + t.TempDir(),
                "PLONK_DIR=" + t.TempDir(),
                "GOCOVERDIR=.coverdata", // Collect coverage
            )
            output, err := cmd.CombinedOutput()
            // Verify exit code and output
        })
    }
}
```

### 2. **Container Tests (Comprehensive, CI-Focused)**
For testing actual package installations and system modifications:
```go
// tests/integration/container_test.go
func TestPackageOperations(t *testing.T) {
    if !runContainerTests() {
        t.Skip("Container tests disabled")
    }

    ctx := context.Background()
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "plonk-test:latest",
            Env: map[string]string{
                "PLONK_TEST_MODE": "true",
            },
            Mounts: testcontainers.Mounts(
                testcontainers.BindMount(getBinaryPath(), "/usr/local/bin/plonk"),
            ),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)

    // Test actual package installation
    code, output, err := container.Exec(ctx, []string{
        "plonk", "install", "brew:jq",
    })
    require.NoError(t, err)
    assert.Equal(t, 0, code)
    assert.Contains(t, string(output), "Successfully installed")
}
```

### 3. **BATS Tests (User-Facing Behavior)**
Keep existing BATS tests for black-box testing:
```bash
# tests/bats/install.bats
@test "install a brew package" {
    run plonk install brew:jq
    assert_success
    assert_output --partial "Successfully installed"

    run plonk status
    assert_success
    assert_output --partial "jq"
}
```

## Implementation Plan

### Core Design Principles

1. **Safety First**: Complete isolation via containers
2. **Performance Matters**: Parallel execution, smart container reuse
3. **Developer Experience**: Fast feedback, easy commands, clear output

### Phase 1: Container Infrastructure (Week 1-2)

#### Parallel Test Architecture
```go
// tests/integration/base_test.go
func TestMain(m *testing.M) {
    // Pre-build test image once for all suites
    buildTestImage()

    // Run test suites in parallel
    os.Exit(m.Run())
}

// Each test file gets its own container
func setupTestSuite(t *testing.T) *TestEnvironment {
    t.Parallel() // Enable parallel execution

    env := NewTestEnvironment(t)
    t.Cleanup(func() { env.Cleanup() })

    return env
}
```

#### Smart Container Reuse
```go
// One container per test suite, not per test
func TestPackageManagerSuite(t *testing.T) {
    env := setupTestSuite(t) // Container created once

    // All package tests share this container
    t.Run("brew_install", func(t *testing.T) { ... })
    t.Run("npm_install", func(t *testing.T) { ... })
    t.Run("pip_install", func(t *testing.T) { ... })
    t.Run("status_command", func(t *testing.T) { ... })
}

func TestDotfileSuite(t *testing.T) {
    env := setupTestSuite(t) // Different container, runs in parallel

    t.Run("add_dotfiles", func(t *testing.T) { ... })
    t.Run("apply_dotfiles", func(t *testing.T) { ... })
    t.Run("remove_dotfiles", func(t *testing.T) { ... })
}
```

#### Developer-Friendly Commands
```bash
# Justfile additions
# Run all integration tests (parallel by default)
test-integration: build
    @echo "🐳 Running integration tests (parallel, containerized)"
    go test -v -parallel 4 -tags=integration ./tests/integration/...

# Run specific test suite
test-integration-packages: build
    go test -v -tags=integration ./tests/integration/packages_test.go

# Quick smoke test - most important tests only
test-integration-smoke: build
    go test -v -short -tags=integration ./tests/integration/...

# Debug mode - serial execution, verbose output
test-integration-debug: build
    go test -v -parallel 1 -tags=integration ./tests/integration/... -debug
```

### Phase 2: Basic Package Operations (Week 3-4)
1. **Installation Tests**
   - Test installing real packages: jq, ripgrep, tree
   - Verify packages are actually usable after installation
   - Test multiple package managers (brew, npm, pip)

2. **Network Error Handling**
   - Test behavior when package downloads fail
   - Simulate DNS resolution failures
   - Ensure graceful error messages

3. **Lock File Operations**
   - Verify lock file updates after installations
   - Test lock file corruption recovery
   - Validate format compatibility

### Phase 3: Dotfile Management (Week 5-6)
1. **Symlink Operations**
   - Test adding and removing dotfiles
   - Verify symlink creation and cleanup
   - Test conflict resolution

2. **Apply/Remove Workflows**
   - Test full apply command with packages and dotfiles
   - Verify remove command cleans up properly
   - Test selective operations (--packages-only, --dotfiles-only)

### Phase 4: CI Integration (Week 7)
1. **GitHub Actions Setup**
   - Create Docker build workflow for test images
   - Add integration test job to CI pipeline
   - Set up test result reporting

2. **Local Developer Experience**
   - Update Justfile with test-integration command
   - Add pre-flight checks for Docker availability
   - Create troubleshooting documentation

## Performance Optimizations

### 1. Image Caching
```dockerfile
# Pre-install common dependencies in base image
FROM ubuntu:22.04 AS plonk-test-base

# Install once, reuse for all tests
RUN apt-get update && apt-get install -y \
    curl git build-essential python3-pip nodejs npm && \
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Cache common test packages
RUN brew install jq ripgrep tree
```

### 2. Parallel Execution Strategy
```go
// Run independent test suites in parallel
// packages_test.go
func TestPackages(t *testing.T) {
    t.Parallel() // Can run alongside other suites
}

// dotfiles_test.go
func TestDotfiles(t *testing.T) {
    t.Parallel() // Independent, runs concurrently
}

// clone_test.go
func TestClone(t *testing.T) {
    t.Parallel() // No shared state needed
}
```

### 3. Test Organization for Speed
```
tests/integration/
├── smoke_test.go      # Critical path tests only (30s)
├── packages_test.go   # All package managers (2-3m)
├── dotfiles_test.go   # Dotfile operations (1-2m)
├── clone_test.go      # Repository cloning (1m)
└── full_test.go       # Complete workflows (5m)
```

### 4. Container Lifecycle Management
```go
type TestEnvironment struct {
    container testcontainers.Container
    mu        sync.Mutex
    cleaned   bool
}

// Reuse container within a test suite
func (e *TestEnvironment) Reset() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    // Quick reset instead of new container
    _, err := e.RunCommand("rm -rf ~/.config/plonk/*")
    return err
}
```

## Test Container Setup

### Base Test Image
We'll use Ubuntu 22.04 as our single test environment:

```dockerfile
# tests/integration/dockerfiles/Dockerfile
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    python3-pip \
    nodejs \
    npm

# Install Homebrew on Linux
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
ENV PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"

# Create test user
RUN useradd -m testuser
USER testuser
WORKDIR /home/testuser
```

## Network Error Handling Strategy

Since we're using real packages, we need robust handling for network issues:

```go
func TestPackageInstallationWithRetry(t *testing.T) {
    env := NewTestEnvironment(t)
    defer env.Cleanup()

    packages := []string{"brew:jq", "brew:ripgrep", "npm:prettier"}

    for _, pkg := range packages {
        t.Run(pkg, func(t *testing.T) {
            // Allow up to 3 attempts for network issues
            var lastErr error
            for i := 0; i < 3; i++ {
                output, err := env.RunPlonk("install", pkg)
                if err == nil {
                    assert.Contains(t, output, "Successfully installed")
                    return
                }
                lastErr = err
                if !isNetworkError(err) {
                    break // Non-network errors shouldn't retry
                }
                time.Sleep(time.Second * time.Duration(i+1))
            }

            // If we get here, check if it's a known network issue
            if isNetworkError(lastErr) {
                t.Skip("Skipping due to network issues: " + lastErr.Error())
            }
            t.Fatalf("Installation failed: %v", lastErr)
        })
    }
}
```

## Developer Experience Enhancements

### 1. Fast Feedback Loop
```bash
# Justfile commands for common workflows
test-watch: ## Run tests on file change
    watchexec -e go -- just test-integration-smoke

test-failed: ## Re-run only failed tests
    go test -v -run $(go test -list | grep FAIL) ./tests/integration/...

test-focus pattern: ## Run tests matching pattern
    go test -v -run {{pattern}} -tags=integration ./tests/integration/...
```

### 2. Clear Test Output
```go
// Helper for readable test output
func (e *TestEnvironment) MustRun(t *testing.T, args ...string) string {
    output, err := e.RunPlonk(args...)
    if err != nil {
        t.Logf("Command failed: plonk %s", strings.Join(args, " "))
        t.Logf("Output: %s", output)
        t.Logf("Error: %v", err)
        t.FailNow()
    }
    return output
}
```

### 3. Pre-flight Checks
```go
func TestMain(m *testing.M) {
    // Check Docker is available
    if err := checkDocker(); err != nil {
        fmt.Fprintf(os.Stderr, "⚠️  Docker required for integration tests\n")
        fmt.Fprintf(os.Stderr, "   Install from: https://docker.com\n")
        os.Exit(1)
    }

    // Pull/build image with progress
    fmt.Println("🐳 Preparing test environment...")
    buildTestImage()

    fmt.Println("✅ Ready to run tests")
    os.Exit(m.Run())
}
```

### 4. Debugging Support
```go
// Environment variable for debugging
func (e *TestEnvironment) RunPlonk(args ...string) (string, error) {
    if os.Getenv("PLONK_TEST_DEBUG") != "" {
        log.Printf("Running: plonk %s", strings.Join(args, " "))
    }
    // ... run command ...
}

// Keep failed containers for inspection
func (e *TestEnvironment) Cleanup() {
    if t.Failed() && os.Getenv("PLONK_TEST_KEEP_FAILED") != "" {
        t.Logf("Container kept for debugging: %s", e.container.ID)
        t.Logf("Connect with: docker exec -it %s /bin/bash", e.container.ID)
        return
    }
    e.container.Terminate(context.Background())
}
```

## Success Metrics

1. **Coverage**: Combined unit + integration tests reach 80%+
2. **Safety**: Zero risk to developer machines (all containerized)
3. **Speed**: Full integration suite runs in < 10 minutes (< 30s for smoke tests)
4. **Reliability**: < 5% flake rate due to network issues
5. **Developer Experience**: Single command, clear output, fast feedback
6. **Parallel Execution**: 4x speedup with parallel test suites

## Final Strategy Summary

### Architecture Overview
```
┌─────────────────────────────────────────────────┐
│              Developer Machine                   │
│  ┌─────────────────────────────────────────┐   │
│  │         just test-integration            │   │
│  └─────────────────┬───────────────────────┘   │
│                    │                            │
│  ┌─────────────────┴───────────────────────┐   │
│  │            Go Test Runner               │   │
│  │         (Parallel Execution)            │   │
│  └────┬──────────┬──────────┬─────────────┘   │
│       │          │          │                  │
│  ┌────┴───┐ ┌───┴────┐ ┌───┴────┐           │
│  │Container│ │Container│ │Container│           │
│  │ Suite 1 │ │ Suite 2 │ │ Suite 3 │           │
│  │(Ubuntu) │ │(Ubuntu) │ │(Alpine) │           │
│  └────────┘ └────────┘ └────────┘           │
│                                                │
│  Developer's actual plonk config remains       │
│  completely untouched and isolated             │
└─────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **Container-Per-Suite**: Balance between isolation and performance
2. **Real Package Testing**: Use actual packages (jq, ripgrep) not mocks
3. **Parallel by Default**: Leverage modern multi-core machines
4. **Network Resilience**: Retry logic for transient failures
5. **Developer First**: Fast smoke tests, clear output, easy debugging

### Test Organization
- **Smoke Tests**: 30 seconds - critical user journeys
- **Feature Tests**: 2-3 minutes each - focused functionality
- **Full Suite**: <10 minutes - comprehensive coverage
- **All Tests**: Completely isolated, safe to run anytime

## Conclusion

This integration testing strategy achieves our goals:
- **Primary**: Absolute safety through complete containerization
- **Secondary**: Fast execution through parallelization and smart grouping
- **Secondary**: Excellent developer experience with clear commands and output

The strategy fills the 54.9% coverage gap from unit tests while maintaining zero risk to developer systems.
