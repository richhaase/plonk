package managers

import (
	"bytes"
	"strings"
)

// HomebrewManager manages Homebrew packages
type HomebrewManager struct {
	executor CommandExecutor
}

// NewHomebrewManager creates a new Homebrew manager
func NewHomebrewManager(executor CommandExecutor) *HomebrewManager {
	return &HomebrewManager{
		executor: executor,
	}
}

// runCommandWithOutput executes a command and returns output + error
func (h *HomebrewManager) runCommandWithOutput(args ...string) (string, error) {
	cmd := h.executor.Execute("brew", args...)
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	return out.String(), err
}

// runCommand executes a command and returns success/error (ignores output)
func (h *HomebrewManager) runCommand(args ...string) error {
	_, err := h.runCommandWithOutput(args...)
	return err
}

// IsAvailable checks if Homebrew is installed
func (h *HomebrewManager) IsAvailable() bool {
	err := h.runCommand("--version")
	return err == nil
}

// Install installs a package via Homebrew
func (h *HomebrewManager) Install(packageName string) error {
	return h.runCommand("install", packageName)
}

// ListInstalled lists all installed Homebrew packages
func (h *HomebrewManager) ListInstalled() ([]string, error) {
	output, err := h.runCommandWithOutput("list")
	if err != nil {
		return nil, err
	}
	
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}
	
	packages := strings.Split(output, "\n")
	
	// Clean up any empty strings
	result := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if trimmed := strings.TrimSpace(pkg); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result, nil
}

// Update updates a specific package via Homebrew
func (h *HomebrewManager) Update(packageName string) error {
	return h.runCommand("upgrade", packageName)
}

// UpdateAll updates all packages via Homebrew
func (h *HomebrewManager) UpdateAll() error {
	return h.runCommand("upgrade")
}

// IsInstalled checks if a specific package is installed via Homebrew
func (h *HomebrewManager) IsInstalled(packageName string) bool {
	err := h.runCommand("list", packageName)
	return err == nil
}