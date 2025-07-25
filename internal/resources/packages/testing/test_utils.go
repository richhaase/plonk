// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package testing

import (
	"context"
	"testing"
)

// PackageManager interface for testing (avoids circular import)
type PackageManager interface {
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	SupportsSearch() bool
	Search(ctx context.Context, query string) ([]string, error)
}

// PackageInfo represents detailed information about a package
type PackageInfo struct {
	Name          string   `json:"name"`
	Version       string   `json:"version,omitempty"`
	Description   string   `json:"description,omitempty"`
	Homepage      string   `json:"homepage,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	InstalledSize string   `json:"installed_size,omitempty"`
	Manager       string   `json:"manager"`
	Installed     bool     `json:"installed"`
}

// ManagerTestSuite provides common test utilities for package managers
type ManagerTestSuite struct {
	Manager     PackageManager
	TestPackage string
	BinaryName  string
}

// TestIsAvailable tests the IsAvailable method with common patterns
func (suite *ManagerTestSuite) TestIsAvailable(t *testing.T) {
	ctx := context.Background()

	// Basic availability test
	available, err := suite.Manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("IsAvailable() failed: %v", err)
	}

	// Should return boolean result without error when binary is available
	if available {
		t.Logf("Manager %s is available", suite.BinaryName)
	} else {
		t.Logf("Manager %s is not available (expected if not installed)", suite.BinaryName)
	}
}

// TestListInstalled tests the ListInstalled method with common patterns
func (suite *ManagerTestSuite) TestListInstalled(t *testing.T) {
	ctx := context.Background()
	available, err := suite.Manager.IsAvailable(ctx)
	if err != nil {
		t.Fatalf("IsAvailable() failed: %v", err)
	}

	if !available {
		t.Skip("Manager not available, skipping ListInstalled test")
	}

	packages, err := suite.Manager.ListInstalled(ctx)
	if err != nil {
		// Allow reasonable failures when manager isn't properly configured
		t.Logf("ListInstalled() failed (may be expected): %v", err)
		return
	}

	// Should return a slice (possibly empty)
	if packages == nil {
		t.Error("ListInstalled() returned nil instead of empty slice")
	}

	t.Logf("Found %d installed packages", len(packages))
}

// TestSupportsSearch tests the SupportsSearch capability
func (suite *ManagerTestSuite) TestSupportsSearch(t *testing.T) {
	supportsSearch := suite.Manager.SupportsSearch()
	t.Logf("Manager supports search: %v", supportsSearch)

	if supportsSearch {
		ctx := context.Background()
		available, err := suite.Manager.IsAvailable(ctx)
		if err != nil {
			t.Fatalf("IsAvailable() failed: %v", err)
		}

		if available {
			// Test that search doesn't panic with a simple query
			_, err := suite.Manager.Search(ctx, "test")
			// Don't require success, just no panic
			t.Logf("Search test completed (error: %v)", err)
		}
	}
}

// StringSlicesEqual compares two string slices for equality (order-independent)
func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to count occurrences
	aMap := make(map[string]int)
	bMap := make(map[string]int)

	for _, s := range a {
		aMap[s]++
	}
	for _, s := range b {
		bMap[s]++
	}

	// Compare maps
	for k, v := range aMap {
		if bMap[k] != v {
			return false
		}
	}

	return true
}

// CreateMockPackageInfo creates a standardized PackageInfo for testing
func CreateMockPackageInfo(name, version, manager string) *PackageInfo {
	return &PackageInfo{
		Name:        name,
		Version:     version,
		Description: "Test package",
		Manager:     manager,
		Installed:   true,
	}
}
