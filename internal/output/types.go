package output

import "fmt"

// ApplyResult represents the top-level result of any apply operation
type ApplyResult struct {
	DryRun        bool            `json:"dry_run" yaml:"dry_run"`
	Success       bool            `json:"success" yaml:"success"`
	Scope         string          `json:"scope" yaml:"scope"` // "packages", "dotfiles", "all"
	Packages      *PackageResults `json:"packages,omitempty" yaml:"packages,omitempty"`
	Dotfiles      *DotfileResults `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
	Error         string          `json:"error,omitempty" yaml:"error,omitempty"`
	PackageErrors []string        `json:"package_errors,omitempty" yaml:"package_errors,omitempty"`
	DotfileErrors []string        `json:"dotfile_errors,omitempty" yaml:"dotfile_errors,omitempty"`
}

// PackageResults represents package apply operation results
type PackageResults struct {
	DryRun            bool             `json:"dry_run" yaml:"dry_run"`
	TotalMissing      int              `json:"total_missing" yaml:"total_missing"`
	TotalInstalled    int              `json:"total_installed" yaml:"total_installed"`
	TotalFailed       int              `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall int              `json:"total_would_install" yaml:"total_would_install"`
	Managers          []ManagerResults `json:"managers" yaml:"managers"`
}

// ManagerResults represents results for a specific package manager
type ManagerResults struct {
	Name         string             `json:"name" yaml:"name"`
	MissingCount int                `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageOperation `json:"packages" yaml:"packages"`
}

// PackageOperation represents a single package operation result
type PackageOperation struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"` // "installed", "failed", "would_install", etc.
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileResults represents dotfile apply operation results
type DotfileResults struct {
	DryRun     bool               `json:"dry_run" yaml:"dry_run"`
	TotalFiles int                `json:"total_files" yaml:"total_files"`
	Actions    []DotfileOperation `json:"actions" yaml:"actions"`
	Summary    DotfileSummary     `json:"summary" yaml:"summary"`
}

// DotfileOperation represents a single dotfile operation result
type DotfileOperation struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"` // "added", "updated", "unchanged", "failed"
	Status      string `json:"status" yaml:"status"` // "success", "failed", "skipped"
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileSummary represents dotfile operation summary
type DotfileSummary struct {
	Added     int `json:"added" yaml:"added"`
	Updated   int `json:"updated" yaml:"updated"`
	Unchanged int `json:"unchanged" yaml:"unchanged"`
	Failed    int `json:"failed" yaml:"failed"`
}

// TableOutput generates human-friendly table output for apply
func (r ApplyResult) TableOutput() string {
	output := ""

	if r.DryRun {
		output += "Plonk Apply (Dry Run)\n"
		output += "=====================\n\n"
	} else {
		output += "Plonk Apply\n"
		output += "===========\n\n"
	}

	// Show detailed results if available

	// Package details
	if r.Packages != nil && len(r.Packages.Managers) > 0 {
		for _, mgr := range r.Packages.Managers {
			if len(mgr.Packages) > 0 {
				output += fmt.Sprintf("%s:\n", mgr.Name)
				for _, pkg := range mgr.Packages {
					switch pkg.Status {
					case "installed":
						output += fmt.Sprintf("  ✓ %s\n", pkg.Name)
					case "would-install":
						output += fmt.Sprintf("  → %s (would install)\n", pkg.Name)
					case "failed":
						output += fmt.Sprintf("  ✗ %s: %s\n", pkg.Name, pkg.Error)
					}
				}
				output += "\n"
			}
		}
	}

	// Dotfile details
	if r.Dotfiles != nil && len(r.Dotfiles.Actions) > 0 {
		output += "Dotfiles:\n"
		for _, action := range r.Dotfiles.Actions {
			switch action.Status {
			case "added":
				output += fmt.Sprintf("  ✓ %s\n", action.Destination)
			case "would-add":
				output += fmt.Sprintf("  → %s (would deploy)\n", action.Destination)
			case "failed":
				output += fmt.Sprintf("  ✗ %s: %s\n", action.Destination, action.Error)
			}
		}
		output += "\n"
	}

	// Summary section
	output += "Summary:\n"
	output += "--------\n"

	totalSucceeded := 0
	totalFailed := 0

	// Package summary
	if r.Packages != nil {
		if r.DryRun {
			output += fmt.Sprintf("Packages: %d would be installed\n", r.Packages.TotalWouldInstall)
		} else {
			if r.Packages.TotalInstalled > 0 || r.Packages.TotalFailed > 0 {
				output += fmt.Sprintf("Packages: %d installed, %d failed\n", r.Packages.TotalInstalled, r.Packages.TotalFailed)
				totalSucceeded += r.Packages.TotalInstalled
				totalFailed += r.Packages.TotalFailed
			} else if r.Packages.TotalMissing == 0 {
				output += "Packages: All up to date\n"
			}
		}
	}

	// Dotfile summary
	if r.Dotfiles != nil {
		if r.DryRun {
			output += fmt.Sprintf("Dotfiles: %d would be deployed\n", r.Dotfiles.Summary.Added)
		} else {
			if r.Dotfiles.Summary.Added > 0 || r.Dotfiles.Summary.Failed > 0 {
				output += fmt.Sprintf("Dotfiles: %d deployed, %d failed\n", r.Dotfiles.Summary.Added, r.Dotfiles.Summary.Failed)
				totalSucceeded += r.Dotfiles.Summary.Added
				totalFailed += r.Dotfiles.Summary.Failed
			} else if r.Dotfiles.TotalFiles == 0 {
				output += "Dotfiles: None configured\n"
			} else {
				output += "Dotfiles: All up to date\n"
			}
		}
	}

	// Overall result
	if !r.DryRun && (totalSucceeded > 0 || totalFailed > 0) {
		output += fmt.Sprintf("\nTotal: %d succeeded, %d failed\n", totalSucceeded, totalFailed)
		if totalFailed > 0 {
			output += "\nSome operations failed. Check the errors above.\n"
		}
	}

	if r.DryRun {
		output += "\nUse 'plonk apply' without --dry-run to apply these changes\n"
	}

	return output
}

// StructuredData returns the data structure for JSON/YAML serialization
func (r ApplyResult) StructuredData() any {
	return r
}
