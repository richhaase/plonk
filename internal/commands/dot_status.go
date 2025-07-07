// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var dotStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show dotfile status summary",
	Long: `Display a summary of dotfile status relative to plonk configuration.

Shows counts of:
- Managed dotfiles (in config and exist in home)
- Missing dotfiles (in config but not in home)  
- Untracked dotfiles (in home but not in config)

For detailed dotfile lists, use:
  plonk dot list managed
  plonk dot list missing
  plonk dot list untracked`,
	RunE: runDotStatus,
	Args:  cobra.NoArgs,
}

func init() {
	dotCmd.AddCommand(dotStatusCmd)
}

func runDotStatus(cmd *cobra.Command, args []string) error {
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

	// Reconcile dotfiles
	managed, missing, untracked, err := reconcileDotfiles(homeDir, configDir)
	if err != nil {
		return fmt.Errorf("failed to reconcile dotfiles: %w", err)
	}

	// Prepare output structure
	outputData := DotfileStatusOutput{
		Summary: DotfileSummary{
			Managed:   len(managed),
			Missing:   len(missing),
			Untracked: len(untracked),
		},
	}

	return RenderOutput(outputData, format)
}

// DotfileStatusOutput represents the output structure for dotfile status command
type DotfileStatusOutput struct {
	Summary DotfileSummary `json:"summary" yaml:"summary"`
}

// DotfileSummary represents the overall dotfile status summary
type DotfileSummary struct {
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked" yaml:"untracked"`
}

// TableOutput generates human-friendly table output for dotfile status
func (d DotfileStatusOutput) TableOutput() string {
	output := "Dotfile Status\n==============\n\n"
	
	if d.Summary.Managed > 0 {
		output += fmt.Sprintf("✅ %d managed dotfiles\n", d.Summary.Managed)
	} else {
		output += "📁 No managed dotfiles\n"
	}
	
	if d.Summary.Missing > 0 {
		output += fmt.Sprintf("❌ %d missing dotfiles\n", d.Summary.Missing)
	}
	
	if d.Summary.Untracked > 0 {
		output += fmt.Sprintf("🔍 %d untracked dotfiles\n", d.Summary.Untracked)
	}
	
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileStatusOutput) StructuredData() any {
	return d
}