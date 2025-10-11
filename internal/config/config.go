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
	DefaultManager    string                   `yaml:"default_manager,omitempty" validate:"omitempty,validmanager"`
	OperationTimeout  int                      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"`
	PackageTimeout    int                      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`
	DotfileTimeout    int                      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`
	ExpandDirectories []string                 `yaml:"expand_directories,omitempty"`
	IgnorePatterns    []string                 `yaml:"ignore_patterns,omitempty"`
	Dotfiles          Dotfiles                 `yaml:"dotfiles,omitempty"`
	DiffTool          string                   `yaml:"diff_tool,omitempty"`
	Managers          map[string]ManagerConfig `yaml:"managers,omitempty"`
}

// Dotfiles contains dotfile-specific configuration
type Dotfiles struct {
	UnmanagedFilters []string `yaml:"unmanaged_filters,omitempty"`
}

// defaultConfig holds the default configuration values
var defaultConfig = Config{
	DefaultManager:   "brew",
	OperationTimeout: 300, // 5 minutes
	PackageTimeout:   180, // 3 minutes
	DotfileTimeout:   60,  // 1 minute
	ExpandDirectories: []string{
		".config",
	},
	Managers: GetDefaultManagers(),
	IgnorePatterns: []string{
		// System files
		".DS_Store",
		".Trash",
		".CFUserTextEncoding",
		".cups",

		// Version control
		".git",

		// Backup/temp files
		"*.backup",
		"*.tmp",
		"*.swp",

		// Plonk files
		"plonk.lock",

		// Tool caches/data (usually not meant to be tracked)
		".cache",
		".npm",
		".gem",
		".cargo",
		".rustup",
		".bundle",
		".local",
		".ollama",

		// History files
		"*_history",
		"*.lesshst",

		// Application data
		".cursor",
		".lima",
		".colima",
		".cdk",
		".magefile",

		// Security sensitive (should be managed separately)
		".ssh",
		".gnupg",

		// Tokens and auth
		"*_token",
		"*.pem",
		"*.key",
	},
	Dotfiles: Dotfiles{
		UnmanagedFilters: []string{
			// File patterns for unmanaged filtering only
			"*.log",
			"*.lock",
			"*.db",
			"*.cache",
			"*.map",
			"*.pid",
			"*.sock",
			"*.socket",
			"*.sqlite",
			"*.sqlite3",
			"*.wasm",
			"*.idx",
			"*.pack",

			// Directory patterns
			"**/node_modules/**",
			"**/plugins/**",
			"**/extensions/**",
			"**/__pycache__/**",
			"**/logs/**",
			"**/tmp/**",
			"**/temp/**",
			"**/dist/**",
			"**/build/**",
			"**/out/**",
			"**/.git/**",

			// Cache patterns
			"**/*cache*/**",
			"**/Cache/**",
			"**/Caches/**",

			// UUID-like directory patterns (generic for extensions, etc.)
			"**/*-*-*-*-*/**",

			// State/session patterns
			"**/*state*",
			"**/*session*",
			"**/*State*",

			// Git internals
			"**/.git*",
			"**/hooks/**",
			"**/objects/**",
			"**/refs/**",

			// Tool-specific patterns
			"**/spec/**",
			"**/test/**",
			"**/tests/**",
			"**/assets/**",
			"**/tools/**",
		},
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
	if err := RegisterValidators(validate); err != nil {
		return nil, err
	}
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
