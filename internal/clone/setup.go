// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
)

// Config represents setup configuration options
type Config struct {
	DryRun bool // Whether to show what would happen without making changes
}

// CloneAndSetup clones a repository and sets up plonk intelligently
func CloneAndSetup(ctx context.Context, gitRepo string, cfg Config) error {
	// Parse and validate git URL
	gitURL, err := parseGitURL(gitRepo)
	if err != nil {
		return fmt.Errorf("invalid git repository: %w", err)
	}

	// Get plonk directory
	plonkDir := config.GetConfigDir()

	// Dry run mode: just show what would happen
	if cfg.DryRun {
		output.Printf("Dry run: would set up plonk with repository: %s\n", gitURL)
		output.Printf("Dry run: would clone to: %s\n", plonkDir)

		// Check if PLONK_DIR already exists
		if _, err := os.Stat(plonkDir); err == nil {
			output.Printf("Dry run: plonk directory already exists at: %s\n", plonkDir)
			output.Printf("Dry run: would skip clone (directory exists)\n")
			return nil
		}

		output.Printf("Dry run: would create default plonk.yaml configuration\n")
		output.Printf("Dry run: would detect required package managers from lock file\n")
		output.Printf("Dry run: would run 'plonk apply' after setup\n")
		output.Printf("Dry run: no changes made\n")
		return nil
	}

	output.Printf("Setting up plonk with repository: %s\n", gitURL)

	// Check if PLONK_DIR already exists
	if _, err := os.Stat(plonkDir); err == nil {
		output.Printf("Plonk directory already exists at: %s\n", plonkDir)
		output.Printf("If you want to clone a repository, manually delete the directory and run setup again.\n")
		return nil
	}

	// Clone repository
	output.StageUpdate("Cloning repository...")
	if err := cloneRepository(gitURL, plonkDir); err != nil {
		// Clean up on failure
		os.RemoveAll(plonkDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	output.Printf("Repository cloned successfully\n")

	// Check for existing plonk.yaml
	configFilePath := filepath.Join(plonkDir, "plonk.yaml")
	hasConfig := false
	if _, err := os.Stat(configFilePath); err == nil {
		hasConfig = true
		output.Printf("Found existing plonk.yaml configuration\n")
	} else {
		// Create default configuration file
		if err := createDefaultConfig(plonkDir); err != nil {
			return fmt.Errorf("failed to create default configuration: %w", err)
		}
		output.Printf("Created default plonk.yaml configuration\n")
	}

	if err := SetupFromClonedRepo(ctx, plonkDir, hasConfig); err != nil {
		return err
	}
	output.Printf("Setup complete! Your dotfiles are now managed by plonk.\n")
	return nil
}

// SetupFromClonedRepo performs post-clone setup: detect managers, install, and apply
func SetupFromClonedRepo(ctx context.Context, plonkDir string, hasConfig bool) error {
	repoCfg := config.LoadWithDefaults(plonkDir)

	// Detect required managers from lock file
	output.StageUpdate("Detecting required package managers...")
	lockPath := filepath.Join(plonkDir, "plonk.lock")
	detectedManagers, err := DetectRequiredManagers(lockPath)
	if err != nil {
		output.Printf("Warning: Could not read lock file: %v\n", err)
		output.Printf("No package managers will be installed. You can run 'plonk init' later if needed.\n")
		detectedManagers = []string{} // Empty list
	}

	missingManagers := []string{}
	if len(detectedManagers) > 0 {
		output.Printf("Detected required package managers from lock file:\n")
		for _, mgr := range detectedManagers {
			output.Printf("- %s\n", getManagerDescription(repoCfg, mgr))
		}

		// Install only detected managers
		var installErr error
		missingManagers, installErr = installDetectedManagers(ctx, repoCfg, detectedManagers)
		if installErr != nil {
			return fmt.Errorf("failed to evaluate required tools: %w", installErr)
		}
		if len(missingManagers) > 0 {
			output.Printf("\nThe package managers listed above are missing. Install them manually and run 'plonk doctor' when ready.\n")
		}
	} else {
		output.Printf("No package managers detected from lock file.\n")
	}

	// Run apply if config exists
	if hasConfig {
		if len(missingManagers) > 0 {
			output.Printf("Some package managers are missing; continuing with 'plonk apply' for everything else.\n")
			output.Printf("After installing the missing managers, re-run 'plonk doctor' and 'plonk apply' to reconcile remaining packages.\n")
		}

		output.StageUpdate("Running plonk apply...")
		homeDir := config.GetHomeDir()
		cfg := repoCfg
		if cfg == nil {
			cfg = config.LoadWithDefaults(plonkDir)
		}
		orch := orchestrator.New(
			orchestrator.WithConfig(cfg),
			orchestrator.WithConfigDir(plonkDir),
			orchestrator.WithHomeDir(homeDir),
			orchestrator.WithDryRun(false),
		)
		result, err := orch.Apply(ctx)
		if err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
		if result.Success {
			output.Printf("Applied configuration successfully\n")
		} else {
			output.Printf("Apply completed with some issues\n")
		}
	}
	return nil
}

// createDefaultConfig creates default plonk.yaml file
func createDefaultConfig(plonkDir string) error {
	// Get default values
	defaults := config.GetDefaults()

	// Create plonk.yaml with defaults
	configContent := fmt.Sprintf(`# Plonk Configuration File
# This file contains your plonk settings. Modify as needed.

# Default package manager to use when installing packages
default_manager: %s

# Timeout settings (in seconds)
operation_timeout: %d
package_timeout: %d
dotfile_timeout: %d

# Directories to expand when listing dotfiles
expand_directories:`, defaults.DefaultManager, defaults.OperationTimeout, defaults.PackageTimeout, defaults.DotfileTimeout)

	// Add expand directories
	for _, dir := range defaults.ExpandDirectories {
		configContent += fmt.Sprintf("\n  - %s", dir)
	}

	configContent += `

# Files and patterns to ignore when discovering dotfiles
ignore_patterns:`

	// Add ignore patterns
	for _, pattern := range defaults.IgnorePatterns {
		configContent += fmt.Sprintf("\n  - %q", pattern)
	}

	configContent += "\n"

	// Write plonk.yaml
	configFilePath := filepath.Join(plonkDir, "plonk.yaml")
	if err := os.WriteFile(configFilePath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Note: The doctor command no longer supports --fix flag.
// Package manager installation is only done by clone command when needed.

// getManagerDescription returns a user-friendly description of the package manager
func getManagerDescription(cfg *config.Config, manager string) string {
	// Use registry metadata for built-in managers
	registry := packages.GetRegistry()
	if meta, ok := registry.GetManagerMetadata(manager); ok && meta.Description != "" {
		return meta.Description
	}

	return fmt.Sprintf("%s package manager", manager)
}

// getManualInstallInstructions returns manual installation instructions
func getManualInstallInstructions(cfg *config.Config, manager string) string {
	// Use registry metadata for built-in managers
	registry := packages.GetRegistry()
	if meta, ok := registry.GetManagerMetadata(manager); ok && meta.InstallHint != "" {
		return meta.InstallHint
	}

	return "See official documentation for installation instructions"
}

// DetectRequiredManagers reads a lock file and returns unique package managers
func DetectRequiredManagers(lockPath string) ([]string, error) {
	lockService := lock.NewYAMLLockService(filepath.Dir(lockPath))
	lockFile, err := lockService.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	// Use a map to track unique managers
	managersMap := make(map[string]bool)

	for _, resource := range lockFile.Resources {
		// Only process package resources
		if resource.Type != "package" {
			continue
		}

		// Extract manager from metadata or ID prefix
		var manager string

		// Try to get manager from metadata first (v2 format)
		if managerVal, ok := resource.Metadata["manager"]; ok {
			if managerStr, ok := managerVal.(string); ok {
				manager = managerStr
			}
		}

		// If not in metadata, extract from ID prefix (fallback)
		if manager == "" && strings.Contains(resource.ID, ":") {
			parts := strings.SplitN(resource.ID, ":", 2)
			manager = parts[0]
		}

		if manager != "" {
			managersMap[manager] = true
		}
	}

	// Convert map to sorted slice
	var managers []string
	for mgr := range managersMap {
		managers = append(managers, mgr)
	}

	return managers, nil
}

// installDetectedManagers evaluates which managers are available and returns the missing ones.
func installDetectedManagers(ctx context.Context, cfgData *config.Config, managers []string) ([]string, error) {
	if len(managers) == 0 {
		return nil, nil
	}

	registry := packages.GetRegistry()

	output.StageUpdate(fmt.Sprintf("Checking package managers (%d total)...", len(managers)))

	// Find which managers are missing
	var missingManagers []string
	for _, mgr := range managers {
		packageManager, err := registry.GetManager(mgr)
		if err != nil {
			output.Printf("Warning: Unknown package manager '%s', skipping\n", mgr)
			missingManagers = append(missingManagers, mgr)
			continue
		}

		available, err := packageManager.IsAvailable(ctx)
		if err != nil {
			output.Printf("Warning: Could not check availability of %s: %v\n", mgr, err)
			missingManagers = append(missingManagers, mgr)
			continue
		}

		if !available {
			missingManagers = append(missingManagers, mgr)
		}
	}

	if len(missingManagers) == 0 {
		output.Printf("All required package managers are already installed\n")
		return nil, nil
	}

	output.Printf("\nMissing package managers (automatic installation not supported):\n")
	for _, manager := range missingManagers {
		output.Printf("- %s\n", getManagerDescription(cfgData, manager))
		output.Printf("  Installation: %s\n", getManualInstallInstructions(cfgData, manager))
	}

	return missingManagers, nil
}
