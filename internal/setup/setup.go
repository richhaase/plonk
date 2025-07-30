// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/diagnostics"
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

	fmt.Printf("Setting up plonk with repository: %s\n", gitURL)

	// Get plonk directory
	plonkDir := config.GetConfigDir()

	// Check if PLONK_DIR already exists
	if _, err := os.Stat(plonkDir); err == nil {
		fmt.Printf("Plonk directory already exists at: %s\n", plonkDir)
		fmt.Printf("If you want to clone a repository, manually delete the directory and run setup again.\n")
		return nil
	}

	// Clone repository
	output.StageUpdate("Cloning repository...")
	if err := cloneRepository(gitURL, plonkDir); err != nil {
		// Clean up on failure
		os.RemoveAll(plonkDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	fmt.Printf("Repository cloned successfully\n")

	// Check for existing plonk.yaml
	configFilePath := filepath.Join(plonkDir, "plonk.yaml")
	hasConfig := false
	if _, err := os.Stat(configFilePath); err == nil {
		hasConfig = true
		fmt.Printf("Found existing plonk.yaml configuration\n")
	} else {
		// Create default configuration file
		if err := createDefaultConfig(plonkDir); err != nil {
			return fmt.Errorf("failed to create default configuration: %w", err)
		}
		fmt.Printf("Created default plonk.yaml configuration\n")
	}

	// For clone command, detect required managers from lock file
	output.StageUpdate("Detecting required package managers...")
	lockPath := filepath.Join(plonkDir, "plonk.lock")
	detectedManagers, err := DetectRequiredManagers(lockPath)
	if err != nil {
		fmt.Printf("Warning: Could not read lock file: %v\n", err)
		fmt.Printf("No package managers will be installed. You can run 'plonk init' later if needed.\n")
		detectedManagers = []string{} // Empty list
	}

	if len(detectedManagers) > 0 {
		fmt.Printf("Detected required package managers from lock file:\n")
		for _, mgr := range detectedManagers {
			fmt.Printf("- %s\n", getManagerDescription(mgr))
		}

		// Install only detected managers
		if err := installDetectedManagers(ctx, detectedManagers, cfg); err != nil {
			return fmt.Errorf("failed to install required tools: %w", err)
		}
	} else {
		fmt.Printf("No package managers detected from lock file.\n")
	}

	// If we had existing config and not skipping apply, run plonk apply
	if hasConfig && !cfg.NoApply {
		output.StageUpdate("Running plonk apply...")

		// Run apply
		homeDir := config.GetHomeDir()
		cfg := config.LoadWithDefaults(plonkDir)

		// Create orchestrator
		orch := orchestrator.New(
			orchestrator.WithConfig(cfg),
			orchestrator.WithConfigDir(plonkDir),
			orchestrator.WithHomeDir(homeDir),
			orchestrator.WithDryRun(false),
		)

		// Run apply
		result, err := orch.Apply(ctx)
		if err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}

		if result.Success {
			fmt.Printf("Applied configuration successfully\n")
		} else {
			fmt.Printf("Apply completed with some issues\n")
		}
	}

	fmt.Printf("Setup complete! Your dotfiles are now managed by plonk.\n")
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
	switch manager {
	case "homebrew", "brew":
		return "Homebrew (macOS/Linux package manager)"
	case "cargo":
		return "Cargo (Rust package manager)"
	case "npm":
		return "npm (Node.js package manager)"
	case "pip":
		return "pip (Python package manager)"
	case "gem":
		return "gem (Ruby package manager)"
	case "go":
		return "go (Go package manager)"
	default:
		return fmt.Sprintf("%s package manager", manager)
	}
}

// getManualInstallInstructions returns manual installation instructions
func getManualInstallInstructions(manager string) string {
	switch manager {
	case "homebrew", "brew":
		return "Visit https://brew.sh for installation instructions (prerequisite)"
	case "cargo":
		return "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
	case "npm":
		return "Install Node.js from https://nodejs.org/ or use brew install node"
	case "pip":
		return "Install Python from https://python.org/ or use brew install python"
	case "gem":
		return "Install Ruby from https://ruby-lang.org/ or use brew install ruby"
	case "go":
		return "Install Go from https://golang.org/dl/ or use brew install go"
	default:
		return "See official documentation for installation instructions"
	}
}

// installSingleManager installs a single package manager with error handling
func installSingleManager(ctx context.Context, manager string, cfg Config) error {
	switch manager {
	case "homebrew", "brew":
		return fmt.Errorf("homebrew must be installed manually as a prerequisite - see https://brew.sh")
	case "cargo":
		return installCargo(ctx, cfg)
	case "npm", "pip", "gem", "go":
		return installLanguagePackage(ctx, manager, cfg)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

// findMissingPackageManagers extracts missing package managers from health report
func findMissingPackageManagers(report diagnostics.HealthReport) []string {
	var missing []string

	for _, check := range report.Checks {
		if check.Category == "package-managers" && check.Name == "Package Manager Availability" {
			// Parse the Details field which contains entries like "brew: not available" or "npm: available"
			for _, detail := range check.Details {
				if strings.Contains(detail, "not available") {
					// Extract manager name before the colon
					parts := strings.Split(detail, ":")
					if len(parts) > 0 {
						managerName := strings.TrimSpace(parts[0])
						missing = append(missing, managerName)
					}
				}
			}
		}
	}

	return missing
}

// installLanguagePackage installs a single language package via plonk's system
func installLanguagePackage(ctx context.Context, manager string, cfg Config) error {
	// Map package manager to package name and description
	var packageName, description string
	switch manager {
	case "npm":
		packageName = "node"
		description = "Node.js (provides npm)"
	case "pip":
		packageName = "python"
		description = "Python (provides pip)"
	case "gem":
		packageName = "ruby"
		description = "Ruby (provides gem)"
	case "go":
		packageName = "go"
		description = "Go language"
	default:
		return fmt.Errorf("unsupported language package manager: %s", manager)
	}

	// Load current config to get default manager
	configDir := config.GetConfigDir()
	currentConfig := config.LoadWithDefaults(configDir)

	// Check if default manager is available
	registry := packages.NewManagerRegistry()
	defaultMgr, err := registry.GetManager(currentConfig.DefaultManager)
	if err != nil {
		return fmt.Errorf("default package manager %s not available: %w", currentConfig.DefaultManager, err)
	}

	available, err := defaultMgr.IsAvailable(ctx)
	if err != nil || !available {
		return fmt.Errorf("default package manager %s is not functional", currentConfig.DefaultManager)
	}

	fmt.Printf("Installing %s via %s...\n", description, currentConfig.DefaultManager)

	// Install the package using plonk's normal installation path (updates lock file)
	opts := packages.InstallOptions{
		Manager: currentConfig.DefaultManager,
		DryRun:  false,
		Force:   false,
	}

	_, err = packages.InstallPackages(ctx, configDir, []string{packageName}, opts)
	if err != nil {
		return fmt.Errorf("failed to install %s via %s: %w", description, currentConfig.DefaultManager, err)
	}

	return nil
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

// installDetectedManagers installs only the specified package managers
func installDetectedManagers(ctx context.Context, managers []string, cfg Config) error {
	if len(managers) == 0 {
		return nil
	}

	output.StageUpdate(fmt.Sprintf("Installing package managers (%d required)...", len(managers)))

	// Run doctor checks to get current state
	healthReport := diagnostics.RunHealthChecks()

	// Find which of the detected managers are missing
	var missingManagers []string
	allMissing := findMissingPackageManagers(healthReport)

	for _, mgr := range managers {
		for _, missing := range allMissing {
			if mgr == missing {
				missingManagers = append(missingManagers, mgr)
				break
			}
		}
	}

	if len(missingManagers) == 0 {
		fmt.Printf("All required package managers are already installed\n")
		return nil
	}

	fmt.Printf("Missing required package managers:\n")
	for _, manager := range missingManagers {
		description := getManagerDescription(manager)
		fmt.Printf("- %s (%s)\n", manager, description)
	}

	if cfg.Interactive {
		if !promptYesNo("Install missing package managers?", true) {
			fmt.Printf("Some required tools are missing. You can install them manually\n")
			return nil
		}
	}

	// Install missing tools
	successful := 0
	failed := 0

	for i, manager := range missingManagers {
		// Show progress for each manager
		output.ProgressUpdate(i+1, len(missingManagers), "Installing", getManagerDescription(manager))

		if err := installSingleManager(ctx, manager, cfg); err != nil {
			failed++
			fmt.Printf("Failed to install %s: %v\n", manager, err)
			fmt.Printf("Manual installation: %s\n", getManualInstallInstructions(manager))
			continue
		}

		successful++
		fmt.Printf("%s installed successfully\n", getManagerDescription(manager))
	}

	if failed > 0 {
		fmt.Printf("Installation summary: %d successful, %d failed\n", successful, failed)
		fmt.Printf("You can retry failed installations manually\n")
		if successful > 0 {
			return nil // Don't treat partial success as failure
		}
		return fmt.Errorf("failed to install %d package managers", failed)
	}

	fmt.Printf("All required package managers installed successfully\n")
	return nil
}
