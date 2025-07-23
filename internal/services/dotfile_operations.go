// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package services

import (
	"context"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/core"
	"github.com/richhaase/plonk/internal/errors"
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
		result, err := core.ProcessDotfileForApply(ctx, core.ProcessDotfileForApplyOptions{
			ConfigDir:   options.ConfigDir,
			HomeDir:     options.HomeDir,
			Source:      item.Name,
			Destination: item.Metadata["destination"].(string),
			DryRun:      options.DryRun,
			Backup:      options.Backup,
		})

		action := DotfileAction{
			Source:      result.Source,
			Destination: result.Destination,
			Action:      result.Action,
			Status:      result.Status,
			Error:       result.Error,
		}

		if err != nil {
			action.Action = "error"
			action.Status = "failed"
			action.Error = err.Error()
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
		return d.cfg.GetIgnorePatterns()
	}
	return []string{}
}

func (d *dotfileConfigAdapter) GetExpandDirectories() []string {
	if d.cfg != nil {
		return d.cfg.GetExpandDirectories()
	}
	return []string{}
}

// CreateDotfileProvider creates a dotfile provider
func CreateDotfileProvider(homeDir string, configDir string, cfg *config.Config) *state.DotfileProvider {
	return state.NewDotfileProvider(homeDir, configDir, &dotfileConfigAdapter{cfg: cfg})
}
