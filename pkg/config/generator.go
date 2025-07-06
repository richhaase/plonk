package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DiscoveryResults contains the results from package and dotfile discovery.
type DiscoveryResults struct {
	HomebrewPackages []string
	AsdfTools        []string
	NpmPackages      []string
	Dotfiles         []string
	ZSHConfig        ZSHConfig
	GitConfig        GitConfig
}

// GenerateConfig creates a Config struct from discovery results.
func GenerateConfig(results DiscoveryResults) Config {
	config := Config{
		Settings: Settings{
			DefaultManager: "homebrew",
		},
		Dotfiles: make([]string, len(results.Dotfiles)),
		Homebrew: HomebrewConfig{
			Brews: make([]HomebrewPackage, len(results.HomebrewPackages)),
		},
		ASDF: make([]ASDFTool, len(results.AsdfTools)),
		NPM:  make([]NPMPackage, len(results.NpmPackages)),
		ZSH:  results.ZSHConfig,
		Git:  results.GitConfig,
	}

	// Copy dotfiles (for any remaining non-special dotfiles)
	copy(config.Dotfiles, results.Dotfiles)

	// Convert Homebrew packages
	for i, pkg := range results.HomebrewPackages {
		config.Homebrew.Brews[i] = HomebrewPackage{Name: pkg}
	}

	// Convert ASDF tools (format: "tool version")
	for i, tool := range results.AsdfTools {
		parts := strings.Fields(tool)
		if len(parts) >= 2 {
			config.ASDF[i] = ASDFTool{
				Name:    parts[0],
				Version: parts[1],
			}
		}
	}

	// Convert NPM packages
	for i, pkg := range results.NpmPackages {
		config.NPM[i] = NPMPackage{Name: pkg}
	}

	return config
}

// SaveConfig marshals a Config struct to YAML and writes it to the specified file.
func SaveConfig(config Config, filePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create YAML encoder with proper settings
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2) // Use 2-space indentation
	defer encoder.Close()

	// Encode the config
	return encoder.Encode(&config)
}
