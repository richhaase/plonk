// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

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
	if err := cloneRepository(ctx, gitURL, plonkDir); err != nil {
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
		hasConfig = true
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
		output.Printf("No package managers will be installed. Run 'plonk doctor' to check system readiness.\n")
		detectedManagers = []string{} // Empty list
	}

	missingManagers := []string{}
	if len(detectedManagers) > 0 {
		output.Printf("Detected required package managers from lock file:\n")
		for _, mgr := range detectedManagers {
			output.Printf("- %s package manager\n", mgr)
		}

		missingManagers = reportMissingManagers(detectedManagers)
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
		homeDir, err := config.GetHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		result, err := orchestrator.Apply(ctx, orchestrator.Options{
			Config:    repoCfg,
			ConfigDir: plonkDir,
			HomeDir:   homeDir,
		})
		result.Scope = "all"
		output.RenderOutput(result)
		if err != nil {
			hasDotfileErrors := len(result.DotfileErrors) > 0
			hasOnlyPackageErrors := len(result.PackageErrors) > 0 && !hasDotfileErrors

			// Re-evaluate manager availability after apply since earlier package installs
			// may have made some managers available during this run.
			currentlyMissing := missingManagersNow(detectedManagers)

			// Only suppress package errors that are fully explained by currently missing managers.
			if hasOnlyPackageErrors && len(currentlyMissing) > 0 && !hasPackageFailuresFromAvailableManagers(result, currentlyMissing) {
				output.Printf("Apply completed with some package errors (expected due to missing managers)\n")
			} else {
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
		} else if result.Success {
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
dotfile_timeout: %d

# Directories to expand when listing dotfiles
expand_directories:`, defaults.DefaultManager, defaults.OperationTimeout, defaults.DotfileTimeout)

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

	// Sort for deterministic output
	sort.Strings(managers)

	return managers, nil
}

// reportMissingManagers reports unavailable managers and how to install them.
func reportMissingManagers(managers []string) []string {
	if len(managers) == 0 {
		return nil
	}
	output.StageUpdate(fmt.Sprintf("Checking package managers (%d total)...", len(managers)))
	missing := missingManagersNow(managers)
	for _, mgr := range managers {
		if !packages.IsSupportedManager(mgr) {
			output.Printf("Warning: %s is not a supported package manager and will be skipped\n", mgr)
		}
	}
	if len(missing) == 0 {
		output.Printf("All required package managers are already installed\n")
		return nil
	}
	output.Printf("\nMissing package managers (automatic installation not supported):\n")
	for _, manager := range missing {
		output.Printf("- %s package manager\n", manager)
		output.Printf("  Installation: See official documentation for installation instructions\n")
	}
	return missing
}

// missingManagersNow returns managers that are unsupported or unavailable on PATH.
func missingManagersNow(managers []string) []string {
	var missing []string
	for _, manager := range managers {
		if !packages.IsSupportedManager(manager) {
			missing = append(missing, manager)
			continue
		}
		if _, err := exec.LookPath(manager); err != nil {
			missing = append(missing, manager)
		}
	}
	return missing
}

// hasPackageFailuresFromAvailableManagers checks if any package failures
// came from managers that are NOT in the missing list.
func hasPackageFailuresFromAvailableManagers(result output.ApplyResult, missingManagers []string) bool {
	if result.Packages == nil {
		return false
	}
	missingSet := make(map[string]bool, len(missingManagers))
	for _, m := range missingManagers {
		missingSet[m] = true
	}
	for _, mgr := range result.Packages.Managers {
		if missingSet[mgr.Name] {
			continue
		}
		for _, pkg := range mgr.Packages {
			if pkg.Status == "failed" {
				return true
			}
		}
	}
	return false
}
