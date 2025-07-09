// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/dotfiles"
	"plonk/internal/errors"
	"plonk/internal/managers"
	"plonk/internal/state"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	manager string
)

var pkgAddCmd = &cobra.Command{
	Use:   "add [package]",
	Short: "Add package(s) to plonk configuration and install them",
	Long: `Add one or more packages to your plonk.yaml configuration and install them.

With package name:
  plonk pkg add htop              # Add htop using default manager
  plonk pkg add git --manager homebrew  # Add git specifically to homebrew
  plonk pkg add lodash --manager npm     # Add lodash to npm global packages

Without arguments:
  plonk pkg add                   # Add all untracked packages
  plonk pkg add --dry-run         # Preview what would be added`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPkgAdd,
}

func init() {
	pkgCmd.AddCommand(pkgAddCmd)
	pkgAddCmd.Flags().StringVar(&manager, "manager", "", "Package manager to use (homebrew|npm)")
	pkgAddCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")
}

func runPkgAdd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if len(args) == 0 {
		// No package specified - add all untracked packages
		return addAllUntrackedPackages(cmd, dryRun)
	}

	packageName := args[0]

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-add", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Load existing configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Determine which manager to use
	targetManager := manager
	if targetManager == "" {
		targetManager = cfg.Settings.DefaultManager
	}

	// Validate manager
	if targetManager != "homebrew" && targetManager != "npm" {
		return errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "validate", fmt.Sprintf("unsupported manager '%s'. Use: homebrew, npm", targetManager))
	}

	// Check if package is already in config
	if isPackageInConfig(cfg, packageName, targetManager) {
		if format == OutputTable {
			fmt.Printf("Package '%s' is already in %s configuration\n", packageName, targetManager)
		}
		return nil
	}

	// Add package to configuration
	err = addPackageToConfig(cfg, packageName, targetManager)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrConfigParseFailure, errors.DomainConfig, "update", packageName, "failed to add package to config")
	}

	// Save updated configuration
	err = saveConfig(cfg, configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainConfig, "save", "failed to save configuration")
	}

	// Install the package
	packageManagers := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
	}

	mgr := packageManagers[targetManager]
	ctx := context.Background()
	available, err := mgr.IsAvailable(ctx)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", targetManager, "failed to check if manager is available")
	}
	if !available {
		return errors.NewError(errors.ErrManagerUnavailable, errors.DomainPackages, "check", fmt.Sprintf("manager '%s' is not available", targetManager))
	}

	// Check if already installed
	installed, err := mgr.IsInstalled(ctx, packageName)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "check", packageName, "failed to check if package is installed")
	}
	if installed {
		if format == OutputTable {
			fmt.Printf("Package '%s' is already installed in %s\n", packageName, targetManager)
			fmt.Printf("Added to configuration: %s\n", packageName)
		}
	} else {
		// Install the package
		if format == OutputTable {
			fmt.Printf("Installing %s using %s...\n", packageName, targetManager)
		}

		err = mgr.Install(ctx, packageName)
		if err != nil {
			return errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", packageName, "failed to install package")
		}

		if format == OutputTable {
			fmt.Printf("Successfully installed and added to configuration: %s\n", packageName)
		}
	}

	// Prepare structured output
	result := AddOutput{
		Package: packageName,
		Manager: targetManager,
		Action:  "added",
	}

	return RenderOutput(result, format)
}

// isPackageInConfig checks if a package is already in the configuration
func isPackageInConfig(cfg *config.Config, packageName, targetManager string) bool {
	switch targetManager {
	case "homebrew":
		for _, pkg := range cfg.Homebrew {
			if pkg.Name == packageName {
				return true
			}
		}
	case "npm":
		for _, pkg := range cfg.NPM {
			if pkg.Name == packageName || pkg.Package == packageName {
				return true
			}
		}
	}
	return false
}

// addPackageToConfig adds a package to the appropriate section of the configuration
func addPackageToConfig(cfg *config.Config, packageName, targetManager string) error {
	switch targetManager {
	case "homebrew":
		// Add to packages section (unified homebrew packages)
		newPackage := config.HomebrewPackage{
			Name: packageName,
		}
		cfg.Homebrew = append(cfg.Homebrew, newPackage)
	case "npm":
		newPackage := config.NPMPackage{
			Name: packageName,
		}
		cfg.NPM = append(cfg.NPM, newPackage)
	default:
		return errors.NewError(errors.ErrInvalidInput, errors.DomainPackages, "validate", fmt.Sprintf("unsupported manager: %s", targetManager))
	}
	return nil
}

// saveConfig saves the configuration back to plonk.yaml atomically
func saveConfig(cfg *config.Config, configDir string) error {
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Create config directory if it doesn't exist
	err := os.MkdirAll(configDir, 0750)
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainConfig, "create", "failed to create config directory")
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigParseFailure, errors.DomainConfig, "marshal", "failed to marshal config")
	}

	// Write to file atomically
	atomicWriter := dotfiles.NewAtomicFileWriter()
	err = atomicWriter.WriteFile(configPath, data, 0644)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainConfig, "write", "failed to write config file")
	}

	return nil
}

// AddOutput represents the output structure for pkg add command
type AddOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
}

// TableOutput generates human-friendly table output for add command
func (a AddOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

// StructuredData returns the structured data for serialization
func (a AddOutput) StructuredData() any {
	return a
}

// addAllUntrackedPackages adds all untracked packages to the configuration
func addAllUntrackedPackages(cmd *cobra.Command, dryRun bool) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "pkg-add-all", "output-format", "invalid output format")
	}

	// Get directories
	configDir := config.GetDefaultConfigDirectory()

	// Load existing configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Create reconciler to get untracked packages
	reconciler := state.NewReconciler()

	// Create package provider (same as status command)
	ctx := context.Background()
	packageProvider, err := createPackageProvider(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternal, errors.DomainPackages, "create", "failed to create package provider")
	}
	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile to get package states
	summary, err := reconciler.ReconcileAll(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile package states")
	}

	// Find package results and collect untracked packages
	var untrackedPackages []state.Item
	for _, result := range summary.Results {
		if result.Domain == "package" {
			untrackedPackages = append(untrackedPackages, result.Untracked...)
		}
	}

	if len(untrackedPackages) == 0 {
		if format == OutputTable {
			fmt.Println("No untracked packages found")
		}
		return nil
	}

	if dryRun {
		if format == OutputTable {
			fmt.Printf("Would add %d untracked packages:\n\n", len(untrackedPackages))
			for _, pkg := range untrackedPackages {
				fmt.Printf("  %s (%s)\n", pkg.Name, pkg.Manager)
			}
		}
		return nil
	}

	// Add packages to configuration
	addedCount := 0
	for _, pkg := range untrackedPackages {
		if !isPackageInConfig(cfg, pkg.Name, pkg.Manager) {
			err = addPackageToConfig(cfg, pkg.Name, pkg.Manager)
			if err != nil {
				return errors.WrapWithItem(err, errors.ErrConfigParseFailure, errors.DomainConfig, "update", pkg.Name, "failed to add package to config")
			}
			addedCount++
		}
	}

	if addedCount == 0 {
		if format == OutputTable {
			fmt.Println("No packages were added (all were already in configuration)")
		}
		return nil
	}

	// Save updated configuration
	err = saveConfig(cfg, configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainConfig, "save", "failed to save configuration")
	}

	if format == OutputTable {
		fmt.Printf("Successfully added %d packages to configuration\n", addedCount)
	}

	// Prepare structured output
	result := AddAllOutput{
		Added:  addedCount,
		Total:  len(untrackedPackages),
		Action: "added-all",
	}

	return RenderOutput(result, format)
}

// AddAllOutput represents the output structure for pkg add-all command
type AddAllOutput struct {
	Added  int    `json:"added" yaml:"added"`
	Total  int    `json:"total" yaml:"total"`
	Action string `json:"action" yaml:"action"`
}

// TableOutput generates human-friendly table output for add-all command
func (a AddAllOutput) TableOutput() string {
	return "" // Table output is handled in the command logic
}

// StructuredData returns the structured data for serialization
func (a AddAllOutput) StructuredData() any {
	return a
}
