// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
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
	Args:         cobra.MinimumNArgs(1),
	RunE:         runInstall,
	SilenceUsage: true,
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

	// Parse and validate all package specifications
	validator := NewPackageSpecValidator(cfg)
	validSpecs, validationErrors := validator.ValidateInstallSpecs(args)

	// Process each package with prefix parsing
	var allResults []resources.OperationResult

	// Add validation errors to results
	allResults = append(allResults, validationErrors...)

	// Process valid specifications
	for i, spec := range validSpecs {
		// Show progress for multi-package operations
		output.ProgressUpdate(i+1, len(validSpecs), "Installing", spec.String())

		// Configure installation options for this package
		opts := packages.InstallOptions{
			Manager: spec.Manager,
			DryRun:  dryRun,
		}

		// Process this package with configurable timeout
		timeout := time.Duration(cfg.PackageTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		results, err := packages.InstallPackages(ctx, configDir, []string{spec.Name}, opts)
		cancel()

		if err != nil {
			allResults = append(allResults, resources.OperationResult{
				Name:    spec.String(),
				Manager: spec.Manager,
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
		output.Printf("%s %s %s\n", icon, result.Status, result.Name)
		// Show error details for failed operations
		if result.Status == "failed" && result.Error != nil {
			output.Printf("   Error: %s\n", result.Error.Error())
		}
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
