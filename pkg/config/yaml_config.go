package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// YAMLConfig represents the new YAML-based configuration
type YAMLConfig struct {
	Settings  YAMLSettings    `yaml:"settings"`
	Dotfiles  []string        `yaml:"dotfiles,omitempty"`
	Homebrew  HomebrewConfig  `yaml:"homebrew,omitempty"`
	ASDF      []ASDFTool      `yaml:"asdf,omitempty"`
	NPM       []NPMPackage    `yaml:"npm,omitempty"`
}

// YAMLSettings contains global configuration settings
type YAMLSettings struct {
	DefaultManager string `yaml:"default_manager"`
}

// HomebrewConfig contains homebrew package lists
type HomebrewConfig struct {
	Brews []HomebrewPackage `yaml:"brews,omitempty"`
	Casks []HomebrewPackage `yaml:"casks,omitempty"`
}

// HomebrewPackage can be a simple string or complex object
type HomebrewPackage struct {
	Name   string `yaml:"name,omitempty"`
	Config string `yaml:"config,omitempty"`
}

// ASDFTool represents an ASDF tool configuration
type ASDFTool struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Config  string `yaml:"config,omitempty"`
}

// NPMPackage represents an NPM package configuration
type NPMPackage struct {
	Name    string `yaml:"name,omitempty"`
	Package string `yaml:"package,omitempty"` // If different from name
	Config  string `yaml:"config,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling for HomebrewPackage
func (h *HomebrewPackage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case
		h.Name = node.Value
		return nil
	}
	
	// Complex object case
	type homebrewPackageAlias HomebrewPackage
	var pkg homebrewPackageAlias
	if err := node.Decode(&pkg); err != nil {
		return err
	}
	*h = HomebrewPackage(pkg)
	return nil
}

// UnmarshalYAML implements custom unmarshaling for NPMPackage
func (n *NPMPackage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case
		n.Name = node.Value
		return nil
	}
	
	// Complex object case
	type npmPackageAlias NPMPackage
	var pkg npmPackageAlias
	if err := node.Decode(&pkg); err != nil {
		return err
	}
	*n = NPMPackage(pkg)
	return nil
}

// LoadYAMLConfig loads configuration from plonk.yaml and optionally plonk.local.yaml
func LoadYAMLConfig(configDir string) (*YAMLConfig, error) {
	config := &YAMLConfig{
		Settings: YAMLSettings{
			DefaultManager: "homebrew", // Default value
		},
	}

	// Load main config file
	mainConfigPath := filepath.Join(configDir, "plonk.yaml")
	if err := loadYAMLConfigFile(mainConfigPath, config); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", mainConfigPath)
		}
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load local config file if it exists
	localConfigPath := filepath.Join(configDir, "plonk.local.yaml")
	if _, err := os.Stat(localConfigPath); err == nil {
		if err := loadYAMLConfigFile(localConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadYAMLConfigFile loads a single YAML config file and merges it into the config
func loadYAMLConfigFile(path string, config *YAMLConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Create a temporary config to decode into
	var tempConfig YAMLConfig
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Merge settings
	if tempConfig.Settings.DefaultManager != "" {
		config.Settings.DefaultManager = tempConfig.Settings.DefaultManager
	}

	// Merge dotfiles
	config.Dotfiles = append(config.Dotfiles, tempConfig.Dotfiles...)

	// Merge homebrew packages
	config.Homebrew.Brews = append(config.Homebrew.Brews, tempConfig.Homebrew.Brews...)
	config.Homebrew.Casks = append(config.Homebrew.Casks, tempConfig.Homebrew.Casks...)

	// Merge ASDF tools
	config.ASDF = append(config.ASDF, tempConfig.ASDF...)

	// Merge NPM packages
	config.NPM = append(config.NPM, tempConfig.NPM...)

	return nil
}

// validate ensures the configuration is valid
func (c *YAMLConfig) validate() error {
	validManagers := map[string]bool{
		"homebrew": true,
		"asdf":     true,
		"npm":      true,
	}

	// Validate default manager
	if !validManagers[c.Settings.DefaultManager] {
		return fmt.Errorf("invalid default_manager: %s (must be: homebrew, asdf, npm)", c.Settings.DefaultManager)
	}

	// Validate ASDF tools have versions
	for _, tool := range c.ASDF {
		if tool.Version == "" {
			return fmt.Errorf("version is required for ASDF tool: %s", tool.Name)
		}
	}

	return nil
}

// GetDotfileTargets returns dotfiles with their target paths
func (c *YAMLConfig) GetDotfileTargets() map[string]string {
	result := make(map[string]string)
	for _, dotfile := range c.Dotfiles {
		result[dotfile] = sourceToTarget(dotfile)
	}
	return result
}

// sourceToTarget converts a source path to target path using our convention
// Examples:
//   config/nvim/ -> ~/.config/nvim/
//   zshrc -> ~/.zshrc
//   dot_gitconfig -> ~/.gitconfig
func sourceToTarget(source string) string {
	// Handle dot_ prefix convention
	if len(source) > 4 && source[:4] == "dot_" {
		return "~/." + source[4:]
	}
	
	// Handle config/ directory
	if len(source) > 7 && source[:7] == "config/" {
		return "~/." + source
	}
	
	// Default: add ~/. prefix
	return "~/." + source
}