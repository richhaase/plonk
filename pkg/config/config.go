package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the main plonk configuration
type Config struct {
	Settings Settings           `toml:"settings"`
	Packages map[string]Package `toml:"packages"`
}

// Settings contains global configuration settings
type Settings struct {
	DefaultManager string `toml:"default_manager"`
}

// Package represents a single package configuration
type Package struct {
	Manager     string       `toml:"manager,omitempty"`
	Name        string       `toml:"name,omitempty"`
	Version     string       `toml:"version,omitempty"`
	Type        string       `toml:"type,omitempty"`
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// ConfigFile represents a configuration file mapping
type ConfigFile struct {
	Source string `toml:"source"`
	Target string `toml:"target"`
}

// LoadConfig loads configuration from plonk.toml and optionally plonk.local.toml
func LoadConfig(configDir string) (*Config, error) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew", // Default value
		},
		Packages: make(map[string]Package),
	}

	// Load main config file
	mainConfigPath := filepath.Join(configDir, "plonk.toml")
	if err := loadConfigFile(mainConfigPath, config); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", mainConfigPath)
		}
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load local config file if it exists
	localConfigPath := filepath.Join(configDir, "plonk.local.toml")
	if _, err := os.Stat(localConfigPath); err == nil {
		if err := loadConfigFile(localConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
	}

	// Validate and set defaults
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadConfigFile loads a single TOML config file and merges it into the config
func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Create a temporary config to decode into
	var tempConfig Config
	if err := toml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Merge settings
	if tempConfig.Settings.DefaultManager != "" {
		config.Settings.DefaultManager = tempConfig.Settings.DefaultManager
	}

	// Merge packages
	for name, pkg := range tempConfig.Packages {
		config.Packages[name] = pkg
	}

	return nil
}

// validate ensures the configuration is valid and sets defaults
func (c *Config) validate() error {
	validManagers := map[string]bool{
		"homebrew": true,
		"asdf":     true,
		"npm":      true,
	}

	// Validate default manager
	if !validManagers[c.Settings.DefaultManager] {
		return fmt.Errorf("invalid default_manager: %s (must be: homebrew, asdf, npm)", c.Settings.DefaultManager)
	}

	// Validate and set defaults for packages
	for name, pkg := range c.Packages {
		// Set default manager if not specified
		if pkg.Manager == "" {
			pkg.Manager = c.Settings.DefaultManager
		}

		// Validate manager
		if !validManagers[pkg.Manager] {
			return fmt.Errorf("invalid manager for package %s: %s (must be: homebrew, asdf, npm)", name, pkg.Manager)
		}

		// Version is required for ASDF packages
		if pkg.Manager == "asdf" && pkg.Version == "" {
			return fmt.Errorf("version is required for ASDF package: %s", name)
		}

		// Update the package in the map
		c.Packages[name] = pkg
	}

	return nil
}