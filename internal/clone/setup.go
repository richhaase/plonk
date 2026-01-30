// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/output"
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
	plonkDir := config.GetDefaultConfigDirectory()

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
		return fmt.Errorf("plonk directory already exists at %s; delete it manually and re-run clone if you want to replace it", plonkDir)
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
func getManagerDescription(_ *config.Config, manager string) string {
	return fmt.Sprintf("%s package manager", manager)
}

// getManualInstallInstructions returns manual installation instructions
func getManualInstallInstructions(_ *config.Config, _ string) string {
	return "See official documentation for installation instructions"
}

// DetectRequiredManagers reads a lock file and returns unique package managers
func DetectRequiredManagers(lockPath string) ([]string, error) {
	lockService := lock.NewLockV3Service(filepath.Dir(lockPath))
	lockFile, err := lockService.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	// Extract unique managers from v3 format (packages grouped by manager)
	var managers []string
	for manager := range lockFile.Packages {
		managers = append(managers, manager)
	}

	return managers, nil
}

// installDetectedManagers evaluates which managers are available and returns the missing ones.
func installDetectedManagers(ctx context.Context, cfgData *config.Config, managers []string) ([]string, error) {
	if len(managers) == 0 {
		return nil, nil
	}

	output.StageUpdate(fmt.Sprintf("Checking package managers (%d total)...", len(managers)))

	// Manager binary names
	managerBinaries := map[string]string{
		"brew":  "brew",
		"cargo": "cargo",
		"go":    "go",
		"pnpm":  "pnpm",
		"uv":    "uv",
	}

	// Find which managers are missing
	var missingManagers []string
	for _, mgr := range managers {
		binary := managerBinaries[mgr]
		if binary == "" {
			binary = mgr
		}

		_, err := exec.LookPath(binary)
		if err != nil {
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
