# Integration Testing Proof-of-Concept Plan

**Created**: 2025-08-04
**Status**: COMPLETED
**Target**: Immediate implementation
**Estimated Duration**: 1 week

## DETOUR: Fix JSON/YAML Output Bug (2025-08-04)

**Issue Found**: Plonk is outputting ANSI color codes/terminal control characters even when using `-o json` or `-o yaml`. This breaks JSON parsing in tests.

**Example**: The output contains `\x01` and other control characters mixed with JSON:
```
Installing: brew:hello
      ✓ added hello
      �{
  "command": "install",
  ...
```

**Action Required**:
1. Fix plonk to ensure NO color output when using structured output formats
2. The `-o json` and `-o yaml` flags should force NO_COLOR behavior
3. This is a prerequisite for reliable integration testing

**Status**: FIXED - All progress/status messages now go to stderr, JSON/YAML output is clean

## Overview

This document outlines a minimal proof-of-concept for containerized integration testing using testcontainers-go. The goal is to establish the infrastructure and validate the approach with a single working test before expanding.

## Goals

1. **Prove the concept** with minimal complexity
2. **Establish safety** through complete isolation
3. **Validate speed** for developer experience
4. **Test efficiency** by focusing on user-facing behavior

## Constraints

- **Safe**: No impact to developer environments (Docker-only on dev machines)
- **Fast**: Single test completes in <30 seconds
- **Efficient**: Test real user behavior with JSON/YAML validation
- **Minimal**: <500 lines of code for entire PoC

## Key Design Decisions

1. **Use testcontainers-go from start** - Avoid migration pain, idiomatic Go approach
2. **Cross-compile for Linux** - Build plonk-linux binary for Docker container
3. **JSON output validation** - More reliable than parsing human-readable output
4. **Docker-only on dev machines** - CI can run directly for speed
5. **Verify via Homebrew** - Check `brew list` to ensure actual installation

## Phase 1: Infrastructure Setup (Day 1-2)

### 1.1 Directory Structure
```
internal/
└── integration/
    ├── Dockerfile
    ├── test_env.go      # Testcontainers wrapper
    ├── basic_test.go    # PoC test
    └── README.md        # Setup instructions
```

### 1.2 Minimal Test Container
```dockerfile
# internal/integration/Dockerfile
FROM ubuntu:22.04

# Minimal dependencies for PoC
RUN apt-get update && apt-get install -y \
    curl \
    git \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Create test user
RUN useradd -m -s /bin/bash testuser && \
    echo 'testuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER testuser
WORKDIR /home/testuser

# Install Homebrew (Linux)
RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
ENV PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"

# Set environment for faster operations
ENV NO_COLOR=1
ENV HOMEBREW_NO_ANALYTICS=1

# Pre-install test packages for speed
RUN brew install tree jq

# Signal that container is ready
RUN echo "Container ready" > /tmp/ready.txt

CMD ["tail", "-f", "/dev/null"]
```

### 1.3 Test Environment with Testcontainers
```go
// internal/integration/test_env.go
package integration

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

type TestEnv struct {
    t         *testing.T
    container testcontainers.Container
    ctx       context.Context
}

func NewTestEnv(t *testing.T) *TestEnv {
    t.Helper()

    ctx := context.Background()

    // Check if we're in CI - run directly without Docker
    if os.Getenv("CI") == "true" {
        t.Skip("CI mode - would run directly without Docker")
        // TODO: Implement CI mode in future
    }

    // Ensure plonk-linux binary exists
    if _, err := os.Stat("./plonk-linux"); os.IsNotExist(err) {
        t.Fatal("Linux binary not found. Run: just build-linux")
    }

    req := testcontainers.ContainerRequest{
        Image: "plonk-test:poc",
        Env: map[string]string{
            "NO_COLOR": "1",
            "HOMEBREW_NO_ANALYTICS": "1",
        },
        Files: []testcontainers.ContainerFile{{
            HostFilePath:      "./plonk-linux",
            ContainerFilePath: "/usr/local/bin/plonk",
            FileMode:         0755,
        }},
        WaitingFor: wait.ForExec([]string{"test", "-f", "/tmp/ready.txt"}).
            WithStartupTimeout(30 * time.Second),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        t.Fatalf("Failed to start container: %v", err)
    }

    // Ensure cleanup
    t.Cleanup(func() {
        if t.Failed() {
            // Get container logs on failure
            logs, _ := container.Logs(ctx)
            t.Logf("Container logs:\n%s", logs)
        }
        container.Terminate(ctx)
    })

    return &TestEnv{
        t:         t,
        container: container,
        ctx:       ctx,
    }
}

// Run executes plonk command in container
func (e *TestEnv) Run(args ...string) (string, error) {
    return e.Exec("plonk", args...)
}

// Exec runs any command in container
func (e *TestEnv) Exec(cmd string, args ...string) (string, error) {
    e.t.Helper()

    exitCode, output, err := e.container.Exec(e.ctx, append([]string{cmd}, args...))
    if err != nil {
        return "", fmt.Errorf("exec failed: %w", err)
    }

    outputStr := string(output)

    if exitCode != 0 {
        return outputStr, fmt.Errorf("command failed with exit code %d", exitCode)
    }

    return outputStr, nil
}

// RunJSON runs plonk command and parses JSON output
func (e *TestEnv) RunJSON(v interface{}, args ...string) error {
    // Append -o json to args
    args = append(args, "-o", "json")

    output, err := e.Run(args...)
    if err != nil {
        return fmt.Errorf("command failed: %w\nOutput: %s", err, output)
    }

    if err := json.Unmarshal([]byte(output), v); err != nil {
        return fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
    }

    return nil
}
```

### 1.4 Single PoC Test
```go
// internal/integration/basic_test.go
//go:build integration

package integration

import (
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestInstallPackage(t *testing.T) {
    env := NewTestEnv(t)

    // Test: Install a package with JSON output
    var installResult struct {
        TotalPackages int `json:"total_packages"`
        Results []struct {
            Name   string `json:"name"`
            Status string `json:"status"`
        } `json:"results"`
    }

    err := env.RunJSON(&installResult, "install", "brew:tree")
    require.NoError(t, err, "Install should succeed")

    // Verify installation result
    assert.Equal(t, 1, installResult.TotalPackages)
    assert.Equal(t, "tree", installResult.Results[0].Name)
    assert.Equal(t, "installed", installResult.Results[0].Status)

    // Test: Verify status shows package
    var statusResult struct {
        Packages []struct {
            Name    string `json:"name"`
            Manager string `json:"manager"`
        } `json:"packages"`
    }

    err = env.RunJSON(&statusResult, "status")
    require.NoError(t, err, "Status should succeed")

    // Find tree in packages
    found := false
    for _, pkg := range statusResult.Packages {
        if pkg.Name == "tree" && pkg.Manager == "brew" {
            found = true
            break
        }
    }
    assert.True(t, found, "Tree should be in status output")

    // Test: Verify brew actually installed it
    brewOut, err := env.Exec("brew", "list")
    require.NoError(t, err, "Brew list should work")
    assert.Contains(t, brewOut, "tree", "Tree should be in brew list")

    // Test: Verify the binary works
    treeOut, err := env.Exec("tree", "--version")
    require.NoError(t, err, "Tree command should work")
    assert.Contains(t, treeOut, "tree v", "Tree should report version")
}
```

## Phase 2: Build and Execution (Day 3-4)

### 2.1 Build Configuration
```justfile
# Build Linux binary for Docker container
build-linux:
    @echo "🔨 Building Linux binary for Docker..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plonk-linux cmd/plonk/main.go

# Build test container
build-test-image:
    @echo "🐳 Building test container..."
    docker build -t plonk-test:poc -f internal/integration/Dockerfile .

# Run integration tests (Docker-only on dev machines)
test-integration: build-linux build-test-image
    @echo "🧪 Running integration tests in Docker..."
    @if [ -z "$$CI" ]; then \
        echo "   Using Docker for safety (required on dev machines)"; \
        go test -v -tags=integration ./internal/integration/...; \
    else \
        echo "   CI mode - would run directly"; \
        go build -o plonk cmd/plonk/main.go && \
        go test -v -tags=integration ./internal/integration/...; \
    fi

# Quick verification of Linux binary
verify-linux-binary: build-linux
    @echo "✓ Testing Linux binary in Docker..."
    @docker run --rm \
        -v $$PWD/plonk-linux:/plonk \
        ubuntu:22.04 \
        /plonk --version || \
        (echo "❌ Linux binary failed" && exit 1)
```

### 2.2 TestMain Setup
```go
// internal/integration/main_test.go
//go:build integration

package integration

import (
    "os"
    "os/exec"
    "testing"
)

func TestMain(m *testing.M) {
    // Require explicit opt-in
    if os.Getenv("PLONK_INTEGRATION") == "" && os.Getenv("CI") == "" {
        os.Exit(0)
    }

    // Verify Docker is available (for dev machines)
    if os.Getenv("CI") == "" {
        if err := exec.Command("docker", "version").Run(); err != nil {
            panic("Docker is required for integration tests")
        }
    }

    os.Exit(m.Run())
}
```

## Phase 3: Validation (Day 5)

### 3.1 Success Criteria

- [ ] Test runs in Docker container via testcontainers-go
- [ ] Test completes in <30 seconds
- [ ] Test validates JSON output structure
- [ ] Test verifies actual package installation via `brew list`
- [ ] Test confirms installed binary works
- [ ] No modifications to host system
- [ ] Clear pass/fail output with logs on failure

### 3.2 Performance Metrics

Measure and document:
- Container image build time (one-time)
- Container startup time (<5s expected)
- Test execution time (<20s expected)
- Total end-to-end time (<30s target)

### 3.3 Developer Experience

Validate:
- Single command: `just test-integration`
- Clear error when Docker missing
- Helpful error when plonk-linux missing
- Container logs on test failure
- Easy debugging with testcontainers

## Next Steps (After PoC Success)

If PoC succeeds, expand gradually:

1. **Add core command tests**:
   - `uninstall` with verification
   - `status` with multiple packages
   - `apply` with packages

2. **Add dotfile test**:
   - `add` dotfile
   - `apply --dotfiles-only`
   - Verify symlink creation

3. **Extract common patterns**:
   - JSON output parsing helpers
   - Common assertions
   - Test data fixtures

4. **Add CI integration**:
   - GitHub Actions workflow
   - Coverage collection
   - Test result reporting

5. **Performance optimization**:
   - Pre-built base image
   - Layer caching
   - Parallel test execution

## What We're NOT Doing (PoC Scope)

- ❌ Multiple package managers (just brew)
- ❌ Coverage collection
- ❌ Parallel execution
- ❌ Error scenarios
- ❌ Network resilience
- ❌ Multiple platforms
- ❌ Complex test patterns

## Implementation Checklist

- [x] Add testcontainers-go to go.mod
- [x] Create tests/integration directory (changed from internal/integration)
- [x] Write Dockerfile.integration
- [x] Implement container_test.go with testcontainers
- [x] Write TestInstallPackage with JSON validation
- [x] Update Justfile with build commands
- [ ] Create README.md with setup instructions
- [x] Test the PoC end-to-end
- [ ] Document results and timings
- [ ] Plan expansion based on learnings

## Implementation Notes (2025-08-04)

### Changes from Original Plan
1. Used `tests/integration` instead of `internal/integration` for test location
2. Combined test_env.go functionality into container_test.go
3. Fixed JSON structure mismatch - status returns `managed_items` not `packages`
4. Fixed test expectation - install returns status "added" not "installed"
5. Used `hello` package instead of `tree` (simpler, works on all architectures)

### Current State (Completed 2025-08-04)
- Full integration test suite implemented
- All commands have happy path tests:
  - Package management: install, uninstall, list, search, status
  - Dotfile management: add, rm, diff
  - Orchestration: apply, clone
  - Configuration: config show
- Infrastructure features:
  - Container file writing via WriteFile helper
  - JSON output validation
  - System state verification
  - Real package operations
