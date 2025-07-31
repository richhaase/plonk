// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"
)

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		expectedManager string
		expectedPackage string
	}{
		{
			name:            "package without prefix",
			spec:            "htop",
			expectedManager: "",
			expectedPackage: "htop",
		},
		{
			name:            "package with brew prefix",
			spec:            "brew:git",
			expectedManager: "brew",
			expectedPackage: "git",
		},
		{
			name:            "package with npm prefix",
			spec:            "npm:lodash",
			expectedManager: "npm",
			expectedPackage: "lodash",
		},
		{
			name:            "go package with module path",
			spec:            "go:golang.org/x/tools/cmd/gopls",
			expectedManager: "go",
			expectedPackage: "golang.org/x/tools/cmd/gopls",
		},
		{
			name:            "empty prefix with colon",
			spec:            ":package",
			expectedManager: "",
			expectedPackage: "package",
		},
		{
			name:            "multiple colons in spec",
			spec:            "npm:@types/node:lts",
			expectedManager: "npm",
			expectedPackage: "@types/node:lts",
		},
		{
			name:            "empty package name",
			spec:            "brew:",
			expectedManager: "brew",
			expectedPackage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, packageName := ParsePackageSpec(tt.spec)
			if manager != tt.expectedManager {
				t.Errorf("ParsePackageSpec(%q) manager = %q, want %q", tt.spec, manager, tt.expectedManager)
			}
			if packageName != tt.expectedPackage {
				t.Errorf("ParsePackageSpec(%q) packageName = %q, want %q", tt.spec, packageName, tt.expectedPackage)
			}
		})
	}
}

func TestIsValidManager(t *testing.T) {
	tests := []struct {
		name     string
		manager  string
		expected bool
	}{
		{
			name:     "valid brew manager",
			manager:  "brew",
			expected: true,
		},
		{
			name:     "valid npm manager",
			manager:  "npm",
			expected: true,
		},
		{
			name:     "valid cargo manager",
			manager:  "cargo",
			expected: true,
		},
		{
			name:     "valid pip manager",
			manager:  "pip",
			expected: true,
		},
		{
			name:     "valid gem manager",
			manager:  "gem",
			expected: true,
		},
		{
			name:     "valid go manager",
			manager:  "go",
			expected: true,
		},
		{
			name:     "invalid manager",
			manager:  "invalid",
			expected: false,
		},
		{
			name:     "empty manager",
			manager:  "",
			expected: false,
		},
		{
			name:     "case sensitive check",
			manager:  "Brew",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidManager(tt.manager)
			if result != tt.expected {
				t.Errorf("IsValidManager(%q) = %v, want %v", tt.manager, result, tt.expected)
			}
		})
	}
}

func TestGetValidManagers(t *testing.T) {
	managers := GetValidManagers()

	// Should return all 6 supported managers
	expectedCount := 6
	if len(managers) != expectedCount {
		t.Errorf("GetValidManagers() returned %d managers, want %d", len(managers), expectedCount)
	}

	// Check that all expected managers are present
	expectedManagers := map[string]bool{
		"brew":  false,
		"npm":   false,
		"cargo": false,
		"pip":   false,
		"gem":   false,
		"go":    false,
	}

	for _, manager := range managers {
		if _, ok := expectedManagers[manager]; ok {
			expectedManagers[manager] = true
		} else {
			t.Errorf("GetValidManagers() returned unexpected manager: %s", manager)
		}
	}

	// Verify all expected managers were found
	for manager, found := range expectedManagers {
		if !found {
			t.Errorf("GetValidManagers() missing expected manager: %s", manager)
		}
	}
}
