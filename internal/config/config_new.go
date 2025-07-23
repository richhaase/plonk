// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// NewConfig represents the plonk configuration
// This will replace the current Config struct
type NewConfig struct {
	DefaultManager    string   `yaml:"default_manager,omitempty" validate:"omitempty,oneof=homebrew npm pip gem go cargo test-unavailable"`
	OperationTimeout  int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`
	PackageTimeout    int      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`
	DotfileTimeout    int      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`
	ExpandDirectories []string `yaml:"expand_directories,omitempty"`
	IgnorePatterns    []string `yaml:"ignore_patterns,omitempty"`
}

// defaultConfig holds the default configuration values
var defaultConfig = NewConfig{
	DefaultManager:   "homebrew",
	OperationTimeout: 300, // 5 minutes
	PackageTimeout:   180, // 3 minutes
	DotfileTimeout:   60,  // 1 minute
	ExpandDirectories: []string{
		".config",
		".ssh",
		".aws",
		".kube",
		".docker",
		".gnupg",
		".local",
	},
	IgnorePatterns: []string{
		".DS_Store",
		".git",
		"*.backup",
		"*.tmp",
		"*.swp",
		"plonk.lock",
	},
}

// LoadNew reads and validates configuration from the standard location
func LoadNew(configDir string) (*NewConfig, error) {
	configPath := filepath.Join(configDir, "plonk.yaml")
	return LoadNewFromPath(configPath)
}

// LoadNewFromPath reads and validates configuration from a specific path
func LoadNewFromPath(configPath string) (*NewConfig, error) {
	// Start with a copy of defaults
	cfg := defaultConfig

	// Read file if it exists
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Zero-config: return defaults if file doesn't exist
			return &cfg, nil
		}
		return nil, err
	}

	// Unmarshal YAML over defaults
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Validate
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadNewWithDefaults provides zero-config behavior matching current LoadConfigWithDefaults
func LoadNewWithDefaults(configDir string) *NewConfig {
	cfg, err := LoadNew(configDir)
	if err != nil {
		// Return copy of defaults on any error
		defaultCopy := defaultConfig
		return &defaultCopy
	}
	return cfg
}

// Resolve returns self for API compatibility
// In the new system, Config IS the resolved config
func (c *NewConfig) Resolve() *NewConfig {
	return c
}

// GetDefaultManager returns the default package manager
func (c *NewConfig) GetDefaultManager() string {
	return c.DefaultManager
}

// GetOperationTimeout returns operation timeout in seconds
func (c *NewConfig) GetOperationTimeout() int {
	return c.OperationTimeout
}

// GetPackageTimeout returns package timeout in seconds
func (c *NewConfig) GetPackageTimeout() int {
	return c.PackageTimeout
}

// GetDotfileTimeout returns dotfile timeout in seconds
func (c *NewConfig) GetDotfileTimeout() int {
	return c.DotfileTimeout
}

// GetExpandDirectories returns directories to expand
func (c *NewConfig) GetExpandDirectories() []string {
	return c.ExpandDirectories
}

// GetIgnorePatterns returns patterns to ignore
func (c *NewConfig) GetIgnorePatterns() []string {
	return c.IgnorePatterns
}
