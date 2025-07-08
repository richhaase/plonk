// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package managers

import (
	"os"
	"path/filepath"

	"plonk/internal/config"
)

// PlonkConfigLoader implements ConfigLoader using plonk's config system
type PlonkConfigLoader struct {
	configDir string
}

// NewPlonkConfigLoader creates a new config loader
func NewPlonkConfigLoader(configDir string) *PlonkConfigLoader {
	return &PlonkConfigLoader{
		configDir: configDir,
	}
}

// GetPackagesForManager returns packages for the specified manager from plonk.yaml
func (p *PlonkConfigLoader) GetPackagesForManager(managerName string) ([]ConfigPackage, error) {
	cfg, err := config.LoadConfig(p.configDir)
	if err != nil {
		// If config doesn't exist, return empty list
		if os.IsNotExist(err) {
			return []ConfigPackage{}, nil
		}
		return nil, err
	}

	var packages []ConfigPackage

	switch managerName {
	case "homebrew":
		// Add brews
		for _, brew := range cfg.Homebrew.Brews {
			packages = append(packages, ConfigPackage{
				Name:    brew.Name,
				Version: "", // Homebrew doesn't use versions in basic config
			})
		}
		// Add casks
		for _, cask := range cfg.Homebrew.Casks {
			packages = append(packages, ConfigPackage{
				Name:    cask.Name,
				Version: "", // Casks don't use versions
			})
		}
	case "npm":
		for _, pkg := range cfg.NPM {
			name := pkg.Name
			if name == "" && pkg.Package != "" {
				name = pkg.Package
			}
			if name != "" {
				packages = append(packages, ConfigPackage{
					Name:    name,
					Version: "", // NPM packages don't have versions in this config structure
				})
			}
		}
	}

	return packages, nil
}

// DefaultConfigDir returns the default plonk configuration directory
func DefaultConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "plonk"), nil
}