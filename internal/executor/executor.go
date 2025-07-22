// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package executor

import (
	"context"
	"os/exec"

	"github.com/richhaase/plonk/internal/interfaces"
)

// CommandExecutor is an alias for the unified interface
type CommandExecutor = interfaces.CommandExecutor

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

// Execute runs a command and returns stdout
func (r *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// ExecuteCombined runs a command and returns combined stdout/stderr
func (r *RealCommandExecutor) ExecuteCombined(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// LookPath checks if a binary exists in PATH
func (r *RealCommandExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
