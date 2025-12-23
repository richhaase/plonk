// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall <packages...>",
	Aliases: []string{"u"},
	Short:   "Uninstall packages and remove from plonk management",
	Long: `Uninstall packages from your system and remove them from your lock file.

This command uninstalls packages from your system using the appropriate package
manager and removes them from plonk management. Use prefix syntax to specify
the package manager, or omit the prefix to use your default manager.

Examples are generated at runtime based on the configured package managers.`,
	Args:         cobra.MinimumNArgs(1),
	RunE:         runUninstall,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Common flags
	uninstallCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")

	// Dynamic examples based on current manager configuration.
	uninstallCmd.Example = buildUninstallExamples()
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get config directory and load configuration
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Parse and validate all package specifications
	validationResult := packages.ValidateSpecs(args, packages.ValidationModeUninstall, "")

	// Convert validation errors to Results
	var validationErrors []operations.Result
	for _, invalid := range validationResult.Invalid {
		validationErrors = append(validationErrors, operations.Result{
			Name:    invalid.OriginalSpec,
			Manager: "",
			Status:  "failed",
			Error:   invalid.Error,
		})
	}

	// Process each package with prefix parsing
	var allResults []operations.Result

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
			result := operations.Result{
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

	// Create formatter and render
	formatter := output.NewPackageOperationFormatter(outputData)
	output.RenderOutput(formatter)

	// Check if all operations failed and return appropriate error
	return operations.ValidateResults(allResults, "uninstall packages")
}
