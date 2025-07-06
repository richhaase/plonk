package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"plonk/internal/utils"
	"plonk/pkg/config"
)

// getRepoDir returns the repository directory path, creating it if it doesn't exist
func getRepoDir() string {
	plonkDir := getPlonkDir()
	repoDir := filepath.Join(plonkDir, "repo")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		// If we can't create it, fall back to plonk dir
		return plonkDir
	}
	
	return repoDir
}

// getBackupsDir returns the backups directory path, respecting config settings
func getBackupsDir() string {
	plonkDir := getPlonkDir()
	
	// Try to load config to check for custom backup location
	cfg, err := config.LoadConfig(plonkDir)
	if err == nil && cfg.Backup.Location != "" && cfg.Backup.Location != "default" {
		// Use custom backup location from config (same as existing getBackupDirectory logic)
		customPath := expandHomeDir(cfg.Backup.Location)
		// Create directory if it doesn't exist
		if err := os.MkdirAll(customPath, 0755); err == nil {
			return customPath
		}
	}
	
	// Use default backups subdirectory (new behavior)
	backupsDir := filepath.Join(plonkDir, "backups")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		// If we can't create it, fall back to old behavior
		if cfg != nil {
			return getBackupDirectory(cfg)
		}
		// Last resort fallback
		return filepath.Join(plonkDir, "backups")
	}
	
	return backupsDir
}

// hasNewDirectoryStructure checks if the new repo/ and backups/ structure exists
func hasNewDirectoryStructure() bool {
	plonkDir := getPlonkDir()
	repoDir := filepath.Join(plonkDir, "repo")
	backupsDir := filepath.Join(plonkDir, "backups")
	
	return utils.FileExists(repoDir) && utils.FileExists(backupsDir)
}

// migrateDirectoryStructure migrates from old flat structure to new subdirectory structure
func migrateDirectoryStructure() error {
	plonkDir := getPlonkDir()
	
	// Skip migration if new structure already exists
	if hasNewDirectoryStructure() {
		return nil
	}
	
	// Create new directories
	repoDir := filepath.Join(plonkDir, "repo")
	backupsDir := filepath.Join(plonkDir, "backups")
	
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}
	
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return fmt.Errorf("failed to create backups directory: %w", err)
	}
	
	// Read all files in plonk directory
	entries, err := os.ReadDir(plonkDir)
	if err != nil {
		return fmt.Errorf("failed to read plonk directory: %w", err)
	}
	
	for _, entry := range entries {
		// Skip the new directories we just created
		if entry.Name() == "repo" || entry.Name() == "backups" {
			continue
		}
		
		// Skip config files - they stay in root
		if entry.Name() == "plonk.yaml" || entry.Name() == "plonk.local.yaml" {
			continue
		}
		
		oldPath := filepath.Join(plonkDir, entry.Name())
		
		// Determine destination based on file type
		var newPath string
		if isBackupFile(entry.Name()) {
			newPath = filepath.Join(backupsDir, entry.Name())
		} else {
			// Everything else goes to repo directory
			newPath = filepath.Join(repoDir, entry.Name())
		}
		
		// Move the file/directory
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", oldPath, newPath, err)
		}
	}
	
	return nil
}

// isBackupFile determines if a filename is a backup file
func isBackupFile(filename string) bool {
	return strings.Contains(filename, ".backup.")
}

// ensureDirectoryStructure ensures the correct directory structure exists
func ensureDirectoryStructure() error {
	// Check if migration is needed
	if !hasNewDirectoryStructure() {
		// Check if there are files in the root that need migration
		plonkDir := getPlonkDir()
		entries, err := os.ReadDir(plonkDir)
		if err != nil {
			// Directory doesn't exist yet, that's fine
			return nil
		}
		
		// Look for files that would need migration
		needsMigration := false
		for _, entry := range entries {
			// Skip config files and directories we would create
			if entry.Name() == "plonk.yaml" || entry.Name() == "plonk.local.yaml" ||
			   entry.Name() == "repo" || entry.Name() == "backups" {
				continue
			}
			needsMigration = true
			break
		}
		
		if needsMigration {
			return migrateDirectoryStructure()
		}
	}
	
	// Ensure directories exist (this will create them if they don't)
	getRepoDir()
	getBackupsDir()
	
	return nil
}