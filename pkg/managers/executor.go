// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import "os/exec"

// RealCommandExecutor implements CommandExecutor for actual command execution
type RealCommandExecutor struct{}

// NewRealCommandExecutor creates a new real command executor
func NewRealCommandExecutor() *RealCommandExecutor {
	return &RealCommandExecutor{}
}

// Execute creates and returns a command for execution
func (r *RealCommandExecutor) Execute(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
