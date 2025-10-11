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
		managers:   make(map[string]*managerEntry),
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

func TestRegistry_V2Disabled_FallsBackToFactory(t *testing.T) {
	registry := &ManagerRegistry{
		managers:   make(map[string]*managerEntry),
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   false,
	}

	v2Config := config.ManagerConfig{Binary: "test"}
	registry.v2Managers["testmgr"] = v2Config

	factory := func(exec CommandExecutor) PackageManager {
		return &mockPackageManagerV2{}
	}
	registry.RegisterV2("testmgr", factory)

	mgr, err := registry.GetManager("testmgr")

	require.NoError(t, err)
	assert.IsType(t, &mockPackageManagerV2{}, mgr)
}

func TestRegistry_LoadV2Configs(t *testing.T) {
	registry := &ManagerRegistry{
		managers:   make(map[string]*managerEntry),
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

	assert.Len(t, registry.v2Managers, 1)
	assert.Contains(t, registry.v2Managers, "custom")
	assert.Equal(t, "custom-bin", registry.v2Managers["custom"].Binary)
}

func TestRegistry_GetAllManagerNames_UnionAndSorted(t *testing.T) {
	registry := &ManagerRegistry{
		managers:   make(map[string]*managerEntry),
		v2Managers: make(map[string]config.ManagerConfig),
	}

	registry.RegisterV2("npm", func(exec CommandExecutor) PackageManager { return nil })
	registry.v2Managers["pipx"] = config.ManagerConfig{}
	registry.v2Managers["cargo"] = config.ManagerConfig{}
	registry.RegisterV2("brew", func(exec CommandExecutor) PackageManager { return nil })

	names := registry.GetAllManagerNames()

	assert.ElementsMatch(t, []string{"npm", "pipx", "cargo", "brew"}, names)
	assert.Equal(t, []string{"brew", "cargo", "npm", "pipx"}, names, "should be sorted")
}

func TestRegistry_HasManager_ChecksBothMaps(t *testing.T) {
	registry := &ManagerRegistry{
		managers:   make(map[string]*managerEntry),
		v2Managers: make(map[string]config.ManagerConfig),
	}

	registry.v2Managers["v2mgr"] = config.ManagerConfig{}
	registry.RegisterV2("gomgr", func(exec CommandExecutor) PackageManager { return nil })

	assert.True(t, registry.HasManager("v2mgr"))
	assert.True(t, registry.HasManager("gomgr"))
	assert.False(t, registry.HasManager("nonexistent"))
}

func TestRegistry_EnableV2_TogglesGetManager(t *testing.T) {
	registry := &ManagerRegistry{
		managers:   make(map[string]*managerEntry),
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   false,
	}

	registry.v2Managers["test"] = config.ManagerConfig{Binary: "test"}
	factory := func(exec CommandExecutor) PackageManager { return &mockPackageManagerV2{} }
	registry.RegisterV2("test", factory)

	registry.EnableV2(false)
	mgr, _ := registry.GetManager("test")
	assert.IsType(t, &mockPackageManagerV2{}, mgr, "v2 disabled should use factory")

	registry.EnableV2(true)
	mgr, _ = registry.GetManager("test")
	assert.IsType(t, &GenericManager{}, mgr, "v2 enabled should use config")
}

func TestRegistry_GetManagerWithExecutor_InjectsExecutor(t *testing.T) {
	registry := &ManagerRegistry{
		managers:   make(map[string]*managerEntry),
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
		managers:   make(map[string]*managerEntry),
		v2Managers: make(map[string]config.ManagerConfig),
		enableV2:   true,
	}

	_, err := registry.GetManager("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported package manager")
}
