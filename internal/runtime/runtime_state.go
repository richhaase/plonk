// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package runtime provides a unified interface for configuration and state management.
// It eliminates the tight coupling between config and state packages by providing
// a single point of interaction for commands.
package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/state"
)

// RuntimeState provides a unified interface for configuration and state management.
// It encapsulates both configuration data and runtime state operations,
// eliminating the need for complex adapter patterns between config and state packages.
type RuntimeState interface {
	// Configuration Management
	LoadConfiguration() error
	SaveConfiguration() error
	ValidateConfiguration() error

	// Domain Operations
	GetDotfileProvider() DotfileProvider
	GetPackageProvider() PackageProvider

	// State Reconciliation
	ReconcileDotfiles(ctx context.Context) (state.Result, error)
	ReconcilePackages(ctx context.Context) (state.Result, error)
	ReconcileAll(ctx context.Context) (map[string]state.Result, error)

	// Configuration Access (for backward compatibility)
	GetConfig() *config.Config
	GetConfigPath() string
	GetConfigDir() string
	GetHomeDir() string
}

// DotfileProvider provides operations for dotfile management
// This is an alias for the state.Provider interface
type DotfileProvider = state.Provider

// PackageProvider provides operations for package management
// This is an alias for the state.Provider interface
type PackageProvider = state.Provider

// RuntimeStateImpl implements the RuntimeState interface
type RuntimeStateImpl struct {
	configDir string
	homeDir   string

	// Configuration management
	configManager *config.ConfigManager
	configData    *config.Config

	// State management
	reconciler      *state.Reconciler
	dotfileProvider DotfileProvider
	packageProvider PackageProvider

	// Lazy initialization flags
	initialized bool
}

// NewRuntimeState creates a new RuntimeState instance
func NewRuntimeState(configDir, homeDir string) RuntimeState {
	return &RuntimeStateImpl{
		configDir:     configDir,
		homeDir:       homeDir,
		configManager: config.NewConfigManager(configDir),
		reconciler:    state.NewReconciler(),
		initialized:   false,
	}
}

// LoadConfiguration loads configuration from the config directory
func (r *RuntimeStateImpl) LoadConfiguration() error {
	var err error
	r.configData, err = r.configManager.LoadOrCreate()
	if err != nil {
		return err
	}

	// Initialize providers with loaded configuration
	return r.initializeProviders()
}

// SaveConfiguration saves the current configuration
func (r *RuntimeStateImpl) SaveConfiguration() error {
	if r.configData == nil {
		// Create default config if none exists
		r.configData = &config.Config{}
	}
	return r.configManager.Save(r.configData)
}

// ValidateConfiguration validates the current configuration
func (r *RuntimeStateImpl) ValidateConfiguration() error {
	if r.configData == nil {
		if err := r.LoadConfiguration(); err != nil {
			return err
		}
	}

	// Use the config manager's validation capabilities
	// This delegates to the YAMLConfigService validator
	_, err := config.LoadConfig(r.configDir)
	return err
}

// GetDotfileProvider returns the dotfile provider, initializing if necessary
func (r *RuntimeStateImpl) GetDotfileProvider() DotfileProvider {
	if !r.initialized {
		r.ensureInitialized()
	}
	return r.dotfileProvider
}

// GetPackageProvider returns the package provider, initializing if necessary
func (r *RuntimeStateImpl) GetPackageProvider() PackageProvider {
	if !r.initialized {
		r.ensureInitialized()
	}
	return r.packageProvider
}

// ReconcileDotfiles reconciles dotfile state
func (r *RuntimeStateImpl) ReconcileDotfiles(ctx context.Context) (state.Result, error) {
	provider := r.GetDotfileProvider()
	r.reconciler.RegisterProvider("dotfile", provider)

	return r.reconciler.ReconcileProvider(ctx, "dotfile")
}

// ReconcilePackages reconciles package state
func (r *RuntimeStateImpl) ReconcilePackages(ctx context.Context) (state.Result, error) {
	provider := r.GetPackageProvider()
	r.reconciler.RegisterProvider("package", provider)

	return r.reconciler.ReconcileProvider(ctx, "package")
}

// ReconcileAll reconciles all domains
func (r *RuntimeStateImpl) ReconcileAll(ctx context.Context) (map[string]state.Result, error) {
	results := make(map[string]state.Result)

	// Reconcile dotfiles
	dotfileResult, err := r.ReconcileDotfiles(ctx)
	if err != nil {
		return nil, err
	}
	results["dotfile"] = dotfileResult

	// Reconcile packages
	packageResult, err := r.ReconcilePackages(ctx)
	if err != nil {
		return nil, err
	}
	results["package"] = packageResult

	return results, nil
}

// GetConfig returns the loaded configuration
func (r *RuntimeStateImpl) GetConfig() *config.Config {
	if r.configData == nil {
		r.ensureInitialized()
	}
	return r.configData
}

// GetConfigPath returns the path to the configuration file
func (r *RuntimeStateImpl) GetConfigPath() string {
	// We need to build the path since ConfigManager doesn't expose ConfigPath
	return r.configDir + "/plonk.yaml"
}

// GetConfigDir returns the configuration directory
func (r *RuntimeStateImpl) GetConfigDir() string {
	return r.configDir
}

// GetHomeDir returns the home directory
func (r *RuntimeStateImpl) GetHomeDir() string {
	return r.homeDir
}

// ensureInitialized ensures the runtime state is properly initialized
func (r *RuntimeStateImpl) ensureInitialized() error {
	if r.initialized {
		return nil
	}

	if r.configData == nil {
		if err := r.LoadConfiguration(); err != nil {
			return err
		}
	}

	return r.initializeProviders()
}

// initializeProviders initializes the dotfile and package providers
func (r *RuntimeStateImpl) initializeProviders() error {
	if r.configData == nil {
		return nil // Cannot initialize without config
	}

	// Create dotfile provider using a runtime config adapter
	configAdapter := NewRuntimeConfigAdapter(r.configData)
	stateAdapter := state.NewConfigAdapter(configAdapter)
	r.dotfileProvider = state.NewDotfileProvider(r.homeDir, r.configDir, stateAdapter)

	// Create package provider (this will need to be updated once we consolidate)
	// For now, use the existing pattern from commands
	ctx := context.Background()
	lockService := lock.NewYAMLLockService(r.configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	registry := managers.NewManagerRegistry()
	packageProvider, err := registry.CreateMultiProvider(ctx, lockAdapter)
	if err != nil {
		return err
	}
	r.packageProvider = packageProvider

	r.initialized = true
	return nil
}

// RuntimeConfigAdapter adapts config.Config to state.ConfigInterface
// This is a temporary bridge that will be eliminated as we consolidate
type RuntimeConfigAdapter struct {
	config *config.Config
}

// NewRuntimeConfigAdapter creates a new runtime config adapter
func NewRuntimeConfigAdapter(cfg *config.Config) *RuntimeConfigAdapter {
	return &RuntimeConfigAdapter{config: cfg}
}

// GetDotfileTargets returns dotfile targets from config
func (r *RuntimeConfigAdapter) GetDotfileTargets() map[string]string {
	result := make(map[string]string)

	// Auto-discover dotfiles from configured directory
	configDir := config.GetDefaultConfigDirectory()
	resolvedConfig := r.config.Resolve()
	ignorePatterns := resolvedConfig.GetIgnorePatterns()

	// Walk the directory to find all files
	_ = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Get relative path from config dir
		relPath, err := filepath.Rel(configDir, path)
		if err != nil {
			return nil
		}

		// Skip if this is the config directory itself
		if relPath == "." {
			return nil
		}

		// Always skip plonk config files (matching working implementation)
		if relPath == "plonk.yaml" || relPath == "plonk.lock" {
			return nil
		}

		// Skip if this matches any ignore pattern
		for _, pattern := range ignorePatterns {
			// Check if the filename matches
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
			// Check if the relative path matches
			if matched, _ := filepath.Match(pattern, relPath); matched {
				return nil
			}
			// Check if any parent directory in the path matches (for directory exclusions like .git)
			pathParts := strings.Split(relPath, string(filepath.Separator))
			for _, part := range pathParts {
				if matched, _ := filepath.Match(pattern, part); matched {
					return nil
				}
			}
		}

		// Don't process directories, only files
		if info.IsDir() {
			return nil
		}

		// Create destination path based on the file location using ~ notation
		var destPath string
		if strings.HasPrefix(relPath, "config/") {
			// Files in config/ subdirectory map to ~/.config/...
			configSubPath := strings.TrimPrefix(relPath, "config/")
			destPath = "~/.config/" + configSubPath
		} else {
			// Files in root map to ~/.<filename>
			destPath = "~/." + relPath
		}

		// Store the mapping (source in config -> destination in home)
		result[relPath] = destPath

		return nil
	})

	return result
}

// GetHomebrewBrews returns homebrew brews from config (legacy - now in lock file)
func (r *RuntimeConfigAdapter) GetHomebrewBrews() []string {
	return []string{} // Packages now managed via lock file
}

// GetHomebrewCasks returns homebrew casks from config (legacy - now in lock file)
func (r *RuntimeConfigAdapter) GetHomebrewCasks() []string {
	return []string{} // Packages now managed via lock file
}

// GetNPMPackages returns npm packages from config (legacy - now in lock file)
func (r *RuntimeConfigAdapter) GetNPMPackages() []string {
	return []string{} // Packages now managed via lock file
}

// GetIgnorePatterns returns ignore patterns from config
func (r *RuntimeConfigAdapter) GetIgnorePatterns() []string {
	if r.config.IgnorePatterns != nil {
		return r.config.IgnorePatterns
	}
	// Return default ignore patterns
	defaults := config.GetDefaults()
	return defaults.IgnorePatterns
}

// GetExpandDirectories returns expand directories from config
func (r *RuntimeConfigAdapter) GetExpandDirectories() []string {
	if r.config.ExpandDirectories != nil {
		return *r.config.ExpandDirectories
	}
	// Return default expand directories
	defaults := config.GetDefaults()
	return defaults.ExpandDirectories
}
