// Package importer provides functionality for discovering existing shell environment
// configurations and generating plonk.yaml files. It supports package discovery
// from Homebrew, ASDF, NPM, and dotfile detection for comprehensive environment
// import capabilities.
package importer

import (
	"strings"

	"plonk/pkg/managers"
)

// HomebrewDiscoverer discovers installed Homebrew packages.
type HomebrewDiscoverer struct {
	executor managers.CommandExecutor
}

// NewHomebrewDiscoverer creates a new Homebrew package discoverer.
func NewHomebrewDiscoverer(executor managers.CommandExecutor) *HomebrewDiscoverer {
	return &HomebrewDiscoverer{
		executor: executor,
	}
}

// DiscoverPackages returns a list of installed Homebrew packages.
func (h *HomebrewDiscoverer) DiscoverPackages() ([]string, error) {
	cmd := h.executor.Execute("brew", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	packages := strings.Split(outputStr, "\n")

	// Clean up any empty strings
	var result []string
	for _, pkg := range packages {
		if pkg != "" {
			result = append(result, pkg)
		}
	}

	return result, nil
}
