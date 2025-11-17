// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
)

// ExecuteCommand runs a command with the given arguments and returns the output.
// This is a simple wrapper around exec.CommandContext for common usage patterns.
func ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return defaultExecutor.Execute(ctx, name, args...)
}

// managerInstallHint returns the install hint for a manager from config/defaults.
func managerInstallHint(cfg *config.Config, manager string) string {
	source := cfg
	if source == nil {
		source = config.LoadWithDefaults(config.GetConfigDir())
	}
	if source != nil && source.Managers != nil {
		if m, ok := source.Managers[manager]; ok && m.InstallHint != "" {
			return m.InstallHint
		}
	}
	return "check installation instructions for " + manager
}
