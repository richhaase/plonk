// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"github.com/richhaase/plonk/internal/config"
)

func init() {
	// Register all managers with the config validation system
	// This runs after all package manager init() functions have registered themselves
	registry := NewManagerRegistry()
	config.SetValidManagers(registry.GetAllManagerNames())
}
