// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"os"

	"github.com/richhaase/plonk/internal/config"
)

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	homeDir, _ := os.UserHomeDir()
	return homeDir
}

// GetConfigDir returns the plonk configuration directory
func GetConfigDir() string {
	return config.GetDefaultConfigDirectory()
}
