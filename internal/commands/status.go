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

	// Load configuration
	cfg, err := config.LoadConfig(configDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, errors.DomainConfig, "load", "failed to load configuration")
	}

	// Create unified state reconciler
	reconciler := state.NewReconciler()

	// Register package provider (multi-manager)
	ctx := context.Background()
	packageProvider, err := createPackageProvider(ctx, cfg)
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

	// Prepare output
	outputData := StatusOutput{
		ConfigPath:   filepath.Join(configDir, "plonk.yaml"),
		ConfigValid:  true,
		StateSummary: summary,
	}

	return RenderOutput(outputData, format)
}

// createPackageProvider creates a multi-manager package provider
func createPackageProvider(ctx context.Context, cfg *config.Config) (*state.MultiManagerPackageProvider, error) {
	provider := state.NewMultiManagerPackageProvider()
	
	// Create config adapter using new clean interface
	configAdapter := config.NewConfigAdapter(cfg)
	packageConfigAdapter := config.NewStatePackageConfigAdapter(configAdapter)
	
	// Add Homebrew manager
	homebrewManager := managers.NewHomebrewManager()
	available, err := homebrewManager.IsAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check homebrew availability: %w", err)
	}
	if available {
		managerAdapter := state.NewManagerAdapter(homebrewManager)
		provider.AddManager("homebrew", managerAdapter, packageConfigAdapter)
	}
	
	// Add NPM manager
	npmManager := managers.NewNpmManager()
	available, err = npmManager.IsAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check npm availability: %w", err)
	}
	if available {
		managerAdapter := state.NewManagerAdapter(npmManager)
		provider.AddManager("npm", managerAdapter, packageConfigAdapter)
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
	ConfigValid  bool          `json:"config_valid" yaml:"config_valid"`
	StateSummary state.Summary `json:"state_summary" yaml:"state_summary"`
}

// TableOutput generates human-friendly table output for status
func (s StatusOutput) TableOutput() string {
	output := "Plonk Status\n"
	output += "============\n\n"

	// Configuration status
	configStatus := "❌"
	if s.ConfigValid {
		configStatus = "✅"
	}
	output += fmt.Sprintf("📁 Config: %s (%s valid)\n\n", s.ConfigPath, configStatus)

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
	return s
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