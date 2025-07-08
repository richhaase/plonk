// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"fmt"
	"testing"
)

// MockPackageManager for testing
type MockPackageManager struct {
	available bool
	packages  []string
}

func (m *MockPackageManager) IsAvailable() bool {
	return m.available
}

func (m *MockPackageManager) ListInstalled() ([]string, error) {
	return m.packages, nil
}

func (m *MockPackageManager) Install(name string, version string) error {
	// Mock implementation - just add to packages list
	m.packages = append(m.packages, name)
	return nil
}

func (m *MockPackageManager) Uninstall(name string) error {
	// Mock implementation - remove from packages list
	for i, pkg := range m.packages {
		if pkg == name {
			m.packages = append(m.packages[:i], m.packages[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockPackageManager) IsInstalled(name string) bool {
	for _, pkg := range m.packages {
		if pkg == name {
			return true
		}
	}
	return false
}

func (m *MockPackageManager) GetVersion(name string) (string, error) {
	if m.IsInstalled(name) {
		return "1.0.0", nil
	}
	return "", fmt.Errorf("package %s not installed", name)
}

// MockConfigLoader for testing
type MockConfigLoader struct {
	packages map[string][]ConfigPackage
}

func (m *MockConfigLoader) GetPackagesForManager(managerName string) ([]ConfigPackage, error) {
	if packages, exists := m.packages[managerName]; exists {
		return packages, nil
	}
	return []ConfigPackage{}, nil
}

func TestStateReconciler_ReconcileManager(t *testing.T) {
	// Setup test data
	mockManager := &MockPackageManager{
		available: true,
		packages:  []string{"git", "curl", "vim"},
	}
	
	mockLoader := &MockConfigLoader{
		packages: map[string][]ConfigPackage{
			"test": {
				{Name: "git", Version: ""},
				{Name: "curl", Version: ""},
				{Name: "missing-package", Version: ""},
			},
		},
	}
	
	mockChecker := &HomebrewVersionChecker{} // Always returns true
	
	managers := map[string]PackageManager{
		"test": mockManager,
	}
	
	checkers := map[string]VersionChecker{
		"test": mockChecker,
	}
	
	reconciler := NewStateReconciler(mockLoader, managers, checkers)
	
	// Test reconciliation
	result, err := reconciler.ReconcileManager("test")
	if err != nil {
		t.Fatalf("ReconcileManager failed: %v", err)
	}
	
	// Verify results
	if len(result.Managed) != 2 {
		t.Errorf("Expected 2 managed packages, got %d", len(result.Managed))
	}
	
	if len(result.Missing) != 1 {
		t.Errorf("Expected 1 missing package, got %d", len(result.Missing))
	}
	
	if len(result.Untracked) != 1 {
		t.Errorf("Expected 1 untracked package, got %d", len(result.Untracked))
	}
	
	// Check specific packages
	managedNames := make([]string, len(result.Managed))
	for i, pkg := range result.Managed {
		managedNames[i] = pkg.Name
	}
	
	expectedManaged := []string{"git", "curl"}
	for _, expected := range expectedManaged {
		found := false
		for _, actual := range managedNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected managed package %s not found", expected)
		}
	}
}

func TestVersionCheckers(t *testing.T) {
	testCases := []struct {
		name     string
		checker  VersionChecker
		config   ConfigPackage
		installed string
		expected bool
	}{
		{
			name:     "Homebrew ignores version",
			checker:  &HomebrewVersionChecker{},
			config:   ConfigPackage{Name: "git", Version: "2.30.0"},
			installed: "2.45.0",
			expected: true,
		},
		{
			name:     "NPM exact version match",
			checker:  &NpmVersionChecker{},
			config:   ConfigPackage{Name: "typescript", Version: "5.0.0"},
			installed: "5.0.0",
			expected: true,
		},
		{
			name:     "NPM version mismatch",
			checker:  &NpmVersionChecker{},
			config:   ConfigPackage{Name: "typescript", Version: "5.0.0"},
			installed: "4.9.0",
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.checker.CheckVersion(tc.config, tc.installed)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}