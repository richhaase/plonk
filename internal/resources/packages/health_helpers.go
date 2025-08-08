// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"fmt"
	"strings"
)

// DefaultCheckHealth provides basic health check implementation
func DefaultCheckHealth(ctx context.Context, manager PackageManager, managerName string) (*HealthCheck, error) {
	check := &HealthCheck{
		Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
		Category: "package-managers",
		Status:   "pass",
		Message:  fmt.Sprintf("%s is available and functional", managerName),
	}

	// Check basic availability
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		if IsContextError(err) {
			return nil, err
		}
		check.Status = "fail"
		check.Message = fmt.Sprintf("%s availability check failed", managerName)
		check.Issues = []string{fmt.Sprintf("Error checking %s: %v", managerName, err)}
		return check, nil
	}

	if !available {
		check.Status = "warn"
		check.Message = fmt.Sprintf("%s is not available", managerName)
		check.Issues = []string{fmt.Sprintf("%s command not found or not functional", managerName)}
		check.Suggestions = []string{fmt.Sprintf("Install %s package manager", managerName)}
		return check, nil
	}

	check.Details = []string{fmt.Sprintf("%s binary found and functional", managerName)}
	return check, nil
}
