// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
)

// executeInstallCommand runs an installation command with proper error handling
func executeInstallCommand(ctx context.Context, name string, args []string, managerName string) error {
	output, err := defaultExecutor.CombinedOutput(ctx, name, args...)
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", managerName, err, string(output))
	}
	return nil
}

// executeInstallScript runs a shell script for installation
func executeInstallScript(ctx context.Context, script string, managerName string) error {
	return executeInstallCommand(ctx, "bash", []string{"-c", script}, managerName)
}

// checkPackageManagerAvailable checks if a package manager is available for delegation
func checkPackageManagerAvailable(ctx context.Context, managerName string) (bool, error) {
	registry := NewManagerRegistry()
	mgr, err := registry.GetManager(managerName)
	if err != nil {
		return false, err
	}
	return mgr.IsAvailable(ctx)
}

// Note: requiresUserConfirmation, describeInstallMethod, and describeSecurityLevel functions removed
// These were planned for interactive prompting which is not supported per project requirements
