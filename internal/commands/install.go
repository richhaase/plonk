// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"time"

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

Examples:
  plonk install htop                      # Install htop using default manager
  plonk install git neovim ripgrep        # Install multiple packages
  plonk install brew:git                  # Install git specifically with Homebrew
  plonk install npm:lodash                # Install lodash with npm global packages
  plonk install cargo:ripgrep             # Install ripgrep with cargo packages
  plonk install pip:black pip:flake8      # Install Python tools with pip
  plonk install gem:bundler gem:rubocop   # Install Ruby tools with gem
  plonk install go:golang.org/x/tools/cmd/gopls  # Install Go tools with go install
  plonk install --dry-run htop neovim     # Preview what would be installed`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Common flags
	installCmd.Flags().BoolP("dry-run", "n", false, "Show what would be installed without making changes")
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get flags (only common flags now)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get directories and config
	configDir := config.GetDefaultConfigDirectory()
	cfg := config.LoadWithDefaults(configDir)

	// Process each package with prefix parsing
	var allResults []resources.OperationResult

	for i, packageSpec := range args {
		// Show progress for multi-package operations
		output.ProgressUpdate(i+1, len(args), "Installing", packageSpec)
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

		// Use default manager if no prefix specified
		if manager == "" {
			if cfg.DefaultManager != "" {
				manager = cfg.DefaultManager
			} else {
				manager = packages.DefaultManager
			}
		}

		// Validate manager
		if !IsValidManager(manager) {
			errorMsg := FormatNotFoundError("package manager", manager, GetValidManagers())
			allResults = append(allResults, resources.OperationResult{
				Name:    packageSpec,
				Manager: manager,
				Status:  "failed",
				Error:   fmt.Errorf("%s", errorMsg),
			})
			continue
		}

		// Configure installation options for this package
		opts := packages.InstallOptions{
			Manager: manager,
			DryRun:  dryRun,
		}

		// Process this package with configurable timeout
		timeout := time.Duration(cfg.PackageTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		results, err := packages.InstallPackages(ctx, configDir, []string{packageName}, opts)
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
		Command:    "install",
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
	return resources.ValidateOperationResults(allResults, "install packages")
}
