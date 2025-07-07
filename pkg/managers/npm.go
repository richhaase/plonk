// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"os/exec"
	"strings"
)

// NpmManager manages NPM packages.
type NpmManager struct{}

// NewNpmManager creates a new NPM manager.
func NewNpmManager() *NpmManager {
	return &NpmManager{}
}

// IsAvailable checks if NPM is installed.
func (n *NpmManager) IsAvailable() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// ListInstalled lists all globally installed NPM packages.
func (n *NpmManager) ListInstalled() ([]string, error) {
	cmd := exec.Command("npm", "list", "-g", "--depth=0", "--parseable")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}

	var packages []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "/") {
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				pkg := parts[len(parts)-1]
				if pkg != "" && pkg != "lib" {
					packages = append(packages, pkg)
				}
			}
		}
	}

	return packages, nil
}