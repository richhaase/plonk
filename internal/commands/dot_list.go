// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

var dotListCmd = &cobra.Command{
	Use:   "list [filter]",
	Short: "List dotfiles across all locations",
	Long: `List dotfiles from your home directory and plonk configuration.

Available filters:
  (no filter)  List all discovered dotfiles
  managed      List dotfiles managed by plonk configuration
  untracked    List dotfiles in home but not in plonk configuration  
  missing      List dotfiles in configuration but not in home

Examples:
  plonk dot list           # List all dotfiles
  plonk dot list managed   # List only dotfiles in plonk.yaml
  plonk dot list untracked # List dotfiles not tracked by plonk`,
	RunE: runDotList,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	dotCmd.AddCommand(dotListCmd)
}

func runDotList(cmd *cobra.Command, args []string) error {
	// Determine filter type
	filter := "all"
	if len(args) > 0 {
		filter = args[0]
		if filter != "managed" && filter != "untracked" && filter != "missing" && filter != "all" {
			return fmt.Errorf("invalid filter '%s'. Use: managed, untracked, missing, or no filter for all", filter)
		}
	}

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "plonk")

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register dotfile provider
	configAdapter := config.NewConfigAdapter(cfg)
	dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
	dotfileProvider := state.NewDotfileProvider(homeDir, configDir, dotfileConfigAdapter)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	// Reconcile dotfile domain
	ctx := context.Background()
	result, err := reconciler.ReconcileProvider(ctx, "dotfile")
	if err != nil {
		return fmt.Errorf("failed to reconcile dotfile state: %w", err)
	}

	// Filter items based on the requested filter
	var filteredItems []state.Item
	switch filter {
	case "all":
		filteredItems = append(filteredItems, result.Managed...)
		filteredItems = append(filteredItems, result.Untracked...)
	case "managed":
		filteredItems = result.Managed
	case "untracked":
		filteredItems = result.Untracked
	case "missing":
		filteredItems = result.Missing
	}

	// Convert to string slice for output compatibility
	dotfiles := make([]string, len(filteredItems))
	for i, item := range filteredItems {
		dotfiles[i] = item.Name
	}

	// Prepare output
	outputData := DotfileListOutput{
		Filter:   filter,
		Count:    len(dotfiles),
		Dotfiles: dotfiles,
	}

	return RenderOutput(outputData, format)
}

// DotfileListOutput represents the output structure for dotfile list commands
type DotfileListOutput struct {
	Filter   string   `json:"filter" yaml:"filter"`
	Count    int      `json:"count" yaml:"count"`
	Dotfiles []string `json:"dotfiles" yaml:"dotfiles"`
}

// TableOutput generates human-friendly table output for dotfiles
func (d DotfileListOutput) TableOutput() string {
	if d.Count == 0 {
		return "No dotfiles found\n"
	}

	output := fmt.Sprintf("# Dotfiles (%d files)\n", d.Count)
	for _, dotfile := range d.Dotfiles {
		output += fmt.Sprintf("%s\n", dotfile)
	}
	return output + "\n"
}

// StructuredData returns the structured data for serialization
func (d DotfileListOutput) StructuredData() any {
	return d
}