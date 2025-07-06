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
	ZSH      ZSHConfig      `yaml:"zsh,omitempty"`
	Git      GitConfig      `yaml:"git,omitempty"`
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

// ZSHConfig represents ZSH shell configuration.
type ZSHConfig struct {
	EnvVars      map[string]string `yaml:"env_vars,omitempty"`
	ShellOptions []string          `yaml:"shell_options,omitempty"`
	Inits        []string          `yaml:"inits,omitempty"`
	Completions  []string          `yaml:"completions,omitempty"`
	Plugins      []string          `yaml:"plugins,omitempty"`
	Aliases      map[string]string `yaml:"aliases,omitempty"`
	Functions    map[string]string `yaml:"functions,omitempty"`
	SourceBefore []string          `yaml:"source_before,omitempty"`
	SourceAfter  []string          `yaml:"source_after,omitempty"`
}

// GitConfig represents Git configuration.
type GitConfig struct {
	User    map[string]string            `yaml:"user,omitempty"`
	Core    map[string]string            `yaml:"core,omitempty"`
	Delta   map[string]string            `yaml:"delta,omitempty"`
	Aliases map[string]string            `yaml:"aliases,omitempty"`
	Color   map[string]string            `yaml:"color,omitempty"`
	Fetch   map[string]string            `yaml:"fetch,omitempty"`
	Pull    map[string]string            `yaml:"pull,omitempty"`
	Push    map[string]string            `yaml:"push,omitempty"`
	Status  map[string]string            `yaml:"status,omitempty"`
	Diff    map[string]string            `yaml:"diff,omitempty"`
	Log     map[string]string            `yaml:"log,omitempty"`
	Init    map[string]string            `yaml:"init,omitempty"`
	Rerere  map[string]string            `yaml:"rerere,omitempty"`
	Branch  map[string]string            `yaml:"branch,omitempty"`
	Rebase  map[string]string            `yaml:"rebase,omitempty"`
	Merge   map[string]string            `yaml:"merge,omitempty"`
	Filter  map[string]map[string]string `yaml:"filter,omitempty"`
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
		ZSH: ZSHConfig{
			EnvVars:   make(map[string]string),
			Aliases:   make(map[string]string),
			Functions: make(map[string]string),
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

	// Merge ZSH configuration.
	mergeZSHConfig(&config.ZSH, &tempConfig.ZSH)

	// Merge Git configuration.
	mergeGitConfig(&config.Git, &tempConfig.Git)

	return nil
}

// mergeZSHConfig merges ZSH configuration from source into target.
func mergeZSHConfig(target, source *ZSHConfig) {
	// Merge environment variables.
	if source.EnvVars != nil {
		if target.EnvVars == nil {
			target.EnvVars = make(map[string]string)
		}
		for key, value := range source.EnvVars {
			target.EnvVars[key] = value
		}
	}

	// Merge shell options.
	target.ShellOptions = append(target.ShellOptions, source.ShellOptions...)

	// Merge inits and completions.
	target.Inits = append(target.Inits, source.Inits...)
	target.Completions = append(target.Completions, source.Completions...)

	// Merge plugins.
	target.Plugins = append(target.Plugins, source.Plugins...)

	// Merge aliases.
	if source.Aliases != nil {
		if target.Aliases == nil {
			target.Aliases = make(map[string]string)
		}
		for key, value := range source.Aliases {
			target.Aliases[key] = value
		}
	}

	// Merge functions.
	if source.Functions != nil {
		if target.Functions == nil {
			target.Functions = make(map[string]string)
		}
		for key, value := range source.Functions {
			target.Functions[key] = value
		}
	}

	// Merge source before/after.
	target.SourceBefore = append(target.SourceBefore, source.SourceBefore...)
	target.SourceAfter = append(target.SourceAfter, source.SourceAfter...)
}

// mergeGitConfig merges Git configuration from source into target.
func mergeGitConfig(target, source *GitConfig) {
	// Helper function to merge string maps.
	mergeStringMap := func(targetMap, sourceMap map[string]string) map[string]string {
		if sourceMap == nil {
			return targetMap
		}
		if targetMap == nil {
			targetMap = make(map[string]string)
		}
		for key, value := range sourceMap {
			targetMap[key] = value
		}
		return targetMap
	}

	// Merge all simple string map sections.
	target.User = mergeStringMap(target.User, source.User)
	target.Core = mergeStringMap(target.Core, source.Core)
	target.Delta = mergeStringMap(target.Delta, source.Delta)
	target.Aliases = mergeStringMap(target.Aliases, source.Aliases)
	target.Color = mergeStringMap(target.Color, source.Color)
	target.Fetch = mergeStringMap(target.Fetch, source.Fetch)
	target.Pull = mergeStringMap(target.Pull, source.Pull)
	target.Push = mergeStringMap(target.Push, source.Push)
	target.Status = mergeStringMap(target.Status, source.Status)
	target.Diff = mergeStringMap(target.Diff, source.Diff)
	target.Log = mergeStringMap(target.Log, source.Log)
	target.Init = mergeStringMap(target.Init, source.Init)
	target.Rerere = mergeStringMap(target.Rerere, source.Rerere)
	target.Branch = mergeStringMap(target.Branch, source.Branch)
	target.Rebase = mergeStringMap(target.Rebase, source.Rebase)
	target.Merge = mergeStringMap(target.Merge, source.Merge)

	// Merge filter sections (nested maps).
	if source.Filter != nil {
		if target.Filter == nil {
			target.Filter = make(map[string]map[string]string)
		}
		for filterName, filterConfig := range source.Filter {
			target.Filter[filterName] = mergeStringMap(target.Filter[filterName], filterConfig)
		}
	}
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
