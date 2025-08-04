# Plonk Integration Tests

This directory contains integration tests that validate plonk's behavior in a real environment using Docker containers.

## Overview

Integration tests run plonk commands inside Ubuntu containers to ensure:
- Commands work correctly with real package managers
- JSON/YAML output is properly formatted
- Side effects (installations, file creation) occur as expected
- No impact on the developer's machine

## Requirements

- Docker (Docker Desktop or Colima)
- Go 1.23+
- Linux binary built (`just build-linux`)

## Running Tests

```bash
# Build the Linux binary
just build-linux

# Build the test container
just build-test-image

# Run all integration tests
just test-integration

# Run specific test
go test -v -tags=integration ./tests/integration/... -run TestInstallPackage

# Run with timeout (recommended for full suite)
go test -v -tags=integration ./tests/integration/... -timeout 10m
```

## Test Files

Each command has its own test file:

- `container_test.go` - Test infrastructure and install test
- `uninstall_test.go` - Package uninstallation
- `status_test.go` - Status with multiple packages
- `list_test.go` - Package listing
- `search_test.go` - Package search
- `apply_test.go` - Apply packages and dotfiles
- `add_test.go` - Add dotfiles
- `rm_test.go` - Remove dotfiles
- `diff_test.go` - Diff dotfiles
- `clone_test.go` - Clone repositories
- `config_test.go` - Configuration management

## Test Infrastructure

### TestEnv

The `TestEnv` struct provides a containerized test environment:

```go
env := NewTestEnv(t)

// Run plonk command
output, err := env.Run("install", "brew:jq")

// Run with JSON output
var result struct{ ... }
err := env.RunJSON(&result, "status")

// Execute any command
output, err := env.Exec("brew", "list")

// Write files in container
err := env.WriteFile("/home/testuser/.testrc", []byte(content))
```

### Container Setup

The test container (`Dockerfile.integration`) includes:
- Ubuntu 22.04 base
- Homebrew pre-installed
- Test user with sudo access
- Environment variables set (NO_COLOR=1)

## Writing New Tests

Follow this pattern:

```go
//go:build integration

package integration_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCommandName(t *testing.T) {
    env := NewTestEnv(t)

    // Setup test data

    // Execute command
    var result struct {
        // Expected JSON fields
    }
    err := env.RunJSON(&result, "command", "args...")
    require.NoError(t, err)

    // Verify JSON output
    assert.Equal(t, expected, result.Field)

    // Verify side effects
    output, err := env.Exec("verify", "command")
    assert.Contains(t, output, "expected")
}
```

## Troubleshooting

### Docker not found
Ensure Docker is running (Docker Desktop or Colima).

### Container startup timeout
Check Docker resources and ensure the base image can be pulled.

### Test failures
Failed tests will print container logs. Check for:
- Missing dependencies
- Network issues
- Incorrect JSON structure expectations

### Slow tests
Integration tests take time (~15-20s per test) due to:
- Container startup
- Package installations
- Network operations

This is normal and expected.
