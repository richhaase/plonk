// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"

	"github.com/spf13/cobra"
)

var dotReAddCmd = &cobra.Command{
	Use:   "re-add <dotfile>",
	Short: "Update an existing managed dotfile from system changes",
	Long: `Update a dotfile that is already managed by plonk with changes from the system.

This command allows you to sync changes you made directly to system dotfiles
back into your plonk configuration. The dotfile must already be managed by plonk.

Examples:
  plonk dot re-add ~/.vimrc           # Update vimrc from system changes
  plonk dot re-add ~/.config/nvim/    # Update nvim config from system changes`,
	RunE: runDotReAdd,
	Args: cobra.ExactArgs(1),
}

func init() {
	dotCmd.AddCommand(dotReAddCmd)
}

func runDotReAdd(cmd *cobra.Command, args []string) error {
	dotfilePath := args[0]

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

	// Resolve and validate dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		return err
	}

	// Check if dotfile exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return fmt.Errorf("dotfile does not exist: %s", resolvedPath)
	}

	// Load existing configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Generate source and destination paths
	source, destination := generatePaths(resolvedPath, homeDir)

	// Check if already managed using auto-discovery
	adapter := config.NewConfigAdapter(cfg)
	dotfileTargets := adapter.GetDotfileTargets()
	if _, exists := dotfileTargets[source]; !exists {
		return fmt.Errorf("dotfile is not currently managed by plonk: %s\nUse 'plonk dot add' to add new dotfiles", destination)
	}

	// Copy dotfile to plonk config directory (overwriting existing)
	sourcePath := filepath.Join(configDir, source)
	if err := copyDotfile(resolvedPath, sourcePath); err != nil {
		return fmt.Errorf("failed to copy dotfile: %w", err)
	}

	// Prepare output
	outputData := DotfileReAddOutput{
		Source:      source,
		Destination: destination,
		Action:      "updated",
		Path:        resolvedPath,
	}

	return RenderOutput(outputData, format)
}

// DotfileReAddOutput represents the output structure for dotfile re-add command
type DotfileReAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
}

// TableOutput generates human-friendly table output for dotfile re-add
func (d DotfileReAddOutput) TableOutput() string {
	output := "Dotfile Re-Add\n==============\n\n"
	output += fmt.Sprintf("✅ Updated dotfile in plonk configuration\n")
	output += fmt.Sprintf("   Source: %s\n", d.Source)
	output += fmt.Sprintf("   Destination: %s\n", d.Destination)
	output += fmt.Sprintf("   System File: %s\n", d.Path)
	output += "\nThe system file has been copied to your plonk config directory, overwriting the previous version\n"
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileReAddOutput) StructuredData() any {
	return d
}