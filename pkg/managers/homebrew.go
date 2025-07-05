package managers

import (
	"strings"
)

// HomebrewManager manages Homebrew packages
type HomebrewManager struct {
	runner *CommandRunner
}

// NewHomebrewManager creates a new Homebrew manager
func NewHomebrewManager(executor CommandExecutor) *HomebrewManager {
	return &HomebrewManager{
		runner: NewCommandRunner(executor, "brew"),
	}
}

// IsAvailable checks if Homebrew is installed
func (h *HomebrewManager) IsAvailable() bool {
	err := h.runner.RunCommand("--version")
	return err == nil
}

// Install installs a package via Homebrew
func (h *HomebrewManager) Install(packageName string) error {
	return h.runner.RunCommand("install", packageName)
}

// ListInstalled lists all installed Homebrew packages
func (h *HomebrewManager) ListInstalled() ([]string, error) {
	output, err := h.runner.RunCommandWithOutput("list")
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
	return h.runner.RunCommand("upgrade", packageName)
}

// UpdateAll updates all packages via Homebrew
func (h *HomebrewManager) UpdateAll() error {
	return h.runner.RunCommand("upgrade")
}

// IsInstalled checks if a specific package is installed via Homebrew
func (h *HomebrewManager) IsInstalled(packageName string) bool {
	err := h.runner.RunCommand("list", packageName)
	return err == nil
}