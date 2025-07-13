// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/spf13/cobra"
)

var dotAddCmd = &cobra.Command{
	Use:   "add <dotfile1> [dotfile2] ...",
	Short: "Add dotfile(s) to plonk configuration and import them",
	Long: `Import one or more existing dotfiles into your plonk configuration.

This command will:
- Copy the dotfiles from their current locations to your plonk dotfiles directory
- Add them to your plonk.yaml configuration
- Preserve the original files in case you need to revert

For directories, plonk will recursively add all files individually, respecting
ignore patterns configured in your plonk.yaml.

Path Resolution:
- Absolute paths: /home/user/.vimrc
- Tilde paths: ~/.vimrc
- Relative paths: First tries current directory, then home directory

Examples:
  plonk dot add ~/.zshrc                    # Add single file
  plonk dot add ~/.zshrc ~/.vimrc           # Add multiple files
  plonk dot add .zshrc .vimrc               # Finds files in home directory
  plonk dot add ~/.config/nvim/ ~/.tmux.conf # Add directory and file
  plonk dot add --dry-run ~/.zshrc ~/.vimrc # Preview what would be added`,
	RunE: runDotAdd,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	dotCmd.AddCommand(dotAddCmd)
	dotAddCmd.Flags().BoolP("dry-run", "n", false, "Show what would be added without making changes")

	// Add file path completion
	dotAddCmd.ValidArgsFunction = completeDotfilePaths
}

func runDotAdd(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Handle single or multiple dotfiles
	return addDotfiles(cmd, args, dryRun)
}

// addDotfiles handles adding one or more dotfiles
func addDotfiles(cmd *cobra.Command, dotfilePaths []string, dryRun bool) error {
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

	// Load config for ignore patterns
	cfg, err := loadOrCreateConfig(configDir)
	if err != nil {
		return err
	}

	// Create operation context with timeout
	opCtx, cancel := operations.CreateOperationContext(5 * time.Minute)
	defer cancel()

	// Process dotfiles sequentially
	results := make([]operations.OperationResult, 0)
	reporter := operations.NewProgressReporter("dotfile", format == OutputTable)

	for _, dotfilePath := range dotfilePaths {
		// Check for cancellation
		if err := operations.CheckCancellation(opCtx, errors.DomainDotfiles, "add-multiple"); err != nil {
			return err
		}

		// Process each dotfile (can result in multiple files for directories)
		dotfileResults := addSingleDotfile(opCtx, cfg, homeDir, configDir, dotfilePath, dryRun)

		// Show progress for each file processed
		for _, result := range dotfileResults {
			results = append(results, result)
			reporter.ShowItemProgress(result)
		}
	}

	// Handle output based on format
	if format == OutputTable {
		// Show summary for table output
		reporter.ShowBatchSummary(results)
	} else {
		// For structured output, create appropriate response
		if len(dotfilePaths) == 1 && len(results) == 1 {
			// Single dotfile/file - use existing DotfileAddOutput format for compatibility
			result := results[0]
			output := DotfileAddOutput{
				Source:      result.Metadata["source"].(string),
				Destination: result.Metadata["destination"].(string),
				Action:      mapStatusToAction(result.Status),
				Path:        result.Name,
			}
			return RenderOutput(output, format)
		} else {
			// Multiple dotfiles/files - use batch output
			batchOutput := DotfileBatchAddOutput{
				TotalFiles: len(results),
				AddedFiles: convertResultsToDotfileAdd(results),
				Errors:     extractErrorMessages(results),
			}
			return RenderOutput(batchOutput, format)
		}
	}

	// Determine exit code
	return operations.DetermineExitCode(results, errors.DomainDotfiles, "add-multiple")
}

// addSingleDotfile processes a single dotfile path and returns results for all files processed
func addSingleDotfile(ctx context.Context, cfg *config.Config, homeDir, configDir, dotfilePath string, dryRun bool) []operations.OperationResult {
	// Resolve and validate dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
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
		return addDirectoryFilesNew(ctx, cfg, resolvedPath, homeDir, configDir, dryRun)
	}

	// Handle single file
	result := addSingleFileNew(ctx, cfg, resolvedPath, homeDir, configDir, dryRun)
	return []operations.OperationResult{result}
}

// addSingleFileNew processes a single file and returns an OperationResult
func addSingleFileNew(ctx context.Context, cfg *config.Config, filePath, homeDir, configDir string, dryRun bool) operations.OperationResult {
	// Generate source and destination paths
	source, destination := generatePaths(filePath, homeDir)

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
	if err := copyFileWithAttributes(filePath, sourcePath); err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", source, "failed to copy dotfile")
		return result
	}

	return result
}

// addDirectoryFilesNew processes all files in a directory and returns results
func addDirectoryFilesNew(ctx context.Context, cfg *config.Config, dirPath, homeDir, configDir string, dryRun bool) []operations.OperationResult {
	var results []operations.OperationResult
	ignorePatterns := cfg.Resolve().GetIgnorePatterns()

	// Walk the directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			results = append(results, operations.OperationResult{
				Name:   path,
				Status: "failed",
				Error:  errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "walk", "failed to walk directory"),
			})
			return nil // Continue with other files
		}

		// Only process files, but DON'T skip directories (let Walk traverse them)
		if !info.IsDir() {
			// Check for cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Get relative path from the base directory being added
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				results = append(results, operations.OperationResult{
					Name:   path,
					Status: "failed",
					Error:  errors.Wrap(err, errors.ErrPathValidation, errors.DomainDotfiles, "relative-path", "failed to get relative path"),
				})
				return nil
			}

			// Check if file should be skipped
			if shouldSkipDotfile(relPath, info, ignorePatterns) {
				return nil
			}

			// Process each file individually
			result := addSingleFileNew(ctx, cfg, path, homeDir, configDir, dryRun)
			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		// If walk failed completely, return a single error result
		return []operations.OperationResult{{
			Name:   dirPath,
			Status: "failed",
			Error:  errors.Wrap(err, errors.ErrFileIO, errors.DomainDotfiles, "walk", "failed to walk directory"),
		}}
	}

	return results
}

// copyFileWithAttributes copies a file while preserving permissions and timestamps
func copyFileWithAttributes(src, dst string) error {
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

// mapStatusToAction converts operation status to legacy action string
func mapStatusToAction(status string) string {
	switch status {
	case "added", "would-add":
		return "added"
	case "updated", "would-update":
		return "updated"
	default:
		return "failed"
	}
}

// convertResultsToDotfileAdd converts OperationResult to DotfileAddOutput for structured output
func convertResultsToDotfileAdd(results []operations.OperationResult) []DotfileAddOutput {
	outputs := make([]DotfileAddOutput, 0, len(results))
	for _, result := range results {
		if result.Status == "failed" {
			continue // Skip failed results, they're handled in errors
		}

		outputs = append(outputs, DotfileAddOutput{
			Source:      result.Metadata["source"].(string),
			Destination: result.Metadata["destination"].(string),
			Action:      mapStatusToAction(result.Status),
			Path:        result.Name,
		})
	}
	return outputs
}

// extractErrorMessages extracts error messages from failed results
func extractErrorMessages(results []operations.OperationResult) []string {
	var errors []string
	for _, result := range results {
		if result.Status == "failed" && result.Error != nil {
			errors = append(errors, fmt.Sprintf("failed to add %s: %v", result.Name, result.Error))
		}
	}
	return errors
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

// completeDotfilePaths provides file path completion for dotfiles
func completeDotfilePaths(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	_, err := os.UserHomeDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	// Define common dotfile suggestions
	commonDotfiles := []string{
		"~/.zshrc", "~/.bashrc", "~/.bash_profile", "~/.profile",
		"~/.vimrc", "~/.vim/", "~/.nvim/",
		"~/.gitconfig", "~/.gitignore_global",
		"~/.tmux.conf", "~/.tmux/",
		"~/.ssh/config", "~/.ssh/",
		"~/.aws/config", "~/.aws/credentials",
		"~/.config/", "~/.config/nvim/", "~/.config/fish/", "~/.config/alacritty/",
		"~/.docker/config.json",
		"~/.zprofile", "~/.zshenv",
		"~/.inputrc", "~/.editorconfig",
	}

	// If no input yet, return all common suggestions
	if toComplete == "" {
		return commonDotfiles, cobra.ShellCompDirectiveNoSpace
	}

	// If starts with tilde, filter common dotfiles
	if strings.HasPrefix(toComplete, "~/") {
		var filtered []string
		for _, suggestion := range commonDotfiles {
			if strings.HasPrefix(suggestion, toComplete) {
				filtered = append(filtered, suggestion)
			}
		}

		if len(filtered) > 0 {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}

		// Fall back to file completion for ~/.config/ style paths
		return nil, cobra.ShellCompDirectiveDefault
	}

	// For relative paths, try to suggest based on common dotfile names
	if !strings.HasPrefix(toComplete, "/") {
		relativeSuggestions := []string{
			".zshrc", ".bashrc", ".bash_profile", ".profile",
			".vimrc", ".gitconfig", ".tmux.conf", ".inputrc",
			".editorconfig", ".zprofile", ".zshenv",
		}

		var filtered []string
		for _, suggestion := range relativeSuggestions {
			if strings.HasPrefix(suggestion, toComplete) {
				filtered = append(filtered, suggestion)
			}
		}

		if len(filtered) > 0 {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}
	}

	// Fall back to default file completion for absolute paths and other cases
	return nil, cobra.ShellCompDirectiveDefault
}
