package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"plonk/pkg/config"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [file] | --list | --all",
	Short: "Restore configuration files from backups",
	Long: `Restore configuration files from timestamped backups.

Examples:
  plonk restore --list                # List all available backups
  plonk restore ~/.zshrc              # Restore latest backup of .zshrc
  plonk restore ~/.zshrc --timestamp 20241206-143022  # Restore specific backup
  plonk restore --all                 # Restore all files from latest backups`,
	RunE: restoreCmdRun,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().Bool("list", false, "List available backups")
	restoreCmd.Flags().String("timestamp", "", "Restore from specific timestamp")
	restoreCmd.Flags().Bool("all", false, "Restore all files from latest backups")
}

func restoreCmdRun(cmd *cobra.Command, args []string) error {
	listFlag, _ := cmd.Flags().GetBool("list")
	timestampFlag, _ := cmd.Flags().GetString("timestamp")
	allFlag, _ := cmd.Flags().GetBool("all")
	
	if listFlag {
		return runRestoreList()
	} else if allFlag {
		return runRestoreAll()
	} else if len(args) == 1 {
		return runRestoreFile(args[0], timestampFlag)
	} else {
		return fmt.Errorf("please specify --list, --all, or a file to restore")
	}
}

func runRestoreList() error {
	plonkDir := getPlonkDir()
	
	// Load configuration to get backup settings
	cfg, err := config.LoadYAMLConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Get backup directory
	backupDir := getBackupDirectory(cfg)
	
	// Find all backup files
	backupPattern := filepath.Join(backupDir, "*.backup.*")
	backupFiles, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("failed to search for backup files: %w", err)
	}
	
	if len(backupFiles) == 0 {
		fmt.Println("No backups found")
		return nil
	}
	
	// Group backups by original file
	backupGroups := groupBackupsByOriginalFile(backupFiles)
	
	// Display the grouped backups
	displayBackupGroups(backupGroups)
	
	return nil
}

func runRestoreAll() error {
	// This will be implemented later
	return fmt.Errorf("restore --all not implemented yet")
}

func runRestoreFile(filePath, timestamp string) error {
	// This will be implemented later
	return fmt.Errorf("restore file not implemented yet")
}

// groupBackupsByOriginalFile groups backup files by their original file path
func groupBackupsByOriginalFile(backupFiles []string) map[string][]string {
	backupGroups := make(map[string][]string)
	
	for _, backupFile := range backupFiles {
		baseName := filepath.Base(backupFile)
		// Parse filename like "zshrc.backup.20241206-143022"
		parts := strings.Split(baseName, ".backup.")
		if len(parts) != 2 {
			continue // Skip malformed backup files
		}
		
		originalFile := backupFilenameToOriginalPath(parts[0])
		backupGroups[originalFile] = append(backupGroups[originalFile], backupFile)
	}
	
	return backupGroups
}

// backupFilenameToOriginalPath converts backup filename to original file path
func backupFilenameToOriginalPath(backupFilename string) string {
	// Convert back to dotfile format (zshrc -> ~/.zshrc)
	if backupFilename == "gitconfig" {
		return "~/.gitconfig"
	}
	return "~/." + backupFilename
}

// displayBackupGroups displays backup files grouped by original file
func displayBackupGroups(backupGroups map[string][]string) {
	// Sort original files for consistent output
	var originalFiles []string
	for originalFile := range backupGroups {
		originalFiles = append(originalFiles, originalFile)
	}
	sort.Strings(originalFiles)
	
	fmt.Println("Available backups:")
	fmt.Println("==================")
	
	for _, originalFile := range originalFiles {
		fmt.Printf("\n%s:\n", originalFile)
		
		// Sort backups by timestamp (newest first)
		backups := backupGroups[originalFile]
		sort.Sort(sort.Reverse(sort.StringSlice(backups)))
		
		for _, backup := range backups {
			timestamp := extractTimestampFromBackup(backup)
			if timestamp != "" {
				fmt.Printf("  %s\n", timestamp)
			}
		}
	}
}

// extractTimestampFromBackup extracts timestamp from backup filename
func extractTimestampFromBackup(backupPath string) string {
	baseName := filepath.Base(backupPath)
	parts := strings.Split(baseName, ".backup.")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}