// Copyright (c) 2025 Plonk Contributors
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"plonk/internal/directories"
	"plonk/internal/utils"
	"plonk/pkg/config"

	"github.com/spf13/cobra"
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
	dryRun := IsDryRun(cmd)

	if listFlag {
		return runRestoreList()
	} else if allFlag {
		return runRestoreAllWithOptions(dryRun)
	} else if len(args) == 1 {
		return runRestoreFileWithOptions(args[0], timestampFlag, dryRun)
	} else {
		return fmt.Errorf("please specify --list, --all, or a file to restore")
	}
}

func runRestoreList() error {
	// Ensure plonk is properly set up
	plonkDir := directories.Default.PlonkDir()
	if _, err := config.LoadConfig(plonkDir); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get backup directory using new directory structure
	backupDir := directories.Default.BackupsDir()

	// Check if backup directory exists
	if !utils.FileExists(backupDir) {
		fmt.Println("No backups found - backup directory does not exist")
		return nil
	}

	// Find all backup files
	backupPattern := filepath.Join(backupDir, "*.backup.*")
	backupFiles, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("failed to search for backup files in %s: %w", backupDir, err)
	}

	if len(backupFiles) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	// Group backups by original file
	backupGroups := groupBackupsByOriginalFile(backupFiles)

	if len(backupGroups) == 0 {
		fmt.Println("No valid backups found")
		return nil
	}

	// Display the grouped backups
	displayBackupGroups(backupGroups)

	return nil
}

func runRestoreAll() error {
	// Get backup directory using new directory structure
	backupDir := directories.Default.BackupsDir()

	// Check if backup directory exists
	if !utils.FileExists(backupDir) {
		fmt.Println("No backups found - backup directory does not exist")
		return nil
	}

	// Find all backup files
	backupPattern := filepath.Join(backupDir, "*.backup.*")
	backupFiles, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("failed to search for backup files in %s: %w", backupDir, err)
	}

	if len(backupFiles) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	// Group backups by original file
	backupGroups := groupBackupsByOriginalFile(backupFiles)

	if len(backupGroups) == 0 {
		fmt.Println("No valid backups found")
		return nil
	}

	// Restore latest backup for each file
	restoredCount := 0
	failedCount := 0
	fmt.Printf("Restoring %d file(s) from latest backups...\n", len(backupGroups))

	for originalFile, backups := range backupGroups {
		// Sort backups by timestamp (newest first)
		sort.Sort(sort.Reverse(sort.StringSlice(backups)))
		latestBackup := backups[0]

		// Restore the backup file
		if err := restoreBackupToFile(latestBackup, originalFile); err != nil {
			fmt.Printf("⚠️  Failed to restore %s: %v\n", originalFile, err)
			failedCount++
			continue
		}

		restoredCount++
	}

	// Provide summary feedback
	if restoredCount > 0 {
		fmt.Printf("✅ Successfully restored %d file(s) from backups\n", restoredCount)
	}
	if failedCount > 0 {
		fmt.Printf("❌ Failed to restore %d file(s)\n", failedCount)
	}

	return nil
}

func runRestoreFile(filePath, timestamp string) error {
	// Get backup directory using new directory structure
	backupDir := directories.Default.BackupsDir()

	// Check if backup directory exists
	if !utils.FileExists(backupDir) {
		return fmt.Errorf("no backups found for %s - backup directory does not exist", filePath)
	}

	// Convert file path to backup filename format
	backupFilename := originalPathToBackupFilename(filePath)

	// Find backup files for this file
	backupPattern := filepath.Join(backupDir, backupFilename+".backup.*")
	backupFiles, err := filepath.Glob(backupPattern)
	if err != nil {
		return fmt.Errorf("failed to search for backup files in %s: %w", backupDir, err)
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
		return fmt.Errorf("failed to restore %s: %w", filePath, err)
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
	expandedPath := directories.Default.ExpandHomeDir(originalPath)

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
		return fmt.Errorf("failed to read backup file %s: %w", backupPath, err)
	}

	// Expand home directory if needed
	expandedTargetPath := directories.Default.ExpandHomeDir(targetPath)

	// Ensure directory exists for target file
	targetDir := filepath.Dir(expandedTargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Write backup content to target file
	if err := os.WriteFile(expandedTargetPath, backupContent, 0644); err != nil {
		return fmt.Errorf("failed to write to %s: %w", expandedTargetPath, err)
	}

	// Extract timestamp for user feedback
	usedTimestamp := extractTimestampFromBackup(backupPath)
	fmt.Printf("Restored %s from backup %s\n", targetPath, usedTimestamp)

	return nil
}

// runRestoreAllWithOptions provides dry-run support for restore all
func runRestoreAllWithOptions(dryRun bool) error {
	if dryRun {
		fmt.Println("Dry-run mode: Showing what files would be restored")

		// Get backup directory
		backupDir := directories.Default.BackupsDir()

		// Check if backup directory exists
		if !utils.FileExists(backupDir) {
			fmt.Println("❌ No backup directory found - no files to restore")
			return nil
		}

		// List all backup files and group by original file
		entries, err := os.ReadDir(backupDir)
		if err != nil {
			return fmt.Errorf("failed to read backup directory: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("❌ No backup files found")
			return nil
		}

		backupGroups := make(map[string][]string)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// Extract original file path from backup filename
			originalFile := backupFilenameToOriginalPath(entry.Name())
			if originalFile != "" {
				backupGroups[originalFile] = append(backupGroups[originalFile], entry.Name())
			}
		}

		fmt.Printf("\nFiles that would be restored (%d files):\n", len(backupGroups))
		for originalFile, backups := range backupGroups {
			// Sort backups by timestamp (newest first)
			sort.Sort(sort.Reverse(sort.StringSlice(backups)))
			latestBackup := backups[0]
			timestamp := extractTimestampFromBackup(latestBackup)

			fmt.Printf("📁 %s (from backup: %s)\n", originalFile, timestamp)
		}

		fmt.Println("\nDry-run complete. No files were restored.")
		return nil
	}

	return runRestoreAll()
}

// runRestoreFileWithOptions provides dry-run support for restore file
func runRestoreFileWithOptions(filePath, timestamp string, dryRun bool) error {
	if dryRun {
		fmt.Printf("Dry-run mode: Showing what would happen when restoring %s\n\n", filePath)

		// Get backup directory
		backupDir := directories.Default.BackupsDir()

		// Check if backup directory exists
		if !utils.FileExists(backupDir) {
			fmt.Println("❌ No backup directory found - cannot restore file")
			return nil
		}

		// Find backup files for this path
		backupFilename := originalPathToBackupFilename(filePath)
		entries, err := os.ReadDir(backupDir)
		if err != nil {
			return fmt.Errorf("failed to read backup directory: %w", err)
		}

		var availableBackups []string
		for _, entry := range entries {
			if strings.Contains(entry.Name(), backupFilename) {
				availableBackups = append(availableBackups, entry.Name())
			}
		}

		if len(availableBackups) == 0 {
			fmt.Printf("❌ No backups found for %s\n", filePath)
			return nil
		}

		// Sort backups by timestamp (newest first)
		sort.Sort(sort.Reverse(sort.StringSlice(availableBackups)))

		var targetBackup string
		if timestamp != "" {
			// Find specific timestamp
			for _, backup := range availableBackups {
				if strings.Contains(backup, timestamp) {
					targetBackup = backup
					break
				}
			}
			if targetBackup == "" {
				fmt.Printf("❌ No backup found with timestamp %s for %s\n", timestamp, filePath)
				return nil
			}
		} else {
			// Use latest backup
			targetBackup = availableBackups[0]
		}

		usedTimestamp := extractTimestampFromBackup(targetBackup)
		fmt.Printf("📁 Would restore %s from backup %s\n", filePath, usedTimestamp)
		fmt.Printf("💾 Backup file: %s\n", targetBackup)

		fmt.Println("\nDry-run complete. No files were restored.")
		return nil
	}

	return runRestoreFile(filePath, timestamp)
}
