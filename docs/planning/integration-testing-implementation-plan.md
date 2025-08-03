# Integration Testing Implementation Plan

**Created**: 2025-08-03
**Status**: READY FOR IMPLEMENTATION
**Target**: v1.1
**Estimated Duration**: 4-6 weeks

## Overview

This document provides a detailed, step-by-step implementation plan for adding containerized integration tests to Plonk. The plan prioritizes safety, performance, and developer experience.

## Implementation Phases

### Phase 0: Foundation Setup (Week 1)

#### 0.1 Project Structure
```bash
# Create directory structure
mkdir -p tests/integration/{dockerfiles,testdata,helpers}
mkdir -p tests/integration/suites/{smoke,packages,dotfiles,workflows}
```

#### 0.2 Dependencies
```go
// go.mod additions
require (
    github.com/testcontainers/testcontainers-go v0.33.0
    github.com/docker/docker v25.0.0
    github.com/docker/go-connections v0.5.0
)
```

#### 0.3 Base Test Image
```dockerfile
# tests/integration/dockerfiles/Dockerfile
FROM ubuntu:22.04

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    sudo \
    locales \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set up locale
RUN locale-gen en_US.UTF-8
ENV LANG=en_US.UTF-8

# Install Node.js and npm
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs

# Install Python and pip
RUN apt-get update && apt-get install -y python3 python3-pip

# Create test user with sudo (no password)
RUN useradd -m -s /bin/bash testuser && \
    echo "testuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Switch to test user
USER testuser
WORKDIR /home/testuser

# Install Homebrew
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Add Homebrew to PATH
ENV PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"

# Pre-install common test packages to speed up tests
RUN brew install jq ripgrep tree

# Set up clean home directory
RUN mkdir -p ~/.config ~/.local/bin
```

#### 0.4 Test Environment Helper
```go
// tests/integration/helpers/environment.go
package helpers

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

type TestEnvironment struct {
    t         *testing.T
    container testcontainers.Container
    ctx       context.Context
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
    t.Helper()

    ctx := context.Background()

    // Get plonk binary path
    plonkBinary := os.Getenv("PLONK_TEST_BINARY")
    if plonkBinary == "" {
        plonkBinary = filepath.Join(projectRoot(), "plonk")
    }

    req := testcontainers.ContainerRequest{
        Image: "plonk-test:latest",
        Env: map[string]string{
            "PLONK_TEST_MODE": "true",
            "HOME":            "/home/testuser",
        },
        Files: []testcontainers.ContainerFile{
            {
                HostFilePath:      plonkBinary,
                ContainerFilePath: "/usr/local/bin/plonk",
                FileMode:          0755,
            },
        },
        WaitingFor: wait.ForLog("ready").WithStartupTimeout(30 * time.Second),
        Cmd:        []string{"sh", "-c", "echo ready && tail -f /dev/null"},
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("Failed to start container: %v", err)
    }

    env := &TestEnvironment{
        t:         t,
        container: container,
        ctx:       ctx,
    }

    // Register cleanup
    t.Cleanup(func() {
        if t.Failed() && os.Getenv("PLONK_TEST_KEEP_FAILED") != "" {
            id := container.GetContainerID()
            t.Logf("Container kept for debugging: %s", id[:12])
            t.Logf("Connect with: docker exec -it %s /bin/bash", id[:12])
            return
        }
        container.Terminate(ctx)
    })

    return env
}

func (e *TestEnvironment) RunPlonk(args ...string) (string, error) {
    return e.RunCommand("plonk", args...)
}

func (e *TestEnvironment) RunCommand(cmd string, args ...string) (string, error) {
    e.t.Helper()

    fullCmd := append([]string{cmd}, args...)

    if os.Getenv("PLONK_TEST_DEBUG") != "" {
        e.t.Logf("Running: %s", strings.Join(fullCmd, " "))
    }

    exitCode, output, err := e.container.Exec(e.ctx, fullCmd)
    if err != nil {
        return "", fmt.Errorf("exec failed: %w", err)
    }

    outputStr := string(output)

    if exitCode != 0 {
        return outputStr, fmt.Errorf("command failed with exit code %d", exitCode)
    }

    return outputStr, nil
}

func (e *TestEnvironment) MustRun(args ...string) string {
    e.t.Helper()

    output, err := e.RunPlonk(args...)
    if err != nil {
        e.t.Logf("Command failed: plonk %s", strings.Join(args, " "))
        e.t.Logf("Output: %s", output)
        e.t.Logf("Error: %v", err)
        e.t.FailNow()
    }
    return output
}

func (e *TestEnvironment) WriteFile(path, content string) error {
    cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", path, content)
    _, err := e.RunCommand("sh", "-c", cmd)
    return err
}

func (e *TestEnvironment) FileExists(path string) bool {
    _, err := e.RunCommand("test", "-f", path)
    return err == nil
}

func (e *TestEnvironment) ReadFile(path string) (string, error) {
    return e.RunCommand("cat", path)
}
```

### Phase 1: Core Test Infrastructure (Week 1-2)

#### 1.1 Test Main Setup
```go
// tests/integration/main_test.go
package integration

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "testing"
)

func TestMain(m *testing.M) {
    // Check Docker availability
    if err := checkDocker(); err != nil {
        fmt.Fprintf(os.Stderr, "⚠️  Docker is required for integration tests\n")
        fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
        fmt.Fprintf(os.Stderr, "   Install Docker from: https://docker.com\n")
        os.Exit(1)
    }

    // Build plonk binary
    fmt.Println("🔨 Building plonk binary...")
    if err := buildPlonk(); err != nil {
        log.Fatalf("Failed to build plonk: %v", err)
    }

    // Build test container image
    fmt.Println("🐳 Building test container image...")
    if err := buildTestImage(); err != nil {
        log.Fatalf("Failed to build test image: %v", err)
    }

    fmt.Println("✅ Test environment ready")

    // Run tests
    code := m.Run()
    os.Exit(code)
}

func checkDocker() error {
    cmd := exec.Command("docker", "version")
    return cmd.Run()
}

func buildPlonk() error {
    cmd := exec.Command("go", "build", "-o", "plonk", "./cmd/plonk")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func buildTestImage() error {
    cmd := exec.Command("docker", "build",
        "-t", "plonk-test:latest",
        "-f", "tests/integration/dockerfiles/Dockerfile",
        ".")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

#### 1.2 Justfile Integration
```justfile
# Run all integration tests (parallel)
test-integration: build
    @echo "🐳 Running integration tests (parallel, containerized)"
    PLONK_TEST_BINARY={{justfile_directory()}}/plonk \
        go test -v -parallel 4 -tags=integration -timeout 15m ./tests/integration/...

# Run smoke tests only (fast)
test-integration-smoke: build
    @echo "🚀 Running smoke tests (fast path)"
    PLONK_TEST_BINARY={{justfile_directory()}}/plonk \
        go test -v -short -tags=integration -timeout 2m ./tests/integration/suites/smoke/...

# Run specific test suite
test-integration-suite suite: build
    @echo "🧪 Running {{suite}} test suite"
    PLONK_TEST_BINARY={{justfile_directory()}}/plonk \
        go test -v -tags=integration ./tests/integration/suites/{{suite}}/...

# Debug mode - keep failed containers
test-integration-debug: build
    @echo "🐛 Running tests in debug mode (failed containers will be kept)"
    PLONK_TEST_KEEP_FAILED=1 PLONK_TEST_DEBUG=1 \
    PLONK_TEST_BINARY={{justfile_directory()}}/plonk \
        go test -v -parallel 1 -tags=integration ./tests/integration/...

# Watch mode - re-run on changes
test-integration-watch:
    watchexec -e go,dockerfile -- just test-integration-smoke
```

### Phase 2: Smoke Tests (Week 2)

#### 2.1 Basic CLI Operations
```go
// tests/integration/suites/smoke/cli_test.go
//go:build integration

package smoke

import (
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/plonk/tests/integration/helpers"
)

func TestCLIBasics(t *testing.T) {
    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    t.Run("help_command", func(t *testing.T) {
        output := env.MustRun("--help")
        assert.Contains(t, output, "Plonk helps you")
        assert.Contains(t, output, "Available Commands:")
    })

    t.Run("version_command", func(t *testing.T) {
        output := env.MustRun("--version")
        assert.Contains(t, output, "plonk version")
    })

    t.Run("status_empty", func(t *testing.T) {
        output := env.MustRun("status")
        assert.Contains(t, output, "No packages or dotfiles")
    })

    t.Run("invalid_command", func(t *testing.T) {
        output, err := env.RunPlonk("invalid-command")
        assert.Error(t, err)
        assert.Contains(t, output, "unknown command")
    })
}

func TestQuickPackageInstall(t *testing.T) {
    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    // Install a small, fast package
    output := env.MustRun("install", "brew:tree")
    assert.Contains(t, output, "Successfully installed")

    // Verify it's in status
    output = env.MustRun("status")
    assert.Contains(t, output, "tree")

    // Verify it actually works
    output, err := env.RunCommand("tree", "--version")
    assert.NoError(t, err)
    assert.Contains(t, output, "tree v")
}
```

### Phase 3: Package Manager Tests (Week 3)

#### 3.1 Homebrew Tests
```go
// tests/integration/suites/packages/homebrew_test.go
//go:build integration

package packages

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/plonk/tests/integration/helpers"
)

func TestHomebrewOperations(t *testing.T) {
    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    packages := []string{"jq", "ripgrep", "tree"}

    t.Run("install_multiple_packages", func(t *testing.T) {
        for _, pkg := range packages {
            output := env.MustRun("install", "brew:"+pkg)
            assert.Contains(t, output, "Successfully installed")
        }
    })

    t.Run("verify_installations", func(t *testing.T) {
        // Check plonk status
        output := env.MustRun("status")
        for _, pkg := range packages {
            assert.Contains(t, output, pkg)
        }

        // Verify packages work
        cmds := map[string]string{
            "jq":      "--version",
            "rg":      "--version",
            "tree":    "--version",
        }

        for cmd, arg := range cmds {
            output, err := env.RunCommand(cmd, arg)
            assert.NoError(t, err, "Command %s should work", cmd)
            assert.NotEmpty(t, output)
        }
    })

    t.Run("uninstall_package", func(t *testing.T) {
        output := env.MustRun("uninstall", "brew:tree")
        assert.Contains(t, output, "Successfully removed")

        // Verify it's gone
        _, err := env.RunCommand("tree", "--version")
        assert.Error(t, err)
    })

    t.Run("reinstall_package", func(t *testing.T) {
        // Install
        env.MustRun("install", "brew:tree")

        // Try to install again
        output, err := env.RunPlonk("install", "brew:tree")
        assert.NoError(t, err)
        assert.Contains(t, output, "already installed")
    })
}

func TestHomebrewNetworkResilience(t *testing.T) {
    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    t.Run("install_with_retry", func(t *testing.T) {
        packages := []string{"brew:jq", "brew:ripgrep"}

        for _, pkg := range packages {
            var lastErr error
            for attempt := 1; attempt <= 3; attempt++ {
                output, err := env.RunPlonk("install", pkg)
                if err == nil {
                    assert.Contains(t, output, "Successfully installed")
                    break
                }

                lastErr = err
                if !isNetworkError(output) {
                    t.Fatalf("Non-network error: %v\nOutput: %s", err, output)
                }

                t.Logf("Network error on attempt %d, retrying...", attempt)
                time.Sleep(time.Second * time.Duration(attempt))
            }

            if lastErr != nil && isNetworkError(lastErr.Error()) {
                t.Skipf("Skipping due to persistent network issues: %v", lastErr)
            }
        }
    })
}

func isNetworkError(output string) bool {
    networkErrors := []string{
        "connection refused",
        "connection reset",
        "no such host",
        "timeout",
        "temporary failure",
    }

    for _, err := range networkErrors {
        if strings.Contains(strings.ToLower(output), err) {
            return true
        }
    }
    return false
}
```

#### 3.2 NPM Tests
```go
// tests/integration/suites/packages/npm_test.go
//go:build integration

package packages

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/plonk/tests/integration/helpers"
)

func TestNPMOperations(t *testing.T) {
    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    t.Run("install_global_package", func(t *testing.T) {
        output := env.MustRun("install", "npm:prettier")
        assert.Contains(t, output, "Successfully installed")

        // Verify it works
        output, err := env.RunCommand("prettier", "--version")
        assert.NoError(t, err)
        assert.NotEmpty(t, output)
    })

    t.Run("multiple_npm_packages", func(t *testing.T) {
        packages := []string{"eslint", "typescript"}

        for _, pkg := range packages {
            output := env.MustRun("install", "npm:"+pkg)
            assert.Contains(t, output, "Successfully installed")
        }

        // Verify status shows all
        output := env.MustRun("status")
        assert.Contains(t, output, "prettier")
        assert.Contains(t, output, "eslint")
        assert.Contains(t, output, "typescript")
    })
}
```

### Phase 4: Dotfile Management Tests (Week 4)

#### 4.1 Dotfile Operations
```go
// tests/integration/suites/dotfiles/dotfile_test.go
//go:build integration

package dotfiles

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/plonk/tests/integration/helpers"
)

func TestDotfileOperations(t *testing.T) {
    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    // Initialize plonk
    env.MustRun("init")

    t.Run("add_dotfiles", func(t *testing.T) {
        // Create test dotfiles
        require.NoError(t, env.WriteFile(".bashrc", "# My bashrc\nexport EDITOR=vim"))
        require.NoError(t, env.WriteFile(".gitconfig", "[user]\nname = Test User"))

        // Add them to plonk
        output := env.MustRun("dot", "add", ".bashrc", ".gitconfig")
        assert.Contains(t, output, "Added")

        // Verify they're tracked
        output = env.MustRun("status")
        assert.Contains(t, output, ".bashrc")
        assert.Contains(t, output, ".gitconfig")
    })

    t.Run("apply_dotfiles", func(t *testing.T) {
        // Remove originals
        env.RunCommand("rm", ".bashrc", ".gitconfig")

        // Apply
        output := env.MustRun("apply", "--dotfiles-only")
        assert.Contains(t, output, "Created")

        // Verify symlinks
        output, err := env.RunCommand("readlink", ".bashrc")
        assert.NoError(t, err)
        assert.Contains(t, output, ".local/share/plonk")
    })

    t.Run("modify_and_diff", func(t *testing.T) {
        // Modify a tracked file
        require.NoError(t, env.WriteFile(".bashrc", "# Modified\nexport EDITOR=nano"))

        // Check diff
        output := env.MustRun("diff")
        assert.Contains(t, output, "modified")
        assert.Contains(t, output, ".bashrc")
    })

    t.Run("remove_dotfile", func(t *testing.T) {
        output := env.MustRun("dot", "remove", ".gitconfig")
        assert.Contains(t, output, "Removed")

        // Verify it's gone from status
        output = env.MustRun("status")
        assert.NotContains(t, output, ".gitconfig")
    })
}
```

### Phase 5: Workflow Tests (Week 5)

#### 5.1 Complete User Workflows
```go
// tests/integration/suites/workflows/complete_workflow_test.go
//go:build integration

package workflows

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/plonk/tests/integration/helpers"
)

func TestCompleteSetupWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping complete workflow test in short mode")
    }

    t.Parallel()
    env := helpers.NewTestEnvironment(t)

    t.Run("clone_repository", func(t *testing.T) {
        // Use a test fixtures repository
        output := env.MustRun("clone", "github.com/plonk-test/fixtures", "--no-apply")
        assert.Contains(t, output, "Repository cloned successfully")
        assert.Contains(t, output, "Detected managers")
    })

    t.Run("review_and_apply", func(t *testing.T) {
        // Check what would be installed
        output := env.MustRun("status")
        assert.Contains(t, output, "Packages")
        assert.Contains(t, output, "Dotfiles")

        // Apply everything
        output = env.MustRun("apply")
        assert.Contains(t, output, "Successfully installed")
        assert.Contains(t, output, "Created")
    })

    t.Run("verify_setup", func(t *testing.T) {
        // Check packages work
        _, err := env.RunCommand("jq", "--version")
        assert.NoError(t, err)

        // Check dotfiles exist
        assert.True(t, env.FileExists(".bashrc"))
        assert.True(t, env.FileExists(".gitconfig"))
    })

    t.Run("clean_removal", func(t *testing.T) {
        // Remove everything
        output := env.MustRun("remove")
        assert.Contains(t, output, "Removed")

        // Verify clean state
        output = env.MustRun("status")
        assert.Contains(t, output, "No packages or dotfiles")
    })
}
```

### Phase 6: CI Integration (Week 6)

#### 6.1 GitHub Actions Workflow
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
    timeout-minutes: 30

    strategy:
      fail-fast: false
      matrix:
        suite: [smoke, packages, dotfiles, workflows]

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.23'

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Cache Docker layers
      uses: actions/cache@v4
      with:
        path: /tmp/.buildx-cache
        key: ${{ runner.os }}-buildx-${{ github.sha }}
        restore-keys: |
          ${{ runner.os }}-buildx-

    - name: Build plonk binary
      run: go build -o plonk ./cmd/plonk

    - name: Build test image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: tests/integration/dockerfiles/Dockerfile
        tags: plonk-test:latest
        load: true
        cache-from: type=local,src=/tmp/.buildx-cache
        cache-to: type=local,dest=/tmp/.buildx-cache-new,mode=max

    - name: Run ${{ matrix.suite }} tests
      run: |
        PLONK_TEST_BINARY=$PWD/plonk \
        go test -v -tags=integration -timeout 10m \
          ./tests/integration/suites/${{ matrix.suite }}/...

    - name: Upload test logs on failure
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: test-logs-${{ matrix.suite }}
        path: |
          tests/integration/**/*.log
          /tmp/plonk-test-*
```

## Testing Patterns and Best Practices

### 1. Test Organization
```go
// Group related tests in the same suite
func TestPackageManagerSuite(t *testing.T) {
    t.Parallel() // Always enable parallel execution
    env := helpers.NewTestEnvironment(t) // One container per suite

    // Logical test progression
    t.Run("install", func(t *testing.T) { /* ... */ })
    t.Run("verify", func(t *testing.T) { /* ... */ })
    t.Run("update", func(t *testing.T) { /* ... */ })
    t.Run("remove", func(t *testing.T) { /* ... */ })
}
```

### 2. Assertions
```go
// Use descriptive assertions
assert.Contains(t, output, "Successfully installed",
    "Installation should report success")

// Verify actual functionality, not just plonk output
output, err := env.RunCommand("jq", "--version")
assert.NoError(t, err, "Installed package should be executable")
```

### 3. Error Handling
```go
// Always handle network errors gracefully
if isNetworkError(err) {
    t.Skipf("Skipping due to network error: %v", err)
}

// Provide helpful debug information
if err != nil {
    t.Logf("Command output: %s", output)
    t.Logf("Container logs: %s", env.GetLogs())
}
```

### 4. Test Data
```go
// Use small, fast packages for tests
fastPackages := []string{"tree", "jq", "bat"}

// Avoid large packages unless testing specific scenarios
// slowPackages := []string{"node", "rust", "docker"} // Don't use these
```

## Maintenance Guidelines

### 1. Adding New Tests
1. Determine correct suite (smoke, packages, dotfiles, workflows)
2. Follow existing patterns in that suite
3. Enable parallel execution unless tests must be serial
4. Add to appropriate test group in CI matrix

### 2. Debugging Failed Tests
```bash
# Run specific failing test with debug output
PLONK_TEST_DEBUG=1 just test-integration-suite packages

# Keep failed container for inspection
PLONK_TEST_KEEP_FAILED=1 just test-integration-debug

# Connect to kept container
docker exec -it <container-id> /bin/bash
```

### 3. Performance Monitoring
- Smoke tests: Must complete in < 30 seconds
- Individual suites: Must complete in < 3 minutes
- Full suite: Must complete in < 10 minutes
- If tests exceed limits, split into smaller suites

## Coverage Reporting

### Combined Coverage Strategy

Go 1.20+ supports coverage collection from binaries using the `GOCOVERDIR` environment variable. This allows us to collect coverage from integration tests that run the actual plonk binary.

#### 1. Coverage Collection Setup
```go
// tests/integration/helpers/coverage.go
package helpers

import (
    "os"
    "path/filepath"
    "testing"
)

func SetupCoverage(t *testing.T) string {
    t.Helper()

    // Create coverage directory
    coverDir := filepath.Join(t.TempDir(), "coverage")
    if err := os.MkdirAll(coverDir, 0755); err != nil {
        t.Fatalf("Failed to create coverage dir: %v", err)
    }

    return coverDir
}

// In TestEnvironment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
    // ... existing code ...

    coverDir := SetupCoverage(t)

    req := testcontainers.ContainerRequest{
        // ... existing config ...
        Env: map[string]string{
            "GOCOVERDIR": "/coverage", // Coverage collection in container
        },
        Mounts: testcontainers.Mounts(
            testcontainers.BindMount(coverDir, "/coverage"),
        ),
    }

    // ... rest of setup ...
}
```

#### 2. Building with Coverage Support
```bash
# Justfile additions
# Build plonk with coverage instrumentation
build-coverage:
    go build -cover -o plonk ./cmd/plonk

# Run all tests with coverage
test-all-coverage: build-coverage
    @echo "🧪 Running all tests with coverage"
    # Clean previous coverage
    rm -rf coverage/
    mkdir -p coverage/integration coverage/unit

    # Run unit tests with coverage
    go test -coverprofile=coverage/unit/coverage.out ./...

    # Run integration tests with coverage
    GOCOVERDIR=coverage/integration PLONK_TEST_BINARY={{justfile_directory()}}/plonk \
        go test -v -tags=integration ./tests/integration/...

    # Merge coverage data
    go tool covdata textfmt -i=coverage/integration -o coverage/integration.txt
    just merge-coverage

# Merge unit and integration coverage
merge-coverage:
    @echo "📊 Merging coverage reports"
    # Convert binary coverage to text format
    go tool covdata textfmt -i=coverage/integration -o coverage/integration.txt

    # Merge all coverage files
    gocovmerge coverage/unit/coverage.out coverage/integration.txt > coverage/combined.out

    # Generate reports
    go tool cover -html=coverage/combined.out -o coverage/report.html
    go tool cover -func=coverage/combined.out | tail -n 1

    @echo "✅ Coverage report: coverage/report.html"
```

#### 3. CI Coverage Collection
```yaml
# .github/workflows/coverage.yml
name: Combined Coverage

on:
  push:
    branches: [main]
  pull_request:

jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.23'

    - name: Install gocovmerge
      run: go install github.com/wadey/gocovmerge@latest

    - name: Build with coverage
      run: go build -cover -o plonk ./cmd/plonk

    - name: Run unit tests
      run: |
        mkdir -p coverage/unit
        go test -coverprofile=coverage/unit/coverage.out ./...

    - name: Build test image
      run: |
        docker build -t plonk-test:latest \
          -f tests/integration/dockerfiles/Dockerfile .

    - name: Run integration tests
      run: |
        mkdir -p coverage/integration
        GOCOVERDIR=coverage/integration \
        PLONK_TEST_BINARY=$PWD/plonk \
          go test -v -tags=integration ./tests/integration/...

    - name: Merge coverage
      run: |
        # Convert binary coverage to text
        go tool covdata textfmt -i=coverage/integration -o coverage/integration.txt

        # Merge all coverage
        gocovmerge coverage/unit/coverage.out coverage/integration.txt > coverage/combined.out

        # Generate summary
        echo "### Coverage Report" >> $GITHUB_STEP_SUMMARY
        echo '```' >> $GITHUB_STEP_SUMMARY
        go tool cover -func=coverage/combined.out | tail -n 10 >> $GITHUB_STEP_SUMMARY
        echo '```' >> $GITHUB_STEP_SUMMARY

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage/combined.out
        flags: combined
        name: combined-coverage

    - name: Upload coverage artifacts
      uses: actions/upload-artifact@v4
      with:
        name: coverage-report
        path: |
          coverage/combined.out
          coverage/report.html
```

#### 4. Local Coverage Visualization
```bash
# Justfile additions
# View coverage in browser
coverage-view: test-all-coverage
    @echo "🌐 Opening coverage report"
    open coverage/report.html

# Coverage by package
coverage-by-package: test-all-coverage
    @echo "📦 Coverage by package:"
    go tool cover -func=coverage/combined.out | grep -E "^github.com/plonk" | sort -k3 -nr

# Check coverage threshold
coverage-check threshold="80": test-all-coverage
    #!/usr/bin/env bash
    total=$(go tool cover -func=coverage/combined.out | tail -1 | awk '{print $3}' | sed 's/%//')
    echo "Total coverage: $total%"
    if (( $(echo "$total < {{threshold}}" | bc -l) )); then
        echo "❌ Coverage below {{threshold}}% threshold"
        exit 1
    else
        echo "✅ Coverage meets {{threshold}}% threshold"
    fi
```

### Coverage Report Format

The combined coverage report will show:
```
github.com/plonk/cmd/plonk/main.go:15:         main            85.7%
github.com/plonk/internal/commands/apply.go:45: runApply        92.3%
github.com/plonk/internal/commands/install.go:  runInstall      88.5%
...
github.com/plonk/internal/packages/npm.go:      NPM.Install     95.2%
github.com/plonk/internal/packages/brew.go:     Brew.Install    93.8%
...
total:                                          (statements)     82.4%
```

### Notes on Coverage Collection

1. **Binary Instrumentation**: The `-cover` flag instruments the binary for coverage
2. **GOCOVERDIR**: Environment variable tells the binary where to write coverage
3. **Coverage Format**: Go 1.20+ uses a binary format that's converted with `covdata`
4. **Merging**: `gocovmerge` combines multiple coverage files into one
5. **Performance Impact**: Coverage collection adds ~10-15% overhead

## Success Criteria

1. **Coverage**: Integration tests + unit tests achieve 80%+ total coverage
2. **Reliability**: < 5% flake rate across 100 runs
3. **Performance**: Full suite completes in < 10 minutes
4. **Safety**: Zero host system modifications ever
5. **Adoption**: Developers run tests locally before pushing
6. **Visibility**: Single combined coverage report showing all test contributions

## Timeline Summary

- **Week 1**: Foundation setup, base image, test helpers
- **Week 2**: Smoke tests, basic CLI operations
- **Week 3**: Package manager tests (all managers)
- **Week 4**: Dotfile management tests
- **Week 5**: Complete workflow tests
- **Week 6**: CI integration, documentation, handoff

Total: 6 weeks to full implementation
