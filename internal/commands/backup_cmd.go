// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/directories"
	"plonk/pkg/config"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup [file...]",
	Short: "Create backups of configuration files",
	Long: `Create timestamped backups of configuration files using the configured backup location.

With no arguments, backs up all files that would be overwritten by 'plonk apply'.
With file arguments, backs up only the specified files.

Examples:
  plonk backup                      # Backup all files that apply would overwrite
  plonk backup ~/.zshrc ~/.vimrc    # Backup specific files`,
	RunE: backupCmdRun,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func backupCmdRun(cmd *cobra.Command, args []string) error {
	dryRun := IsDryRun(cmd)
	if len(args) == 0 {
		// Backup all files that apply would overwrite
		return backupFilesForApplyWithOptions(dryRun)
	} else {
		// Backup specific files
		return backupConfigurationFilesWithOptions(args, dryRun)
	}
}

func backupFilesForApplyWithOptions(dryRun bool) error {
	plonkDir := directories.Default.PlonkDir()

	// Load configuration
	cfg, err := config.LoadConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if dryRun {
		fmt.Printf("Dry-run mode: Showing what files would be backed up (files that apply would overwrite)\n\n")

		// Preview what files would be backed up using apply logic
		return previewBackupsBeforeApply(cfg, []string{})
	}

	// Use the same logic as apply --backup to determine which files to backup
	return createBackupsBeforeApply(cfg, []string{})
}

func backupConfigurationFilesWithOptions(filePaths []string, dryRun bool) error {
	if dryRun {
		fmt.Printf("Dry-run mode: Showing what files would be backed up\n\n")

		backupDir := directories.Default.BackupsDir()
		fmt.Printf("📁 Backup directory: %s\n", backupDir)
		fmt.Printf("🕐 Timestamp format: YYYYMMDD-HHMMSS\n\n")

		for _, filePath := range filePaths {
			expandedPath := directories.Default.ExpandHomeDir(filePath)
			if _, err := os.Stat(expandedPath); err == nil {
				fmt.Printf("📄 %s (would backup to timestamped file)\n", filePath)
			} else {
				fmt.Printf("⚠️  %s (file not found - would skip)\n", filePath)
			}
		}

		fmt.Printf("\nDry-run complete. No files were backed up.\n")
		return nil
	}

	return BackupConfigurationFiles(filePaths)
}

// previewBackupsBeforeApply shows what files would be backed up before apply
func previewBackupsBeforeApply(cfg *config.Config, args []string) error {
	var filesToBackup []string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	if len(args) == 0 {
		// Preview backing up for full apply - check all files that will be written

		// Add dotfiles
		dotfileTargets := cfg.GetDotfileTargets()
		for _, target := range dotfileTargets {
			targetPath := directories.Default.ExpandHomeDir(target)
			filesToBackup = append(filesToBackup, targetPath)
		}

		// Add ZSH configuration files if ZSH config exists
		if len(cfg.ZSH.EnvVars) > 0 || len(cfg.ZSH.Aliases) > 0 ||
			len(cfg.ZSH.Inits) > 0 || len(cfg.ZSH.Completions) > 0 ||
			len(cfg.ZSH.Functions) > 0 || len(cfg.ZSH.ShellOptions) > 0 {
			filesToBackup = append(filesToBackup, filepath.Join(homeDir, ".zshrc"))

			// Only preview backup .zshenv if there are environment variables
			if len(cfg.ZSH.EnvVars) > 0 {
				filesToBackup = append(filesToBackup, filepath.Join(homeDir, ".zshenv"))
			}
		}

		// Add Git configuration file if Git config exists
		if len(cfg.Git.User) > 0 || len(cfg.Git.Core) > 0 || len(cfg.Git.Aliases) > 0 ||
			len(cfg.Git.Color) > 0 || len(cfg.Git.Delta) > 0 || len(cfg.Git.Fetch) > 0 ||
			len(cfg.Git.Pull) > 0 || len(cfg.Git.Push) > 0 || len(cfg.Git.Status) > 0 ||
			len(cfg.Git.Diff) > 0 || len(cfg.Git.Log) > 0 || len(cfg.Git.Init) > 0 ||
			len(cfg.Git.Rerere) > 0 || len(cfg.Git.Branch) > 0 || len(cfg.Git.Rebase) > 0 ||
			len(cfg.Git.Merge) > 0 || len(cfg.Git.Filter) > 0 {
			filesToBackup = append(filesToBackup, filepath.Join(homeDir, ".gitconfig"))
		}

		// Add package configuration files
		filesToBackup = append(filesToBackup, getPackageConfigFilesToBackup(cfg)...)
	} else {
		// Preview backing up for specific package apply - only preview that package's config
		packageName := args[0]
		packageConfig := findPackageConfig(cfg, packageName)
		if packageConfig != "" {
			targetPath := directories.Default.ExpandHomeDir("~/." + packageConfig)
			filesToBackup = append(filesToBackup, targetPath)
		}
	}

	backupDir := directories.Default.BackupsDir()
	fmt.Printf("📁 Backup directory: %s\n", backupDir)
	fmt.Printf("🕐 Timestamp format: YYYYMMDD-HHMMSS\n\n")

	if len(filesToBackup) == 0 {
		fmt.Printf("ℹ️  No files would be backed up\n")
		return nil
	}

	fmt.Printf("Files that would be backed up (%d files):\n", len(filesToBackup))
	for _, filePath := range filesToBackup {
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("📄 %s (would backup to timestamped file)\n", filePath)
		} else {
			fmt.Printf("⚠️  %s (file not found - would skip)\n", filePath)
		}
	}

	fmt.Printf("\nDry-run complete. No files were backed up.\n")
	return nil
}
