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
	"github.com/richhaase/plonk/internal/resources/packages"
)

// Config represents setup configuration options
type Config struct {
	Interactive bool // Whether to prompt user for confirmations
	Verbose     bool // Whether to show verbose output
	NoApply     bool // Whether to skip running apply (for clone)
}

// CloneAndSetup clones a repository and sets up plonk intelligently
func CloneAndSetup(ctx context.Context, gitRepo string, cfg Config) error {
	// Parse and validate git URL
	gitURL, err := parseGitURL(gitRepo)
	if err != nil {
		return fmt.Errorf("invalid git repository: %w", err)
	}

	output.Printf("Setting up plonk with repository: %s\n", gitURL)

	// Get plonk directory
	plonkDir := config.GetConfigDir()

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

	if err := SetupFromClonedRepo(ctx, plonkDir, hasConfig, cfg.NoApply); err != nil {
		return err
	}
	output.Printf("Setup complete! Your dotfiles are now managed by plonk.\n")
	return nil
}

// SetupFromClonedRepo performs post-clone setup: detect managers, install, and apply
func SetupFromClonedRepo(ctx context.Context, plonkDir string, hasConfig bool, noApply bool) error {
	// Detect required managers from lock file
	output.StageUpdate("Detecting required package managers...")
	lockPath := filepath.Join(plonkDir, "plonk.lock")
	detectedManagers, err := DetectRequiredManagers(lockPath)
	if err != nil {
		output.Printf("Warning: Could not read lock file: %v\n", err)
		output.Printf("No package managers will be installed. You can run 'plonk init' later if needed.\n")
		detectedManagers = []string{} // Empty list
	}

	if len(detectedManagers) > 0 {
		output.Printf("Detected required package managers from lock file:\n")
		for _, mgr := range detectedManagers {
			output.Printf("- %s\n", getManagerDescription(mgr))
		}

		// Install only detected managers
		if err := installDetectedManagers(ctx, detectedManagers, Config{}); err != nil {
			return fmt.Errorf("failed to install required tools: %w", err)
		}
	} else {
		output.Printf("No package managers detected from lock file.\n")
	}

	// Optionally run apply
	if hasConfig && !noApply {
		output.StageUpdate("Running plonk apply...")
		homeDir := config.GetHomeDir()
		cfg := config.LoadWithDefaults(plonkDir)
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
func getManagerDescription(manager string) string {
	// Prefer descriptions from default manager configs when available.
	for name, cfg := range config.GetDefaultManagers() {
		if name == manager && cfg.Description != "" {
			return cfg.Description
		}
	}

	return fmt.Sprintf("%s package manager", manager)
}

// getManualInstallInstructions returns manual installation instructions
func getManualInstallInstructions(manager string) string {
	for name, cfg := range config.GetDefaultManagers() {
		if name == manager && cfg.InstallHint != "" {
			return cfg.InstallHint
		}
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

// installDetectedManagers installs package managers in the order provided
func installDetectedManagers(ctx context.Context, managers []string, cfg Config) error {
	if len(managers) == 0 {
		return nil
	}

	registry := packages.NewManagerRegistry()

	output.StageUpdate(fmt.Sprintf("Checking package managers (%d total)...", len(managers)))

	// Find which managers are missing
	var missingManagers []string
	for _, mgr := range managers {
		packageManager, err := registry.GetManager(mgr)
		if err != nil {
			output.Printf("Warning: Unknown package manager '%s', skipping\n", mgr)
			continue
		}

		available, err := packageManager.IsAvailable(ctx)
		if err != nil {
			output.Printf("Warning: Could not check availability of %s: %v\n", mgr, err)
			continue
		}

		if !available {
			missingManagers = append(missingManagers, mgr)
		}
	}

	if len(missingManagers) == 0 {
		output.Printf("All required package managers are already installed\n")
		return nil
	}

	output.Printf("\nMissing package managers (automatic installation not supported):\n")
	for _, manager := range missingManagers {
		output.Printf("- %s\n", getManagerDescription(manager))
		output.Printf("  Installation: %s\n", getManualInstallInstructions(manager))
	}

	return fmt.Errorf(
		"automatic installation of package managers is not supported for security reasons\n" +
			"Please install the missing package managers manually using the instructions above\n" +
			"Run 'plonk doctor' for more detailed installation instructions",
	)
}
