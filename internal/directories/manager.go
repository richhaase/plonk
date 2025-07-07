// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package directories provides centralized directory management for Plonk.
// It handles path resolution, home directory expansion, and ensures proper
// directory structure for configuration files, backups, and repositories.
//
// The package supports environment variable overrides (PLONK_DIR) and
// automatic migration from legacy directory structures.
package directories

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/utils"
	"plonk/pkg/config"
)

// Manager provides centralized directory management for plonk
type Manager struct {
	// Cache computed paths to avoid repeated filesystem operations
	plonkDir   string
	repoDir    string
	backupsDir string
}

// NewManager creates a new directory manager
func NewManager() *Manager {
	return &Manager{}
}

// PlonkDir returns the main plonk directory location
// Defaults to ~/.config/plonk but can be overridden with PLONK_DIR environment variable
func (m *Manager) PlonkDir() string {
	if m.plonkDir != "" {
		return m.plonkDir
	}

	if dir := os.Getenv("PLONK_DIR"); dir != "" {
		m.plonkDir = dir
		return dir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		m.plonkDir = ".plonk"
		return m.plonkDir
	}

	m.plonkDir = filepath.Join(homeDir, ".config", "plonk")
	return m.plonkDir
}

// RepoDir returns the repository directory path, creating it if it doesn't exist
func (m *Manager) RepoDir() string {
	if m.repoDir != "" {
		return m.repoDir
	}

	plonkDir := m.PlonkDir()
	repoDir := filepath.Join(plonkDir, "repo")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		// If we can't create it, fall back to plonk dir
		m.repoDir = plonkDir
		return plonkDir
	}

	m.repoDir = repoDir
	return repoDir
}

// BackupsDir returns the backups directory path, respecting config settings
func (m *Manager) BackupsDir() string {
	if m.backupsDir != "" {
		return m.backupsDir
	}

	plonkDir := m.PlonkDir()

	// Try to load config to check for custom backup location
	cfg, err := config.LoadConfig(plonkDir)
	if err == nil && cfg.Backup.Location != "" && cfg.Backup.Location != "default" {
		// Use custom backup location from config
		customPath := m.ExpandHomeDir(cfg.Backup.Location)
		// Create directory if it doesn't exist
		if err := os.MkdirAll(customPath, 0755); err == nil {
			m.backupsDir = customPath
			return customPath
		}
	}

	// Use default backups subdirectory
	backupsDir := filepath.Join(plonkDir, "backups")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		// If we can't create it, fall back to legacy location
		if cfg != nil {
			fallback := m.getLegacyBackupDirectory(cfg)
			m.backupsDir = fallback
			return fallback
		}
		// Last resort fallback
		m.backupsDir = filepath.Join(plonkDir, "backups")
		return m.backupsDir
	}

	m.backupsDir = backupsDir
	return backupsDir
}

// ExpandHomeDir expands ~ at the beginning of a path to the user's home directory
func (m *Manager) ExpandHomeDir(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path // fallback to original path
	}

	if len(path) == 1 || path[1] == '/' {
		return filepath.Join(homeDir, path[1:])
	}

	// Handle ~user syntax (though we don't use it in plonk)
	return path
}

// getLegacyBackupDirectory returns the configured backup directory with default fallback (legacy logic)
func (m *Manager) getLegacyBackupDirectory(cfg *config.Config) string {
	if cfg.Backup.Location != "" {
		return m.ExpandHomeDir(cfg.Backup.Location)
	}

	// Default to ~/.config/plonk/backups
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".plonk/backups" // fallback
	}

	return filepath.Join(homeDir, ".config", "plonk", "backups")
}

// EnsureStructure ensures the correct directory structure exists, performing migration if needed
func (m *Manager) EnsureStructure() error {
	// Check if migration is needed
	if !m.hasNewDirectoryStructure() {
		// Check if there are files in the root that need migration
		plonkDir := m.PlonkDir()
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
			return m.migrateDirectoryStructure()
		}
	}

	// Ensure directories exist (this will create them if they don't)
	m.RepoDir()
	m.BackupsDir()

	return nil
}

// hasNewDirectoryStructure checks if the new repo/ and backups/ structure exists
func (m *Manager) hasNewDirectoryStructure() bool {
	plonkDir := m.PlonkDir()
	repoDir := filepath.Join(plonkDir, "repo")
	backupsDir := filepath.Join(plonkDir, "backups")

	return utils.FileExists(repoDir) && utils.FileExists(backupsDir)
}

// migrateDirectoryStructure migrates from old flat structure to new subdirectory structure
func (m *Manager) migrateDirectoryStructure() error {
	plonkDir := m.PlonkDir()

	// Skip migration if new structure already exists
	if m.hasNewDirectoryStructure() {
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
		if m.isBackupFile(entry.Name()) {
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
func (m *Manager) isBackupFile(filename string) bool {
	return strings.Contains(filename, ".backup.")
}

// Reset clears cached paths (useful for testing or when PLONK_DIR changes)
func (m *Manager) Reset() {
	m.plonkDir = ""
	m.repoDir = ""
	m.backupsDir = ""
}
