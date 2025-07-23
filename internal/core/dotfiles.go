// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// Package core contains the core business logic for plonk.
// This package should never import from internal/commands or internal/cli.
package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/paths"
)

// AddSingleDotfile processes a single dotfile path and returns results for all files processed
func AddSingleDotfile(ctx context.Context, cfg *config.Config, homeDir, configDir, dotfilePath string, dryRun bool) []operations.OperationResult {
	// Resolve and validate dotfile path
	resolvedPath, err := ResolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		return []operations.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainDotfiles, "resolve", dotfilePath, "failed to resolve dotfile path"),
		}}
	}

	// Check if dotfile exists
	info, err := os.Stat(resolvedPath)
	if os.IsNotExist(err) {
		return []operations.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile does not exist: %s", resolvedPath)).WithSuggestionMessage("Check if path exists: ls -la " + resolvedPath),
		}}
	}
	if err != nil {
		return []operations.OperationResult{{
			Name:   dotfilePath,
			Status: "failed",
			Error:  errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "check", "failed to check dotfile"),
		}}
	}

	// Check if it's a directory and handle accordingly
	if info.IsDir() {
		return AddDirectoryFiles(ctx, cfg, resolvedPath, homeDir, configDir, dryRun)
	}

	// Handle single file
	result := AddSingleFile(ctx, cfg, resolvedPath, homeDir, configDir, dryRun)
	return []operations.OperationResult{result}
}

// AddSingleFile processes a single file and returns an OperationResult
func AddSingleFile(ctx context.Context, cfg *config.Config, filePath, homeDir, configDir string, dryRun bool) operations.OperationResult {
	// Generate source and destination paths
	source, destination := GeneratePaths(filePath, homeDir)

	result := operations.OperationResult{
		Name: filePath,
		Metadata: map[string]interface{}{
			"source":      source,
			"destination": destination,
		},
		FilesProcessed: 1,
	}

	// Check if already managed by checking if source file exists in config dir
	adapter := config.NewConfigAdapter(cfg)
	dotfileTargets := adapter.GetDotfileTargets()
	if _, exists := dotfileTargets[source]; exists {
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
	sourcePath := filepath.Join(configDir, source)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0750); err != nil {
		result.Status = "failed"
		result.Error = errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainDotfiles, "create-dirs", "failed to create parent directories")
		return result
	}

	// Copy file with attribute preservation
	if err := CopyFileWithAttributes(filePath, sourcePath); err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", source, "failed to copy dotfile")
		return result
	}

	return result
}

// AddDirectoryFiles processes all files in a directory and returns results
func AddDirectoryFiles(ctx context.Context, cfg *config.Config, dirPath, homeDir, configDir string, dryRun bool) []operations.OperationResult {
	var results []operations.OperationResult
	ignorePatterns := cfg.Resolve().GetIgnorePatterns()

	// Use PathResolver to expand directory
	resolver := paths.NewPathResolver(homeDir, configDir)
	validator := paths.NewPathValidator(homeDir, configDir, ignorePatterns)

	entries, err := resolver.ExpandDirectory(dirPath)
	if err != nil {
		return []operations.OperationResult{{
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
			results = append(results, operations.OperationResult{
				Name:   entry.FullPath,
				Status: "failed",
				Error:  errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "stat", "failed to get file info"),
			})
			continue
		}

		// Check if file should be skipped
		if validator.ShouldSkipPath(entry.RelativePath, info) {
			continue
		}

		// Process each file individually
		result := AddSingleFile(ctx, cfg, entry.FullPath, homeDir, configDir, dryRun)
		results = append(results, result)
	}

	return results
}

// RemoveSingleDotfile removes a single dotfile
func RemoveSingleDotfile(homeDir, configDir string, cfg *config.Config, dotfilePath string, dryRun bool) operations.OperationResult {
	result := operations.OperationResult{
		Name: dotfilePath,
	}

	// Resolve dotfile path
	resolvedPath, err := ResolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainDotfiles, "resolve", dotfilePath, "failed to resolve dotfile path")
		return result
	}

	// Get the source file path in config directory
	_, destination := GeneratePaths(resolvedPath, homeDir)
	source := config.TargetToSource(destination)
	sourcePath := filepath.Join(configDir, source)

	// Check if file is managed (has corresponding file in config directory)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		result.Status = "skipped"
		result.Error = errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile '%s' is not managed by plonk", dotfilePath))
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
			result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "remove", dotfilePath, "failed to remove deployed dotfile")
			return result
		}
	}

	// Remove the source file from config directory
	if err := os.Remove(sourcePath); err != nil {
		// If we can't remove the source file, the deployed file is already gone
		// so we report partial success
		result.Status = "removed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "remove-source", source, "deployed file removed but failed to remove source file from config")
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

// CopyFileWithAttributes copies a file while preserving permissions and timestamps
func CopyFileWithAttributes(src, dst string) error {
	// Get source file info for preserving attributes
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Read source file
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination with source permissions
	if err := os.WriteFile(dst, content, srcInfo.Mode()); err != nil {
		return err
	}

	// Preserve timestamps
	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

// ResolveDotfilePath resolves relative paths and validates the dotfile path
func ResolveDotfilePath(path, homeDir string) (string, error) {
	// Create a path resolver instance
	resolver := paths.NewPathResolver(homeDir, config.GetDefaultConfigDirectory())
	return resolver.ResolveDotfilePath(path)
}

// GeneratePaths generates source and destination paths for the dotfile
func GeneratePaths(resolvedPath, homeDir string) (string, string) {
	// Create a path resolver instance
	resolver := paths.NewPathResolver(homeDir, config.GetDefaultConfigDirectory())
	source, destination, err := resolver.GeneratePaths(resolvedPath)
	if err != nil {
		// Fallback to manual generation if there's an error
		relPath := filepath.Base(resolvedPath)
		destination = "~/" + relPath
		source = config.TargetToSource(destination)
	}
	return source, destination
}
