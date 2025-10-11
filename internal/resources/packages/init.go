// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"github.com/richhaase/plonk/internal/config"
)

func init() {
	// Load default manager configs into registry
	registry := NewManagerRegistry()
	cfg := &config.Config{
		Managers: config.GetDefaultManagers(),
	}
	registry.LoadV2Configs(cfg)

	// Register all managers with the config validation system
	config.SetValidManagers(registry.GetAllManagerNames())
}
