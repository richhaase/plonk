// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManagerRegistry_HasManager(t *testing.T) {
	registry := NewManagerRegistry()

	tests := []struct {
		name     string
		manager  string
		expected bool
	}{
		{
			name:     "brew manager exists",
			manager:  "brew",
			expected: true,
		},
		{
			name:     "npm manager exists",
			manager:  "npm",
			expected: true,
		},
		{
			name:     "cargo manager exists",
			manager:  "cargo",
			expected: true,
		},
		{
			name:     "uv manager exists",
			manager:  "uv",
			expected: true,
		},
		{
			name:     "gem manager exists",
			manager:  "gem",
			expected: true,
		},
		{
			name:     "go manager exists",
			manager:  "go",
			expected: true,
		},
		{
			name:     "invalid manager does not exist",
			manager:  "invalid",
			expected: false,
		},
		{
			name:     "empty manager does not exist",
			manager:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.HasManager(tt.manager)
			assert.Equal(t, tt.expected, result)
		})
	}
}
