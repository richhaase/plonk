package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"plonk/internal/directories"
	"plonk/pkg/config"
)

// BackupExistingFile creates a timestamped backup of an existing file.
// Returns the backup path if a backup was created, empty string if file doesn't exist.
func BackupExistingFile(filePath string) (string, error) {
	// Check if file exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil // File doesn't exist, no backup needed.
	} else if err != nil {
		return "", fmt.Errorf("failed to check file status: %w", err)
	}

	// Generate backup path with timestamp.
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", filePath, timestamp)

	// Read original file.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read original file: %w", err)
	}

	// Write backup file.
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupPath, nil
}

// BackupConfigurationFiles creates backups of specified files using configuration settings.
func BackupConfigurationFiles(filePaths []string) error {
	// Load configuration to get backup settings.
	plonkDir := directories.Default.PlonkDir()
	cfg, err := config.LoadConfig(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get backup configuration with defaults.
	backupDir := getBackupDirectory(cfg)
	keepCount := getKeepCount(cfg)

	// Ensure backup directory exists.
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup each file.
	for _, filePath := range filePaths {
		if err := backupFileToDirectory(filePath, backupDir, keepCount); err != nil {
			return fmt.Errorf("failed to backup %s: %w", filePath, err)
		}
	}

	return nil
}

// getBackupDirectory returns the configured backup directory with default fallback.
func getBackupDirectory(cfg *config.Config) string {
	if cfg.Backup.Location != "" {
		return directories.Default.ExpandHomeDir(cfg.Backup.Location)
	}

	// Default to ~/.config/plonk/backups.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".plonk/backups" // fallback.
	}

	return filepath.Join(homeDir, ".config", "plonk", "backups")
}

// getKeepCount returns the configured keep count with default fallback.
func getKeepCount(cfg *config.Config) int {
	if cfg.Backup.KeepCount > 0 {
		return cfg.Backup.KeepCount
	}

	return 5 // default keep count.
}

// backupFileToDirectory backs up a file to a specific directory and manages cleanup.
func backupFileToDirectory(filePath, backupDir string, keepCount int) error {
	// Check if file exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, no backup needed.
	} else if err != nil {
		return fmt.Errorf("failed to check file status: %w", err)
	}

	// Get base filename without path.
	baseFilename := filepath.Base(filePath)

	// Remove leading dot for backup filename (e.g., .zshrc -> zshrc).
	if strings.HasPrefix(baseFilename, ".") {
		baseFilename = baseFilename[1:]
	}

	// Generate backup filename with timestamp.
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("%s.backup.%s", baseFilename, timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Read and write backup.
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	// Clean up old backups.
	if err := cleanupOldBackups(backupDir, baseFilename, keepCount); err != nil {
		return fmt.Errorf("failed to cleanup old backups: %w", err)
	}

	return nil
}

// cleanupOldBackups removes old backup files beyond the keep count.
func cleanupOldBackups(backupDir, baseFilename string, keepCount int) error {
	// Find all backup files for this base filename.
	pattern := filepath.Join(backupDir, fmt.Sprintf("%s.backup.*", baseFilename))
	backupFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find backup files: %w", err)
	}

	// Sort by filename (which includes timestamp) to get chronological order.
	sort.Strings(backupFiles)

	// Remove old backups if we exceed keep count.
	if len(backupFiles) > keepCount {
		filesToRemove := backupFiles[:len(backupFiles)-keepCount]
		for _, fileToRemove := range filesToRemove {
			if err := os.Remove(fileToRemove); err != nil {
				return fmt.Errorf("failed to remove old backup %s: %w", fileToRemove, err)
			}
		}
	}

	return nil
}
