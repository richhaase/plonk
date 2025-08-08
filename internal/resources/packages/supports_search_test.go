// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManagers_SupportsSearch(t *testing.T) {
	tests := []struct {
		name     string
		manager  PackageManager
		expected bool
	}{
		{
			name:     "Cargo supports search",
			manager:  &CargoManager{},
			expected: true,
		},
		{
			name:     "Gem supports search",
			manager:  &GemManager{},
			expected: true,
		},
		{
			name:     "GoInstall does not support search",
			manager:  &GoInstallManager{},
			expected: false,
		},
		{
			name:     "Homebrew supports search",
			manager:  &HomebrewManager{},
			expected: true,
		},
		{
			name:     "NPM supports search",
			manager:  &NpmManager{},
			expected: true,
		},
		{
			name:     "UV does not support search",
			manager:  &UvManager{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.manager.SupportsSearch())
		})
	}
}
