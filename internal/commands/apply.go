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
)

var (
	applyDryRun bool
	applyBackup bool
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply entire plonk configuration (packages and dotfiles)",
	Long: `Apply the complete plonk configuration to your system.

This command will:
1. Install all missing packages from your configuration
2. Deploy all dotfiles from your configuration
3. Report the results for both operations

This is equivalent to running 'plonk pkg apply' followed by 'plonk dot apply'.

Examples:
  plonk apply           # Apply all configuration
  plonk apply --dry-run # Show what would be applied without making changes
  plonk apply --backup  # Create backups before overwriting existing dotfiles`,
	RunE: runApply,
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be applied without making changes")
	applyCmd.Flags().BoolVar(&applyBackup, "backup", false, "Create backups before overwriting existing dotfiles")
}

func runApply(cmd *cobra.Command, args []string) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "apply", "output-format", "invalid output format")
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "apply", "failed to get home directory")
	}

	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Apply packages first
	packageResult, err := applyPackages(configDir, cfg, applyDryRun, format)
	if err != nil {
		return err
	}

	// Apply dotfiles second
	dotfileResult, err := applyDotfiles(configDir, homeDir, cfg, applyDryRun, applyBackup, format)
	if err != nil {
		return err
	}

	// Prepare combined output
	outputData := CombinedApplyOutput{
		DryRun:    applyDryRun,
		Packages:  packageResult,
		Dotfiles:  dotfileResult,
	}

	return RenderOutput(outputData, format)
}

// applyPackages applies package configuration and returns the result
func applyPackages(configDir string, cfg *config.Config, dryRun bool, format OutputFormat) (ApplyOutput, error) {
	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider (multi-manager)
	ctx := context.Background()
	packageProvider := state.NewMultiManagerPackageProvider()
	configAdapter := config.NewConfigAdapter(cfg)
	packageConfigAdapter := config.NewStatePackageConfigAdapter(configAdapter)
	
	// Add Homebrew manager
	homebrewManager := managers.NewHomebrewManager()
	available, err := homebrewManager.IsAvailable(ctx)
	if err != nil {
		return ApplyOutput{}, fmt.Errorf("failed to check homebrew availability: %w", err)
	}
	if available {
		managerAdapter := state.NewManagerAdapter(homebrewManager)
		packageProvider.AddManager("homebrew", managerAdapter, packageConfigAdapter)
	}
	
	// Add NPM manager
	npmManager := managers.NewNpmManager()
	available, err = npmManager.IsAvailable(ctx)
	if err != nil {
		return ApplyOutput{}, fmt.Errorf("failed to check npm availability: %w", err)
	}
	if available {
		managerAdapter := state.NewManagerAdapter(npmManager)
		packageProvider.AddManager("npm", managerAdapter, packageConfigAdapter)
	}
	
	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile package domain to find missing packages
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return ApplyOutput{}, errors.Wrap(err, errors.ErrReconciliation, errors.DomainPackages, "reconcile", "failed to reconcile package state")
	}

	// Group missing packages by manager
	missingByManager := make(map[string][]state.Item)
	for _, item := range result.Missing {
		manager := item.Manager
		if manager == "" {
			manager = "unknown"
		}
		missingByManager[manager] = append(missingByManager[manager], item)
	}

	// Prepare output structure
	outputData := ApplyOutput{
		DryRun:         dryRun,
		TotalMissing:   len(result.Missing),
		Managers:       make([]ManagerApplyResult, 0, len(missingByManager)),
	}

	// Handle case where no packages are missing
	if len(result.Missing) == 0 {
		if format == OutputTable {
			fmt.Println("📦 All packages up to date")
		}
		return outputData, nil
	}

	// Process each manager that has missing packages
	managerInstances := map[string]managers.PackageManager{
		"homebrew": managers.NewHomebrewManager(),
		"npm":      managers.NewNpmManager(),
	}

	for managerName, missingItems := range missingByManager {
		managerInstance, exists := managerInstances[managerName]
		if !exists {
			if format == OutputTable {
				fmt.Printf("📦 %s: Unknown manager, skipping\n", managerName)
			}
			continue
		}

		available, err := managerInstance.IsAvailable(ctx)
		if err != nil {
			return ApplyOutput{}, fmt.Errorf("failed to check if manager '%s' is available: %w", managerName, err)
		}
		if !available {
			if format == OutputTable {
				fmt.Printf("📦 %s: Not available, skipping\n", managerName)
			}
			continue
		}

		// Convert manager name for display
		displayName := managerName
		switch managerName {
		case "homebrew":
			displayName = "Homebrew"
		case "npm":
			displayName = "NPM"
		}

		// Process missing packages for this manager
		managerResult := ManagerApplyResult{
			Name:         displayName,
			MissingCount: len(missingItems),
			Packages:     make([]PackageApplyResult, 0, len(missingItems)),
		}

		for _, item := range missingItems {
			packageResult := PackageApplyResult{
				Name:   item.Name,
				Status: "pending",
			}

			if dryRun {
				packageResult.Status = "would-install"
				if format == OutputTable {
					fmt.Printf("📦 Would install: %s (%s)\n", item.Name, displayName)
				}
				outputData.TotalWouldInstall++
			} else {
				// Actually install the package
				err := managerInstance.Install(ctx, item.Name)
				if err != nil {
					packageResult.Status = "failed"
					// Use structured error for better user messages
					plonkErr := errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", item.Name, "failed to install package")
					packageResult.Error = plonkErr.UserMessage()
					if format == OutputTable {
						fmt.Printf("📦 Failed to install %s: %v\n", item.Name, plonkErr.UserMessage())
					}
					outputData.TotalFailed++
				} else {
					packageResult.Status = "installed"
					if format == OutputTable {
						fmt.Printf("📦 Installed: %s (%s)\n", item.Name, displayName)
					}
					outputData.TotalInstalled++
				}
			}

			managerResult.Packages = append(managerResult.Packages, packageResult)
		}

		outputData.Managers = append(outputData.Managers, managerResult)
	}

	// Output summary for table format
	if format == OutputTable {
		if dryRun {
			fmt.Printf("📦 Package summary: %d packages would be installed\n", outputData.TotalWouldInstall)
		} else {
			fmt.Printf("📦 Package summary: %d installed, %d failed\n", outputData.TotalInstalled, outputData.TotalFailed)
		}
		fmt.Println()
	}

	return outputData, nil
}

// applyDotfiles applies dotfile configuration and returns the result
func applyDotfiles(configDir, homeDir string, cfg *config.Config, dryRun, backup bool, format OutputFormat) (DotfileApplyOutput, error) {
	// Create unified state reconciler
	reconciler := state.NewReconciler()
	
	// Register dotfile provider
	configAdapter := config.NewConfigAdapter(cfg)
	dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
	dotfileProvider := state.NewDotfileProvider(homeDir, configDir, dotfileConfigAdapter)
	reconciler.RegisterProvider("dotfile", dotfileProvider)
	
	// Reconcile dotfile domain to get expanded file list
	ctx := context.Background()
	result, err := reconciler.ReconcileProvider(ctx, "dotfile")
	if err != nil {
		return DotfileApplyOutput{}, errors.Wrap(err, errors.ErrReconciliation, errors.DomainDotfiles, "reconcile", "failed to reconcile dotfile state")
	}
	
	// Process each dotfile from the reconciled state
	var actions []DotfileAction
	deployedCount := 0
	skippedCount := 0
	
	// Process all items (managed, missing, and untracked)
	allItems := append(result.Managed, result.Missing...)
	
	for _, item := range allItems {
		// Get source and destination from metadata
		source, _ := item.Metadata["source"].(string)
		destination, _ := item.Metadata["destination"].(string)
		
		if source == "" || destination == "" {
			continue
		}
		
		action, err := processDotfileForApply(ctx, configDir, homeDir, source, destination, dryRun, backup, format)
		if err != nil {
			return DotfileApplyOutput{}, errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "deploy", source, "failed to process dotfile")
		}
		
		actions = append(actions, action)
		
		if action.Status == "deployed" || action.Status == "would-deploy" {
			deployedCount++
		} else {
			skippedCount++
		}
	}

	// Output summary for table format
	if format == OutputTable {
		if dryRun {
			fmt.Printf("📄 Dotfile summary: %d dotfiles would be deployed, %d would be skipped\n", deployedCount, skippedCount)
		} else {
			fmt.Printf("📄 Dotfile summary: %d deployed, %d skipped\n", deployedCount, skippedCount)
		}
	}

	// Prepare output
	outputData := DotfileApplyOutput{
		DryRun:   dryRun,
		Deployed: deployedCount,
		Skipped:  skippedCount,
		Actions:  actions,
	}

	return outputData, nil
}

// processDotfileForApply handles the deployment of a single dotfile (similar to processDotfile but with table output)
func processDotfileForApply(ctx context.Context, configDir, homeDir, source, destination string, dryRun, backup bool, format OutputFormat) (DotfileAction, error) {
	// Create dotfiles manager and file operations
	manager := dotfiles.NewManager(homeDir, configDir)
	fileOps := dotfiles.NewFileOperations(manager)

	action := DotfileAction{
		Source:      source,
		Destination: destination,
		Status:      "skipped",
		Reason:      "",
	}

	// Validate paths
	if err := manager.ValidatePaths(source, destination); err != nil {
		action.Status = "error"
		action.Reason = err.Error()
		return action, nil
	}

	// Check if source is a directory (should have been expanded)
	if manager.IsDirectory(manager.GetSourcePath(source)) {
		action.Status = "error"
		action.Reason = "unexpected directory (should have been expanded)"
		return action, nil
	}

	// Check if destination exists and is a directory
	destPath := manager.GetDestinationPath(destination)
	if manager.FileExists(destPath) && manager.IsDirectory(destPath) {
		action.Status = "error"
		action.Reason = "destination is a directory, expected file"
		return action, nil
	}

	// Check if file needs update
	needsUpdate, err := fileOps.FileNeedsUpdate(ctx, source, destination)
	if err != nil {
		return action, errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "check", source, "failed to check if file needs update")
	}

	if !needsUpdate {
		action.Status = "skipped"
		action.Reason = "files are identical"
		if format == OutputTable {
			fmt.Printf("📄 Skipped: %s (files are identical)\n", source)
		}
		return action, nil
	}

	// Need to deploy
	action.Status = "deployed"
	action.Reason = "copying from source"
	
	// Add backup indication if backup is requested and file exists
	if backup && manager.FileExists(destPath) {
		action.Reason = "copying from source (with backup)"
	}

	if dryRun {
		action.Status = "would-deploy"
		if format == OutputTable {
			fmt.Printf("📄 Would deploy: %s -> %s\n", source, destination)
		}
		return action, nil
	}

	// Configure copy options
	options := dotfiles.CopyOptions{
		CreateBackup:      backup,
		BackupSuffix:      ".backup",
		OverwriteExisting: true,
	}

	// Copy file using dotfiles operations
	if err := fileOps.CopyFile(ctx, source, destination, options); err != nil {
		return action, errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", source, "failed to copy dotfile")
	}

	if format == OutputTable {
		fmt.Printf("📄 Deployed: %s -> %s\n", source, destination)
	}

	return action, nil
}

// CombinedApplyOutput represents the output structure for the combined apply command
type CombinedApplyOutput struct {
	DryRun   bool                `json:"dry_run" yaml:"dry_run"`
	Packages ApplyOutput         `json:"packages" yaml:"packages"`
	Dotfiles DotfileApplyOutput  `json:"dotfiles" yaml:"dotfiles"`
}

// TableOutput generates human-friendly table output for combined apply
func (c CombinedApplyOutput) TableOutput() string {
	output := ""
	
	if c.DryRun {
		output += "Plonk Apply (Dry Run)\n"
		output += "=====================\n\n"
	} else {
		output += "Plonk Apply\n"
		output += "===========\n\n"
	}
	
	// Summary
	if c.DryRun {
		output += fmt.Sprintf("📦 Packages: %d would be installed\n", c.Packages.TotalWouldInstall)
		output += fmt.Sprintf("📄 Dotfiles: %d would be deployed, %d would be skipped\n", c.Dotfiles.Deployed, c.Dotfiles.Skipped)
	} else {
		output += fmt.Sprintf("📦 Packages: %d installed, %d failed\n", c.Packages.TotalInstalled, c.Packages.TotalFailed)
		output += fmt.Sprintf("📄 Dotfiles: %d deployed, %d skipped\n", c.Dotfiles.Deployed, c.Dotfiles.Skipped)
	}
	
	return output
}

// StructuredData returns the structured data for serialization
func (c CombinedApplyOutput) StructuredData() any {
	return c
}