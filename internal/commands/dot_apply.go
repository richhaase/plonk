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

var dotApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply dotfile configuration by deploying missing dotfiles",
	Long: `Deploy dotfiles from your plonk configuration to their target locations.

This command will:
- Copy dotfiles from source locations to their destinations
- Create necessary parent directories
- Skip files that are already in sync
- Optionally backup existing files before overwriting

Examples:
  plonk dot apply           # Deploy all configured dotfiles
  plonk dot apply --dry-run # Show what would be deployed without making changes
  plonk dot apply --backup  # Create backups before overwriting existing files`,
	RunE: runDotApply,
}

var (
	dotApplyDryRun bool
	dotApplyBackup bool
)

func init() {
	dotCmd.AddCommand(dotApplyCmd)
	dotApplyCmd.Flags().BoolVar(&dotApplyDryRun, "dry-run", false, "Show what would be deployed without making changes")
	dotApplyCmd.Flags().BoolVar(&dotApplyBackup, "backup", false, "Create backups before overwriting existing files")
}

func runDotApply(cmd *cobra.Command, args []string) error {
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

	// Get dotfile targets
	targets := cfg.GetDotfileTargets()
	if len(targets) == 0 {
		outputData := DotfileApplyOutput{
			DryRun:   dotApplyDryRun,
			Deployed: 0,
			Skipped:  0,
			Actions:  []DotfileAction{},
		}
		return RenderOutput(outputData, format)
	}

	// Process each dotfile
	var actions []DotfileAction
	deployedCount := 0
	skippedCount := 0

	for source, destination := range targets {
		action, err := processDotfile(configDir, homeDir, source, destination, dotApplyDryRun, dotApplyBackup)
		if err != nil {
			return fmt.Errorf("failed to process dotfile %s: %w", source, err)
		}
		
		actions = append(actions, action)
		
		if action.Status == "deployed" {
			deployedCount++
		} else {
			skippedCount++
		}
	}

	// Prepare output
	outputData := DotfileApplyOutput{
		DryRun:   dotApplyDryRun,
		Deployed: deployedCount,
		Skipped:  skippedCount,
		Actions:  actions,
	}

	return RenderOutput(outputData, format)
}

// processDotfile handles the deployment of a single dotfile
func processDotfile(configDir, homeDir, source, destination string, dryRun, backup bool) (DotfileAction, error) {
	// Resolve paths
	sourcePath := filepath.Join(configDir, source)
	destPath := expandPath(destination, homeDir)

	action := DotfileAction{
		Source:      source,
		Destination: destination,
		Status:      "skipped",
		Reason:      "",
	}

	// Check if source exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			action.Status = "error"
			action.Reason = "source file not found"
			return action, nil
		}
		return action, err
	}

	// Check if destination exists
	destInfo, err := os.Stat(destPath)
	if err != nil && !os.IsNotExist(err) {
		return action, err
	}

	// Determine action needed
	if err == nil {
		// Destination exists - check if different
		if sourceInfo.IsDir() != destInfo.IsDir() {
			action.Status = "error"
			action.Reason = "source and destination have different types (file vs directory)"
			return action, nil
		}

		if !sourceInfo.IsDir() {
			// Compare file contents
			same, err := filesAreSame(sourcePath, destPath)
			if err != nil {
				return action, err
			}
			if same {
				action.Status = "skipped"
				action.Reason = "files are identical"
				return action, nil
			}
		}
	}

	// Need to deploy
	action.Status = "deployed"
	action.Reason = "copying from source"

	if dryRun {
		action.Status = "would-deploy"
		return action, nil
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return action, err
	}

	// Backup existing file if requested
	if backup && err == nil {
		backupPath := destPath + ".backup"
		if err := copyFile(destPath, backupPath); err != nil {
			return action, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Copy file or directory
	if sourceInfo.IsDir() {
		err = copyDir(sourcePath, destPath)
	} else {
		err = copyFile(sourcePath, destPath)
	}

	if err != nil {
		return action, err
	}

	return action, nil
}

// expandPath expands ~ to home directory
func expandPath(path, homeDir string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// filesAreSame checks if two files have the same content
func filesAreSame(path1, path2 string) (bool, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return false, err
	}

	content2, err := os.ReadFile(path2)
	if err != nil {
		return false, err
	}

	return string(content1) == string(content2), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, 0644)
}

// copyDir copies a directory from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}

// DotfileApplyOutput represents the output structure for dotfile apply command
type DotfileApplyOutput struct {
	DryRun   bool             `json:"dry_run" yaml:"dry_run"`
	Deployed int              `json:"deployed" yaml:"deployed"`
	Skipped  int              `json:"skipped" yaml:"skipped"`
	Actions  []DotfileAction  `json:"actions" yaml:"actions"`
}

// DotfileAction represents a single dotfile deployment action
type DotfileAction struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Status      string `json:"status" yaml:"status"`
	Reason      string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// TableOutput generates human-friendly table output for dotfile apply
func (d DotfileApplyOutput) TableOutput() string {
	if d.DryRun {
		output := "Dotfile Apply (Dry Run)\n========================\n\n"
		if d.Deployed == 0 && d.Skipped == 0 {
			return output + "No dotfiles configured\n"
		}
		
		output += fmt.Sprintf("Would deploy: %d\n", d.Deployed)
		output += fmt.Sprintf("Would skip: %d\n", d.Skipped)
		
		if len(d.Actions) > 0 {
			output += "\nActions:\n"
			for _, action := range d.Actions {
				status := "❓"
				if action.Status == "would-deploy" {
					status = "🚀"
				} else if action.Status == "skipped" {
					status = "⏭️"
				} else if action.Status == "error" {
					status = "❌"
				}
				
				output += fmt.Sprintf("  %s %s -> %s", status, action.Source, action.Destination)
				if action.Reason != "" {
					output += fmt.Sprintf(" (%s)", action.Reason)
				}
				output += "\n"
			}
		}
		
		return output
	}

	output := "Dotfile Apply\n=============\n\n"
	if d.Deployed == 0 && d.Skipped == 0 {
		return output + "No dotfiles configured\n"
	}

	if d.Deployed > 0 {
		output += fmt.Sprintf("✅ Deployed: %d dotfiles\n", d.Deployed)
	}
	if d.Skipped > 0 {
		output += fmt.Sprintf("⏭️ Skipped: %d dotfiles\n", d.Skipped)
	}

	if len(d.Actions) > 0 {
		output += "\nActions:\n"
		for _, action := range d.Actions {
			status := "❓"
			if action.Status == "deployed" {
				status = "✅"
			} else if action.Status == "skipped" {
				status = "⏭️"
			} else if action.Status == "error" {
				status = "❌"
			}
			
			output += fmt.Sprintf("  %s %s -> %s", status, action.Source, action.Destination)
			if action.Reason != "" {
				output += fmt.Sprintf(" (%s)", action.Reason)
			}
			output += "\n"
		}
	}

	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileApplyOutput) StructuredData() any {
	return d
}