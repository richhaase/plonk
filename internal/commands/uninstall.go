// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/richhaase/plonk/internal/config"
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
	Args: cobra.MinimumNArgs(1),
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Common flags
	uninstallCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags (only common flags now)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get config directory and load configuration
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Process each package with prefix parsing
	var allResults []resources.OperationResult

	for _, packageSpec := range args {
		// Parse the package specification
		manager, packageName := ParsePackageSpec(packageSpec)

		// Validate package specification
		if packageName == "" {
			errorMsg := FormatValidationError("package specification", packageSpec, "package name cannot be empty")
			allResults = append(allResults, resources.OperationResult{
				Name:   packageSpec,
				Status: "failed",
				Error:  fmt.Errorf("%s", errorMsg),
			})
			continue
		}

		if manager == "" && packageSpec != packageName {
			// This means there was a colon but empty prefix
			errorMsg := FormatValidationError("package specification", packageSpec, "manager prefix cannot be empty")
			allResults = append(allResults, resources.OperationResult{
				Name:   packageSpec,
				Status: "failed",
				Error:  fmt.Errorf("%s", errorMsg),
			})
			continue
		}

		// Only validate manager if explicitly specified
		if manager != "" && !IsValidManager(manager) {
			errorMsg := FormatNotFoundError("package manager", manager, GetValidManagers())
			allResults = append(allResults, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   fmt.Errorf("%s", errorMsg),
			})
			continue
		}

		// Configure uninstallation options for this package
		// Pass empty manager if not specified to let UninstallPackages determine it
		opts := packages.UninstallOptions{
			Manager: manager,
			DryRun:  dryRun,
		}

		// Process this package with configurable timeout
		timeout := time.Duration(cfg.PackageTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		results, err := packages.UninstallPackages(ctx, configDir, []string{packageName}, opts)
		cancel()

		if err != nil {
			allResults = append(allResults, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   err,
			})
			continue
		}

		allResults = append(allResults, results...)
	}

	// Show progress for each result
	for _, result := range allResults {
		icon := GetStatusIcon(result.Status)
		fmt.Printf("%s %s %s\n", icon, result.Status, result.Name)
	}

	// Create output data using standardized format
	summary := CalculatePackageOperationSummary(allResults)
	outputData := PackageOperationOutput{
		Command:    "uninstall",
		TotalItems: len(allResults),
		Results:    ConvertOperationResults(allResults),
		Summary:    summary,
		DryRun:     dryRun,
	}

	// Render output
	if err := RenderOutput(outputData, format); err != nil {
		return err
	}

	// Check if all operations failed and return appropriate error
	return resources.ValidateOperationResults(allResults, "uninstall packages")
}
