package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"plonk/internal/utils"
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
	plonkDir := getPlonkDir()
	
	// Load configuration to get backup settings
	cfg, err := config.LoadYAMLConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Get backup directory
	backupDir := getBackupDirectory(cfg)
	
	// Convert file path to backup filename format
	backupFilename := originalPathToBackupFilename(filePath)
	
	// Find backup files for this file
	backupPattern := filepath.Join(backupDir, backupFilename+".backup.*")
	backupFiles, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("failed to search for backup files: %w", err)
	}
	
	if len(backupFiles) == 0 {
		return fmt.Errorf("no backups found for %s", filePath)
	}
	
	// Find the backup to restore
	backupToRestore, err := selectBackupToRestore(backupDir, backupFilename, backupFiles, timestamp)
	if err != nil {
		return err
	}
	
	// Restore the backup file
	if err := restoreBackupToFile(backupToRestore, filePath); err != nil {
		return err
	}
	
	return nil
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

// originalPathToBackupFilename converts original file path to backup filename
func originalPathToBackupFilename(originalPath string) string {
	// Expand ~ to full path first
	expandedPath := expandHomeDir(originalPath)
	
	// Get just the filename without directory
	filename := filepath.Base(expandedPath)
	
	// Remove leading dot (e.g., .zshrc -> zshrc)
	if strings.HasPrefix(filename, ".") {
		filename = filename[1:]
	}
	
	return filename
}

// selectBackupToRestore selects which backup file to restore based on timestamp preference
func selectBackupToRestore(backupDir, backupFilename string, backupFiles []string, timestamp string) (string, error) {
	if timestamp != "" {
		// Look for specific timestamp
		targetBackup := filepath.Join(backupDir, backupFilename+".backup."+timestamp)
		if !utils.FileExists(targetBackup) {
			return "", fmt.Errorf("backup with timestamp %s not found", timestamp)
		}
		return targetBackup, nil
	}
	
	// Use latest backup (sort by timestamp, newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(backupFiles)))
	return backupFiles[0], nil
}

// restoreBackupToFile restores a backup file to the target location
func restoreBackupToFile(backupPath, targetPath string) error {
	// Read backup content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}
	
	// Ensure directory exists for target file
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
	}
	
	// Write backup content to target file
	if err := os.WriteFile(targetPath, backupContent, 0644); err != nil {
		return fmt.Errorf("failed to restore %s: %w", targetPath, err)
	}
	
	// Extract timestamp for user feedback
	usedTimestamp := extractTimestampFromBackup(backupPath)
	fmt.Printf("Restored %s from backup %s\n", targetPath, usedTimestamp)
	
	return nil
}

