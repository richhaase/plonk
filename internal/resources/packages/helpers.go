// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"strings"
)

// ExtractExitCode attempts to extract the exit code from an exec error.
// Returns the exit code and true if successful, 0 and false otherwise.
func ExtractExitCode(err error) (int, bool) {
	if execErr, ok := err.(interface{ ExitCode() int }); ok {
		return execErr.ExitCode(), true
	}
	return 0, false
}

// SplitLines splits output into lines, filtering out empty lines.
func SplitLines(output []byte) []string {
	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}
	}

	lines := strings.Split(result, "\n")
	var filtered []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return filtered
}

// ExecuteCommand runs a command with the given arguments and returns the output.
// This is a simple wrapper around exec.CommandContext for common usage patterns.
func ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return defaultExecutor.Execute(ctx, name, args...)
}

// IsContextError checks if an error is a context cancellation or deadline error.
func IsContextError(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}
