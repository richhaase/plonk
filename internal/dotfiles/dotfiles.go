// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/ignore"
)

// DotfileManager manages dotfiles in a single config directory
type DotfileManager struct {
	configDir string     // $PLONK_DIR
	homeDir   string     // $HOME
	fs        FileSystem // file operations
	matcher   *ignore.Matcher
}

// NewDotfileManager creates a manager using the real filesystem
func NewDotfileManager(configDir, homeDir string, ignorePatterns []string) *DotfileManager {
	return NewDotfileManagerWithFS(configDir, homeDir, ignorePatterns, OSFileSystem{})
}

// NewDotfileManagerWithFS creates a manager with a custom filesystem (for testing)
func NewDotfileManagerWithFS(configDir, homeDir string, ignorePatterns []string, fs FileSystem) *DotfileManager {
	return &DotfileManager{
		configDir: configDir,
		homeDir:   homeDir,
		fs:        fs,
		matcher:   ignore.NewMatcher(ignorePatterns),
	}
}

// List returns all dotfiles in the config directory
func (m *DotfileManager) List() ([]Dotfile, error) {
	var dotfiles []Dotfile

	err := m.walkDir(m.configDir, func(sourcePath string, isDir bool) error {
		if isDir {
			return nil // skip directories, only return files
		}

		relPath, err := filepath.Rel(m.configDir, sourcePath)
		if err != nil {
			return err
		}

		if m.shouldIgnore(relPath) {
			return nil
		}

		dotfiles = append(dotfiles, Dotfile{
			Name:   relPath,
			Source: sourcePath,
			Target: m.toTarget(relPath),
		})
		return nil
	})

	return dotfiles, err
}

// Add copies a file from $HOME to $PLONK_DIR
func (m *DotfileManager) Add(targetPath string) error {
	// Resolve to absolute path
	absTarget := targetPath
	if !filepath.IsAbs(targetPath) {
		absTarget = filepath.Join(m.homeDir, targetPath)
	}

	// Verify source exists
	info, err := m.fs.Stat(absTarget)
	if err != nil {
		return fmt.Errorf("%s does not exist", absTarget)
	}

	if info.IsDir() {
		return m.addDirectory(absTarget)
	}

	return m.addFile(absTarget)
}

// addFile adds a single file
func (m *DotfileManager) addFile(absTarget string) error {
	relPath := m.toSource(absTarget)
	destPath := filepath.Join(m.configDir, relPath)

	// Read source
	content, err := m.fs.ReadFile(absTarget)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", absTarget, err)
	}

	// Create parent directories
	if err := m.fs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to config dir
	if err := m.fs.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}

	return nil
}

// addDirectory recursively adds all files in a directory
func (m *DotfileManager) addDirectory(absTarget string) error {
	return m.walkDir(absTarget, func(path string, isDir bool) error {
		if isDir {
			return nil
		}
		return m.addFile(path)
	})
}

// Remove deletes a file or directory from $PLONK_DIR
func (m *DotfileManager) Remove(name string) error {
	sourcePath := filepath.Join(m.configDir, name)

	// Verify it exists
	info, err := m.fs.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("dotfile not found: %s", name)
	}

	// Use RemoveAll for directories to handle non-empty directories
	if info.IsDir() {
		if err := m.fs.RemoveAll(sourcePath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}
	} else {
		if err := m.fs.Remove(sourcePath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", name, err)
		}
	}

	return nil
}

// Deploy copies a file from $PLONK_DIR to $HOME (atomic write)
func (m *DotfileManager) Deploy(name string) error {
	sourcePath := filepath.Join(m.configDir, name)
	targetPath := m.toTarget(name)

	// Read source
	content, err := m.fs.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	// Create parent directories
	if err := m.fs.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpPath := targetPath + ".plonk.tmp"
	if err := m.fs.WriteFile(tmpPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := m.fs.Rename(tmpPath, targetPath); err != nil {
		// Clean up temp file on failure
		if cleanupErr := m.fs.Remove(tmpPath); cleanupErr != nil {
			log.Printf("Warning: failed to clean up temp file %s: %v", tmpPath, cleanupErr)
		}
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// IsDrifted returns true if the target differs from source
func (m *DotfileManager) IsDrifted(d Dotfile) (bool, error) {
	sourceContent, err := m.fs.ReadFile(d.Source)
	if err != nil {
		return false, fmt.Errorf("failed to read source: %w", err)
	}

	targetContent, err := m.fs.ReadFile(d.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // missing, not drifted
		}
		return false, fmt.Errorf("failed to read target: %w", err)
	}

	return !bytes.Equal(sourceContent, targetContent), nil
}

// Diff returns the difference between source and target
func (m *DotfileManager) Diff(d Dotfile) (string, error) {
	sourceContent, err := m.fs.ReadFile(d.Source)
	if err != nil {
		return "", fmt.Errorf("failed to read source: %w", err)
	}

	targetContent, err := m.fs.ReadFile(d.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("(target missing, source has %d bytes)", len(sourceContent)), nil
		}
		return "", fmt.Errorf("failed to read target: %w", err)
	}

	if bytes.Equal(sourceContent, targetContent) {
		return "", nil // no diff
	}

	// Simple line-by-line diff
	sourceLines := strings.Split(string(sourceContent), "\n")
	targetLines := strings.Split(string(targetContent), "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s (source)\n", d.Source))
	diff.WriteString(fmt.Sprintf("+++ %s (target)\n", d.Target))

	// Find differences
	maxLen := len(sourceLines)
	if len(targetLines) > maxLen {
		maxLen = len(targetLines)
	}

	for i := 0; i < maxLen; i++ {
		var srcLine, tgtLine string
		if i < len(sourceLines) {
			srcLine = sourceLines[i]
		}
		if i < len(targetLines) {
			tgtLine = targetLines[i]
		}

		if srcLine != tgtLine {
			if i < len(sourceLines) {
				diff.WriteString(fmt.Sprintf("-%s\n", srcLine))
			}
			if i < len(targetLines) {
				diff.WriteString(fmt.Sprintf("+%s\n", tgtLine))
			}
		}
	}

	return diff.String(), nil
}

// toTarget converts a relative source path to an absolute target path
// e.g., "zshrc" -> "/home/user/.zshrc"
// e.g., "config/nvim/init.lua" -> "/home/user/.config/nvim/init.lua"
func (m *DotfileManager) toTarget(relPath string) string {
	// Add dot prefix to the first path component
	parts := strings.SplitN(relPath, string(os.PathSeparator), 2)
	parts[0] = "." + parts[0]
	dotPath := strings.Join(parts, string(os.PathSeparator))
	return filepath.Join(m.homeDir, dotPath)
}

// toSource converts an absolute target path to a relative source path
// e.g., "/home/user/.zshrc" -> "zshrc"
// e.g., "/home/user/.config/nvim/init.lua" -> "config/nvim/init.lua"
func (m *DotfileManager) toSource(absTarget string) string {
	// Remove home prefix
	relPath, err := filepath.Rel(m.homeDir, absTarget)
	if err != nil {
		// This should only happen if paths are on different drives (Windows) or malformed
		log.Printf("Warning: failed to compute relative path from %s to %s: %v", m.homeDir, absTarget, err)
		return absTarget // fallback - may cause issues
	}

	// Remove dot prefix from the first component
	parts := strings.SplitN(relPath, string(os.PathSeparator), 2)
	if len(parts[0]) > 0 && parts[0][0] == '.' {
		parts[0] = parts[0][1:]
	}
	return strings.Join(parts, string(os.PathSeparator))
}

// shouldIgnore returns true if the path should be ignored
func (m *DotfileManager) shouldIgnore(relPath string) bool {
	// Ignore files/dirs that start with a dot in the config directory
	// These are internal files like .git, .gitignore, .beads, etc.
	// In $PLONK_DIR, only files WITHOUT a dot prefix are managed dotfiles
	// e.g., $PLONK_DIR/zshrc -> ~/.zshrc (dot added on deploy)
	if len(relPath) > 0 && relPath[0] == '.' {
		return true
	}

	// Ignore plonk.yaml and plonk.lock
	base := filepath.Base(relPath)
	if base == "plonk.yaml" || base == "plonk.lock" {
		return true
	}

	// Check custom ignore patterns
	if m.matcher == nil {
		return false
	}
	return m.matcher.ShouldIgnore(relPath, false)
}

// walkDir recursively walks a directory
func (m *DotfileManager) walkDir(root string, fn func(path string, isDir bool) error) error {
	entries, err := m.fs.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())

		if err := fn(path, entry.IsDir()); err != nil {
			return err
		}

		if entry.IsDir() {
			if err := m.walkDir(path, fn); err != nil {
				return err
			}
		}
	}

	return nil
}
