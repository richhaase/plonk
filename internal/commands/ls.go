// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/orchestrator"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List managed items",
	Long: `Show overview of managed packages and dotfiles.

By default, shows a smart overview with both packages and dotfiles.
Use flags to filter to specific types or show detailed information.

Examples:
  plonk ls                    # Show smart overview (packages + dotfiles summary)
  plonk ls -v                 # Show detailed view with all items including untracked
  plonk ls --packages         # Show packages only
  plonk ls --dotfiles         # Show dotfiles only
  plonk ls -a                 # Include untracked items in overview
  plonk ls --brew             # Show only Homebrew packages
  plonk ls --npm              # Show only NPM packages
  plonk ls --cargo            # Show only Cargo packages
  plonk ls --pip              # Show only pip packages
  plonk ls --gem              # Show only gem packages
  plonk ls --go               # Show only go packages`,
	RunE: runLs,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Type filter flags
	lsCmd.Flags().Bool("packages", false, "Show packages only")
	lsCmd.Flags().Bool("dotfiles", false, "Show dotfiles only")
	lsCmd.MarkFlagsMutuallyExclusive("packages", "dotfiles")

	// Manager filter flags (mutually exclusive)
	lsCmd.Flags().Bool("brew", false, "Show Homebrew packages only")
	lsCmd.Flags().Bool("npm", false, "Show NPM packages only")
	lsCmd.Flags().Bool("cargo", false, "Show Cargo packages only")
	lsCmd.Flags().Bool("pip", false, "Show pip packages only")
	lsCmd.Flags().Bool("gem", false, "Show gem packages only")
	lsCmd.Flags().Bool("go", false, "Show go packages only")
	lsCmd.MarkFlagsMutuallyExclusive("brew", "npm", "cargo", "pip", "gem", "go")

	// Detail flags
	lsCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
	lsCmd.Flags().BoolP("all", "a", false, "Include untracked items")
}

func runLs(cmd *cobra.Command, args []string) error {
	// Parse flags
	flags, err := ParseSimpleFlags(cmd)
	if err != nil {
		return fmt.Errorf("invalid flag combination: %w", err)
	}

	// Parse output format
	format, err := ParseOutputFormat(flags.Output)
	if err != nil {
		return err
	}

	// Get filter flags
	packagesOnly, _ := cmd.Flags().GetBool("packages")
	dotfilesOnly, _ := cmd.Flags().GetBool("dotfiles")
	showAll, _ := cmd.Flags().GetBool("all")

	// Handle specific filters
	if packagesOnly {
		return runPackageList(cmd, flags, format)
	}
	if dotfilesOnly {
		return runDotfileList(cmd, flags, format)
	}

	// Handle manager-specific filters
	if flags.Manager != "" {
		return runManagerSpecificList(cmd, flags, format)
	}

	// Default: smart overview
	return runSmartOverview(cmd, flags, format, showAll)
}

// runSmartOverview provides a unified view of packages and dotfiles
func runSmartOverview(cmd *cobra.Command, flags *SimpleFlags, format OutputFormat, showAll bool) error {
	// Get directories
	homeDir := orchestrator.GetHomeDir()
	configDir := orchestrator.GetConfigDir()

	// Reconcile all domains
	ctx := context.Background()
	results, err := orchestrator.ReconcileAll(ctx, homeDir, configDir)
	if err != nil {
		return fmt.Errorf("failed to reconcile state: %w", err)
	}

	// Convert results to summary
	summary := &resources.Summary{
		TotalManaged:   0,
		TotalMissing:   0,
		TotalUntracked: 0,
		Results:        make([]resources.Result, 0),
	}

	for _, result := range results {
		result.AddToSummary(summary)
	}
	if err != nil {
		return fmt.Errorf("failed to reconcile state: %w", err)
	}

	// Prepare smart overview data
	overviewData := SmartOverviewOutput{
		TotalManaged:   summary.TotalManaged,
		TotalMissing:   summary.TotalMissing,
		TotalUntracked: summary.TotalUntracked,
		Verbose:        flags.Verbose,
		ShowAll:        showAll,
		Domains:        make([]DomainOverview, 0),
	}

	// Process each domain
	for _, result := range summary.Results {
		if result.IsEmpty() {
			continue
		}

		domain := DomainOverview{
			Name:           result.Domain,
			Manager:        result.Manager,
			ManagedCount:   len(result.Managed),
			MissingCount:   len(result.Missing),
			UntrackedCount: len(result.Untracked),
			Items:          make([]SmartOverviewItem, 0),
		}

		// Add items based on verbose and showAll flags
		allItems := append(result.Managed, result.Missing...)
		if showAll {
			allItems = append(allItems, result.Untracked...)
		}

		for _, item := range allItems {
			domain.Items = append(domain.Items, SmartOverviewItem{
				Name:    item.Name,
				State:   item.State.String(),
				Manager: item.Manager,
			})
		}

		// Sort items by state then name
		sort.Slice(domain.Items, func(i, j int) bool {
			stateOrder := map[string]int{"managed": 0, "missing": 1, "untracked": 2}
			if stateOrder[domain.Items[i].State] != stateOrder[domain.Items[j].State] {
				return stateOrder[domain.Items[i].State] < stateOrder[domain.Items[j].State]
			}
			return strings.ToLower(domain.Items[i].Name) < strings.ToLower(domain.Items[j].Name)
		})

		overviewData.Domains = append(overviewData.Domains, domain)
	}

	// Sort domains by name
	sort.Slice(overviewData.Domains, func(i, j int) bool {
		return overviewData.Domains[i].Name < overviewData.Domains[j].Name
	})

	return RenderOutput(overviewData, format)
}

// runPackageList shows packages only (reuses existing logic)
func runPackageList(cmd *cobra.Command, flags *SimpleFlags, format OutputFormat) error {
	// Delegate to the existing package list implementation
	return runPkgList(cmd, []string{})
}

// runDotfileList shows dotfiles only (delegates to existing implementation)
func runDotfileList(cmd *cobra.Command, flags *SimpleFlags, format OutputFormat) error {
	// Delegate to the shared dotfile listing implementation
	return runDotList(cmd, []string{})
}

// runManagerSpecificList shows packages for a specific manager
func runManagerSpecificList(cmd *cobra.Command, flags *SimpleFlags, format OutputFormat) error {
	// Set the manager flag and delegate to package list
	cmd.Flags().Set("manager", flags.Manager)
	return runPackageList(cmd, flags, format)
}

// SmartOverviewOutput represents the output structure for smart overview
type SmartOverviewOutput struct {
	TotalManaged   int              `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int              `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int              `json:"total_untracked" yaml:"total_untracked"`
	Verbose        bool             `json:"verbose" yaml:"verbose"`
	ShowAll        bool             `json:"show_all" yaml:"show_all"`
	Domains        []DomainOverview `json:"domains" yaml:"domains"`
}

// DomainOverview represents overview information for a specific domain
type DomainOverview struct {
	Name           string              `json:"name" yaml:"name"`
	Manager        string              `json:"manager,omitempty" yaml:"manager,omitempty"`
	ManagedCount   int                 `json:"managed_count" yaml:"managed_count"`
	MissingCount   int                 `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int                 `json:"untracked_count" yaml:"untracked_count"`
	Items          []SmartOverviewItem `json:"items,omitempty" yaml:"items,omitempty"`
}

// SmartOverviewItem represents an item in the smart overview
type SmartOverviewItem struct {
	Name    string `json:"name" yaml:"name"`
	State   string `json:"state" yaml:"state"`
	Manager string `json:"manager,omitempty" yaml:"manager,omitempty"`
}

// TableOutput generates human-friendly table output for smart overview
func (s SmartOverviewOutput) TableOutput() string {
	var output strings.Builder

	// Header
	output.WriteString("Plonk Overview\n")
	output.WriteString("==============\n")
	output.WriteString(fmt.Sprintf("Total: %d managed | %d missing | %d untracked\n\n",
		s.TotalManaged, s.TotalMissing, s.TotalUntracked))

	// If nothing to show
	if len(s.Domains) == 0 {
		output.WriteString("No managed items found\n")
		return output.String()
	}

	// Domain summaries
	for _, domain := range s.Domains {
		displayName := strings.Title(domain.Name)
		if domain.Manager != "" {
			switch domain.Manager {
			case "homebrew":
				displayName = "Homebrew"
			case "npm":
				displayName = "NPM"
			case "cargo":
				displayName = "Cargo"
			default:
				displayName = strings.Title(domain.Manager)
			}
		}

		output.WriteString(fmt.Sprintf("%s:\n", displayName))

		// Summary line
		parts := make([]string, 0)
		if domain.ManagedCount > 0 {
			parts = append(parts, fmt.Sprintf("%d managed", domain.ManagedCount))
		}
		if domain.MissingCount > 0 {
			parts = append(parts, fmt.Sprintf("%d missing", domain.MissingCount))
		}
		if domain.UntrackedCount > 0 && s.ShowAll {
			parts = append(parts, fmt.Sprintf("%d untracked", domain.UntrackedCount))
		}

		if len(parts) > 0 {
			output.WriteString(fmt.Sprintf("  %s\n", strings.Join(parts, ", ")))
		}

		// Show items if verbose or if there are issues
		if s.Verbose || domain.MissingCount > 0 {
			if len(domain.Items) > 0 {
				itemsToShow := domain.Items
				if !s.Verbose && !s.ShowAll {
					// Show only managed and missing when not verbose
					itemsToShow = make([]SmartOverviewItem, 0)
					for _, item := range domain.Items {
						if item.State == "managed" || item.State == "missing" {
							itemsToShow = append(itemsToShow, item)
						}
					}
				}

				for _, item := range itemsToShow {
					var statusIcon string
					switch item.State {
					case "managed":
						statusIcon = "✓"
					case "missing":
						statusIcon = "⚠"
					case "untracked":
						statusIcon = "?"
					default:
						statusIcon = "-"
					}
					output.WriteString(fmt.Sprintf("    %s %s\n", statusIcon, item.Name))
				}
			}
		}

		output.WriteString("\n")
	}

	// Hints
	if !s.Verbose && s.TotalUntracked > 0 {
		output.WriteString(fmt.Sprintf("Use 'plonk ls -v' to see all items or 'plonk ls -a' to include %d untracked items\n", s.TotalUntracked))
	}

	return output.String()
}

// StructuredData returns the structured data for serialization
func (s SmartOverviewOutput) StructuredData() any {
	return s
}
