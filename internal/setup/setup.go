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
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// Config represents setup configuration options
type Config struct {
	Interactive bool // Whether to prompt user for confirmations
	Verbose     bool // Whether to show verbose output
}

// SetupWithoutRepo initializes plonk without cloning a repository
func SetupWithoutRepo(ctx context.Context, cfg Config) error {
	fmt.Println("Setting up plonk configuration...")

	// Get plonk directory
	plonkDir := config.GetConfigDir()

	// Check if PLONK_DIR already exists
	if _, err := os.Stat(plonkDir); err == nil {
		fmt.Printf("Plonk directory already exists at: %s\n", plonkDir)
		fmt.Printf("If you want to start fresh, manually delete the directory and run setup again.\n")
		return nil
	}

	// Create plonk directory
	if err := os.MkdirAll(plonkDir, 0750); err != nil {
		return fmt.Errorf("failed to create plonk directory %s: %w", plonkDir, err)
	}
	fmt.Printf("✅ Created plonk directory: %s\n", plonkDir)

	// Create default configuration
	if err := createDefaultConfig(plonkDir); err != nil {
		return fmt.Errorf("failed to create default configuration: %w", err)
	}
	fmt.Printf("✅ Created default configuration files\n")

	// Run doctor checks and install missing tools
	if err := checkAndInstallTools(ctx, cfg); err != nil {
		return fmt.Errorf("failed to install required tools: %w", err)
	}

	fmt.Printf("✅ Setup complete! Run 'plonk status' to see current state.\n")
	return nil
}

// SetupWithRepo clones a repository and sets up plonk
func SetupWithRepo(ctx context.Context, gitRepo string, cfg Config) error {
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
	if err := cloneRepository(gitURL, plonkDir); err != nil {
		// Clean up on failure
		os.RemoveAll(plonkDir)
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	fmt.Printf("✅ Repository cloned successfully\n")

	// Check for existing plonk.yaml
	configFilePath := filepath.Join(plonkDir, "plonk.yaml")
	hasConfig := false
	if _, err := os.Stat(configFilePath); err == nil {
		hasConfig = true
		fmt.Printf("✅ Found existing plonk.yaml configuration\n")
	} else {
		// Create default configuration files
		if err := createDefaultConfig(plonkDir); err != nil {
			return fmt.Errorf("failed to create default configuration: %w", err)
		}
		fmt.Printf("✅ Created default configuration files\n")
	}

	// Run doctor checks and install missing tools
	if err := checkAndInstallTools(ctx, cfg); err != nil {
		return fmt.Errorf("failed to install required tools: %w", err)
	}

	// If we had existing config, run plonk apply
	if hasConfig {
		fmt.Println("Running 'plonk apply' to configure your system...")

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
			fmt.Printf("✅ Applied configuration successfully\n")
		} else {
			fmt.Printf("⚠️ Apply completed with some issues\n")
		}
	}

	fmt.Printf("✅ Setup complete! Your dotfiles are now managed by plonk.\n")
	return nil
}

// createDefaultConfig creates default plonk.yaml and plonk.lock files
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

	// Create empty plonk.lock file
	lockPath := filepath.Join(plonkDir, "plonk.lock")
	lockContent := `# Plonk Lock File
# This file tracks the exact versions of packages and dotfiles managed by plonk.
# Do not edit this file manually - it is maintained automatically by plonk.

version: 1
packages: {}
dotfiles: {}
`

	if err := os.WriteFile(lockPath, []byte(lockContent), 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// checkAndInstallTools runs doctor checks and offers to install missing tools
func checkAndInstallTools(ctx context.Context, cfg Config) error {
	fmt.Println("Checking system requirements...")

	// Run doctor checks
	healthReport := diagnostics.RunHealthChecks()

	return CheckAndInstallToolsFromReport(ctx, healthReport, cfg)
}

// CheckAndInstallToolsFromReport analyzes a health report and installs missing tools
func CheckAndInstallToolsFromReport(ctx context.Context, healthReport diagnostics.HealthReport, cfg Config) error {
	// Find missing package managers
	missingManagers := findMissingPackageManagers(healthReport)

	if len(missingManagers) == 0 {
		fmt.Printf("✅ All required tools are available\n")
		return nil
	}

	fmt.Printf("Missing package managers:\n")
	for _, manager := range missingManagers {
		description := getManagerDescription(manager)
		fmt.Printf("- %s (%s)\n", manager, description)
	}

	if cfg.Interactive {
		if !promptYesNo("Install missing package managers?", true) {
			fmt.Printf("⚠️ Some tools are missing. You can install them later with 'plonk doctor --fix'\n")
			fmt.Printf("💡 Manual installation options:\n")
			for _, manager := range missingManagers {
				instructions := getManualInstallInstructions(manager)
				fmt.Printf("   %s: %s\n", manager, instructions)
			}
			return nil
		}
	}

	// Install missing tools with detailed progress
	successful := 0
	failed := 0

	for _, manager := range missingManagers {
		fmt.Printf("🔄 Installing %s...\n", getManagerDescription(manager))

		if err := installSingleManager(ctx, manager, cfg); err != nil {
			failed++
			fmt.Printf("❌ Failed to install %s: %v\n", manager, err)
			fmt.Printf("💡 Manual installation: %s\n", getManualInstallInstructions(manager))

			// Continue with other managers instead of failing completely
			continue
		}

		successful++
		fmt.Printf("✅ %s installed successfully\n", getManagerDescription(manager))
	}

	// Provide summary
	if failed > 0 {
		fmt.Printf("⚠️ Installation summary: %d successful, %d failed\n", successful, failed)
		fmt.Printf("💡 You can retry failed installations with 'plonk doctor --fix'\n")
		if successful > 0 {
			return nil // Don't treat partial success as failure
		}
		return fmt.Errorf("failed to install %d package managers", failed)
	}

	fmt.Printf("✅ All tools installed successfully (%d managers)\n", successful)
	return nil
}

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
		return "curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | bash"
	case "cargo":
		return "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
	case "npm":
		return "Install Node.js from https://nodejs.org/ or use package manager"
	case "pip":
		return "Install Python from https://python.org/ or use system package manager"
	case "gem":
		return "Install Ruby from https://ruby-lang.org/ or use system package manager"
	case "go":
		return "Install Go from https://golang.org/dl/ or use package manager"
	default:
		return "See official documentation for installation instructions"
	}
}

// installSingleManager installs a single package manager with error handling
func installSingleManager(ctx context.Context, manager string, cfg Config) error {
	switch manager {
	case "homebrew", "brew":
		return installHomebrew(ctx, cfg)
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
			// Parse the Details field which contains entries like "brew: ❌" or "npm: ✅"
			for _, detail := range check.Details {
				if strings.Contains(detail, "❌") {
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
