// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"errors"
	"fmt"
	"strings"
)

// ApplyResult represents the top-level result of any apply operation
type ApplyResult struct {
	DryRun        bool            `json:"dry_run" yaml:"dry_run"`
	Success       bool            `json:"success" yaml:"success"` // True if no errors occurred (includes clean no-op)
	Changed       bool            `json:"changed" yaml:"changed"` // True if any changes were made
	Scope         string          `json:"scope" yaml:"scope"`     // "packages", "dotfiles", "all"
	Packages      *PackageResults `json:"packages,omitempty" yaml:"packages,omitempty"`
	Dotfiles      *DotfileResults `json:"dotfiles,omitempty" yaml:"dotfiles,omitempty"`
	Error         string          `json:"error,omitempty" yaml:"error,omitempty"`
	PackageErrors []error         `json:"-" yaml:"-"`
	DotfileErrors []error         `json:"-" yaml:"-"`
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

// TableOutput generates compact results for apply without hiding partial failures.
func (r ApplyResult) TableOutput() string {
	var w strings.Builder
	if r.DryRun {
		w.WriteString("Dry run · no changes will be made\n\n")
	}
	installed, deployed, failed, planned := 0, 0, 0, 0
	if r.Packages != nil {
		installed = r.Packages.TotalInstalled
		failed += r.Packages.TotalFailed
		planned += r.Packages.TotalWouldInstall
		for _, manager := range r.Packages.Managers {
			for _, pkg := range manager.Packages {
				WriteAction(&w, pkg.Status, manager.Name+":"+pkg.Name, pkg.Error, false)
			}
		}
	}
	if r.Dotfiles != nil {
		deployed = r.Dotfiles.Summary.Added + r.Dotfiles.Summary.Updated
		failed += r.Dotfiles.Summary.Failed
		if r.DryRun {
			planned += deployed
			deployed = 0
		}
		for _, item := range r.Dotfiles.Actions {
			WriteAction(&w, item.Status, item.Destination, item.Error, false)
		}
	}
	// Domain-level failures may have no per-item action to render.
	for _, err := range r.PackageErrors {
		WriteAction(&w, "error", "packages", err.Error(), false)
	}
	for _, err := range r.DotfileErrors {
		WriteAction(&w, "error", "dotfiles", err.Error(), false)
	}
	if r.Error != "" {
		WriteAction(&w, "error", "apply", r.Error, false)
	}
	if installed+deployed+planned+failed == 0 && !r.HasErrors() && r.Error == "" {
		w.WriteString("Already up to date. No changes.\n")
	} else {
		if installed+deployed+planned+failed > 0 {
			if r.DryRun {
				fmt.Fprintf(&w, "\n%d planned, %d failed\n", planned, failed)
			} else {
				fmt.Fprintf(&w, "\n%d installed, %d deployed, %d failed\n", installed, deployed, failed)
			}
		}
		if failed > 0 || r.HasErrors() || r.Error != "" {
			w.WriteString(ColorState("failed") + "  Completed with errors.\n")
		}
	}
	if r.DryRun {
		w.WriteString("No changes made.\n")
	}
	return w.String()
}

// AddPackageError adds an error to the package errors list
func (r *ApplyResult) AddPackageError(err error) {
	if err != nil {
		r.PackageErrors = append(r.PackageErrors, err)
	}
}

// AddDotfileError adds an error to the dotfile errors list
func (r *ApplyResult) AddDotfileError(err error) {
	if err != nil {
		r.DotfileErrors = append(r.DotfileErrors, err)
	}
}

// GetCombinedError returns all errors as a single error using errors.Join
func (r *ApplyResult) GetCombinedError() error {
	var allErrors []error
	allErrors = append(allErrors, r.PackageErrors...)
	allErrors = append(allErrors, r.DotfileErrors...)
	return errors.Join(allErrors...)
}

// HasErrors returns true if there are any errors
func (r *ApplyResult) HasErrors() bool {
	return len(r.PackageErrors) > 0 || len(r.DotfileErrors) > 0
}
