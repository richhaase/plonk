// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestGetApplyScope(t *testing.T) {
	tests := []struct {
		name         string
		packagesOnly bool
		dotfilesOnly bool
		expected     string
	}{
		{
			name:         "packages only",
			packagesOnly: true,
			dotfilesOnly: false,
			expected:     "packages",
		},
		{
			name:         "dotfiles only",
			packagesOnly: false,
			dotfilesOnly: true,
			expected:     "dotfiles",
		},
		{
			name:         "neither flag set returns all",
			packagesOnly: false,
			dotfilesOnly: false,
			expected:     "all",
		},
		{
			name:         "both flags set (packages takes precedence)",
			packagesOnly: true,
			dotfilesOnly: true,
			expected:     "packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getApplyScope(tt.packagesOnly, tt.dotfilesOnly)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertApplyResult(t *testing.T) {
	tests := []struct {
		name     string
		result   orchestrator.ApplyResult
		scope    string
		expected CombinedApplyOutput
	}{
		{
			name: "successful apply with packages",
			result: orchestrator.ApplyResult{
				DryRun:  false,
				Success: true,
				Packages: map[string]interface{}{
					"installed": 3,
					"removed":   1,
				},
				Dotfiles: nil,
				Error:    "",
			},
			scope: "packages",
			expected: CombinedApplyOutput{
				DryRun:  false,
				Success: true,
				Packages: map[string]interface{}{
					"installed": 3,
					"removed":   1,
				},
				Dotfiles:      nil,
				Scope:         "packages",
				PackageErrors: nil,
				DotfileErrors: nil,
			},
		},
		{
			name: "dry run with dotfiles",
			result: orchestrator.ApplyResult{
				DryRun:   true,
				Success:  true,
				Packages: nil,
				Dotfiles: map[string]interface{}{
					"added":   5,
					"updated": 2,
				},
			},
			scope: "dotfiles",
			expected: CombinedApplyOutput{
				DryRun:   true,
				Success:  true,
				Packages: nil,
				Dotfiles: map[string]interface{}{
					"added":   5,
					"updated": 2,
				},
				Scope:         "dotfiles",
				PackageErrors: nil,
				DotfileErrors: nil,
			},
		},
		{
			name: "apply with errors",
			result: orchestrator.ApplyResult{
				DryRun:        false,
				Success:       false,
				Packages:      map[string]interface{}{"installed": 1},
				Dotfiles:      map[string]interface{}{"added": 1},
				PackageErrors: []string{"failed to install vim"},
				DotfileErrors: []string{"permission denied: ~/.zshrc"},
			},
			scope: "all",
			expected: CombinedApplyOutput{
				DryRun:        false,
				Success:       false,
				Packages:      map[string]interface{}{"installed": 1},
				Dotfiles:      map[string]interface{}{"added": 1},
				Scope:         "all",
				PackageErrors: []string{"failed to install vim"},
				DotfileErrors: []string{"permission denied: ~/.zshrc"},
			},
		},
		{
			name: "empty result",
			result: orchestrator.ApplyResult{
				DryRun:   false,
				Success:  true,
				Packages: nil,
				Dotfiles: nil,
			},
			scope: "all",
			expected: CombinedApplyOutput{
				DryRun:        false,
				Success:       true,
				Packages:      nil,
				Dotfiles:      nil,
				Scope:         "all",
				PackageErrors: nil,
				DotfileErrors: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertApplyResult(tt.result, tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}
