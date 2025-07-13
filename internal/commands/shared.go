// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richhaase/plonk/internal/business"
	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
	"github.com/richhaase/plonk/internal/paths"
	"github.com/richhaase/plonk/internal/state"
	"github.com/spf13/cobra"
)

// Shared types from the original apply.go

// ApplyOutput represents the output structure for package apply operations
type ApplyOutput struct {
	DryRun            bool                 `json:"dry_run" yaml:"dry_run"`
	TotalMissing      int                  `json:"total_missing" yaml:"total_missing"`
	TotalInstalled    int                  `json:"total_installed" yaml:"total_installed"`
	TotalFailed       int                  `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall int                  `json:"total_would_install" yaml:"total_would_install"`
	Managers          []ManagerApplyResult `json:"managers" yaml:"managers"`
}

// ManagerApplyResult represents the result for a specific manager
type ManagerApplyResult struct {
	Name         string               `json:"name" yaml:"name"`
	MissingCount int                  `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageApplyResult `json:"packages" yaml:"packages"`
}

// PackageApplyResult represents the result for a specific package
type PackageApplyResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileApplyOutput represents the output structure for dotfile apply operations
type DotfileApplyOutput struct {
	DryRun   bool            `json:"dry_run" yaml:"dry_run"`
	Deployed int             `json:"deployed" yaml:"deployed"`
	Skipped  int             `json:"skipped" yaml:"skipped"`
	Actions  []DotfileAction `json:"actions" yaml:"actions"`
}

// DotfileAction represents a single dotfile deployment action
type DotfileAction struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Status      string `json:"status" yaml:"status"`
	Reason      string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// TableOutput methods for shared types

// TableOutput generates human-friendly table output for apply results
func (a ApplyOutput) TableOutput() string {
	// Table output is handled inline in the command
	return ""
}

// StructuredData returns the structured data for serialization
func (a ApplyOutput) StructuredData() any {
	return a
}

// TableOutput generates human-friendly table output for dotfile apply
func (d DotfileApplyOutput) TableOutput() string {
	if d.DryRun {
		output := "Dotfile Apply (Dry Run)\n========================\n\n"
		if d.Deployed == 0 && d.Skipped == 0 {
			return output + "No dotfiles configured\n"
		}

		output += fmt.Sprintf("Would deploy: %d\n", d.Deployed)
		output += fmt.Sprintf("Would skip: %d\n", d.Skipped)

		if len(d.Actions) > 0 {
			output += "\nActions:\n"
			for _, action := range d.Actions {
				status := "❓"
				if action.Status == "would-deploy" {
					status = "🚀"
				} else if action.Status == "skipped" {
					status = "⏭️"
				} else if action.Status == "error" {
					status = "❌"
				}

				output += fmt.Sprintf("  %s %s -> %s", status, action.Source, action.Destination)
				if action.Reason != "" {
					output += fmt.Sprintf(" (%s)", action.Reason)
				}
				output += "\n"
			}
		}

		return output
	}

	output := "Dotfile Apply\n=============\n\n"
	if d.Deployed == 0 && d.Skipped == 0 {
		return output + "No dotfiles configured\n"
	}

	if d.Deployed > 0 {
		output += fmt.Sprintf("✅ Deployed: %d dotfiles\n", d.Deployed)
	}
	if d.Skipped > 0 {
		output += fmt.Sprintf("⏭️ Skipped: %d dotfiles\n", d.Skipped)
	}

	if len(d.Actions) > 0 {
		output += "\nActions:\n"
		for _, action := range d.Actions {
			status := "❓"
			if action.Status == "deployed" {
				status = "✅"
			} else if action.Status == "skipped" {
				status = "⏭️"
			} else if action.Status == "error" {
				status = "❌"
			}

			output += fmt.Sprintf("  %s %s -> %s", status, action.Source, action.Destination)
			if action.Reason != "" {
				output += fmt.Sprintf(" (%s)", action.Reason)
			}
			output += "\n"
		}
	}

	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileApplyOutput) StructuredData() any {
	return d
}

// DotfileListOutput represents the output structure for dotfile listing operations
type DotfileListOutput struct {
	Summary  DotfileListSummary `json:"summary" yaml:"summary"`
	Dotfiles []DotfileInfo      `json:"dotfiles" yaml:"dotfiles"`
}

// DotfileListSummary provides summary information for dotfile listing
type DotfileListSummary struct {
	Total     int  `json:"total" yaml:"total"`
	Managed   int  `json:"managed" yaml:"managed"`
	Missing   int  `json:"missing" yaml:"missing"`
	Untracked int  `json:"untracked" yaml:"untracked"`
	Verbose   bool `json:"verbose" yaml:"verbose"`
}

// DotfileInfo represents information about a single dotfile
type DotfileInfo struct {
	Name   string `json:"name" yaml:"name"`
	State  string `json:"state" yaml:"state"`
	Target string `json:"target" yaml:"target"`
	Source string `json:"source" yaml:"source"`
}

// TableOutput generates human-friendly table output for dotfile listing
func (d DotfileListOutput) TableOutput() string {
	output := "Dotfiles Summary\n================\n"

	if d.Summary.Total == 0 {
		return output + "No dotfiles found\n"
	}

	// Summary line
	output += fmt.Sprintf("Total: %d files", d.Summary.Total)
	if !d.Summary.Verbose {
		if d.Summary.Managed > 0 {
			output += fmt.Sprintf(" | ✓ Managed: %d", d.Summary.Managed)
		}
		if d.Summary.Missing > 0 {
			output += fmt.Sprintf(" | ⚠ Missing: %d", d.Summary.Missing)
		}
		if d.Summary.Untracked > 0 {
			output += fmt.Sprintf(" | ? Untracked: %d", d.Summary.Untracked)
		}
	}
	output += "\n\n"

	if len(d.Dotfiles) == 0 {
		return output + "No dotfiles to display\n"
	}

	// Table headers
	output += "  Status Target                                    Source\n"
	output += "  ------ ----------------------------------------- --------------------------------------\n"

	// Table rows
	for _, dotfile := range d.Dotfiles {
		var statusIcon string
		switch dotfile.State {
		case "managed":
			statusIcon = "✓"
		case "missing":
			statusIcon = "⚠"
		case "untracked":
			statusIcon = "?"
		default:
			statusIcon = "-"
		}

		target := dotfile.Target
		if target == "" {
			target = "-"
		}
		source := dotfile.Source
		if source == "" {
			source = "-"
		}

		output += fmt.Sprintf("  %-6s %-41s %s\n", statusIcon, target, source)
	}

	// Show untracked hint if not verbose
	if !d.Summary.Verbose && d.Summary.Untracked > 0 {
		output += fmt.Sprintf("\n%d untracked files (use --verbose to show details)\n", d.Summary.Untracked)
	}

	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileListOutput) StructuredData() any {
	return d
}

// Shared functions from the original commands

// applyPackages applies package configuration and returns the result (refactored to use business module)
func applyPackages(configDir string, cfg *config.Config, dryRun bool, format OutputFormat) (ApplyOutput, error) {
	ctx := context.Background()

	// Use business module for package operations
	options := business.PackageApplyOptions{
		ConfigDir: configDir,
		Config:    cfg,
		DryRun:    dryRun,
	}

	result, err := business.ApplyPackages(ctx, options)
	if err != nil {
		return ApplyOutput{}, err
	}

	// Convert business result to command output format
	outputData := ApplyOutput{
		DryRun:            result.DryRun,
		TotalMissing:      result.TotalMissing,
		TotalInstalled:    result.TotalInstalled,
		TotalFailed:       result.TotalFailed,
		TotalWouldInstall: result.TotalWouldInstall,
		Managers:          make([]ManagerApplyResult, len(result.Managers)),
	}

	// Convert manager results
	for i, mgr := range result.Managers {
		packages := make([]PackageApplyResult, len(mgr.Packages))
		for j, pkg := range mgr.Packages {
			packages[j] = PackageApplyResult{
				Name:   pkg.Name,
				Status: pkg.Status,
				Error:  pkg.Error,
			}
		}
		outputData.Managers[i] = ManagerApplyResult{
			Name:         mgr.Name,
			MissingCount: mgr.MissingCount,
			Packages:     packages,
		}
	}

	// Output summary for table format
	if format == OutputTable {
		if result.TotalMissing == 0 {
			fmt.Println("📦 All packages up to date")
		} else {
			if dryRun {
				fmt.Printf("📦 Package summary: %d packages would be installed\n", outputData.TotalWouldInstall)
			} else {
				fmt.Printf("📦 Package summary: %d installed, %d failed\n", outputData.TotalInstalled, outputData.TotalFailed)
			}
		}
		fmt.Println()
	}

	return outputData, nil
}

// applyDotfiles applies dotfile configuration and returns the result (refactored to use business module)
func applyDotfiles(configDir, homeDir string, cfg *config.Config, dryRun, backup bool, format OutputFormat) (DotfileApplyOutput, error) {
	ctx := context.Background()

	// Use business module for dotfile operations
	options := business.DotfileApplyOptions{
		ConfigDir: configDir,
		HomeDir:   homeDir,
		Config:    cfg,
		DryRun:    dryRun,
		Backup:    backup,
	}

	result, err := business.ApplyDotfiles(ctx, options)
	if err != nil {
		return DotfileApplyOutput{}, err
	}

	// Convert business result to command output format
	actions := make([]DotfileAction, len(result.Actions))
	for i, action := range result.Actions {
		actions[i] = DotfileAction{
			Source:      action.Source,
			Destination: action.Destination,
			Status:      action.Status,
			Reason:      "", // Business module uses Action field differently
		}
	}

	outputData := DotfileApplyOutput{
		DryRun:   result.DryRun,
		Deployed: result.Summary.Added + result.Summary.Updated,
		Skipped:  result.Summary.Unchanged,
		Actions:  actions,
	}

	// Output summary for table format
	if format == OutputTable {
		if result.TotalFiles == 0 {
			fmt.Println("📄 No dotfiles configured")
		} else {
			if dryRun {
				fmt.Printf("📄 Dotfile summary: %d dotfiles would be deployed, %d would be skipped\n", outputData.Deployed, outputData.Skipped)
			} else {
				fmt.Printf("📄 Dotfile summary: %d deployed, %d skipped\n", outputData.Deployed, outputData.Skipped)
			}
		}
	}

	return outputData, nil
}

// Shared functions from pkg_add.go and dot_add.go

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

// loadOrCreateConfig loads existing config or creates a new one
func loadOrCreateConfig(configDir string) (*config.Config, error) {
	manager := config.NewConfigManager(configDir)
	return manager.LoadOrCreate()
}

// createPackageProvider creates a multi-manager package provider using lock file
// TODO: Replace with RuntimeState in future refactoring
func createPackageProvider(ctx context.Context, configDir string) (*state.MultiManagerPackageProvider, error) {
	// Create lock file adapter
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create package provider using registry
	registry := managers.NewManagerRegistry()
	return registry.CreateMultiProvider(ctx, lockAdapter)
}

// createDotfileProvider creates a dotfile provider
// TODO: Replace with RuntimeState in future refactoring
func createDotfileProvider(homeDir string, configDir string, cfg *config.Config) *state.DotfileProvider {
	configAdapter := config.NewConfigAdapter(cfg)
	dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
	return state.NewDotfileProvider(homeDir, configDir, dotfileConfigAdapter)
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
		result := addSingleFileNew(ctx, cfg, entry.FullPath, homeDir, configDir, dryRun)
		results = append(results, result)
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

// convertToDotfileAddOutput converts OperationResult to DotfileAddOutput for structured output
func convertToDotfileAddOutput(results []operations.OperationResult) []DotfileAddOutput {
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
	// Create a path resolver instance
	resolver := paths.NewPathResolver(homeDir, config.GetDefaultConfigDirectory())
	return resolver.ResolveDotfilePath(path)
}

// generatePaths generates source and destination paths for the dotfile
func generatePaths(resolvedPath, homeDir string) (string, string) {
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

// Shared output types from dot_add.go

// DotfileAddOutput represents the output structure for dotfile add command
type DotfileAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
}

// DotfileBatchAddOutput represents the output structure for batch dotfile add operations
type DotfileBatchAddOutput struct {
	TotalFiles int                `json:"total_files" yaml:"total_files"`
	AddedFiles []DotfileAddOutput `json:"added_files" yaml:"added_files"`
	Errors     []string           `json:"errors,omitempty" yaml:"errors,omitempty"`
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

// Shared functions for pkg and dot list operations
// Note: runPkgList implementation deferred pending requirements clarification
func runPkgList(cmd *cobra.Command, args []string) error {
	// This would need to be implemented based on the original pkg_list.go logic
	// For now, return an error indicating it's not implemented
	return fmt.Errorf("runPkgList function needs to be implemented")
}

func runDotList(cmd *cobra.Command, args []string) error {
	// Note: Full dotfiles layer integration deferred to maintain current functionality
	// Current implementation delegates to the state reconciliation system

	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "dotfiles", "output-format", "invalid output format")
	}

	// Get directories and use the existing state reconciliation system
	configDir := config.GetDefaultConfigDirectory()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "dotfiles", "failed to get home directory")
	}

	// Load configuration using LoadOrDefault for consistent zero-config behavior
	cfg := config.LoadConfigWithDefaults(configDir)

	// Use the same state reconciliation system as status command
	reconciler := state.NewReconciler()
	dotfileProvider := createDotfileProvider(homeDir, configDir, cfg)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	ctx := context.Background()
	domainResult, err := reconciler.ReconcileProvider(ctx, "dotfile")
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile dotfiles")
	}

	// Parse filter flags
	showManaged, _ := cmd.Flags().GetBool("managed")
	showMissing, _ := cmd.Flags().GetBool("missing")
	showUntracked, _ := cmd.Flags().GetBool("untracked")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Filter based on flags
	var items []state.Item
	if showManaged {
		items = domainResult.Managed
	} else if showMissing {
		items = domainResult.Missing
	} else if showUntracked {
		items = domainResult.Untracked
	} else {
		// Default: show managed + missing, optionally untracked
		items = append(items, domainResult.Managed...)
		items = append(items, domainResult.Missing...)
		if verbose {
			items = append(items, domainResult.Untracked...)
		}
	}

	// Convert to output format using existing dotfiles types
	output := DotfileListOutput{
		Summary: DotfileListSummary{
			Total:     len(items),
			Managed:   len(domainResult.Managed),
			Missing:   len(domainResult.Missing),
			Untracked: len(domainResult.Untracked),
			Verbose:   verbose,
		},
		Dotfiles: convertToDotfileInfo(items),
	}

	return RenderOutput(output, format)
}

// convertToDotfileInfo converts state.Item to DotfileInfo for display
func convertToDotfileInfo(items []state.Item) []DotfileInfo {
	result := make([]DotfileInfo, len(items))
	for i, item := range items {
		// Map state.Item fields to DotfileInfo
		target := item.Path
		source := item.Name

		// Extract additional info from metadata if available
		if item.Metadata != nil {
			if t, ok := item.Metadata["target"].(string); ok && t != "" {
				target = t
			}
			if s, ok := item.Metadata["source"].(string); ok && s != "" {
				source = s
			}
		}

		result[i] = DotfileInfo{
			Name:   item.Name,
			State:  item.State.String(),
			Target: target,
			Source: source,
		}
	}
	return result
}

// removeSingleDotfile removes a single dotfile
func removeSingleDotfile(homeDir, configDir string, cfg *config.Config, dotfilePath string, dryRun bool) operations.OperationResult {
	result := operations.OperationResult{
		Name: dotfilePath,
	}

	// Resolve dotfile path
	resolvedPath, err := resolveDotfilePath(dotfilePath, homeDir)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainDotfiles, "resolve", dotfilePath, "failed to resolve dotfile path")
		return result
	}

	// Check if file is managed (has a symlink)
	if !isSymlink(resolvedPath) {
		result.Status = "skipped"
		result.Error = errors.NewError(errors.ErrFileNotFound, errors.DomainDotfiles, "check", fmt.Sprintf("dotfile '%s' is not a managed symlink", dotfilePath))
		return result
	}

	if dryRun {
		result.Status = "would-unlink"
		return result
	}

	// Remove the symlink
	err = os.Remove(resolvedPath)
	if err != nil {
		result.Status = "failed"
		result.Error = errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "unlink", dotfilePath, "failed to remove symlink")
		return result
	}

	result.Status = "unlinked"
	result.Metadata = map[string]interface{}{
		"source":      dotfilePath,
		"destination": resolvedPath,
	}
	return result
}

// isSymlink checks if a path is a symbolic link
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// SimpleFlags represents basic command flags without detection logic
type SimpleFlags struct {
	Manager string
	DryRun  bool
	Force   bool
	Verbose bool
	Output  string
}

// ParseSimpleFlags parses basic flags for package commands
func ParseSimpleFlags(cmd *cobra.Command) (*SimpleFlags, error) {
	flags := &SimpleFlags{}

	// Parse manager flags with precedence
	if brew, _ := cmd.Flags().GetBool("brew"); brew {
		flags.Manager = "homebrew"
	} else if npm, _ := cmd.Flags().GetBool("npm"); npm {
		flags.Manager = "npm"
	} else if cargo, _ := cmd.Flags().GetBool("cargo"); cargo {
		flags.Manager = "cargo"
	}

	// Parse common flags
	flags.DryRun, _ = cmd.Flags().GetBool("dry-run")
	flags.Force, _ = cmd.Flags().GetBool("force")
	flags.Verbose, _ = cmd.Flags().GetBool("verbose")
	flags.Output, _ = cmd.Flags().GetString("output")

	return flags, nil
}
