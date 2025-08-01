// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/orchestrator"
	plonkoutput "github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/spf13/cobra"
)

// Status command implementation using unified state management system

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Display overall plonk status",
	Long: `Display a detailed list of all plonk-managed items and their status.

Shows:
- All managed packages and dotfiles
- Missing items that need to be installed
- Configuration and lock file status
- Unmanaged items (with --unmanaged flag)

Flag Behavior:
- --packages and --dotfiles: Filter by resource type (both shown if neither specified)
- --unmanaged and --missing: Filter by state (mutually exclusive)
- Combinations work as expected: --packages --missing shows only missing packages

Examples:
  plonk status              # Show all managed items
  plonk status --packages   # Show only packages
  plonk status --dotfiles   # Show only dotfiles
  plonk status --unmanaged  # Show only unmanaged items
  plonk status --missing    # Show only missing resources
  plonk status --missing --packages  # Show only missing packages
  plonk status -o json      # Show as JSON
  plonk status -o yaml      # Show as YAML`,
	RunE:         runStatus,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("packages", false, "Show only package status")
	statusCmd.Flags().Bool("dotfiles", false, "Show only dotfile status")
	statusCmd.Flags().Bool("unmanaged", false, "Show only unmanaged items")
	statusCmd.Flags().Bool("missing", false, "Show only missing resources")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return err
	}

	// Get filter flags
	showPackages, _ := cmd.Flags().GetBool("packages")
	showDotfiles, _ := cmd.Flags().GetBool("dotfiles")
	showUnmanaged, _ := cmd.Flags().GetBool("unmanaged")
	showMissing, _ := cmd.Flags().GetBool("missing")

	// Validate mutually exclusive flags
	if showUnmanaged && showMissing {
		return fmt.Errorf("--unmanaged and --missing are mutually exclusive: items cannot be both untracked and missing")
	}

	// If neither flag is set, show both
	if !showPackages && !showDotfiles {
		showPackages = true
		showDotfiles = true
	}

	// Get directories
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	// Load configuration (may fail if config is invalid, but we handle this gracefully)
	_, configLoadErr := config.Load(configDir)

	// Reconcile all domains
	ctx := context.Background()
	results, err := orchestrator.ReconcileAll(ctx, homeDir, configDir)
	if err != nil {
		return err
	}

	// Convert results to summary for compatibility with existing output logic
	summary := resources.ConvertResultsToSummary(results)

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
		ConfigPath:    configPath,
		LockPath:      lockPath,
		ConfigExists:  configExists,
		ConfigValid:   configValid,
		LockExists:    lockExists,
		StateSummary:  summary,
		ShowPackages:  showPackages,
		ShowDotfiles:  showDotfiles,
		ShowUnmanaged: showUnmanaged,
		ShowMissing:   showMissing,
		ConfigDir:     configDir,
	}

	return RenderOutput(outputData, format)
}

// sortItems sorts a slice of resources.Item alphabetically by name (case-insensitive)
func sortItems(items []resources.Item) {
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}

// sortItemsByManager sorts items first by manager, then by name (case-insensitive)
func sortItemsByManager(items map[string][]resources.Item) []string {
	// Get sorted manager names
	managers := make([]string, 0, len(items))
	for manager := range items {
		managers = append(managers, manager)
	}
	sort.Strings(managers)

	// Sort items within each manager
	for _, manager := range managers {
		sortItems(items[manager])
	}

	return managers
}

// Removed - using config.ConfigAdapter instead

// StatusOutput represents the output structure for status command
type StatusOutput struct {
	ConfigPath    string            `json:"config_path" yaml:"config_path"`
	LockPath      string            `json:"lock_path" yaml:"lock_path"`
	ConfigExists  bool              `json:"config_exists" yaml:"config_exists"`
	ConfigValid   bool              `json:"config_valid" yaml:"config_valid"`
	LockExists    bool              `json:"lock_exists" yaml:"lock_exists"`
	StateSummary  resources.Summary `json:"state_summary" yaml:"state_summary"`
	ShowPackages  bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowDotfiles  bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowUnmanaged bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ShowMissing   bool              `json:"-" yaml:"-"` // Not included in JSON/YAML output
	ConfigDir     string            `json:"-" yaml:"-"` // Not included in JSON/YAML output
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
	TotalManaged   int                       `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int                       `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int                       `json:"total_untracked" yaml:"total_untracked"`
	Domains        []resources.DomainSummary `json:"domains" yaml:"domains"`
}

// ManagedItem represents a currently managed item
type ManagedItem struct {
	Name    string `json:"name" yaml:"name"`
	Domain  string `json:"domain" yaml:"domain"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
}

// TableOutput generates human-friendly table output for status
func (s StatusOutput) TableOutput() string {
	var output strings.Builder

	// Title
	output.WriteString("Plonk Status\n")
	output.WriteString("============\n\n")

	// Process results by domain
	var packageResult, dotfileResult *resources.Result
	for i := range s.StateSummary.Results {
		if s.StateSummary.Results[i].Domain == "package" {
			packageResult = &s.StateSummary.Results[i]
		} else if s.StateSummary.Results[i].Domain == "dotfile" {
			dotfileResult = &s.StateSummary.Results[i]
		}
	}

	// Show packages table if requested
	if s.ShowPackages && packageResult != nil {
		// Group packages by manager
		packagesByManager := make(map[string][]resources.Item)
		missingPackages := []resources.Item{}
		untrackedPackages := make(map[string][]resources.Item)

		if s.ShowUnmanaged {
			// Show only untracked items
			for _, item := range packageResult.Untracked {
				untrackedPackages[item.Manager] = append(untrackedPackages[item.Manager], item)
			}
		} else if s.ShowMissing {
			// Show only missing items
			for _, item := range packageResult.Missing {
				missingPackages = append(missingPackages, item)
			}
		} else {
			// Show managed and missing items
			for _, item := range packageResult.Managed {
				packagesByManager[item.Manager] = append(packagesByManager[item.Manager], item)
			}
			for _, item := range packageResult.Missing {
				missingPackages = append(missingPackages, item)
			}
		}

		// Sort missing packages
		sortItems(missingPackages)

		// Build packages table
		if len(packagesByManager) > 0 || len(missingPackages) > 0 || len(untrackedPackages) > 0 {
			output.WriteString("PACKAGES\n")
			output.WriteString("--------\n")

			// Create a table for packages
			pkgBuilder := NewStandardTableBuilder("")
			pkgBuilder.SetHeaders("NAME", "MANAGER", "STATUS")

			// Show managed packages by manager (unless showing only missing)
			if !s.ShowMissing {
				sortedManagers := sortItemsByManager(packagesByManager)
				for _, manager := range sortedManagers {
					packages := packagesByManager[manager]
					for _, pkg := range packages {
						pkgBuilder.AddRow(pkg.Name, manager, plonkoutput.Managed())
					}
				}
			}

			// Show missing packages
			for _, pkg := range missingPackages {
				pkgBuilder.AddRow(pkg.Name, pkg.Manager, plonkoutput.Missing())
			}

			// Show untracked packages when --unmanaged flag is set
			sortedUntrackedManagers := sortItemsByManager(untrackedPackages)
			for _, manager := range sortedUntrackedManagers {
				packages := untrackedPackages[manager]
				for _, pkg := range packages {
					pkgBuilder.AddRow(pkg.Name, manager, plonkoutput.Unmanaged())
				}
			}

			output.WriteString(pkgBuilder.Build())
			output.WriteString("\n")
		}
	}

	// Show dotfiles table if requested
	if s.ShowDotfiles && dotfileResult != nil {
		// Determine which items to show based on flags
		var itemsToShow []resources.Item
		if s.ShowUnmanaged {
			itemsToShow = dotfileResult.Untracked
		} else if s.ShowMissing {
			itemsToShow = dotfileResult.Missing
		} else {
			// Include managed and missing items
			itemsToShow = append(dotfileResult.Managed, dotfileResult.Missing...)
			// Also need to check for drifted items (they won't be in Managed due to GroupItemsByState)
			// We'll handle drifted items separately below
		}

		if len(itemsToShow) > 0 {
			output.WriteString("DOTFILES\n")
			output.WriteString("--------\n")

			// Create a table for dotfiles
			dotBuilder := NewStandardTableBuilder("")

			if s.ShowUnmanaged {
				// For unmanaged, use single column showing just the path
				dotBuilder.SetHeaders("UNMANAGED DOTFILES")

				// Load config to check expand directories
				cfg := config.LoadWithDefaults(s.ConfigDir)

				// Sort untracked dotfiles
				sortItems(dotfileResult.Untracked)

				// Show untracked dotfiles
				for _, item := range dotfileResult.Untracked {
					// Show the dotfile path with ~ notation
					path := "~/" + item.Name

					// Add trailing slash for unexpanded directories
					if itemPath, ok := item.Metadata["path"].(string); ok {
						if info, err := os.Stat(itemPath); err == nil && info.IsDir() {
							// Check if this directory is in ExpandDirectories
							isExpanded := false
							for _, expandDir := range cfg.ExpandDirectories {
								if item.Name == expandDir {
									isExpanded = true
									break
								}
							}
							// Add trailing slash if not expanded
							if !isExpanded {
								path += "/"
							}
						}
					}

					dotBuilder.AddRow(path)
				}
			} else {
				// For managed/missing, use the three-column format
				dotBuilder.SetHeaders("SOURCE", "TARGET", "STATUS")

				// Sort managed and missing dotfiles
				sortItems(dotfileResult.Managed)
				sortItems(dotfileResult.Missing)

				// Show managed dotfiles (unless showing only missing)
				if !s.ShowMissing {
					for _, item := range dotfileResult.Managed {
						// Use source from metadata if available, otherwise fall back to Name
						source := item.Name
						if src, ok := item.Metadata["source"].(string); ok {
							source = src
						}
						target := ""
						if dest, ok := item.Metadata["destination"].(string); ok {
							target = dest
						}
						// Check if this is actually a drifted file
						status := plonkoutput.Deployed()
						if item.State == resources.StateDegraded {
							status = plonkoutput.Drifted()
						}
						dotBuilder.AddRow(source, target, status)
					}
				}

				// Show missing dotfiles
				for _, item := range dotfileResult.Missing {
					// Use source from metadata if available, otherwise fall back to Name
					source := item.Name
					if src, ok := item.Metadata["source"].(string); ok {
						source = src
					}
					target := ""
					if dest, ok := item.Metadata["destination"].(string); ok {
						target = dest
					}
					dotBuilder.AddRow(source, target, plonkoutput.Missing())
				}
			}

			output.WriteString(dotBuilder.Build())
			output.WriteString("\n")
		}
	}

	// Add summary (skip for unmanaged or missing to avoid misleading counts)
	if !s.ShowUnmanaged && !s.ShowMissing {
		summary := s.StateSummary

		// Count drifted items separately
		driftedCount := 0
		for _, result := range s.StateSummary.Results {
			if result.Domain == "dotfile" {
				for _, item := range result.Managed {
					if item.State == resources.StateDegraded {
						driftedCount++
					}
				}
			}
		}

		// Adjust managed count to exclude drifted
		managedCount := summary.TotalManaged - driftedCount

		output.WriteString("Summary: ")
		output.WriteString(fmt.Sprintf("%d managed", managedCount))
		if summary.TotalMissing > 0 {
			output.WriteString(fmt.Sprintf(", %d missing", summary.TotalMissing))
		}
		if driftedCount > 0 {
			output.WriteString(fmt.Sprintf(", %d drifted", driftedCount))
		}
		output.WriteString("\n")
	}

	// If no output was generated (except for title), show helpful message
	outputStr := output.String()
	if outputStr == "Plonk Status\n============\n\n" || outputStr == "" {
		output.Reset()
		output.WriteString("Plonk Status\n")
		output.WriteString("============\n\n")
		output.WriteString("No items match the specified filters.\n")
		if s.ShowMissing {
			output.WriteString("(Great! Everything tracked is installed/deployed)\n")
		}
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (s StatusOutput) StructuredData() any {
	// Filter items based on flags
	var items []ManagedItem

	// Sort all items in results first
	for i := range s.StateSummary.Results {
		sortItems(s.StateSummary.Results[i].Managed)
		sortItems(s.StateSummary.Results[i].Missing)
		sortItems(s.StateSummary.Results[i].Untracked)
	}

	for _, result := range s.StateSummary.Results {
		// Apply domain filter
		if result.Domain == "package" && !s.ShowPackages {
			continue
		}
		if result.Domain == "dotfile" && !s.ShowDotfiles {
			continue
		}

		// Apply status filter
		if s.ShowUnmanaged {
			// Show only untracked items
			for _, item := range result.Untracked {
				items = append(items, ManagedItem{
					Name:    item.Name,
					Domain:  result.Domain,
					Manager: item.Manager,
				})
			}
		} else if s.ShowMissing {
			// Show only missing items
			for _, item := range result.Missing {
				items = append(items, ManagedItem{
					Name:    item.Name,
					Domain:  result.Domain,
					Manager: item.Manager,
				})
			}
		} else {
			// Show managed items
			for _, item := range result.Managed {
				items = append(items, ManagedItem{
					Name:    item.Name,
					Domain:  result.Domain,
					Manager: item.Manager,
				})
			}
		}
	}

	// Sort the final items list
	sort.Slice(items, func(i, j int) bool {
		// First sort by domain
		if items[i].Domain != items[j].Domain {
			return items[i].Domain < items[j].Domain
		}
		// Then by manager (for packages)
		if items[i].Manager != items[j].Manager {
			return items[i].Manager < items[j].Manager
		}
		// Finally by name (case-insensitive)
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

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
			Domains:        resources.CreateDomainSummary(s.StateSummary.Results),
		},
		ManagedItems: items,
	}
}
