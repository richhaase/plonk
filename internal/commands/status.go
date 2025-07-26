// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/spf13/cobra"
)

// Status command implementation using unified state management system

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show the current state of plonk-managed resources",
	Long: `Display a summary of all packages and dotfiles managed by plonk,
including any that are missing and need to be installed.

Shows:
- Summary of managed packages (count by manager)
- Summary of managed dotfiles (count)
- Missing packages/dotfiles that need to be installed/linked
- Quick overview of plonk's current state

Examples:
  plonk status         # Show managed resource status
  plonk st             # Same as above (alias)
  plonk status -o json # Show as JSON
  plonk status -o yaml # Show as YAML`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Reconcile all domains to get managed state
	ctx := context.Background()
	results, err := orchestrator.ReconcileAll(ctx, homeDir, configDir)
	if err != nil {
		return err
	}

	// Convert results to summary
	summary := resources.ConvertResultsToSummary(results)

	// Prepare focused status output
	outputData := StatusOutput{
		StateSummary: summary,
	}

	return RenderOutput(outputData, format)
}

// Removed - using config.ConfigAdapter instead

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	StateSummary resources.Summary `json:"state_summary" yaml:"state_summary"`
}

// PackageStatus represents package management status
type PackageStatus struct {
	Managed   int            `json:"managed" yaml:"managed"`
	Missing   int            `json:"missing" yaml:"missing"`
	ByManager map[string]int `json:"by_manager" yaml:"by_manager"`
}

// DotfileStatus represents dotfile management status
type DotfileStatus struct {
	Managed int `json:"managed" yaml:"managed"`
	Missing int `json:"missing" yaml:"missing"`
	Linked  int `json:"linked" yaml:"linked"`
}

// ManagedItem represents a currently managed item
type ManagedItem struct {
	Name    string `json:"name" yaml:"name"`
	Domain  string `json:"domain" yaml:"domain"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
}

// TableOutput generates human-friendly table output for status
func (s StatusOutput) TableOutput() string {
	summary := s.StateSummary
	output := "Plonk Status\n\n"

	// Package summary
	packagesByManager := make(map[string]int)
	packagesMissing := 0
	for _, result := range summary.Results {
		if result.Domain == "packages" {
			packagesMissing += len(result.Missing)
			for _, item := range result.Managed {
				packagesByManager[item.Manager]++
			}
		}
	}

	totalPackages := 0
	for _, count := range packagesByManager {
		totalPackages += count
	}

	if totalPackages > 0 || packagesMissing > 0 {
		output += fmt.Sprintf("Packages: %d managed", totalPackages)
		if packagesMissing > 0 {
			output += fmt.Sprintf(", %d missing", packagesMissing)
		}
		output += "\n"

		for manager, count := range packagesByManager {
			output += fmt.Sprintf("  %s: %d packages\n", manager, count)
		}
		output += "\n"
	}

	// Dotfile summary
	dotfilesManaged := 0
	dotfilesMissing := 0
	dotfilesLinked := 0
	for _, result := range summary.Results {
		if result.Domain == "dotfiles" {
			dotfilesManaged += len(result.Managed)
			dotfilesMissing += len(result.Missing)
			for range result.Managed {
				// For dotfiles, assume managed means linked
				dotfilesLinked++
			}
		}
	}

	if dotfilesManaged > 0 || dotfilesMissing > 0 {
		output += fmt.Sprintf("Dotfiles: %d managed", dotfilesManaged)
		if dotfilesMissing > 0 {
			output += fmt.Sprintf(", %d missing", dotfilesMissing)
		}
		output += "\n"
		output += fmt.Sprintf("  linked: %d files\n", dotfilesLinked)
		if dotfilesMissing > 0 {
			output += fmt.Sprintf("  missing: %d files\n", dotfilesMissing)
		}
		output += "\n"
	}

	// Action item
	if summary.TotalMissing > 0 {
		output += "Run 'plonk apply' to install missing items.\n"
	}

	return output
}

// StructuredData returns the structured data for serialization
func (s StatusOutput) StructuredData() any {
	return s
}
