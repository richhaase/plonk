// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"testing"

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
