// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"fmt"
	"os/exec"
	"strings"
)

// AsdfManager manages ASDF tools.
type AsdfManager struct{}

// NewAsdfManager creates a new ASDF manager.
func NewAsdfManager() *AsdfManager {
	return &AsdfManager{}
}

// IsAvailable checks if ASDF is installed.
func (a *AsdfManager) IsAvailable() bool {
	_, err := exec.LookPath("asdf")
	return err == nil
}

// ListInstalled lists all installed ASDF tools.
func (a *AsdfManager) ListInstalled() ([]string, error) {
	cmd := exec.Command("asdf", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	var tools []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, " ") {
			tools = append(tools, line)
		}
	}

	return tools, nil
}

// Install installs an ASDF tool with a specific version.
func (a *AsdfManager) Install(name string, version string) error {
	if version == "" {
		return fmt.Errorf("version is required for ASDF tool %s", name)
	}
	
	// First, ensure the plugin is installed
	cmd := exec.Command("asdf", "plugin", "add", name)
	cmd.Run() // Ignore error as plugin might already be installed
	
	// Install the specific version
	cmd = exec.Command("asdf", "install", name, version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s %s: %w\nOutput: %s", name, version, err, string(output))
	}
	
	// Set as global version
	cmd = exec.Command("asdf", "global", name, version)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set global version for %s %s: %w\nOutput: %s", name, version, err, string(output))
	}
	
	return nil
}

// Uninstall removes an ASDF tool.
func (a *AsdfManager) Uninstall(name string) error {
	// Get current version to uninstall
	version, err := a.GetVersion(name)
	if err != nil {
		return fmt.Errorf("failed to get version for %s: %w", name, err)
	}
	
	cmd := exec.Command("asdf", "uninstall", name, version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall %s %s: %w\nOutput: %s", name, version, err, string(output))
	}
	
	return nil
}

// IsInstalled checks if a specific tool is installed.
func (a *AsdfManager) IsInstalled(name string) bool {
	cmd := exec.Command("asdf", "current", name)
	err := cmd.Run()
	return err == nil
}

// GetVersion gets the currently active version of a tool.
func (a *AsdfManager) GetVersion(name string) (string, error) {
	cmd := exec.Command("asdf", "current", name)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version for %s: %w", name, err)
	}
	
	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("no version found for %s", name)
	}
	
	// Parse output format: "nodejs 20.0.0 (set by /path/to/.tool-versions)"
	parts := strings.Fields(result)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	
	return "", fmt.Errorf("could not parse version for %s", name)
}