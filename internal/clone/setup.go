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
		output.Println("Dry run · no changes will be made")
		output.StageUpdate("Setup")
		output.PrintAction("would-clone", gitURL, "to "+plonkDir)
		if _, err := os.Stat(plonkDir); err == nil {
			output.PrintAction("skipped", "Clone", "directory already exists: "+plonkDir)
			output.Println("No changes made.")
			return nil
		}
		output.PrintAction("would-create", "plonk.yaml", "default configuration")
		output.PrintAction("would-check", "Package managers", "required managers from plonk.lock")
		output.PrintAction("would-apply", "Configuration", "run plonk apply after setup")
		output.Println("No changes made.")
		return nil
	}

	output.StageUpdate("Setup")
	output.PrintAction("info", "Repository", gitURL)

	// Check if PLONK_DIR already exists
	if _, err := os.Stat(plonkDir); err == nil {
		return fmt.Errorf("plonk directory already exists at %s; delete it manually and re-run clone if you want to replace it", plonkDir)
	}

	// Clone repository
	output.PrintAction("info", "Cloning repository", "")
	if err := cloneRepository(ctx, gitURL, plonkDir); err != nil {
		// Clean up on failure
		os.RemoveAll(plonkDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	output.PrintAction("done", "Clone repository", plonkDir)

	// Check for existing plonk.yaml
	configFilePath := filepath.Join(plonkDir, "plonk.yaml")
	hasConfig := false
	if _, err := os.Stat(configFilePath); err == nil {
		hasConfig = true
		output.PrintAction("done", "Read configuration", "plonk.yaml loaded")
	} else {
		// Create default configuration file
		if err := createDefaultConfig(plonkDir); err != nil {
			return fmt.Errorf("failed to create default configuration: %w", err)
		}
		hasConfig = true
		output.PrintAction("created", "plonk.yaml", "default configuration")
	}

	if err := SetupFromClonedRepo(ctx, plonkDir, hasConfig); err != nil {
		return err
	}
	output.PrintAction("done", "Setup complete", "")
	return nil
}

// SetupFromClonedRepo performs post-clone setup: detect managers, install, and apply
func SetupFromClonedRepo(ctx context.Context, plonkDir string, hasConfig bool) error {
	repoCfg := config.LoadWithDefaults(plonkDir)

	// Detect required managers from lock file
	output.StageUpdate("Package managers")
	lockPath := filepath.Join(plonkDir, "plonk.lock")
	detectedManagers, err := DetectRequiredManagers(lockPath)
	if err != nil {
		output.PrintAction("warn", "plonk.lock", err.Error())
		output.Printf("No package managers will be installed. Run 'plonk doctor' to check system readiness.\n")
		detectedManagers = []string{} // Empty list
	}

	missingManagers := []string{}
	if len(detectedManagers) > 0 {
		for _, mgr := range detectedManagers {
			output.PrintAction("info", mgr, "required by plonk.lock")
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

		output.StageUpdate("Apply")
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
				output.PrintAction("warn", "Apply", "package errors due to missing managers")
			} else {
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
		} else if !result.Success {
			output.PrintAction("warn", "Apply", "completed with some issues")
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
	missing := missingManagersNow(managers)
	for _, mgr := range managers {
		if !packages.IsSupportedManager(mgr) {
			output.PrintAction("warn", mgr, "unsupported package manager; skipped")
		}
	}
	if len(missing) == 0 {
		output.PrintAction("pass", "Required package managers", "available on PATH")
		return nil
	}
	for _, manager := range missing {
		output.PrintAction("missing", manager, "install manually; see official installation instructions")
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
