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

// IsAvailable checks if Homebrew is installed
func (h *HomebrewManager) IsAvailable() bool {
	cmd := h.executor.Execute("brew", "--version")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	return err == nil
}

// Install installs a package via Homebrew
func (h *HomebrewManager) Install(packageName string) error {
	cmd := h.executor.Execute("brew", "install", packageName)
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	return cmd.Run()
}

// ListInstalled lists all installed Homebrew packages
func (h *HomebrewManager) ListInstalled() ([]string, error) {
	cmd := h.executor.Execute("brew", "list")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	
	output := strings.TrimSpace(out.String())
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