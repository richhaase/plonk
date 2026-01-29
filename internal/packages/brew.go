// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os/exec"
	"strings"
)

// BrewSimple implements Manager for Homebrew
type BrewSimple struct{}

// NewBrewSimple creates a new Homebrew manager
func NewBrewSimple() *BrewSimple {
	return &BrewSimple{}
}

// IsInstalled checks if a package is installed via brew
func (b *BrewSimple) IsInstalled(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "brew", "list", "--formula", name)
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means not installed, not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Install installs a package via brew
func (b *BrewSimple) Install(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "brew", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already installed (idempotent)
		if strings.Contains(strings.ToLower(string(output)), "already installed") {
			return nil
		}
		return err
	}
	return nil
}
