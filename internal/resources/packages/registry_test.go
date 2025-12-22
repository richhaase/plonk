// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerRegistry_GetManager_ReturnsCorrectType(t *testing.T) {
	registry := GetRegistry()

	tests := []struct {
		name         string
		manager      string
		expectedType interface{}
	}{
		{"brew", "brew", &BrewManager{}},
		{"cargo", "cargo", &CargoManager{}},
		{"go", "go", &GoManager{}},
		{"npm", "npm", &NPMManager{}},
		{"pnpm", "pnpm", &NPMManager{}},
		{"bun", "bun", &NPMManager{}},
		{"uv", "uv", &UVManager{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := registry.GetManager(tt.manager)
			require.NoError(t, err)
			assert.IsType(t, tt.expectedType, mgr)
		})
	}
}

func TestManagerRegistry_UnknownManager_Error(t *testing.T) {
	registry := GetRegistry()
	_, err := registry.GetManager("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported package manager")
}

func TestManagerRegistry_GetAllManagerNames_Sorted(t *testing.T) {
	registry := GetRegistry()
	names := registry.GetAllManagerNames()

	// Should be sorted alphabetically
	expected := []string{"brew", "bun", "cargo", "go", "npm", "pnpm", "uv"}
	assert.Equal(t, expected, names)
}

func TestManagerRegistry_HasManager(t *testing.T) {
	registry := GetRegistry()

	// Supported managers
	assert.True(t, registry.HasManager("brew"))
	assert.True(t, registry.HasManager("npm"))
	assert.True(t, registry.HasManager("cargo"))
	assert.True(t, registry.HasManager("go"))
	assert.True(t, registry.HasManager("uv"))

	// Unsupported managers
	assert.False(t, registry.HasManager("unknown"))
	assert.False(t, registry.HasManager("pip"))
	assert.False(t, registry.HasManager("gem"))
}

func TestManagerRegistry_GetManagerWithExecutor(t *testing.T) {
	registry := GetRegistry()
	mock := &MockCommandExecutor{
		Responses: map[string]CommandResponse{
			"brew --version": {Output: []byte("Homebrew 4.0.0")},
		},
	}

	mgr, err := registry.GetManagerWithExecutor("brew", mock)
	require.NoError(t, err)
	assert.NotNil(t, mgr)

	// The manager should use our mock executor
	brewMgr, ok := mgr.(*BrewManager)
	require.True(t, ok)
	assert.NotNil(t, brewMgr)
}
