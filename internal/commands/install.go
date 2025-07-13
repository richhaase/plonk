// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/errors"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <items...>",
	Short: "Add and sync items immediately",
	Long: `Add packages or dotfiles and apply changes in one command.

This is a convenience command that combines 'plonk add' and 'plonk sync':
1. First, it adds the specified items to your configuration (like 'plonk add')
2. Then, it immediately syncs all pending changes (like 'plonk sync')

This is perfect for quickly installing new tools and getting them ready to use.

Packages (detected automatically):
  plonk install htop                    # Add htop and sync all changes
  plonk install git neovim ripgrep      # Add multiple packages and sync
  plonk install git --brew              # Add git via Homebrew and sync

Dotfiles (detected automatically):
  plonk install ~/.zshrc                # Add dotfile and sync all changes
  plonk install ~/.zshrc ~/.vimrc       # Add multiple dotfiles and sync

Mixed operations:
  plonk install git ~/.vimrc            # Add package + dotfile and sync
  plonk install --dry-run git ~/.zshrc  # Preview add + sync operations

Force type interpretation:
  plonk install config --package       # Force 'config' as package

Examples:
  plonk install ripgrep                 # Add ripgrep to config and install it
  plonk install ~/.config/nvim/         # Add nvim config and deploy it
  plonk install git ~/.gitconfig        # Add both package and dotfile, then sync
  plonk install --dry-run htop          # Preview what would be added and synced`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Manager-specific flags (mutually exclusive)
	installCmd.Flags().Bool("brew", false, "Use Homebrew package manager")
	installCmd.Flags().Bool("npm", false, "Use NPM package manager")
	installCmd.Flags().Bool("cargo", false, "Use Cargo package manager")
	installCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo")

	// Type override flags (mutually exclusive)
	installCmd.Flags().Bool("package", false, "Force all items to be treated as packages")
	installCmd.Flags().Bool("dotfile", false, "Force all items to be treated as dotfiles")
	installCmd.MarkFlagsMutuallyExclusive("package", "dotfile")

	// Behavior flags
	installCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added and synced without making changes")
	installCmd.Flags().Bool("backup", false, "Create backups before overwriting existing dotfiles")
	installCmd.Flags().BoolP("force", "f", false, "Force addition even if already managed")

	// Add intelligent completion (same as add command)
	installCmd.ValidArgsFunction = completeAddItems
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "install", "output-format", "invalid output format")
	}

	// Step 1: Add items to configuration (reuse add command logic)
	if format == OutputTable {
		if dryRun {
			fmt.Println("Step 1: Adding items to configuration (dry run)")
			fmt.Println("===============================================")
		} else {
			fmt.Println("Step 1: Adding items to configuration")
			fmt.Println("=====================================")
		}
	}

	err = runAdd(cmd, args)
	if err != nil {
		return errors.Wrap(err, errors.ErrCommandExecution, errors.DomainCommands, "install", "failed to add items")
	}

	if format == OutputTable {
		fmt.Println()
		if dryRun {
			fmt.Println("Step 2: Syncing all changes (dry run)")
			fmt.Println("=====================================")
		} else {
			fmt.Println("Step 2: Syncing all changes")
			fmt.Println("===========================")
		}
	}

	// Step 2: Sync all changes (reuse sync command logic)
	// Create a new command instance for sync to avoid flag conflicts
	syncCmd := &cobra.Command{}

	// Copy relevant flags from install to sync
	if dryRun {
		syncCmd.Flags().Bool("dry-run", true, "")
		syncCmd.Flags().Set("dry-run", "true")
	}

	if backup, _ := cmd.Flags().GetBool("backup"); backup {
		syncCmd.Flags().Bool("backup", true, "")
		syncCmd.Flags().Set("backup", "true")
	}

	err = runSync(syncCmd, []string{})
	if err != nil {
		return errors.Wrap(err, errors.ErrCommandExecution, errors.DomainCommands, "install", "failed to sync changes")
	}

	// For structured output, we could create a combined result
	if format != OutputTable {
		installOutput := InstallOutput{
			DryRun:  dryRun,
			Items:   args,
			Status:  "completed",
			Message: "Items added to configuration and changes synced",
		}

		if dryRun {
			installOutput.Status = "would-complete"
			installOutput.Message = "Items would be added to configuration and changes would be synced"
		}

		return RenderOutput(installOutput, format)
	}

	return nil
}

// InstallOutput represents the output structure for install command
type InstallOutput struct {
	DryRun  bool     `json:"dry_run" yaml:"dry_run"`
	Items   []string `json:"items" yaml:"items"`
	Status  string   `json:"status" yaml:"status"`
	Message string   `json:"message" yaml:"message"`
}

// TableOutput generates human-friendly table output for install
func (i InstallOutput) TableOutput() string {
	// Table output is handled inline in the command
	return ""
}

// StructuredData returns the structured data for serialization
func (i InstallOutput) StructuredData() any {
	return i
}
