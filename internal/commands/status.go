// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"plonk/internal/config"
	"plonk/internal/errors"
	"plonk/internal/lock"
	"plonk/internal/managers"
	"plonk/internal/state"

	"github.com/spf13/cobra"
)

// Status command implementation using unified state management system

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display overall plonk status",
	Long: `Display the complete status of your plonk-managed environment.

Shows:
- Configuration file status
- Package management state (managed/missing/untracked)
- Dotfile management state (managed/missing/untracked)
- List of all managed items

Examples:
  plonk status           # Show overall status
  plonk status -o json   # Show as JSON
  plonk status -o yaml   # Show as YAML`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "status", "output-format", "invalid output format")
	}

	// Get directories
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, errors.ErrFilePermission, errors.DomainCommands, "status", "failed to get home directory")
	}

	configDir := config.GetDefaultConfigDirectory()

	// Load configuration (may fail if config is invalid)
	cfg, configLoadErr := config.LoadConfig(configDir)
	if configLoadErr != nil {
		// For invalid config, use empty config and continue
		cfg = &config.Config{}
	}

	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider (multi-manager) - using lock file
	ctx := context.Background()
	packageProvider, err := createPackageProvider(ctx, configDir)
	if err != nil {
		return err
	}
	reconciler.RegisterProvider("package", packageProvider)

	// Register dotfile provider
	dotfileProvider := createDotfileProvider(homeDir, configDir, cfg)
	reconciler.RegisterProvider("dotfile", dotfileProvider)

	// Reconcile all domains
	summary, err := reconciler.ReconcileAll(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrReconciliation, errors.DomainState, "reconcile", "failed to reconcile state")
	}

	// Check file existence and validity
	configPath := filepath.Join(configDir, "plonk.yaml")
	lockPath := filepath.Join(configDir, "plonk.lock")

	configExists := false
	configValid := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		// Config is valid only if it loaded without error
		configValid = (configLoadErr == nil)
	}

	lockExists := false
	if _, err := os.Stat(lockPath); err == nil {
		lockExists = true
	}

	// Prepare output
	outputData := StatusOutput{
		ConfigPath:   configPath,
		LockPath:     lockPath,
		ConfigExists: configExists,
		ConfigValid:  configValid,
		LockExists:   lockExists,
		StateSummary: summary,
	}

	return RenderOutput(outputData, format)
}

// createPackageProvider creates a multi-manager package provider using lock file
func createPackageProvider(ctx context.Context, configDir string) (*state.MultiManagerPackageProvider, error) {
	provider := state.NewMultiManagerPackageProvider()

	// Create lock file adapter
	lockService := lock.NewYAMLLockService(configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Add Homebrew manager
	homebrewManager := managers.NewHomebrewManager()
	available, err := homebrewManager.IsAvailable(ctx)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "homebrew", "failed to check homebrew availability")
	}
	if available {
		managerAdapter := state.NewManagerAdapter(homebrewManager)
		provider.AddManager("homebrew", managerAdapter, lockAdapter)
	}

	// Add NPM manager
	npmManager := managers.NewNpmManager()
	available, err = npmManager.IsAvailable(ctx)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "npm", "failed to check npm availability")
	}
	if available {
		managerAdapter := state.NewManagerAdapter(npmManager)
		provider.AddManager("npm", managerAdapter, lockAdapter)
	}

	// Add Cargo manager
	cargoManager := managers.NewCargoManager()
	available, err = cargoManager.IsAvailable(ctx)
	if err != nil {
		return nil, errors.WrapWithItem(err, errors.ErrManagerUnavailable, errors.DomainPackages, "check", "cargo", "failed to check cargo availability")
	}
	if available {
		managerAdapter := state.NewManagerAdapter(cargoManager)
		provider.AddManager("cargo", managerAdapter, lockAdapter)
	}

	return provider, nil
}

// createDotfileProvider creates a dotfile provider
func createDotfileProvider(homeDir string, configDir string, cfg *config.Config) *state.DotfileProvider {
	configAdapter := config.NewConfigAdapter(cfg)
	dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
	return state.NewDotfileProvider(homeDir, configDir, dotfileConfigAdapter)
}

// Removed - using config.ConfigAdapter instead

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	ConfigPath   string        `json:"config_path" yaml:"config_path"`
	LockPath     string        `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool          `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool          `json:"config_valid" yaml:"config_valid"`
	LockExists   bool          `json:"lock_exists" yaml:"lock_exists"`
	StateSummary state.Summary `json:"state_summary" yaml:"state_summary"`
}

// StatusOutputSummary represents a summary-focused version for JSON/YAML output
type StatusOutputSummary struct {
	ConfigPath   string            `json:"config_path" yaml:"config_path"`
	LockPath     string            `json:"lock_path" yaml:"lock_path"`
	ConfigExists bool              `json:"config_exists" yaml:"config_exists"`
	ConfigValid  bool              `json:"config_valid" yaml:"config_valid"`
	LockExists   bool              `json:"lock_exists" yaml:"lock_exists"`
	Summary      StatusSummaryData `json:"summary" yaml:"summary"`
	ManagedItems []ManagedItem     `json:"managed_items" yaml:"managed_items"`
}

// StatusSummaryData represents aggregate counts and domain summaries
type StatusSummaryData struct {
	TotalManaged   int             `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int             `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int             `json:"total_untracked" yaml:"total_untracked"`
	Domains        []DomainSummary `json:"domains" yaml:"domains"`
}

// DomainSummary represents counts for a specific domain/manager
type DomainSummary struct {
	Domain         string `json:"domain" yaml:"domain"`
	Manager        string `json:"manager,omitempty" yaml:"manager,omitempty"`
	ManagedCount   int    `json:"managed_count" yaml:"managed_count"`
	MissingCount   int    `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int    `json:"untracked_count" yaml:"untracked_count"`
}

// ManagedItem represents a currently managed item
type ManagedItem struct {
	Name    string `json:"name" yaml:"name"`
	Domain  string `json:"domain" yaml:"domain"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
}

// TableOutput generates human-friendly table output for status
func (s StatusOutput) TableOutput() string {
	output := "Plonk Status\n"
	output += "============\n\n"

	// Configuration status
	configStatus := "ℹ️ using defaults"
	if s.ConfigExists {
		if s.ConfigValid {
			configStatus = "✅ valid"
		} else {
			configStatus = "❌ invalid"
		}
	}

	lockStatus := "ℹ️ using defaults"
	if s.LockExists {
		lockStatus = "✅ exists"
	}

	output += fmt.Sprintf("📁 Config: %s (%s)\n", s.ConfigPath, configStatus)
	output += fmt.Sprintf("🔒 Lock:   %s (%s)\n\n", s.LockPath, lockStatus)

	// Overall state summary
	summary := s.StateSummary
	output += "Overall State:\n"
	if summary.TotalManaged > 0 {
		output += fmt.Sprintf("  ✅ %d managed items\n", summary.TotalManaged)
	}
	if summary.TotalMissing > 0 {
		output += fmt.Sprintf("  ❌ %d missing items\n", summary.TotalMissing)
	}
	if summary.TotalUntracked > 0 {
		output += fmt.Sprintf("  🔍 %d untracked items\n", summary.TotalUntracked)
	}

	// Domain-specific details
	if len(summary.Results) > 0 {
		output += "\nDomain Details:\n"
		for _, result := range summary.Results {
			if result.IsEmpty() {
				continue
			}

			domainName := result.Domain
			if result.Manager != "" {
				domainName = fmt.Sprintf("%s (%s)", result.Domain, result.Manager)
			}

			output += fmt.Sprintf("  %s: ", domainName)
			parts := []string{}
			if len(result.Managed) > 0 {
				parts = append(parts, fmt.Sprintf("%d managed", len(result.Managed)))
			}
			if len(result.Missing) > 0 {
				parts = append(parts, fmt.Sprintf("%d missing", len(result.Missing)))
			}
			if len(result.Untracked) > 0 {
				parts = append(parts, fmt.Sprintf("%d untracked", len(result.Untracked)))
			}

			for i, part := range parts {
				if i > 0 {
					output += ", "
				}
				output += part
			}
			output += "\n"
		}
	}

	// Currently managed items
	output += "\nCurrently Managing:\n"
	hasItems := false
	for _, result := range summary.Results {
		if len(result.Managed) > 0 {
			hasItems = true
			emoji := "📦"
			if result.Domain == "dotfile" {
				emoji = "📄"
			}

			domainLabel := result.Domain
			if result.Manager != "" {
				domainLabel = result.Manager
			}

			itemNames := make([]string, len(result.Managed))
			for i, item := range result.Managed {
				itemNames[i] = item.Name
			}

			output += fmt.Sprintf("  %s %s: %s\n", emoji, domainLabel,
				joinWithLimit(itemNames, ", ", 5))
		}
	}

	if !hasItems {
		output += "  No items currently managed\n"
	}

	return output
}

// StructuredData returns the structured data for serialization
func (s StatusOutput) StructuredData() any {
	// Create a summary-focused version for structured output
	return StatusOutputSummary{
		ConfigPath:   s.ConfigPath,
		LockPath:     s.LockPath,
		ConfigExists: s.ConfigExists,
		ConfigValid:  s.ConfigValid,
		LockExists:   s.LockExists,
		Summary: StatusSummaryData{
			TotalManaged:   s.StateSummary.TotalManaged,
			TotalMissing:   s.StateSummary.TotalMissing,
			TotalUntracked: s.StateSummary.TotalUntracked,
			Domains:        createDomainSummary(s.StateSummary.Results),
		},
		ManagedItems: extractManagedItems(s.StateSummary.Results),
	}
}

// createDomainSummary creates domain summaries with counts only
func createDomainSummary(results []state.Result) []DomainSummary {
	var domains []DomainSummary
	for _, result := range results {
		if result.IsEmpty() {
			continue
		}
		domains = append(domains, DomainSummary{
			Domain:         result.Domain,
			Manager:        result.Manager,
			ManagedCount:   len(result.Managed),
			MissingCount:   len(result.Missing),
			UntrackedCount: len(result.Untracked),
		})
	}
	return domains
}

// extractManagedItems extracts only the managed items without full metadata
func extractManagedItems(results []state.Result) []ManagedItem {
	var items []ManagedItem
	for _, result := range results {
		for _, managed := range result.Managed {
			items = append(items, ManagedItem{
				Name:    managed.Name,
				Domain:  managed.Domain,
				Manager: managed.Manager,
			})
		}
	}
	return items
}

// joinWithLimit joins strings with a separator, truncating if too many
func joinWithLimit(items []string, sep string, limit int) string {
	if len(items) <= limit {
		return fmt.Sprintf("%s", items)
	}

	truncated := make([]string, limit)
	copy(truncated, items[:limit])
	result := fmt.Sprintf("%s", truncated)
	result += fmt.Sprintf(" (and %d more)", len(items)-limit)
	return result
}
