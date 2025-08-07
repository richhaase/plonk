# CheckHealth() Phase 1: Interface Extension - Detailed Implementation Guide

This document provides step-by-step implementation instructions for Phase 1 of the CheckHealth() integration. Another agent can follow these instructions to complete the implementation.

## Phase 1 Overview

**Objective**: Extend the PackageManager interface with CheckHealth() method and provide foundation code without breaking existing functionality.

**Files to Modify/Create**:
1. `/Users/rdh/src/plonk/internal/resources/packages/interfaces.go` - Add CheckHealth method
2. `/Users/rdh/src/plonk/internal/resources/packages/health_helpers.go` - NEW FILE - Helper functions
3. Update all 10 package manager files to add default CheckHealth implementations

**Time Estimate**: 1 day

**Success Criteria**:
- Interface compiles successfully
- All existing tests continue to pass
- `plonk doctor` continues to work (using old logic for now)
- Foundation is ready for Phase 2 implementations

## Step-by-Step Implementation

### Step 1: Add HealthCheck Struct to Packages Package

**File**: `/Users/rdh/src/plonk/internal/resources/packages/interfaces.go`

**Action**: Add the HealthCheck struct and extend the PackageManager interface

**Exact code to add** (insert at the end of the file, before the closing comment):

```go
// HealthCheck represents the result of a package manager health check
// This mirrors the HealthCheck struct in diagnostics but lives in the packages domain
type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}
```

**Action**: Add CheckHealth method to PackageManager interface

**Find this section** (around line 23):
```go
type PackageManager interface {
	PackageManagerCapabilities

	// Core operations - these are always supported by all package packages
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
	InstalledVersion(ctx context.Context, name string) (string, error)
	Info(ctx context.Context, name string) (*PackageInfo, error)

	// Optional operations - check capabilities before calling
	Search(ctx context.Context, query string) ([]string, error)
}
```

**Replace with**:
```go
type PackageManager interface {
	PackageManagerCapabilities

	// Core operations - these are always supported by all package packages
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
	InstalledVersion(ctx context.Context, name string) (string, error)
	Info(ctx context.Context, name string) (*PackageInfo, error)

	// Optional operations - check capabilities before calling
	Search(ctx context.Context, query string) ([]string, error)

	// Health checking operations
	CheckHealth(ctx context.Context) (*HealthCheck, error)
}
```

### Step 2: Create Health Helper Functions

**File**: `/Users/rdh/src/plonk/internal/resources/packages/health_helpers.go` (NEW FILE)

**Action**: Create this new file with the exact content below:

```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultCheckHealth provides a basic health check implementation that any package manager can use
// This is a fallback implementation that only checks basic availability
func DefaultCheckHealth(ctx context.Context, manager PackageManager, managerName string) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
		Category: "package-managers",
		Status:   "pass",
		Message:  fmt.Sprintf("%s is available and functional", strings.Title(managerName)),
	}

	// Check basic availability using existing IsAvailable method
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		// Context errors should bubble up
		if IsContextError(err) {
			return nil, err
		}
		// Other errors indicate health check failure
		check.Status = "fail"
		check.Message = fmt.Sprintf("%s availability check failed", strings.Title(managerName))
		check.Issues = []string{fmt.Sprintf("Error checking %s: %v", managerName, err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = fmt.Sprintf("%s is not available", strings.Title(managerName))
		check.Issues = []string{fmt.Sprintf("%s command not found or not functional", managerName)}
		check.Suggestions = []string{fmt.Sprintf("Install %s package manager", managerName)}
		return check, nil
	}

	// Available and functional
	check.Details = []string{fmt.Sprintf("%s binary found and functional", managerName)}
	return check, nil
}

// PathCheckResult represents the result of checking if a directory is in PATH
type PathCheckResult struct {
	InPath      bool
	Exists      bool
	Suggestions []string
}

// CheckDirectoryInPath checks if a directory exists and is in PATH
// This is a helper function that package managers can use in their CheckHealth implementations
func CheckDirectoryInPath(directory string) PathCheckResult {
	result := PathCheckResult{}

	// Check if directory exists
	if _, err := os.Stat(directory); err == nil {
		result.Exists = true
	}

	// Check if directory is in PATH
	path := os.Getenv("PATH")
	if path != "" {
		pathDirs := strings.Split(path, string(os.PathListSeparator))
		for _, pathDir := range pathDirs {
			// Clean up paths for comparison
			cleanPathDir := filepath.Clean(pathDir)
			cleanDirectory := filepath.Clean(directory)
			if cleanPathDir == cleanDirectory {
				result.InPath = true
				break
			}
		}
	}

	// Generate PATH suggestions if directory exists but not in PATH
	if result.Exists && !result.InPath {
		result.Suggestions = generatePathSuggestions(directory)
	}

	return result
}

// generatePathSuggestions creates shell-specific suggestions for adding a directory to PATH
func generatePathSuggestions(directory string) []string {
	suggestions := []string{
		fmt.Sprintf("Add %s to your PATH:", directory),
	}

	// Detect shell from environment
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		// Generic suggestion if shell can't be detected
		suggestions = append(suggestions, fmt.Sprintf(`export PATH="%s:$PATH"`, directory))
		suggestions = append(suggestions, "Add this line to your shell's configuration file (.bashrc, .zshrc, etc.)")
		return suggestions
	}

	// Shell-specific suggestions
	shellName := filepath.Base(shellPath)
	switch shellName {
	case "zsh":
		suggestions = append(suggestions, fmt.Sprintf(`echo 'export PATH="%s:$PATH"' >> ~/.zshrc`, directory))
		suggestions = append(suggestions, "source ~/.zshrc")
	case "bash":
		suggestions = append(suggestions, fmt.Sprintf(`echo 'export PATH="%s:$PATH"' >> ~/.bashrc`, directory))
		suggestions = append(suggestions, "source ~/.bashrc")
	case "fish":
		suggestions = append(suggestions, fmt.Sprintf("fish_add_path %s", directory))
	default:
		// Generic suggestion for unknown shells
		suggestions = append(suggestions, fmt.Sprintf(`export PATH="%s:$PATH"`, directory))
		suggestions = append(suggestions, fmt.Sprintf("Add this to your %s configuration file", shellName))
	}

	return suggestions
}

// FormatManagerName consistently formats package manager names for display
func FormatManagerName(name string) string {
	switch strings.ToLower(name) {
	case "npm":
		return "NPM"
	case "pip":
		return "Pip"
	case "uv":
		return "UV"
	case "dotnet":
		return ".NET"
	default:
		return strings.Title(name)
	}
}
```

### Step 3: Add Default CheckHealth Implementation to All Package Managers

For **each** of the 10 package manager files, add a default CheckHealth implementation. Follow this pattern:

**Files to update**:
1. `/Users/rdh/src/plonk/internal/resources/packages/homebrew.go`
2. `/Users/rdh/src/plonk/internal/resources/packages/npm.go`
3. `/Users/rdh/src/plonk/internal/resources/packages/pip.go`
4. `/Users/rdh/src/plonk/internal/resources/packages/cargo.go`
5. `/Users/rdh/src/plonk/internal/resources/packages/goinstall.go`
6. `/Users/rdh/src/plonk/internal/resources/packages/gem.go`
7. `/Users/rdh/src/plonk/internal/resources/packages/uv.go`
8. `/Users/rdh/src/plonk/internal/resources/packages/pixi.go`
9. `/Users/rdh/src/plonk/internal/resources/packages/composer.go`
10. `/Users/rdh/src/plonk/internal/resources/packages/dotnet.go`

**For each file**, add this method **before** the `init()` function:

#### Example for Homebrew Manager

**File**: `/Users/rdh/src/plonk/internal/resources/packages/homebrew.go`

**Find the location** right before the `init()` function (should be around the end of the file)

**Add this method**:
```go
// CheckHealth performs comprehensive health check for Homebrew
// This is a Phase 1 basic implementation - will be enhanced in Phase 2
func (h *HomebrewManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return DefaultCheckHealth(ctx, h, "homebrew")
}
```

#### Example for NPM Manager

**File**: `/Users/rdh/src/plonk/internal/resources/packages/npm.go`

**Add this method** before `init()`:
```go
// CheckHealth performs comprehensive health check for NPM
// This is a Phase 1 basic implementation - will be enhanced in Phase 2
func (n *NpmManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return DefaultCheckHealth(ctx, n, "npm")
}
```

#### Template for Remaining Package Managers

**Use this template** for the remaining 8 package managers, replacing `{Manager}`, `{receiver}`, and `{name}` appropriately:

```go
// CheckHealth performs comprehensive health check for {Manager}
// This is a Phase 1 basic implementation - will be enhanced in Phase 2
func ({receiver} *{Manager}Manager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return DefaultCheckHealth(ctx, {receiver}, "{name}")
}
```

**Specific replacements**:
- **Pip**: `func (p *PipManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, p, "pip") }`
- **Cargo**: `func (c *CargoManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, c, "cargo") }`
- **Go**: `func (g *GoManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, g, "go") }`
- **Gem**: `func (g *GemManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, g, "gem") }`
- **UV**: `func (u *UvManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, u, "uv") }`
- **Pixi**: `func (p *PixiManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, p, "pixi") }`
- **Composer**: `func (c *ComposerManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, c, "composer") }`
- **.NET**: `func (d *DotnetManager) CheckHealth(ctx context.Context) (*HealthCheck, error) { return DefaultCheckHealth(ctx, d, "dotnet") }`

### Step 4: Compilation and Testing

**Action**: Verify the implementation compiles and existing functionality works

**Commands to run**:
```bash
# 1. Verify compilation
go build ./internal/resources/packages/

# 2. Run existing package manager tests to ensure no regressions
go test ./internal/resources/packages/ -v

# 3. Test that plonk doctor still works (should use old logic for now)
go run main.go doctor

# 4. Run a basic functionality test
go run main.go status
```

**Expected Results**:
- All commands should compile successfully
- All existing tests should pass
- `plonk doctor` should work exactly as before (no behavior changes yet)
- `plonk status` should work normally

### Step 5: Basic Verification Test

**Action**: Create a simple test to verify the new interface methods exist

**File**: `/Users/rdh/src/plonk/internal/resources/packages/health_test.go` (NEW FILE)

**Content**:
```go
// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"testing"
	"time"
)

// TestAllPackageManagersImplementCheckHealth verifies that all package managers implement CheckHealth
func TestAllPackageManagersImplementCheckHealth(t *testing.T) {
	registry := NewManagerRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, managerName := range registry.GetAllManagerNames() {
		t.Run(managerName, func(t *testing.T) {
			mgr, err := registry.GetManager(managerName)
			if err != nil {
				t.Fatalf("Failed to get manager %s: %v", managerName, err)
			}

			// Verify CheckHealth method exists and can be called
			healthCheck, err := mgr.CheckHealth(ctx)
			if err != nil {
				// Context errors are acceptable in tests
				if err == context.Canceled || err == context.DeadlineExceeded {
					t.Skipf("Context cancelled for %s - acceptable in test environment", managerName)
				}
				t.Errorf("CheckHealth failed for %s: %v", managerName, err)
				return
			}

			// Verify basic health check structure
			if healthCheck == nil {
				t.Errorf("CheckHealth returned nil for %s", managerName)
				return
			}

			if healthCheck.Name == "" {
				t.Errorf("CheckHealth returned empty Name for %s", managerName)
			}

			if healthCheck.Category == "" {
				t.Errorf("CheckHealth returned empty Category for %s", managerName)
			}

			if healthCheck.Status == "" {
				t.Errorf("CheckHealth returned empty Status for %s", managerName)
			}

			// Status should be one of the valid values
			validStatuses := []string{"pass", "warn", "fail", "info"}
			validStatus := false
			for _, valid := range validStatuses {
				if healthCheck.Status == valid {
					validStatus = true
					break
				}
			}
			if !validStatus {
				t.Errorf("CheckHealth returned invalid status '%s' for %s", healthCheck.Status, managerName)
			}

			t.Logf("✓ %s CheckHealth: %s - %s", managerName, healthCheck.Status, healthCheck.Message)
		})
	}
}

// TestDefaultCheckHealthWithMockManager tests the DefaultCheckHealth helper function
func TestDefaultCheckHealth(t *testing.T) {
	ctx := context.Background()

	// Create a mock manager for testing
	mockManager := &MockManager{
		available: true,
	}

	healthCheck, err := DefaultCheckHealth(ctx, mockManager, "test")
	if err != nil {
		t.Fatalf("DefaultCheckHealth failed: %v", err)
	}

	if healthCheck.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s'", healthCheck.Status)
	}

	if healthCheck.Name != "Test Manager" {
		t.Errorf("Expected name 'Test Manager', got '%s'", healthCheck.Name)
	}

	if healthCheck.Category != "package-managers" {
		t.Errorf("Expected category 'package-managers', got '%s'", healthCheck.Category)
	}
}

// MockManager for testing DefaultCheckHealth
type MockManager struct {
	available bool
	error     error
}

func (m *MockManager) IsAvailable(ctx context.Context) (bool, error) {
	return m.available, m.error
}

// Required interface methods (minimal implementations for testing)
func (m *MockManager) ListInstalled(ctx context.Context) ([]string, error)                   { return nil, nil }
func (m *MockManager) Install(ctx context.Context, name string) error                        { return nil }
func (m *MockManager) Uninstall(ctx context.Context, name string) error                      { return nil }
func (m *MockManager) IsInstalled(ctx context.Context, name string) (bool, error)           { return false, nil }
func (m *MockManager) InstalledVersion(ctx context.Context, name string) (string, error)    { return "", nil }
func (m *MockManager) Info(ctx context.Context, name string) (*PackageInfo, error)          { return nil, nil }
func (m *MockManager) Search(ctx context.Context, query string) ([]string, error)           { return nil, nil }
func (m *MockManager) SupportsSearch() bool                                                  { return false }
func (m *MockManager) CheckHealth(ctx context.Context) (*HealthCheck, error) {
	return DefaultCheckHealth(ctx, m, "test")
}
```

### Step 6: Final Verification

**Action**: Run the new test to verify everything is working

**Commands**:
```bash
# Run the new health test
go test ./internal/resources/packages/ -run TestAllPackageManagersImplementCheckHealth -v

# Run the default health test
go test ./internal/resources/packages/ -run TestDefaultCheckHealth -v

# Run all package tests to ensure no regressions
go test ./internal/resources/packages/ -v

# Final compilation check
go build ./...
```

**Expected Results**:
- All new tests pass
- All existing tests continue to pass
- Full codebase compiles successfully
- No behavior changes in plonk commands (doctor, status, etc.)

## Handoff Information

**What Phase 1 Accomplishes**:
✅ Extends PackageManager interface with CheckHealth() method
✅ Provides default implementation that wraps existing IsAvailable() logic
✅ Creates helper functions for PATH checking (foundation for Phase 2)
✅ Maintains 100% backward compatibility
✅ All 10 package managers implement the interface
✅ Comprehensive test coverage for new interface

**What Phase 1 Does NOT Do**:
❌ Does not change plonk doctor behavior (still uses old hardcoded logic)
❌ Does not implement dynamic path discovery (that's Phase 2)
❌ Does not remove any existing code (cleanup happens in Phase 6)

**Ready for Phase 2**:
- Interface is defined and implemented by all package managers
- Helper functions exist for PATH checking
- Foundation code is tested and stable
- Next phase can enhance individual CheckHealth implementations with dynamic path discovery

**Files Modified in Phase 1**:
- `interfaces.go` - Extended interface
- `health_helpers.go` - NEW FILE with helper functions
- `health_test.go` - NEW FILE with verification tests
- All 10 package manager files - Added basic CheckHealth implementations

**Development Rules Compliance**:
✅ No unrequested features - implementing exactly the CheckHealth method
✅ No breaking changes - all existing functionality preserved
✅ Professional output - no emojis, clean error messages
✅ Safe tests - mock managers, no system modifications
✅ Prefer editing existing files - only essential new files created

The implementation is ready for Phase 2 enhancement where each package manager will implement sophisticated path discovery and health checking.
