// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/state"
	"github.com/spf13/cobra"
)

var dotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dotfiles across all locations",
	Long: `List dotfiles from your home directory and plonk configuration.

By default, shows missing and managed files with a count of untracked files.
Use --verbose to see all files including the full list of untracked files.

Examples:
  plonk dot list           # Show missing + managed + untracked count
  plonk dot list --verbose # Show all files including full untracked list`,
	RunE: runDotList,
	Args: cobra.NoArgs,
}

func init() {
	dotListCmd.Flags().BoolVarP(&verboseOutput, "verbose", "v", false, "Show all files including untracked ones")
	dotCmd.AddCommand(dotListCmd)
}

var verboseOutput bool

func runDotList(cmd *cobra.Command, args []string) error {
	// No more filter arguments - just verbose flag determines behavior

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

	configDir := config.GetDefaultConfigDirectory()

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

	// Build items list based on verbose flag
	var filteredItems []state.Item

	// Always show missing first (needs action), then managed
	filteredItems = append(filteredItems, result.Missing...)
	filteredItems = append(filteredItems, result.Managed...)

	// Only show untracked files if verbose flag is set
	if verboseOutput {
		filteredItems = append(filteredItems, result.Untracked...)
	}

	// Sort items alphabetically by target path for consistent display
	sort.Slice(filteredItems, func(i, j int) bool {
		targetI := getTargetPath(filteredItems[i])
		targetJ := getTargetPath(filteredItems[j])
		return targetI < targetJ
	})

	// Prepare enhanced output with full item information
	outputData := DotfileListOutput{
		ManagedCount:   len(result.Managed),
		MissingCount:   len(result.Missing),
		UntrackedCount: len(result.Untracked),
		Items:          filteredItems,
		Verbose:        verboseOutput,
	}

	return RenderOutput(outputData, format)
}

// DotfileListOutput represents the output structure for dotfile list commands
type DotfileListOutput struct {
	ManagedCount   int          `json:"managed_count" yaml:"managed_count"`
	MissingCount   int          `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int          `json:"untracked_count" yaml:"untracked_count"`
	Items          []state.Item `json:"items" yaml:"items"`
	Verbose        bool         `json:"verbose" yaml:"verbose"`
}

// TableOutput generates human-friendly table output for dotfiles
func (d DotfileListOutput) TableOutput() string {
	totalCount := d.ManagedCount + d.MissingCount + d.UntrackedCount

	if totalCount == 0 {
		return "No dotfiles found\n"
	}

	// Build summary header
	output := "Dotfiles Summary\n"
	output += "================\n"
	output += fmt.Sprintf("Total: %d files | ✓ Managed: %d | ⚠ Missing: %d | ? Untracked: %d\n\n",
		totalCount, d.ManagedCount, d.MissingCount, d.UntrackedCount)

	if len(d.Items) == 0 && d.UntrackedCount == 0 {
		return output + "No dotfiles found.\n"
	}

	// Find the maximum lengths for proper column alignment
	maxTargetLen := len("Target")
	maxSourceLen := len("Source")

	for _, item := range d.Items {
		// Get target path
		targetPath := getTargetPath(item)
		if len(targetPath) > maxTargetLen {
			maxTargetLen = len(targetPath)
		}

		// Get source path
		sourcePath := getSourcePath(item)
		if len(sourcePath) > maxSourceLen {
			maxSourceLen = len(sourcePath)
		}
	}

	// Add padding
	maxTargetLen += 2
	maxSourceLen += 2

	// Build table header
	output += fmt.Sprintf("  %-6s %-*s %-*s\n", "Status", maxTargetLen, "Target", maxSourceLen, "Source")
	output += fmt.Sprintf("  %-6s %s %s\n", "------",
		repeatChar("-", maxTargetLen),
		repeatChar("-", maxSourceLen))

	// Build table rows
	for _, item := range d.Items {
		status := getStatusIcon(item.State)
		targetPath := getTargetPath(item)
		sourcePath := getSourcePath(item)

		output += fmt.Sprintf("  %-6s %-*s %-*s\n", status, maxTargetLen, targetPath, maxSourceLen, sourcePath)
	}

	// Add untracked summary if not in verbose mode
	if !d.Verbose && d.UntrackedCount > 0 {
		output += fmt.Sprintf("\n%d untracked files (use --verbose to show details)\n", d.UntrackedCount)
	}

	return output
}

// Helper functions for table formatting
func getStatusIcon(itemState state.ItemState) string {
	switch itemState {
	case state.StateManaged:
		return "✓"
	case state.StateMissing:
		return "⚠"
	case state.StateUntracked:
		return "?"
	default:
		return "-"
	}
}

func getTargetPath(item state.Item) string {
	// For dotfiles, the name is the relative path from home
	target := "~/" + item.Name

	// If we have a destination in metadata, use that
	if dest, ok := item.Metadata["destination"].(string); ok {
		target = dest
	}

	return target
}

func getSourcePath(item state.Item) string {
	// For untracked items, there's no source
	if item.State == state.StateUntracked {
		return "-"
	}

	// Check metadata for source
	if source, ok := item.Metadata["source"].(string); ok {
		return source
	}

	// Default to item name if no source found
	return item.Name
}

func repeatChar(char string, count int) string {
	return strings.Repeat(char, count)
}

// StructuredData returns the structured data for serialization
func (d DotfileListOutput) StructuredData() any {
	// For JSON/YAML output, convert to the simpler format
	type structuredOutput struct {
		Summary struct {
			Total     int  `json:"total" yaml:"total"`
			Managed   int  `json:"managed" yaml:"managed"`
			Missing   int  `json:"missing" yaml:"missing"`
			Untracked int  `json:"untracked" yaml:"untracked"`
			Verbose   bool `json:"verbose" yaml:"verbose"`
		} `json:"summary" yaml:"summary"`
		Dotfiles []struct {
			Name   string `json:"name" yaml:"name"`
			State  string `json:"state" yaml:"state"`
			Target string `json:"target" yaml:"target"`
			Source string `json:"source" yaml:"source"`
		} `json:"dotfiles" yaml:"dotfiles"`
	}

	output := structuredOutput{}
	output.Summary.Total = d.ManagedCount + d.MissingCount + d.UntrackedCount
	output.Summary.Managed = d.ManagedCount
	output.Summary.Missing = d.MissingCount
	output.Summary.Untracked = d.UntrackedCount
	output.Summary.Verbose = d.Verbose

	for _, item := range d.Items {
		dotfile := struct {
			Name   string `json:"name" yaml:"name"`
			State  string `json:"state" yaml:"state"`
			Target string `json:"target" yaml:"target"`
			Source string `json:"source" yaml:"source"`
		}{
			Name:   item.Name,
			State:  item.State.String(),
			Target: getTargetPath(item),
			Source: getSourcePath(item),
		}
		output.Dotfiles = append(output.Dotfiles, dotfile)
	}

	return output
}
