// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/dotfiles"
	"plonk/internal/state"

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

	// Create unified state reconciler
	reconciler := state.NewReconciler()
	
	// Register dotfile provider
	dotfileProvider := createDotfileProvider(homeDir, configDir, cfg)
	reconciler.RegisterProvider("dotfile", dotfileProvider)
	
	// Reconcile dotfile domain to get expanded file list
	result, err := reconciler.ReconcileProvider("dotfile")
	if err != nil {
		return fmt.Errorf("failed to reconcile dotfile state: %w", err)
	}
	
	// Process each dotfile from the reconciled state
	var actions []DotfileAction
	deployedCount := 0
	skippedCount := 0
	
	// Process all items (managed, missing, and untracked)
	allItems := append(result.Managed, result.Missing...)
	
	for _, item := range allItems {
		// Get source and destination from metadata
		source, _ := item.Metadata["source"].(string)
		destination, _ := item.Metadata["destination"].(string)
		
		if source == "" || destination == "" {
			continue
		}
		
		action, err := processDotfile(configDir, homeDir, source, destination, dotApplyDryRun, dotApplyBackup)
		if err != nil {
			return fmt.Errorf("failed to process dotfile %s: %w", source, err)
		}
		
		actions = append(actions, action)
		
		if action.Status == "deployed" || action.Status == "would-deploy" {
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
	// Create dotfiles manager and file operations
	manager := dotfiles.NewManager(homeDir, configDir)
	fileOps := dotfiles.NewFileOperations(manager)

	action := DotfileAction{
		Source:      source,
		Destination: destination,
		Status:      "skipped",
		Reason:      "",
	}

	// Validate paths
	if err := manager.ValidatePaths(source, destination); err != nil {
		action.Status = "error"
		action.Reason = err.Error()
		return action, nil
	}

	// Check if source is a directory (should have been expanded)
	if manager.IsDirectory(manager.GetSourcePath(source)) {
		action.Status = "error"
		action.Reason = "unexpected directory (should have been expanded)"
		return action, nil
	}

	// Check if destination exists and is a directory
	destPath := manager.GetDestinationPath(destination)
	if manager.FileExists(destPath) && manager.IsDirectory(destPath) {
		action.Status = "error"
		action.Reason = "destination is a directory, expected file"
		return action, nil
	}

	// Check if file needs update
	needsUpdate, err := fileOps.FileNeedsUpdate(source, destination)
	if err != nil {
		return action, err
	}

	if !needsUpdate {
		action.Status = "skipped"
		action.Reason = "files are identical"
		return action, nil
	}

	// Need to deploy
	action.Status = "deployed"
	action.Reason = "copying from source"
	
	// Add backup indication if backup is requested and file exists
	if backup && manager.FileExists(destPath) {
		action.Reason = "copying from source (with backup)"
	}

	if dryRun {
		action.Status = "would-deploy"
		return action, nil
	}

	// Configure copy options
	options := dotfiles.CopyOptions{
		CreateBackup:      backup,
		BackupSuffix:      ".backup",
		OverwriteExisting: true,
	}

	// Copy file using dotfiles operations
	if err := fileOps.CopyFile(source, destination, options); err != nil {
		return action, err
	}

	return action, nil
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