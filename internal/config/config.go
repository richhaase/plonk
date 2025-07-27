// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import (
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Config represents the plonk configuration
type Config struct {
	DefaultManager    string   `yaml:"default_manager,omitempty" validate:"omitempty,oneof=brew npm pip gem go cargo test-unavailable"`
	OperationTimeout  int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`
	PackageTimeout    int      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`
	DotfileTimeout    int      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`
	ExpandDirectories []string `yaml:"expand_directories,omitempty"`
	IgnorePatterns    []string `yaml:"ignore_patterns,omitempty"`
	Hooks             Hooks    `yaml:"hooks,omitempty"`
}

// Hooks contains pre and post apply hooks
type Hooks struct {
	PreApply  []Hook `yaml:"pre_apply,omitempty"`
	PostApply []Hook `yaml:"post_apply,omitempty"`
}

// Hook represents a single hook command
type Hook struct {
	Command         string `yaml:"command" validate:"required"`
	Timeout         string `yaml:"timeout,omitempty"`
	ContinueOnError bool   `yaml:"continue_on_error,omitempty"`
}

// defaultConfig holds the default configuration values
var defaultConfig = Config{
	DefaultManager:   "brew",
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

// Load reads and validates configuration from the standard location
func Load(configDir string) (*Config, error) {
	configPath := filepath.Join(configDir, "plonk.yaml")
	return LoadFromPath(configPath)
}

// LoadFromPath reads and validates configuration from a specific path
func LoadFromPath(configPath string) (*Config, error) {
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

// LoadWithDefaults provides zero-config behavior matching current LoadConfigWithDefaults
func LoadWithDefaults(configDir string) *Config {
	cfg, err := Load(configDir)
	if err != nil {
		// Return copy of defaults on any error
		defaultCopy := defaultConfig
		return &defaultCopy
	}
	return cfg
}

// applyDefaults applies default values to a config
func applyDefaults(cfg *Config) {
	if cfg.DefaultManager == "" {
		cfg.DefaultManager = defaultConfig.DefaultManager
	}
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = defaultConfig.OperationTimeout
	}
	if cfg.PackageTimeout == 0 {
		cfg.PackageTimeout = defaultConfig.PackageTimeout
	}
	if cfg.DotfileTimeout == 0 {
		cfg.DotfileTimeout = defaultConfig.DotfileTimeout
	}
	if len(cfg.ExpandDirectories) == 0 {
		cfg.ExpandDirectories = defaultConfig.ExpandDirectories
	}
	if len(cfg.IgnorePatterns) == 0 {
		cfg.IgnorePatterns = defaultConfig.IgnorePatterns
	}
}

// Utility functions for directory management

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	homeDir, _ := os.UserHomeDir()
	return homeDir
}

// GetConfigDir returns the plonk configuration directory
func GetConfigDir() string {
	return GetDefaultConfigDirectory()
}
