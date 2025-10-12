// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"testing"

	"github.com/richhaase/plonk/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerRegistry_V2_GetManager_ReturnsGeneric(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: map[string]config.ManagerConfig{
			"testmgr": {Binary: "test"},
		},
		enableV2: true,
	}

	mgr, err := registry.GetManager("testmgr")
	require.NoError(t, err)
	assert.IsType(t, &GenericManager{}, mgr)
}

func TestManagerRegistry_V2_UnknownManager_Error(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: map[string]config.ManagerConfig{},
		enableV2:   true,
	}
	_, err := registry.GetManager("unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported package manager")
}

func TestManagerRegistry_GetAllManagerNames_Sorted(t *testing.T) {
	registry := &ManagerRegistry{
		v2Managers: map[string]config.ManagerConfig{
			"npm":  {},
			"brew": {},
			"uv":   {},
		},
		enableV2: true,
	}
	names := registry.GetAllManagerNames()
	assert.Equal(t, []string{"brew", "npm", "uv"}, names)
}
