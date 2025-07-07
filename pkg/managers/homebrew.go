// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"os/exec"
	"strings"
)

// HomebrewManager manages Homebrew packages.
type HomebrewManager struct{}

// NewHomebrewManager creates a new Homebrew manager.
func NewHomebrewManager() *HomebrewManager {
	return &HomebrewManager{}
}

// IsAvailable checks if Homebrew is installed.
func (h *HomebrewManager) IsAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// ListInstalled lists all installed Homebrew packages.
func (h *HomebrewManager) ListInstalled() ([]string, error) {
	cmd := exec.Command("brew", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	return strings.Split(result, "\n"), nil
}