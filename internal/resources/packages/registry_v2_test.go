// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_V2ConfigPreferred_WhenEnabled(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   true,
	}

	v2Config := config.ManagerConfig{
		Binary: "test",
		List: config.ListConfig{
			Command: []string{"test", "list"},
			Parse:   "lines",
		},
	}
	registry.v2Managers["testmgr"] = v2Config

	mgr, err := registry.GetManager("testmgr")

	require.NoError(t, err)
	assert.IsType(t, &GenericManager{}, mgr)
}

func TestRegistry_V2Disabled_ReturnsError(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   false,
	}

	v2Config := config.ManagerConfig{Binary: "test"}
	registry.v2Managers["testmgr"] = v2Config

	_, err := registry.GetManager("testmgr")
	require.Error(t, err)
}

func TestRegistry_LoadV2Configs(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   true,
	}

	cfg := &config.Config{
		Managers: map[string]config.ManagerConfig{
			"custom": {
				Binary: "custom-bin",
				List: config.ListConfig{
					Command: []string{"custom-bin", "list"},
					Parse:   "lines",
				},
			},
		},
	}

	registry.LoadV2Configs(cfg)

	// Should have defaults + custom manager
	assert.Greater(t, len(registry.v2Managers), 1, "should have defaults loaded")
	assert.Contains(t, registry.v2Managers, "custom")
	assert.Contains(t, registry.v2Managers, "brew", "should have default managers")
	assert.Equal(t, "custom-bin", registry.v2Managers["custom"].Binary)
}

func TestRegistry_GetAllManagerNames_UnionAndSorted(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
	}

	registry.v2Managers["npm"] = config.ManagerConfig{}
	registry.v2Managers["pipx"] = config.ManagerConfig{}
	registry.v2Managers["cargo"] = config.ManagerConfig{}
	registry.v2Managers["brew"] = config.ManagerConfig{}

	names := registry.GetAllManagerNames()

	assert.Equal(t, []string{"brew", "cargo", "npm", "pipx"}, names)
}

func TestRegistry_HasManager_ChecksV2Map(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
	}

	registry.v2Managers["v2mgr"] = config.ManagerConfig{}

	assert.True(t, registry.HasManager("v2mgr"))
	assert.False(t, registry.HasManager("nonexistent"))
}

func TestRegistry_EnableV2_TogglesGetManager(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   false,
	}

	registry.v2Managers["test"] = config.ManagerConfig{Binary: "test"}

	registry.EnableV2(false)
	_, err := registry.GetManager("test")
	assert.Error(t, err, "v2 disabled should not return a manager")

	registry.EnableV2(true)
	mgr, _ := registry.GetManager("test")
	assert.IsType(t, &GenericManager{}, mgr, "v2 enabled should use config")
}

func TestRegistry_GetManagerWithExecutor_InjectsExecutor(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   true,
	}

	registry.v2Managers["test"] = config.ManagerConfig{Binary: "test"}

	customExec := &MockCommandExecutor{}
	mgr, err := registry.GetManagerWithExecutor("test", customExec)

	require.NoError(t, err)
	generic, ok := mgr.(*GenericManager)
	require.True(t, ok)
	assert.Equal(t, customExec, generic.exec)
}

func TestRegistry_UnsupportedManager_ReturnsError(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   true,
	}

	_, err := registry.GetManager("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported package manager")
}
