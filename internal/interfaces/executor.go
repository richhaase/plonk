// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package interfaces

import "context"

// CommandExecutor abstracts command execution for testing
type CommandExecutor interface {
	// Execute runs a command and returns stdout
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)

	// ExecuteCombined runs a command and returns combined stdout/stderr
	ExecuteCombined(ctx context.Context, name string, args ...string) ([]byte, error)

	// LookPath checks if a binary exists in PATH
	LookPath(name string) (string, error)
}
