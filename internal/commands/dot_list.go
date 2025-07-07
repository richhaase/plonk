// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/config"

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

	// Reconcile dotfiles
	managed, missing, untracked, err := reconcileDotfiles(homeDir, configDir)
	if err != nil {
		return fmt.Errorf("failed to reconcile dotfiles: %w", err)
	}

	// Select dotfiles based on filter
	var dotfiles []string
	switch filter {
	case "all":
		dotfiles = append(dotfiles, managed...)
		dotfiles = append(dotfiles, untracked...)
	case "managed":
		dotfiles = managed
	case "untracked":
		dotfiles = untracked
	case "missing":
		dotfiles = missing
	}

	// Prepare output
	outputData := DotfileListOutput{
		Filter:   filter,
		Count:    len(dotfiles),
		Dotfiles: dotfiles,
	}

	return RenderOutput(outputData, format)
}

// listDotfiles finds all dotfiles in home directory
func listDotfiles(homeDir string) ([]string, error) {
	var dotfiles []string

	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != "." && entry.Name() != ".." {
			dotfiles = append(dotfiles, entry.Name())
		}
	}

	return dotfiles, nil
}

// getConfigDotfiles gets dotfiles from plonk.yaml
func getConfigDotfiles(configDir string) ([]string, error) {
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return cfg.Dotfiles, nil
}

// reconcileDotfiles performs simple set operations like package reconciliation
func reconcileDotfiles(homeDir, configDir string) ([]string, []string, []string, error) {
	// Get actual dotfiles in home
	actualDotfiles, err := listDotfiles(homeDir)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get config dotfiles
	configDotfiles, err := getConfigDotfiles(configDir)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create lookup sets
	actualSet := make(map[string]bool)
	for _, dotfile := range actualDotfiles {
		actualSet[dotfile] = true
	}

	configSet := make(map[string]bool)
	for _, dotfile := range configDotfiles {
		configSet[dotfile] = true
	}

	// Classify dotfiles
	var managed, missing, untracked []string
	
	for _, dotfile := range configDotfiles {
		if actualSet[dotfile] {
			managed = append(managed, dotfile)
		} else {
			missing = append(missing, dotfile)
		}
	}

	for _, dotfile := range actualDotfiles {
		if !configSet[dotfile] {
			untracked = append(untracked, dotfile)
		}
	}

	return managed, missing, untracked, nil
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