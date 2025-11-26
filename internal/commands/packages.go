// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources/packages"
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
- Unmanaged packages (with --unmanaged flag)

Examples:
  plonk packages              # Show all managed packages
  plonk packages --missing   # Show only missing packages
  plonk packages --unmanaged # Show only unmanaged packages
  plonk packages -o json     # Show as JSON
  plonk packages -o yaml     # Show as YAML`,
	RunE:         runPackages,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(packagesCmd)
	packagesCmd.Flags().Bool("unmanaged", false, "Show only unmanaged packages")
	packagesCmd.Flags().Bool("missing", false, "Show only missing packages")
}

func runPackages(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get filter flags
	showUnmanaged, _ := cmd.Flags().GetBool("unmanaged")
	showMissing, _ := cmd.Flags().GetBool("missing")

	// Validate mutually exclusive flags
	if err := validateStatusFlags(showUnmanaged, showMissing); err != nil {
		return err
	}

	// Get directories
	configDir := config.GetConfigDir()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile packages
	packageResult, err := packages.ReconcileWithConfig(ctx, configDir, cfg)
	if err != nil {
		return err
	}

	// Convert to output format
	outputResult := output.Result{
		Domain:    packageResult.Domain,
		Managed:   convertItemsToOutput(packageResult.Managed),
		Missing:   convertItemsToOutput(packageResult.Missing),
		Untracked: convertItemsToOutput(packageResult.Untracked),
	}

	// Prepare output
	outputData := output.PackagesStatusOutput{
		Result:        outputResult,
		ShowMissing:   showMissing,
		ShowUnmanaged: showUnmanaged,
	}

	// Create formatter and render
	formatter := output.NewPackagesStatusFormatter(outputData)
	return output.RenderOutput(formatter, format)
}
