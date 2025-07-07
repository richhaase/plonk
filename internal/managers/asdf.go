// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
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