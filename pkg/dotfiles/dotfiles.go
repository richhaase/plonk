// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"plonk/pkg/config"
)

// DotfileStatus represents the state of a dotfile
type DotfileStatus int

const (
	DotfileManaged   DotfileStatus = iota // In plonk.yaml and exists in $HOME
	DotfileUntracked                      // Exists in $HOME but not in plonk.yaml
	DotfileMissing                        // In plonk.yaml but not in $HOME
	DotfileModified                       // Managed but differs from plonk version
)

// String returns the string representation of DotfileStatus
func (s DotfileStatus) String() string {
	switch s {
	case DotfileManaged:
		return "managed"
	case DotfileUntracked:
		return "untracked"
	case DotfileMissing:
		return "missing"
	case DotfileModified:
		return "modified"
	default:
		return "unknown"
	}
}

// DotfileInfo contains metadata about a dotfile
type DotfileInfo struct {
	Path         string        // Full path (e.g., "/Users/rdh/.zshrc")
	Name         string        // Basename (e.g., ".zshrc")
	Status       DotfileStatus // Current status
	Source       string        // Path in plonk repo (if managed)
	LastModified time.Time     // File modification time
	Size         int64         // File size in bytes
	IsDir        bool          // true if directory
}

// DotfilesManager defines operations for managing dotfiles
type DotfilesManager interface {
	ListManaged() ([]DotfileInfo, error)   // dotfiles in plonk.yaml
	ListUntracked() ([]DotfileInfo, error) // dotfiles in $HOME but not managed
	ListMissing() ([]DotfileInfo, error)   // dotfiles in plonk.yaml but not in $HOME
	ListModified() ([]DotfileInfo, error)  // managed dotfiles that differ from plonk version
	ListAll() ([]DotfileInfo, error)       // all dotfiles with their statuses
}

// DotfilesManagerInfo holds a dotfiles manager and its display name
type DotfilesManagerInfo struct {
	Name    string
	Manager DotfilesManager
}

// DefaultIgnorePatterns contains sensible default patterns to ignore
var DefaultIgnorePatterns = []string{
	".DS_Store",
	".git",
	".Trash",
	".cache",
	".tmp",
	"*.tmp",
	"*.log",
	".npm",
	".node_modules",
	".vscode",
	".idea",
}

// DotfilesManagerImpl implements DotfilesManager for dotfiles management
type DotfilesManagerImpl struct {
	homeDir        string
	plonkDir       string
	ignorePatterns []string
	maxFileSize    int64
}

// NewDotfilesManager creates a new dotfiles manager
func NewDotfilesManager(homeDir, plonkDir string) *DotfilesManagerImpl {
	return &DotfilesManagerImpl{
		homeDir:        homeDir,
		plonkDir:       plonkDir,
		ignorePatterns: DefaultIgnorePatterns,
		maxFileSize:    10 * 1024 * 1024, // 10MB default limit
	}
}

// SetIgnorePatterns sets custom ignore patterns
func (m *DotfilesManagerImpl) SetIgnorePatterns(patterns []string) {
	m.ignorePatterns = patterns
}

// SetMaxFileSize sets the maximum file size to consider
func (m *DotfilesManagerImpl) SetMaxFileSize(size int64) {
	m.maxFileSize = size
}

// shouldIgnore checks if a file should be ignored based on patterns
func (m *DotfilesManagerImpl) shouldIgnore(path string, info os.FileInfo) bool {
	name := filepath.Base(path)

	// Check file size limit
	if !info.IsDir() && info.Size() > m.maxFileSize {
		return true
	}

	// Check ignore patterns
	for _, pattern := range m.ignorePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		// Check if it's a directory exclusion
		if strings.HasSuffix(pattern, "/") && info.IsDir() {
			if name == strings.TrimSuffix(pattern, "/") {
				return true
			}
		}
	}

	return false
}

// discoverHomeDotfiles scans $HOME for dotfiles
func (m *DotfilesManagerImpl) discoverHomeDotfiles() ([]DotfileInfo, error) {
	var dotfiles []DotfileInfo

	entries, err := os.ReadDir(m.homeDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()

		// Only consider files/directories starting with '.'
		if !strings.HasPrefix(name, ".") {
			continue
		}

		path := filepath.Join(m.homeDir, name)

		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// Check if should ignore
		if m.shouldIgnore(path, info) {
			continue
		}

		dotfile := DotfileInfo{
			Path:         path,
			Name:         name,
			LastModified: info.ModTime(),
			Size:         info.Size(),
			IsDir:        info.IsDir(),
		}

		dotfiles = append(dotfiles, dotfile)
	}

	return dotfiles, nil
}

// getManagedDotfiles returns dotfiles listed in plonk.yaml
func (m *DotfilesManagerImpl) getManagedDotfiles() ([]string, error) {
	cfg, err := config.LoadConfig(m.plonkDir)
	if err != nil {
		// If no config exists, return empty list
		return []string{}, nil
	}

	return cfg.Dotfiles, nil
}

// ListManaged returns dotfiles that are managed by plonk
func (m *DotfilesManagerImpl) ListManaged() ([]DotfileInfo, error) {
	managedNames, err := m.getManagedDotfiles()
	if err != nil {
		return nil, err
	}

	var managed []DotfileInfo

	for _, name := range managedNames {
		path := filepath.Join(m.homeDir, name)

		// Check if file exists
		info, err := os.Stat(path)
		if err != nil {
			continue // Skip missing files (they'll show up in ListMissing)
		}

		dotfile := DotfileInfo{
			Path:         path,
			Name:         name,
			Status:       DotfileManaged,
			Source:       filepath.Join(m.plonkDir, "repo", name),
			LastModified: info.ModTime(),
			Size:         info.Size(),
			IsDir:        info.IsDir(),
		}

		managed = append(managed, dotfile)
	}

	return managed, nil
}

// ListUntracked returns dotfiles in $HOME that are not managed by plonk
func (m *DotfilesManagerImpl) ListUntracked() ([]DotfileInfo, error) {
	allDotfiles, err := m.discoverHomeDotfiles()
	if err != nil {
		return nil, err
	}

	managedNames, err := m.getManagedDotfiles()
	if err != nil {
		return nil, err
	}

	// Create a map of managed names for quick lookup
	managedMap := make(map[string]bool)
	for _, name := range managedNames {
		managedMap[name] = true
	}

	var untracked []DotfileInfo
	for _, dotfile := range allDotfiles {
		if !managedMap[dotfile.Name] {
			dotfile.Status = DotfileUntracked
			untracked = append(untracked, dotfile)
		}
	}

	return untracked, nil
}

// ListMissing returns dotfiles in plonk.yaml that don't exist in $HOME
func (m *DotfilesManagerImpl) ListMissing() ([]DotfileInfo, error) {
	managedNames, err := m.getManagedDotfiles()
	if err != nil {
		return nil, err
	}

	var missing []DotfileInfo

	for _, name := range managedNames {
		path := filepath.Join(m.homeDir, name)

		// Check if file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			dotfile := DotfileInfo{
				Path:   path,
				Name:   name,
				Status: DotfileMissing,
				Source: filepath.Join(m.plonkDir, "repo", name),
			}
			missing = append(missing, dotfile)
		}
	}

	return missing, nil
}

// ListModified returns managed dotfiles that differ from plonk version
func (m *DotfilesManagerImpl) ListModified() ([]DotfileInfo, error) {
	// For now, return empty list - modification detection requires file comparison
	// This can be implemented later with file hashing or timestamp comparison
	return []DotfileInfo{}, nil
}

// ListAll returns all dotfiles with their statuses
func (m *DotfilesManagerImpl) ListAll() ([]DotfileInfo, error) {
	var all []DotfileInfo

	managed, err := m.ListManaged()
	if err != nil {
		return nil, err
	}
	all = append(all, managed...)

	untracked, err := m.ListUntracked()
	if err != nil {
		return nil, err
	}
	all = append(all, untracked...)

	missing, err := m.ListMissing()
	if err != nil {
		return nil, err
	}
	all = append(all, missing...)

	modified, err := m.ListModified()
	if err != nil {
		return nil, err
	}
	all = append(all, modified...)

	return all, nil
}
