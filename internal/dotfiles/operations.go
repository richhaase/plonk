// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package dotfiles provides core dotfile management operations including
// file discovery, path resolution, directory expansion, and file operations.
package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/paths"
	"github.com/richhaase/plonk/internal/state"
)

// Manager handles dotfile operations and path management
type Manager struct {
	homeDir      string
	configDir    string
	pathResolver *paths.PathResolver
}

// NewManager creates a new dotfile manager
func NewManager(homeDir, configDir string) *Manager {
	return &Manager{
		homeDir:      homeDir,
		configDir:    configDir,
		pathResolver: paths.NewPathResolver(homeDir, configDir),
	}
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

// ExpandDirectory walks a directory and returns individual file entries
func (m *Manager) ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error) {
	var items []DotfileInfo
	sourcePath := filepath.Join(m.configDir, sourceDir)

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from source directory
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		// Build source and destination paths
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

// DestinationToName converts a destination path to a standardized name
func (m *Manager) DestinationToName(destination string) string {
	// Remove ~/ prefix if present
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}

// ResolvePath resolves a dotfile path using PathResolver with full validation
func (m *Manager) ResolvePath(path string) (string, error) {
	return m.pathResolver.ResolveDotfilePath(path)
}

// GetSourcePath returns the full source path for a dotfile
func (m *Manager) GetSourcePath(source string) string {
	return filepath.Join(m.configDir, source)
}

// GetDestinationPath returns the full destination path for a dotfile
func (m *Manager) GetDestinationPath(destination string) (string, error) {
	// Delegate to the centralized PathResolver
	return m.pathResolver.GetDestinationPath(destination)
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
	// Check if source exists in config directory
	sourcePath := m.GetSourcePath(source)
	if !m.FileExists(sourcePath) {
		return fmt.Errorf("source file %s does not exist at %s", source, sourcePath)
	}

	// Validate destination path format
	if !strings.HasPrefix(destination, "~/") && !filepath.IsAbs(destination) {
		return fmt.Errorf("destination %s must start with ~/ or be absolute", destination)
	}

	return nil
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

// AddFiles processes multiple dotfile paths and returns results for all files processed
func (m *Manager) AddFiles(ctx context.Context, cfg *config.Config, dotfilePaths []string, opts AddOptions) ([]state.OperationResult, error) {
	var allResults []state.OperationResult

	for _, dotfilePath := range dotfilePaths {
		results := m.AddSingleDotfile(ctx, cfg, dotfilePath, opts.DryRun)
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// AddSingleDotfile processes a single dotfile path and returns results for all files processed
func (m *Manager) AddSingleDotfile(ctx context.Context, cfg *config.Config, dotfilePath string, dryRun bool) []state.OperationResult {
	// Resolve and validate dotfile path
	resolvedPath, err := m.pathResolver.ResolveDotfilePath(dotfilePath)
	if err != nil {
		return []state.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  fmt.Errorf("failed to resolve dotfile path %s: %w", dotfilePath, err),
		}}
	}

	// Check if dotfile exists
	info, err := os.Stat(resolvedPath)
	if os.IsNotExist(err) {
		return []state.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  fmt.Errorf("dotfile does not exist: %s", resolvedPath),
		}}
	}
	if err != nil {
		return []state.OperationResult{{
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
	return []state.OperationResult{result}
}

// AddSingleFile processes a single file and returns an state.OperationResult
func (m *Manager) AddSingleFile(ctx context.Context, cfg *config.Config, filePath string, dryRun bool) state.OperationResult {
	result := state.OperationResult{
		Name: filePath,
	}

	// Generate source and destination paths
	source, destination, err := m.pathResolver.GeneratePaths(filePath)
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
func (m *Manager) AddDirectoryFiles(ctx context.Context, cfg *config.Config, dirPath string, dryRun bool) []state.OperationResult {
	var results []state.OperationResult
	ignorePatterns := cfg.IgnorePatterns

	// Use PathResolver to expand directory
	validator := paths.NewPathValidator(m.homeDir, m.configDir, ignorePatterns)

	entries, err := m.pathResolver.ExpandDirectory(dirPath)
	if err != nil {
		return []state.OperationResult{{
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
			results = append(results, state.OperationResult{
				Name:   entry.FullPath,
				Status: "failed",
				Error:  fmt.Errorf("failed to get file info: %w", err),
			})
			continue
		}

		// Check if file should be skipped
		if validator.ShouldSkipPath(entry.RelativePath, info) {
			continue
		}

		// Process each file individually
		result := m.AddSingleFile(ctx, cfg, entry.FullPath, dryRun)
		results = append(results, result)
	}

	return results
}

// RemoveFiles processes multiple dotfile paths for removal
func (m *Manager) RemoveFiles(cfg *config.Config, dotfilePaths []string, opts RemoveOptions) ([]state.OperationResult, error) {
	var allResults []state.OperationResult

	for _, dotfilePath := range dotfilePaths {
		result := m.RemoveSingleDotfile(cfg, dotfilePath, opts.DryRun)
		allResults = append(allResults, result)
	}

	return allResults, nil
}

// RemoveSingleDotfile removes a single dotfile
func (m *Manager) RemoveSingleDotfile(cfg *config.Config, dotfilePath string, dryRun bool) state.OperationResult {
	result := state.OperationResult{
		Name: dotfilePath,
	}

	// Resolve dotfile path
	resolvedPath, err := m.pathResolver.ResolveDotfilePath(dotfilePath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to resolve dotfile path %s: %w", dotfilePath, err)
		return result
	}

	// Get the source file path in config directory
	_, destination, err := m.pathResolver.GeneratePaths(resolvedPath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to generate paths for %s: %w", dotfilePath, err)
		return result
	}
	source := config.TargetToSource(destination)
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
		// If we can't remove the source file, the deployed file is already gone
		// so we report partial success
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

// ApplyResult represents an action taken on a dotfile
type ApplyResult struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Status      string `json:"status" yaml:"status"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ProcessDotfileForApply processes a single dotfile for apply operations
func (m *Manager) ProcessDotfileForApply(ctx context.Context, source, destination string, opts ApplyOptions) (ApplyResult, error) {
	sourcePath := filepath.Join(m.configDir, source)
	destinationPath, err := m.pathResolver.ResolveDotfilePath(destination)
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
