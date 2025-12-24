// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePackageSpec(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *PackageSpec
		expectError bool
		errorMsg    string
	}{
		{
			name:  "simple package name",
			input: "git",
			expected: &PackageSpec{
				Name:         "git",
				Manager:      "",
				OriginalSpec: "git",
			},
		},
		{
			name:  "brew package",
			input: "brew:wget",
			expected: &PackageSpec{
				Name:         "wget",
				Manager:      "brew",
				OriginalSpec: "brew:wget",
			},
		},
		{
			name:  "npm scoped package",
			input: "npm:@types/node",
			expected: &PackageSpec{
				Name:         "@types/node",
				Manager:      "npm",
				OriginalSpec: "npm:@types/node",
			},
		},
		{
			name:        "empty specification",
			input:       "",
			expectError: true,
			errorMsg:    "package specification cannot be empty",
		},
		{
			name:        "empty manager prefix",
			input:       ":package",
			expectError: true,
			errorMsg:    "manager prefix cannot be empty",
		},
		{
			name:        "empty package name with manager",
			input:       "brew:",
			expectError: true,
			errorMsg:    "package name cannot be empty",
		},
		{
			name:  "multiple colons",
			input: "manager:pkg:extra",
			expected: &PackageSpec{
				Name:         "pkg:extra",
				Manager:      "manager",
				OriginalSpec: "manager:pkg:extra",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePackageSpec(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Manager, result.Manager)
				assert.Equal(t, tt.expected.OriginalSpec, result.OriginalSpec)
			}
		})
	}
}

func TestPackageSpec_ValidateManager(t *testing.T) {
	tests := []struct {
		name        string
		spec        *PackageSpec
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty manager is valid",
			spec: &PackageSpec{
				Name:    "git",
				Manager: "",
			},
			expectError: false,
		},
		{
			name: "brew manager is valid",
			spec: &PackageSpec{
				Name:    "wget",
				Manager: "brew",
			},
			expectError: false,
		},
		{
			name: "npm manager is valid",
			spec: &PackageSpec{
				Name:    "prettier",
				Manager: "npm",
			},
			expectError: false,
		},
		{
			name: "invalid manager",
			spec: &PackageSpec{
				Name:    "package",
				Manager: "invalid",
			},
			expectError: true,
			errorMsg:    `unknown package manager "invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.ValidateManager()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPackageSpec_ResolveManager(t *testing.T) {
	tests := []struct {
		name            string
		spec            *PackageSpec
		defaultManager  string
		expectedManager string
	}{
		{
			name: "manager already set",
			spec: &PackageSpec{
				Name:    "wget",
				Manager: "brew",
			},
			defaultManager:  "npm",
			expectedManager: "brew",
		},
		{
			name: "no manager, use default",
			spec: &PackageSpec{
				Name:    "git",
				Manager: "",
			},
			defaultManager:  "brew",
			expectedManager: "brew",
		},
		{
			name: "no manager, no default",
			spec: &PackageSpec{
				Name:    "git",
				Manager: "",
			},
			defaultManager:  "",
			expectedManager: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.spec.ResolveManager(tt.defaultManager)
			assert.Equal(t, tt.expectedManager, tt.spec.Manager)
		})
	}
}

func TestPackageSpec_RequireManager(t *testing.T) {
	tests := []struct {
		name            string
		spec            *PackageSpec
		defaultManager  string
		expectError     bool
		errorMsg        string
		expectedManager string
	}{
		{
			name: "manager already set and valid",
			spec: &PackageSpec{
				Name:    "wget",
				Manager: "brew",
			},
			defaultManager:  "npm",
			expectError:     false,
			expectedManager: "brew",
		},
		{
			name: "no manager, use valid default",
			spec: &PackageSpec{
				Name:    "prettier",
				Manager: "",
			},
			defaultManager:  "npm",
			expectError:     false,
			expectedManager: "npm",
		},
		{
			name: "no manager, no default",
			spec: &PackageSpec{
				Name:    "git",
				Manager: "",
			},
			defaultManager: "",
			expectError:    true,
			errorMsg:       "no package manager specified and no default configured",
		},
		{
			name: "invalid manager",
			spec: &PackageSpec{
				Name:    "package",
				Manager: "invalid",
			},
			defaultManager: "brew",
			expectError:    true,
			errorMsg:       `unknown package manager "invalid"`,
		},
		{
			name: "no manager, invalid default",
			spec: &PackageSpec{
				Name:    "git",
				Manager: "",
			},
			defaultManager: "invalid",
			expectError:    true,
			errorMsg:       `unknown package manager "invalid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.RequireManager(tt.defaultManager)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedManager, tt.spec.Manager)
			}
		})
	}
}

func TestPackageSpec_String(t *testing.T) {
	tests := []struct {
		name     string
		spec     *PackageSpec
		expected string
	}{
		{
			name: "package without manager",
			spec: &PackageSpec{
				Name:    "git",
				Manager: "",
			},
			expected: "git",
		},
		{
			name: "package with manager",
			spec: &PackageSpec{
				Name:    "wget",
				Manager: "brew",
			},
			expected: "brew:wget",
		},
		{
			name: "scoped npm package",
			spec: &PackageSpec{
				Name:    "@types/node",
				Manager: "npm",
			},
			expected: "npm:@types/node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSpecs_RejectsUnknownManager(t *testing.T) {
	// Custom managers are no longer supported - only hardcoded managers work
	result := ValidateSpecs([]string{"custom:tool"}, ValidationModeInstall, "")
	require.Len(t, result.Invalid, 1)
	assert.Contains(t, result.Invalid[0].Error.Error(), "unknown package manager")
}
