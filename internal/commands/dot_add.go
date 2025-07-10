// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plonk/internal/config"
	"plonk/internal/errors"

	"github.com/spf13/cobra"
)

var dotAddCmd = &cobra.Command{
	Use:   "add <dotfile>",
	Short: "Add a dotfile to plonk configuration and import it",
	Long: `Import an existing dotfile into your plonk configuration.

This command will:
- Copy the dotfile from its current location to your plonk dotfiles directory
- Add it to your plonk.yaml configuration
- Preserve the original file in case you need to revert

For directories, plonk will recursively add all files individually, respecting
ignore patterns configured in your plonk.yaml.

Path Resolution:
- Absolute paths: /home/user/.vimrc
- Tilde paths: ~/.vimrc
- Relative paths: First tries current directory, then home directory

Examples:
  plonk dot add ~/.zshrc           # Add single file
  plonk dot add .zshrc             # Finds ~/.zshrc (if not in current dir)
  plonk dot add ~/.config/nvim/    # Add all files in directory recursively
  cd ~/.config/nvim && plonk dot add init.lua  # Finds ./init.lua`,
	RunE: runDotAdd,
	Args: cobra.ExactArgs(1),
}

func init() {
	dotCmd.AddCommand(dotAddCmd)
}

func runDotAdd(cmd *cobra.Command, args []string) error {
	dotfilePath := args[0]

	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "dot-add", "output-format", "invalid output format")
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "dot-add", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Resolve and validate dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainDotfiles, "resolve", dotfilePath, "failed to resolve dotfile path")
	}

	// Check if dotfile exists
	info, err := os.Stat(resolvedPath)
	if os.IsNotExist(err) {
		return errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile does not exist: %s", resolvedPath))
	}
	if err != nil {
		return errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "check", "failed to check dotfile")
	}

	// Check if it's a directory and handle accordingly
	if info.IsDir() {
		return addDirectoryFiles(resolvedPath, homeDir, configDir, format)
	}

	// Handle single file (existing logic)
	return addSingleFile(resolvedPath, homeDir, configDir, format)
}

// resolveDotfilePath resolves relative paths and validates the dotfile path
func resolveDotfilePath(path, homeDir string) (string, error) {
	var resolvedPath string

	// Handle different path types
	if strings.HasPrefix(path, "~/") {
		// Expand ~ to home directory
		resolvedPath = filepath.Join(homeDir, path[2:])
	} else if filepath.IsAbs(path) {
		// Already absolute path
		resolvedPath = path
	} else {
		// Relative path - try to resolve relative to current working directory first
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "resolve", "failed to resolve path")
		}

		if _, err := os.Stat(absPath); err == nil {
			// File exists relative to current working directory
			resolvedPath = absPath
		} else {
			// Fall back to home directory
			homeRelativePath := filepath.Join(homeDir, path)
			if _, err := os.Stat(homeRelativePath); err == nil {
				// File exists relative to home directory
				resolvedPath = homeRelativePath
			} else {
				// Neither location has the file, use the current working directory path
				// so the error message will be more intuitive
				resolvedPath = absPath
			}
		}
	}

	// Ensure it's within the home directory
	if !strings.HasPrefix(resolvedPath, homeDir) {
		return "", errors.NewError(errors.ErrInvalidInput, errors.DomainDotfiles, "validate", fmt.Sprintf("dotfile must be within home directory: %s", resolvedPath))
	}

	return resolvedPath, nil
}

// generatePaths generates source and destination paths for the dotfile
func generatePaths(resolvedPath, homeDir string) (string, string) {
	// Get relative path from home directory
	relPath, err := filepath.Rel(homeDir, resolvedPath)
	if err != nil {
		// Fallback to just the filename
		relPath = filepath.Base(resolvedPath)
	}

	// Generate destination (always relative to home with ~ prefix)
	destination := "~/" + relPath

	// Generate source path using our naming convention
	source := targetToSource(destination)

	return source, destination
}

// targetToSource converts a target path to source path using our convention
func targetToSource(target string) string {
	// Use the config package implementation
	return config.TargetToSource(target)
}

// copyFileContents copies a file from src to dst
func copyFileContents(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0600)
}

// DotfileAddOutput represents the output structure for dotfile add command
type DotfileAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
}

// TableOutput generates human-friendly table output for dotfile add
func (d DotfileAddOutput) TableOutput() string {
	var actionText string
	if d.Action == "updated" {
		actionText = "Updated existing dotfile in plonk configuration"
	} else {
		actionText = "Added dotfile to plonk configuration"
	}

	output := "Dotfile Add\n===========\n\n"
	output += fmt.Sprintf("✅ %s\n", actionText)
	output += fmt.Sprintf("   Source: %s\n", d.Source)
	output += fmt.Sprintf("   Destination: %s\n", d.Destination)
	output += fmt.Sprintf("   Original: %s\n", d.Path)

	if d.Action == "updated" {
		output += "\nThe system file has been copied to your plonk config directory, overwriting the previous version\n"
	} else {
		output += "\nThe dotfile has been copied to your plonk config directory\n"
	}
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileAddOutput) StructuredData() any {
	return d
}

// addSingleFile handles adding a single file to plonk management
func addSingleFile(filePath, homeDir, configDir string, format OutputFormat) error {
	// Load existing configuration
	cfg, err := loadOrCreateConfig(configDir)
	if err != nil {
		return err
	}

	// Generate source and destination paths
	source, destination := generatePaths(filePath, homeDir)

	// Check if already managed by checking if source file exists in config dir
	adapter := config.NewConfigAdapter(cfg)
	dotfileTargets := adapter.GetDotfileTargets()
	action := "added"
	if _, exists := dotfileTargets[source]; exists {
		action = "updated"
	}

	// Copy dotfile to plonk config directory
	sourcePath := filepath.Join(configDir, source)
	if err := copyFileContents(filePath, sourcePath); err != nil {
		return errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", source, "failed to copy dotfile")
	}

	// Configuration doesn't need to be updated since we use auto-discovery
	// The dotfile will be automatically detected once it's in the config directory

	// Prepare output
	outputData := DotfileAddOutput{
		Source:      source,
		Destination: destination,
		Action:      action,
		Path:        filePath,
	}

	return RenderOutput(outputData, format)
}

// addDirectoryFiles handles adding all files in a directory recursively
func addDirectoryFiles(dirPath, homeDir, configDir string, format OutputFormat) error {
	var addedFiles []DotfileAddOutput
	var errorList []string

	// Load config to get ignore patterns
	cfg, err := loadOrCreateConfig(configDir)
	if err != nil {
		return err
	}

	ignorePatterns := cfg.Resolve().GetIgnorePatterns()

	// Walk the directory tree
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process files, but DON'T skip directories (let Walk traverse them)
		if !info.IsDir() {
			// Get relative path from the base directory being added
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}

			// Use existing shouldSkipDotfile function
			if shouldSkipDotfile(relPath, info, ignorePatterns) {
				return nil
			}

			// Process each file individually
			result, err := addSingleFileInternal(path, homeDir, configDir, cfg)
			if err != nil {
				// Log error but continue with other files
				errorList = append(errorList, fmt.Sprintf("failed to add %s: %v", path, err))
				return nil
			}

			addedFiles = append(addedFiles, result)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Render output for all added files
	outputData := DotfileBatchAddOutput{
		TotalFiles: len(addedFiles),
		AddedFiles: addedFiles,
		Errors:     errorList,
	}

	return RenderOutput(outputData, format)
}

// addSingleFileInternal handles adding a single file with an existing config
func addSingleFileInternal(filePath, homeDir, configDir string, cfg *config.Config) (DotfileAddOutput, error) {
	// Generate source and destination paths
	source, destination := generatePaths(filePath, homeDir)

	// Check if already managed by checking if source file exists in config dir
	adapter := config.NewConfigAdapter(cfg)
	dotfileTargets := adapter.GetDotfileTargets()
	action := "added"
	if _, exists := dotfileTargets[source]; exists {
		action = "updated"
	}

	// Copy file to plonk config directory
	sourcePath := filepath.Join(configDir, source)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0750); err != nil {
		return DotfileAddOutput{}, fmt.Errorf("failed to create parent directories: %v", err)
	}

	if err := copyFileContents(filePath, sourcePath); err != nil {
		return DotfileAddOutput{}, fmt.Errorf("failed to copy dotfile: %v", err)
	}

	return DotfileAddOutput{
		Source:      source,
		Destination: destination,
		Action:      action,
		Path:        filePath,
	}, nil
}

// loadOrCreateConfig loads existing config or creates a new one
func loadOrCreateConfig(configDir string) (*config.Config, error) {
	return config.GetOrCreateConfig(configDir)
}

// shouldSkipDotfile uses the existing function from config package
func shouldSkipDotfile(relPath string, info os.FileInfo, ignorePatterns []string) bool {
	// Always skip plonk config file
	if relPath == "plonk.yaml" {
		return true
	}

	// Check against configured ignore patterns
	for _, pattern := range ignorePatterns {
		// Check exact match for file/directory name
		if pattern == info.Name() || pattern == relPath {
			return true
		}
		// Check glob pattern match
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}

	return false
}

// DotfileBatchAddOutput represents the output structure for batch dotfile add operations
type DotfileBatchAddOutput struct {
	TotalFiles int                `json:"total_files" yaml:"total_files"`
	AddedFiles []DotfileAddOutput `json:"added_files" yaml:"added_files"`
	Errors     []string           `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// TableOutput generates human-friendly table output for batch dotfile add
func (d DotfileBatchAddOutput) TableOutput() string {
	output := fmt.Sprintf("Dotfile Directory Add\n=====================\n\n")

	// Count added vs updated
	var addedCount, updatedCount int
	for _, file := range d.AddedFiles {
		if file.Action == "updated" {
			updatedCount++
		} else {
			addedCount++
		}
	}

	if addedCount > 0 && updatedCount > 0 {
		output += fmt.Sprintf("✅ Processed %d files (%d added, %d updated)\n\n", d.TotalFiles, addedCount, updatedCount)
	} else if updatedCount > 0 {
		output += fmt.Sprintf("✅ Updated %d files in plonk configuration\n\n", d.TotalFiles)
	} else {
		output += fmt.Sprintf("✅ Added %d files to plonk configuration\n\n", d.TotalFiles)
	}

	for _, file := range d.AddedFiles {
		actionIndicator := "+"
		if file.Action == "updated" {
			actionIndicator = "↻"
		}
		output += fmt.Sprintf("   %s %s → %s\n", actionIndicator, file.Destination, file.Source)
	}

	if len(d.Errors) > 0 {
		output += fmt.Sprintf("\n⚠️  Warnings:\n")
		for _, err := range d.Errors {
			output += fmt.Sprintf("   %s\n", err)
		}
	}

	output += "\nAll files have been copied to your plonk config directory\n"
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileBatchAddOutput) StructuredData() any {
	return d
}
