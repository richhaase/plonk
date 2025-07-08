// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"fmt"
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

// Install installs a Homebrew package.
func (h *HomebrewManager) Install(name string) error {
	cmd := exec.Command("brew", "install", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// Uninstall removes a Homebrew package.
func (h *HomebrewManager) Uninstall(name string) error {
	cmd := exec.Command("brew", "uninstall", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %w\nOutput: %s", name, err, string(output))
	}
	return nil
}

// IsInstalled checks if a specific package is installed.
func (h *HomebrewManager) IsInstalled(name string) bool {
	cmd := exec.Command("brew", "list", name)
	err := cmd.Run()
	return err == nil
}

