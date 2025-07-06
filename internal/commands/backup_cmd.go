package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"plonk/pkg/config"
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
	if len(args) == 0 {
		// Backup all files that apply would overwrite
		return backupFilesForApply()
	} else {
		// Backup specific files
		return BackupConfigurationFiles(args)
	}
}

// backupFilesForApply backs up all files that would be overwritten by apply
func backupFilesForApply() error {
	plonkDir := getPlonkDir()
	
	// Load configuration
	cfg, err := config.LoadYAMLConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Use the same logic as apply --backup to determine which files to backup
	return createBackupsBeforeApply(cfg, []string{})
}