// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <packages...>",
	Short: "Install packages and add them to plonk management",
	Long: `Install packages on your system and add them to your lock file for management.

This command installs packages using the specified package manager and adds them
to your lock file so they can be managed by plonk. Use prefix syntax to specify
the package manager, or omit the prefix to use your default manager.

To install a package manager itself, follow the instructions from 'plonk doctor'
or the install hints defined in your plonk.yaml. Bare manager names no longer
trigger automatic bootstrapping.

Examples are generated at runtime based on the configured package managers.`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runInstall,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Common flags
	installCmd.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")

	// Dynamic examples based on current manager configuration.
	installCmd.Example = buildInstallExamples()
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags (only common flags now)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get directories and config
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Check for package manager self-installation requests before normal processing
	registry := packages.NewManagerRegistry()
	registry.LoadV2Configs(cfg)
	var managerSelfInstallResults []resources.OperationResult
	var remainingArgs []string

	// Create spinner manager for all operations
	spinnerManager := output.NewSpinnerManager(len(args))

	for _, arg := range args {
		// Only treat as manager self-install if it's a bare name (no prefix)
		if !strings.Contains(arg, ":") && registry.HasManager(arg) {
			// Start spinner for manager self-installation
			spinner := spinnerManager.StartSpinner("Installing", arg+" (self-install)")

			// Handle self-installation
			result := handleManagerSelfInstall(context.Background(), arg, dryRun, cfg)
			managerSelfInstallResults = append(managerSelfInstallResults, result)

			// Stop spinner and show result
			if result.Status == "failed" && result.Error != nil {
				spinner.Error(fmt.Sprintf("Failed to install %s: %s", arg, result.Error.Error()))
			} else {
				spinner.Success(fmt.Sprintf("%s %s", result.Status, result.Name))
			}
		} else {
			// Process normally (includes prefixed packages like "brew:npm")
			remainingArgs = append(remainingArgs, arg)
		}
	}

	// Parse and validate remaining package specifications
	defaultManager := packages.DefaultManager
	if cfg != nil && cfg.DefaultManager != "" {
		defaultManager = cfg.DefaultManager
	}

	var validationResult packages.BatchValidationResult
	if len(remainingArgs) > 0 {
		validationResult = packages.ValidateSpecs(remainingArgs, packages.ValidationModeInstall, defaultManager)
	}

	// Convert validation errors to OperationResults
	var validationErrors []resources.OperationResult
	for _, invalid := range validationResult.Invalid {
		validationErrors = append(validationErrors, resources.OperationResult{
			Name:    invalid.OriginalSpec,
			Manager: "",
			Status:  "failed",
			Error:   invalid.Error,
		})
	}

	// Combine all results
	var allResults []resources.OperationResult

	// Add manager self-install results first
	allResults = append(allResults, managerSelfInstallResults...)

	// Add validation errors to results
	allResults = append(allResults, validationErrors...)

	// Process valid specifications
	for _, spec := range validationResult.Valid {
		// Start spinner for package installation
		spinner := spinnerManager.StartSpinner("Installing", spec.String())

		// Configure installation options for this package
		opts := packages.InstallOptions{
			Manager: spec.Manager,
			DryRun:  dryRun,
		}

		// Process this package with configurable timeout
		t := config.GetTimeouts(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), t.Package)
		results, err := packages.InstallPackages(ctx, configDir, []string{spec.Name}, opts)
		cancel()

		if err != nil {
			result := resources.OperationResult{
				Name:    spec.String(),
				Manager: spec.Manager,
				Status:  "failed",
				Error:   err,
			}
			allResults = append(allResults, result)
			spinner.Error(fmt.Sprintf("Failed to install %s: %s", spec.String(), err.Error()))
			continue
		}

		allResults = append(allResults, results...)

		// Show results for installed packages
		for _, result := range results {
			if result.Status == "failed" && result.Error != nil {
				spinner.Error(fmt.Sprintf("Failed to install %s: %s", result.Name, result.Error.Error()))
			} else {
				spinner.Success(fmt.Sprintf("%s %s", result.Status, result.Name))
			}
			break // Only show first result since we're installing one package at a time
		}
	}

	// Note: Results are now shown immediately after each operation via spinners
	// This section is kept for any validation errors that weren't processed above
	for _, result := range validationErrors {
		icon := output.GetStatusIcon(result.Status)
		output.Printf("%s %s %s\n", icon, result.Status, result.Name)
		// Show error details for failed operations
		if result.Status == "failed" && result.Error != nil {
			output.Printf("   Error: %s\n", result.Error.Error())
		}
	}

	// Create output data using standardized format
	summary := calculatePackageOperationSummary(allResults)
	outputData := output.PackageOperationOutput{
		Command:    "install",
		TotalItems: len(allResults),
		Results:    convertOperationResults(allResults),
		Summary:    summary,
		DryRun:     dryRun,
	}

	// Create formatter
	formatter := output.NewPackageOperationFormatter(outputData)
	if err := output.RenderOutput(formatter, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	return resources.ValidateOperationResults(allResults, "install packages")
}

// handleManagerSelfInstall reports that self-installation is no longer supported
func handleManagerSelfInstall(ctx context.Context, managerName string, dryRun bool, cfg *config.Config) resources.OperationResult {
	result := resources.OperationResult{
		Name:    managerName,
		Manager: "self-install",
		Status:  "failed",
		Error: fmt.Errorf(
			"automatic installation of package managers is not supported for security reasons\n"+
				"Please install %s manually or via another package manager\n"+
				"Run 'plonk doctor' for installation instructions",
			managerName,
		),
	}

	return result
}
