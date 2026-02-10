// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var packagesCmd = &cobra.Command{
	Use:     "packages",
	Aliases: []string{"p"},
	Short:   "Display package status",
	Long: `Display the status of all plonk-managed packages.

Shows:
- All managed packages
- Missing packages that need to be installed

Examples:
  plonk packages    # Show all managed packages
  plonk p           # Short alias`,
	RunE:         runPackages,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(packagesCmd)
}

func runPackages(cmd *cobra.Command, args []string) error {
	// Get directories
	configDir := config.GetDefaultConfigDirectory()
	ctx := context.Background()

	// Get package status from lock file
	pkgResult, err := getPackageStatus(ctx, configDir)
	if err != nil {
		return err
	}

	// Convert to output format
	outputResult := output.Result{
		Domain:  "package",
		Managed: pkgResult.Managed,
		Missing: pkgResult.Missing,
		Errors:  pkgResult.Errors,
	}

	// Prepare output
	outputData := output.PackagesStatusOutput{
		Result: outputResult,
	}

	// Create formatter and render
	formatter := output.NewPackagesStatusFormatter(outputData)
	output.RenderOutput(formatter)
	return nil
}
