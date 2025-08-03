// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "released version shows only version",
			setup: func() {
				versionInfo.Version = "1.0.0"
				versionInfo.Commit = "abc123"
				versionInfo.Date = "2025-08-03"
			},
			expected: "1.0.0",
		},
		{
			name: "dev version shows dev with commit",
			setup: func() {
				versionInfo.Version = "dev"
				versionInfo.Commit = "abc123"
				versionInfo.Date = ""
			},
			expected: "dev-abc123",
		},
		{
			name: "dev version without commit",
			setup: func() {
				versionInfo.Version = "dev"
				versionInfo.Commit = ""
				versionInfo.Date = ""
			},
			expected: "dev-",
		},
		{
			name: "empty version returns empty",
			setup: func() {
				versionInfo.Version = ""
				versionInfo.Commit = "abc123"
				versionInfo.Date = "2025-08-03"
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersionInfo := versionInfo
			defer func() {
				versionInfo = origVersionInfo
			}()

			tt.setup()
			result := formatVersion()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompleteOutputFormats(t *testing.T) {
	tests := []struct {
		name       string
		toComplete string
		want       []string
		directive  cobra.ShellCompDirective
	}{
		{
			name:       "always returns all formats",
			toComplete: "",
			want:       []string{"table", "json", "yaml"},
			directive:  cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:       "returns all formats even with partial input",
			toComplete: "j",
			want:       []string{"table", "json", "yaml"},
			directive:  cobra.ShellCompDirectiveNoFileComp,
		},
		{
			name:       "returns all formats with any input",
			toComplete: "xyz",
			want:       []string{"table", "json", "yaml"},
			directive:  cobra.ShellCompDirectiveNoFileComp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, directive := completeOutputFormats(nil, nil, tt.toComplete)
			assert.Equal(t, tt.want, suggestions)
			assert.Equal(t, tt.directive, directive)
		})
	}
}
