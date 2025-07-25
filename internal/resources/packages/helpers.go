// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os/exec"
	"strings"
)

// ExecuteCommandCombined runs a command and returns combined stdout and stderr.
// Useful for commands where error output is important for diagnostics.
func ExecuteCommandCombined(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

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
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// CheckCommandAvailable checks if a command exists in PATH and is executable.
func CheckCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// VerifyBinary verifies that a binary is functional by running a version command.
// Returns nil if the binary executes successfully, error otherwise.
func VerifyBinary(ctx context.Context, binary string, versionArgs []string) error {
	if len(versionArgs) == 0 {
		versionArgs = []string{"--version"} // Default version argument
	}

	cmd := exec.CommandContext(ctx, binary, versionArgs...)
	_, err := cmd.Output()
	if err != nil {
		// Check for context cancellation - return directly without wrapping
		if err == context.Canceled || err == context.DeadlineExceeded {
			return err
		}
		// Return the error wrapped with context
		return err
	}
	return nil
}

// IsContextError checks if an error is a context cancellation or deadline error.
func IsContextError(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded
}

// FindAvailableBinary checks a list of binary names and returns the first one that is
// available and functional. Returns empty string if none are found.
func FindAvailableBinary(ctx context.Context, binaries []string, versionArgs []string) string {
	for _, binary := range binaries {
		if _, err := exec.LookPath(binary); err == nil {
			// Verify the binary works
			cmd := exec.CommandContext(ctx, binary, versionArgs...)
			if _, err := cmd.Output(); err == nil {
				return binary
			}
			// Check for context cancellation
			if ctx.Err() != nil {
				return ""
			}
		}
	}
	return ""
}
