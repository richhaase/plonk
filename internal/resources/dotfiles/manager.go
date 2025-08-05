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
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
)

// Manager coordinates dotfile operations using focused components
type Manager struct {
	homeDir          string
	configDir        string
	pathResolver     PathResolver
	pathValidator    PathValidator
	directoryScanner DirectoryScanner
	configHandler    ConfigHandler
	fileComparator   FileComparator
	fileOperations   *FileOperations
}

// NewManager creates a new dotfile manager with all components
func NewManager(homeDir, configDir string) *Manager {
	pathResolver := NewPathResolver(homeDir, configDir)
	pathValidator := NewPathValidator(homeDir, configDir)
	directoryScanner := NewDirectoryScanner(homeDir, configDir, pathValidator, pathResolver)
	fileComparator := NewFileComparator()
	configHandler := NewConfigHandler(homeDir, configDir, pathResolver, directoryScanner, fileComparator)
	fileOperations := NewFileOperations(pathResolver)

	return &Manager{
		homeDir:          homeDir,
		configDir:        configDir,
		pathResolver:     pathResolver,
		pathValidator:    pathValidator,
		directoryScanner: directoryScanner,
		configHandler:    configHandler,
		fileComparator:   fileComparator,
		fileOperations:   fileOperations,
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

// Delegate path resolution methods
func (m *Manager) ResolveDotfilePath(path string) (string, error) {
	return m.pathResolver.ResolveDotfilePath(path)
}

func (m *Manager) GetSourcePath(source string) string {
	return m.pathResolver.GetSourcePath(source)
}

func (m *Manager) GetDestinationPath(destination string) (string, error) {
	return m.pathResolver.GetDestinationPath(destination)
}

func (m *Manager) GenerateDestinationPath(resolvedPath string) (string, error) {
	return m.pathResolver.GenerateDestinationPath(resolvedPath)
}

func (m *Manager) GenerateSourcePath(destination string) string {
	return m.pathResolver.GenerateSourcePath(destination)
}

func (m *Manager) GeneratePaths(resolvedPath string) (source, destination string, err error) {
	return m.pathResolver.GeneratePaths(resolvedPath)
}

// Delegate validation methods
func (m *Manager) ValidatePath(path string) error {
	return m.pathValidator.ValidatePath(path)
}

func (m *Manager) ValidatePaths(source, destination string) error {
	return m.pathValidator.ValidatePaths(source, destination)
}

func (m *Manager) ShouldSkipPath(relPath string, info os.FileInfo, ignorePatterns []string) bool {
	return m.pathValidator.ShouldSkipPath(relPath, info, ignorePatterns)
}

// Delegate directory operations
func (m *Manager) ExpandDirectoryPaths(dirPath string) ([]DirectoryEntry, error) {
	return m.directoryScanner.ExpandDirectoryPaths(dirPath)
}

func (m *Manager) ExpandConfigDirectory(ignorePatterns []string) (map[string]string, error) {
	return m.directoryScanner.ExpandConfigDirectory(ignorePatterns)
}

func (m *Manager) ListDotfiles(dir string) ([]string, error) {
	return m.directoryScanner.ListDotfiles(dir)
}

func (m *Manager) ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error) {
	return m.directoryScanner.ExpandDirectory(sourceDir, destDir)
}

// Delegate config operations
func (m *Manager) GetConfiguredDotfiles() ([]resources.Item, error) {
	return m.configHandler.GetConfiguredDotfiles()
}

func (m *Manager) GetActualDotfiles(ctx context.Context) ([]resources.Item, error) {
	return m.configHandler.GetActualDotfiles(ctx)
}

// Delegate file comparison
func (m *Manager) CompareFiles(path1, path2 string) (bool, error) {
	return m.fileComparator.CompareFiles(path1, path2)
}

// Utility methods that need to stay in manager
func (m *Manager) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m *Manager) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (m *Manager) DestinationToName(destination string) string {
	if strings.HasPrefix(destination, "~/") {
		return destination[2:]
	}
	return destination
}

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

// High-level operations that coordinate components
func (m *Manager) AddFiles(ctx context.Context, cfg *config.Config, dotfilePaths []string, opts AddOptions) ([]resources.OperationResult, error) {
	var allResults []resources.OperationResult

	for _, dotfilePath := range dotfilePaths {
		results := m.AddSingleDotfile(ctx, cfg, dotfilePath, opts.DryRun)
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func (m *Manager) AddSingleDotfile(ctx context.Context, cfg *config.Config, dotfilePath string, dryRun bool) []resources.OperationResult {
	// Resolve and validate dotfile path
	resolvedPath, err := m.pathResolver.ResolveDotfilePath(dotfilePath)
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

func (m *Manager) AddSingleFile(ctx context.Context, cfg *config.Config, filePath string, dryRun bool) resources.OperationResult {
	result := resources.OperationResult{
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
	configured, err := m.configHandler.GetConfiguredDotfiles()
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
	sourcePath := m.pathResolver.GetSourcePath(source)

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

func (m *Manager) AddDirectoryFiles(ctx context.Context, cfg *config.Config, dirPath string, dryRun bool) []resources.OperationResult {
	var results []resources.OperationResult
	ignorePatterns := cfg.IgnorePatterns

	entries, err := m.directoryScanner.ExpandDirectoryPaths(dirPath)
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
		if m.pathValidator.ShouldSkipPath(entry.RelativePath, info, ignorePatterns) {
			continue
		}

		// Process each file individually
		result := m.AddSingleFile(ctx, cfg, entry.FullPath, dryRun)
		results = append(results, result)
	}

	return results
}

func (m *Manager) RemoveFiles(cfg *config.Config, dotfilePaths []string, opts RemoveOptions) ([]resources.OperationResult, error) {
	var allResults []resources.OperationResult

	for _, dotfilePath := range dotfilePaths {
		result := m.RemoveSingleDotfile(cfg, dotfilePath, opts.DryRun)
		allResults = append(allResults, result)
	}

	return allResults, nil
}

func (m *Manager) RemoveSingleDotfile(cfg *config.Config, dotfilePath string, dryRun bool) resources.OperationResult {
	result := resources.OperationResult{
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
	source := TargetToSource(destination)
	sourcePath := m.pathResolver.GetSourcePath(source)

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

	// Only remove the source file from config directory
	// We never touch the deployed file in the user's environment
	if info, err := os.Stat(sourcePath); err == nil {
		if info.IsDir() {
			err = os.RemoveAll(sourcePath)
		} else {
			err = os.Remove(sourcePath)
		}
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Errorf("failed to remove source file %s from config: %w", source, err)
			return result
		}
	} else if os.IsNotExist(err) {
		result.Status = "skipped"
		result.Error = fmt.Errorf("dotfile '%s' is not managed by plonk", dotfilePath)
		return result
	} else {
		result.Status = "failed"
		result.Error = fmt.Errorf("failed to check source file %s: %w", source, err)
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

func (m *Manager) ProcessDotfileForApply(ctx context.Context, source, destination string, opts ApplyOptions) (output.DotfileOperation, error) {
	sourcePath := m.pathResolver.GetSourcePath(source)
	destinationPath, err := m.pathResolver.ResolveDotfilePath(destination)
	if err != nil {
		return output.DotfileOperation{}, fmt.Errorf("failed to resolve destination path: %w", err)
	}

	// Check if source exists
	if !m.FileExists(sourcePath) {
		return output.DotfileOperation{
			Source:      source,
			Destination: destination,
			Action:      "error",
			Status:      "failed",
			Error:       "source file does not exist",
		}, nil
	}

	// Check if destination already exists
	destExists := m.FileExists(destinationPath)

	action := output.DotfileOperation{
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

	// Configure copy options
	copyOptions := DefaultCopyOptions()
	copyOptions.CreateBackup = opts.Backup

	// Perform the copy
	err = m.fileOperations.CopyFile(ctx, source, destination, copyOptions)
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

// Type definitions that were moved from old manager

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

// Type aliases and backward compatibility

// ApplyResult is now defined in output package
type ApplyResult = output.DotfileOperation

// Backward compatibility functions for external usage

// GetConfiguredDotfiles is a convenience function that creates a Manager and calls GetConfiguredDotfiles
func GetConfiguredDotfiles(homeDir, configDir string) ([]resources.Item, error) {
	manager := NewManager(homeDir, configDir)
	return manager.GetConfiguredDotfiles()
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

// computeFileHash computes the SHA256 hash of a file (for backward compatibility)
func (m *Manager) computeFileHash(path string) (string, error) {
	return m.fileComparator.ComputeFileHash(path)
}

// createCompareFunc creates a comparison function for a dotfile (for backward compatibility)
func (m *Manager) createCompareFunc(source, destination string) func() (bool, error) {
	return func() (bool, error) {
		sourcePath := m.pathResolver.GetSourcePath(source)
		destPath, err := m.pathResolver.GetDestinationPath(destination)
		if err != nil {
			return false, err
		}
		// If destination doesn't exist, they're not the same
		if !m.FileExists(destPath) {
			return false, nil
		}
		return m.fileComparator.CompareFiles(sourcePath, destPath)
	}
}
