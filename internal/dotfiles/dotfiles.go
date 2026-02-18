// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/richhaase/plonk/internal/ignore"
)

// errSkipDir is returned by walkDir callbacks to skip a directory
var errSkipDir = errors.New("skip directory")

const templateExtension = ".tmpl"

var templateVarPattern = regexp.MustCompile(`\{\{([A-Za-z_][A-Za-z0-9_]*)\}\}`)

func isTemplate(name string) bool {
	return strings.HasSuffix(name, templateExtension)
}

func renderTemplate(content []byte, lookupEnv func(string) (string, bool)) ([]byte, error) {
	matches := templateVarPattern.FindAllSubmatch(content, -1)
	if len(matches) == 0 {
		return content, nil
	}

	var missing []string
	seen := make(map[string]bool)
	for _, match := range matches {
		varName := string(match[1])
		if seen[varName] {
			continue
		}
		seen[varName] = true
		if _, ok := lookupEnv(varName); !ok {
			missing = append(missing, varName)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}

	result := templateVarPattern.ReplaceAllFunc(content, func(match []byte) []byte {
		varName := string(templateVarPattern.FindSubmatch(match)[1])
		val, _ := lookupEnv(varName)
		return []byte(val)
	})

	return result, nil
}

// DotfileManager manages dotfiles in a single config directory
type DotfileManager struct {
	configDir string     // $PLONK_DIR
	homeDir   string     // $HOME
	fs        FileSystem // file operations
	matcher   *ignore.Matcher
	lookupEnv func(string) (string, bool)
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
		lookupEnv: os.LookupEnv,
	}
}

// List returns all dotfiles in the config directory
func (m *DotfileManager) List() ([]Dotfile, error) {
	var dotfiles []Dotfile

	err := m.walkDir(m.configDir, func(sourcePath string, isDir bool) error {
		relPath, err := filepath.Rel(m.configDir, sourcePath)
		if err != nil {
			return err
		}

		// Check ignore patterns for both files and directories
		if m.shouldIgnoreWithDir(relPath, isDir) {
			if isDir {
				return errSkipDir // Skip entire ignored directory
			}
			return nil // Skip ignored file
		}

		if isDir {
			return nil // Continue into non-ignored directory
		}

		dotfiles = append(dotfiles, Dotfile{
			Name:   relPath,
			Source: sourcePath,
			Target: m.toTarget(relPath),
		})
		return nil
	})

	// Treat missing config directory as empty dotfile set (first-run scenario)
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Check for template/plain file conflicts (same target path)
	targets := make(map[string]string) // target -> source name
	for _, d := range dotfiles {
		if existing, ok := targets[d.Target]; ok {
			return nil, fmt.Errorf("conflict: %s and %s both target %s", existing, d.Name, d.Target)
		}
		targets[d.Target] = d.Name
	}

	return dotfiles, nil
}

// Add copies a file from $HOME to $PLONK_DIR
func (m *DotfileManager) Add(targetPath string) error {
	// Resolve to absolute path
	absTarget := targetPath
	if !filepath.IsAbs(targetPath) {
		absTarget = filepath.Join(m.homeDir, targetPath)
	}

	// Security: validate path is under $HOME to prevent path traversal attacks
	if err := m.validatePathUnderHome(absTarget); err != nil {
		return err
	}

	// Require dot-prefixed path (dotfile manager only manages dotfiles)
	if err := m.requireDotPrefix(absTarget); err != nil {
		return err
	}

	// Security: reject paths under configDir to prevent self-referential adds
	if err := m.rejectPathUnderConfigDir(absTarget); err != nil {
		return err
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

	// Get source file info to preserve permissions
	info, err := m.fs.Stat(absTarget)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", absTarget, err)
	}
	mode := info.Mode().Perm()

	// Read source
	content, err := m.fs.ReadFile(absTarget)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", absTarget, err)
	}

	// Create parent directories
	if err := m.fs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to config dir, preserving original permissions
	if err := m.fs.WriteFile(destPath, content, mode); err != nil {
		return fmt.Errorf("failed to write %s: %w", destPath, err)
	}

	return nil
}

// addDirectory recursively adds all files in a directory
func (m *DotfileManager) addDirectory(absTarget string) error {
	return m.walkDir(absTarget, func(path string, isDir bool) error {
		// Get path relative to the target directory being added (preserves dots)
		relToTarget, err := filepath.Rel(absTarget, path)
		if err != nil {
			return err
		}

		// Check ignore patterns (with dots preserved)
		if m.shouldIgnoreWithDot(relToTarget, isDir) {
			if isDir {
				return errSkipDir // Skip entire directory
			}
			return nil // Skip file
		}

		if isDir {
			return nil // Continue into non-ignored directory
		}

		return m.addFile(path)
	})
}

// Remove deletes a file or directory from $PLONK_DIR
func (m *DotfileManager) Remove(name string) error {
	if err := m.ValidateRemove(name); err != nil {
		return err
	}

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

// ValidateRemove checks if a name can be removed without actually removing it
func (m *DotfileManager) ValidateRemove(name string) error {
	// Reject empty names (would resolve to configDir itself)
	if name == "" || name == "." {
		return fmt.Errorf("invalid dotfile name: %q", name)
	}

	sourcePath := filepath.Join(m.configDir, name)

	// Security: validate path is under configDir to prevent path traversal attacks
	if err := m.validatePathUnderConfigDir(sourcePath); err != nil {
		return err
	}

	// Reject internal config files
	if name == "plonk.lock" || name == "plonk.yaml" {
		return fmt.Errorf("cannot remove internal file: %s", name)
	}

	// Verify it exists
	if _, err := m.fs.Stat(sourcePath); err != nil {
		return fmt.Errorf("dotfile not found: %s", name)
	}

	return nil
}

// Deploy copies a file from $PLONK_DIR to $HOME (atomic write)
func (m *DotfileManager) Deploy(name string) error {
	sourcePath := filepath.Join(m.configDir, name)
	targetPath := m.toTarget(name)

	// Get source file info to preserve permissions
	info, err := m.fs.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}
	mode := info.Mode().Perm()

	// Read source
	content, err := m.fs.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	// Render template if needed
	if isTemplate(name) {
		content, err = renderTemplate(content, m.lookupEnv)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", name, err)
		}
	}

	// Create parent directories
	if err := m.fs.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	// Use restrictive permissions for temp file, final permissions set after rename
	tmpPath := targetPath + ".plonk.tmp"
	if err := m.fs.WriteFile(tmpPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := m.fs.Rename(tmpPath, targetPath); err != nil {
		// Clean up temp file on failure
		if cleanupErr := m.fs.Remove(tmpPath); cleanupErr != nil {
			log.Printf("Warning: failed to clean up temp file %s: %v", tmpPath, cleanupErr)
		}
		return fmt.Errorf("failed to rename: %w", err)
	}

	// Set final permissions after rename (rename preserves temp file permissions)
	if err := m.fs.Chmod(targetPath, mode); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// IsDrifted returns true if the target differs from source
func (m *DotfileManager) IsDrifted(d Dotfile) (bool, error) {
	sourceContent, err := m.fs.ReadFile(d.Source)
	if err != nil {
		return false, fmt.Errorf("failed to read source: %w", err)
	}

	// Render template if needed
	if isTemplate(d.Name) {
		sourceContent, err = renderTemplate(sourceContent, m.lookupEnv)
		if err != nil {
			return false, fmt.Errorf("failed to render template %s: %w", d.Name, err)
		}
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

	// Render template if needed
	if isTemplate(d.Name) {
		sourceContent, err = renderTemplate(sourceContent, m.lookupEnv)
		if err != nil {
			return "", fmt.Errorf("failed to render template %s: %w", d.Name, err)
		}
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
	fmt.Fprintf(&diff, "--- %s (source)\n", d.Source)
	fmt.Fprintf(&diff, "+++ %s (target)\n", d.Target)

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
				fmt.Fprintf(&diff, "-%s\n", srcLine)
			}
			if i < len(targetLines) {
				fmt.Fprintf(&diff, "+%s\n", tgtLine)
			}
		}
	}

	return diff.String(), nil
}

// toTarget converts a relative source path to an absolute target path
// e.g., "zshrc" -> "/home/user/.zshrc"
// e.g., "config/nvim/init.lua" -> "/home/user/.config/nvim/init.lua"
func (m *DotfileManager) toTarget(relPath string) string {
	// Strip .tmpl extension for template files
	if isTemplate(relPath) {
		relPath = strings.TrimSuffix(relPath, templateExtension)
	}

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
		// This error is unreachable in normal operation since absTarget comes from
		// walking m.homeDir. Log for debugging but continue with fallback.
		log.Printf("Warning: failed to compute relative path from %s to %s: %v", m.homeDir, absTarget, err)
		return absTarget
	}

	// Remove dot prefix from the first component
	parts := strings.SplitN(relPath, string(os.PathSeparator), 2)
	if len(parts[0]) > 0 && parts[0][0] == '.' {
		parts[0] = parts[0][1:]
	}
	return strings.Join(parts, string(os.PathSeparator))
}

// validatePathUnderHome ensures the path is under $HOME to prevent path traversal attacks
func (m *DotfileManager) validatePathUnderHome(absPath string) error {
	// Clean the path to resolve any .. components
	cleanPath := filepath.Clean(absPath)
	cleanHome := filepath.Clean(m.homeDir)

	// Check if the path is under home directory
	rel, err := filepath.Rel(cleanHome, cleanPath)
	if err != nil {
		return fmt.Errorf("path %s is not accessible from home directory: %w", absPath, err)
	}

	// If the relative path escapes via "..", the path is outside the home directory
	if relEscapes(rel) {
		return fmt.Errorf("path %s is outside home directory %s", absPath, m.homeDir)
	}

	return nil
}

// validatePathUnderConfigDir ensures the path is under $PLONK_DIR to prevent path traversal
func (m *DotfileManager) validatePathUnderConfigDir(absPath string) error {
	cleanPath := filepath.Clean(absPath)
	cleanConfig := filepath.Clean(m.configDir)

	rel, err := filepath.Rel(cleanConfig, cleanPath)
	if err != nil {
		return fmt.Errorf("path %s is not accessible from config directory: %w", absPath, err)
	}

	if relEscapes(rel) {
		return fmt.Errorf("path %s is outside config directory %s", absPath, m.configDir)
	}

	return nil
}

// rejectPathUnderConfigDir returns an error if the path is under $PLONK_DIR
func (m *DotfileManager) rejectPathUnderConfigDir(absPath string) error {
	cleanPath := filepath.Clean(absPath)
	cleanConfig := filepath.Clean(m.configDir)

	rel, err := filepath.Rel(cleanConfig, cleanPath)
	if err != nil {
		return nil // Different drives, not under configDir
	}

	if !relEscapes(rel) {
		return fmt.Errorf("cannot add files from config directory %s", m.configDir)
	}

	return nil
}

// requireDotPrefix ensures the first path component under $HOME starts with a dot.
// Plonk manages dotfiles; non-dot paths would be deployed to the wrong location
// because toTarget always adds a dot prefix.
func (m *DotfileManager) requireDotPrefix(absPath string) error {
	rel, err := filepath.Rel(m.homeDir, absPath)
	if err != nil {
		return fmt.Errorf("cannot compute relative path: %w", err)
	}

	// Reject home directory itself ("." means the path equals homeDir)
	if rel == "." {
		return fmt.Errorf("path %s is the home directory, not a dotfile", absPath)
	}

	parts := strings.SplitN(rel, string(os.PathSeparator), 2)
	if len(parts[0]) == 0 || parts[0][0] != '.' {
		return fmt.Errorf("path %s is not a dotfile (first component must start with '.')", absPath)
	}

	return nil
}

// relEscapes returns true if a filepath.Rel result represents path traversal
// (i.e., ".." or "../something"), but NOT files that merely start with ".."
// (e.g., "..env" is a valid filename, not traversal).
func relEscapes(rel string) bool {
	return rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// shouldIgnore returns true if the path should be ignored (assumes it's a file)
func (m *DotfileManager) shouldIgnore(relPath string) bool {
	return m.shouldIgnoreWithDir(relPath, false)
}

// shouldIgnoreWithDir returns true if the path should be ignored, with explicit isDir flag
func (m *DotfileManager) shouldIgnoreWithDir(relPath string, isDir bool) bool {
	// Ignore files/dirs that start with a dot in the config directory
	// These are internal files like .git, .gitignore, .beads, etc.
	// In $PLONK_DIR, only files WITHOUT a dot prefix are managed dotfiles
	// e.g., $PLONK_DIR/zshrc -> ~/.zshrc (dot added on deploy)
	if len(relPath) > 0 && relPath[0] == '.' {
		return true
	}

	// Ignore root-level plonk.yaml and plonk.lock (plonk's own config files)
	// Don't ignore nested files like config/plonk.yaml that users may want to manage
	if relPath == "plonk.yaml" || relPath == "plonk.lock" {
		return true
	}

	// Check custom ignore patterns
	if m.matcher == nil {
		return false
	}
	return m.matcher.ShouldIgnore(relPath, isDir)
}

// shouldIgnoreWithDot checks if a path should be ignored when adding from $HOME.
// Unlike shouldIgnore (for configDir paths), this preserves dots and only ignores
// specific VCS/system files, not all dotfiles.
func (m *DotfileManager) shouldIgnoreWithDot(relPath string, isDir bool) bool {
	// List of always-ignored file/directory names (VCS and system files)
	alwaysIgnore := map[string]bool{
		".git":           true,
		".gitignore":     true,
		".gitattributes": true,
		".gitmodules":    true,
		".svn":           true,
		".hg":            true,
		".DS_Store":      true,
		".Trash":         true,
		".cache":         true,
		".localized":     true,
	}

	// Check each path component against the always-ignore list
	parts := strings.Split(relPath, string(os.PathSeparator))
	for _, part := range parts {
		if alwaysIgnore[part] {
			return true
		}
	}

	// Check custom ignore patterns from matcher
	if m.matcher != nil && m.matcher.ShouldIgnore(relPath, isDir) {
		return true
	}

	return false
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
			if errors.Is(err, errSkipDir) {
				continue // Skip this directory, don't recurse
			}
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

// RenderSource reads a source file and renders it if it's a template.
// Returns the rendered content suitable for diffing against the deployed target.
func (m *DotfileManager) RenderSource(name string) ([]byte, error) {
	sourcePath := filepath.Join(m.configDir, name)
	content, err := m.fs.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read source: %w", err)
	}

	if isTemplate(name) {
		content, err = renderTemplate(content, m.lookupEnv)
		if err != nil {
			return nil, fmt.Errorf("failed to render template %s: %w", name, err)
		}
	}

	return content, nil
}

// ValidateAdd checks if a path can be added without actually adding it
func (m *DotfileManager) ValidateAdd(targetPath string) error {
	// Resolve to absolute path
	absTarget := targetPath
	if !filepath.IsAbs(targetPath) {
		absTarget = filepath.Join(m.homeDir, targetPath)
	}

	// Security: validate path is under $HOME
	if err := m.validatePathUnderHome(absTarget); err != nil {
		return err
	}

	// Require dot-prefixed path (dotfile manager only manages dotfiles)
	if err := m.requireDotPrefix(absTarget); err != nil {
		return err
	}

	// Security: reject paths under configDir
	if err := m.rejectPathUnderConfigDir(absTarget); err != nil {
		return err
	}

	// Verify target exists
	if _, err := m.fs.Stat(absTarget); err != nil {
		return fmt.Errorf("%s does not exist", absTarget)
	}

	return nil
}
