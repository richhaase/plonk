// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"github.com/richhaase/plonk/internal/config"
)

func init() {
	// Load default manager configs into the registry so that package operations
	// have a sensible baseline even before any user configuration is loaded.
	registry := GetRegistry()
	cfg := &config.Config{
		Managers: config.GetDefaultManagers(),
	}
	registry.LoadV2Configs(cfg)
}
