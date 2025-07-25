// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
)

// Manager handles all dotfile operations including path resolution, validation, and file operations
type Manager struct {
	homeDir   string
	configDir string
}

// NewManager creates a new dotfile manager
func NewManager(homeDir, configDir string) *Manager {
	return &Manager{
		homeDir:   homeDir,
		configDir: configDir,
	}
}

// HomeDir returns the home directory path
func (m *Manager) HomeDir() string {
	return m.homeDir
}

// ConfigDir returns the config directory path
func (m *Manager) ConfigDir() string {
	return m.configDir
}

// DotfileInfo represents information about a dotfile
type DotfileInfo struct {
	Name        string
	Source      string // Path in config directory
	Destination string // Path in home directory
	IsDirectory bool
	ParentDir   string // For files expanded from directories
	Metadata    map[string]interface{}
}

// DirectoryEntry represents a file found during directory expansion
type DirectoryEntry struct {
	RelativePath string // Path relative to the expanded directory
	FullPath     string // Full absolute path to the file
	ParentDir    string // Original directory path that was expanded
}

// AddOptions configures dotfile addition operations
type AddOptions struct {
	DryRun bool
	Force  bool
}

// RemoveOptions configures dotfile removal operations
type RemoveOptions struct {
	DryRun bool
}

// ApplyOptions configures dotfile apply operations
type ApplyOptions struct {
	DryRun bool
	Backup bool
}

// ApplyResult represents an action taken on a dotfile
type ApplyResult struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Status      string `json:"status" yaml:"status"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ResolveDotfilePath resolves a dotfile path to an absolute path within the home directory
func (m *Manager) ResolveDotfilePath(path string) (string, error) {
	var resolvedPath string

	if strings.HasPrefix(path, "~/") {
		resolvedPath = filepath.Join(m.homeDir, path[2:])
	} else if filepath.IsAbs(path) {
		resolvedPath = path
	} else {
		// Relative path - try current directory first, then home directory
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(absPath); err == nil {
			resolvedPath = absPath
		} else {
			homeRelativePath := filepath.Join(m.homeDir, path)
			if _, err := os.Stat(homeRelativePath); err == nil {
				resolvedPath = homeRelativePath
			} else {
				resolvedPath = absPath
			}
		}
	}

	// Ensure it's within the home directory
	if !strings.HasPrefix(resolvedPath, m.homeDir) {
		return "", fmt.Errorf("dotfile must be within home directory: %s", resolvedPath)
	}

	return resolvedPath, nil
}

// GetSourcePath returns the full path to a source file in the config directory
func (m *Manager) GetSourcePath(source string) string {
	return filepath.Join(m.configDir, source)
}

// GetDestinationPath converts a destination path to an absolute path in the home directory
func (m *Manager) GetDestinationPath(destination string) (string, error) {
	if strings.HasPrefix(destination, "~/") {
		return filepath.Join(m.homeDir, destination[2:]), nil
	}
	if filepath.IsAbs(destination) {
		return destination, nil
	}
	// Relative destination, assume it's relative to home
	return filepath.Join(m.homeDir, destination), nil
}

// GenerateDestinationPath converts a resolved absolute path to a destination path (~/relative/path)
func (m *Manager) GenerateDestinationPath(resolvedPath string) (string, error) {
	relPath, err := filepath.Rel(m.homeDir, resolvedPath)
	if err != nil {
		return "", err
	}
	return "~/" + relPath, nil
}

// GenerateSourcePath converts a destination path to a source path using plonk's naming convention
func (m *Manager) GenerateSourcePath(destination string) string {
	return TargetToSource(destination)
}

// GeneratePaths generates both source and destination paths for a resolved dotfile path
func (m *Manager) GeneratePaths(resolvedPath string) (source, destination string, err error) {
	destination, err = m.GenerateDestinationPath(resolvedPath)
	if err != nil {
		return "", "", err
	}
	source = m.GenerateSourcePath(destination)
	return source, destination, nil
}

// ValidatePath validates that a path is safe and within allowed boundaries
func (m *Manager) ValidatePath(path string) error {
	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	// Clean and resolve the path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return err
	}

	// Ensure path is within home directory
	if !strings.HasPrefix(absPath, m.homeDir) {
		return fmt.Errorf("path is outside home directory: %s", absPath)
	}

	// Ensure we're not managing plonk's own config
	if strings.HasPrefix(absPath, m.configDir) {
		return fmt.Errorf("cannot manage plonk configuration directory")
	}

	return nil
}

// ShouldSkipPath determines if a path should be skipped based on ignore patterns
func (m *Manager) ShouldSkipPath(relPath string, info os.FileInfo, ignorePatterns []string) bool {
	// Always skip plonk config files
	if relPath == "plonk.yaml" || relPath == "plonk.lock" {
		return true
	}

	// Check against ignore patterns
	for _, pattern := range ignorePatterns {
		if strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if info.IsDir() && strings.Contains(relPath, dirPattern) {
				return true
			}
			if strings.Contains(relPath, dirPattern+"/") {
				return true
			}
		} else {
			matched, err := filepath.Match(pattern, filepath.Base(relPath))
			if err == nil && matched {
				return true
			}
			matched, err = filepath.Match(pattern, relPath)
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

// ExpandDirectoryPaths walks a directory and returns individual file paths
func (m *Manager) ExpandDirectoryPaths(dirPath string) ([]DirectoryEntry, error) {
	var entries []DirectoryEntry

	resolvedDirPath, err := m.ResolveDotfilePath(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	info, err := os.Stat(resolvedDirPath)
	if err != nil {
		return nil, fmt.Errorf("directory does not exist: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	err = filepath.Walk(resolvedDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(resolvedDirPath, path)
		if err != nil {
			return err
		}

		entries = append(entries, DirectoryEntry{
			RelativePath: relPath,
			FullPath:     path,
			ParentDir:    dirPath,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return entries, nil
}

// ExpandConfigDirectory walks the config directory and returns all files suitable for dotfile management
func (m *Manager) ExpandConfigDirectory(ignorePatterns []string) (map[string]string, error) {
	result := make(map[string]string)

	err := filepath.Walk(m.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		relPath, err := filepath.Rel(m.configDir, path)
		if err != nil {
			return nil
		}

		// Always skip plonk config file
		if relPath == "plonk.yaml" {
			return nil
		}

		// Skip files based on ignore patterns
		if m.ShouldSkipPath(relPath, info, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories themselves (we'll get the files inside)
		if info.IsDir() {
			return nil
		}

		// Add to results with proper mapping
		source := relPath
		target := SourceToTarget(source)
		result[source] = target

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk config directory: %w", err)
	}

	return result, nil
}

// FileExists checks if a file exists at the given path
func (m *Manager) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func (m *Manager) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// DestinationToName converts a destination path to a standardized name
func (m *Manager) DestinationToName(destination string) string {
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}

// CreateDotfileInfo creates a DotfileInfo from source and destination paths
func (m *Manager) CreateDotfileInfo(source, destination string) DotfileInfo {
	sourcePath := m.GetSourcePath(source)
	isDir := m.IsDirectory(sourcePath)

	return DotfileInfo{
		Name:        m.DestinationToName(destination),
		Source:      source,
		Destination: destination,
		IsDirectory: isDir,
		Metadata: map[string]interface{}{
			"source":      source,
			"destination": destination,
		},
	}
}

// ValidatePaths validates that source and destination paths are valid
func (m *Manager) ValidatePaths(source, destination string) error {
	sourcePath := m.GetSourcePath(source)
	if !m.FileExists(sourcePath) {
		return fmt.Errorf("source file %s does not exist at %s", source, sourcePath)
	}

	if !strings.HasPrefix(destination, "~/") && !filepath.IsAbs(destination) {
		return fmt.Errorf("destination %s must start with ~/ or be absolute", destination)
	}

	return nil
}

// SourceToTarget converts a source path to target path using plonk's convention
// Prepends ~/. to make all files/directories hidden
func SourceToTarget(source string) string {
	return "~/." + source
}

// TargetToSource converts a target path to source path using plonk's convention
// Removes the ~/. prefix
func TargetToSource(target string) string {
	if len(target) > 3 && target[:3] == "~/." {
		return target[3:]
	}
	return target
}

// ListDotfiles finds all dotfiles in the specified directory
func (m *Manager) ListDotfiles(dir string) ([]string, error) {
	var dotfiles []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != "." && entry.Name() != ".." {
			dotfiles = append(dotfiles, entry.Name())
		}
	}

	return dotfiles, nil
}

// ExpandDirectory walks a directory and returns individual file entries as DotfileInfo
func (m *Manager) ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error) {
	var items []DotfileInfo
	sourcePath := filepath.Join(m.configDir, sourceDir)

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		source := filepath.Join(sourceDir, relPath)
		destination := filepath.Join(destDir, relPath)
		name := m.DestinationToName(destination)

		items = append(items, DotfileInfo{
			Name:        name,
			Source:      source,
			Destination: destination,
			IsDirectory: false,
			ParentDir:   sourceDir,
			Metadata: map[string]interface{}{
				"source":      source,
				"destination": destination,
				"parent_dir":  sourceDir,
			},
		})

		return nil
	})

	return items, err
}

// AddFiles processes multiple dotfile paths and returns results for all files processed
func (m *Manager) AddFiles(ctx context.Context, cfg *config.Config, dotfilePaths []string, opts AddOptions) ([]resources.OperationResult, error) {
	var allResults []resources.OperationResult

	for _, dotfilePath := range dotfilePaths {
		results := m.AddSingleDotfile(ctx, cfg, dotfilePath, opts.DryRun)
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// AddSingleDotfile processes a single dotfile path and returns results for all files processed
func (m *Manager) AddSingleDotfile(ctx context.Context, cfg *config.Config, dotfilePath string, dryRun bool) []resources.OperationResult {
	// Resolve and validate dotfile path
	resolvedPath, err := m.ResolveDotfilePath(dotfilePath)
	if err != nil {
		return []resources.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  fmt.Errorf("failed to resolve dotfile path %s: %w", dotfilePath, err),
		}}
	}

	// Check if dotfile exists
	info, err := os.Stat(resolvedPath)
	if os.IsNotExist(err) {
		return []resources.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  fmt.Errorf("dotfile does not exist: %s", resolvedPath),
		}}
	}
	if err != nil {
		return []resources.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  fmt.Errorf("failed to check dotfile: %w", err),
		}}
	}

	// Check if it's a directory and handle accordingly
	if info.IsDir() {
		return m.AddDirectoryFiles(ctx, cfg, resolvedPath, dryRun)
	}

	// Handle single file
	result := m.AddSingleFile(ctx, cfg, resolvedPath, dryRun)
	return []resources.OperationResult{result}
}

// AddSingleFile processes a single file and returns an resources.OperationResult
func (m *Manager) AddSingleFile(ctx context.Context, cfg *config.Config, filePath string, dryRun bool) resources.OperationResult {
	result := resources.OperationResult{
		Name: filePath,
	}

	// Generate source and destination paths
	source, destination, err := m.GeneratePaths(filePath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to generate paths: %w", err)
		return result
	}

	result.Metadata = map[string]interface{}{
		"source":      source,
		"destination": destination,
	}
	result.FilesProcessed = 1

	// Check if already managed by checking if source file exists in config dir
	configured, err := GetConfiguredDotfiles(m.homeDir, m.configDir)
	if err == nil {
		for _, item := range configured {
			if meta, ok := item.Metadata["source"].(string); ok && meta == source {
				result.AlreadyManaged = true
				break
			}
		}
	}
	if result.AlreadyManaged {
		if dryRun {
			result.Status = "would-update"
		} else {
			result.Status = "updated"
		}
	} else {
		if dryRun {
			result.Status = "would-add"
		} else {
			result.Status = "added"
		}
	}

	// If dry run, just return the result
	if dryRun {
		return result
	}

	// Copy file to plonk config directory
	sourcePath := filepath.Join(m.configDir, source)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0750); err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to create parent directories: %w", err)
		return result
	}

	// Copy file with attribute preservation
	if err := CopyFileWithAttributes(filePath, sourcePath); err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to copy dotfile %s: %w", source, err)
		return result
	}

	return result
}

// AddDirectoryFiles processes all files in a directory and returns results
func (m *Manager) AddDirectoryFiles(ctx context.Context, cfg *config.Config, dirPath string, dryRun bool) []resources.OperationResult {
	var results []resources.OperationResult
	ignorePatterns := cfg.IgnorePatterns

	entries, err := m.ExpandDirectoryPaths(dirPath)
	if err != nil {
		return []resources.OperationResult{{
			Name:   dirPath,
			Status: "failed",
			Error:  err,
		}}
	}

	// Process each file found in the directory
	for _, entry := range entries {
		// Check for cancellation
		if ctx.Err() != nil {
			break
		}

		// Get file info for skip checking
		info, err := os.Stat(entry.FullPath)
		if err != nil {
			results = append(results, resources.OperationResult{
				Name:   entry.FullPath,
				Status: "failed",
				Error:  fmt.Errorf("failed to get file info: %w", err),
			})
			continue
		}

		// Check if file should be skipped
		if m.ShouldSkipPath(entry.RelativePath, info, ignorePatterns) {
			continue
		}

		// Process each file individually
		result := m.AddSingleFile(ctx, cfg, entry.FullPath, dryRun)
		results = append(results, result)
	}

	return results
}

// RemoveFiles processes multiple dotfile paths for removal
func (m *Manager) RemoveFiles(cfg *config.Config, dotfilePaths []string, opts RemoveOptions) ([]resources.OperationResult, error) {
	var allResults []resources.OperationResult

	for _, dotfilePath := range dotfilePaths {
		result := m.RemoveSingleDotfile(cfg, dotfilePath, opts.DryRun)
		allResults = append(allResults, result)
	}

	return allResults, nil
}

// RemoveSingleDotfile removes a single dotfile
func (m *Manager) RemoveSingleDotfile(cfg *config.Config, dotfilePath string, dryRun bool) resources.OperationResult {
	result := resources.OperationResult{
		Name: dotfilePath,
	}

	// Resolve dotfile path
	resolvedPath, err := m.ResolveDotfilePath(dotfilePath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to resolve dotfile path %s: %w", dotfilePath, err)
		return result
	}

	// Get the source file path in config directory
	_, destination, err := m.GeneratePaths(resolvedPath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to generate paths for %s: %w", dotfilePath, err)
		return result
	}
	source := TargetToSource(destination)
	sourcePath := filepath.Join(m.configDir, source)

	// Check if file is managed (has corresponding file in config directory)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		result.Status = "skipped"
		result.Error = fmt.Errorf("dotfile '%s' is not managed by plonk", dotfilePath)
		return result
	}

	if dryRun {
		result.Status = "would-remove"
		result.Metadata = map[string]interface{}{
			"source":      source,
			"destination": destination,
			"path":        resolvedPath,
		}
		return result
	}

	// Remove the deployed file first (if it exists)
	if _, err := os.Stat(resolvedPath); err == nil {
		err = os.Remove(resolvedPath)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Errorf("failed to remove deployed dotfile %s: %w", dotfilePath, err)
			return result
		}
	}

	// Remove the source file from config directory
	if err := os.Remove(sourcePath); err != nil {
		result.Status = "removed"
		result.Error = fmt.Errorf("deployed file removed but failed to remove source file %s from config: %w", source, err)
		result.Metadata = map[string]interface{}{
			"source":      source,
			"destination": destination,
			"path":        resolvedPath,
			"partial":     true,
		}
		return result
	}

	result.Status = "removed"
	result.Metadata = map[string]interface{}{
		"source":      source,
		"destination": destination,
		"path":        resolvedPath,
	}
	return result
}

// ProcessDotfileForApply processes a single dotfile for apply operations
func (m *Manager) ProcessDotfileForApply(ctx context.Context, source, destination string, opts ApplyOptions) (ApplyResult, error) {
	sourcePath := filepath.Join(m.configDir, source)
	destinationPath, err := m.ResolveDotfilePath(destination)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return ApplyResult{
			Source:      source,
			Destination: destination,
			Action:      "error",
			Status:      "failed",
			Error:       "source file does not exist",
		}, nil
	}

	// Check if destination already exists
	destExists := false
	if _, err := os.Stat(destinationPath); err == nil {
		destExists = true
	}

	action := ApplyResult{
		Source:      source,
		Destination: destination,
	}

	if opts.DryRun {
		if destExists {
			action.Action = "would-update"
			action.Status = "would-update"
		} else {
			action.Action = "would-add"
			action.Status = "would-add"
		}
		return action, nil
	}

	// Create file operations handler
	fileOps := NewFileOperations(m)

	// Configure copy options
	copyOptions := DefaultCopyOptions()
	copyOptions.CreateBackup = opts.Backup

	// Perform the copy
	err = fileOps.CopyFile(ctx, source, destination, copyOptions)
	if err != nil {
		action.Action = "error"
		action.Status = "failed"
		action.Error = err.Error()
		return action, nil
	}

	if destExists {
		action.Action = "update"
		action.Status = "updated"
	} else {
		action.Action = "add"
		action.Status = "added"
	}

	return action, nil
}

// GetConfiguredDotfiles returns dotfiles defined in configuration
func (m *Manager) GetConfiguredDotfiles() ([]resources.Item, error) {
	// Load config to get ignore patterns
	cfg := config.LoadWithDefaults(m.configDir)

	targets, err := m.ExpandConfigDirectory(cfg.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("expanding config directory: %w", err)
	}

	items := make([]resources.Item, 0)

	for source, destination := range targets {
		// Check if source is a directory
		sourcePath := m.GetSourcePath(source)
		info, err := os.Stat(sourcePath)
		if err != nil {
			// Source doesn't exist yet, treat as single file
			name := m.DestinationToName(destination)
			items = append(items, resources.Item{
				Name:   name,
				State:  resources.StateMissing,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
				},
			})
			continue
		}

		if info.IsDir() {
			// For directories, just use the directory itself as one item
			name := m.DestinationToName(destination)
			items = append(items, resources.Item{
				Name:   name,
				State:  resources.StateManaged,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
					"isDirectory": true,
				},
			})
		} else {
			// Single file
			name := m.DestinationToName(destination)
			items = append(items, resources.Item{
				Name:   name,
				State:  resources.StateManaged,
				Domain: "dotfile",
				Path:   destination,
				Metadata: map[string]interface{}{
					"source":      source,
					"destination": destination,
				},
			})
		}
	}

	return items, nil
}

// GetActualDotfiles returns dotfiles currently present in the home directory
func (m *Manager) GetActualDotfiles(ctx context.Context) ([]resources.Item, error) {
	// Load config to get ignore patterns and expand directories
	cfg := config.LoadWithDefaults(m.configDir)

	// Create filter with ignore patterns
	filter := NewFilter(cfg.IgnorePatterns, m.configDir, true)

	// Create scanner
	scanner := NewScanner(m.homeDir, filter)

	// Create expander
	expander := NewExpander(m.homeDir, cfg.ExpandDirectories, scanner)

	var items []resources.Item

	// Scan home directory for dotfiles
	scanResults, err := scanner.ScanDotfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanning dotfiles: %w", err)
	}

	// Process scan results
	for _, result := range scanResults {
		// Check if directory should be expanded
		shouldExpand := false
		for _, dir := range cfg.ExpandDirectories {
			if result.Name == dir {
				shouldExpand = true
				break
			}
		}
		if result.Info.IsDir() && shouldExpand {
			// For expanded directories, treat as single item
			expander.CheckDuplicate(result.Name)
			items = append(items, resources.Item{
				Name:   result.Name,
				State:  resources.StateUntracked,
				Domain: "dotfile",
				Path:   result.Path,
				Metadata: map[string]interface{}{
					"isDirectory": true,
					"expanded":    true,
				},
			})
		} else {
			// Single file or unexpanded directory
			expander.CheckDuplicate(result.Name)
			items = append(items, resources.Item{
				Name:     result.Name,
				State:    resources.StateUntracked,
				Domain:   "dotfile",
				Path:     result.Path,
				Metadata: result.Metadata,
			})
		}
	}

	return items, nil
}

// Backward compatibility functions for external usage

// GetConfiguredDotfiles is a convenience function that creates a Manager and calls GetConfiguredDotfiles
func GetConfiguredDotfiles(homeDir, configDir string) ([]resources.Item, error) {
	manager := NewManager(homeDir, configDir)
	return manager.GetConfiguredDotfiles()
}
