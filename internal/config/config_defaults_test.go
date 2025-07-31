// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    *Config
		expected *Config
	}{
		{
			name:  "empty config gets all defaults",
			input: &Config{},
			expected: &Config{
				DefaultManager:    "brew",
				OperationTimeout:  300, // 5 minutes
				PackageTimeout:    180, // 3 minutes
				DotfileTimeout:    60,  // 1 minute
				ExpandDirectories: []string{".config"},
				IgnorePatterns:    defaultConfig.IgnorePatterns,
			},
		},
		{
			name: "partial config preserves set values",
			input: &Config{
				DefaultManager:   "npm",
				OperationTimeout: 60, // 1 minute
			},
			expected: &Config{
				DefaultManager:    "npm",
				OperationTimeout:  60,  // 1 minute
				PackageTimeout:    180, // 3 minutes
				DotfileTimeout:    60,  // 1 minute
				ExpandDirectories: []string{".config"},
				IgnorePatterns:    defaultConfig.IgnorePatterns,
			},
		},
		{
			name: "custom patterns preserved",
			input: &Config{
				IgnorePatterns:    []string{"custom.txt"},
				ExpandDirectories: []string{"custom-dir"},
			},
			expected: &Config{
				DefaultManager:    "brew",
				OperationTimeout:  300, // 5 minutes
				PackageTimeout:    180, // 3 minutes
				DotfileTimeout:    60,  // 1 minute
				ExpandDirectories: []string{"custom-dir"},
				IgnorePatterns:    []string{"custom.txt"},
			},
		},
		{
			name: "all values set preserves everything",
			input: &Config{
				DefaultManager:    "cargo",
				OperationTimeout:  300,  // 5 minutes
				PackageTimeout:    1200, // 20 minutes
				DotfileTimeout:    60,   // 1 minute
				ExpandDirectories: []string{"my-config"},
				IgnorePatterns:    []string{"*.bak"},
			},
			expected: &Config{
				DefaultManager:    "cargo",
				OperationTimeout:  300,  // 5 minutes
				PackageTimeout:    1200, // 20 minutes
				DotfileTimeout:    60,   // 1 minute
				ExpandDirectories: []string{"my-config"},
				IgnorePatterns:    []string{"*.bak"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyDefaults(tt.input)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestGetDefaultConfigDirectory(t *testing.T) {
	// This test verifies the function exists and returns a non-empty path
	result := GetDefaultConfigDirectory()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, ".config/plonk")
}
