// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
	"github.com/spf13/cobra"
)

var dotfilesCmd = &cobra.Command{
	Use:     "dotfiles",
	Aliases: []string{"d"},
	Short:   "Display dotfile status",
	Long: `Display the status of all plonk-managed dotfiles.

Shows:
- All managed dotfiles
- Missing dotfiles that need to be deployed
- Drifted dotfiles (modified after deployment)
- Unmanaged dotfiles (with --unmanaged flag)

Examples:
  plonk dotfiles              # Show all managed dotfiles
  plonk dotfiles --missing    # Show only missing dotfiles
  plonk dotfiles --unmanaged  # Show only unmanaged dotfiles
  plonk dotfiles -o json      # Show as JSON
  plonk dotfiles -o yaml      # Show as YAML`,
	RunE:         runDotfiles,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)
	dotfilesCmd.Flags().Bool("unmanaged", false, "Show only unmanaged dotfiles")
	dotfilesCmd.Flags().Bool("untracked", false, "Show only untracked dotfiles (alias for --unmanaged)")
	dotfilesCmd.Flags().Bool("missing", false, "Show only missing dotfiles")
	dotfilesCmd.Flags().Bool("managed", false, "Show only managed dotfiles")
	dotfilesCmd.Flags().BoolP("verbose", "v", false, "Show verbose output (includes untracked items)")
}

func runDotfiles(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := output.ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get filter flags
	showUnmanaged, _ := cmd.Flags().GetBool("unmanaged")
	showUntracked, _ := cmd.Flags().GetBool("untracked")
	showMissing, _ := cmd.Flags().GetBool("missing")
	showManaged, _ := cmd.Flags().GetBool("managed")

	// Handle aliases: --untracked is alias for --unmanaged
	if showUntracked {
		showUnmanaged = true
	}

	// Validate mutually exclusive flags
	if err := validateStatusFlags(showUnmanaged, showMissing); err != nil {
		return err
	}

	// If --managed is set, show only managed (exclude missing and untracked)
	if showManaged {
		showMissing = false
		showUnmanaged = false
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration
	cfg := config.LoadWithDefaults(configDir)
	ctx := context.Background()

	// Reconcile dotfiles
	dotfileResult, err := dotfiles.ReconcileWithConfig(ctx, homeDir, configDir, cfg)
	if err != nil {
		return err
	}

	// Convert to output format
	outputResult := output.Result{
		Domain:    dotfileResult.Domain,
		Managed:   convertItemsToOutput(dotfileResult.Managed),
		Missing:   convertItemsToOutput(dotfileResult.Missing),
		Untracked: convertItemsToOutput(dotfileResult.Untracked),
	}

	// Prepare output
	// If --managed flag is set, we need to filter to show only managed items
	// Otherwise use the existing filter logic
	outputData := output.DotfilesStatusOutput{
		Result:        outputResult,
		ShowMissing:   showMissing,
		ShowUnmanaged: showUnmanaged,
		ConfigDir:     configDir,
	}

	// For --managed flag, we need to adjust the result to show only managed items
	if showManaged {
		outputData.Result.Missing = []output.Item{}
		outputData.Result.Untracked = []output.Item{}
		outputData.ShowMissing = false
		outputData.ShowUnmanaged = false
	}

	// Create formatter and render
	formatter := output.NewDotfilesStatusFormatter(outputData)
	return output.RenderOutput(formatter, format)
}
