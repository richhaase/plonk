// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <packages...>",
	Short: "Uninstall packages and remove from plonk management",
	Long: `Uninstall packages from your system and remove them from your lock file.

This command uninstalls packages from your system using the appropriate package
manager and removes them from plonk management. Use prefix syntax to specify
the package manager, or omit the prefix to use your default manager.

Examples:
  plonk uninstall htop                    # Uninstall htop using default manager
  plonk uninstall git neovim              # Uninstall multiple packages
  plonk uninstall brew:git                # Uninstall git specifically with Homebrew
  plonk uninstall npm:lodash              # Uninstall lodash from npm global packages
  plonk uninstall --dry-run htop          # Preview what would be uninstalled`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runUninstall,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Common flags
	uninstallCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags (only common flags now)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get config directory and load configuration
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Parse and validate all package specifications
	validationResult := packages.ValidateSpecs(args, packages.ValidationModeUninstall, "")

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

	// Process each package with prefix parsing
	var allResults []resources.OperationResult

	// Add validation errors to results
	allResults = append(allResults, validationErrors...)

	// Create spinner manager for all uninstall operations
	spinnerManager := output.NewSpinnerManager(len(validationResult.Valid))

	// Process valid specifications
	for _, spec := range validationResult.Valid {
		// Start spinner for package uninstallation
		spinner := spinnerManager.StartSpinner("Uninstalling", spec.String())

		// Configure uninstallation options for this package
		// Pass empty manager if not specified to let UninstallPackages determine it
		opts := packages.UninstallOptions{
			Manager: spec.Manager,
			DryRun:  dryRun,
		}

		// Process this package with configurable timeout
		t := config.GetTimeouts(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), t.Package)
		results, err := packages.UninstallPackages(ctx, configDir, []string{spec.Name}, opts)
		cancel()

		if err != nil {
			result := resources.OperationResult{
				Name:    spec.String(),
				Manager: spec.Manager,
				Status:  "failed",
				Error:   err,
			}
			allResults = append(allResults, result)
			spinner.Error(fmt.Sprintf("Failed to uninstall %s: %s", spec.String(), err.Error()))
			continue
		}

		allResults = append(allResults, results...)

		// Show results for uninstalled packages
		for _, result := range results {
			if result.Status == "failed" && result.Error != nil {
				spinner.Error(fmt.Sprintf("Failed to uninstall %s: %s", result.Name, result.Error.Error()))
			} else {
				spinner.Success(fmt.Sprintf("%s %s", result.Status, result.Name))
			}
			break // Only show first result since we're uninstalling one package at a time
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
		Command:    "uninstall",
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
	return resources.ValidateOperationResults(allResults, "uninstall packages")
}
