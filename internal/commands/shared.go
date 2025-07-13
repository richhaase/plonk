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
	"github.com/richhaase/plonk/internal/dotfiles"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/operations"
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

// applyPackages applies package configuration and returns the result (from apply.go)
func applyPackages(configDir string, cfg *config.Config, dryRun bool, format OutputFormat) (ApplyOutput, error) {
	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider (multi-manager) - using lock file
	ctx := context.Background()

	// Create lock file adapter
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create package provider using registry
	registry := managers.NewManagerRegistry()
	packageProvider, err := registry.CreateMultiProvider(ctx, lockAdapter)
	if err != nil {
		return ApplyOutput{}, errors.Wrap(err, errors.ErrProviderNotFound, errors.DomainPackages, "apply",
			"failed to create package provider")
	}

	reconciler.RegisterProvider("package", packageProvider)

	// Reconcile package domain to find missing packages
	result, err := reconciler.ReconcileProvider(ctx, "package")
	if err != nil {
		return ApplyOutput{}, errors.Wrap(err, errors.ErrReconciliation, errors.DomainPackages, "reconcile", "failed to reconcile package state")
	}

	// Group missing packages by manager
	missingByManager := make(map[string][]state.Item)
	for _, item := range result.Missing {
		manager := item.Manager
		if manager == "" {
			manager = "unknown"
		}
		missingByManager[manager] = append(missingByManager[manager], item)
	}

	// Prepare output structure
	outputData := ApplyOutput{
		DryRun:       dryRun,
		TotalMissing: len(result.Missing),
		Managers:     make([]ManagerApplyResult, 0, len(missingByManager)),
	}

	// Handle case where no packages are missing
	if len(result.Missing) == 0 {
		if format == OutputTable {
			fmt.Println("📦 All packages up to date")
		}
		return outputData, nil
	}

	// Process each manager that has missing packages
	managerInstances := make(map[string]managers.PackageManager)
	for _, name := range registry.GetAllManagerNames() {
		manager, err := registry.GetManager(name)
		if err == nil {
			managerInstances[name] = manager
		}
	}

	for managerName, missingItems := range missingByManager {
		managerInstance, exists := managerInstances[managerName]
		if !exists {
			if format == OutputTable {
				fmt.Printf("📦 %s: Unknown manager, skipping\n", managerName)
			}
			continue
		}

		available, err := managerInstance.IsAvailable(ctx)
		if err != nil {
			// Log the error but continue without this manager
			// Note: Structured logging deferred to future enhancement
			continue
		}
		if !available {
			if format == OutputTable {
				fmt.Printf("📦 %s: Not available, skipping\n", managerName)
			}
			continue
		}

		// Convert manager name for display
		displayName := managerName
		switch managerName {
		case "homebrew":
			displayName = "Homebrew"
		case "npm":
			displayName = "NPM"
		case "cargo":
			displayName = "Cargo"
		}

		// Process missing packages for this manager
		managerResult := ManagerApplyResult{
			Name:         displayName,
			MissingCount: len(missingItems),
			Packages:     make([]PackageApplyResult, 0, len(missingItems)),
		}

		for _, item := range missingItems {
			packageResult := PackageApplyResult{
				Name:   item.Name,
				Status: "pending",
			}

			if dryRun {
				packageResult.Status = "would-install"
				if format == OutputTable {
					fmt.Printf("📦 Would install: %s (%s)\n", item.Name, displayName)
				}
				outputData.TotalWouldInstall++
			} else {
				// Actually install the package
				err := managerInstance.Install(ctx, item.Name)
				if err != nil {
					packageResult.Status = "failed"
					// Use structured error for better user messages
					plonkErr := errors.WrapWithItem(err, errors.ErrPackageInstall, errors.DomainPackages, "install", item.Name, "failed to install package")
					packageResult.Error = plonkErr.UserMessage()
					if format == OutputTable {
						fmt.Printf("📦 Failed to install %s: %v\n", item.Name, plonkErr.UserMessage())
					}
					outputData.TotalFailed++
				} else {
					packageResult.Status = "installed"
					if format == OutputTable {
						fmt.Printf("📦 Installed: %s (%s)\n", item.Name, displayName)
					}
					outputData.TotalInstalled++
				}
			}

			managerResult.Packages = append(managerResult.Packages, packageResult)
		}

		outputData.Managers = append(outputData.Managers, managerResult)
	}

	// Output summary for table format
	if format == OutputTable {
		if dryRun {
			fmt.Printf("📦 Package summary: %d packages would be installed\n", outputData.TotalWouldInstall)
		} else {
			fmt.Printf("📦 Package summary: %d installed, %d failed\n", outputData.TotalInstalled, outputData.TotalFailed)
		}
		fmt.Println()
	}

	return outputData, nil
}

// applyDotfiles applies dotfile configuration and returns the result (from apply.go)
func applyDotfiles(configDir, homeDir string, cfg *config.Config, dryRun, backup bool, format OutputFormat) (DotfileApplyOutput, error) {
	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register dotfile provider
	configAdapter := config.NewConfigAdapter(cfg)
	dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
	dotfileProvider := state.NewDotfileProvider(homeDir, configDir, dotfileConfigAdapter)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	// Reconcile dotfile domain to get expanded file list
	ctx := context.Background()
	result, err := reconciler.ReconcileProvider(ctx, "dotfile")
	if err != nil {
		return DotfileApplyOutput{}, errors.Wrap(err, errors.ErrReconciliation, errors.DomainDotfiles, "reconcile", "failed to reconcile dotfile state")
	}

	// Process each dotfile from the reconciled state
	var actions []DotfileAction
	deployedCount := 0
	skippedCount := 0

	// Process both missing and managed items that may need deployment
	allItems := append(result.Missing, result.Managed...)

	for _, item := range allItems {
		// Get source and destination from metadata
		source, _ := item.Metadata["source"].(string)
		destination, _ := item.Metadata["destination"].(string)

		if source == "" || destination == "" {
			continue
		}

		action, err := processDotfileForApply(ctx, configDir, homeDir, source, destination, dryRun, backup, format)
		if err != nil {
			return DotfileApplyOutput{}, errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "deploy", source, "failed to process dotfile")
		}

		actions = append(actions, action)

		if action.Status == "deployed" || action.Status == "would-deploy" {
			deployedCount++
		} else {
			skippedCount++
		}
	}

	// Output summary for table format
	if format == OutputTable {
		if dryRun {
			fmt.Printf("📄 Dotfile summary: %d dotfiles would be deployed, %d would be skipped\n", deployedCount, skippedCount)
		} else {
			fmt.Printf("📄 Dotfile summary: %d deployed, %d skipped\n", deployedCount, skippedCount)
		}
	}

	// Prepare output
	outputData := DotfileApplyOutput{
		DryRun:   dryRun,
		Deployed: deployedCount,
		Skipped:  skippedCount,
		Actions:  actions,
	}

	return outputData, nil
}

// processDotfileForApply handles the deployment of a single dotfile (from apply.go)
func processDotfileForApply(ctx context.Context, configDir, homeDir, source, destination string, dryRun, backup bool, format OutputFormat) (DotfileAction, error) {
	// Create dotfiles manager and file operations
	manager := dotfiles.NewManager(homeDir, configDir)
	fileOps := dotfiles.NewFileOperations(manager)

	action := DotfileAction{
		Source:      source,
		Destination: destination,
		Status:      "skipped",
		Reason:      "",
	}

	// Validate paths
	if err := manager.ValidatePaths(source, destination); err != nil {
		action.Status = "error"
		action.Reason = err.Error()
		return action, nil
	}

	// Check if source is a directory (should have been expanded)
	if manager.IsDirectory(manager.GetSourcePath(source)) {
		action.Status = "error"
		action.Reason = "unexpected directory (should have been expanded)"
		return action, nil
	}

	// Check if destination exists and is a directory
	destPath := manager.GetDestinationPath(destination)
	if manager.FileExists(destPath) && manager.IsDirectory(destPath) {
		action.Status = "error"
		action.Reason = "destination is a directory, expected file"
		return action, nil
	}

	// Check if file needs update
	needsUpdate, err := fileOps.FileNeedsUpdate(ctx, source, destination)
	if err != nil {
		return action, errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "check", source, "failed to check if file needs update")
	}

	if !needsUpdate {
		action.Status = "skipped"
		action.Reason = "files are identical"
		if format == OutputTable {
			fmt.Printf("📄 Skipped: %s (files are identical)\n", source)
		}
		return action, nil
	}

	// Need to deploy
	action.Status = "deployed"
	action.Reason = "copying from source"

	// Add backup indication if backup is requested and file exists
	if backup && manager.FileExists(destPath) {
		action.Reason = "copying from source (with backup)"
	}

	if dryRun {
		action.Status = "would-deploy"
		if format == OutputTable {
			fmt.Printf("📄 Would deploy: %s -> %s\n", source, destination)
		}
		return action, nil
	}

	// Configure copy options
	options := dotfiles.CopyOptions{
		CreateBackup:      backup,
		BackupSuffix:      ".backup",
		OverwriteExisting: true,
	}

	// Copy file using dotfiles operations
	if err := fileOps.CopyFile(ctx, source, destination, options); err != nil {
		return action, errors.WrapWithItem(err, errors.ErrFileIO, errors.DomainDotfiles, "copy", source, "failed to copy dotfile")
	}

	if format == OutputTable {
		fmt.Printf("📄 Deployed: %s -> %s\n", source, destination)
	}

	return action, nil
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
	return config.GetOrCreateConfig(configDir)
}

// addDotfiles handles adding one or more dotfiles (from dot_add.go)
func addDotfiles(cmd *cobra.Command, dotfilePaths []string, dryRun bool) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
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
	var results []operations.OperationResult
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
				AddedFiles: convertToDotfileAddOutput(results),
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

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		// Use empty config if not found
		cfg = &config.Config{}
	}

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
