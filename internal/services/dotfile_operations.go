// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/paths"
	"github.com/richhaase/plonk/internal/state"
)

// DotfileApplyOptions configures dotfile apply operations
type DotfileApplyOptions struct {
	ConfigDir string
	HomeDir   string
	Config    *config.Config
	DryRun    bool
	Backup    bool
}

// DotfileApplyResult represents the result of dotfile apply operations
type DotfileApplyResult struct {
	DryRun     bool            `json:"dry_run" yaml:"dry_run"`
	Backup     bool            `json:"backup" yaml:"backup"`
	TotalFiles int             `json:"total_files" yaml:"total_files"`
	Actions    []DotfileAction `json:"actions" yaml:"actions"`
	Summary    DotfileSummary  `json:"summary" yaml:"summary"`
}

// DotfileAction represents an action taken on a dotfile
type DotfileAction struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Status      string `json:"status" yaml:"status"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileSummary provides summary statistics
type DotfileSummary struct {
	Added     int `json:"added" yaml:"added"`
	Updated   int `json:"updated" yaml:"updated"`
	Unchanged int `json:"unchanged" yaml:"unchanged"`
	Failed    int `json:"failed" yaml:"failed"`
}

// AddDotfileOptions configures dotfile addition operations
type AddDotfileOptions struct {
	Config      *config.Config
	HomeDir     string
	ConfigDir   string
	DotfilePath string
	DryRun      bool
}

// ApplyDotfiles applies dotfile configuration and returns the result
func ApplyDotfiles(ctx context.Context, options DotfileApplyOptions) (DotfileApplyResult, error) {
	// Create dotfile provider
	provider := CreateDotfileProvider(options.HomeDir, options.ConfigDir, options.Config)

	// Get configured dotfiles
	configuredItems, err := provider.GetConfiguredItems()
	if err != nil {
		return DotfileApplyResult{}, errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainDotfiles, "apply",
			"failed to get configured dotfiles")
	}

	var actions []DotfileAction
	summary := DotfileSummary{}

	// Process each configured dotfile
	for _, item := range configuredItems {
		action, err := ProcessDotfileForApply(ctx, ProcessDotfileOptions{
			ConfigDir:   options.ConfigDir,
			HomeDir:     options.HomeDir,
			Source:      item.Name,
			Destination: item.Metadata["destination"].(string),
			DryRun:      options.DryRun,
			Backup:      options.Backup,
		})

		if err != nil {
			action = DotfileAction{
				Source:      item.Name,
				Destination: item.Metadata["destination"].(string),
				Action:      "error",
				Status:      "failed",
				Error:       err.Error(),
			}
			summary.Failed++
		} else {
			switch action.Status {
			case "added":
				summary.Added++
			case "updated":
				summary.Updated++
			case "unchanged":
				summary.Unchanged++
			case "failed":
				summary.Failed++
			}
		}

		actions = append(actions, action)
	}

	return DotfileApplyResult{
		DryRun:     options.DryRun,
		Backup:     options.Backup,
		TotalFiles: len(configuredItems),
		Actions:    actions,
		Summary:    summary,
	}, nil
}

// ProcessDotfileOptions configures individual dotfile processing
type ProcessDotfileOptions struct {
	ConfigDir   string
	HomeDir     string
	Source      string
	Destination string
	DryRun      bool
	Backup      bool
}

// ProcessDotfileForApply processes a single dotfile for apply operations
func ProcessDotfileForApply(ctx context.Context, options ProcessDotfileOptions) (DotfileAction, error) {
	// Resolve paths
	resolver := paths.NewPathResolver(options.HomeDir, options.ConfigDir)

	sourcePath := filepath.Join(options.ConfigDir, options.Source)
	destinationPath, err := resolver.ResolveDotfilePath(options.Destination)
	if err != nil {
		return DotfileAction{}, errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "apply",
			"failed to resolve destination path")
	}

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return DotfileAction{
			Source:      options.Source,
			Destination: options.Destination,
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

	action := DotfileAction{
		Source:      options.Source,
		Destination: options.Destination,
	}

	if options.DryRun {
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
	manager := dotfiles.NewManager(options.HomeDir, options.ConfigDir)
	fileOps := dotfiles.NewFileOperations(manager)

	// Configure copy options
	copyOptions := dotfiles.DefaultCopyOptions()
	copyOptions.CreateBackup = options.Backup

	// Perform the copy
	err = fileOps.CopyFile(ctx, options.Source, options.Destination, copyOptions)
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

// AddSingleDotfile adds a single dotfile to configuration
func AddSingleDotfile(ctx context.Context, options AddDotfileOptions) []operations.OperationResult {
	resolver := paths.NewPathResolver(options.HomeDir, options.ConfigDir)
	resolvedPath, err := resolver.ResolveDotfilePath(options.DotfilePath)
	if err != nil {
		return []operations.OperationResult{{
			Name:   options.DotfilePath,
			Status: "failed",
			Error:  errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "add", "failed to resolve dotfile path"),
		}}
	}

	// Check if path exists
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return []operations.OperationResult{{
			Name:   options.DotfilePath,
			Status: "failed",
			Error:  errors.Wrap(err, errors.ErrFileNotFound, errors.DomainDotfiles, "add", "dotfile does not exist"),
		}}
	}

	if info.IsDir() {
		// Handle directory
		return AddDirectoryFiles(ctx, AddDirectoryOptions{
			Config:    options.Config,
			DirPath:   resolvedPath,
			HomeDir:   options.HomeDir,
			ConfigDir: options.ConfigDir,
			DryRun:    options.DryRun,
		})
	} else {
		// Handle single file
		result := AddSingleFile(ctx, AddSingleFileOptions{
			Config:    options.Config,
			FilePath:  resolvedPath,
			HomeDir:   options.HomeDir,
			ConfigDir: options.ConfigDir,
			DryRun:    options.DryRun,
		})
		return []operations.OperationResult{result}
	}
}

// AddSingleFileOptions configures single file addition
type AddSingleFileOptions struct {
	Config    *config.Config
	FilePath  string
	HomeDir   string
	ConfigDir string
	DryRun    bool
}

// AddSingleFile adds a single file to dotfile management
func AddSingleFile(ctx context.Context, options AddSingleFileOptions) operations.OperationResult {
	result := operations.OperationResult{
		Name: options.FilePath,
	}

	// Generate source and destination paths
	_, destPath := GeneratePaths(options.FilePath, options.HomeDir)

	if options.DryRun {
		result.Status = "would-add"
		return result
	}

	// Copy file with attributes
	err := CopyFileWithAttributes(options.FilePath, filepath.Join(options.ConfigDir, destPath))
	if err != nil {
		result.Status = "failed"
		result.Error = errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "add", "failed to copy file")
		return result
	}

	result.Status = "added"
	return result
}

// AddDirectoryOptions configures directory addition
type AddDirectoryOptions struct {
	Config    *config.Config
	DirPath   string
	HomeDir   string
	ConfigDir string
	DryRun    bool
}

// AddDirectoryFiles adds all files in a directory to dotfile management
func AddDirectoryFiles(ctx context.Context, options AddDirectoryOptions) []operations.OperationResult {
	var results []operations.OperationResult

	resolver := paths.NewPathResolver(options.HomeDir, options.ConfigDir)

	files, err := resolver.ExpandDirectory(options.DirPath)
	if err != nil {
		return []operations.OperationResult{{
			Name:   options.DirPath,
			Status: "failed",
			Error:  errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "add", "failed to expand directory"),
		}}
	}

	for _, file := range files {
		result := AddSingleFile(ctx, AddSingleFileOptions{
			Config:    options.Config,
			FilePath:  file.FullPath,
			HomeDir:   options.HomeDir,
			ConfigDir: options.ConfigDir,
			DryRun:    options.DryRun,
		})
		results = append(results, result)
	}

	return results
}

// dotfileConfigAdapter adapts config.Config to DotfileConfigLoader interface
type dotfileConfigAdapter struct {
	cfg *config.Config
}

func (d *dotfileConfigAdapter) GetDotfileTargets() map[string]string {
	// This would need to be implemented based on how dotfiles are configured
	// For now, return empty map as a placeholder
	return make(map[string]string)
}

func (d *dotfileConfigAdapter) GetIgnorePatterns() []string {
	if d.cfg != nil {
		return d.cfg.IgnorePatterns
	}
	return []string{}
}

func (d *dotfileConfigAdapter) GetExpandDirectories() []string {
	if d.cfg != nil && d.cfg.ExpandDirectories != nil {
		return *d.cfg.ExpandDirectories
	}
	return []string{}
}

// CreateDotfileProvider creates a dotfile provider
func CreateDotfileProvider(homeDir string, configDir string, cfg *config.Config) *state.DotfileProvider {
	return state.NewDotfileProvider(homeDir, configDir, &dotfileConfigAdapter{cfg: cfg})
}

// GeneratePaths generates source and destination paths for a dotfile
func GeneratePaths(resolvedPath, homeDir string) (string, string) {
	// Calculate relative path from home directory
	relPath, err := filepath.Rel(homeDir, resolvedPath)
	if err != nil {
		// If we can't make it relative, use the base name
		relPath = filepath.Base(resolvedPath)
	}

	// Remove leading dot from filename for storage
	destPath := relPath
	if strings.HasPrefix(filepath.Base(destPath), ".") {
		dir := filepath.Dir(destPath)
		base := filepath.Base(destPath)[1:] // Remove leading dot
		if dir == "." {
			destPath = base
		} else {
			destPath = filepath.Join(dir, base)
		}
	}

	// For config directory storage, we want to maintain the structure
	// but without the leading home path
	if strings.HasPrefix(relPath, ".config/") {
		// Store .config files in a config/ subdirectory
		destPath = "config/" + relPath[8:] // Remove ".config/" and add "config/"
	}

	return relPath, destPath
}

// CopyFileWithAttributes copies a file preserving attributes
func CopyFileWithAttributes(src, dst string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Wrap(err, errors.ErrDirectoryCreate, errors.DomainDotfiles, "copy", "failed to create destination directory")
	}

	// Get source file info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to stat source file")
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to read source file")
	}

	// Write to destination with same permissions
	err = os.WriteFile(dst, data, srcInfo.Mode())
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", "failed to write destination file")
	}

	return nil
}
