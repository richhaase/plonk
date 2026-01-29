// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/packages"
)

func init() {
	// Set up ManagerChecker for config validation during tests
	config.ManagerChecker = packages.IsSupportedManager
}
