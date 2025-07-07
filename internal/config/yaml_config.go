// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package config provides configuration management for Plonk, including YAML
// configuration parsing, validation, and generation of shell configuration files.
//
// The package supports loading configuration from plonk.yaml and plonk.local.yaml,
// validating package definitions and file paths, and generating shell-specific
// configuration files like .zshrc, .zshenv, and .gitconfig.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure.
type Config struct {
	Settings Settings       `yaml:"settings" validate:"required"`
	Backup   BackupConfig   `yaml:"backup,omitempty"`
	Dotfiles []string       `yaml:"dotfiles,omitempty" validate:"dive,file_path"`
	Homebrew HomebrewConfig `yaml:"homebrew,omitempty"`
	ASDF     []ASDFTool     `yaml:"asdf,omitempty" validate:"dive"`
	NPM      []NPMPackage   `yaml:"npm,omitempty" validate:"dive"`
}

// Settings contains global configuration settings.
type Settings struct {
	DefaultManager string `yaml:"default_manager" validate:"required,oneof=homebrew asdf npm"`
}

// BackupConfig contains backup configuration settings.
type BackupConfig struct {
	Location  string `yaml:"location,omitempty"`
	KeepCount int    `yaml:"keep_count,omitempty"`
}

// HomebrewConfig contains homebrew package lists.
type HomebrewConfig struct {
	Brews []HomebrewPackage `yaml:"brews,omitempty" validate:"dive"`
	Casks []HomebrewPackage `yaml:"casks,omitempty" validate:"dive"`
}

// HomebrewPackage can be a simple string or complex object.
type HomebrewPackage struct {
	Name   string `yaml:"name,omitempty" validate:"required,package_name"`
	Config string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}

// ASDFTool represents an ASDF tool configuration.
type ASDFTool struct {
	Name    string `yaml:"name" validate:"required,package_name"`
	Version string `yaml:"version" validate:"required"`
	Config  string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}

// NPMPackage represents an NPM package configuration.
type NPMPackage struct {
	Name    string `yaml:"name,omitempty" validate:"omitempty,package_name"`
	Package string `yaml:"package,omitempty" validate:"omitempty,package_name"` // If different from name.
	Config  string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}

// MarshalYAML implements custom marshaling for HomebrewPackage.
func (h HomebrewPackage) MarshalYAML() (interface{}, error) {
	// If only Name is set, marshal as a simple string
	if h.Config == "" {
		return h.Name, nil
	}
	// Otherwise, marshal as a full object
	type homebrewPackageAlias HomebrewPackage
	return homebrewPackageAlias(h), nil
}

// UnmarshalYAML implements custom unmarshaling for HomebrewPackage.
func (h *HomebrewPackage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case.
		h.Name = node.Value
		return nil
	}

	// Complex object case.
	type homebrewPackageAlias HomebrewPackage
	var pkg homebrewPackageAlias
	if err := node.Decode(&pkg); err != nil {
		return err
	}
	*h = HomebrewPackage(pkg)
	return nil
}

// MarshalYAML implements custom marshaling for NPMPackage.
func (n NPMPackage) MarshalYAML() (interface{}, error) {
	// If only Name is set (no Package or Config), marshal as a simple string
	if n.Package == "" && n.Config == "" {
		return n.Name, nil
	}
	// Otherwise, marshal as a full object
	type npmPackageAlias NPMPackage
	return npmPackageAlias(n), nil
}

// UnmarshalYAML implements custom unmarshaling for NPMPackage.
func (n *NPMPackage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		// Simple string case.
		n.Name = node.Value
		return nil
	}

	// Complex object case.
	type npmPackageAlias NPMPackage
	var pkg npmPackageAlias
	if err := node.Decode(&pkg); err != nil {
		return err
	}
	*n = NPMPackage(pkg)
	return nil
}

// LoadConfig loads configuration from plonk.yaml and optionally plonk.local.yaml.
func LoadConfig(configDir string) (*Config, error) {
	config := &Config{
		Settings: Settings{
			DefaultManager: "homebrew", // Default value.
		},
	}

	// Load main config file - check both main directory and repo subdirectory.
	mainConfigPath := filepath.Join(configDir, "plonk.yaml")
	repoConfigPath := filepath.Join(configDir, "repo", "plonk.yaml")

	// Try main directory first.
	if err := loadConfigFile(mainConfigPath, config); err != nil {
		if os.IsNotExist(err) {
			// Try repo subdirectory.
			if err := loadConfigFile(repoConfigPath, config); err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("config file not found in %s or %s", mainConfigPath, repoConfigPath)
				}
				return nil, fmt.Errorf("failed to load config from repo directory: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Load local config file if it exists.
	localConfigPath := filepath.Join(configDir, "plonk.local.yaml")
	if _, err := os.Stat(localConfigPath); err == nil {
		if err := loadConfigFile(localConfigPath, config); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
	}

	// Validate configuration with new unified validator.
	validator := NewSimpleValidator()
	result := validator.ValidateConfig(config)
	if !result.IsValid() {
		return nil, fmt.Errorf("config validation failed: %s", strings.Join(result.Errors, "; "))
	}

	return config, nil
}

// loadConfigFile loads a single YAML config file and merges it into the config.
func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Create a temporary config to decode into.
	var tempConfig Config
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Merge settings.
	if tempConfig.Settings.DefaultManager != "" {
		config.Settings.DefaultManager = tempConfig.Settings.DefaultManager
	}

	// Merge dotfiles.
	config.Dotfiles = append(config.Dotfiles, tempConfig.Dotfiles...)

	// Merge homebrew packages.
	config.Homebrew.Brews = append(config.Homebrew.Brews, tempConfig.Homebrew.Brews...)
	config.Homebrew.Casks = append(config.Homebrew.Casks, tempConfig.Homebrew.Casks...)

	// Merge ASDF tools.
	config.ASDF = append(config.ASDF, tempConfig.ASDF...)

	// Merge NPM packages.
	config.NPM = append(config.NPM, tempConfig.NPM...)

	// Merge backup configuration.
	if tempConfig.Backup.Location != "" {
		config.Backup.Location = tempConfig.Backup.Location
	}
	if tempConfig.Backup.KeepCount > 0 {
		config.Backup.KeepCount = tempConfig.Backup.KeepCount
	}

	return nil
}

// GetDotfileTargets returns dotfiles with their target paths.
func (c *Config) GetDotfileTargets() map[string]string {
	result := make(map[string]string)
	for _, dotfile := range c.Dotfiles {
		result[dotfile] = sourceToTarget(dotfile)
	}
	return result
}

// sourceToTarget converts a source path to target path using our convention
// Examples:
//
//	config/nvim/ -> ~/.config/nvim/
//	zshrc -> ~/.zshrc
//	dot_gitconfig -> ~/.gitconfig
func sourceToTarget(source string) string {
	// Handle dot_ prefix convention.
	if len(source) > 4 && source[:4] == "dot_" {
		return "~/." + source[4:]
	}

	// Handle config/ directory.
	if len(source) > 7 && source[:7] == "config/" {
		return "~/." + source
	}

	// Default: add ~/. prefix.
	return "~/." + source
}
