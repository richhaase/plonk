// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"

	"github.com/richhaase/plonk/internal/resources"
)

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

// TableOutput generates human-friendly table output for apply results
func (a ApplyOutput) TableOutput() string {
	// Table output is handled inline in the command
	return ""
}

// StructuredData returns the structured data for serialization
func (a ApplyOutput) StructuredData() any {
	return a
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

// DotfileAddOutput represents the output structure for dotfile add command
type DotfileAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileBatchAddOutput represents the output structure for batch dotfile add operations
type DotfileBatchAddOutput struct {
	TotalFiles int                `json:"total_files" yaml:"total_files"`
	AddedFiles []DotfileAddOutput `json:"added_files" yaml:"added_files"`
	Errors     []string           `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// TableOutput generates human-friendly table output for dotfile add
func (d DotfileAddOutput) TableOutput() string {
	output := "Dotfile Add\n===========\n\n"

	// Handle failed action
	if d.Action == "failed" {
		output += fmt.Sprintf("✗ %s - %s\n", d.Path, d.Error)
		return output
	}

	var actionText string
	var isDryRun bool
	switch d.Action {
	case "would-add":
		actionText = "Would add dotfile to plonk configuration"
		isDryRun = true
	case "would-update":
		actionText = "Would update existing dotfile in plonk configuration"
		isDryRun = true
	case "updated":
		actionText = "Updated existing dotfile in plonk configuration"
	case "added":
		actionText = "Added dotfile to plonk configuration"
	default:
		actionText = d.Action
	}

	if isDryRun {
		output += fmt.Sprintf("🔍 %s (dry-run)\n", actionText)
	} else {
		output += fmt.Sprintf("✅ %s\n", actionText)
	}
	output += fmt.Sprintf("   Source: %s\n", d.Source)
	output += fmt.Sprintf("   Destination: %s\n", d.Destination)
	output += fmt.Sprintf("   Original: %s\n", d.Path)

	if !isDryRun {
		if d.Action == "updated" {
			output += "\nThe system file has been copied to your plonk config directory, overwriting the previous version\n"
		} else {
			output += "\nThe dotfile has been copied to your plonk config directory\n"
		}
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
	var addedCount, updatedCount, wouldAddCount, wouldUpdateCount int
	for _, file := range d.AddedFiles {
		switch file.Action {
		case "updated":
			updatedCount++
		case "added":
			addedCount++
		case "would-update":
			wouldUpdateCount++
		case "would-add":
			wouldAddCount++
		}
	}

	isDryRun := wouldAddCount > 0 || wouldUpdateCount > 0

	if isDryRun {
		if wouldAddCount > 0 && wouldUpdateCount > 0 {
			output += fmt.Sprintf("🔍 Would process %d files (%d add, %d update) - dry-run\n\n", d.TotalFiles, wouldAddCount, wouldUpdateCount)
		} else if wouldUpdateCount > 0 {
			output += fmt.Sprintf("🔍 Would update %d files in plonk configuration - dry-run\n\n", d.TotalFiles)
		} else {
			output += fmt.Sprintf("🔍 Would add %d files to plonk configuration - dry-run\n\n", d.TotalFiles)
		}
	} else {
		if addedCount > 0 && updatedCount > 0 {
			output += fmt.Sprintf("✅ Processed %d files (%d added, %d updated)\n\n", d.TotalFiles, addedCount, updatedCount)
		} else if updatedCount > 0 {
			output += fmt.Sprintf("✅ Updated %d files in plonk configuration\n\n", d.TotalFiles)
		} else {
			output += fmt.Sprintf("✅ Added %d files to plonk configuration\n\n", d.TotalFiles)
		}
	}

	for _, file := range d.AddedFiles {
		var actionIndicator string
		switch file.Action {
		case "updated":
			actionIndicator = "↻"
		case "added":
			actionIndicator = "+"
		case "would-update":
			actionIndicator = "↻"
		case "would-add":
			actionIndicator = "+"
		}
		output += fmt.Sprintf("   %s %s → %s\n", actionIndicator, file.Destination, file.Source)
	}

	if len(d.Errors) > 0 {
		output += fmt.Sprintf("\n⚠️  Warnings:\n")
		for _, err := range d.Errors {
			output += fmt.Sprintf("   %s\n", err)
		}
	}

	if !isDryRun {
		output += "\nAll files have been copied to your plonk config directory\n"
	}
	return output
}

// StructuredData returns the structured data for serialization
func (d DotfileBatchAddOutput) StructuredData() any {
	return d
}

// ExtractErrorMessages extracts error messages from failed results
func ExtractErrorMessages(results []resources.OperationResult) []string {
	var errors []string
	for _, result := range results {
		if result.Status == "failed" && result.Error != nil {
			errors = append(errors, fmt.Sprintf("failed to add %s: %v", result.Name, result.Error))
		}
	}
	return errors
}

// MapStatusToAction converts operation status to legacy action string
func MapStatusToAction(status string) string {
	switch status {
	case "added", "updated", "would-add", "would-update":
		return status
	default:
		return "failed"
	}
}

// ConvertToDotfileAddOutput converts OperationResult to DotfileAddOutput for structured output
func ConvertToDotfileAddOutput(results []resources.OperationResult) []DotfileAddOutput {
	outputs := make([]DotfileAddOutput, 0, len(results))
	for _, result := range results {
		if result.Status == "failed" {
			continue // Skip failed results, they're handled in errors
		}

		outputs = append(outputs, DotfileAddOutput{
			Source:      result.Metadata["source"].(string),
			Destination: result.Metadata["destination"].(string),
			Action:      MapStatusToAction(result.Status),
			Path:        result.Name,
		})
	}
	return outputs
}
